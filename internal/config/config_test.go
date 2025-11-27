package config

import (
	"strings"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				ProjectName:  "my-project",
				ModulePath:   "github.com/user/my-project",
				GoVersion:    "1.23",
				Framework:    "chi",
				Logger:       "slog",
				ConfigFormat: "env",
			},
			wantErr: false,
		},
		{
			name: "empty project name",
			config: Config{
				ProjectName: "",
				ModulePath:  "github.com/user/my-project",
				GoVersion:   "1.23",
			},
			wantErr: true,
			errMsg:  "project name is required",
		},
		{
			name: "invalid project name with uppercase",
			config: Config{
				ProjectName: "MyProject",
				ModulePath:  "github.com/user/my-project",
				GoVersion:   "1.23",
			},
			wantErr: true,
			errMsg:  "project name must start with lowercase",
		},
		{
			name: "empty module path",
			config: Config{
				ProjectName: "my-project",
				ModulePath:  "",
				GoVersion:   "1.23",
			},
			wantErr: true,
			errMsg:  "module path is required",
		},
		{
			name: "invalid module path",
			config: Config{
				ProjectName: "my-project",
				ModulePath:  "not-a-valid-path",
				GoVersion:   "1.23",
			},
			wantErr: true,
			errMsg:  "module path must be a valid Go module path",
		},
		{
			name: "empty go version",
			config: Config{
				ProjectName: "my-project",
				ModulePath:  "github.com/user/my-project",
				GoVersion:   "",
			},
			wantErr: true,
			errMsg:  "Go version is required",
		},
		{
			name: "invalid go version",
			config: Config{
				ProjectName: "my-project",
				ModulePath:  "github.com/user/my-project",
				GoVersion:   "1.18",
			},
			wantErr: true,
			errMsg:  "Go version must be one of",
		},
		{
			name: "invalid framework",
			config: Config{
				ProjectName: "my-project",
				ModulePath:  "github.com/user/my-project",
				GoVersion:   "1.23",
				Framework:   "invalid",
			},
			wantErr: true,
			errMsg:  "framework must be one of",
		},
		{
			name: "invalid logger",
			config: Config{
				ProjectName: "my-project",
				ModulePath:  "github.com/user/my-project",
				GoVersion:   "1.23",
				Logger:      "invalid",
			},
			wantErr: true,
			errMsg:  "logger must be one of",
		},
		{
			name: "project name with underscore is valid",
			config: Config{
				ProjectName: "my_project",
				ModulePath:  "github.com/user/my_project",
				GoVersion:   "1.23",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestConfig_HasDatabase(t *testing.T) {
	tests := []struct {
		name      string
		databases []string
		check     string
		want      bool
	}{
		{
			name:      "has postgres",
			databases: []string{"postgres", "redis"},
			check:     "postgres",
			want:      true,
		},
		{
			name:      "does not have mysql",
			databases: []string{"postgres", "redis"},
			check:     "mysql",
			want:      false,
		},
		{
			name:      "empty databases",
			databases: []string{},
			check:     "postgres",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Databases: tt.databases}
			if got := c.HasDatabase(tt.check); got != tt.want {
				t.Errorf("Config.HasDatabase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_NeedsCache(t *testing.T) {
	tests := []struct {
		name      string
		databases []string
		want      bool
	}{
		{
			name:      "needs cache with redis",
			databases: []string{"postgres", "redis"},
			want:      true,
		},
		{
			name:      "no cache without redis",
			databases: []string{"postgres", "mysql"},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Databases: tt.databases}
			if got := c.NeedsCache(); got != tt.want {
				t.Errorf("Config.NeedsCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_NeedsSQL(t *testing.T) {
	tests := []struct {
		name      string
		databases []string
		want      bool
	}{
		{
			name:      "needs SQL with postgres",
			databases: []string{"postgres"},
			want:      true,
		},
		{
			name:      "needs SQL with mysql",
			databases: []string{"mysql"},
			want:      true,
		},
		{
			name:      "no SQL with only mongodb",
			databases: []string{"mongodb"},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Databases: tt.databases}
			if got := c.NeedsSQL(); got != tt.want {
				t.Errorf("Config.NeedsSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_NeedsNoSQL(t *testing.T) {
	tests := []struct {
		name      string
		databases []string
		want      bool
	}{
		{
			name:      "needs NoSQL with mongodb",
			databases: []string{"mongodb"},
			want:      true,
		},
		{
			name:      "no NoSQL with only postgres",
			databases: []string{"postgres"},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Databases: tt.databases}
			if got := c.NeedsNoSQL(); got != tt.want {
				t.Errorf("Config.NeedsNoSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
