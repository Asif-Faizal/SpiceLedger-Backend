package rest

import "net/http"

func NewHandler(server *Server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/accounts/check-email", server.handleCheckEmail)
	mux.HandleFunc("/accounts/login", server.handleLogin)
	mux.HandleFunc("/accounts/logout", server.handleLogout)
	mux.HandleFunc("/accounts/refresh", server.handleRefreshToken)
	return mux
}
