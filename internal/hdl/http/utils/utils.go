package utils

import (
	"encoding/json"
	"github.com/JMURv/avito-spring/internal/hdl"
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

func TextResponse(w http.ResponseWriter, statusCode int, data []byte) {
	w.WriteHeader(statusCode)
	w.Write(data)
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

func Parse(r *http.Request, dst any) error {
	var err error
	if err = json.NewDecoder(r.Body).Decode(dst); err != nil {
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.Error(err),
		)
		return hdl.ErrDecodeRequest
	}

	return nil
}
