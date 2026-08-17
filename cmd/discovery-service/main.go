package main

import (
	"context"
	discoveryv1 "ecommerce/gen/discovery/v1"
	productv1 "ecommerce/gen/product/v1"
	"ecommerce/internal/interceptor"
	"ecommerce/internal/product"
	"ecommerce/internal/registry"
	"ecommerce/internal/tlsutil"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	discoveryCreds, err := tlsutil.LoadClientTLSCredentials(
		"certs/client.crt",
		"certs/client.key",
		"certs/ca.crt",
	)
	if err != nil {
		log.Fatal(err)
	}

	discoveryConn, err := grpc.NewClient(
		"127.0.0.1:50054",
		grpc.WithTransportCredentials(discoveryCreds),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer discoveryConn.Close()

	discoveryClient := discoveryv1.NewDiscoveryServiceClient(discoveryConn)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal(err)
	}

	serverCreds, err := tlsutil.LoadServerTLSCredentials(
		"certs/server.crt",
		"certs/server.key",
		"certs/ca.crt",
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
		grpc.ChainStreamInterceptor(
			interceptor.LoggingStreamInterceptor,
			interceptor.RecoveryStreamInterceptor,
		),
	)

	store := product.NewStore()
	productv1.RegisterProductServiceServer(
		grpcServer,
		product.NewServer(store),
	)

	// for postman
	reflection.Register(grpcServer)

	go registry.RegisterAndHeartbeat(
		context.Background(),
		discoveryClient,
		"product-service",
		"127.0.0.1:50052",
	)

	log.Println("50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
