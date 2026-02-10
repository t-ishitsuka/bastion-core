package parallel

import (
	"os"
	"os/exec"
	"testing"
	"time"
)

// tmux が利用可能かチェック
func isTmuxAvailable() bool {
	// CI環境ではtmuxテストをスキップ（PTYがないため）
	if os.Getenv("CI") != "" {
		return false
	}
	cmd := exec.Command("tmux", "-V")
	return cmd.Run() == nil
}

// テスト後のクリーンアップ
func cleanupSession(t *testing.T, sm *SessionManager) {
	if exists, _ := sm.SessionExists(); exists {
		if err := sm.KillSession(); err != nil {
			t.Logf("cleanup warning: failed to kill session: %v", err)
		}
	}
}

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager()
	if sm == nil {
		t.Fatal("NewSessionManager returned nil")
	}
	if sm.sessionName != SessionName {
		t.Errorf("expected session name %s, got %s", SessionName, sm.sessionName)
	}
}

func TestSessionManager_CreateSession(t *testing.T) {
	if !isTmuxAvailable() {
		t.Skip("tmux is not available")
	}

	sm := NewSessionManager()
	defer cleanupSession(t, sm)

	err := sm.CreateSession()
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// セッションが存在することを確認
	exists, err := sm.SessionExists()
	if err != nil {
		t.Fatalf("failed to check session existence: %v", err)
	}
	if !exists {
		t.Error("session should exist after creation")
	}
}

func TestSessionManager_SessionExists(t *testing.T) {
	if !isTmuxAvailable() {
		t.Skip("tmux is not available")
	}

	sm := NewSessionManager()
	defer cleanupSession(t, sm)

	// セッション作成前は存在しない
	exists, err := sm.SessionExists()
	if err != nil {
		t.Fatalf("failed to check session existence: %v", err)
	}
	if exists {
		t.Error("session should not exist before creation")
	}

	// セッション作成
	if err := sm.CreateSession(); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// セッション作成後は存在する
	exists, err = sm.SessionExists()
	if err != nil {
		t.Fatalf("failed to check session existence: %v", err)
	}
	if !exists {
		t.Error("session should exist after creation")
	}
}

func TestSessionManager_CreateWindow(t *testing.T) {
	if !isTmuxAvailable() {
		t.Skip("tmux is not available")
	}

	sm := NewSessionManager()
	defer cleanupSession(t, sm)

	// セッション作成
	if err := sm.CreateSession(); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// ウィンドウ作成
	windowName := "test-window"
	if err := sm.CreateWindow(windowName); err != nil {
		t.Fatalf("failed to create window: %v", err)
	}

	// ウィンドウ一覧を取得して確認
	windows, err := sm.ListWindows()
	if err != nil {
		t.Fatalf("failed to list windows: %v", err)
	}

	found := false
	for _, w := range windows {
		if w == windowName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("window %s not found in list: %v", windowName, windows)
	}
}

func TestSessionManager_SplitPane(t *testing.T) {
	if !isTmuxAvailable() {
		t.Skip("tmux is not available")
	}

	sm := NewSessionManager()
	defer cleanupSession(t, sm)

	// セッション作成
	if err := sm.CreateSession(); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// 初期状態は1ペイン
	panes, err := sm.ListPanes(WindowEnvoy)
	if err != nil {
		t.Fatalf("failed to list panes: %v", err)
	}
	if len(panes) != 1 {
		t.Errorf("expected 1 pane initially, got %d", len(panes))
	}

	// 水平分割
	if err := sm.SplitPaneHorizontal(WindowEnvoy); err != nil {
		t.Fatalf("failed to split pane horizontally: %v", err)
	}

	// ペイン数が増えていることを確認
	panes, err = sm.ListPanes(WindowEnvoy)
	if err != nil {
		t.Fatalf("failed to list panes: %v", err)
	}
	if len(panes) != 2 {
		t.Errorf("expected 2 panes after split, got %d", len(panes))
	}

	// 垂直分割
	if err := sm.SplitPaneVertical(WindowEnvoy); err != nil {
		t.Fatalf("failed to split pane vertically: %v", err)
	}

	panes, err = sm.ListPanes(WindowEnvoy)
	if err != nil {
		t.Fatalf("failed to list panes: %v", err)
	}
	if len(panes) != 3 {
		t.Errorf("expected 3 panes after second split, got %d", len(panes))
	}
}

