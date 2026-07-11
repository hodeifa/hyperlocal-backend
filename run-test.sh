#!/bin/bash

# Definisi Warna untuk Terminal (Agar tracing lebih mudah dibaca)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}============================================================${NC}"
echo -e "${BLUE}       🧪 HYPERLOCAL BACKEND TEST RUNNER (PER MODULE)       ${NC}"
echo -e "${BLUE}============================================================${NC}"
echo ""

# 1. Pengecekan Docker (Wajib untuk testcontainers-go)
if ! docker info > /dev/null 2>&1; then
  echo -e "${RED}❌ ERROR: Docker tidak berjalan atau tidak terinstal!${NC}"
  echo -e "${YELLOW}⚠️  Karena Anda menggunakan testcontainers-go, Docker WAJIB aktif.${NC}"
  echo -e "${YELLOW}Silakan nyalakan Docker Desktop / Docker Daemon terlebih dahulu.${NC}"
  exit 1
fi
echo -e "${GREEN}✅ Docker terdeteksi aktif. Memulai pemindaian module...${NC}"
echo ""

FAILED_MODULES=()
PASSED_MODULES=()

# 2. Mencari semua module secara dinamis berdasarkan file go.mod
# Mengabaikan folder vendor, .git, dan node_modules (jika ada)
MODULES=$(find . -type f -name "go.mod" -not -path "*/vendor/*" -not -path "*/.git/*" -not -path "*/node_modules/*" | sort | xargs -n 1 dirname)

TOTAL_MODULES=$(echo "$MODULES" | wc -l)
CURRENT=0

# 3. Looping per Module
for module_dir in $MODULES; do
    CURRENT=$((CURRENT + 1))
    
    # Bersihkan prefix './' agar nama module lebih rapi di output
    module_name=${module_dir#./}
    if [ "$module_name" = "." ]; then
        module_name="root"
    fi

    echo -e "${YELLOW}------------------------------------------------------------${NC}"
    echo -e "${CYAN}📦 [$CURRENT/$TOTAL_MODULES] Memulai Test: ${GREEN}$module_name${NC}"
    echo -e "${YELLOW}------------------------------------------------------------${NC}"
    
    # Masuk ke direktori module dan jalankan test
    # -v: Verbose (menampilkan nama fungsi yang sedang ditest)
    # -count=1: Disable cache (memastikan test dijalankan fresh, bukan hasil cache)
    # -race: Mendeteksi race condition pada concurrency
    (cd "$module_dir" && go test -v -count=1 -race ./...)
    TEST_EXIT_CODE=$?
    
    echo ""
    if [ $TEST_EXIT_CODE -eq 0 ]; then
        echo -e "${GREEN}✅ SUKSES: $module_name${NC}"
        PASSED_MODULES+=("$module_name")
    else
        echo -e "${RED}❌ GAGAL: $module_name${NC}"
        FAILED_MODULES+=("$module_name")
    fi
    echo ""
done

# 4. Laporan Akhir (Summary)
echo -e "${BLUE}============================================================${NC}"
echo -e "${BLUE}                       📊 LAPORAN AKHIR                     ${NC}"
echo -e "${BLUE}============================================================${NC}"
echo -e "${GREEN}✅ LULUS (${#PASSED_MODULES[@]}):${NC}"
for p in "${PASSED_MODULES[@]}"; do echo "   - $p"; done

echo ""

if [ ${#FAILED_MODULES[@]} -gt 0 ]; then
    echo -e "${RED}❌ GAGAL (${#FAILED_MODULES[@]}):${NC}"
    for f in "${FAILED_MODULES[@]}"; do echo "   - $f"; done
    echo ""
    echo -e "${RED}⛔ Ada test yang gagal. Silakan scroll ke atas untuk tracing error pada module tersebut.${NC}"
    exit 1
else
    echo -e "${GREEN}🎉 SEMUA MODULE BERHASIL DILALUI! 🎉${NC}"
    exit 0
fi