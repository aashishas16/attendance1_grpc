# builder
FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /attendance1 .

# final
FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /attendance1 /attendance1
ENV TZ=Asia/Kolkata
EXPOSE 50051 8080
CMD ["/attendance1"]

# # builder
# FROM golang:1.25-alpine3.18 AS builder
# WORKDIR /app
# COPY . .
# RUN go build -o attendance1 .

# # final image
# FROM alpine:3.17
# RUN apk update && apk add --no-cache ca-certificates tzdata git
# WORKDIR /app
# COPY --from=builder /app/attendance1 /attendance1
# ENV TZ=Asia/Kolkata
# EXPOSE 50051 8080

# ENTRYPOINT ["/attendance1"]
