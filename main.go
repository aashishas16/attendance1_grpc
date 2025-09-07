package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	log.Println("Starting Attendance API server...")

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Mongo connect error:", err)
	}
	log.Println("MongoDB connected successfully")

	collection := client.Database("attendance_db").Collection("records")
	s := &server{collection: collection}

	r := mux.NewRouter()
	r.HandleFunc("/v1/checkin", s.checkIn).Methods("POST")
	r.HandleFunc("/v1/checkout/{record_id}", s.checkOut).Methods("PUT")
	r.HandleFunc("/v1/attendance/{user_id}", s.getAttendance).Methods("GET")
	r.HandleFunc("/v1/attendance", s.getAllAttendance).Methods("GET")

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	log.Println("HTTP server running on port", httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, r))
}
