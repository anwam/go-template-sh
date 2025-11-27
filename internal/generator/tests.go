package generator

import (
	"fmt"
	"strings"
)

func (g *Generator) generateTestFiles() error {
	if err := g.generateHandlerTests(); err != nil {
		return err
	}

	// Generate mock interfaces if database is enabled
	if g.config.NeedsSQL() || g.config.NeedsNoSQL() {
		if err := g.generateMockInterfaces(); err != nil {
			return err
		}
	}

	if err := g.generateTestSuiteExample(); err != nil {
		return err
	}

	return g.generateTestingReadme()
}

func (g *Generator) generateHandlerTests() error {
	content := g.getHandlerTestContent()
	return g.writeFile("internal/handlers/handlers_test.go", content)
}

func (g *Generator) getHandlerTestContent() string {
	return fmt.Sprintf(`package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"%[1]s/internal/config"
	"%[1]s/internal/observability"
%[2]s
)

type HandlerTestSuite struct {
	suite.Suite
	handler *Handler
	config  *config.Config
	obs     *observability.Observability
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.config = &config.Config{
%[3]s
	}
	
	// Initialize observability (in test mode)
	suite.obs = &observability.Observability{
		Logger: observability.NewLogger(suite.config),
	}
	
	suite.handler = NewHandler(suite.config, suite.obs)
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (suite *HandlerTestSuite) TestHealth() {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	
	suite.handler.Health(w, req)
	
	suite.Equal(http.StatusOK, w.Code)
	suite.Equal("application/json", w.Header().Get("Content-Type"))
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("ok", response.Status)
	suite.Equal("Service is healthy", response.Message)
}

func (suite *HandlerTestSuite) TestReady() {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	
	suite.handler.Ready(w, req)
	
	suite.Equal(http.StatusOK, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("ready", response.Status)
}

func (suite *HandlerTestSuite) TestIndex() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	
	suite.handler.Index(w, req)
	
	suite.Equal(http.StatusOK, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("ok", response.Status)
	suite.NotNil(response.Data)
	suite.Equal("1.0.0", response.Data["version"])
	suite.Equal(%[4]s, response.Data["environment"])
}

%[5]s

// Table-driven test example
func TestResponseJSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		wantErr  bool
	}{
		{
			name: "valid response with data",
			response: Response{
				Status:  "ok",
				Message: "test",
				Data:    map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "response without data",
			response: Response{
				Status:  "ok",
				Message: "test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, data)
				
				var decoded Response
				err = json.Unmarshal(data, &decoded)
				require.NoError(t, err)
				assert.Equal(t, tt.response.Status, decoded.Status)
			}
		})
	}
}
`,
		g.config.ModulePath,
		g.getTestImports(),
		g.getTestConfigFields(),
		g.getConfigFieldReference("Environment"),
		g.getFrameworkSpecificTests(),
	)
}

func (g *Generator) getTestImports() string {
	switch g.config.Framework {
	case "gin":
		return `	"github.com/gin-gonic/gin"`
	case "echo":
		return `	"github.com/labstack/echo/v4"`
	case "fiber":
		return `	"github.com/gofiber/fiber/v2"`
	default:
		return ""
	}
}

func (g *Generator) getTestConfigFields() string {
	if g.config.ConfigFormat == "" || g.config.ConfigFormat == "env" {
		return `		Environment: "test",
		Port:        "8080",`
	}

	return `		App: config.AppConfig{
			Environment: "test",
			Port:        8080,
		},`
}

func (g *Generator) getFrameworkSpecificTests() string {
	switch g.config.Framework {
	case "gin":
		return `
func (suite *HandlerTestSuite) TestHealthGin() {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	suite.handler.HealthGin(c)
	
	suite.Equal(http.StatusOK, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("ok", response.Status)
}`
	case "echo":
		return `
func (suite *HandlerTestSuite) TestHealthEcho() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	err := suite.handler.HealthEcho(c)
	suite.NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
}`
	case "fiber":
		return `
func (suite *HandlerTestSuite) TestHealthFiber() {
	app := fiber.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	
	resp, err := app.Test(req)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
}`
	default:
		return ""
	}
}

func (g *Generator) generateMockInterfaces() error {
	content := g.getMockInterfacesContent()
	if content == "" {
		fmt.Println("[go-template-sh] Skipping mock interface generation: no database configured, so no mock interfaces needed.")
		return nil
	}
	return g.writeFile("internal/mocks/interfaces.go", content)
}

func (g *Generator) getMockInterfacesContent() string {
	var interfaces []string

	if g.config.HasDatabase("postgres") {
		interfaces = append(interfaces, `
// DatabaseInterface defines methods for database operations
type DatabaseInterface interface {
	Ping(ctx context.Context) error
	Close() error
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	Exec(ctx context.Context, query string, args ...interface{}) (CommandTag, error)
}

type Row interface {
	Scan(dest ...interface{}) error
}

type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close()
	Err() error
}

type CommandTag interface {
	RowsAffected() int64
}
`)
	}

	if g.config.NeedsCache() {
		interfaces = append(interfaces, `
// CacheInterface defines methods for cache operations
type CacheInterface interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
}
`)
	}

	if len(interfaces) == 0 {
		return ""
	}

	return fmt.Sprintf(`package mocks

import (
	"context"
	"time"
)

// This file contains interface definitions for generating mocks
// Run 'make generate-mocks' to generate mock implementations

%s

//go:generate mockgen -source=interfaces.go -destination=mocks.go -package=mocks
`, strings.Join(interfaces, "\n"))
}

