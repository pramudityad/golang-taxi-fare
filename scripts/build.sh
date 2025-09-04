#!/bin/bash
# Cross-platform build script for Go Taxi Fare Calculator

set -e

# Configuration
VERSION=${VERSION:-"1.0.0"}
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build settings
BINARY_NAME="taxi-fare"
LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime='${BUILD_TIME}' -X main.GitCommit=${GIT_COMMIT}"

# Platform configurations
declare -A PLATFORMS=(
    ["linux-amd64"]="linux amd64"
    ["linux-arm64"]="linux arm64"
    ["darwin-amd64"]="darwin amd64"
    ["darwin-arm64"]="darwin arm64"
    ["windows-amd64"]="windows amd64"
    ["freebsd-amd64"]="freebsd amd64"
)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create build directory
BUILD_DIR="builds"
mkdir -p "${BUILD_DIR}"

log_info "Starting cross-platform build..."
log_info "Version: ${VERSION}"
log_info "Build Time: ${BUILD_TIME}"
log_info "Git Commit: ${GIT_COMMIT}"

# Build counter
BUILT_COUNT=0
FAILED_COUNT=0

# Build for each platform
for platform_key in "${!PLATFORMS[@]}"; do
    platform_info="${PLATFORMS[$platform_key]}"
    GOOS=$(echo $platform_info | cut -d' ' -f1)
    GOARCH=$(echo $platform_info | cut -d' ' -f2)
    
    OUTPUT_NAME="${BINARY_NAME}-${platform_key}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    OUTPUT_PATH="${BUILD_DIR}/${OUTPUT_NAME}"
    
    log_info "Building for ${GOOS}/${GOARCH}..."
    
    if env GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags "$LDFLAGS" -o "$OUTPUT_PATH" .; then
        # Get file size
        if command -v stat >/dev/null 2>&1; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                SIZE=$(stat -f%z "$OUTPUT_PATH")
            else
                SIZE=$(stat -c%s "$OUTPUT_PATH")
            fi
            SIZE_MB=$(echo "scale=2; $SIZE / 1024 / 1024" | bc -l 2>/dev/null || echo "unknown")
            log_success "Built ${OUTPUT_NAME} (${SIZE_MB}MB)"
        else
            log_success "Built ${OUTPUT_NAME}"
        fi
        ((BUILT_COUNT++))
    else
        log_error "Failed to build for ${GOOS}/${GOARCH}"
        ((FAILED_COUNT++))
    fi
done

# Generate checksums
log_info "Generating checksums..."
CHECKSUM_FILE="${BUILD_DIR}/checksums.txt"
> "$CHECKSUM_FILE"

if command -v sha256sum >/dev/null 2>&1; then
    (cd "${BUILD_DIR}" && sha256sum taxi-fare-* >> checksums.txt)
elif command -v shasum >/dev/null 2>&1; then
    (cd "${BUILD_DIR}" && shasum -a 256 taxi-fare-* >> checksums.txt)
else
    log_error "No checksum utility found"
fi

# Create archive
log_info "Creating release archive..."
ARCHIVE_NAME="taxi-fare-${VERSION}-all-platforms.tar.gz"
tar -czf "${ARCHIVE_NAME}" -C "${BUILD_DIR}" .

# Summary
echo
echo "========================================"
log_info "Build Summary:"
echo "  Successfully built: ${BUILT_COUNT}"
echo "  Failed: ${FAILED_COUNT}"
echo "  Binaries location: ./${BUILD_DIR}/"
echo "  Archive created: ./${ARCHIVE_NAME}"

if [ "$FAILED_COUNT" -eq 0 ]; then
    log_success "All builds completed successfully! 🎉"
else
    log_error "Some builds failed! ❌"
    exit 1
fi

# Optional: Test local binary
if [ -f "${BUILD_DIR}/taxi-fare-linux-amd64" ] && [[ "$OSTYPE" == "linux"* ]]; then
    log_info "Testing Linux binary..."
    echo "12:34:56.789 12345678.5" | "${BUILD_DIR}/taxi-fare-linux-amd64" > /dev/null && log_success "Linux binary test passed"
elif [ -f "${BUILD_DIR}/taxi-fare-darwin-amd64" ] && [[ "$OSTYPE" == "darwin"* ]]; then
    log_info "Testing macOS binary..."
    echo "12:34:56.789 12345678.5" | "${BUILD_DIR}/taxi-fare-darwin-amd64" > /dev/null && log_success "macOS binary test passed"
fi

echo "🚀 Cross-platform build completed!"