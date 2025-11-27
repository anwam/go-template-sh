# Testing Infrastructure Features

This document describes the comprehensive testing infrastructure added to generated projects.

## Overview

All generated projects now include professional-grade testing setup with:
- **testify** suite for structured tests
- **uber-go/mock** for generating mocks
- Complete test examples
- Automated test workflows
- Comprehensive testing documentation

## What's Included

### 1. Testing Dependencies

Added to `go.mod`:
```go
github.com/stretchr/testify v1.8.4
go.uber.org/mock v0.4.0
```

### 2. Handler Tests (`internal/handlers/handlers_test.go`)

**Features:**
- Uses testify suite pattern for organized tests
- Tests all HTTP handlers (Health, Ready, Index)
- Framework-specific handler tests (Gin, Echo, Fiber)
- Table-driven test examples
- JSON marshaling/unmarshaling tests

**Example:**
```go
type HandlerTestSuite struct {
    suite.Suite
    handler *Handler
    config  *config.Config
    obs     *observability.Observability
}

func (suite *HandlerTestSuite) SetupTest() {
    // Runs before each test
}

func TestHandlerSuite(t *testing.T) {
    suite.Run(t, new(HandlerTestSuite))
}
```

### 3. Mock Interfaces (`internal/mocks/interfaces.go`)

**Generated for projects with databases:**
- Interface definitions for mocking
- `go:generate` directives for mockgen
- Database operation interfaces
- Cache operation interfaces

**Example:**
```go
//go:generate mockgen -source=interfaces.go -destination=mocks.go -package=mocks

type DatabaseInterface interface {
    Ping(ctx context.Context) error
    QueryRow(ctx context.Context, query string, args ...interface{}) Row
    Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
    Exec(ctx context.Context, query string, args ...interface{}) (CommandTag, error)
}
```

### 4. Database Test Suite (`internal/database/database_test.go`)

**Generated for projects with databases:**
- Example database tests using mocks
- Shows how to use gomock controllers
- Demonstrates context and timeout usage
- Ready to extend with actual DB operations

**Example:**
```go
func (suite *DatabaseTestSuite) SetupTest() {
    ctrl := gomock.NewController(suite.T())
    suite.mockDB = mocks.NewMockDatabaseInterface(ctrl)
    
    suite.mockDB.EXPECT().Ping(gomock.Any()).Return(nil).AnyTimes()
}
```

### 5. Enhanced Makefile

**New test targets:**
- `make test` - Run all tests with coverage
- `make test-coverage` - Generate and open HTML coverage report
- `make test-unit` - Run unit tests only (with `-short` flag)
- `make test-integration` - Run integration tests only
- `make generate-mocks` - Generate mock implementations
- `make install-tools` - Install mockgen and other dev tools

**Example usage:**
```bash
# Run all tests
make test

# Generate coverage report
make test-coverage
# Opens coverage.html in browser

# Generate mocks after modifying interfaces
make generate-mocks

# Install development tools
make install-tools
```

### 6. Testing Documentation (`docs/TESTING.md`)

**Comprehensive guide covering:**
- Testing stack overview
- How to run tests
- How to generate mocks
- Test structure explanation
- Writing tests with testify suite
- Using mocks with gomock
- Table-driven test patterns
- Best practices
- CI/CD integration
- Complete code examples

## Usage Examples

### Running Tests

```bash
# All tests with coverage
make test

# Unit tests only (fast)
make test-unit

# Integration tests only
make test-integration

# Generate HTML coverage report
make test-coverage
```

### Generating Mocks

1. Define interface in `internal/mocks/interfaces.go`:
```go
type MyInterface interface {
    DoSomething(ctx context.Context, id string) error
}
```

2. Generate mocks:
```bash
make generate-mocks
```

3. Use in tests:
```go
ctrl := gomock.NewController(t)
defer ctrl.Finish()

mockObj := mocks.NewMockMyInterface(ctrl)
mockObj.EXPECT().DoSomething(gomock.Any(), "123").Return(nil)
```

### Writing Suite-Based Tests

```go
type MyTestSuite struct {
    suite.Suite
    myService *MyService
}

func (suite *MyTestSuite) SetupTest() {
    suite.myService = NewMyService()
}

func (suite *MyTestSuite) TestMyFeature() {
    result := suite.myService.DoSomething()
    suite.Equal("expected", result)
    suite.NoError(err)
}

func TestMySuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}
```

