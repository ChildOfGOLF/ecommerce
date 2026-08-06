package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userv1 "ecommerce/gen/user/v1"
)

func main() {
	conn, err := grpc.NewClient(
		"127.0.0.1:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := userv1.NewUserServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	created, err := client.CreateUser(ctx, &userv1.CreateUserRequest{
		Email:    "don@corleo.ne",
		Password: "password123",
		FullName: "Vito Corleone",
	})
	if err != nil {
		log.Fatal(err)
	}

	user := created.GetUser()

	log.Printf(
		"created: id=%s email=%s name=%s",
		user.GetUserId(),
		user.GetEmail(),
		user.GetFullName(),
	)

	got, err := client.GetUser(ctx, &userv1.GetUserRequest{
		UserId: user.GetUserId(),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(
		"fetched: id=%s email=%s name=%s",
		got.GetUser().GetUserId(),
		got.GetUser().GetEmail(),
		got.GetUser().GetFullName(),
	)

	list, err := client.ListUsers(ctx, &userv1.ListUsersRequest{})
	if err != nil {
		log.Fatal(err)
	}

	log.Print(len(list.GetUsers()))

	for _, u := range list.GetUsers() {
		log.Print(
			u.GetUserId(),
			u.GetFullName(),
			u.GetEmail(),
		)
	}
}
