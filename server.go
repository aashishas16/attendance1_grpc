package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
)

// Attendance record structure
type AttendanceRecord struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       string             `bson:"user_id" json:"user_id"`
	Username     string             `bson:"username" json:"username"`
	CheckinTime  time.Time          `bson:"checkin_time" json:"checkin_time"`
	CheckoutTime *time.Time         `bson:"checkout_time,omitempty" json:"checkout_time,omitempty"`
}

// Server holds Mongo collection
type server struct {
	collection *mongo.Collection
}

// Utility: format time in IST
func formatIST(t time.Time) string {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	return t.In(loc).Format("2006-01-02 15:04:05 MST")
}

// --------- Handlers ---------

func (s *server) checkIn(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UserID == "" || body.Username == "" {
		http.Error(w, "user_id and username required", http.StatusBadRequest)
		return
	}

	rec := AttendanceRecord{
		ID:          primitive.NewObjectID(),
		UserID:      body.UserID,
		Username:    body.Username,
		CheckinTime: time.Now().UTC(),
	}

	if _, err := s.collection.InsertOne(r.Context(), rec); err != nil {
		http.Error(w, "failed to insert record", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Checked in successfully",
		"id":      rec.ID.Hex(),
	})
}

func (s *server) checkOut(w http.ResponseWriter, r *http.Request) {
	recID := mux.Vars(r)["record_id"]
	if recID == "" {
		http.Error(w, "record_id required", http.StatusBadRequest)
		return
	}

	oid, _ := primitive.ObjectIDFromHex(recID)
	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{"checkout_time": now}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated AttendanceRecord
	err := s.collection.FindOneAndUpdate(r.Context(), bson.M{"_id": oid}, update, opts).Decode(&updated)
	if err != nil {
		http.Error(w, "record not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message":      "Checked out successfully",
		"id":           updated.ID.Hex(),
		"checkoutTime": formatIST(*updated.CheckoutTime),
	})
}

