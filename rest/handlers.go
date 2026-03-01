package rest

import (
	"encoding/json"
	"net/http"

	pb "github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	util.WriteJSONResponse(w, http.StatusOK, true, "", map[string]string{
		"service": "rest",
	})
}

func (s *Server) handleCheckEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	resp, err := s.accountClient.CheckEmailExists(r.Context(), email)
	if err != nil {
		s.logger.Service().Error().Err(err).Msg("failed to check if email exists")
		util.WriteJSONResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		return
	}

	message := ""
	if resp.Exists {
		message = "Email already exists"
	} else {
		message = "Email is available"
	}

	util.WriteJSONResponse(w, http.StatusOK, true, message, map[string]bool{
		"exists": resp.Exists,
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSONResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	resp, err := s.accountClient.Login(r.Context(), req.Email, req.Password, req.DeviceID)
	if err != nil {
		util.WriteJSONResponse(w, http.StatusUnauthorized, false, err.Error(), nil)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Login successful", toAuthenticatedResponse(resp))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		util.WriteJSONResponse(w, http.StatusUnauthorized, false, "unauthorized", nil)
		return
	}
	accessToken := authHeader[7:]

	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSONResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	if _, err := s.accountClient.Logout(r.Context(), accessToken, req.DeviceID); err != nil {
		util.WriteJSONResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Logged out successfully", nil)
}

func (s *Server) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSONResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	resp, err := s.accountClient.RefreshToken(r.Context(), req.RefreshToken, req.DeviceID)
	if err != nil {
		util.WriteJSONResponse(w, http.StatusUnauthorized, false, err.Error(), nil)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Token refreshed successfully", toAuthenticatedResponse(resp))
}

func toAuthenticatedResponse(resp interface{}) *AuthenticatedResponse {
	switch r := resp.(type) {
	case *pb.LoginResponse:
		return &AuthenticatedResponse{
			Account:      toAccount(r.Account),
			AccessToken:  r.AccessToken,
			RefreshToken: r.RefreshToken,
		}
	case *pb.RefreshTokenResponse:
		return &AuthenticatedResponse{
			Account:      toAccount(r.Account),
			AccessToken:  r.AccessToken,
			RefreshToken: r.RefreshToken,
		}
	default:
		return &AuthenticatedResponse{}
	}
}
