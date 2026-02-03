package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v5"
)

type OpenAPIValidator struct {
	router routers.Router
}

func NewOpenAPIValidator(ctx context.Context, specPath string) (*OpenAPIValidator, error) {
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("load openapi: %w", err)
	}
	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("validate openapi: %w", err)
	}
	r, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("build openapi router: %w", err)
	}
	return &OpenAPIValidator{router: r}, nil
}

func (v *OpenAPIValidator) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Валидируем только публичный API (всё, что под /api/).
			if !strings.HasPrefix(c.Request().URL.Path, "/api/") {
				return next(c)
			}

			req := c.Request()
			var bodyCopy []byte
			if req.Body != nil && req.Body != http.NoBody {
				// Ограничение на размер для валидации/перемотки.
				b, err := io.ReadAll(io.LimitReader(req.Body, 1<<20))
				if err != nil {
					return fmt.Errorf("read request body: %w", err)
				}
				bodyCopy = b
				req.Body = io.NopCloser(bytes.NewReader(bodyCopy))
			}

			route, pathParams, err := v.router.FindRoute(req)
			if err != nil {
				// Если маршрут не найден в OpenAPI, отдаём управление обычному роутингу.
				// (например, ws/docs/health/metrics не описаны в OpenAPI)
				if req.Body != nil && bodyCopy != nil {
					req.Body = io.NopCloser(bytes.NewReader(bodyCopy))
				}
				return next(c)
			}

			in := &openapi3filter.RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			}
			if err := openapi3filter.ValidateRequest(c.Request().Context(), in); err != nil {
				return err
			}

			if req.Body != nil && bodyCopy != nil {
				req.Body = io.NopCloser(bytes.NewReader(bodyCopy))
			}
			return next(c)
		}
	}
}
