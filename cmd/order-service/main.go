package main

import (
	orderv1 "ecommerce/gen/order/v1"
	productv1 "ecommerce/gen/product/v1"
	userv1 "ecommerce/gen/user/v1"
	"ecommerce/internal/interceptor"
	"ecommerce/internal/order"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	userConn, err := grpc.NewClient("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer userConn.Close()

	productConn, err := grpc.NewClient("127.0.0.1:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer productConn.Close()

	userClient := userv1.NewUserServiceClient(userConn)
	productClient := productv1.NewProductServiceClient(productConn)

	lis, err := net.Listen("tcp", ":50053")
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
	store := order.NewStore()
	orderv1.RegisterOrderServiceServer(grpcServer, order.NewServer(store, userClient, productClient))
	reflection.Register(grpcServer)

	log.Println("50053")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