func (s *server) getAttendance(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["user_id"]
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	var record AttendanceRecord
	err := s.collection.FindOne(r.Context(), bson.M{"user_id": userID}).Decode(&record)
	if err != nil {
		http.Error(w, "no record found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(record)
}

func (s *server) getAllAttendance(w http.ResponseWriter, r *http.Request) {
	cursor, err := s.collection.Find(r.Context(), bson.D{})
	if err != nil {
		http.Error(w, "failed to fetch records", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(r.Context())

	var records []AttendanceRecord
	if err := cursor.All(r.Context(), &records); err != nil {
		http.Error(w, "failed to decode records", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(records)
}

// --------- Main ---------

func main() {
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

// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"log"
// 	"net"
// 	"net/http"
// 	"os"
// 	"time"

// 	pb "attendance1/proto"

// 	"github.com/gorilla/mux"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"

// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/reflection"
// 	"google.golang.org/grpc/status"
// )

// type AttendanceRecord struct {
// 	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
// 	UserID       string             `bson:"user_id" json:"user_id"`
// 	Username     string             `bson:"username" json:"username"`
// 	CheckinTime  time.Time          `bson:"checkin_time" json:"checkin_time"`
// 	CheckoutTime *time.Time         `bson:"checkout_time,omitempty" json:"checkout_time,omitempty"`
// }

// type server struct {
// 	pb.UnimplementedAttendanceServiceServer
// 	collection *mongo.Collection
// 	loc        *time.Location
// }

// // helper to format time in IST display
// func formatIST(t time.Time, loc *time.Location) string {
// 	return t.In(loc).Format("2006-01-02 15:04:05 MST")
// }

// // ---------------- gRPC methods ----------------

// func (s *server) CheckIn(ctx context.Context, req *pb.CheckInRequest) (*pb.AttendanceRecordResponse, error) {
// 	log.Printf("CheckIn called: user_id=%s username=%s", req.GetUserId(), req.GetUsername())
// 	if req == nil {
// 		return nil, status.Error(codes.InvalidArgument, "nil request")
// 	}
// 	if req.GetUserId() == "" || req.GetUsername() == "" {
// 		return nil, status.Error(codes.InvalidArgument, "user_id and username are required")
// 	}
// 	rec := AttendanceRecord{
// 		ID:          primitive.NewObjectID(),
// 		UserID:      req.GetUserId(),
// 		Username:    req.GetUsername(),
// 		CheckinTime: time.Now().UTC(),
// 	}
// 	_, err := s.collection.InsertOne(ctx, rec)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to insert record: %v", err)
// 	}

// 	log.Println("Inserted record ID:", rec.ID.Hex())
// 	return &pb.AttendanceRecordResponse{
// 		Id:            rec.ID.Hex(),
// 		UserId:        rec.UserID,
// 		Username:      rec.Username,
// 		CheckinTime:   formatIST(rec.CheckinTime, s.loc),
// 		StatusMessage: "User checked in successfully.",
// 	}, nil
// }

// func (s *server) CheckOut(ctx context.Context, req *pb.CheckOutRequest) (*pb.AttendanceRecordResponse, error) {
// 	if req == nil || req.GetRecordId() == "" {
// 		return nil, status.Error(codes.InvalidArgument, "record_id required")
// 	}

// 	oid, err := primitive.ObjectIDFromHex(req.GetRecordId())
// 	if err != nil {
// 		return nil, status.Error(codes.InvalidArgument, "invalid record id")
// 	}

// 	update := bson.M{"$set": bson.M{"checkout_time": time.Now().UTC()}}
// 	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

// 	var updated AttendanceRecord
// 	err = s.collection.FindOneAndUpdate(ctx, bson.M{"_id": oid}, update, opts).Decode(&updated)
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return nil, status.Error(codes.NotFound, "record not found")
// 		}
// 		return nil, status.Errorf(codes.Internal, "update error: %v", err)
// 	}

// 	var checkoutStr string
// 	if updated.CheckoutTime != nil {
// 		checkoutStr = formatIST(*updated.CheckoutTime, s.loc)
// 	}

// 	return &pb.AttendanceRecordResponse{
// 		Id:            updated.ID.Hex(),
// 		UserId:        updated.UserID,
// 		Username:      updated.Username,
// 		CheckinTime:   formatIST(updated.CheckinTime, s.loc),
// 		CheckoutTime:  checkoutStr,
// 		StatusMessage: "User checked out successfully.",
// 	}, nil
// }

// func (s *server) GetAttendance(ctx context.Context, req *pb.GetAttendanceRequest) (*pb.AttendanceRecordResponse, error) {
// 	if req == nil || req.GetUserId() == "" {
// 		return nil, status.Error(codes.InvalidArgument, "user_id required")
// 	}
// 	filter := bson.M{"user_id": req.GetUserId()}
// 	opts := options.FindOne().SetSort(bson.D{{Key: "checkin_time", Value: -1}})

// 	var r AttendanceRecord
// 	if err := s.collection.FindOne(ctx, filter, opts).Decode(&r); err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return nil, status.Error(codes.NotFound, "no records found for this user")
// 		}
// 		return nil, status.Errorf(codes.Internal, "db error: %v", err)
// 	}

// 	var checkoutStr string
// 	if r.CheckoutTime != nil {
// 		checkoutStr = formatIST(*r.CheckoutTime, s.loc)
// 	}

// 	return &pb.AttendanceRecordResponse{
// 		Id:            r.ID.Hex(),
// 		UserId:        r.UserID,
// 		Username:      r.Username,
// 		CheckinTime:   formatIST(r.CheckinTime, s.loc),
// 		CheckoutTime:  checkoutStr,
// 		StatusMessage: "Record found.",
// 	}, nil
// }

// func (s *server) GetAllAttendance(ctx context.Context, req *pb.GetAllAttendanceRequest) (*pb.GetAllAttendanceResponse, error) {
// 	cursor, err := s.collection.Find(ctx, bson.D{})
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "find error: %v", err)
// 	}
// 	defer cursor.Close(ctx)

// 	out := []*pb.AttendanceRecordResponse{}
// 	for cursor.Next(ctx) {
// 		var r AttendanceRecord
// 		if err := cursor.Decode(&r); err != nil {
// 			log.Printf("decode error: %v", err)
// 			continue
// 		}
// 		var checkoutStr string
// 		if r.CheckoutTime != nil {
// 			checkoutStr = formatIST(*r.CheckoutTime, s.loc)
// 		}
// 		out = append(out, &pb.AttendanceRecordResponse{
// 			Id:            r.ID.Hex(),
// 			UserId:        r.UserID,
// 			Username:      r.Username,
// 			CheckinTime:   formatIST(r.CheckinTime, s.loc),
// 			CheckoutTime:  checkoutStr,
// 			StatusMessage: "Record retrieved.",
// 		})
// 	}
// 	if err := cursor.Err(); err != nil {
// 		return nil, status.Errorf(codes.Internal, "cursor error: %v", err)
// 	}
// 	return &pb.GetAllAttendanceResponse{Records: out}, nil
// }

// // ---------------- HTTP handlers (call the same business logic) ----------------

// func (s *server) httpCheckIn(w http.ResponseWriter, r *http.Request) {
// 	var body struct {
// 		UserID   string `json:"user_id"`
// 		Username string `json:"username"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
// 		http.Error(w, "invalid json body", http.StatusBadRequest)
// 		return
// 	}
// 	resp, err := s.CheckIn(r.Context(), &pb.CheckInRequest{UserId: body.UserID, Username: body.Username})
// 	if err != nil {
// 		http.Error(w, err.Error(), httpgrpcCode(err))
// 		return
// 	}
// 	writeJSON(w, resp)
// }

// func (s *server) httpCheckOut(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	recID := vars["record_id"]
// 	resp, err := s.CheckOut(r.Context(), &pb.CheckOutRequest{RecordId: recID})
// 	if err != nil {
// 		http.Error(w, err.Error(), httpgrpcCode(err))
// 		return
// 	}
// 	writeJSON(w, resp)
// }

// func (s *server) httpGetAttendance(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	userID := vars["user_id"]
// 	resp, err := s.GetAttendance(r.Context(), &pb.GetAttendanceRequest{UserId: userID})
// 	if err != nil {
// 		http.Error(w, err.Error(), httpgrpcCode(err))
// 		return
// 	}
// 	writeJSON(w, resp)
// }

// func (s *server) httpGetAll(w http.ResponseWriter, r *http.Request) {
// 	resp, err := s.GetAllAttendance(r.Context(), &pb.GetAllAttendanceRequest{})
// 	if err != nil {
// 		http.Error(w, err.Error(), httpgrpcCode(err))
// 		return
// 	}
// 	writeJSON(w, resp)
// }

// func writeJSON(w http.ResponseWriter, v interface{}) {
// 	w.Header().Set("Content-Type", "application/json")
// 	_ = json.NewEncoder(w).Encode(v)
// }

// // map gRPC status -> HTTP code (basic)
// func httpgrpcCode(err error) int {
// 	if err == nil {
// 		return http.StatusOK
// 	}
// 	st, ok := status.FromError(err)
// 	if !ok {
// 		return http.StatusInternalServerError
// 	}
// 	switch st.Code() {
// 	case codes.InvalidArgument:
// 		return http.StatusBadRequest
// 	case codes.NotFound:
// 		return http.StatusNotFound
// 	case codes.Internal, codes.ResourceExhausted:
// 		return http.StatusInternalServerError
// 	default:
// 		return http.StatusInternalServerError
// 	}
// }

// // ---------------- Main ----------------

// func main() {
// 	// Mongo URI
// 	mongoURI := os.Getenv("MONGO_URI")
// 	if mongoURI == "" {
// 		mongoURI = "mongodb://localhost:27017"
// 	}

// 	// Connect to Mongo
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
// 	if err != nil {
// 		log.Fatalf("failed connect mongodb: %v", err)
// 	}
// 	collection := client.Database("attendance_db").Collection("records")

// 	// Load IST
// 	loc, err := time.LoadLocation("Asia/Kolkata")
// 	if err != nil {
// 		log.Printf("warning: failed load tz: %v — using UTC", err)
// 		loc = time.UTC
// 	}

// 	srv := &server{collection: collection, loc: loc}

// 	// Start gRPC server
// 	grpcPort := getEnv("GRPC_PORT", "50051")
// 	lis, err := net.Listen("tcp", ":"+grpcPort)
// 	if err != nil {
// 		log.Fatalf("failed listen: %v", err)
// 	}
// 	grpcServer := grpc.NewServer()
// 	pb.RegisterAttendanceServiceServer(grpcServer, srv)
// 	reflection.Register(grpcServer)

// 	go func() {
// 		log.Printf("gRPC listening on %s", grpcPort)
// 		if err := grpcServer.Serve(lis); err != nil {
// 			log.Fatalf("grpc serve: %v", err)
// 		}
// 	}()

// 	// Start HTTP server
// 	httpPort := getEnv("HTTP_PORT", "8080")
// 	r := mux.NewRouter()
// 	r.HandleFunc("/v1/checkin", srv.httpCheckIn).Methods("POST")
// 	r.HandleFunc("/v1/checkout/{record_id}", srv.httpCheckOut).Methods("PUT")
// 	r.HandleFunc("/v1/attendance/{user_id}", srv.httpGetAttendance).Methods("GET")
// 	r.HandleFunc("/v1/attendance", srv.httpGetAll).Methods("GET")

// 	log.Printf("HTTP API listening on %s", httpPort)
// 	if err := http.ListenAndServe(":"+httpPort, r); err != nil {
// 		log.Fatalf("http serve: %v", err)
// 	}
// }

// func getEnv(k, def string) string {
// 	if v := os.Getenv(k); v != "" {
// 		return v
// 	}
// 	return def
// }
