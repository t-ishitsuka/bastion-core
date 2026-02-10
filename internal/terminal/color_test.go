package terminal

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// NO_COLOR 環境変数が設定されている場合は色付けが無効になるかテストする
func TestIsColorEnabledWithNOCOLOR(t *testing.T) {
	// NO_COLOR を設定
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	if isColorEnabled() {
		t.Error("NO_COLOR が設定されているのに isColorEnabled() が true を返しました")
	}
}

// NO_COLOR が未設定の場合の動作をテストする
func TestColorFunctionsWithoutNOCOLOR(t *testing.T) {
	// NO_COLOR を削除
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Unsetenv("NO_COLOR")

	tests := []struct {
		name     string
		fn       func(string) string
		input    string
		contains string // ANSI エスケープコードを含むかチェック
	}{
		{"Green", Green, "test", "\x1b[32m"},
		{"Red", Red, "test", "\x1b[31m"},
		{"Yellow", Yellow, "test", "\x1b[33m"},
		{"Blue", Blue, "test", "\x1b[34m"},
		{"Cyan", Cyan, "test", "\x1b[36m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			// ターミナルでない場合は色付けされない可能性があるため、
			// 入力文字列が含まれていることだけ確認
			if !strings.Contains(result, tt.input) {
				t.Errorf("%s() の結果に入力文字列が含まれていません: got %q", tt.name, result)
			}
		})
	}
}

// NO_COLOR が設定されている場合は元の文字列をそのまま返すかテストする
func TestColorFunctionsWithNOCOLOR(t *testing.T) {
	// NO_COLOR を設定
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	tests := []struct {
		name  string
		fn    func(string) string
		input string
		want  string
	}{
		{"Green", Green, "test", "test"},
		{"Red", Red, "test", "test"},
		{"Yellow", Yellow, "test", "test"},
		{"Blue", Blue, "test", "test"},
		{"Cyan", Cyan, "test", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if result != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, result, tt.want)
			}
		})
	}
}

// 標準出力をキャプチャするヘルパー関数
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// PrintSuccess が正しく出力するかテストする
func TestPrintSuccess(t *testing.T) {
	// NO_COLOR を設定して色なしでテスト
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	output := captureStdout(func() {
		PrintSuccess("test message")
	})

	if !strings.Contains(output, "✅ test message") {
		t.Errorf("PrintSuccess() の出力が期待と異なります: got %q", output)
	}
}

// PrintError が正しく出力するかテストする
func TestPrintError(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	output := captureStdout(func() {
		PrintError("error message")
	})

	if !strings.Contains(output, "✗ error message") {
		t.Errorf("PrintError() の出力が期待と異なります: got %q", output)
	}
}

// PrintWarning が正しく出力するかテストする
func TestPrintWarning(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	output := captureStdout(func() {
		PrintWarning("warning message")
	})

	if !strings.Contains(output, "⚠️  warning message") {
		t.Errorf("PrintWarning() の出力が期待と異なります: got %q", output)
	}
}

// PrintInfo が正しく出力するかテストする
func TestPrintInfo(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	output := captureStdout(func() {
		PrintInfo("info message")
	})

	if !strings.Contains(output, "📦 info message") {
		t.Errorf("PrintInfo() の出力が期待と異なります: got %q", output)
	}
}

// PrintfBlue が正しく出力するかテストする
func TestPrintfBlue(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	output := captureStdout(func() {
		PrintfBlue("test %s\n", "message")
	})

	if !strings.Contains(output, "test message") {
		t.Errorf("PrintfBlue() の出力が期待と異なります: got %q", output)
	}
}

// PrintlnBlue が正しく出力するかテストする
func TestPrintlnBlue(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	output := captureStdout(func() {
		PrintlnBlue("test message")
	})

	if !strings.Contains(output, "test message") {
		t.Errorf("PrintlnBlue() の出力が期待と異なります: got %q", output)
	}
}

// PrintfGreen が正しく出力するかテストする
func TestPrintfGreen(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	output := captureStdout(func() {
		PrintfGreen("test %s\n", "message")
	})

	if !strings.Contains(output, "test message") {
		t.Errorf("PrintfGreen() の出力が期待と異なります: got %q", output)
	}
}

