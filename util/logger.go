package util

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type Logger interface {
	Transport() *zerolog.Logger
	Database() *zerolog.Logger
	Service() *zerolog.Logger
}

type zerologLogger struct {
	logger zerolog.Logger
}

func NewLogger(level string) Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	return &zerologLogger{logger: log.Logger}
}

func (l *zerologLogger) Transport() *zerolog.Logger {
	sub := l.logger.With().Str("layer", "transport").Logger()
	return &sub
}

func (l *zerologLogger) Database() *zerolog.Logger {
	sub := l.logger.With().Str("layer", "database").Logger()
	return &sub
}

func (l *zerologLogger) Service() *zerolog.Logger {
	sub := l.logger.With().Str("layer", "service").Logger()
	return &sub
}

// UnaryServerInterceptor returns a new unary server interceptor that logs gRPC requests
func UnaryServerInterceptor(logger Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		st, _ := status.FromError(err)

		var logEvent *zerolog.Event
		if err != nil {
			logEvent = logger.Transport().Error().Err(err)
		} else {
			logEvent = logger.Transport().Info()
		}

		logEvent.
			Str("method", info.FullMethod).
			Str("duration", duration.String()).
			Str("code", st.Code().String()).
			Msg("gRPC Request")

		return resp, err
	}
}
