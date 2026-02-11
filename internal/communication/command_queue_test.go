package communication

import (
	"fmt"
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

	// 個別ファイルが作成されたことを確認
	taskPath := filepath.Join(tmpDir, "tasks", "cmd_001.yaml")
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		t.Fatalf("task file was not created: %s", taskPath)
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

	// 並行して書き込みを行う（異なる ID を使用）
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(n int) {
			cmd := Command{
				ID:        fmt.Sprintf("cmd_%d", n),
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

func TestCommandQueueManager_ReadByID(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewCommandQueueManager(tmpDir)

	// 指令を書き込む
	cmd := Command{
		ID:        "cmd_test",
		Timestamp: time.Now(),
		Purpose:   "特定タスク取得テスト",
		Command:   "テスト",
		Status:    CommandStatusPending,
	}

	_ = manager.Write(cmd)

	// ID で読み込む
	readCmd, err := manager.ReadByID("cmd_test")
	if err != nil {
		t.Fatalf("ReadByID failed: %v", err)
	}

	if readCmd.ID != "cmd_test" {
		t.Errorf("expected ID 'cmd_test', got '%s'", readCmd.ID)
	}
}

func TestCommandQueueManager_UpdateStatus(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewCommandQueueManager(tmpDir)

	// 指令を書き込む
	cmd := Command{
		ID:        "cmd_update",
		Timestamp: time.Now(),
		Purpose:   "状態更新テスト",
		Command:   "テスト",
		Status:    CommandStatusPending,
	}

	_ = manager.Write(cmd)

	// 状態を更新
	err := manager.UpdateStatus("cmd_update", CommandStatusInProgress)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	// 読み込んで確認
	readCmd, err := manager.ReadByID("cmd_update")
	if err != nil {
		t.Fatalf("ReadByID failed: %v", err)
	}

	if readCmd.Status != CommandStatusInProgress {
		t.Errorf("expected status 'in_progress', got '%s'", readCmd.Status)
	}
}

func TestCommandQueueManager_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewCommandQueueManager(tmpDir)

	// 指令を書き込む
	cmd := Command{
		ID:        "cmd_delete",
		Timestamp: time.Now(),
		Purpose:   "削除テスト",
		Command:   "テスト",
		Status:    CommandStatusPending,
	}

	_ = manager.Write(cmd)

	// 削除
	err := manager.Delete("cmd_delete")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 削除されたことを確認
	_, err = manager.ReadByID("cmd_delete")
	if err == nil {
		t.Fatalf("expected error when reading deleted task")
	}

	// 存在しないタスクの削除はエラーにならない
	err = manager.Delete("cmd_nonexistent")
	if err != nil {
		t.Fatalf("Delete of non-existent task should not error: %v", err)
	}
}
