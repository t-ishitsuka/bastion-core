package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/t-ishitsuka/bastion-core/internal/parallel"
)

func TestStatusCommand_SessionRunning(t *testing.T) {
	if !isTmuxAvailableForCmd() {
		t.Skip("tmux is not available")
	}

	sm := parallel.NewSessionManager()
	defer cleanupSession(t, sm)

	// セッションを作成
	if err := parallel.SetupBastionSession(); err != nil {
		t.Fatalf("failed to setup session: %v", err)
	}

	// status コマンドを実行
	err := runStatus(&cobra.Command{}, []string{})
	if err != nil {
		t.Errorf("status command failed: %v", err)
	}
}

func TestStatusCommand_SessionNotRunning(t *testing.T) {
	if !isTmuxAvailableForCmd() {
		t.Skip("tmux is not available")
	}

	sm := parallel.NewSessionManager()

	// セッションが存在しない状態にする
	if exists, _ := sm.SessionExists(); exists {
		_ = sm.KillSession()
	}

	// status コマンドを実行（エラーにならないはず）
	err := runStatus(&cobra.Command{}, []string{})
	if err != nil {
		t.Errorf("status command should not fail when session doesn't exist: %v", err)
	}
}
