package control

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	pb "github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

type GrpcServer struct {
	accountService Service
	logger         util.Logger
	config         *util.Config
	pb.UnimplementedControlServiceServer
}

func ListenGrpcServer(service Service, logger util.Logger, config *util.Config) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.ControlGrpcPort))
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			util.UnaryServerInterceptor(logger),
			util.AuthInterceptor(config.JWTSecret, config.BasicAuthUser, config.BasicAuthPass),
			SessionInterceptor(service, logger),
		)),
	)

	server := &GrpcServer{
		accountService: service,
		logger:         logger,
		config:         config,
	}
	pb.RegisterControlServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	logger.Transport().Info().Int("port", config.ControlGrpcPort).Msg("gRPC server listening")
	return grpcServer.Serve(lis)
}

func SessionInterceptor(service Service, logger util.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		accessToken, _ := ctx.Value(util.AccessTokenKey).(string)
		if accessToken != "" {
			accountService := service.(*AccountService)
			session, err := accountService.repository.GetSessionByAccessToken(ctx, accessToken)
			if err != nil || session == nil || session.IsRevoked {
				logger.Transport().Warn().Str("token", accessToken).Msg("Rejected revoked or missing session")
				return nil, status.Error(codes.Unauthenticated, "session revoked or invalid")
			}
		}

		return handler(ctx, req)
	}
}

func (server *GrpcServer) checkAdmin(ctx context.Context) error {
	isAdmin, ok := ctx.Value(util.IsAdminKey).(bool)
	if !ok || !isAdmin {
		return status.Error(codes.PermissionDenied, "admin access required")
	}
	return nil
}

func (server *GrpcServer) checkMerchant(ctx context.Context) error {
	isMerchant, ok := ctx.Value(util.IsMerchantKey).(bool)
	if !ok || !isMerchant {
		return status.Error(codes.PermissionDenied, "merchant access required")
	}
	return nil
}

func (server *GrpcServer) checkAuthenticated(ctx context.Context) error {
	isAuthenticated, ok := ctx.Value(util.IsAuthenticatedKey).(bool)
	if !ok || !isAuthenticated {
		return status.Error(codes.Unauthenticated, "authentication required")
	}
	return nil
}

