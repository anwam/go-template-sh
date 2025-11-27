package prompt

import (
	"testing"
)

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid lowercase name",
			input:   "my-project",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "project123",
			wantErr: false,
		},
		{
			name:    "valid with underscore",
			input:   "my_project",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "project name cannot be empty",
		},
		{
			name:    "uppercase letters",
			input:   "MyProject",
			wantErr: true,
			errMsg:  "lowercase letters",
		},
		{
			name:    "special characters",
			input:   "my@project",
			wantErr: true,
			errMsg:  "lowercase letters",
		},
		{
			name:    "invalid type - int",
			input:   123,
			wantErr: true,
			errMsg:  "invalid type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProjectName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProjectName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("validateProjectName() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestParseFramework(t *testing.T) {
	tests := []struct {
		name     string
		choice   string
		expected string
	}{
		{"net/http", "net/http (standard library)", "stdlib"},
		{"Chi", "Chi", "chi"},
		{"Gin", "Gin", "gin"},
		{"Echo", "Echo", "echo"},
		{"Fiber", "Fiber", "fiber"},
		{"unknown", "unknown", "stdlib"},
		{"empty", "", "stdlib"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFramework(tt.choice)
			if result != tt.expected {
				t.Errorf("parseFramework(%q) = %q, want %q", tt.choice, result, tt.expected)
			}
		})
	}
}

func TestParseDatabases(t *testing.T) {
	tests := []struct {
		name     string
		choices  []string
		expected []string
	}{
		{"postgres", []string{"PostgreSQL"}, []string{"postgres"}},
		{"mysql", []string{"MySQL"}, []string{"mysql"}},
		{"mongodb", []string{"MongoDB"}, []string{"mongodb"}},
		{"redis", []string{"Redis (cache)"}, []string{"redis"}},
		{"multiple", []string{"PostgreSQL", "Redis (cache)"}, []string{"postgres", "redis"}},
		{"empty", []string{}, nil},
		{"none", []string{"None"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDatabases(tt.choices)
			if !slicesEqual(result, tt.expected) {
				t.Errorf("parseDatabases(%v) = %v, want %v", tt.choices, result, tt.expected)
			}
		})
	}
}

func TestParseLogger(t *testing.T) {
	tests := []struct {
		name     string
		choice   string
		expected string
	}{
		{"slog", "slog (standard library)", "slog"},
		{"Zap", "Zap", "zap"},
		{"Zerolog", "Zerolog", "zerolog"},
		{"unknown", "unknown", "slog"},
		{"empty", "", "slog"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogger(tt.choice)
			if result != tt.expected {
				t.Errorf("parseLogger(%q) = %q, want %q", tt.choice, result, tt.expected)
			}
		})
	}
}

func TestParseCI(t *testing.T) {
	tests := []struct {
		name     string
		choice   string
		expected string
	}{
		{"GitHub Actions", "GitHub Actions", "github"},
		{"GitLab CI", "GitLab CI", "gitlab"},
		{"None", "None", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCI(tt.choice)
			if result != tt.expected {
				t.Errorf("parseCI(%q) = %q, want %q", tt.choice, result, tt.expected)
			}
		})
	}
}

func TestParseConfigFormat(t *testing.T) {
	tests := []struct {
		name     string
		choice   string
		expected string
	}{
		{"YAML", "YAML (config.yaml)", "yaml"},
		{"JSON", "JSON (config.json)", "json"},
		{"TOML", "TOML (config.toml)", "toml"},
		{"env", "Environment variables (.env)", "env"},
		{"empty", "", "env"},
		{"unknown", "Unknown format", "env"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConfigFormat(tt.choice)
			if result != tt.expected {
				t.Errorf("parseConfigFormat(%q) = %q, want %q", tt.choice, result, tt.expected)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
