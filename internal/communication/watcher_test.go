package communication

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// テスト後のクリーンアップ
func cleanupWatcher(t *testing.T, watcher *Watcher) {
	if err := watcher.Stop(); err != nil {
		t.Logf("cleanup warning: failed to stop watcher: %v", err)
	}
}

func TestNewWatcher(t *testing.T) {
	watcher, err := NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer cleanupWatcher(t, watcher)

	if watcher.watcher == nil {
		t.Error("watcher.watcher should not be nil")
	}
	if watcher.events == nil {
		t.Error("watcher.events should not be nil")
	}
	if watcher.errors == nil {
		t.Error("watcher.errors should not be nil")
	}
	if watcher.done == nil {
		t.Error("watcher.done should not be nil")
	}
}

func TestWatcher_Watch(t *testing.T) {
	watcher, err := NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer cleanupWatcher(t, watcher)

	tmpDir := t.TempDir()

	err = watcher.Watch(tmpDir)
	if err != nil {
		t.Errorf("failed to watch directory: %v", err)
	}
}

func TestWatcher_FileChangeDetection(t *testing.T) {
	watcher, err := NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer cleanupWatcher(t, watcher)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// ディレクトリを監視対象に追加
	err = watcher.Watch(tmpDir)
	if err != nil {
		t.Fatalf("failed to watch directory: %v", err)
	}

	// イベント処理を開始
	watcher.Start()

	// ファイルを作成
	if err := os.WriteFile(testFile, []byte("initial content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// ファイルを更新
	if err := os.WriteFile(testFile, []byte("updated content"), 0644); err != nil {
		t.Fatalf("failed to update test file: %v", err)
	}

	// イベントを待機（タイムアウト付き）
	timeout := time.After(2 * time.Second)
	eventReceived := false

	for !eventReceived {
		select {
		case event := <-watcher.Events():
			if filepath.Base(event.Path) == "test.txt" && event.Operation == "write" {
				eventReceived = true
			}
		case err := <-watcher.Errors():
			t.Errorf("watcher error: %v", err)
		case <-timeout:
			t.Fatal("timeout waiting for file change event")
		}
	}
}

func TestWatcher_Stop(t *testing.T) {
	watcher, err := NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}

	tmpDir := t.TempDir()
	err = watcher.Watch(tmpDir)
	if err != nil {
		t.Fatalf("failed to watch directory: %v", err)
	}

	watcher.Start()

	err = watcher.Stop()
	if err != nil {
		t.Errorf("failed to stop watcher: %v", err)
	}

	// Stop 後にイベントが送信されないことを確認
	select {
	case _, ok := <-watcher.Events():
		if ok {
			t.Error("events channel should be closed after Stop")
		}
	case <-time.After(100 * time.Millisecond):
		// チャンネルがクローズされている場合、すぐに受信できる
	}
}

func TestWatcher_StopIdempotent(t *testing.T) {
	watcher, err := NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}

	watcher.Start()

	// 1回目の Stop
	err = watcher.Stop()
	if err != nil {
		t.Errorf("first Stop failed: %v", err)
	}

	// 2回目の Stop（エラーにならないことを確認）
	err = watcher.Stop()
	if err != nil {
		t.Errorf("second Stop failed: %v", err)
	}
}

func TestWatchInbox(t *testing.T) {
	tmpDir := t.TempDir()
	inboxDir := filepath.Join(tmpDir, "inbox")

	// inbox ディレクトリを作成
	if err := os.MkdirAll(inboxDir, 0755); err != nil {
		t.Fatalf("failed to create inbox directory: %v", err)
	}

	watcher, err := WatchInbox(tmpDir)
	if err != nil {
		t.Fatalf("failed to watch inbox: %v", err)
	}
	defer cleanupWatcher(t, watcher)

	// inbox ディレクトリ内でファイルを更新
	testFile := filepath.Join(inboxDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte("test: data"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// イベントを待機
	timeout := time.After(2 * time.Second)
	eventReceived := false

	for !eventReceived {
		select {
		case event := <-watcher.Events():
			if filepath.Base(event.Path) == "test.yaml" && event.Operation == "write" {
				eventReceived = true
			}
		case err := <-watcher.Errors():
			t.Errorf("watcher error: %v", err)
		case <-timeout:
			t.Fatal("timeout waiting for inbox file change event")
		}
	}
}

func TestWatchInbox_NonExistentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// inbox ディレクトリを作成しない

	_, err := WatchInbox(tmpDir)
	if err == nil {
		t.Error("WatchInbox should fail when inbox directory does not exist")
	}
}
