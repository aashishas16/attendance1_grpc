package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	pb "attendance1/proto"
)

func main() {
	log.Println("Starting Attendance Service (gRPC + REST)")

	// MongoDB
	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Mongo connect error:", err)
	}
	log.Println("MongoDB connected successfully")

	collection := client.Database("attendance_db").Collection("records")
	loc, _ := time.LoadLocation("Asia/Kolkata")

	// gRPC Server
	grpcPort := getEnv("GRPC_PORT", "50052")
	grpcServer := grpc.NewServer()
	s := &attendanceServer{collection: collection, loc: loc}
	pb.RegisterAttendanceServiceServer(grpcServer, s)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Println("gRPC server running on port", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// REST Gateway
	httpPort := getEnv("HTTP_PORT", "8080")
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = pb.RegisterAttendanceServiceHandlerFromEndpoint(context.Background(), mux, "localhost:"+grpcPort, opts)
	if err != nil {
		log.Fatalf("Failed to start HTTP gateway: %v", err)
	}
	log.Println("REST gateway running on port", httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, mux))
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
