package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/t-ishitsuka/bastion-core/internal/parallel"
	"github.com/t-ishitsuka/bastion-core/internal/terminal"
)

var (
	specialists int
)

// start コマンド
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Bastion セッションを起動",
	Long: `Bastion マルチエージェントセッションを起動します。

tmux セッションを作成し、以下のウィンドウを起動します:
  - envoy: ユーザーとの対話窓口
  - marshall: タスク管理・並列実行制御
  - specialists: 複数の専門エージェント`,
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVarP(&specialists, "specialists", "s", 2, "Specialist エージェント数")
}

func runStart(cmd *cobra.Command, args []string) error {
	terminal.PrintInfo("Bastion セッションを起動しています...")

	// tmux セッションをセットアップ
	if err := parallel.SetupBastionSession(); err != nil {
		terminal.PrintError("セッションの作成に失敗しました: %v", err)
		return err
	}

	terminal.PrintSuccess("✓ tmux セッションを作成しました")

	// セッション情報を表示
	sm := parallel.NewSessionManager()
	windows, err := sm.ListWindows()
	if err != nil {
		terminal.PrintWarning("ウィンドウ一覧の取得に失敗: %v", err)
	} else {
		fmt.Println()
		terminal.PrintInfo("作成されたウィンドウ:")
		for _, w := range windows {
			terminal.PrintfGreen("  • %s\n", w)
		}
	}

	fmt.Println()
	terminal.PrintSuccess("Bastion セッションが起動しました")
	terminal.PrintInfo("セッションにアタッチ: tmux attach -t %s", parallel.SessionName)
	terminal.PrintInfo("セッションを確認: bastion status")

	return nil
}
