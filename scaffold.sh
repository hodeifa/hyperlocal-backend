#!/bin/bash
# ==============================================================================
# HYPERLOCAL BACKEND - SCAFFOLDING SCRIPT (Sprint 2 Issue #25)
# Men-generate struktur Monorepo Microservices + Clean Architecture + Versioning
# ==============================================================================

set -e

echo "🚀 Memulai scaffolding struktur folder & Go Workspace..."

# 1. Definisi Services
SERVICES=("location" "order" "customer" "driver" "chat" "chat-worker" "map")
PKG_DIRS=("logger" "database" "cache" "storage" "middleware" "grpcclient" "response" "errors" "metrics" "fcm" "vatopup")

# 2. Buat Struktur API Gateway
mkdir -p api-gateway/{cmd/server,config,internal/delivery/http/v1,internal/delivery/http/v2,internal/middleware}
echo "package main" > api-gateway/cmd/server/main.go
echo "package config" > api-gateway/config/config.go
echo "package v1" > api-gateway/internal/delivery/http/v1/doc.go
echo "package v2" > api-gateway/internal/delivery/http/v2/doc.go
echo "package middleware" > api-gateway/internal/middleware/doc.go
cat <<EOF > api-gateway/go.mod
module github.com/hodeifa/hyperlocal-backend/api-gateway
go 1.22
EOF

# 3. Buat Struktur Services (Microservices)
for svc in "${SERVICES[@]}"; do
    mkdir -p services/$svc/{cmd/server,config}
    echo "package main" > services/$svc/cmd/server/main.go
    echo "package config" > services/$svc/config/config.go
    
    cat <<EOF > services/$svc/go.mod
module github.com/hodeifa/hyperlocal-backend/services/$svc
go 1.22
EOF

    # chat-worker adalah background worker, tidak punya delivery layer
    if [ "$svc" == "chat-worker" ]; then
        mkdir -p services/$svc/internal/{usecase/v1,repository}
        echo "package v1" > services/$svc/internal/usecase/v1/doc.go
        echo "package repository" > services/$svc/internal/repository/doc.go
    else
        # Delivery Layer (Versioned)
        if [ "$svc" == "location" ]; then
            mkdir -p services/$svc/internal/delivery/{grpc/v1,ws/v1}
            echo "package v1" > services/$svc/internal/delivery/grpc/v1/doc.go
            echo "package v1" > services/$svc/internal/delivery/ws/v1/doc.go
        elif [ "$svc" == "chat" ]; then
            mkdir -p services/$svc/internal/delivery/{http/v1,ws/v1}
            echo "package v1" > services/$svc/internal/delivery/http/v1/doc.go
            echo "package v1" > services/$svc/internal/delivery/ws/v1/doc.go
        else
            mkdir -p services/$svc/internal/delivery/grpc/v1
            echo "package v1" > services/$svc/internal/delivery/grpc/v1/doc.go
        fi

        # Usecase & Repository
        mkdir -p services/$svc/internal/{usecase/v1,repository}
        echo "package v1" > services/$svc/internal/usecase/v1/doc.go
        echo "package repository" > services/$svc/internal/repository/doc.go
    fi
done

# 4. Buat Struktur Proto (gRPC Contracts - Versioned)
PROTO_SVCS=("location" "order" "customer" "driver" "chat" "map")
for p in "${PROTO_SVCS[@]}"; do
    mkdir -p proto/$p/v1
    # Buat dummy .proto file agar folder terisi dan siap untuk protoc
    cat <<EOF > proto/$p/v1/$p.proto
syntax = "proto3";
package $p.v1;
option go_package = "github.com/hodeifa/hyperlocal-backend/proto/$p/v1;$p";
EOF
done

# 5. Buat Struktur pkg (Shared Libraries)
mkdir -p pkg
cat <<EOF > pkg/go.mod
module github.com/hodeifa/hyperlocal-backend/pkg
go 1.22
EOF
for dir in "${PKG_DIRS[@]}"; do
    mkdir -p pkg/$dir
    echo "package $dir" > pkg/$dir/doc.go
done

# 6. Buat Folder Pendukung Lainnya
mkdir -p migrations docs bruno/collections scripts

# 7. Setup Go Workspace (go.work)
cat <<EOF > go.work
go 1.22

use (
    ./api-gateway
    ./pkg
    ./services/location
    ./services/order
    ./services/customer
    ./services/driver
    ./services/chat
    ./services/chat-worker
    ./services/map
)
EOF

echo "✅ Scaffolding selesai!"
echo "📂 Struktur folder Clean Architecture + Versioning (v1/v2) telah dibuat."
echo "🛠️  Silakan jalankan: go build ./... untuk memverifikasi."