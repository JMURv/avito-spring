package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/JMURv/avito-spring/internal/auth"
	"github.com/JMURv/avito-spring/internal/hdl/http/utils"
	metrics "github.com/JMURv/avito-spring/internal/observability/metrics/prometheus"
	"net/http"
	"slices"
	"strings"
	"time"
)

var ErrNotAuthorized = errors.New("not authorized")
var ErrAuthHeaderIsMissing = errors.New("authorization header is missing")
var ErrInvalidTokenFormat = errors.New("invalid token format")

func Auth(au auth.Core, allowedRole ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				header := r.Header.Get("Authorization")
				if header == "" {
					utils.ErrResponse(w, http.StatusForbidden, ErrAuthHeaderIsMissing)
					return
				}

				token := strings.TrimPrefix(header, "Bearer ")
				if token == header {
					utils.ErrResponse(w, http.StatusForbidden, ErrInvalidTokenFormat)
					return
				}

				claims, err := au.ParseClaims(r.Context(), token)
				if err != nil {
					utils.ErrResponse(w, http.StatusForbidden, err)
					return
				}

				if len(allowedRole) > 0 {
					if !slices.Contains(allowedRole, claims.Role) {
						utils.ErrResponse(w, http.StatusForbidden, ErrNotAuthorized)
						return
					}
				}

				ctx := context.WithValue(r.Context(), "role", claims.Role)
				ctx = context.WithValue(ctx, "uid", claims.UID)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{w, http.StatusOK}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func PromMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			s := time.Now()
			lrw := NewLoggingResponseWriter(w)
			next.ServeHTTP(lrw, r)
			metrics.ObserveRequest(time.Since(s), lrw.statusCode, fmt.Sprintf("%s %s", r.Method, r.RequestURI))
		},
	)
}
