#!/bin/bash

echo "🔍 Mencari semua file go.mod di repository (mengabaikan folder vendor)..."

# Cari semua go.mod, lalu update versi Go-nya menggunakan go mod edit
find . -type f -name "go.mod" -not -path "*/vendor/*" | while read -r file; do
    dir=$(dirname "$file")
    echo "⚙️  Bumping $dir ke Go 1.25.0..."
    (cd "$dir" && go mod edit -go=1.25.0)
done

echo ""
echo "✅ Semua go.mod berhasil di-bump ke 1.25.0!"
echo "🧹 Menjalankan go mod tidy di semua module..."

# Jalankan go mod tidy untuk merapikan dependencies
find . -type f -name "go.mod" -not -path "*/vendor/*" | while read -r file; do
    dir=$(dirname "$file")
    (cd "$dir" && go mod tidy)
done

echo ""
echo "🔄 Sinkronisasi go.work..."
go work sync

echo "🎉 Selesai! Workspace sekarang konsisten di Go 1.25.0."