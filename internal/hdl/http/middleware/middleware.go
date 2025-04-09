package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/JMURv/avito-spring/internal/auth"
	"github.com/JMURv/avito-spring/internal/hdl/http/utils"
	metrics "github.com/JMURv/avito-spring/internal/observability/metrics/prometheus"
	"go.uber.org/zap"
	"net/http"
	"slices"
	"strings"
	"time"
)

var ErrNotAuthorized = errors.New("not authorized")
var ErrAuthHeaderIsMissing = errors.New("authorization header is missing")
var ErrInvalidTokenFormat = errors.New("invalid token format")

func Apply(h http.HandlerFunc, middleware ...func(http.Handler) http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var handler http.Handler = h
		for _, m := range middleware {
			handler = m(handler)
		}
		handler.ServeHTTP(w, r)
	}
}

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

func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					zap.L().Error("panic", zap.Any("err", err))
					utils.ErrResponse(w, http.StatusInternalServerError, errors.New("internal error"))
				}
			}()
			next.ServeHTTP(w, r)
		},
	)
}

var ErrMethodNotAllowed = errors.New("method not allowed")

func AllowedMethods(methods ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if ok := slices.Contains(methods, r.Method); !ok {
					utils.ErrResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
					return
				}
				next.ServeHTTP(w, r)
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

func LogMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			s := time.Now()
			lrw := NewLoggingResponseWriter(w)
			next.ServeHTTP(lrw, r)
			metrics.ObserveRequest(time.Since(s), lrw.statusCode, fmt.Sprintf("%s %s", r.Method, r.RequestURI))

			zap.L().Info(
				"<--",
				zap.String("method", r.Method),
				zap.Int("status", lrw.statusCode),
				zap.Any("duration", time.Since(s)),
				zap.String("uri", r.RequestURI),
			)
		},
	)
}
