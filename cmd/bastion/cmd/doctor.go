package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/t-ishitsuka/bastion-core/internal/terminal"
)

const (
	minGoMajor = 1
	minGoMinor = 22
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "環境チェック（必須 CLI ツールの確認）",
	Long:  `Bastion の実行に必要な CLI ツール（claude, tmux, git, go）がインストールされているかをチェックします。`,
	Run:   runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) {
	terminal.PrintInfo("Bastion 環境チェック")
	fmt.Println()

	allOK := true

	// claude CLI のチェック
	if checkCommand("claude", "--version") {
		terminal.PrintSuccess("claude: インストール済み")
	} else {
		terminal.PrintError("claude: 未インストール")
		fmt.Println("  インストール: https://claude.com/claude-code")
		allOK = false
	}

	// tmux のチェック
	if checkCommand("tmux", "-V") {
		version := getCommandOutput("tmux", "-V")
		terminal.PrintSuccess("tmux: インストール済み (%s)", strings.TrimSpace(version))
	} else {
		terminal.PrintError("tmux: 未インストール")
		fmt.Println("  インストール: apt install tmux / brew install tmux")
		allOK = false
	}

	// git のチェック
	if checkCommand("git", "--version") {
		version := getCommandOutput("git", "--version")
		terminal.PrintSuccess("git: インストール済み (%s)", strings.TrimSpace(version))
	} else {
		terminal.PrintError("git: 未インストール")
		fmt.Println("  インストール: apt install git / brew install git")
		allOK = false
	}

	// Go バージョンのチェック
	goVersion := getCommandOutput("go", "version")
	if goVersion == "" {
		terminal.PrintError("go: 未インストール")
		allOK = false
	} else {
		major, minor, ok := parseGoVersion(goVersion)
		if !ok {
			terminal.PrintError("go: バージョン情報を取得できません (%s)", strings.TrimSpace(goVersion))
			allOK = false
		} else if major < minGoMajor || (major == minGoMajor && minor < minGoMinor) {
			terminal.PrintError("go: バージョンが古いです (必要: %d.%d 以上, 現在: %d.%d)",
				minGoMajor, minGoMinor, major, minor)
			allOK = false
		} else {
			terminal.PrintSuccess("go: バージョン OK (%d.%d)", major, minor)
		}
	}

	fmt.Println()

	if allOK {
		terminal.PrintlnGreen("すべての必須ツールがインストールされています")
		os.Exit(0)
	} else {
		terminal.PrintlnRed("いくつかの必須ツールが不足しています")
		os.Exit(1)
	}
}

// コマンドが存在し実行可能かをチェック
func checkCommand(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	err := cmd.Run()
	return err == nil
}

// コマンドを実行し、その出力を返す
func getCommandOutput(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return string(output)
}

// "go version go1.22.0 ..." をパースし、(major, minor, ok) を返す
func parseGoVersion(version string) (int, int, bool) {
	// 正規表現でバージョン番号を抽出
	re := regexp.MustCompile(`go(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(version)

	if len(matches) < 3 {
		return 0, 0, false
	}

	major, err1 := strconv.Atoi(matches[1])
	minor, err2 := strconv.Atoi(matches[2])

	if err1 != nil || err2 != nil {
		return 0, 0, false
	}

	return major, minor, true
}
