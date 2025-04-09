package http

import (
	"context"
	"fmt"
	"github.com/JMURv/avito-spring/internal/auth"
	"github.com/JMURv/avito-spring/internal/ctrl"
	mid "github.com/JMURv/avito-spring/internal/hdl/http/middleware"
	"github.com/JMURv/avito-spring/internal/hdl/http/utils"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Handler struct {
	srv  *http.Server
	ctrl ctrl.AppCtrl
	au   auth.Core
}

func New(ctrl ctrl.AppCtrl, au auth.Core) *Handler {
	return &Handler{
		ctrl: ctrl,
		au:   au,
	}
}

func (h *Handler) Start(port int) {
	mux := http.NewServeMux()

	RegisterRoutes(mux, h, h.au)
	mux.HandleFunc(
		"/health", func(w http.ResponseWriter, r *http.Request) {
			utils.SuccessResponse(w, http.StatusOK, "OK")
		},
	)

	handler := mid.LogMetrics(mux)
	handler = mid.RecoverPanic(handler)
	h.srv = &http.Server{
		Handler:      handler,
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
