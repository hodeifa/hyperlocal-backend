#!/bin/bash

# Definisi Warna
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${MAGENTA}============================================================${NC}"
echo -e "${MAGENTA}   🐳 HYPERLOCAL INTEGRATION TEST RUNNER (TESTCONTAINERS)   ${NC}"
echo -e "${MAGENTA}============================================================${NC}"
echo ""

# 1. Pengecekan Docker (Sangat Krusial untuk testcontainers-go)
if ! docker info > /dev/null 2>&1; then
  echo -e "${RED}❌ FATAL: Docker tidak berjalan!${NC}"
  echo -e "${YELLOW}⚠️  Integration test mewajibkan Docker Daemon aktif untuk testcontainers-go.${NC}"
  exit 1
fi
echo -e "${GREEN}✅ Docker aktif. Memulai pemindaian module...${NC}"
echo ""

FAILED_MODULES=()
PASSED_MODULES=()
SKIPPED_MODULES=()

# 2. Auto-discovery module
MODULES=$(find . -type f -name "go.mod" -not -path "*/vendor/*" -not -path "*/.git/*" | sort | xargs -n 1 dirname)

TOTAL_MODULES=$(echo "$MODULES" | wc -l)
CURRENT=0

# 3. Looping per Module
for module_dir in $MODULES; do
    CURRENT=$((CURRENT + 1))
    module_name=${module_dir#./}
    [ "$module_name" = "." ] && module_name="root"

    echo -e "${YELLOW}------------------------------------------------------------${NC}"
    echo -e "${CYAN}📦 [$CURRENT/$TOTAL_MODULES] Memulai Integration Test: ${GREEN}$module_name${NC}"
    echo -e "${YELLOW}------------------------------------------------------------${NC}"
    
    # Cek apakah module ini memiliki file dengan tag 'integration'
    # Kita gunakan grep sederhana untuk memastikan agar tidak ada warning "no packages to test"
    HAS_INTEGRATION_TEST=$(cd "$module_dir" && grep -rl "//go:build integration" . --include="*_test.go" 2>/dev/null)
    
    if [ -z "$HAS_INTEGRATION_TEST" ]; then
        echo -e "${BLUE}ℹ️  SKIP: Tidak ada file integration test di $module_name${NC}"
        SKIPPED_MODULES+=("$module_name")
        echo ""
        continue
    fi

    # Jalankan test khusus integration
    # -tags=integration : Hanya compile dan run file yang punya tag ini
    # -timeout 15m      : Memberi waktu lebih untuk pull image & spin-up container
    # -p 1              : Menjalankan test package secara sequential (menghindari OOM/Resource hogging jika banyak container nyala bareng)
    (cd "$module_dir" && go test -tags=integration -v -count=1 -timeout 15m -p 1 ./...)
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
echo -e "${MAGENTA}============================================================${NC}"
echo -e "${MAGENTA}                       📊 LAPORAN AKHIR                     ${NC}"
echo -e "${MAGENTA}============================================================${NC}"

echo -e "${GREEN}✅ LULUS (${#PASSED_MODULES[@]}):${NC}"
for p in "${PASSED_MODULES[@]}"; do echo "   - $p"; done

if [ ${#SKIPPED_MODULES[@]} -gt 0 ]; then
    echo -e "\n${BLUE}ℹ️  DILEWATI (${#SKIPPED_MODULES[@]} - Tidak ada integration test):${NC}"
    for s in "${SKIPPED_MODULES[@]}"; do echo "   - $s"; done
fi

echo ""

if [ ${#FAILED_MODULES[@]} -gt 0 ]; then
    echo -e "${RED}❌ GAGAL (${#FAILED_MODULES[@]}):${NC}"
    for f in "${FAILED_MODULES[@]}"; do echo "   - $f"; done
    echo ""
    echo -e "${RED}⛔ Ada container atau logic DB yang gagal. Cek log di atas.${NC}"
    exit 1
else
    echo -e "${GREEN}🎉 SEMUA INTEGRATION TEST BERHASIL DILALUI! 🎉${NC}"
    exit 0
fi