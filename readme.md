

* What the project does
* How gRPC + gRPC-Gateway + Protobuf fit together
* How to install/run with **Docker**, **Kubernetes (Minikube)**, and **Helm**

Here’s a solid starting point:

---

```markdown
# Attendance Management Service

## 📌 Overview
The **Attendance Management Service** is a microservice that manages employee check-in/check-out data.  
It is built using:
- **Go (Golang)** for backend service
- **Protocol Buffers (Protobuf)** for API definitions
- **gRPC** for high-performance internal service communication
- **gRPC-Gateway** for REST/HTTP to gRPC translation
- **MongoDB** as the database
- **Docker** for containerization
- **Kubernetes (Minikube)** for deployment
- **Helm** for Kubernetes packaging and management

This service can be part of a larger ecosystem (e.g., Employee Data Service, Payroll Service) where services communicate via gRPC.

---

## 🏗️ Project Architecture

```

+---------------------+       +----------------------+
\|   REST Client       | --->  | gRPC-Gateway (REST→gRPC) |
+---------------------+       +----------------------+
|
v
+---------------------+
\| Attendance Service  |
\|   (gRPC Server)     |
+---------------------+
|
v
+---------------------+
\|     MongoDB DB      |
+---------------------+

```

- Clients can use **REST (HTTP/JSON)** or **gRPC**.  
- **gRPC-Gateway** automatically translates REST requests into gRPC.  
- Data is stored in **MongoDB**.

---

## ⚙️ Requirements

- [Go](https://go.dev/) (>=1.20)
- [Protocol Buffers Compiler (`protoc`)](https://grpc.io/docs/protoc-installation/)
- [Docker](https://www.docker.com/products/docker-desktop/)
- [Minikube](https://minikube.sigs.k8s.io/docs/)
- [Helm](https://helm.sh/docs/intro/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

---

## 🚀 Setup & Installation

### 1. Generate gRPC and Gateway Code
```bash
protoc -I proto -I proto/googleapis \
  proto/attendance.proto \
  --go_out=. --go-grpc_out=. \
  --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative
````

This generates:

* `attendance.pb.go` (messages)
* `attendance_grpc.pb.go` (gRPC server + client)
* `attendance.pb.gw.go` (REST handlers via gRPC-Gateway)

---

### 2. Run Locally (Go only)

```bash
go mod tidy
go run main.go
```

Test REST endpoint:

```bash
curl -X POST http://localhost:8080/v1/checkin \
  -H "Content-Type: application/json" \
  -d '{"user_id": "emp01", "username": "Nemo"}'
```

Test gRPC directly:

```bash
grpcurl -plaintext -d '{"user_id": "emp01"}' \
  localhost:50051 attendance.AttendanceService/GetAttendance
```

---

### 3. Run with Docker

Build Docker image:

```bash
docker build -t attendance-service:latest .
```

Run container:

```bash
docker run -d -p 50051:50051 -p 8080:8080 attendance-service:latest
```

---

### 4. Run with Kubernetes (Minikube)

Start Minikube:

```bash
minikube start
```

Deploy MongoDB Secret:

```bash
kubectl apply -f k8s/mongo-secret.yaml
```

Deploy Attendance Service + MongoDB:

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

Check pods:

```bash
kubectl get pods
```

Port-forward service:

```bash
kubectl port-forward svc/attendance-service 8080:8080 50051:50051
```

---

### 5. Run with Helm

From project root:

```bash
helm install attendance charts/
```

Upgrade:

```bash
helm upgrade attendance charts/
```

Uninstall:

```bash
helm uninstall attendance
```

---

## 📡 API Endpoints

### gRPC

* `CheckIn(CheckInRequest) returns (CheckInResponse)`
* `CheckOut(CheckOutRequest) returns (CheckOutResponse)`
* `GetAttendance(GetAttendanceRequest) returns (GetAttendanceResponse)`

### REST (via gRPC-Gateway)

* `POST /v1/checkin`
* `POST /v1/checkout`
* `GET /v1/attendance/{user_id}`

---

## 🛠️ Notes

* gRPC is **faster** and strongly typed; REST support is for external clients.
* `google/api/annotations.proto` is needed for gRPC-Gateway. Clone [googleapis](https://github.com/googleapis/googleapis) into `proto/googleapis/`.
* Use `minikube service attendance-service --url` to get service URL in Kubernetes.

---

## 👨‍💻 Author

Built by **Aashish Singune** as part of microservices learning project.

```

---
