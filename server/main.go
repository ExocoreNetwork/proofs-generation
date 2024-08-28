package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ChainID     uint64 `yaml:"chain_id"`
	Provider    string `yaml:"provider"`
	GRPCPort    string `yaml:"grpc_port"`
	JSONRPCPort string `yaml:"json_rpc_port"`
}

func main() {
	config := loadConfig("config.yaml")

	// Start gRPC server
	go startGRPCServer(config)

	// Start HTTP server (gRPC-Gateway)
	go startHTTPServer(config)

	// Keep the main goroutine alive
	select {}
}

func loadConfig(filename string) Config {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(fmt.Sprintf("Error parsing config file: %v", err))
	}

	return config
}

func startGRPCServer(config Config) {
	lis, err := net.Listen("tcp", ":"+config.GRPCPort)
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	s := grpc.NewServer()
	proofServer := NewProofServer(config.ChainID, config.Provider)
	RegisterProofServiceServer(s, proofServer)
	fmt.Printf("gRPC server listening on :%s\n", config.GRPCPort)
	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}

func startHTTPServer(config Config) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := RegisterProofServiceHandlerFromEndpoint(ctx, mux, "localhost:"+config.GRPCPort, opts)
	if err != nil {
		panic(err)
	}

	fmt.Printf("gRPC-Gateway server listening on :%s\n", config.JSONRPCPort)
	if err := http.ListenAndServe(":"+config.JSONRPCPort, mux); err != nil {
		panic(err)
	}
}
