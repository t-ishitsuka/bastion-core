package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunInit(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// 一時ディレクトリに移動
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("現在のディレクトリの取得に失敗: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("一時ディレクトリへの移動に失敗: %v", err)
	}

	// init コマンドを実行
	if err := runInit(nil, nil); err != nil {
		t.Fatalf("runInit() failed: %v", err)
	}

	// agents/ ディレクトリが作成されたか確認
	agentsDir := filepath.Join(tmpDir, "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Error("agents/ ディレクトリが作成されていません")
	}

	// queue/ ディレクトリ構造が作成されたか確認
	queueDirs := []string{
		filepath.Join(tmpDir, "queue"),
		filepath.Join(tmpDir, "queue", "inbox"),
		filepath.Join(tmpDir, "queue", "tasks"),
		filepath.Join(tmpDir, "queue", "reports"),
	}

	for _, dir := range queueDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("ディレクトリが作成されていません: %s", dir)
		}
	}

	// .gitignore が作成されたか確認
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error(".gitignore が作成されていません")
	} else {
		content, err := os.ReadFile(gitignorePath)
		if err != nil {
			t.Errorf(".gitignore の読み取りに失敗: %v", err)
		}
		if !contains(string(content), "queue/") {
			t.Error(".gitignore に queue/ が追加されていません")
		}
	}
}

func TestRunInitNoClobber(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// 一時ディレクトリに移動
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("現在のディレクトリの取得に失敗: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("一時ディレクトリへの移動に失敗: %v", err)
	}

	// agents/ ディレクトリに既存ファイルを作成
	agentsDir := filepath.Join(tmpDir, "agents")
	envoyDir := filepath.Join(agentsDir, "envoy")
	if err := os.MkdirAll(envoyDir, 0755); err != nil {
		t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
	}

	existingFile := filepath.Join(envoyDir, "CLAUDE.md")
	existingContent := "# Custom content"
	if err := os.WriteFile(existingFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("テスト用ファイルの作成に失敗: %v", err)
	}

	// init コマンドを実行
	if err := runInit(nil, nil); err != nil {
		t.Fatalf("runInit() failed: %v", err)
	}

	// 既存ファイルが上書きされていないか確認
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("ファイルの読み取りに失敗: %v", err)
	}

	if string(content) != existingContent {
		t.Error("既存ファイルが上書きされました（no-clobber モードが機能していません）")
	}
}

func TestUpdateGitignore(t *testing.T) {
	tests := []struct {
		name            string
		existingContent string
		wantContains    string
	}{
		{
			name:            ".gitignore が存在しない場合",
			existingContent: "",
			wantContains:    "queue/",
		},
		{
			name:            ".gitignore が存在する場合",
			existingContent: "node_modules/\n",
			wantContains:    "queue/",
		},
		{
			name:            "既に queue/ が記載されている場合",
			existingContent: "queue/\nnode_modules/\n",
			wantContains:    "queue/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// 既存の .gitignore を作成
			if tt.existingContent != "" {
				gitignorePath := filepath.Join(tmpDir, ".gitignore")
				if err := os.WriteFile(gitignorePath, []byte(tt.existingContent), 0644); err != nil {
					t.Fatalf("テスト用 .gitignore の作成に失敗: %v", err)
				}
			}

			// updateGitignore を実行
			if err := updateGitignore(tmpDir); err != nil {
				t.Fatalf("updateGitignore() failed: %v", err)
			}

			// .gitignore の内容を確認
			gitignorePath := filepath.Join(tmpDir, ".gitignore")
			content, err := os.ReadFile(gitignorePath)
			if err != nil {
				t.Fatalf(".gitignore の読み取りに失敗: %v", err)
			}

			if !contains(string(content), tt.wantContains) {
				t.Errorf(".gitignore に %s が含まれていません", tt.wantContains)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}