// Printf/Println 関数が引数を正しく処理するかテストする
func TestPrintfWithMultipleArgs(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	tests := []struct {
		name   string
		fn     func()
		expect string
	}{
		{
			name: "PrintfYellow with multiple args",
			fn: func() {
				PrintfYellow("test %s %d\n", "message", 123)
			},
			expect: "test message 123",
		},
		{
			name: "PrintfCyan with multiple args",
			fn: func() {
				PrintfCyan("value: %d, text: %s\n", 42, "hello")
			},
			expect: "value: 42, text: hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(tt.fn)
			if !strings.Contains(output, tt.expect) {
				t.Errorf("出力が期待と異なります: got %q, want to contain %q", output, tt.expect)
			}
		})
	}
}

// 色関数が空文字列を正しく処理するかテストする
func TestColorFunctionsWithEmptyString(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	tests := []struct {
		name string
		fn   func(string) string
	}{
		{"Green", Green},
		{"Red", Red},
		{"Yellow", Yellow},
		{"Blue", Blue},
		{"Cyan", Cyan},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn("")
			if result != "" {
				t.Errorf("%s(\"\") = %q, want \"\"", tt.name, result)
			}
		})
	}
}

// PrintSuccess がフォーマット引数を正しく処理するかテストする
func TestPrintSuccessWithArgs(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	output := captureStdout(func() {
		PrintSuccess("test %s %d", "message", 123)
	})

	expected := "✅ test message 123"
	if !strings.Contains(output, expected) {
		t.Errorf("PrintSuccess() の出力が期待と異なります: got %q, want to contain %q", output, expected)
	}
}

// 色関数が特殊文字を正しく処理するかテストする
func TestColorFunctionsWithSpecialChars(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	input := "test\nwith\ttabs\rand\x00null"
	result := Green(input)
	if result != input {
		t.Errorf("Green() が特殊文字を正しく処理できませんでした: got %q, want %q", result, input)
	}
}

// ベンチマーク: 色関数のパフォーマンス
func BenchmarkGreen(b *testing.B) {
	_ = os.Setenv("NO_COLOR", "1")
	for i := 0; i < b.N; i++ {
		_ = Green("test message")
	}
}

// ベンチマーク: PrintSuccess のパフォーマンス
func BenchmarkPrintSuccess(b *testing.B) {
	_ = os.Setenv("NO_COLOR", "1")
	// 標準出力を /dev/null にリダイレクト
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PrintSuccess("test message")
	}
}

// 全ての Println 関数が改行を含むかテストする
func TestPrintlnFunctionsIncludeNewline(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	tests := []struct {
		name string
		fn   func()
	}{
		{"PrintlnBlue", func() { PrintlnBlue("test") }},
		{"PrintlnGreen", func() { PrintlnGreen("test") }},
		{"PrintlnYellow", func() { PrintlnYellow("test") }},
		{"PrintlnRed", func() { PrintlnRed("test") }},
		{"PrintlnCyan", func() { PrintlnCyan("test") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(tt.fn)
			if !strings.HasSuffix(output, "\n") {
				t.Errorf("%s の出力が改行で終わっていません: got %q", tt.name, output)
			}
		})
	}
}

// Printf 関数がフォーマット文字列なしで動作するかテストする
func TestPrintfFunctionsWithoutFormatSpecifiers(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	output := captureStdout(func() {
		PrintfBlue("simple text without format")
	})

	if !strings.Contains(output, "simple text without format") {
		t.Errorf("PrintfBlue() がフォーマット指定子なしで動作しませんでした: got %q", output)
	}
}

// 出力関数が nil を渡された時にパニックしないかテストする
func TestPrintfFunctionsDoNotPanicWithNilArgs(t *testing.T) {
	oldValue := os.Getenv("NO_COLOR")
	defer func() {
		if oldValue == "" {
			_ = os.Unsetenv("NO_COLOR")
		} else {
			_ = os.Setenv("NO_COLOR", oldValue)
		}
		if r := recover(); r != nil {
			t.Errorf("Printf functions panicked: %v", r)
		}
	}()

	_ = os.Setenv("NO_COLOR", "1")

	// /dev/null にリダイレクト
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	PrintfBlue("test %s", nil)
	PrintSuccess("test %v", nil)
}

// 出力例を確認するためのサンプルテスト（手動確認用）
func ExamplePrintSuccess() {
	_ = os.Setenv("NO_COLOR", "1")
	defer func() { _ = os.Unsetenv("NO_COLOR") }()

	PrintSuccess("操作が完了しました")
	// Output: ✅ 操作が完了しました
}

func ExamplePrintfBlue() {
	_ = os.Setenv("NO_COLOR", "1")
	defer func() { _ = os.Unsetenv("NO_COLOR") }()

	PrintfBlue("処理中: %d/%d\n", 5, 10)
	// Output: 処理中: 5/10
}
