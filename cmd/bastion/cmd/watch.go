package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/t-ishitsuka/bastion-core/internal/orchestrator"
	"github.com/t-ishitsuka/bastion-core/internal/terminal"
)

// watch コマンド
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "inbox 監視を開始",
	Long: `inbox ディレクトリを監視し、ファイル変更を検知したらエージェントに通知します。

このコマンドは通常、bastion start によって自動的に起動されます。`,
	RunE: runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	terminal.PrintInfo("inbox 監視を開始しています...")

	// プロジェクトルートを取得
	projectRoot, err := os.Getwd()
	if err != nil {
		terminal.PrintError("プロジェクトルートの取得に失敗: %v", err)
		return err
	}

	// Orchestrator を作成
	orch := orchestrator.NewOrchestrator(projectRoot, 0)

	// watcher を起動
	if err := orch.StartWatcher(); err != nil {
		terminal.PrintError("watcher の起動に失敗: %v", err)
		return err
	}

	terminal.PrintSuccess("✓ inbox 監視を開始しました")
	terminal.PrintInfo("監視ディレクトリ: %s/agents/queue/inbox", projectRoot)
	terminal.PrintInfo("終了: Ctrl+C")

	// シグナルハンドラーをセットアップ
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// watcher が動作し続ける
	<-sigChan

	fmt.Println()
	terminal.PrintInfo("watcher を停止しています...")

	// watcher を停止
	if err := orch.StopWatcher(); err != nil {
		log.Printf("watcher の停止に失敗: %v", err)
	}

	terminal.PrintSuccess("✓ watcher を停止しました")
	return nil
}
