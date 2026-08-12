package main

import (
	"ecommerce/internal/interceptor"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	userv1 "ecommerce/gen/user/v1"
	"ecommerce/internal/user"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.LoggingUnaryInterceptor,
			interceptor.RecoveryUnaryInterceptor,
		),
	)

	store := user.NewStore()
	userv1.RegisterUserServiceServer(grpcServer, user.NewServer(store))

	// for postman
	reflection.Register(grpcServer)

	log.Println("50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
