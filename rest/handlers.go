package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	pb "github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

	resp, err := s.controlClient.CheckEmailExists(s.withAuth(r), email)
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

	resp, err := s.controlClient.Login(s.withAuth(r), req.Email, req.Password, req.DeviceID)
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

	if _, err := s.controlClient.Logout(s.withAuth(r), accessToken, req.DeviceID); err != nil {
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

	resp, err := s.controlClient.RefreshToken(s.withAuth(r), req.RefreshToken, req.DeviceID)
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

	resp, err := s.controlClient.CreateOrUpdateAccount(s.withAuth(r), req.ID, req.Name, req.UserType, req.Email, req.Password)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Account created/updated successfully", &Account{
		ID:       resp.Account.Id,
		Name:     resp.Account.Name,
		UserType: resp.Account.Usertype,
		Email:    resp.Account.Email,
	})
}

func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	resp, err := s.controlClient.ListAccounts(s.withAuth(r), 0, 100)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Accounts listed successfully", ListAccountsResponse{
		Accounts: func() []*Account {
			accounts := make([]*Account, len(resp.Accounts))
			for i, a := range resp.Accounts {
				accounts[i] = &Account{
					ID:       a.Id,
					Name:     a.Name,
					UserType: a.Usertype,
					Email:    a.Email,
				}
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

	resp, err := s.controlClient.GetAccountByID(s.withAuth(r), id)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Account retrieved successfully", &Account{
		ID:       resp.Account.Id,
		Name:     resp.Account.Name,
		UserType: resp.Account.Usertype,
		Email:    resp.Account.Email,
	})
}

func (s *Server) handleGetAccountInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, err := s.controlClient.GetAccountInfo(s.withAuth(r), &pb.GetAccountInfoRequest{})
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Account info retrieved successfully", &Account{
		ID:       resp.Account.Id,
		Name:     resp.Account.Name,
		UserType: resp.Account.Usertype,
		Email:    resp.Account.Email,
	})
}

func (s *Server) handleMerchantDetails(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetMerchantDetails(w, r)
	case http.MethodPost:
		s.handleCreateOrUpdateMerchantDetails(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCreateOrUpdateMerchantDetails(w http.ResponseWriter, r *http.Request) {

	var req CreateOrUpdateMerchantDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSONResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	resp, err := s.controlClient.CreateOrUpdateMerchantDetails(s.withAuth(r), req.ID, req.AccountID, req.PhoneNumber, req.Address, req.City, req.State, req.Pincode)
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

func (s *Server) handleMerchantInfo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetMerchantInfo(w, r)
	case http.MethodPost:
		s.handleCreateOrUpdateMerchantInfo(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCreateOrUpdateMerchantInfo(w http.ResponseWriter, r *http.Request) {
	var req CreateOrUpdateMerchantInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSONResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	resp, err := s.controlClient.CreateOrUpdateMerchantInfo(s.withAuth(r), &pb.CreateOrUpdateMerchantInfoRequest{
		Id:          req.ID,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Pincode:     req.Pincode,
	})
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

	id := strings.TrimPrefix(r.URL.Path, "/accounts/merchant-details/")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	resp, err := s.controlClient.GetMerchantDetails(s.withAuth(r), id)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			util.WriteJSONResponse(w, http.StatusOK, true, "No merchant details found", map[string]interface{}{})
			return
		}
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

func (s *Server) handleGetMerchantInfo(w http.ResponseWriter, r *http.Request) {
	resp, err := s.controlClient.GetMerchantInfo(s.withAuth(r), &pb.GetMerchantInfoRequest{})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			util.WriteJSONResponse(w, http.StatusOK, true, "No merchant details found", map[string]interface{}{})
			return
		}
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

func (s *Server) handleCreateOrUpdateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateOrUpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSONResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	resp, err := s.controlClient.CreateOrUpdateProduct(s.withAuth(r), req.ID, req.Name, req.Category, req.Description, req.Status)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Product created/updated successfully", &Product{
		ID:          resp.Product.Id,
		Name:        resp.Product.Name,
		Category:    resp.Product.Category,
		Description: resp.Product.Description,
		Status:      resp.Product.Status,
	})
}

func (s *Server) handleListProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, err := s.controlClient.ListProducts(s.withAuth(r), 0, 100)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Products listed successfully", ListProductsResponse{
		Products: func() []*Product {
			products := make([]*Product, len(resp.Products))
			for i, p := range resp.Products {
				products[i] = &Product{
					ID:          p.Id,
					Name:        p.Name,
					Category:    p.Category,
					Description: p.Description,
					Status:      p.Status,
				}
			}
			return products
		}(),
	})
}

func (s *Server) handleCreateOrUpdateGrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateOrUpdateGradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSONResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	resp, err := s.controlClient.CreateOrUpdateGrade(s.withAuth(r), req.ID, req.ProductID, req.Name, req.Description, req.Status)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Grade created/updated successfully", &Grade{
		ID:          resp.Grade.Id,
		ProductID:   resp.Grade.ProductId,
		Name:        resp.Grade.Name,
		Description: resp.Grade.Description,
		Status:      resp.Grade.Status,
	})
}

