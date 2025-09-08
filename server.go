package main

import (
	"context"
	"log"
	"time"

	pb "attendance1/proto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mongo Model
type AttendanceRecord struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserID       string             `bson:"user_id"`
	Username     string             `bson:"username"`
	CheckinTime  time.Time          `bson:"checkin_time"`
	CheckoutTime *time.Time         `bson:"checkout_time,omitempty"`
}

// gRPC server struct
type attendanceServer struct {
	pb.UnimplementedAttendanceServiceServer
	collection *mongo.Collection
	loc        *time.Location
}

// Format IST
func formatIST(t time.Time, loc *time.Location) string {
	return t.In(loc).Format("2006-01-02 15:04:05 MST")
}

// --- gRPC Methods ---
func (s *attendanceServer) CheckIn(ctx context.Context, req *pb.CheckInRequest) (*pb.AttendanceRecordResponse, error) {
	log.Println("[CheckIn]", req)
	if req.GetUserId() == "" || req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and username required")
	}

	rec := AttendanceRecord{
		ID:          primitive.NewObjectID(),
		UserID:      req.GetUserId(),
		Username:    req.GetUsername(),
		CheckinTime: time.Now().UTC(),
	}

	_, err := s.collection.InsertOne(ctx, rec)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "insert error: %v", err)
	}

	return &pb.AttendanceRecordResponse{
		Id:            rec.ID.Hex(),
		UserId:        rec.UserID,
		Username:      rec.Username,
		CheckinTime:   formatIST(rec.CheckinTime, s.loc),
		StatusMessage: "User checked in successfully",
	}, nil
}

func (s *attendanceServer) CheckOut(ctx context.Context, req *pb.CheckOutRequest) (*pb.AttendanceRecordResponse, error) {
	log.Println("[CheckOut]", req)
	if req.GetRecordId() == "" {
		return nil, status.Error(codes.InvalidArgument, "record_id required")
	}

	oid, err := primitive.ObjectIDFromHex(req.GetRecordId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid record_id")
	}

	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{"checkout_time": now}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated AttendanceRecord
	err = s.collection.FindOneAndUpdate(ctx, bson.M{"_id": oid}, update, opts).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "record not found")
		}
		return nil, status.Errorf(codes.Internal, "update error: %v", err)
	}

	checkoutStr := ""
	if updated.CheckoutTime != nil {
		checkoutStr = formatIST(*updated.CheckoutTime, s.loc)
	}

	return &pb.AttendanceRecordResponse{
		Id:            updated.ID.Hex(),
		UserId:        updated.UserID,
		Username:      updated.Username,
		CheckinTime:   formatIST(updated.CheckinTime, s.loc),
		CheckoutTime:  checkoutStr,
		StatusMessage: "User checked out successfully",
	}, nil
}

func (s *attendanceServer) GetAttendance(ctx context.Context, req *pb.GetAttendanceRequest) (*pb.AttendanceRecordResponse, error) {
	log.Println("[GetAttendance]", req)
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}

	filter := bson.M{"user_id": req.GetUserId()}
	opts := options.FindOne().SetSort(bson.D{{Key: "checkin_time", Value: -1}})
	var r AttendanceRecord
	if err := s.collection.FindOne(ctx, filter, opts).Decode(&r); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "no records found")
		}
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	checkoutStr := ""
	if r.CheckoutTime != nil {
		checkoutStr = formatIST(*r.CheckoutTime, s.loc)
	}

	return &pb.AttendanceRecordResponse{
		Id:            r.ID.Hex(),
		UserId:        r.UserID,
		Username:      r.Username,
		CheckinTime:   formatIST(r.CheckinTime, s.loc),
		CheckoutTime:  checkoutStr,
		StatusMessage: "Record found",
	}, nil
}

func (s *attendanceServer) GetAllAttendance(ctx context.Context, req *pb.GetAllAttendanceRequest) (*pb.GetAllAttendanceResponse, error) {
	log.Println("[GetAllAttendance] request received")
	cursor, err := s.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "find error: %v", err)
	}
	defer cursor.Close(ctx)

	var records []*pb.AttendanceRecordResponse
	for cursor.Next(ctx) {
		var r AttendanceRecord
		if err := cursor.Decode(&r); err != nil {
			continue
		}
		checkoutStr := ""
		if r.CheckoutTime != nil {
			checkoutStr = formatIST(*r.CheckoutTime, s.loc)
		}
		records = append(records, &pb.AttendanceRecordResponse{
			Id:            r.ID.Hex(),
			UserId:        r.UserID,
			Username:      r.Username,
			CheckinTime:   formatIST(r.CheckinTime, s.loc),
			CheckoutTime:  checkoutStr,
			StatusMessage: "Record retrieved",
		})
	}

	return &pb.GetAllAttendanceResponse{Records: records}, nil
}
