# Go Taxi Fare Calculator

A high-performance, production-ready taxi fare calculation system written in Go. Processes time-stamped distance records from stdin and calculates fares based on Japanese taxi fare structure with precise decimal arithmetic.

[![CI](https://github.com/user/golang-taxi-fare/workflows/CI/badge.svg)](https://github.com/user/golang-taxi-fare/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/user/golang-taxi-fare)](https://goreportcard.com/report/github.com/user/golang-taxi-fare)
[![Coverage](https://codecov.io/gh/user/golang-taxi-fare/branch/main/graph/badge.svg)](https://codecov.io/gh/user/golang-taxi-fare)

## Features

- **Accurate Fare Calculation**: Implements Japanese taxi fare structure with tiered pricing
- **Streaming Processing**: Memory-efficient processing of large datasets via stdin
- **Precision Arithmetic**: Uses `shopspring/decimal` for exact monetary calculations
- **Robust Validation**: Comprehensive timing and format validation with detailed error reporting
- **Structured Logging**: JSON-structured logs to stderr for observability
- **Cross-Platform**: Builds for Linux, macOS, Windows, and FreeBSD
- **Production-Ready**: Signal handling, graceful shutdown, and comprehensive error handling

## Quick Start

```bash
# Clone and build
git clone https://github.com/user/golang-taxi-fare.git
cd golang-taxi-fare
make build

# Run with sample data
echo -e "12:34:56.789 12345678.5\n12:34:57.123 12345679.1" | ./taxi-fare
```

## Architecture

The application follows a modular architecture with clean separation of concerns:

```
├── inputparser/     # Streaming stdin parser with robust error handling
├── datavalidator/   # Timing constraints and format validation  
├── farecalculator/  # Japanese taxi fare calculation logic
├── outputformatter/ # Console output formatting with tabwriter
├── loggingsystem/   # Structured JSON logging to stderr
├── errorhandler/    # Centralized error processing with exit codes
├── models/          # Core data structures and types
└── main.go          # Application orchestration and signal handling
```

## Input Format

The application expects time-stamped distance records in the format:
```
hh:mm:ss.fff xxxxxxxx.f
```

Where:
- `hh:mm:ss.fff`: Timestamp (24-hour format with millisecond precision)
- `xxxxxxxx.f`: Distance reading (8+ digits before decimal, 1+ after)

### Example Input
```
12:34:56.789 12345678.5
12:34:57.123 12345679.1  
12:34:58.456 12345680.3
```

## Fare Calculation

Implements Japanese taxi fare structure:
- **Base Fare**: 400 yen for distance ≤ 1km
- **Standard Rate**: 40 yen per 400m for distances 1-10km
- **Extended Rate**: 40 yen per 350m for distances > 10km

## Validation Rules

- **Sequential Timestamps**: Records must be in chronological order
- **Maximum Interval**: No more than 5 minutes between consecutive records
- **Distance Progression**: Distance readings must be non-decreasing
- **Format Compliance**: Strict adherence to input format specification

## Output

The application provides:
1. **Fare Amount**: Integer fare in yen to stdout
2. **Processing Summary**: Record count, processing time, and total fare
3. **Structured Logs**: JSON logs to stderr for monitoring and debugging

### Example Output
```
400

Processing Summary:
Records processed: 3
Processing time: 1.234ms
Total fare: 400 yen
```

## Error Handling

The application uses specific exit codes for different error conditions:
- **0**: Success
- **1**: Format error (invalid input format)
- **2**: Timing error (timing constraint violation)
- **3**: Insufficient data (no valid records)
- **4**: Calculation error (computational failure)
- **5**: General error (other failures)

## Building and Testing

### Prerequisites
- Go 1.21 or later
- Make (optional, for build automation)

### Build Commands

```bash
# Development build
make build

# Cross-platform builds
make build-all

# With custom build script
./scripts/build.sh
```

### Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run comprehensive test suite
./scripts/test-runner.sh

# Run specific test types
make test-race          # Race condition detection
make benchmark          # Performance benchmarks
make test-integration   # Integration tests
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Static analysis
make vet

# Complete quality check
make check
```

## Performance

The application is optimized for high throughput with:
- **Streaming Processing**: Memory usage independent of input size
- **Buffered I/O**: Optimized for large datasets  
- **Concurrent Safety**: Race-condition free with proper synchronization
- **Decimal Precision**: Exact arithmetic without floating-point errors

### Benchmarks
```bash
make benchmark
```

Typical performance characteristics:
- **Throughput**: 10,000+ records/second
- **Memory Usage**: < 10MB baseline
- **Precision**: Exact decimal arithmetic to 18+ decimal places

## Configuration

### Environment Variables
- `LOG_LEVEL`: Set logging level (DEBUG, INFO, WARN, ERROR)
- `MAX_INTERVAL`: Override default 5-minute maximum interval

### Signal Handling
- **SIGINT/SIGTERM**: Graceful shutdown with cleanup
- **Context Cancellation**: Proper resource cleanup on termination

## Development

### Project Structure
```
├── .github/workflows/  # CI/CD configuration
├── scripts/           # Build and test automation scripts  
├── test-data/         # Sample data files for testing
├── Makefile          # Build automation
└── README.md         # This file
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Run the full test suite: `make validate`
5. Submit a pull request

### Code Standards
- Go 1.21+ with modules
- gofmt formatted code
- Comprehensive test coverage (>70%)
- Documentation for exported functions
- Structured error handling

## CI/CD

The project includes comprehensive CI/CD with GitHub Actions:
- **Multi-version testing** (Go 1.21, 1.22)
- **Cross-platform builds** (Linux, macOS, Windows)
- **Code quality checks** (vet, fmt, staticcheck)
- **Performance benchmarks**
- **Integration testing**

## Monitoring and Observability

### Structured Logging
All logs are output as JSON to stderr:
```json
{
  "time": "2024-01-15T12:34:56.789Z",
  "level": "INFO", 
  "msg": "Processing completed",
  "component": "main",
  "record_count": 1000,
  "processing_time_ms": 123
}
```

### Metrics
The application logs key metrics:
- Processing time and throughput
- Record counts and error rates  
- Memory usage and performance characteristics

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or contributions:
- **Issues**: [GitHub Issues](https://github.com/user/golang-taxi-fare/issues)
- **Documentation**: This README and inline code documentation
- **CI Status**: [GitHub Actions](https://github.com/user/golang-taxi-fare/actions)

---

Built with ❤️ in Go | Production-ready taxi fare calculation system