package main

import (
	"context"
	orderv1 "ecommerce/gen/order/v1"
	productv1 "ecommerce/gen/product/v1"
	userv1 "ecommerce/gen/user/v1"
	"ecommerce/internal/tlsutil"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	mux := runtime.NewServeMux()

	clientCreds, err := tlsutil.LoadClientTLSCredentials("certs/client.crt", "certs/client.key", "certs/ca.crt")
	if err != nil {
		log.Fatal(err)
	}

	err = productv1.RegisterProductServiceHandlerFromEndpoint(
		ctx, mux, "127.0.0.1:50052", []grpc.DialOption{grpc.WithTransportCredentials(clientCreds)},
	)
	if err != nil {
		log.Fatal(err)
	}

	err = userv1.RegisterUserServiceHandlerFromEndpoint(
		ctx, mux, "127.0.0.1:50051", []grpc.DialOption{grpc.WithTransportCredentials(clientCreds)},
	)
	if err != nil {
		log.Fatal(err)
	}

	err = orderv1.RegisterOrderServiceHandlerFromEndpoint(
		ctx, mux, "127.0.0.1:50053", []grpc.DialOption{grpc.WithTransportCredentials(clientCreds)},
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(":8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
