package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
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

	// Custom Console Writer for Tabular Output
	output := zerolog.ConsoleWriter{
		Out:          os.Stdout,
		TimeFormat:   "2006-01-02 15:04:05",
		PartsExclude: []string{"layer", "method", "request", "response", "duration", "query", "result", "code", "email", "exists"},
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s |", i))
		},
		FormatMessage: func(i interface{}) string {
			if i == "API" || i == "DB" {
				return ""
			}
			return fmt.Sprintf("%-15s |", i)
		},
		FormatCaller: func(i interface{}) string {
			file := fmt.Sprintf("%v", i)
			if idx := strings.LastIndex(file, "/"); idx != -1 {
				file = file[idx+1:]
			}
			return fmt.Sprintf("%-20s |", file)
		},
	}

	// Override the default formatter to create a table-like experience
	output.FormatExtra = func(m map[string]interface{}, b *bytes.Buffer) error {
		layer, _ := m["layer"].(string)
		duration, _ := m["duration"].(string)

		if layer == "transport" || layer == "rest" {
			method, _ := m["method"].(string)
			req, _ := m["request"].(string)
			resp, _ := m["response"].(string)
			fmt.Fprintf(b, " %-30s | %-12s | REQ: %-40s | RESP: %-40s |", method, duration, req, resp)
		} else if layer == "database" {
			query, _ := m["query"].(string)
			result, _ := m["result"].(string)
			fmt.Fprintf(b, " %-35s | %-12s | RESULT: %-40s |", query, duration, result)
		}
		return nil
	}

	logger := zerolog.New(output).With().Timestamp().Caller().Logger()
	return &zerologLogger{logger: logger}
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

		reqJSON, _ := json.Marshal(req)
		respJSON, _ := json.Marshal(resp)

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
			Str("request", string(reqJSON)).
			Str("response", string(respJSON)).
			Msg("API")

		return resp, err
	}
}
