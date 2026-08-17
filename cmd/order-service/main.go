package main

import (
	orderv1 "ecommerce/gen/order/v1"
	productv1 "ecommerce/gen/product/v1"
	userv1 "ecommerce/gen/user/v1"
	"ecommerce/internal/interceptor"
	"ecommerce/internal/order"
	"ecommerce/internal/tlsutil"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	userCreds, err := tlsutil.LoadClientTLSCredentials(
		"certs/client.crt", "certs/client.key", "certs/ca.crt",
	)
	if err != nil {
		log.Fatal(err)
	}
	userConn, err := grpc.NewClient(getEnv("USER_SERVICE_ADDR", "127.0.0.1:50051"), grpc.WithTransportCredentials(userCreds))
	if err != nil {
		log.Fatal(err)
	}
	defer userConn.Close()

	productCreds, err := tlsutil.LoadClientTLSCredentials(
		"certs/client.crt",
		"certs/client.key",
		"certs/ca.crt",
	)
	if err != nil {
		log.Fatal(err)
	}

	productConn, err := grpc.NewClient(getEnv("PRODUCT_SERVICE_ADDR", "127.0.0.1:50052"), grpc.WithTransportCredentials(productCreds))
	if err != nil {
		log.Fatal(err)
	}
	defer productConn.Close()

	userClient := userv1.NewUserServiceClient(userConn)
	productClient := productv1.NewProductServiceClient(productConn)

	lis, err := net.Listen("tcp", ":"+getEnv("PORT", "50053"))
	if err != nil {
		log.Fatal(err)
	}

	serverCreds, err := tlsutil.LoadServerTLSCredentials("certs/server.crt", "certs/server.key", "certs/ca.crt")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.ChainUnaryInterceptor(
			interceptor.LoggingUnaryInterceptor,
			interceptor.RecoveryUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			interceptor.LoggingStreamInterceptor,
			interceptor.RecoveryStreamInterceptor,
		),
	)
	store := order.NewStore()
	orderv1.RegisterOrderServiceServer(grpcServer, order.NewServer(store, userClient, productClient))
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	log.Println(getEnv("PORT", "50053"))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
