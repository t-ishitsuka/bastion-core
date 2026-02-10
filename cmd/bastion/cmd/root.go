package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bastion",
	Short: "Bastion - Claude Code マルチエージェントオーケストレーター",
	Long: `Bastion は「一人開発会社」を実現する Claude Code マルチエージェントオーケストレーターです。`,
}

// ルートコマンドを実行
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
