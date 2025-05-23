package http

import (
	"context"
	"fmt"
	"github.com/JMURv/avito-spring/internal/auth"
	"github.com/JMURv/avito-spring/internal/ctrl"
	mid "github.com/JMURv/avito-spring/internal/hdl/http/middleware"
	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Handler struct {
	Router *chi.Mux
	srv    *http.Server
	ctrl   ctrl.AppCtrl
	au     auth.Core
}

func New(ctrl ctrl.AppCtrl, au auth.Core) *Handler {
	r := chi.NewRouter()
	return &Handler{
		Router: r,
		ctrl:   ctrl,
		au:     au,
	}
}

func (h *Handler) Start(port int) {
	h.Router.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Recoverer,
		middleware.Logger,
		mid.PromMetrics,
	)

	h.RegisterRoutes()
	h.srv = &http.Server{
		Handler:      h.Router,
		Addr:         fmt.Sprintf(":%v", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	zap.L().Info(
		"Starting HTTP server",
		zap.String("addr", h.srv.Addr),
	)
	err := h.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		zap.L().Debug("Server error", zap.Error(err))
	}
}

func (h *Handler) Close(ctx context.Context) error {
	return h.srv.Shutdown(ctx)
}