### Table-Driven Tests

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid input", "valid", false},
        {"invalid input", "invalid", true},
        {"empty input", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Benefits

### 1. Production-Ready Testing
- Industry-standard tools (testify, gomock)
- Complete test examples
- Ready to extend

### 2. Improved Code Quality
- Easy to write tests
- Encourages TDD
- Mock external dependencies

### 3. Developer Experience
- Simple commands (`make test`)
- Clear documentation
- Working examples

### 4. CI/CD Ready
- Tests run with `-race` flag
- Coverage reports generated
- Exit codes for CI integration

### 5. Maintainability
- Suite pattern for test organization
- Setup/teardown per test
- Centralized mock generation

## Architecture

### Test Organization

```
project/
├── internal/
│   ├── handlers/
│   │   ├── handlers.go
│   │   └── handlers_test.go       # Handler tests with suite
│   ├── database/
│   │   ├── postgres.go
│   │   └── database_test.go       # DB tests with mocks
│   └── mocks/
│       ├── interfaces.go          # Interface definitions
│       └── mocks.go               # Generated mocks (gitignore)
├── docs/
│   └── TESTING.md                 # Complete testing guide
├── Makefile                       # Test targets
└── coverage.out                   # Coverage data (gitignore)
```

### Test Flow

1. **Write Interface** (`internal/mocks/interfaces.go`)
2. **Generate Mocks** (`make generate-mocks`)
3. **Write Tests** (using testify suite + mocks)
4. **Run Tests** (`make test`)
5. **Check Coverage** (`make test-coverage`)

## Configuration

### Test Flags

Tests run with recommended flags:
- `-v` - Verbose output
- `-race` - Race condition detection
- `-coverprofile=coverage.out` - Coverage tracking

### Short Flag

Use `-short` for unit tests:
```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    // Integration test code
}
```

Run with: `make test-unit` (includes `-short`)

## Best Practices

### 1. Test Organization
✅ **DO:**
- Use suite pattern for tests needing setup/teardown
- Group related tests in same file
- Use table-driven tests for multiple scenarios

❌ **DON'T:**
- Mix unit and integration tests without skip conditions
- Hardcode test data
- Test implementation details

### 2. Mocking
✅ **DO:**
- Mock external dependencies (DB, API, filesystem)
- Use interfaces for mockable components
- Set clear expectations with EXPECT()

❌ **DON'T:**
- Mock everything (don't mock value objects)
- Create mocks manually
- Forget to call ctrl.Finish()

### 3. Assertions
✅ **DO:**
- Use testify assertions for readability
- Use require for fatal errors
- Use assert for non-fatal checks

❌ **DON'T:**
- Use plain if statements
- Continue testing after fatal errors
- Mix assertion styles

### 4. Coverage
✅ **DO:**
- Aim for high coverage (>80%)
- Focus on critical paths
- Check coverage reports regularly

❌ **DON'T:**
- Chase 100% coverage at all costs
- Test trivial getters/setters
- Ignore untested error paths

## Integration with CI/CD

### GitHub Actions Example

```yaml
- name: Run tests
  run: make test

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```

### GitLab CI Example

```yaml
test:
  script:
    - make test
  coverage: '/coverage: \d+\.\d+% of statements/'
  artifacts:
    paths:
      - coverage.out
      - coverage.html
```

## Troubleshooting

### Mock Generation Fails

**Issue:** `mockgen: command not found`

**Solution:**
```bash
make install-tools
```

### Tests Hang

**Issue:** Tests waiting for timeout

**Solution:** Check for:
- Missing `defer ctrl.Finish()`
- Incorrect mock expectations
- Long-running operations without timeout

### Race Condition Detected

**Issue:** `-race` flag reports data race

**Solution:**
- Fix concurrent access to shared variables
- Use sync.Mutex or channels
- Avoid shared mutable state

## Future Enhancements

Possible additions:
- Benchmark test examples
- Integration test helpers
- Test fixtures/factories
- HTTP client testing utilities
- Database seeding helpers

## Resources

- [testify Documentation](https://github.com/stretchr/testify)
- [uber-go/mock Documentation](https://github.com/uber-go/mock)
- [Go Testing Best Practices](https://golang.org/doc/effective_go#testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)

## Summary

Every generated project now includes:
✅ Complete testing setup  
✅ Working test examples  
✅ Mock generation capability  
✅ Comprehensive documentation  
✅ CI/CD ready configuration  

**No additional setup required - just run `make test`!** 🎉
