# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Step 1: Copy ONLY go.mod and go.sum
COPY go.mod go.sum ./

# Step 2: Download dependencies (this layer will be cached if mod/sum haven't changed)
RUN go mod download

# Step 3: Copy ALL code (after dependencies)
COPY . .

# Step 4: Build binaries
RUN go build -o main cmd/main/main.go
RUN go build -o worker cmd/worker/main.go

# Stage 2: Minimal image
FROM alpine:latest

RUN apk update && apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binaries and .env
COPY --from=builder /app/main ./main
COPY --from=builder /app/worker ./worker
COPY --from=builder /app/.env .env

RUN chmod +x ./main ./worker

EXPOSE 8080 50051 8081 8082 8083

# Command will be specified in docker-compose