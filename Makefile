# ==============================================================================
# HYPERLOCAL BACKEND - MAKEFILE
# ==============================================================================

# Daftar semua modul dalam workspace
MODULES := api-gateway pkg \
           services/location \
           services/order \
           services/customer \
           services/driver \
           services/chat \
           services/chat-worker \
           services/map

# Target default: Build semua modul
.PHONY: build
build:
	@echo "🚀 Starting build for all modules..."
	@for mod in $(MODULES); do \
		echo "📦 Building $$mod..."; \
		(cd $$mod && go build ./...) || exit 1; \
	done
	@echo "✅ All modules built successfully."

# Target untuk tidy dependencies (berguna nanti saat import library)
.PHONY: tidy
tidy:
	@for mod in $(MODULES); do \
		echo "🧹 Tidy $$mod..."; \
		(cd $$mod && go mod tidy); \
	done

# Target untuk menjalankan test di semua modul
.PHONY: test
test:
	@for mod in $(MODULES); do \
		echo "🧪 Testing $$mod..."; \
		(cd $$mod && go test ./...); \
	done

# Target untuk clean binary files (jika ada)
.PHONY: clean
clean:
	@echo "🧹 Cleaning up..."
	@find . -type f -name '*.exe' -delete
	@find . -type f -name '*.test' -delete
	@find . -type f -name '*.out' -delete
