package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/codex-k8s/project-example/libs/go/common/logger"
	"github.com/codex-k8s/project-example/services/external/chat-gateway/internal/domain/errs"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v5"
)

// ErrorHandler is the single HTTP error boundary: it maps errors to status + safe payload and logs once.
func ErrorHandler(log *slog.Logger) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		if r, ok := c.Response().(*echo.Response); ok && r.Committed {
			return
		}

		// Context cancellation is not treated as an application error.
		if errors.Is(err, context.Canceled) {
			_ = c.NoContent(499) // 499 (nginx: client closed request)
			return
		}

		// OpenAPI validation errors.
		var reqErr *openapi3filter.RequestError
		if errors.As(err, &reqErr) {
			_ = c.JSON(http.StatusBadRequest, map[string]any{
				"code":    "openapi_validation_failed",
				"message": "invalid request",
			})
			return
		}

		// echo HTTPError (e.g. bind/json decode).
		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			code := httpErr.Code
			msg := "request error"
			if code >= 500 {
				logger.WithContext(c.Request().Context(), log).Error("http error", "status", code, "err", err)
				msg = "internal error"
			}
			_ = c.JSON(code, map[string]any{"code": "http_error", "message": msg})
			return
		}

		// Gateway domain errors.
		var v errs.Validation
		if errors.As(err, &v) {
			_ = c.JSON(http.StatusBadRequest, map[string]any{"code": "validation_error", "message": "invalid request"})
			return
		}
		var u errs.Unauthorized
		if errors.As(err, &u) {
			_ = c.JSON(http.StatusUnauthorized, map[string]any{"code": "unauthorized", "message": "unauthorized"})
			return
		}
		var f errs.Forbidden
		if errors.As(err, &f) {
			_ = c.JSON(http.StatusForbidden, map[string]any{"code": "forbidden", "message": "forbidden"})
			return
		}
		var nf errs.NotFound
		if errors.As(err, &nf) {
			_ = c.JSON(http.StatusNotFound, map[string]any{"code": "not_found", "message": "not found"})
			return
		}
		var cf errs.Conflict
		if errors.As(err, &cf) {
			_ = c.JSON(http.StatusConflict, map[string]any{"code": "conflict", "message": "conflict"})
			return
		}

		logger.WithContext(c.Request().Context(), log).Error("request failed", "err", err)
		_ = c.JSON(http.StatusInternalServerError, map[string]any{"code": "internal_error", "message": "internal error"})
	}
}
