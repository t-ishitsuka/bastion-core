package terminal

import (
	"fmt"
	"os"
)

// ANSI エスケープコード
const (
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorBlue   = "\x1b[34m"
	colorCyan   = "\x1b[36m"
)

// 色付けが有効かどうかを判定する
func isColorEnabled() bool {
	// NO_COLOR 環境変数が設定されていたら無効
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	// ターミナルかどうかをチェック
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}
	return true
}

// Green は緑色のテキストを返す（成功メッセージ用）
func Green(text string) string {
	if !isColorEnabled() {
		return text
	}
	return colorGreen + text + colorReset
}

// Red は赤色のテキストを返す（エラーメッセージ用）
func Red(text string) string {
	if !isColorEnabled() {
		return text
	}
	return colorRed + text + colorReset
}

// Yellow は黄色のテキストを返す（警告メッセージ用）
func Yellow(text string) string {
	if !isColorEnabled() {
		return text
	}
	return colorYellow + text + colorReset
}

// Blue は青色のテキストを返す（情報メッセージ用）
func Blue(text string) string {
	if !isColorEnabled() {
		return text
	}
	return colorBlue + text + colorReset
}

// Cyan はシアン色のテキストを返す（進捗表示用）
func Cyan(text string) string {
	if !isColorEnabled() {
		return text
	}
	return colorCyan + text + colorReset
}

// PrintSuccess は成功メッセージを表示する
func PrintSuccess(format string, args ...interface{}) {
	fmt.Printf(Green("✅ "+format)+"\n", args...)
}

// PrintError はエラーメッセージを表示する
func PrintError(format string, args ...interface{}) {
	fmt.Printf(Red("✗ "+format)+"\n", args...)
}

// PrintWarning は警告メッセージを表示する
func PrintWarning(format string, args ...interface{}) {
	fmt.Printf(Yellow("⚠️  "+format)+"\n", args...)
}

// PrintInfo は情報メッセージを表示する
func PrintInfo(format string, args ...interface{}) {
	fmt.Printf(Blue("📦 "+format)+"\n", args...)
}

// PrintfBlue は青色でフォーマット出力する
func PrintfBlue(format string, args ...interface{}) {
	fmt.Printf(Blue(format), args...)
}

// PrintlnBlue は青色で改行付き出力する
func PrintlnBlue(text string) {
	fmt.Println(Blue(text))
}

// PrintfGreen は緑色でフォーマット出力する
func PrintfGreen(format string, args ...interface{}) {
	fmt.Printf(Green(format), args...)
}

// PrintlnGreen は緑色で改行付き出力する
func PrintlnGreen(text string) {
	fmt.Println(Green(text))
}

// PrintfYellow は黄色でフォーマット出力する
func PrintfYellow(format string, args ...interface{}) {
	fmt.Printf(Yellow(format), args...)
}

// PrintlnYellow は黄色で改行付き出力する
func PrintlnYellow(text string) {
	fmt.Println(Yellow(text))
}

// PrintfRed は赤色でフォーマット出力する
func PrintfRed(format string, args ...interface{}) {
	fmt.Printf(Red(format), args...)
}

// PrintlnRed は赤色で改行付き出力する
func PrintlnRed(text string) {
	fmt.Println(Red(text))
}

// PrintfCyan はシアン色でフォーマット出力する
func PrintfCyan(format string, args ...interface{}) {
	fmt.Printf(Cyan(format), args...)
}

// PrintlnCyan はシアン色で改行付き出力する
func PrintlnCyan(text string) {
	fmt.Println(Cyan(text))
}
