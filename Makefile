# ==============================================================================
# HYPERLOCAL BACKEND - MAKEFILE
# ==============================================================================

# 1. Load file .env jika ada. 
# Tanda '-' di depan membuat Make TIDAK akan error/crash jika file .env belum ada.
-include .env

# 2. Export SEMUA variabel ke environment shell
export

# Daftar semua modul dalam workspace
MODULES := api-gateway pkg \
           services/location \
           services/order \
           services/customer \
           services/driver \
           services/chat \
           services/chat-worker \
           services/map

# Variabel Database (Menggunakan sintaks Make $(VAR) - Dari .env)
DB_URL := "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(POSTGRES_DB)?sslmode=$(DB_SSLMODE)"

# Variabel Database KHUSUS untuk CLI golang-migrate (Hardcoded ke localhost & sslmode=disable)
MIGRATE_DB_URL := "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(DB_PORT)/$(POSTGRES_DB)?sslmode=disable"

# Target default: Build semua modul
.PHONY: build
build:
	@echo "🚀 Starting build for all modules..."
	@for mod in $(MODULES); do \
		echo "📦 Building $$mod..."; \
		(cd $$mod && go build ./...) || exit 1; \
	done
	@echo "✅ All modules built successfully."

# Target untuk tidy dependencies
.PHONY: tidy
tidy:
	@for mod in $(MODULES); do \
		echo "🧹 Tidy $$mod..."; \
		(cd $$mod && go mod tidy) || exit 1; \
	done

# Target untuk menjalankan test di semua modul (dengan race detector)
.PHONY: test
test:
	@for mod in $(MODULES); do \
		echo "🧪 Testing $$mod..."; \
		(cd $$mod && go test -race ./...) || exit 1; \
	done

# Target untuk clean binary files
.PHONY: clean
clean:
	@echo "🧹 Cleaning up..."
	@find . -type f -name '*.exe' -delete
	@find . -type f -name '*.test' -delete
	@find . -type f -name '*.out' -delete

# ==============================================================================
# DATABASE MIGRATIONS (golang-migrate)
# ==============================================================================

# Jalankan semua migrasi yang belum dieksekusi
.PHONY: migrate-up
migrate-up:
	@echo "🚀 Running migrations UP..."
	migrate -path ./migrations -database $(MIGRATE_DB_URL) up

# Rollback 1 langkah migrasi terakhir
.PHONY: migrate-down
migrate-down:
	@echo "⏪ Running migrations DOWN (1 step)..."
	migrate -path ./migrations -database $(MIGRATE_DB_URL) down 1

# Cek status migrasi
.PHONY: migrate-status
migrate-status:
	@echo "📊 Migration status:"
	migrate -path ./migrations -database $(MIGRATE_DB_URL) version

# Buat file migrasi baru (contoh: make migrate-create name=create_users)
.PHONY: migrate-create
migrate-create:
	@echo "📝 Creating new migration files..."
	migrate create -ext sql -dir ./migrations -seq $(name)

# ==============================================================================
# LINTING (Code Quality & Security)
# ==============================================================================

.PHONY: lint
lint:
	@echo "🔍 Running golangci-lint across all modules..."
	@for mod in $(MODULES); do \
		echo "🧹 Linting $$mod..."; \
		(cd $$mod && golangci-lint run ./...) || exit 1; \
	done
	@echo "✅ All modules passed linting!"