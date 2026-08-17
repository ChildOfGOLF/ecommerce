package main

import (
	"ecommerce/internal/interceptor"
	"ecommerce/internal/tlsutil"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	userv1 "ecommerce/gen/user/v1"
	"ecommerce/internal/user"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	lis, err := net.Listen("tcp", ":"+getEnv("PORT", "50051"))
	if err != nil {
		log.Fatal(err)
	}

	serverCreds, err := tlsutil.LoadServerTLSCredentials(
		"certs/server.crt", "certs/server.key", "certs/ca.crt",
	)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.ChainUnaryInterceptor(
			interceptor.LoggingUnaryInterceptor,
			interceptor.RecoveryUnaryInterceptor,
		),
	)

	store := user.NewStore()
	userv1.RegisterUserServiceServer(grpcServer, user.NewServer(store))

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	log.Println(getEnv("PORT", "50051"))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
