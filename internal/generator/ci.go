package generator

import "fmt"

func (g *Generator) generateCIFiles() error {
	switch g.config.CI {
	case "github":
		return g.generateGitHubActions()
	case "gitlab":
		return g.generateGitLabCI()
	default:
		return nil
	}
}

func (g *Generator) generateGitHubActions() error {
	content := fmt.Sprintf(`name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    
    services:%s
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '%s'

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

  lint:
    name: Lint
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '%s'

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [test, lint]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '%s'

      - name: Build
        run: go build -v ./cmd/%s
`, g.getGitHubServicesConfig(), g.config.GoVersion, g.config.GoVersion, g.config.GoVersion, g.config.ProjectName)

	return g.writeFile(".github/workflows/ci.yml", content)
}

func (g *Generator) getGitHubServicesConfig() string {
	if !g.config.HasDatabase("postgres") && !g.config.HasDatabase("mysql") && !g.config.HasDatabase("mongodb") && !g.config.HasDatabase("redis") {
		return ""
	}

	services := []string{}

	if g.config.HasDatabase("postgres") {
		services = append(services, `      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: user
          POSTGRES_PASSWORD: password
          POSTGRES_DB: testdb
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5`)
	}

	if g.config.HasDatabase("mysql") {
		services = append(services, `      mysql:
        image: mysql:8
        env:
          MYSQL_ROOT_PASSWORD: password
          MYSQL_DATABASE: testdb
        ports:
          - 3306:3306
        options: >-
          --health-cmd="mysqladmin ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=3`)
	}

	if g.config.HasDatabase("mongodb") {
		services = append(services, `      mongodb:
        image: mongo:7
        ports:
          - 27017:27017`)
	}

	if g.config.HasDatabase("redis") {
		services = append(services, `      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5`)
	}

	if len(services) > 0 {
		return "\n" + joinWithIndent(services, "\n", "      ")
	}

	return ""
}

func (g *Generator) generateGitLabCI() error {
	content := fmt.Sprintf(`stages:
  - test
  - build

variables:
  GO_VERSION: "%s"

test:
  stage: test
  image: golang:${GO_VERSION}
  
  services:%s
  
  before_script:
    - go mod download
  
  script:
    - go test -v -race -coverprofile=coverage.out ./...
    - go tool cover -func=coverage.out
  
  coverage: '/total:\s+\(statements\)\s+(\d+\.\d+%%)/)'
  
  artifacts:
    paths:
      - coverage.out
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.out

lint:
  stage: test
  image: golangci/golangci-lint:latest
  
  script:
    - golangci-lint run

build:
  stage: build
  image: golang:${GO_VERSION}
  
  dependencies:
    - test
  
  before_script:
    - go mod download
  
  script:
    - go build -v ./cmd/%s
  
  artifacts:
    paths:
      - %s
`, g.config.GoVersion, g.getGitLabServicesConfig(), g.config.ProjectName, g.config.ProjectName)

	return g.writeFile(".gitlab-ci.yml", content)
}

func (g *Generator) getGitLabServicesConfig() string {
	if !g.config.HasDatabase("postgres") && !g.config.HasDatabase("mysql") && !g.config.HasDatabase("mongodb") && !g.config.HasDatabase("redis") {
		return ""
	}

	services := []string{}

	if g.config.HasDatabase("postgres") {
		services = append(services, "    - postgres:16-alpine")
	}
	if g.config.HasDatabase("mysql") {
		services = append(services, "    - mysql:8")
	}
	if g.config.HasDatabase("mongodb") {
		services = append(services, "    - mongo:7")
	}
	if g.config.HasDatabase("redis") {
		services = append(services, "    - redis:7-alpine")
	}

	if len(services) > 0 {
		return "\n" + joinWithIndent(services, "\n", "")
	}

	return ""
}

func joinWithIndent(items []string, sep, indent string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += sep
		}
		result += indent + item
	}
	return result
}
