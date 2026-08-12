package interceptor

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

func LoggingUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	if err != nil {
		log.Printf("method=%s duration=%s error=true err=%s", info.FullMethod, duration, err.Error())
	} else {
		log.Printf("method=%s duration=%s error=false", info.FullMethod, duration)
	}

	return resp, err
}
