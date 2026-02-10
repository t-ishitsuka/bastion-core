package cmd

import (
	"testing"
)

func TestParseGoVersion(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		expectMajor   int
		expectMinor   int
		expectSuccess bool
	}{
		{
			name:          "正常なバージョン",
			version:       "go version go1.22.0 linux/amd64",
			expectMajor:   1,
			expectMinor:   22,
			expectSuccess: true,
		},
		{
			name:          "古いバージョン",
			version:       "go version go1.20.5 linux/amd64",
			expectMajor:   1,
			expectMinor:   20,
			expectSuccess: true,
		},
		{
			name:          "新しいバージョン",
			version:       "go version go1.23.1 darwin/arm64",
			expectMajor:   1,
			expectMinor:   23,
			expectSuccess: true,
		},
		{
			name:          "不正なフォーマット",
			version:       "invalid version string",
			expectMajor:   0,
			expectMinor:   0,
			expectSuccess: false,
		},
		{
			name:          "空文字列",
			version:       "",
			expectMajor:   0,
			expectMinor:   0,
			expectSuccess: false,
		},
		{
			name:          "バージョン番号なし",
			version:       "go version",
			expectMajor:   0,
			expectMinor:   0,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, ok := parseGoVersion(tt.version)

			if ok != tt.expectSuccess {
				t.Errorf("parseGoVersion(%q) success = %v, want %v", tt.version, ok, tt.expectSuccess)
			}

			if major != tt.expectMajor {
				t.Errorf("parseGoVersion(%q) major = %d, want %d", tt.version, major, tt.expectMajor)
			}

			if minor != tt.expectMinor {
				t.Errorf("parseGoVersion(%q) minor = %d, want %d", tt.version, minor, tt.expectMinor)
			}
		})
	}
}

func TestCheckCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		wantOK  bool
	}{
		{
			name:    "存在するコマンド（go）",
			command: "go",
			args:    []string{"version"},
			wantOK:  true,
		},
		{
			name:    "存在しないコマンド",
			command: "nonexistent-command-xyz",
			args:    []string{},
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok := checkCommand(tt.command, tt.args...)
			if ok != tt.wantOK {
				t.Errorf("checkCommand(%q, %v) = %v, want %v", tt.command, tt.args, ok, tt.wantOK)
			}
		})
	}
}

func TestGetCommandOutput(t *testing.T) {
	tests := []struct {
		name       string
		command    string
		args       []string
		wantEmpty  bool
		wantPrefix string
	}{
		{
			name:       "go version の出力",
			command:    "go",
			args:       []string{"version"},
			wantEmpty:  false,
			wantPrefix: "go version",
		},
		{
			name:      "存在しないコマンド",
			command:   "nonexistent-command-xyz",
			args:      []string{},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := getCommandOutput(tt.command, tt.args...)

			if tt.wantEmpty {
				if output != "" {
					t.Errorf("getCommandOutput(%q, %v) = %q, want empty string", tt.command, tt.args, output)
				}
			} else {
				if output == "" {
					t.Errorf("getCommandOutput(%q, %v) returned empty, want non-empty", tt.command, tt.args)
				}
				if tt.wantPrefix != "" && len(output) >= len(tt.wantPrefix) {
					prefix := output[:len(tt.wantPrefix)]
					if prefix != tt.wantPrefix {
						t.Errorf("getCommandOutput(%q, %v) prefix = %q, want %q", tt.command, tt.args, prefix, tt.wantPrefix)
					}
				}
			}
		})
	}
}