func TestSessionManager_SendKeys(t *testing.T) {
	if !isTmuxAvailable() {
		t.Skip("tmux is not available")
	}

	sm := NewSessionManager()
	defer cleanupSession(t, sm)

	// セッション作成
	if err := sm.CreateSession(); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// コマンド送信（エラーが発生しないことを確認）
	if err := sm.SendKeys(WindowEnvoy, "echo test", true); err != nil {
		t.Errorf("failed to send keys: %v", err)
	}

	// 少し待機（コマンド実行のため）
	time.Sleep(100 * time.Millisecond)
}

func TestSessionManager_KillSession(t *testing.T) {
	if !isTmuxAvailable() {
		t.Skip("tmux is not available")
	}

	sm := NewSessionManager()

	// セッション作成
	if err := sm.CreateSession(); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// セッション停止
	if err := sm.KillSession(); err != nil {
		t.Fatalf("failed to kill session: %v", err)
	}

	// セッションが存在しないことを確認
	exists, err := sm.SessionExists()
	if err != nil {
		t.Fatalf("failed to check session existence: %v", err)
	}
	if exists {
		t.Error("session should not exist after kill")
	}
}

func TestSessionManager_KillSession_Idempotent(t *testing.T) {
	if !isTmuxAvailable() {
		t.Skip("tmux is not available")
	}

	sm := NewSessionManager()

	// セッションが存在しない状態で Kill を呼ぶ
	if err := sm.KillSession(); err != nil {
		t.Errorf("KillSession should not error when session doesn't exist: %v", err)
	}
}

func TestSetupBastionSession(t *testing.T) {
	if !isTmuxAvailable() {
		t.Skip("tmux is not available")
	}

	sm := NewSessionManager()
	defer cleanupSession(t, sm)

	// Bastion セッションをセットアップ
	if err := SetupBastionSession(); err != nil {
		t.Fatalf("failed to setup bastion session: %v", err)
	}

	// セッションが存在することを確認
	exists, err := sm.SessionExists()
	if err != nil {
		t.Fatalf("failed to check session existence: %v", err)
	}
	if !exists {
		t.Error("bastion session should exist after setup")
	}

	// 期待するウィンドウが作成されていることを確認
	windows, err := sm.ListWindows()
	if err != nil {
		t.Fatalf("failed to list windows: %v", err)
	}

	expectedWindows := []string{WindowEnvoy, WindowMarshall, WindowSpecialists}
	for _, expected := range expectedWindows {
		found := false
		for _, w := range windows {
			if w == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected window %s not found in list: %v", expected, windows)
		}
	}
}

func TestSetupBastionSession_Idempotent(t *testing.T) {
	if !isTmuxAvailable() {
		t.Skip("tmux is not available")
	}

	sm := NewSessionManager()
	defer cleanupSession(t, sm)

	// 1回目
	if err := SetupBastionSession(); err != nil {
		t.Fatalf("first setup failed: %v", err)
	}

	// 2回目（既存セッションを削除してから作成されるはず）
	if err := SetupBastionSession(); err != nil {
		t.Fatalf("second setup failed: %v", err)
	}

	// セッションが存在することを確認
	exists, err := sm.SessionExists()
	if err != nil {
		t.Fatalf("failed to check session existence: %v", err)
	}
	if !exists {
		t.Error("bastion session should exist after second setup")
	}
}
