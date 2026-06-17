package util

import (
	"encoding/json"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"`
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

func HTTPStatusFromGRPCCode(code codes.Code) int {
	switch code {
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Unimplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

func HTTPStatusFromGRPCCodeName(codeName string) int {
	switch codeName {
	case codes.Unauthenticated.String():
		return http.StatusUnauthorized
	case codes.PermissionDenied.String():
		return http.StatusForbidden
	case codes.InvalidArgument.String():
		return http.StatusBadRequest
	case codes.NotFound.String():
		return http.StatusNotFound
	case codes.AlreadyExists.String():
		return http.StatusConflict
	case codes.DeadlineExceeded.String():
		return http.StatusGatewayTimeout
	case codes.Unimplemented.String():
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

// CleanErrorMessage strips gRPC wrappers like "rpc error: code = X desc = Y".
func CleanErrorMessage(msg string) string {
	if idx := strings.Index(msg, "desc = "); idx >= 0 {
		return strings.TrimSpace(msg[idx+len("desc = "):])
	}
	return msg
}

func WriteGRPCErrorResponse(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		WriteJSONResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		return
	}

	WriteJSONResponse(w, HTTPStatusFromGRPCCode(st.Code()), false, st.Message(), nil)
}
