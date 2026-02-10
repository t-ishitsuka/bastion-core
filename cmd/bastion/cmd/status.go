package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/t-ishitsuka/bastion-core/internal/parallel"
	"github.com/t-ishitsuka/bastion-core/internal/terminal"
)

// status コマンド
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Bastion セッションの状態を確認",
	Long: `Bastion セッションの実行状態を表示します。

tmux セッション、ウィンドウ、ペインの状態を確認できます。`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	sm := parallel.NewSessionManager()

	// セッションの存在確認
	exists, err := sm.SessionExists()
	if err != nil {
		terminal.PrintError("セッションの確認に失敗しました: %v", err)
		return err
	}

	if !exists {
		terminal.PrintWarning("Bastion セッションは起動していません")
		terminal.PrintInfo("起動するには: bastion start")
		return nil
	}

	terminal.PrintSuccess("✓ Bastion セッションは起動中です")
	fmt.Println()

	// ウィンドウ一覧を取得
	windows, err := sm.ListWindows()
	if err != nil {
		terminal.PrintError("ウィンドウ一覧の取得に失敗: %v", err)
		return err
	}

	terminal.PrintInfo("ウィンドウ一覧:")
	for _, w := range windows {
		// 各ウィンドウのペイン数を取得
		panes, err := sm.ListPanes(w)
		if err != nil {
			terminal.PrintfYellow("  • %s (ペイン情報取得失敗)\n", w)
			continue
		}
		terminal.PrintfGreen("  • %s (%d ペイン)\n", w, len(panes))
	}

	fmt.Println()
	terminal.PrintInfo("セッションにアタッチ: tmux attach -t %s", parallel.SessionName)
	terminal.PrintInfo("セッションを停止: bastion stop")

	return nil
}
