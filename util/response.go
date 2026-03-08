package util

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteJSONResponse(w http.ResponseWriter, code int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{
		Success: success,
		Message: message,
		Data:    data,
	})
}

func WriteGRPCErrorResponse(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		WriteJSONResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		return
	}

	code := http.StatusInternalServerError
	switch st.Code() {
	case codes.Unauthenticated:
		code = http.StatusUnauthorized
	case codes.PermissionDenied:
		code = http.StatusForbidden
	case codes.InvalidArgument:
		code = http.StatusBadRequest
	case codes.NotFound:
		code = http.StatusNotFound
	case codes.AlreadyExists:
		code = http.StatusConflict
	case codes.DeadlineExceeded:
		code = http.StatusGatewayTimeout
	case codes.Unimplemented:
		code = http.StatusNotImplemented
	}

	WriteJSONResponse(w, code, false, st.Message(), nil)
}
