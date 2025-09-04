#!/bin/bash
# Comprehensive test runner script for Go Taxi Fare Calculator
# Provides detailed testing with CI-friendly output

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TEST_OUTPUT_DIR="${PROJECT_ROOT}/test-results"
COVERAGE_THRESHOLD=70

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test result tracking
TESTS_PASSED=0
TESTS_FAILED=0

# Setup function
setup() {
    log_info "Setting up test environment..."
    
    cd "${PROJECT_ROOT}"
    
    # Create test results directory
    mkdir -p "${TEST_OUTPUT_DIR}"
    
    # Ensure we have test data
    make test-data > /dev/null 2>&1 || true
    
    log_success "Test environment ready"
}

# Unit test runner
run_unit_tests() {
    log_info "Running unit tests..."
    
    local output_file="${TEST_OUTPUT_DIR}/unit-tests.out"
    
    if go test -v -json ./datavalidator/... ./errorhandler/... ./farecalculator/... ./inputparser/... ./loggingsystem/... ./models/... ./outputformatter/... > "${output_file}" 2>&1; then
        log_success "Unit tests passed"
        ((TESTS_PASSED++))
    else
        log_error "Unit tests failed"
        cat "${output_file}"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Coverage test runner
run_coverage_tests() {
    log_info "Running coverage tests..."
    
    local coverage_file="${TEST_OUTPUT_DIR}/coverage.out"
    local coverage_report="${TEST_OUTPUT_DIR}/coverage-report.txt"
    
    if go test -coverprofile="${coverage_file}" ./datavalidator/... ./errorhandler/... ./farecalculator/... ./inputparser/... ./loggingsystem/... ./models/... ./outputformatter/...; then
        # Generate coverage report
        go tool cover -func="${coverage_file}" > "${coverage_report}"
        
        # Extract total coverage
        local total_coverage=$(go tool cover -func="${coverage_file}" | grep "total:" | awk '{print $3}' | sed 's/%//')
        
        if (( $(echo "${total_coverage} >= ${COVERAGE_THRESHOLD}" | bc -l) )); then
            log_success "Coverage tests passed (${total_coverage}% >= ${COVERAGE_THRESHOLD}%)"
            ((TESTS_PASSED++))
        else
            log_error "Coverage tests failed (${total_coverage}% < ${COVERAGE_THRESHOLD}%)"
            ((TESTS_FAILED++))
            return 1
        fi
    else
        log_error "Coverage tests failed to run"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Race condition test runner
run_race_tests() {
    log_info "Running race condition tests..."
    
    local output_file="${TEST_OUTPUT_DIR}/race-tests.out"
    
    if go test -race -v ./datavalidator/... ./errorhandler/... ./farecalculator/... ./inputparser/... ./loggingsystem/... ./models/... ./outputformatter/... > "${output_file}" 2>&1; then
        log_success "Race condition tests passed"
        ((TESTS_PASSED++))
    else
        log_error "Race condition tests failed"
        cat "${output_file}"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Integration test runner
run_integration_tests() {
    log_info "Running integration tests..."
    
    # Build the application first
    if ! make build > /dev/null 2>&1; then
        log_error "Failed to build application for integration tests"
        ((TESTS_FAILED++))
        return 1
    fi
    
    local failed=0
    
    # Test 1: Valid input processing
    log_info "Testing valid input processing..."
    if echo -e "12:34:56.789 12345678.5\n12:34:57.123 12345679.1" | ./taxi-fare > /dev/null 2>&1; then
        log_success "Valid input test passed"
    else
        log_error "Valid input test failed"
        ((failed++))
    fi
    
    # Test 2: Test with file input
    log_info "Testing file input processing..."
    if ./taxi-fare < test-data/valid_input.txt > /dev/null 2>&1; then
        log_success "File input test passed"
    else
        log_error "File input test failed"
        ((failed++))
    fi
    
    # Test 3: Invalid input handling
    log_info "Testing invalid input handling..."
    if ! echo "invalid input" | ./taxi-fare > /dev/null 2>&1; then
        log_success "Invalid input test passed (correctly failed)"
    else
        log_error "Invalid input test failed (should have failed)"
        ((failed++))
    fi
    
    # Test 4: Empty input handling
    log_info "Testing empty input handling..."
    if ! echo "" | ./taxi-fare > /dev/null 2>&1; then
        log_success "Empty input test passed (correctly failed)"
    else
        log_error "Empty input test failed (should have failed)"
        ((failed++))
    fi
    
    if [ $failed -eq 0 ]; then
        log_success "All integration tests passed"
        ((TESTS_PASSED++))
    else
        log_error "Integration tests failed ($failed/$4 failed)"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Benchmark test runner
run_benchmark_tests() {
    log_info "Running benchmark tests..."
    
    local output_file="${TEST_OUTPUT_DIR}/benchmark-results.txt"
    
    if go test -bench=. -benchmem ./datavalidator/... ./errorhandler/... ./farecalculator/... ./inputparser/... ./loggingsystem/... ./models/... ./outputformatter/... > "${output_file}" 2>&1; then
        log_success "Benchmark tests completed"
        log_info "Results saved to ${output_file}"
        ((TESTS_PASSED++))
    else
        log_error "Benchmark tests failed"
        cat "${output_file}"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Code quality checks
run_quality_checks() {
    log_info "Running code quality checks..."
    
    # Go vet
    log_info "Running go vet..."
    if go vet ./...; then
        log_success "Go vet passed"
    else
        log_error "Go vet failed"
        ((TESTS_FAILED++))
        return 1
    fi
    
    # Go fmt check
    log_info "Checking code formatting..."
    if [ -z "$(gofmt -l .)" ]; then
        log_success "Code formatting check passed"
    else
        log_error "Code formatting check failed"
        gofmt -l .
        ((TESTS_FAILED++))
        return 1
    fi
    
    log_success "Code quality checks passed"
    ((TESTS_PASSED++))
}

# Platform compatibility checks
run_platform_tests() {
    log_info "Running platform compatibility tests..."
    
    local platforms=("linux/amd64" "darwin/amd64" "windows/amd64")
    local failed=0
    
    for platform in "${platforms[@]}"; do
        local goos=$(echo "$platform" | cut -d'/' -f1)
        local goarch=$(echo "$platform" | cut -d'/' -f2)
        
        log_info "Testing build for $platform..."
        
        if GOOS="$goos" GOARCH="$goarch" go build -o "taxi-fare-$goos-$goarch" .; then
            log_success "Build successful for $platform"
            rm -f "taxi-fare-$goos-$goarch" "taxi-fare-$goos-$goarch.exe"
        else
            log_error "Build failed for $platform"
            ((failed++))
        fi
    done
    
    if [ $failed -eq 0 ]; then
        log_success "All platform builds succeeded"
        ((TESTS_PASSED++))
    else
        log_error "Platform tests failed ($failed/${#platforms[@]} failed)"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Performance validation
run_performance_tests() {
    log_info "Running performance validation..."
    
    # Build application
    make build > /dev/null 2>&1 || return 1
    
    # Create large test data
    local large_input="${TEST_OUTPUT_DIR}/large_input.txt"
    local start_time="12:34:56.789"
    local start_distance="12345678.5"
    
    # Generate 1000 records
    > "$large_input"
    for i in $(seq 1 1000); do
        # Calculate timestamp (add seconds)
        local seconds=$((56 + i))
        local minutes=34
        local hours=12
        
        if [ $seconds -ge 60 ]; then
            minutes=$((34 + seconds / 60))
            seconds=$((seconds % 60))
        fi
        
        if [ $minutes -ge 60 ]; then
            hours=$((12 + minutes / 60))
            minutes=$((minutes % 60))
        fi
        
        printf "%02d:%02d:%02d.789 %d.%d\n" $hours $minutes $seconds $((12345678 + i)) $((i % 10)) >> "$large_input"
    done
    
    # Test performance
    log_info "Testing with 1000 records..."
    local start_time=$(date +%s.%N)
    if ./taxi-fare < "$large_input" > /dev/null 2>&1; then
        local end_time=$(date +%s.%N)
        local elapsed=$(echo "$end_time - $start_time" | bc)
        log_success "Performance test completed in ${elapsed}s"
        ((TESTS_PASSED++))
    else
        log_error "Performance test failed"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Generate test report
generate_report() {
    log_info "Generating test report..."
    
    local report_file="${TEST_OUTPUT_DIR}/test-report.txt"
    local total_tests=$((TESTS_PASSED + TESTS_FAILED))
    
    cat > "${report_file}" << EOF
Go Taxi Fare Calculator - Test Report
=====================================

Test Summary:
- Total Tests: $total_tests
- Passed: $TESTS_PASSED
- Failed: $TESTS_FAILED
- Success Rate: $(echo "scale=2; $TESTS_PASSED * 100 / $total_tests" | bc -l)%

Test Results Directory: $TEST_OUTPUT_DIR

Generated on: $(date)
EOF
    
    log_info "Test report saved to ${report_file}"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test artifacts..."
    rm -f taxi-fare taxi-fare-* *.prof || true
}

# Main execution
main() {
    log_info "Starting comprehensive test suite for Go Taxi Fare Calculator..."
    
    setup
    
    # Run all test suites
    run_unit_tests || true
    run_coverage_tests || true
    run_race_tests || true
    run_integration_tests || true
    run_benchmark_tests || true
    run_quality_checks || true
    run_platform_tests || true
    run_performance_tests || true
    
    # Generate report
    generate_report
    
    # Cleanup
    cleanup
    
    # Final summary
    echo
    echo "========================================"
    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All test suites passed! ($TESTS_PASSED/$((TESTS_PASSED + TESTS_FAILED)))"
        echo "🎉 Test suite execution completed successfully!"
        exit 0
    else
        log_error "Some test suites failed! ($TESTS_FAILED/$((TESTS_PASSED + TESTS_FAILED)))"
        echo "❌ Test suite execution completed with failures."
        exit 1
    fi
}

# Handle script arguments
case "${1:-all}" in
    unit)
        setup && run_unit_tests
        ;;
    coverage)
        setup && run_coverage_tests
        ;;
    race)
        setup && run_race_tests
        ;;
    integration)
        setup && run_integration_tests
        ;;
    benchmark)
        setup && run_benchmark_tests
        ;;
    quality)
        setup && run_quality_checks
        ;;
    platform)
        setup && run_platform_tests
        ;;
    performance)
        setup && run_performance_tests
        ;;
    all|*)
        main
        ;;
esac