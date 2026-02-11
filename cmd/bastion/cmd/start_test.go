package cmd

import (
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
	"github.com/t-ishitsuka/bastion-core/internal/parallel"
)

// tmux が利用可能かチェック
func isTmuxAvailableForCmd() bool {
	// CI環境ではtmuxテストをスキップ（PTYがないため）
	if os.Getenv("CI") != "" {
		return false
	}
	cmd := exec.Command("tmux", "-V")
	return cmd.Run() == nil
}

// テスト後のクリーンアップ
func cleanupSession(t *testing.T, sm *parallel.SessionManager) {
	if exists, _ := sm.SessionExists(); exists {
		if err := sm.KillSession(); err != nil {
			t.Logf("cleanup warning: failed to kill session: %v", err)
		}
	}
}

// テスト用に環境変数を設定し、テスト後に元に戻す
func setTestEnv(t *testing.T, key, value string) {
	t.Helper()
	oldValue := os.Getenv(key)
	_ = os.Setenv(key, value)
	t.Cleanup(func() {
		if oldValue == "" {
			_ = os.Unsetenv(key)
		} else {
			_ = os.Setenv(key, oldValue)
		}
	})
}

func TestStartCommand(t *testing.T) {
	if !isTmuxAvailableForCmd() {
		t.Skip("tmux is not available")
	}

	// テストモードを有効化
	setTestEnv(t, "BASTION_TEST_MODE", "1")

	sm := parallel.NewSessionManager()
	defer cleanupSession(t, sm)

	// 既存セッションをクリーンアップ
	if exists, _ := sm.SessionExists(); exists {
		_ = sm.KillSession()
	}

	// start コマンドを実行（RunE を直接呼ぶ）
	err := runStart(&cobra.Command{}, []string{})
	if err != nil {
		t.Fatalf("start command failed: %v", err)
	}

	// セッションが作成されていることを確認
	exists, err := sm.SessionExists()
	if err != nil {
		t.Fatalf("failed to check session: %v", err)
	}
	if !exists {
		t.Error("session should exist after start command")
	}
}

func TestStartCommand_AlreadyRunning(t *testing.T) {
	if !isTmuxAvailableForCmd() {
		t.Skip("tmux is not available")
	}

	// テストモードを有効化
	setTestEnv(t, "BASTION_TEST_MODE", "1")

	sm := parallel.NewSessionManager()
	defer cleanupSession(t, sm)

	// セッションを作成
	if err := parallel.SetupBastionSession(); err != nil {
		t.Fatalf("failed to setup session: %v", err)
	}

	// start コマンドを再度実行（既存セッションを削除して再作成される）
	err := runStart(&cobra.Command{}, []string{})
	if err != nil {
		t.Errorf("start command should recreate session: %v", err)
	}

	// セッションが存在することを確認
	exists, err := sm.SessionExists()
	if err != nil {
		t.Fatalf("failed to check session: %v", err)
	}
	if !exists {
		t.Error("session should exist after recreating")
	}
}
