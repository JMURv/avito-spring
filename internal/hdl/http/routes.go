package http

import (
	"errors"
	"github.com/JMURv/avito-spring/internal/auth"
	"github.com/JMURv/avito-spring/internal/ctrl"
	"github.com/JMURv/avito-spring/internal/dto"
	"github.com/JMURv/avito-spring/internal/hdl"
	mid "github.com/JMURv/avito-spring/internal/hdl/http/middleware"
	"github.com/JMURv/avito-spring/internal/hdl/http/utils"
	md "github.com/JMURv/avito-spring/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (h *Handler) registerRoutes() {
	h.router.Get(
		"/health", func(w http.ResponseWriter, r *http.Request) {
			utils.SuccessResponse(w, http.StatusOK, "OK")
		},
	)

	h.router.Post("/dummyLogin", h.dummyLogin)
	h.router.Post("/register", h.register)
	h.router.Post("/login", h.login)
	h.router.Route(
		"/pvz", func(r chi.Router) {
			r.With(mid.Auth(h.au, md.ModeratorRole, md.EmployeeRole)).Get("/", h.getPVZ)
			r.With(mid.Auth(h.au, md.ModeratorRole)).Post("/", h.createPVZ)

			r.Route(
				"/{id}", func(r chi.Router) {
					r.With(mid.Auth(h.au)).Post("/close_last_reception", h.closeLastReception)
					r.With(mid.Auth(h.au, md.EmployeeRole)).Post("/delete_last_product", h.deleteLastProduct)
				},
			)
		},
	)

	h.router.With(mid.Auth(h.au, md.EmployeeRole)).Post("/receptions", h.createReception)
	h.router.With(mid.Auth(h.au, md.EmployeeRole)).Post("/products", h.addItemToReception)
}

func (h *Handler) dummyLogin(w http.ResponseWriter, r *http.Request) {
	req := &dto.DummyLoginRequest{}
	if err := utils.ParseAndValidate(r, req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.ctrl.DummyLogin(r.Context(), req)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	req := &dto.RegisterRequest{}
	if err := utils.ParseAndValidate(r, req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.ctrl.Register(r.Context(), req)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, res)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	req := &dto.LoginRequest{}
	if err := utils.ParseAndValidate(r, req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.ctrl.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			utils.ErrResponse(w, http.StatusUnauthorized, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) getPVZ(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil || limit < 1 {
		limit = 10
	}

	var startDate, endDate time.Time
	startDate, err = time.Parse(time.RFC3339, r.URL.Query().Get("startDate"))
	if err != nil {
		startDate = time.Now().AddDate(-1000, 0, 0)
		zap.L().Debug(
			"Invalid date format. Please, use RFC3339 format. No search by date will be applied.",
			zap.String("date", r.URL.Query().Get("startDate")),
		)
	}

	endDate, err = time.Parse(time.RFC3339, r.URL.Query().Get("endDate"))
	if err != nil {
		endDate = time.Now()
		zap.L().Debug(
			"Invalid date format. Please, use RFC3339 format. No search by date will be applied.",
			zap.String("date", r.URL.Query().Get("endDate")),
		)
	}

	res, err := h.ctrl.GetPVZ(r.Context(), page, limit, startDate, endDate)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) createPVZ(w http.ResponseWriter, r *http.Request) {
	req := &dto.CreatePVZRequest{}
	if err := utils.ParseAndValidate(r, req); err != nil {
		utils.ErrResponse(w, http.StatusUnauthorized, err)
		return
	}

	res, err := h.ctrl.CreatePVZ(r.Context(), req)
	if err != nil {
		if errors.Is(err, ctrl.ErrCityIsNotValid) {
			utils.ErrResponse(w, http.StatusBadRequest, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, res)
}

func (h *Handler) closeLastReception(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		utils.ErrResponse(w, http.StatusBadRequest, ErrInvalidPathSegments)
		return
	}

	pvzID, err := uuid.Parse(parts[2])
	if err != nil || pvzID == uuid.Nil {
		utils.ErrResponse(w, http.StatusBadRequest, ErrFailedToParseUUID)
		return
	}

	res, err := h.ctrl.CloseLastReception(r.Context(), pvzID)
	if err != nil {
		if errors.Is(err, ctrl.ErrReceptionAlreadyClosed) {
			utils.ErrResponse(w, http.StatusBadRequest, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, res)
}

func (h *Handler) deleteLastProduct(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		utils.ErrResponse(w, http.StatusBadRequest, ErrInvalidPathSegments)
		return
	}

	pvzID, err := uuid.Parse(parts[2])
	if err != nil || pvzID == uuid.Nil {
		zap.L().Debug("Failed to parse uuid", zap.String("uuid", parts[2]), zap.Error(err))
		utils.ErrResponse(w, http.StatusBadRequest, ErrFailedToParseUUID)
		return
	}

	err = h.ctrl.DeleteLastProduct(r.Context(), pvzID)
	if err != nil {
		if errors.Is(err, ctrl.ErrNoActiveReception) || errors.Is(err, ctrl.ErrNoItems) {
			utils.ErrResponse(w, http.StatusBadRequest, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.StatusResponse(w, http.StatusOK)
}

func (h *Handler) createReception(w http.ResponseWriter, r *http.Request) {
	req := &dto.CreateReceptionRequest{}
	if err := utils.ParseAndValidate(r, req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.ctrl.CreateReception(r.Context(), req)
	if err != nil {
		if errors.Is(err, ctrl.ErrReceptionStillOpen) {
			utils.ErrResponse(w, http.StatusBadRequest, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, res)
}

func (h *Handler) addItemToReception(w http.ResponseWriter, r *http.Request) {
	req := &dto.AddItemRequest{}
	if err := utils.ParseAndValidate(r, req); err != nil {
		utils.ErrResponse(w, http.StatusBadRequest, err)
		return
	}
	res, err := h.ctrl.AddItemToReception(r.Context(), req)
	if err != nil {
		if errors.Is(err, ctrl.ErrNoActiveReception) || errors.Is(err, ctrl.ErrTypeIsNotValid) {
			utils.ErrResponse(w, http.StatusBadRequest, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SuccessResponse(w, http.StatusCreated, res)
}
