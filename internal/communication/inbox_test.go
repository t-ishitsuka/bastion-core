package communication

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInboxManager_Write(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	manager := NewInboxManager(tmpDir)

	// メッセージを書き込む
	err := manager.Write("marshall", "新規タスク割当", MessageTypeTaskAssigned, "envoy")
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// ファイルが作成されたことを確認
	inboxPath := filepath.Join(tmpDir, "inbox", "marshall.yaml")
	if _, err := os.Stat(inboxPath); os.IsNotExist(err) {
		t.Fatalf("inbox file was not created: %s", inboxPath)
	}

	// メッセージを読み込んで確認
	messages, err := manager.Read("marshall")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.From != "envoy" {
		t.Errorf("expected from 'envoy', got '%s'", msg.From)
	}
	if msg.Type != MessageTypeTaskAssigned {
		t.Errorf("expected type 'task_assigned', got '%s'", msg.Type)
	}
	if msg.Message != "新規タスク割当" {
		t.Errorf("expected message '新規タスク割当', got '%s'", msg.Message)
	}
	if msg.Status != MessageStatusPending {
		t.Errorf("expected status 'pending', got '%s'", msg.Status)
	}
}

func TestInboxManager_MultipleWrites(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewInboxManager(tmpDir)

	// 複数のメッセージを書き込む
	err := manager.Write("specialist_1", "タスク1", MessageTypeTaskAssigned, "marshall")
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	err = manager.Write("specialist_1", "タスク2", MessageTypeTaskAssigned, "marshall")
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// メッセージを読み込んで確認
	messages, err := manager.Read("specialist_1")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
}

func TestInboxManager_ReadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewInboxManager(tmpDir)

	// 存在しない inbox を読み込む
	messages, err := manager.Read("nonexistent")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(messages) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(messages))
	}
}

func TestInboxManager_MarkAsProcessed(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewInboxManager(tmpDir)

	// メッセージを書き込む
	err := manager.Write("marshall", "テストメッセージ", MessageTypeWakeUp, "system")
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// メッセージを読み込む
	messages, err := manager.Read("marshall")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(messages) == 0 {
		t.Fatal("no messages found")
	}

	messageID := messages[0].ID

	// 処理済みにする
	err = manager.MarkAsProcessed("marshall", messageID)
	if err != nil {
		t.Fatalf("MarkAsProcessed failed: %v", err)
	}

	// 再度読み込んで確認
	messages, err = manager.Read("marshall")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if messages[0].Status != MessageStatusProcessed {
		t.Errorf("expected status 'processed', got '%s'", messages[0].Status)
	}
}

func TestInboxManager_GetPendingMessages(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewInboxManager(tmpDir)

	// 複数のメッセージを書き込む
	_ = manager.Write("marshall", "メッセージ1", MessageTypeTaskAssigned, "envoy")
	_ = manager.Write("marshall", "メッセージ2", MessageTypeTaskAssigned, "envoy")
	_ = manager.Write("marshall", "メッセージ3", MessageTypeTaskAssigned, "envoy")

	// メッセージを読み込む
	messages, _ := manager.Read("marshall")

	// 1つ目を処理済みにする
	_ = manager.MarkAsProcessed("marshall", messages[0].ID)

	// 未処理のメッセージを取得
	pending, err := manager.GetPendingMessages("marshall")
	if err != nil {
		t.Fatalf("GetPendingMessages failed: %v", err)
	}

	if len(pending) != 2 {
		t.Fatalf("expected 2 pending messages, got %d", len(pending))
	}
}

func TestInboxManager_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewInboxManager(tmpDir)

	// 並行して書き込みを行う
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			_ = manager.Write("marshall", "並行書き込みテスト", MessageTypeWakeUp, "test")
			done <- true
		}(i)
	}

	// すべての goroutine が完了するのを待つ
	for i := 0; i < 10; i++ {
		<-done
	}

	// メッセージを読み込んで確認
	messages, err := manager.Read("marshall")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(messages) != 10 {
		t.Fatalf("expected 10 messages, got %d", len(messages))
	}
}
