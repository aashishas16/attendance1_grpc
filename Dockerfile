# -------------------- Builder --------------------
    FROM golang:1.25-alpine AS builder

    WORKDIR /app
    
    # Install git for Go modules
    RUN apk add --no-cache git
    
    # Copy go.mod and go.sum and download dependencies
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the rest of the project
    COPY . .
    
    # Build the Go binary for Linux
    RUN CGO_ENABLED=0 GOOS=linux go build -o /attendance1 .
    
    # -------------------- Final Image --------------------
    FROM alpine:3.18
    
    # Install ca-certificates and timezone data
    RUN apk add --no-cache ca-certificates tzdata
    
    WORKDIR /app
    
    # Copy the built binary
    COPY --from=builder /attendance1 /attendance1
    
    # Set timezone
    ENV TZ=Asia/Kolkata
    
    # Expose gRPC and REST Gateway ports
    EXPOSE 50052 8080
    
    # Start the application
    CMD ["/attendance1"]
    