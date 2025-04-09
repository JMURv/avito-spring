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
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func RegisterRoutes(mux *http.ServeMux, h *Handler, au auth.Core) {
	mux.HandleFunc(
		"/dummyLogin", mid.Apply(
			h.dummyLogin,
			mid.AllowedMethods(http.MethodPost),
		),
	)
	mux.HandleFunc(
		"/register", mid.Apply(
			h.register,
			mid.AllowedMethods(http.MethodPost),
		),
	)
	mux.HandleFunc(
		"/login", mid.Apply(
			h.login,
			mid.AllowedMethods(http.MethodPost),
		),
	)

	mux.HandleFunc(
		"/pvz", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				mid.Apply(
					h.getPVZ,
					mid.AllowedMethods(http.MethodGet),
					mid.Auth(au, md.ModeratorRole, md.EmployeeRole),
				)(w, r)
			case http.MethodPost:
				mid.Apply(
					h.createPVZ,
					mid.AllowedMethods(http.MethodPost),
					mid.Auth(au, md.ModeratorRole),
				)(w, r)
			}
		},
	)

	mux.HandleFunc(
		"/pvz/{id}/close_last_reception", mid.Apply(
			h.closeLastReception,
			mid.AllowedMethods(http.MethodPost),
			mid.Auth(au),
		),
	)
	mux.HandleFunc(
		"/pvz/{id}/delete_last_product", mid.Apply(
			h.deleteLastProduct,
			mid.AllowedMethods(http.MethodPost),
			mid.Auth(au, md.EmployeeRole),
		),
	)

	mux.HandleFunc(
		"/receptions", mid.Apply(
			h.createReception,
			mid.AllowedMethods(http.MethodPost),
			mid.Auth(au, md.EmployeeRole),
		),
	)

	mux.HandleFunc(
		"/products", mid.Apply(
			h.addItemToReception,
			mid.AllowedMethods(http.MethodPost),
			mid.Auth(au, md.EmployeeRole),
		),
	)
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
