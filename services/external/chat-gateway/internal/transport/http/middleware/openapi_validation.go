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

// OpenAPIValidator validates incoming HTTP requests against the OpenAPI spec.
type OpenAPIValidator struct {
	router routers.Router
}

// NewOpenAPIValidator loads and validates an OpenAPI document and builds a router for request matching.
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

// Middleware returns an Echo middleware that validates requests for routes described in the OpenAPI spec.
func (v *OpenAPIValidator) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Validate only public API (everything under /api/).
			if !strings.HasPrefix(c.Request().URL.Path, "/api/") {
				return next(c)
			}

			req := c.Request()
			var bodyCopy []byte
			if req.Body != nil && req.Body != http.NoBody {
				// Keep a bounded copy so we can rewind body after validation.
				b, err := io.ReadAll(io.LimitReader(req.Body, 1<<20))
				if err != nil {
					return fmt.Errorf("read request body: %w", err)
				}
				bodyCopy = b
				req.Body = io.NopCloser(bytes.NewReader(bodyCopy))
			}

			route, pathParams, err := v.router.FindRoute(req)
			if err != nil {
				// If not found in OpenAPI, let regular routing handle it
				// (ws/docs/health/metrics are intentionally not part of OpenAPI).
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
