package main

import (
	"ecommerce/internal/discovery"
	"ecommerce/internal/interceptor"
	"ecommerce/internal/tlsutil"
	"log"
	"net"

	discoveryv1 "ecommerce/gen/discovery/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50054")
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
		grpc.ChainStreamInterceptor(
			interceptor.LoggingStreamInterceptor,
			interceptor.RecoveryStreamInterceptor,
		),
	)

	store := discovery.NewStore()
	discoveryv1.RegisterDiscoveryServiceServer(
		grpcServer,
		discovery.NewServer(store),
	)

	// for postman
	reflection.Register(grpcServer)

	log.Println("50054")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