func (server *GrpcServer) GetSystemMetrics(ctx context.Context, request *pb.GetSystemMetricsRequest) (*pb.GetSystemMetricsResponse, error) {
	if err := server.checkAdmin(ctx); err != nil {
		return nil, err
	}
	userCount, productCount, err := server.accountService.GetSystemMetrics(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GetSystemMetricsResponse{
		TotalUsers:    userCount,
		TotalProducts: productCount,
	}, nil
}

func (server *GrpcServer) CheckEmailExists(ctx context.Context, request *pb.CheckEmailExistsRequest) (*pb.CheckEmailExistsResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	exists, err := server.accountService.CheckEmailExists(ctx, request.Email)
	if err != nil {
		return nil, err
	}
	return &pb.CheckEmailExistsResponse{Exists: exists}, nil
}

func (server *GrpcServer) CreateOrUpdateAccount(ctx context.Context, request *pb.CreateOrUpdateAccountRequest) (*pb.CreateOrUpdateAccountResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	if request.Usertype == "" {
		return nil, fmt.Errorf("usertype is required")
	}
	if request.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	account, err := server.accountService.CreateOrUpdateAccount(ctx, &Account{
		ID:       request.Id,
		Name:     request.Name,
		UserType: request.Usertype,
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrUpdateAccountResponse{
		Account: &pb.Account{
			Id:       account.ID,
			Name:     account.Name,
			Usertype: account.UserType,
			Email:    account.Email,
		},
	}, nil
}

func (server *GrpcServer) GetAccountByID(ctx context.Context, request *pb.GetAccountByIDRequest) (*pb.GetAccountByIDResponse, error) {
	if err := server.checkAdmin(ctx); err != nil {
		return nil, err
	}
	account, err := server.accountService.GetAccountByID(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetAccountByIDResponse{
		Account: &pb.Account{
			Id:       account.ID,
			Name:     account.Name,
			Usertype: account.UserType,
			Email:    account.Email,
		},
	}, nil
}

func (server *GrpcServer) GetAccountInfo(ctx context.Context, request *pb.GetAccountInfoRequest) (*pb.GetAccountByIDResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	accountID, ok := ctx.Value(util.AccountIDKey).(string)
	if !ok || accountID == "" {
		return nil, status.Error(codes.Unauthenticated, "account id not found in context")
	}
	account, err := server.accountService.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return &pb.GetAccountByIDResponse{
		Account: &pb.Account{
			Id:       account.ID,
			Name:     account.Name,
			Usertype: account.UserType,
			Email:    account.Email,
		},
	}, nil
}

func (server *GrpcServer) ListAccounts(ctx context.Context, request *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	if err := server.checkAdmin(ctx); err != nil {
		return nil, err
	}
	domainAccounts, err := server.accountService.ListAccounts(ctx, uint(request.Skip), uint(request.Take))
	if err != nil {
		return nil, err
	}
	accounts := []*pb.Account{}
	for _, account := range domainAccounts {
		accounts = append(accounts, &pb.Account{
			Id:       account.ID,
			Name:     account.Name,
			Usertype: account.UserType,
			Email:    account.Email,
		})
	}
	return &pb.ListAccountsResponse{Accounts: accounts}, nil
}

func (server *GrpcServer) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	resp, err := server.accountService.Login(ctx, request.Email, request.Password, request.DeviceId)
	if err != nil {
		return nil, err
	}

	var account *pb.Account
	if resp.Account != nil {
		account = &pb.Account{
			Id:       resp.Account.ID,
			Name:     resp.Account.Name,
			Usertype: resp.Account.UserType,
			Email:    resp.Account.Email,
		}
	}

	return &pb.LoginResponse{
		Account:      account,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (server *GrpcServer) Logout(ctx context.Context, request *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	if err := server.accountService.Logout(ctx, request.AccessToken, request.DeviceId); err != nil {
		return nil, err
	}
	return &pb.LogoutResponse{Success: true}, nil
}

func (server *GrpcServer) RefreshToken(ctx context.Context, request *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	resp, err := server.accountService.RefreshToken(ctx, request.RefreshToken, request.DeviceId)
	if err != nil {
		return nil, err
	}

	var account *pb.Account
	if resp.Account != nil {
		account = &pb.Account{
			Id:       resp.Account.ID,
			Name:     resp.Account.Name,
			Usertype: resp.Account.UserType,
			Email:    resp.Account.Email,
		}
	}

	return &pb.RefreshTokenResponse{
		Account:      account,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (server *GrpcServer) CreateOrUpdateMerchantDetails(ctx context.Context, request *pb.CreateOrUpdateMerchantDetailsRequest) (*pb.CreateOrUpdateMerchantDetailsResponse, error) {
	if err := server.checkMerchant(ctx); err != nil {
		return nil, err
	}
	merchantDetails, err := server.accountService.CreateOrUpdateMerchantDetails(ctx, &MerchantDetails{
		ID:        request.Id,
		AccountID: request.AccountId,
		Phone:     request.PhoneNumber,
		Address:   request.Address,
		City:      request.City,
		State:     request.State,
		Pincode:   request.Pincode,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrUpdateMerchantDetailsResponse{
		MerchantDetails: &pb.MerchantDetails{
			Id:          merchantDetails.ID,
			AccountId:   merchantDetails.AccountID,
			PhoneNumber: merchantDetails.Phone,
			Address:     merchantDetails.Address,
			City:        merchantDetails.City,
			State:       merchantDetails.State,
			Pincode:     merchantDetails.Pincode,
		},
	}, nil
}

func (server *GrpcServer) CreateOrUpdateMerchantInfo(ctx context.Context, request *pb.CreateOrUpdateMerchantInfoRequest) (*pb.CreateOrUpdateMerchantDetailsResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	accountId, ok := ctx.Value(util.AccountIDKey).(string)
	if !ok || accountId == "" {
		return nil, status.Error(codes.Unauthenticated, "account id not found in context")
	}
	merchantDetails, err := server.accountService.CreateOrUpdateMerchantDetails(ctx, &MerchantDetails{
		ID:        request.Id,
		AccountID: accountId,
		Phone:     request.PhoneNumber,
		Address:   request.Address,
		City:      request.City,
		State:     request.State,
		Pincode:   request.Pincode,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrUpdateMerchantDetailsResponse{
		MerchantDetails: &pb.MerchantDetails{
			Id:          merchantDetails.ID,
			AccountId:   merchantDetails.AccountID,
			PhoneNumber: merchantDetails.Phone,
			Address:     merchantDetails.Address,
			City:        merchantDetails.City,
			State:       merchantDetails.State,
			Pincode:     merchantDetails.Pincode,
		},
	}, nil
}

func (server *GrpcServer) GetMerchantDetails(ctx context.Context, request *pb.GetMerchantDetailsRequest) (*pb.GetMerchantDetailsResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	merchantDetails, err := server.accountService.GetMerchantDetails(ctx, request.AccountId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "merchant details not found")
		}
		return nil, err
	}
	return &pb.GetMerchantDetailsResponse{
		MerchantDetails: &pb.MerchantDetails{
			Id:          merchantDetails.ID,
			AccountId:   merchantDetails.AccountID,
			PhoneNumber: merchantDetails.Phone,
			Address:     merchantDetails.Address,
			City:        merchantDetails.City,
			State:       merchantDetails.State,
			Pincode:     merchantDetails.Pincode,
		},
	}, nil
}

func (server *GrpcServer) GetMerchantInfo(ctx context.Context, request *pb.GetMerchantInfoRequest) (*pb.GetMerchantDetailsResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	accountId, ok := ctx.Value(util.AccountIDKey).(string)
	if !ok || accountId == "" {
		return nil, status.Error(codes.Unauthenticated, "account id not found in context")
	}
	merchantDetails, err := server.accountService.GetMerchantDetails(ctx, accountId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "merchant details not found")
		}
		return nil, err
	}
	return &pb.GetMerchantDetailsResponse{
		MerchantDetails: &pb.MerchantDetails{
			Id:          merchantDetails.ID,
			AccountId:   merchantDetails.AccountID,
			PhoneNumber: merchantDetails.Phone,
			Address:     merchantDetails.Address,
			City:        merchantDetails.City,
			State:       merchantDetails.State,
			Pincode:     merchantDetails.Pincode,
		},
	}, nil
}

// Products
func (server *GrpcServer) CreateOrUpdateProduct(ctx context.Context, request *pb.CreateOrUpdateProductRequest) (*pb.CreateOrUpdateProductResponse, error) {
	if err := server.checkAdmin(ctx); err != nil {
		return nil, err
	}
	product, err := server.accountService.CreateOrUpdateProduct(ctx, &Product{
		ID:          request.Id,
		Name:        request.Name,
		Category:    request.Category,
		Description: request.Description,
		Status:      request.Status,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrUpdateProductResponse{
		Product: &pb.Product{
			Id:          product.ID,
			Name:        product.Name,
			Category:    product.Category,
			Description: product.Description,
			Status:      product.Status,
		},
	}, nil
}

func (server *GrpcServer) ListProducts(ctx context.Context, request *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	products, err := server.accountService.ListProducts(ctx, uint(request.Skip), uint(request.Take))
	if err != nil {
		return nil, err
	}
	protoProducts := make([]*pb.Product, len(products))
	for i, p := range products {
		protoProducts[i] = &pb.Product{
			Id:          p.ID,
			Name:        p.Name,
			Category:    p.Category,
			Description: p.Description,
			Status:      p.Status,
		}
	}
	return &pb.ListProductsResponse{
		Products: protoProducts,
	}, nil
}

// Grades
func (server *GrpcServer) CreateOrUpdateGrade(ctx context.Context, request *pb.CreateOrUpdateGradeRequest) (*pb.CreateOrUpdateGradeResponse, error) {
	if err := server.checkAdmin(ctx); err != nil {
		return nil, err
	}
	grade, err := server.accountService.CreateOrUpdateGrade(ctx, &Grade{
		ID:          request.Id,
		ProductID:   request.ProductId,
		Name:        request.Name,
		Description: request.Description,
		Status:      request.Status,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrUpdateGradeResponse{
		Grade: &pb.Grade{
			Id:          grade.ID,
			ProductId:   grade.ProductID,
			Name:        grade.Name,
			Description: grade.Description,
			Status:      grade.Status,
		},
	}, nil
}

func (server *GrpcServer) ListGradesByProductId(ctx context.Context, request *pb.ListGradesByProductIdRequest) (*pb.ListGradesByProductIdResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	grades, err := server.accountService.ListGradesByProductId(ctx, request.ProductId, uint(request.Skip), uint(request.Take))
	if err != nil {
		return nil, err
	}
	protoGrades := make([]*pb.Grade, len(grades))
	for i, g := range grades {
		protoGrades[i] = &pb.Grade{
			Id:          g.ID,
			ProductId:   g.ProductID,
			Name:        g.Name,
			Description: g.Description,
			Status:      g.Status,
		}
	}
	return &pb.ListGradesByProductIdResponse{
		Grades: protoGrades,
	}, nil
}

// Daily Price
func (server *GrpcServer) CreateOrUpdateDailyPrice(ctx context.Context, request *pb.CreateOrUpdateDailyPriceRequest) (*pb.CreateOrUpdateDailyPriceResponse, error) {
	if err := server.checkAdmin(ctx); err != nil {
		return nil, err
	}
	date, err := time.Parse("2006-01-02", request.Date)
	if err != nil {
		date = time.Now()
	}
	t, err := time.Parse("15:04:05", request.Time)
	if err != nil {
		t = time.Now()
	}

	dailyPrice, err := server.accountService.CreateOrUpdateDailyPrice(ctx, &DailyPrice{
		ID:        request.Id,
		ProductID: request.ProductId,
		GradeID:   request.GradeId,
		Price:     request.Price,
		Date:      date,
		Time:      t,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrUpdateDailyPriceResponse{
		DailyPrice: &pb.DailyPrice{
			Id:        dailyPrice.ID,
			ProductId: dailyPrice.ProductID,
			GradeId:   dailyPrice.GradeID,
			Price:     dailyPrice.Price,
			Date:      dailyPrice.Date.Format("2006-01-02"),
			Time:      dailyPrice.Time.Format("15:04:05"),
		},
	}, nil
}

func (server *GrpcServer) ListDailyPrices(ctx context.Context, request *pb.ListDailyPricesRequest) (*pb.ListDailyPricesResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	today, _ := time.Parse("2006-01-02", request.Today)
	prices, err := server.accountService.ListDailyPricesByGradeId(ctx, request.GradeId, today, int(request.Duration))
	if err != nil {
		return nil, err
	}
	protoPrices := make([]*pb.DailyPrice, len(prices))
	for i, p := range prices {
		protoPrices[i] = &pb.DailyPrice{
			Id:        p.ID,
			ProductId: p.ProductID,
			GradeId:   p.GradeID,
			Price:     p.Price,
			Date:      p.Date.Format("2006-01-02"),
			Time:      p.Time.Format("15:04:05"),
		}
	}
	return &pb.ListDailyPricesResponse{
		DailyPrices: protoPrices,
	}, nil
}

func (server *GrpcServer) GetTodaysPrice(ctx context.Context, request *pb.GetTodaysPriceRequest) (*pb.GetTodaysPriceResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	date, _ := time.Parse("2006-01-02", request.Date)
	prices, err := server.accountService.GetTodaysByGradeId(ctx, request.GradeId, date)
	if err != nil {
		return nil, err
	}
	protoPrices := make([]*pb.DailyPrice, len(prices))
	for i, p := range prices {
		protoPrices[i] = &pb.DailyPrice{
			Id:        p.ID,
			ProductId: p.ProductID,
			GradeId:   p.GradeID,
			Price:     p.Price,
			Date:      p.Date.Format("2006-01-02"),
			Time:      p.Time.Format("15:04:05"),
		}
	}
	return &pb.GetTodaysPriceResponse{
		DailyPrices: protoPrices,
	}, nil
}

func (server *GrpcServer) GetTodaysByProductId(ctx context.Context, request *pb.GetTodaysByProductIdRequest) (*pb.GetTodaysByProductIdResponse, error) {
	if err := server.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	date, _ := time.Parse("2006-01-02", request.Date)
	prices, err := server.accountService.GetTodaysByProductId(ctx, request.ProductId, date)
	if err != nil {
		return nil, err
	}
	protoPrices := make([]*pb.DailyPrice, len(prices))
	for i, p := range prices {
		protoPrices[i] = &pb.DailyPrice{
			Id:        p.ID,
			ProductId: p.ProductID,
			GradeId:   p.GradeID,
			Price:     p.Price,
			Date:      p.Date.Format("2006-01-02"),
			Time:      p.Time.Format("15:04:05"),
		}
	}
	return &pb.GetTodaysByProductIdResponse{
		DailyPrices: protoPrices,
	}, nil
}

func (s *GrpcServer) GetProductsWithGradesAndPrices(ctx context.Context, req *pb.GetProductsWithGradesAndPricesRequest) (*pb.GetProductsWithGradesAndPricesResponse, error) {
	if err := s.checkAuthenticated(ctx); err != nil {
		return nil, err
	}
	dateStr := req.Date
	var date time.Time
	var err error
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid date format: %v", err)
		}
	}

	products, err := s.accountService.GetProductsWithGradesAndPrices(ctx, date, req.Search)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get products: %v", err)
	}

	pbProducts := make([]*pb.ProductWithGrades, len(products))
	for i, p := range products {
		pbGrades := make([]*pb.GradeWithPrice, len(p.Grades))
		for j, g := range p.Grades {
			pbGrades[j] = &pb.GradeWithPrice{
				Id:          g.ID,
				ProductId:   g.ProductID,
				Name:        g.Name,
				Description: g.Description,
				Status:      g.Status,
				Price:       g.Price,
			}
		}
		pbProducts[i] = &pb.ProductWithGrades{
			Id:          p.ID,
			Name:        p.Name,
			Category:    p.Category,
			Description: p.Description,
			Status:      p.Status,
			Grades:      pbGrades,
		}
	}

	return &pb.GetProductsWithGradesAndPricesResponse{
		Products: pbProducts,
	}, nil
}
