package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	helloworldpb "github.com/napat/gprc-gateway/proto/helloworld"
	pingpongpb "github.com/napat/gprc-gateway/proto/pingpong"
)

type server struct {
	helloworldpb.UnimplementedGreeterServer
	pingpongpb.UnimplementedPingPongServiceServer
}

func NewServer() *server {
	return &server{}
}

func (s *server) SayHello(ctx context.Context, in *helloworldpb.HelloRequest) (*helloworldpb.HelloReply, error) {
	respMessage := fmt.Sprintf("%s %s", in.Name, " world")
	return &helloworldpb.HelloReply{Message: respMessage}, nil
}

func (s *server) Pingpong(context.Context, *pingpongpb.Ping) (*pingpongpb.Pong, error) {
	return &pingpongpb.Pong{Result: "ok"}, nil
}

func runGrpcServer() (*grpc.Server, net.Listener, error) {
	var err error

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
		return nil, nil, err
	}

	enableTLS := false
	opts := []grpc.ServerOption{} // default serverOptions is skip SSL
	if enableTLS {
		certFile := "../cert/server-crt.pem"
		keyFile := "../cert/server-key-pkcs8.pem"
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			log.Fatalf("Failed loading certificates: %v\n", err)
			return nil, nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Create a gRPC server object
	s := grpc.NewServer(opts...)

	// Attach the services to the server
	helloworldpb.RegisterGreeterServer(s, &server{})
	pingpongpb.RegisterPingPongServiceServer(s, &server{})

	// Register reflection service on gRPC server.
	// reflection.Register(s)

	// Serve gRPC server
	log.Println("Serving gRPC on 0.0.0.0:8080")
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to listen %v\n", err)
		}
	}()

	return s, lis, nil
}

// runGrpcGatewayProxy REST Reverse Proxy Gateway plugin for gRPC Server
func runGrpcGatewayProxy() (*http.Server, error) {
	// --------------------------------
	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests

	enableTLS := false
	opts := grpc.WithInsecure() // Default dial with grpc.WithInsecure() option to skip SSL certificate init
	if enableTLS {
		caCertFile := "../cert/ca-crt.pem" // Certicate Authority Trust certificate
		creds, err := credentials.NewClientTLSFromFile(caCertFile, "")
		if err != nil {
			log.Fatalf("Failed loading trust certificate: %v\n", err)
			return nil, err
		}

		opts = grpc.WithTransportCredentials(creds)
	}

	conn, err := grpc.DialContext(
		context.Background(),
		"0.0.0.0:8080",
		grpc.WithBlock(),
		grpc.WithInsecure(),
		opts,
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
		return nil, err
	}

	gwmux := runtime.NewServeMux()

	// helloworldpb: Register Greeter
	err = helloworldpb.RegisterGreeterHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
		return nil, err
	}

	// pingpongpb: Register PingPong
	err = pingpongpb.RegisterPingPongServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
		return nil, err
	}

	gwServer := &http.Server{
		Addr:    ":8090",
		Handler: gwmux,
	}

	// log.Fatalln(gwServer.ListenAndServe())
	go func() {
		if err := gwServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8090")

	return gwServer, nil
}

func main() {

	grpcServer, grpcLis, err := runGrpcServer()
	if err != nil {
		log.Fatalf("runGrpcServer: %s\n", err)
	}

	gwServer, err := runGrpcGatewayProxy()
	if err != nil {
		log.Fatalf("runGrpcGatewayProxy: %s\n", err)
	}

	// Graceful Shudown: Wait for ctrl+c to exit
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM) // SIGINT --> CTRL+C

	select {
	case sig := <-sigs:
		log.Printf("Server stop via CTRL+C (signal: %v)\n", sig)
	}

	log.Println("Stop gateway proxy")
	// Using context timeout threshold to prevent Shutdown()'s zombie issues
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// releases resources if slowOperation completes before timeout elapses
		cancel()
		// extra handling here: Close database, redis, truncate message queues, etc.
		// ...
	}()

	if err := gwServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	log.Println("Stop gRPC Server")
	grpcServer.Stop()

	log.Println("Closing the gRPC Listenner")
	grpcLis.Close()

	log.Println("End Of Program")
	os.Exit(0)
}
