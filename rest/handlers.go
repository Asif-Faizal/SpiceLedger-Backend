package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	pb "github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"google.golang.org/grpc/metadata"
)

func (s *Server) withAuth(r *http.Request) context.Context {
	ctx := r.Context()
	auth := r.Header.Get("Authorization")
	if auth != "" {
		return metadata.AppendToOutgoingContext(ctx, "authorization", auth)
	}
	// Fallback to internal basic auth if no header provided
	return s.authCtx(ctx)
}

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

	resp, err := s.accountClient.CheckEmailExists(s.withAuth(r), email)
	if err != nil {
		s.logger.Service().Error().Err(err).Msg("failed to check if email exists")
		util.WriteGRPCErrorResponse(w, err)
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

	resp, err := s.accountClient.Login(s.withAuth(r), req.Email, req.Password, req.DeviceID)
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

	if _, err := s.accountClient.Logout(s.withAuth(r), accessToken, req.DeviceID); err != nil {
		util.WriteGRPCErrorResponse(w, err)
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

	resp, err := s.accountClient.RefreshToken(s.withAuth(r), req.RefreshToken, req.DeviceID)
	if err != nil {
		util.WriteJSONResponse(w, http.StatusUnauthorized, false, err.Error(), nil)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Token refreshed successfully", toAuthenticatedResponse(resp))
}

func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateOrUpdateAccount(w, r)
	case http.MethodGet:
		s.handleListAccounts(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCreateOrUpdateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateOrUpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSONResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	resp, err := s.accountClient.CreateOrUpdateAccount(s.withAuth(r), req.ID, req.Name, req.UserType, req.Email, req.Password)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Account created/updated successfully", toAccount(resp.Account))
}

func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	resp, err := s.accountClient.ListAccounts(s.withAuth(r), 0, 100)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Accounts listed successfully", ListAccountsResponse{
		Accounts: func() []*Account {
			accounts := make([]*Account, len(resp.Accounts))
			for i, a := range resp.Accounts {
				accounts[i] = toAccount(a)
			}
			return accounts
		}(),
	})
}

func (s *Server) handleAccountByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/accounts/")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	resp, err := s.accountClient.GetAccountByID(s.withAuth(r), id)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Account retrieved successfully", toAccount(resp.Account))
}

func (s *Server) handleCreateOrUpdateMerchantDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateOrUpdateMerchantDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSONResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	resp, err := s.accountClient.CreateOrUpdateMerchantDetails(s.withAuth(r), req.ID, req.AccountID, req.PhoneNumber, req.Address, req.City, req.State, req.Pincode)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Merchant details created/updated successfully", &MerchantDetails{
		ID:          resp.MerchantDetails.Id,
		AccountID:   resp.MerchantDetails.AccountId,
		PhoneNumber: resp.MerchantDetails.PhoneNumber,
		Address:     resp.MerchantDetails.Address,
		City:        resp.MerchantDetails.City,
		State:       resp.MerchantDetails.State,
		Pincode:     resp.MerchantDetails.Pincode,
	})
}

func (s *Server) handleGetMerchantDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/accounts/merchant-details/")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	resp, err := s.accountClient.GetMerchantDetails(s.withAuth(r), id)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Merchant details retrieved successfully", &MerchantDetails{
		ID:          resp.MerchantDetails.Id,
		AccountID:   resp.MerchantDetails.AccountId,
		PhoneNumber: resp.MerchantDetails.PhoneNumber,
		Address:     resp.MerchantDetails.Address,
		City:        resp.MerchantDetails.City,
		State:       resp.MerchantDetails.State,
		Pincode:     resp.MerchantDetails.Pincode,
	})
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
