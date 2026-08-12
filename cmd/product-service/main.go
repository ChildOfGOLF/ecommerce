package main

import (
	"ecommerce/internal/interceptor"
	"log"
	"net"

	productv1 "ecommerce/gen/product/v1"
	"ecommerce/internal/product"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(
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
	productv1.RegisterProductServiceServer(grpcServer, product.NewServer(store))

	// for postman
	reflection.Register(grpcServer)

	log.Println("50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
