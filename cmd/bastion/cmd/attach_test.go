package cmd

import (
	"testing"

	"github.com/t-ishitsuka/bastion-core/internal/parallel"
)

func TestAttachCommand_SessionNotExists(t *testing.T) {
	if !isTmuxAvailableForCmd() {
		t.Skip("tmux is not available")
	}

	sm := parallel.NewSessionManager()
	defer cleanupSession(t, sm)

	// セッションが存在しないことを確認
	exists, err := sm.SessionExists()
	if err != nil {
		t.Fatalf("failed to check session: %v", err)
	}
	if exists {
		// 既存セッションをクリーンアップ
		_ = sm.KillSession()
	}

	// attach コマンドは syscall.Exec() を使うため、
	// 実際の実行テストはスキップし、セッション存在確認のみテスト
	exists, err = sm.SessionExists()
	if err != nil {
		t.Fatalf("failed to check session: %v", err)
	}
	if exists {
		t.Error("session should not exist")
	}
}

func TestAttachCommand_SessionExists(t *testing.T) {
	if !isTmuxAvailableForCmd() {
		t.Skip("tmux is not available")
	}

	sm := parallel.NewSessionManager()
	defer cleanupSession(t, sm)

	// セッションを作成
	if err := parallel.SetupBastionSession(); err != nil {
		t.Fatalf("failed to setup session: %v", err)
	}

	// セッションが存在することを確認
	exists, err := sm.SessionExists()
	if err != nil {
		t.Fatalf("failed to check session: %v", err)
	}
	if !exists {
		t.Error("session should exist after setup")
	}

	// attach コマンドは syscall.Exec() を使うため、
	// 実際の実行テストはスキップし、セッション存在確認のみテスト
}

func TestAttachCommandHelp(t *testing.T) {
	// help オプションのテスト
	rootCmd.SetArgs([]string{"attach", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("attach --help が失敗しました: %v", err)
	}
}