func (g *Generator) generateTestSuiteExample() error {
	if !g.config.NeedsSQL() && !g.config.NeedsNoSQL() {
		return nil
	}
	content := g.getTestSuiteExampleContent()
	return g.writeFile("internal/database/database_test.go", content)
}

func (g *Generator) getTestSuiteExampleContent() string {
	var mockSetup string
	var mockImport string

	if g.config.HasDatabase("postgres") {
		mockImport = fmt.Sprintf(`	"go.uber.org/mock/gomock"
	"%s/internal/mocks"`, g.config.ModulePath)
		mockSetup = `	ctrl := gomock.NewController(suite.T())
	suite.mockDB = mocks.NewMockDatabaseInterface(ctrl)
	
	// Setup mock expectations
	suite.mockDB.EXPECT().Ping(gomock.Any()).Return(nil).AnyTimes()`
	}

	return fmt.Sprintf(`package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
%s
)

type DatabaseTestSuite struct {
	suite.Suite
	mockDB  *mocks.MockDatabaseInterface
}

func (suite *DatabaseTestSuite) SetupTest() {
%s
}

func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}

func (suite *DatabaseTestSuite) TestDatabaseConnection() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := suite.mockDB.Ping(ctx)
	suite.NoError(err)
}

// Example of testing with mock
func (suite *DatabaseTestSuite) TestQueryExecution() {
	// This is an example - implement based on your actual database operations
	suite.T().Skip("Implement based on your database operations")
}
`, mockImport, mockSetup)
}

func (g *Generator) generateTestingReadme() error {
	content := g.getTestingReadmeContent()
	return g.writeFile("docs/TESTING.md", content)
}

func (g *Generator) getTestingReadmeContent() string {
	return fmt.Sprintf(`# Testing Guide

This project uses industry-standard testing tools and practices.

## Testing Stack

- **[testify](https://github.com/stretchr/testify)** - Assertion library and test suite
- **[uber-go/mock](https://github.com/uber-go/mock)** - Mock generation framework

## Running Tests

### All Tests
%[1]s

### Unit Tests Only
%[2]s

### Integration Tests Only
%[3]s

### With Coverage Report
%[4]s
This generates `+"`coverage.html`"+` that you can open in a browser.

## Generating Mocks

When you modify interface definitions in `+"`internal/mocks/interfaces.go`"+`:

%[5]s

## Test Structure

### Handler Tests
Example in `+"`internal/handlers/handlers_test.go`"+`:
- Uses testify suite for setup/teardown
- Tests all HTTP handlers
- Demonstrates table-driven tests

### Database Tests
Example in `+"`internal/database/database_test.go`"+`:
- Uses uber-go/mock for database mocking
- Demonstrates context usage
- Shows async operation testing

## Writing Tests

### Using Testify Suite

%[6]s

### Using Mocks

1. Define interface in `+"`internal/mocks/interfaces.go`"+`
2. Run `+"`make generate-mocks`"+`
3. Use in tests:

%[7]s

### Table-Driven Tests

%[8]s

## Best Practices

1. **Use testify assertions**: More readable than plain Go comparisons
2. **Use suite.Suite**: For tests that need setup/teardown
3. **Use table-driven tests**: For testing multiple scenarios
4. **Mock external dependencies**: Use uber-go/mock for interfaces
5. **Test behavior, not implementation**: Focus on what, not how
6. **Keep tests fast**: Mock slow operations (DB, API calls)
7. **Use `+"`-short`"+` flag**: To skip integration tests during development

## Continuous Integration

Tests run automatically on CI. Make sure:
- All tests pass: `+"`make test`"+`
- No race conditions: Tests run with `+"`-race`"+` flag
- Coverage is tracked: `+"`coverage.out`"+` is generated

## Examples

Check these files for examples:
- `+"`internal/handlers/handlers_test.go`"+` - Handler testing with suite
- `+"`internal/database/database_test.go`"+` - Mock usage examples
- `+"`internal/mocks/interfaces.go`"+` - Interface definitions for mocks
`,
		codeBlock("bash", "make test"),
		codeBlock("bash", "make test-unit"),
		codeBlock("bash", "make test-integration"),
		codeBlock("bash", "make test-coverage"),
		codeBlock("bash", "make generate-mocks"),
		codeBlock("go", `type MyTestSuite struct {
    suite.Suite
    // Add fields needed for tests
}

func (suite *MyTestSuite) SetupTest() {
    // Runs before each test
}

func (suite *MyTestSuite) TearDownTest() {
    // Runs after each test
}

func (suite *MyTestSuite) TestSomething() {
    suite.Equal(expected, actual)
    suite.NoError(err)
}

func TestMySuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}`),
		codeBlock("go", `ctrl := gomock.NewController(t)
defer ctrl.Finish()

mockDB := mocks.NewMockDatabaseInterface(ctrl)
mockDB.EXPECT().
    Query(gomock.Any(), "SELECT *").
    Return(mockRows, nil)

// Use mockDB in your test`),
		codeBlock("go", `func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"case 1", "input1", "output1", false},
        {"case 2", "input2", "output2", false},
        {"error case", "bad", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}`),
	)
}

func codeBlock(lang, code string) string {
	return fmt.Sprintf("```%s\n%s\n```", lang, strings.TrimSpace(code))
}
