package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/t-ishitsuka/bastion-core/internal/orchestrator"
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
	startCmd.Flags().IntVarP(&specialists, "specialists", "s", 4, "Specialist エージェント数")
}

func runStart(cmd *cobra.Command, args []string) error {
	terminal.PrintInfo("Bastion セッションを起動しています...")

	// tmux セッションをセットアップ
	if err := parallel.SetupBastionSession(); err != nil {
		terminal.PrintError("セッションの作成に失敗しました: %v", err)
		return err
	}

	terminal.PrintSuccess("✓ tmux セッションを作成しました")

	// プロジェクトルートを取得
	projectRoot, err := os.Getwd()
	if err != nil {
		terminal.PrintError("プロジェクトルートの取得に失敗: %v", err)
		return err
	}

	// Orchestrator を作成
	orch := orchestrator.NewOrchestrator(projectRoot, specialists)

	terminal.PrintInfo("エージェントを起動しています...")

	// すべてのエージェントを起動
	if err := orch.StartAll(); err != nil {
		terminal.PrintWarning("エージェントの起動に失敗: %v", err)
		terminal.PrintInfo("手動で起動してください: tmux attach -t %s", parallel.SessionName)
	} else {
		terminal.PrintSuccess("✓ エージェントを起動しました")
	}

	// watcher ウィンドウで bastion watch を起動
	terminal.PrintInfo("inbox 監視を開始しています...")
	if err := orch.StartWatcherWindow(); err != nil {
		terminal.PrintWarning("watcher ウィンドウの起動に失敗: %v", err)
	} else {
		terminal.PrintSuccess("✓ inbox 監視を開始しました")
	}

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
	fmt.Println()
	terminal.PrintInfo("各エージェントで Claude Code が起動しています")
	terminal.PrintInfo("Envoy: ユーザー対話窓口")
	terminal.PrintInfo("Marshall: タスク管理")
	terminal.PrintInfo("Specialists: タスク実行 (x%d)", specialists)
	fmt.Println()

	// テストモードでは tmux へのアタッチをスキップ
	if os.Getenv("BASTION_TEST_MODE") == "1" {
		terminal.PrintInfo("テストモード: tmux へのアタッチをスキップします")
		return nil
	}

	terminal.PrintInfo("tmux セッションにアタッチしています...")

	// tmux セッションにアタッチ（現在のプロセスを置き換え）
	binary, err := exec.LookPath("tmux")
	if err != nil {
		terminal.PrintError("tmux が見つかりません: %v", err)
		terminal.PrintInfo("手動でアタッチしてください: tmux attach -t %s", parallel.SessionName)
		return err
	}

	tmuxArgs := []string{"tmux", "attach", "-t", parallel.SessionName}
	env := os.Environ()

	// 現在のプロセスを tmux attach に置き換える
	// この呼び出しが成功すると、この関数は戻らない
	err = syscall.Exec(binary, tmuxArgs, env)
	if err != nil {
		// エラーが発生した場合のみここに到達する
		terminal.PrintError("tmux へのアタッチに失敗: %v", err)
		terminal.PrintInfo("手動でアタッチしてください: tmux attach -t %s", parallel.SessionName)
		return err
	}

	// syscall.Exec() が成功すると、ここには到達しない
	return nil
}
