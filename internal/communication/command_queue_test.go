package communication

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCommandQueueManager_Write(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewCommandQueueManager(tmpDir)

	// 指令を書き込む
	cmd := Command{
		ID:        "cmd_001",
		Timestamp: time.Now(),
		Purpose:   "JWT認証を実装",
		AcceptanceCriteria: []string{
			"POST /auth/login が JWT を返す",
			"テストがパスする",
		},
		Command:  "JWT認証を実装する",
		Project:  "api-server",
		Priority: "high",
		Status:   CommandStatusPending,
	}

	err := manager.Write(cmd)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// ファイルが作成されたことを確認
	commandPath := filepath.Join(tmpDir, "envoy_to_marshall.yaml")
	if _, err := os.Stat(commandPath); os.IsNotExist(err) {
		t.Fatalf("command file was not created: %s", commandPath)
	}

	// 指令を読み込んで確認
	commands, err := manager.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}

	readCmd := commands[0]
	if readCmd.ID != "cmd_001" {
		t.Errorf("expected ID 'cmd_001', got '%s'", readCmd.ID)
	}
	if readCmd.Purpose != "JWT認証を実装" {
		t.Errorf("expected purpose 'JWT認証を実装', got '%s'", readCmd.Purpose)
	}
	if len(readCmd.AcceptanceCriteria) != 2 {
		t.Errorf("expected 2 acceptance criteria, got %d", len(readCmd.AcceptanceCriteria))
	}
}

func TestCommandQueueManager_MultipleWrites(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewCommandQueueManager(tmpDir)

	// 複数の指令を書き込む
	cmd1 := Command{
		ID:        "cmd_001",
		Timestamp: time.Now(),
		Purpose:   "認証機能",
		Command:   "認証を実装",
		Status:    CommandStatusPending,
	}

	cmd2 := Command{
		ID:        "cmd_002",
		Timestamp: time.Now(),
		Purpose:   "API実装",
		Command:   "APIを実装",
		Status:    CommandStatusPending,
	}

	_ = manager.Write(cmd1)
	_ = manager.Write(cmd2)

	// 指令を読み込んで確認
	commands, err := manager.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(commands))
	}
}

func TestCommandQueueManager_ReadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewCommandQueueManager(tmpDir)

	// 存在しないキューを読み込む
	commands, err := manager.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(commands) != 0 {
		t.Fatalf("expected 0 commands, got %d", len(commands))
	}
}

func TestCommandQueueManager_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewCommandQueueManager(tmpDir)

	// 並行して書き込みを行う
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(n int) {
			cmd := Command{
				ID:        "cmd_concurrent",
				Timestamp: time.Now(),
				Purpose:   "並行書き込みテスト",
				Command:   "テスト",
				Status:    CommandStatusPending,
			}
			_ = manager.Write(cmd)
			done <- true
		}(i)
	}

	// すべての goroutine が完了するのを待つ
	for i := 0; i < 5; i++ {
		<-done
	}

	// 指令を読み込んで確認
	commands, err := manager.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(commands) != 5 {
		t.Fatalf("expected 5 commands, got %d", len(commands))
	}
}
