#!/bin/bash
# Cross-platform build script for Go Taxi Fare Calculator

set -e

# Configuration
VERSION=${VERSION:-"1.0.0"}

# Cross-platform date command
if command -v date >/dev/null 2>&1; then
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" || "$OSTYPE" == "cygwin" ]]; then
        BUILD_TIME=$(powershell -Command "Get-Date -Format 'yyyy-MM-dd HH:mm:ss' -AsUTC" 2>/dev/null || echo "unknown")
    else
        BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S' 2>/dev/null || echo "unknown")
    fi
else
    BUILD_TIME="unknown"
fi

# Safe git commit retrieval
if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
else
    GIT_COMMIT="unknown"
fi

# Colors and logging functions
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

# Build settings
BINARY_NAME="taxi-fare"

# Check if main.go has version variables for LDFLAGS injection
if grep -q "var.*Version\|var.*BuildTime\|var.*GitCommit" main.go 2>/dev/null; then
    LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime='${BUILD_TIME}' -X main.GitCommit=${GIT_COMMIT}"
    log_info "Version injection enabled"
else
    LDFLAGS=""
    log_info "No version variables found, building without version injection"
fi

# Create build directory
BUILD_DIR="builds"
mkdir -p "${BUILD_DIR}"

log_info "Starting cross-platform build..."
log_info "Version: ${VERSION}"
log_info "Build Time: ${BUILD_TIME}"
log_info "Git Commit: ${GIT_COMMIT}"

# Platform configurations - simple approach
PLATFORMS=("linux-amd64" "linux-arm64" "darwin-amd64" "darwin-arm64" "windows-amd64" "freebsd-amd64")
OS_VALUES=("linux" "linux" "darwin" "darwin" "windows" "freebsd")
ARCH_VALUES=("amd64" "arm64" "amd64" "arm64" "amd64" "amd64")

# Build counter
BUILT_COUNT=0
FAILED_COUNT=0

# Build for each platform
for i in $(seq 0 $((${#PLATFORMS[@]} - 1))); do
    platform_key="${PLATFORMS[$i]}"
    GOOS="${OS_VALUES[$i]}"
    GOARCH="${ARCH_VALUES[$i]}"
    
    OUTPUT_NAME="${BINARY_NAME}-${platform_key}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    OUTPUT_PATH="${BUILD_DIR}/${OUTPUT_NAME}"
    
    log_info "Building for ${GOOS}/${GOARCH}..."
    
    # Build with appropriate flags
    if [ -n "$LDFLAGS" ]; then
        if env GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags "$LDFLAGS" -o "$OUTPUT_PATH" . 2>/dev/null; then
            BUILD_SUCCESS=true
        else
            BUILD_SUCCESS=false
        fi
    else
        if env GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$OUTPUT_PATH" . 2>/dev/null; then
            BUILD_SUCCESS=true
        else
            BUILD_SUCCESS=false
        fi
    fi
    
    if [ "$BUILD_SUCCESS" = true ]; then
        # Simple file size calculation
        if [ -f "$OUTPUT_PATH" ]; then
            SIZE=$(ls -l "$OUTPUT_PATH" | awk '{print $5}')
            if [ "$SIZE" -gt 0 ]; then
                SIZE_MB=$((SIZE / 1048576))
                if [ $SIZE_MB -eq 0 ]; then
                    SIZE_MB="0.$(( (SIZE * 100) / 1048576 ))"
                fi
                log_success "Built ${OUTPUT_NAME} (${SIZE_MB}MB)"
            else
                log_success "Built ${OUTPUT_NAME}"
            fi
        else
            log_success "Built ${OUTPUT_NAME}"
        fi
        BUILT_COUNT=$((BUILT_COUNT + 1))
    else
        log_error "Failed to build for ${GOOS}/${GOARCH}"
        FAILED_COUNT=$((FAILED_COUNT + 1))
    fi
done

# Generate checksums
log_info "Generating checksums..."
CHECKSUM_FILE="${BUILD_DIR}/checksums.txt"
> "$CHECKSUM_FILE"

if command -v sha256sum >/dev/null 2>&1; then
    (cd "${BUILD_DIR}" && sha256sum taxi-fare-* >> checksums.txt 2>/dev/null)
elif command -v shasum >/dev/null 2>&1; then
    (cd "${BUILD_DIR}" && shasum -a 256 taxi-fare-* >> checksums.txt 2>/dev/null)
else
    log_error "No checksum utility found"
fi

# Create archive
log_info "Creating release archive..."
ARCHIVE_NAME="taxi-fare-${VERSION}-all-platforms.tar.gz"
if command -v tar >/dev/null 2>&1; then
    tar -czf "${ARCHIVE_NAME}" -C "${BUILD_DIR}" . 2>/dev/null || log_error "Failed to create archive"
else
    log_error "tar not found, skipping archive creation"
fi

# Summary
echo
echo "========================================"
log_info "Build Summary:"
echo "  Successfully built: ${BUILT_COUNT}"
echo "  Failed: ${FAILED_COUNT}"
echo "  Binaries location: ./${BUILD_DIR}/"
if [ -f "${ARCHIVE_NAME}" ]; then
    echo "  Archive created: ./${ARCHIVE_NAME}"
fi

if [ "$FAILED_COUNT" -eq 0 ]; then
    log_success "All builds completed successfully! 🎉"
else
    log_error "Some builds failed! ❌"
    exit 1
fi

# Optional: Test local binary
CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
CURRENT_ARCH=$(uname -m)
if [ "$CURRENT_ARCH" = "x86_64" ]; then
    CURRENT_ARCH="amd64"
fi

TEST_BINARY="${BUILD_DIR}/taxi-fare-${CURRENT_OS}-${CURRENT_ARCH}"
if [ "$CURRENT_OS" = "windows" ] || [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" || "$OSTYPE" == "cygwin" ]]; then
    TEST_BINARY="${TEST_BINARY}.exe"
fi

if [ -f "$TEST_BINARY" ]; then
    log_info "Testing local binary..."
    if echo "12:34:56.789 12345678.5" | "$TEST_BINARY" > /dev/null 2>&1; then
        log_success "Local binary test passed"
    else
        log_error "Local binary test failed"
    fi
fi

echo "🚀 Cross-platform build completed!"