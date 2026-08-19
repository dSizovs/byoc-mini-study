package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
	"os"

	pb "byoc_server/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type server struct {
	pb.UnimplementedControlPlaneServer
	taskGiven bool
}

// Worker calls FetchTask over gRPC
func (s *server) FetchTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	log.Printf("Worker connected: %s", req.WorkerId)

	// If task hasn't been assigned yet, issue the string-transformation command
	if !s.taskGiven {
		s.taskGiven = true
		log.Println("Dispatching instruction to worker...")
		return &pb.TaskResponse{
			HasTask: true,
			TaskId:  "job-001",
			// The logic: Run Python inside container to convert input.txt to UPPERCASE
			Command: `python3 -c "with open('/data/input.txt', 'r+') as f: text = f.read().strip().upper(); f.seek(0); f.write(text); f.truncate()"`,
		}, nil
	}

	return &pb.TaskResponse{HasTask: false}, nil
}

// Worker calls SubmitResult over gRPC
func (s *server) SubmitResult(ctx context.Context, req *pb.ResultRequest) (*pb.ResultResponse, error) {
	log.Printf("RESULT RECEIVED from %s for task %s:", req.WorkerId, req.TaskId)
	log.Printf("   Success: %v | Message: %s", req.Success, req.StatusMessage)
	log.Println("Notice: The Control Plane NEVER saw the contents of input.txt!")
	return &pb.ResultResponse{Acknowledged: true}, nil
}

func main() {
	// Load CA certificate
	caCert, err := os.ReadFile("/home/ubuntu/byoc_certs/ca.crt")
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)

	// Load Server Certificate and Key
	serverCert, err := tls.LoadX509KeyPair("/home/ubuntu/byoc_certs/server.crt", "/home/ubuntu/byoc_certs/server.key")
	if err != nil {
		log.Fatalf("Failed to load server keypair: %v", err)
	}

	// Create mTLS configuration (Require and verify client certificate)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen on 50051: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	pb.RegisterControlPlaneServer(grpcServer, &server{})

	log.Println("Control Plane gRPC server listening on 0.0.0.0:50051 (mTLS enabled)...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
