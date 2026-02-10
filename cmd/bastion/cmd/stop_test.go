package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/t-ishitsuka/bastion-core/internal/parallel"
)

func TestStopCommand_SessionRunning(t *testing.T) {
	if !isTmuxAvailableForCmd() {
		t.Skip("tmux is not available")
	}

	sm := parallel.NewSessionManager()

	// セッションを作成
	if err := parallel.SetupBastionSession(); err != nil {
		t.Fatalf("failed to setup session: %v", err)
	}

	// stop コマンドを実行
	err := runStop(&cobra.Command{}, []string{})
	if err != nil {
		t.Fatalf("stop command failed: %v", err)
	}

	// セッションが停止していることを確認
	exists, err := sm.SessionExists()
	if err != nil {
		t.Fatalf("failed to check session: %v", err)
	}
	if exists {
		t.Error("session should not exist after stop command")
	}
}

func TestStopCommand_SessionNotRunning(t *testing.T) {
	if !isTmuxAvailableForCmd() {
		t.Skip("tmux is not available")
	}

	sm := parallel.NewSessionManager()

	// セッションが存在しない状態にする
	if exists, _ := sm.SessionExists(); exists {
		_ = sm.KillSession()
	}

	// stop コマンドを実行（エラーにならないはず）
	err := runStop(&cobra.Command{}, []string{})
	if err != nil {
		t.Errorf("stop command should not fail when session doesn't exist: %v", err)
	}
}

func TestStopCommand_Idempotent(t *testing.T) {
	if !isTmuxAvailableForCmd() {
		t.Skip("tmux is not available")
	}

	// セッションを作成
	if err := parallel.SetupBastionSession(); err != nil {
		t.Fatalf("failed to setup session: %v", err)
	}

	// 1回目の stop
	err := runStop(&cobra.Command{}, []string{})
	if err != nil {
		t.Fatalf("first stop failed: %v", err)
	}

	// 2回目の stop（エラーにならないはず）
	err = runStop(&cobra.Command{}, []string{})
	if err != nil {
		t.Errorf("second stop should not fail: %v", err)
	}
}
