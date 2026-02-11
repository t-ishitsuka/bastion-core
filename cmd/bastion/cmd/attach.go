package cmd

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/t-ishitsuka/bastion-core/internal/parallel"
	"github.com/t-ishitsuka/bastion-core/internal/terminal"
)

// attach コマンド
var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Bastion セッションにアタッチ",
	Long: `既存の Bastion tmux セッションにアタッチします。

セッションが存在しない場合は、先に 'bastion start' を実行してください。`,
	RunE: runAttach,
}

func init() {
	rootCmd.AddCommand(attachCmd)
}

func runAttach(cmd *cobra.Command, args []string) error {
	sm := parallel.NewSessionManager()

	// セッションの存在確認
	exists, err := sm.SessionExists()
	if err != nil {
		terminal.PrintError("セッションの確認に失敗しました: %v", err)
		return err
	}

	if !exists {
		terminal.PrintError("Bastion セッションが見つかりません")
		terminal.PrintInfo("先に起動してください: bastion start")
		return nil
	}

	terminal.PrintInfo("Bastion セッションにアタッチしています...")

	// tmux のパスを取得
	binary, err := exec.LookPath("tmux")
	if err != nil {
		terminal.PrintError("tmux が見つかりません: %v", err)
		return err
	}

	// 現在のプロセスを tmux attach に置き換える
	tmuxArgs := []string{"tmux", "attach", "-t", parallel.SessionName}
	env := os.Environ()

	err = syscall.Exec(binary, tmuxArgs, env)
	if err != nil {
		terminal.PrintError("tmux へのアタッチに失敗しました: %v", err)
		return err
	}

	// syscall.Exec() が成功すると、ここには到達しない
	return nil
}
