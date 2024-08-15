package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			panic(err)
		}
		s := grpc.NewServer()
		proofServer := NewProofServer(17000, "")
		// Register your gRPC service here
		RegisterProofServiceServer(s, proofServer)
		fmt.Println("gRPC server listening on :50051")
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	// Start HTTP server (gRPC-Gateway)
	go func() {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		mux := runtime.NewServeMux()
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

		// Register your gRPC-Gateway handler here
		err := RegisterProofServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
		if err != nil {
			panic(err)
		}

		fmt.Println("gRPC-Gateway server listening on :8080")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			panic(err)
		}
	}()

	// Keep the main goroutine alive
	select {}
}
