package entity

import "time"

type Message struct {
	ID        int64      `db:"id"`
	UserID    int64      `db:"user_id"`
	Text      string     `db:"text"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}