func (s *Server) handleListGradesByProductId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID := r.URL.Query().Get("product_id")
	resp, err := s.controlClient.ListGradesByProductId(s.withAuth(r), productID, 0, 100)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Grades listed successfully", ListGradesByProductIdResponse{
		Grades: func() []*Grade {
			grades := make([]*Grade, len(resp.Grades))
			for i, g := range resp.Grades {
				grades[i] = &Grade{
					ID:          g.Id,
					ProductID:   g.ProductId,
					Name:        g.Name,
					Description: g.Description,
					Status:      g.Status,
				}
			}
			return grades
		}(),
	})
}

func (s *Server) handleCreateOrUpdateDailyPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateOrUpdateDailyPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSONResponse(w, http.StatusBadRequest, false, "invalid request body", nil)
		return
	}

	resp, err := s.controlClient.CreateOrUpdateDailyPrice(s.withAuth(r), req.ID, req.ProductID, req.GradeID, req.Price, req.Date, req.Time)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Daily price created/updated successfully", &DailyPrice{
		ID:        resp.DailyPrice.Id,
		ProductID: resp.DailyPrice.ProductId,
		GradeID:   resp.DailyPrice.GradeId,
		Price:     resp.DailyPrice.Price,
		Date:      resp.DailyPrice.Date,
		Time:      resp.DailyPrice.Time,
	})
}

func (s *Server) handleListDailyPricesByGradeId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gradeID := r.URL.Query().Get("grade_id")
	durationStr := r.URL.Query().Get("duration")
	dateStr := r.URL.Query().Get("date")

	var duration int
	if durationStr != "" {
		fmt.Sscanf(durationStr, "%d", &duration)
	}

	resp, err := s.controlClient.ListDailyPrices(s.withAuth(r), gradeID, dateStr, int32(duration))
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Daily prices listed successfully", ListDailyPricesResponse{
		DailyPrices: func() []*DailyPrice {
			dailyPrices := make([]*DailyPrice, len(resp.DailyPrices))
			for i, dp := range resp.DailyPrices {
				dailyPrices[i] = &DailyPrice{
					ID:        dp.Id,
					ProductID: dp.ProductId,
					GradeID:   dp.GradeId,
					Price:     dp.Price,
					Date:      dp.Date,
					Time:      dp.Time,
				}
			}
			return dailyPrices
		}(),
	})
}

func (s *Server) handleGetTodaysByGradeId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gradeID := r.URL.Query().Get("grade_id")
	dateStr := r.URL.Query().Get("date")

	resp, err := s.controlClient.GetTodaysPrice(s.withAuth(r), gradeID, dateStr)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Daily prices for grade listed successfully", GetTodaysPriceResponse{
		DailyPrices: func() []*DailyPrice {
			dailyPrices := make([]*DailyPrice, len(resp.DailyPrices))
			for i, dp := range resp.DailyPrices {
				dailyPrices[i] = &DailyPrice{
					ID:        dp.Id,
					ProductID: dp.ProductId,
					GradeID:   dp.GradeId,
					Price:     dp.Price,
					Date:      dp.Date,
					Time:      dp.Time,
				}
			}
			return dailyPrices
		}(),
	})
}

func (s *Server) handleGetTodaysByProductId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	productID := r.URL.Query().Get("product_id")
	dateStr := r.URL.Query().Get("date")

	resp, err := s.controlClient.GetTodaysByProductId(s.withAuth(r), productID, dateStr)
	if err != nil {
		util.WriteGRPCErrorResponse(w, err)
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, true, "Daily prices for product listed successfully", GetTodaysPriceByProductIdResponse{
		DailyPrices: func() []*DailyPrice {
			dailyPrices := make([]*DailyPrice, len(resp.DailyPrices))
			for i, dp := range resp.DailyPrices {
				dailyPrices[i] = &DailyPrice{
					ID:        dp.Id,
					ProductID: dp.ProductId,
					GradeID:   dp.GradeId,
					Price:     dp.Price,
					Date:      dp.Date,
					Time:      dp.Time,
				}
			}
			return dailyPrices
		}(),
	})
}

func toAuthenticatedResponse(resp interface{}) *AuthenticatedResponse {
	switch r := resp.(type) {
	case *pb.LoginResponse:
		return &AuthenticatedResponse{
			Account: &Account{
				ID:       r.Account.Id,
				Name:     r.Account.Name,
				UserType: r.Account.Usertype,
				Email:    r.Account.Email,
			},
			AccessToken:  r.AccessToken,
			RefreshToken: r.RefreshToken,
		}
	case *pb.RefreshTokenResponse:
		return &AuthenticatedResponse{
			Account: &Account{
				ID:       r.Account.Id,
				Name:     r.Account.Name,
				UserType: r.Account.Usertype,
				Email:    r.Account.Email,
			},
			AccessToken:  r.AccessToken,
			RefreshToken: r.RefreshToken,
		}
	default:
		return &AuthenticatedResponse{}
	}
}
