package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type server struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

type registerRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type loginRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token    string `json:"token"`
	UserID   int64  `json:"userId"`
	Nickname string `json:"nickname"`
}

type messageRequest struct {
	Text string `json:"text"`
}

type messageResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId"`
	Nickname  string    `json:"nickname"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

func main() {
	ctx := context.Background()

	db, err := initPostgres(ctx)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	defer db.Close()

	rdb, err := initRedis(ctx)
	if err != nil {
		log.Fatalf("failed to init redis: %v", err)
	}
	defer func() {
		_ = rdb.Close()
	}()

	s := &server{
		db:    db,
		redis: rdb,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health/livez", s.handleLive)
	mux.HandleFunc("/health/readyz", s.handleReady)
	mux.HandleFunc("/api/register", s.handleRegister)
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/messages", s.handleMessages)

	port := getenv("HTTP_PORT", "8080")
	addr := ":" + port

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("chat-backend listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("http server error: %v", err)
	}
}

func initPostgres(ctx context.Context) (*pgxpool.Pool, error) {
	host := getenv("POSTGRES_HOST", "postgres")
	port := getenv("POSTGRES_PORT", "5432")
	dbname := getenv("POSTGRES_DB", "chat")
	user := getenv("POSTGRES_USER", "chat")
	password := getenv("POSTGRES_PASSWORD", "chat")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnIdleTime = time.Minute * 5

	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

func initRedis(ctx context.Context) (*redis.Client, error) {
	host := getenv("REDIS_HOST", "redis")
	port := getenv("REDIS_PORT", "6379")
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := getenv("REDIS_DB", "0")
	dbNum, err := strconv.Atoi(dbStr)
	if err != nil {
		dbNum = 0
	}

	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       dbNum,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

func (s *server) handleLive(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database is not ready")
		return
	}
	if err := s.redis.Ping(ctx).Err(); err != nil {
		writeError(w, http.StatusServiceUnavailable, "redis is not ready")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "nickname and password must be provided")
		return
	}

	if len(req.Nickname) > 64 {
		writeError(w, http.StatusBadRequest, "nickname is too long")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var existingID int64
	err := s.db.QueryRow(ctx, "SELECT id FROM chat_user WHERE nickname = $1", req.Nickname).Scan(&existingID)
	if err == nil {
		writeError(w, http.StatusConflict, "nickname is already taken")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	var newID int64
	err = s.db.QueryRow(
		ctx,
		"INSERT INTO chat_user (nickname, password_hash, created_at) VALUES ($1, $2, NOW()) RETURNING id",
		req.Nickname,
		string(hash),
	).Scan(&newID)
	if err != nil {
		log.Printf("failed to insert user: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	token, err := generateToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate session token")
		return
	}

	sessionKey := "session:" + token
	if err := s.redis.Set(ctx, sessionKey, newID, 24*time.Hour).Err(); err != nil {
		log.Printf("failed to store session: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to store session")
		return
	}

	resp := loginResponse{
		Token:    token,
		UserID:   newID,
		Nickname: req.Nickname,
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "nickname and password must be provided")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var (
		userID int64
		hash   string
	)
	err := s.db.QueryRow(ctx, "SELECT id, password_hash FROM chat_user WHERE nickname = $1", req.Nickname).
		Scan(&userID, &hash)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := generateToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate session token")
		return
	}

	sessionKey := "session:" + token
	if err := s.redis.Set(ctx, sessionKey, userID, 24*time.Hour).Err(); err != nil {
		log.Printf("failed to store session: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to store session")
		return
	}

	resp := loginResponse{
		Token:    token,
		UserID:   userID,
		Nickname: req.Nickname,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) handleMessages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListMessages(w, r)
	case http.MethodPost:
		s.handlePostMessage(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *server) handlePostMessage(w http.ResponseWriter, r *http.Request) {
	userID, nickname, err := s.authenticate(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req messageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		writeError(w, http.StatusBadRequest, "message text must be provided")
		return
	}
	if len(req.Text) > 2000 {
		writeError(w, http.StatusBadRequest, "message is too long")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var (
		id        int64
		createdAt time.Time
	)
	err = s.db.QueryRow(
		ctx,
		"INSERT INTO chat_message (user_id, text, created_at) VALUES ($1, $2, NOW()) RETURNING id, created_at",
		userID,
		req.Text,
	).Scan(&id, &createdAt)
	if err != nil {
		log.Printf("failed to insert message: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to store message")
		return
	}

	resp := messageResponse{
		ID:        id,
		UserID:    userID,
		Nickname:  nickname,
		Text:      req.Text,
		CreatedAt: createdAt,
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *server) handleListMessages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	afterIDStr := q.Get("after_id")
	limitStr := q.Get("limit")

	var (
		afterID int64
		limit   int64 = 50
		err     error
	)

	if afterIDStr != "" {
		afterID, err = strconv.ParseInt(afterIDStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "after_id must be an integer")
			return
		}
	}

	if limitStr != "" {
		limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil || limit <= 0 {
			writeError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if limit > 200 {
			limit = 200
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := s.db.Query(
		ctx,
		`SELECT m.id, u.id, u.nickname, m.text, m.created_at
         FROM chat_message m
         JOIN chat_user u ON m.user_id = u.id
         WHERE m.id > $1
         ORDER BY m.id ASC
         LIMIT $2`,
		afterID,
		limit,
	)
	if err != nil {
		log.Printf("failed to query messages: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load messages")
		return
	}
	defer rows.Close()

	var messages []messageResponse
	for rows.Next() {
		var m messageResponse
		if err := rows.Scan(&m.ID, &m.UserID, &m.Nickname, &m.Text, &m.CreatedAt); err != nil {
			log.Printf("failed to scan message row: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to read messages")
			return
		}
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		log.Printf("rows error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to read messages")
		return
	}

	writeJSON(w, http.StatusOK, messages)
}

func (s *server) authenticate(r *http.Request) (int64, string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, "", errors.New("authorization header is missing")
	}

	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return 0, "", errors.New("authorization header must be in the format 'Bearer <token>'")
	}

	token := parts[1]
	if token == "" {
		return 0, "", errors.New("token is empty")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	sessionKey := "session:" + token
	val, err := s.redis.Get(ctx, sessionKey).Result()
	if err == redis.Nil {
		return 0, "", errors.New("session not found")
	}
	if err != nil {
		log.Printf("redis error: %v", err)
		return 0, "", errors.New("failed to validate session")
	}

	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, "", errors.New("invalid session value")
	}

	var nickname string
	err = s.db.QueryRow(ctx, "SELECT nickname FROM chat_user WHERE id = $1", userID).Scan(&nickname)
	if err != nil {
		return 0, "", errors.New("user not found")
	}

	return userID, nickname, nil
}

func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func getenv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
