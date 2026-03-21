package rest

import (
	"net/http"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

func NewHandler(server *Server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/accounts/check-email", server.handleCheckEmail)
	mux.HandleFunc("/accounts/login", server.handleLogin)
	mux.HandleFunc("/accounts/logout", server.handleLogout)
	mux.HandleFunc("/accounts/refresh", server.handleRefreshToken)
	mux.HandleFunc("/accounts", server.handleAccounts)
	mux.HandleFunc("/accounts/info", server.handleGetAccountInfo)
	mux.HandleFunc("/accounts/", server.handleAccountByID)
	mux.HandleFunc("/accounts/merchant-details", server.handleMerchantDetails)
	mux.HandleFunc("/accounts/merchant-info", server.handleMerchantInfo)
	mux.HandleFunc("/products", server.handleCreateOrUpdateProduct)
	mux.HandleFunc("/products/", server.handleListProducts)
	mux.HandleFunc("/grades", server.handleCreateOrUpdateGrade)
	mux.HandleFunc("/grades/", server.handleListGradesByProductId)
	mux.HandleFunc("/daily-prices", server.handleCreateOrUpdateDailyPrice)
	mux.HandleFunc("/daily-prices/", server.handleListDailyPricesByGradeId)
	mux.HandleFunc("/daily-prices/product/today/", server.handleGetTodaysByProductId)
	mux.HandleFunc("/daily-prices/grade/today/", server.handleGetTodaysByGradeId)
	return util.LoggingMiddleware(server.logger)(mux)
}
