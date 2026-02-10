package cmd

import (
	"github.com/spf13/cobra"
	"github.com/t-ishitsuka/bastion-core/internal/parallel"
	"github.com/t-ishitsuka/bastion-core/internal/terminal"
)

// stop コマンド
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Bastion セッションを停止",
	Long: `Bastion セッションを停止します。

tmux セッションを終了し、すべてのウィンドウとペインを閉じます。`,
	RunE: runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
	sm := parallel.NewSessionManager()

	// セッションの存在確認
	exists, err := sm.SessionExists()
	if err != nil {
		terminal.PrintError("セッションの確認に失敗しました: %v", err)
		return err
	}

	if !exists {
		terminal.PrintWarning("Bastion セッションは起動していません")
		return nil
	}

	terminal.PrintInfo("Bastion セッションを停止しています...")

	// セッションを停止
	if err := sm.KillSession(); err != nil {
		terminal.PrintError("セッションの停止に失敗しました: %v", err)
		return err
	}

	terminal.PrintSuccess("✓ Bastion セッションを停止しました")
	return nil
}
