package generator

import (
"strings"
"testing"
)

func TestExecuteTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     any
		want     string
		wantErr  bool
	}{
		{
			name:     "simple substitution",
			template: "Hello, {{.Name}}!",
			data:     struct{ Name string }{"World"},
			want:     "Hello, World!",
		},
		{
			name:     "conditional true",
			template: "{{if .Enabled}}enabled{{else}}disabled{{end}}",
			data:     struct{ Enabled bool }{true},
			want:     "enabled",
		},
		{
			name:     "conditional false",
			template: "{{if .Enabled}}enabled{{else}}disabled{{end}}",
			data:     struct{ Enabled bool }{false},
			want:     "disabled",
		},
		{
			name:     "range over slice",
			template: "{{range .Items}}{{.}} {{end}}",
			data:     struct{ Items []string }{[]string{"a", "b", "c"}},
			want:     "a b c ",
		},
		{
			name:     "join function",
			template: `{{join .Items ", "}}`,
			data:     struct{ Items []string }{[]string{"a", "b", "c"}},
			want:     "a, b, c",
		},
		{
			name:     "indent function",
			template: `{{indent 4 "line1\nline2"}}`,
			data:     nil,
			want:     "    line1\n    line2",
		},
		{
			name:     "nested struct",
			template: "{{.Outer.Inner}}",
			data: struct{ Outer struct{ Inner string } }{
				Outer: struct{ Inner string }{"value"},
			},
			want: "value",
		},
		{
			name:     "invalid template syntax",
			template: "{{.Name",
			data:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
got, err := executeTemplate(tt.name, tt.template, tt.data)
if (err != nil) != tt.wantErr {
t.Errorf("executeTemplate() error = %v, wantErr %v", err, tt.wantErr)
return
}
if !tt.wantErr && got != tt.want {
				t.Errorf("executeTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTemplateData(t *testing.T) {
	cfg := createTestConfig()
	cfg.Databases = []string{"postgres", "redis"}
	cfg.EnableTracing = true
	cfg.EnableMetrics = true

	gen := New(cfg, "/tmp")
	data := gen.NewTemplateData()

	if data.ProjectName != "test-project" {
		t.Errorf("ProjectName = %q, want %q", data.ProjectName, "test-project")
	}
	if !data.HasPostgres {
		t.Error("expected HasPostgres to be true")
	}
	if !data.HasRedis {
		t.Error("expected HasRedis to be true")
	}
	if data.HasMySQL {
		t.Error("expected HasMySQL to be false")
	}
	if !data.EnableTracing {
		t.Error("expected EnableTracing to be true")
	}
	if !data.NeedsSQL {
		t.Error("expected NeedsSQL to be true")
	}
	if !data.NeedsCache {
		t.Error("expected NeedsCache to be true")
	}
}

func TestLoggerType(t *testing.T) {
	tests := []struct {
		logger   string
		wantType string
	}{
		{"slog", "*slog.Logger"},
		{"zap", "*zap.Logger"},
		{"zerolog", "*zerolog.Logger"},
		{"", "*slog.Logger"},
	}

	for _, tt := range tests {
		t.Run(tt.logger, func(t *testing.T) {
cfg := createTestConfig()
			cfg.Logger = tt.logger
			gen := New(cfg, "/tmp")
			got := gen.getLoggerType()
			if got != tt.wantType {
				t.Errorf("getLoggerType() = %q, want %q", got, tt.wantType)
			}
		})
	}
}

func TestLoggerImport(t *testing.T) {
	tests := []struct {
		logger     string
		wantImport string
	}{
		{"slog", `"log/slog"`},
		{"zap", `"go.uber.org/zap"`},
		{"zerolog", `"github.com/rs/zerolog"`},
		{"", `"log/slog"`},
	}

	for _, tt := range tests {
		t.Run(tt.logger, func(t *testing.T) {
cfg := createTestConfig()
			cfg.Logger = tt.logger
			gen := New(cfg, "/tmp")
			got := gen.getLoggerImport()
			if got != tt.wantImport {
				t.Errorf("getLoggerImport() = %q, want %q", got, tt.wantImport)
			}
		})
	}
}

func TestWriteTemplate(t *testing.T) {
	cfg := createTestConfig()
	mfs := createMemoryFS()
	gen := New(cfg, "/output", WithFileSystem(mfs))

	tmpl := `package main

import "fmt"

func main() {
	fmt.Println("Hello, {{.ProjectName}}!")
}
`
	err := gen.writeTemplate("cmd/main.go", "main", tmpl, gen.NewTemplateData())
	if err != nil {
		t.Fatalf("writeTemplate() error = %v", err)
	}

	content := mfs.FileContent("/output/test-project/cmd/main.go")
	if !strings.Contains(content, `fmt.Println("Hello, test-project!")`) {
		t.Errorf("template not rendered correctly:\n%s", content)
	}
}
