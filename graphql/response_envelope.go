package graphql

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

type graphQLPayload struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError    `json:"errors"`
}

type graphQLError struct {
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions"`
}

func restResponseEnvelopeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)

		var payload graphQLPayload
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(rec.Body.Bytes())
			return
		}

		if len(payload.Errors) == 0 {
			util.WriteJSONResponse(w, http.StatusOK, true, "", rawJSONToValue(payload.Data))
			return
		}

		first := payload.Errors[0]
		message := util.CleanErrorMessage(first.Message)
		statusCode := httpStatusFromGraphQLError(first)
		util.WriteJSONResponse(w, statusCode, false, message, nil)
	})
}

func httpStatusFromGraphQLError(gqlErr graphQLError) int {
	if gqlErr.Extensions != nil {
		if codeName, ok := gqlErr.Extensions["grpc_code"].(string); ok {
			return util.HTTPStatusFromGRPCCodeName(codeName)
		}
	}
	return http.StatusBadRequest
}

func rawJSONToValue(raw json.RawMessage) interface{} {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	return value
}
