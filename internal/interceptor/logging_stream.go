package interceptor

import (
	"log"
	"time"

	"google.golang.org/grpc"
)

func LoggingStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()

	err := handler(srv, ss)

	duration := time.Since(start)
	if err != nil {
		log.Printf("method=%s duration=%s error=true err=%s", info.FullMethod, duration, err.Error())
	} else {
		log.Printf("method=%s duration=%s error=false", info.FullMethod, duration)
	}

	return err
}
