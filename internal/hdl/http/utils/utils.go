package utils

import (
	"encoding/json"
	"github.com/JMURv/avito-spring/internal/hdl"
	"github.com/JMURv/avito-spring/internal/hdl/validation"
	"go.uber.org/zap"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func StatusResponse(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
}

func SuccessResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func ErrResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(
		&ErrorResponse{
			Message: err.Error(),
		},
	)
}

func ParseAndValidate(r *http.Request, dst any) error {
	var err error
	if err = json.NewDecoder(r.Body).Decode(dst); err != nil {
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.Error(err),
		)
		return hdl.ErrDecodeRequest
	}

	if err = validation.V.Struct(dst); err != nil {
		return err
	}

	return nil
}
