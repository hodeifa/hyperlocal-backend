#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

echo "🚀 Starting Infrastructure Tests..."

# 1. Check Container Status
echo -e "\n🔍 [1/4] Checking Container Status..."
CONTAINERS=("hyperlocal-postgres" "hyperlocal-redis" "hyperlocal-minio")
for c in "${CONTAINERS[@]}"; do
    if [ "$(docker inspect -f '{{.State.Running}}' $c 2>/dev/null)" == "true" ]; then
        echo -e "${GREEN}✅ $c is running${NC}"
    else
        echo -e "${RED}❌ $c is NOT running. Did you run 'docker compose up -d'?${NC}"
        exit 1
    fi
done

# 2. Test PostgreSQL + PostGIS
echo -e "\n🐘 [2/4] Testing PostgreSQL & PostGIS..."

if docker exec hyperlocal-postgres psql -U hyperlocal -d hyperlocal -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ PostgreSQL connection successful${NC}"
else
    echo -e "${RED}❌ PostgreSQL connection failed${NC}"
fi

# Check PostGIS extension (akan gagal jika migrasi 001 belum dijalankan)
POSTGIS_RESULT=$(docker exec hyperlocal-postgres psql -U hyperlocal -d hyperlocal -t -c "SELECT PostGIS_Version();" 2>&1)
if [[ "$POSTGIS_RESULT" != *"ERROR"* ]] && [[ "$POSTGIS_RESULT" == *" "* ]]; then
    echo -e "${GREEN}✅ PostGIS is active (Version: $(echo $POSTGIS_RESULT | xargs))${NC}"
else
    echo -e "${YELLOW}⚠️ PostGIS is NOT enabled yet. Run migration 001_create_extensions.sql${NC}"
fi

# 3. Test Redis
echo -e "\n🔴 [3/4] Testing Redis..."
REDIS_PING=$(docker exec hyperlocal-redis redis-cli ping 2>/dev/null)
if [ "$REDIS_PING" == "PONG" ]; then
    echo -e "${GREEN}✅ Redis connection successful (PONG)${NC}"
else
    echo -e "${RED}❌ Redis connection failed${NC}"
fi

# Test SET/GET
docker exec hyperlocal-redis redis-cli set test:infra "hyperlocal_ok" > /dev/null 2>&1
REDIS_GET=$(docker exec hyperlocal-redis redis-cli get test:infra 2>/dev/null)
if [ "$REDIS_GET" == "hyperlocal_ok" ]; then
    echo -e "${GREEN}✅ Redis SET/GET operations successful${NC}"
else
    echo -e "${RED}❌ Redis SET/GET operations failed${NC}"
fi

# 4. Test MinIO
echo -e "\n📦 [4/4] Testing MinIO..."
MINIO_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9000/minio/health/live)
if [ "$MINIO_HEALTH" == "200" ]; then
    echo -e "${GREEN}✅ MinIO API is healthy (HTTP 200)${NC}"
else
    echo -e "${RED}❌ MinIO API is NOT healthy (HTTP $MINIO_HEALTH)${NC}"
fi

MINIO_CONSOLE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9001)
if [ "$MINIO_CONSOLE" == "200" ] || [ "$MINIO_CONSOLE" == "302" ]; then
    echo -e "${GREEN}✅ MinIO Console is accessible${NC}"
else
    echo -e "${RED}❌ MinIO Console is NOT accessible${NC}"
fi

echo -e "\n🎉 Infrastructure tests completed!"
