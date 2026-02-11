package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/t-ishitsuka/bastion-core/templates"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "プロジェクトに Bastion エージェント環境を初期化",
	Long: `プロジェクトに Bastion エージェント環境を初期化します。

このコマンドは以下を実行します:
  - テンプレートから agents/ ディレクトリを作成
  - queue/ ディレクトリ構造を作成（inbox/, tasks/, reports/）
  - .gitignore に queue/ を追加

既存のファイルは上書きせず、新しいファイルのみを追加します（--no-clobber モード）。`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// 現在のディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("現在のディレクトリの取得に失敗: %w", err)
	}

	agentsDir := filepath.Join(currentDir, "agents")
	queueDir := filepath.Join(currentDir, "queue")

	fmt.Println("Bastion エージェント環境を初期化しています...")

	// 埋め込まれたテンプレートから agents/ を作成（no-clobber モード）
	copied, skipped, err := copyEmbeddedDirNoClobber(templates.FS, "agents", agentsDir)
	if err != nil {
		return fmt.Errorf("agents/ の作成に失敗: %w", err)
	}
	if copied > 0 {
		fmt.Printf("✓ agents/ ディレクトリに %d 個のファイルを追加しました\n", copied)
	}
	if skipped > 0 {
		fmt.Printf("  (%d 個の既存ファイルはスキップしました)\n", skipped)
	}

	// queue/ ディレクトリ構造を作成
	queueSubDirs := []string{"inbox", "tasks", "reports"}
	createdDirs := 0
	for _, subDir := range queueSubDirs {
		dirPath := filepath.Join(queueDir, subDir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return fmt.Errorf("queue/%s/ の作成に失敗: %w", subDir, err)
			}
			createdDirs++
		}
	}
	if createdDirs > 0 {
		fmt.Printf("✓ queue/ ディレクトリ構造を作成しました\n")
	}

	// .gitignore に queue/ を追加
	if err := updateGitignore(currentDir); err != nil {
		fmt.Printf("警告: .gitignore の更新に失敗しました: %v\n", err)
	} else {
		fmt.Printf("✓ .gitignore に queue/ を追加しました\n")
	}

	fmt.Println("\n初期化が完了しました！")
	fmt.Println("\n次のステップ:")
	fmt.Println("  1. agents/ ディレクトリ内の各エージェントの .claude/settings.local.json を確認")
	fmt.Println("  2. bastion start でオーケストレーターを起動")

	return nil
}

// 埋め込まれたディレクトリを再帰的にコピー（no-clobber モード）
// 既存のファイルは上書きせず、新しいファイルのみを追加
// 返り値: (コピーしたファイル数, スキップしたファイル数, エラー)
func copyEmbeddedDirNoClobber(fsys fs.FS, srcPath, dstPath string) (int, int, error) {
	copied := 0
	skipped := 0

	// プロジェクトルートを取得
	projectRoot, err := os.Getwd()
	if err != nil {
		return 0, 0, fmt.Errorf("プロジェクトルートの取得に失敗: %w", err)
	}

	err = fs.WalkDir(fsys, srcPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 相対パスを計算
		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dstPath, relPath)

		if d.IsDir() {
			// ディレクトリを作成（既に存在する場合はスキップ）
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				if err := os.MkdirAll(targetPath, 0755); err != nil {
					return err
				}
			}
			return nil
		}

		// .template ファイルの場合、拡張子を削除
		isTemplate := strings.HasSuffix(targetPath, ".template")
		if isTemplate {
			targetPath = strings.TrimSuffix(targetPath, ".template")
		}

		// ファイルが既に存在するかチェック
		if _, err := os.Stat(targetPath); err == nil {
			// 既存ファイルはスキップ
			skipped++
			return nil
		}

		// ファイルをコピー
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		// .template ファイルの場合、プレースホルダーを置換
		if isTemplate {
			content := string(data)
			content = strings.ReplaceAll(content, "{{PROJECT_ROOT}}", projectRoot)
			data = []byte(content)
		}

		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return err
		}

		copied++
		return nil
	})

	return copied, skipped, err
}

// .gitignore に queue/ を追加
func updateGitignore(projectDir string) error {
	gitignorePath := filepath.Join(projectDir, ".gitignore")

	// .gitignore が存在するか確認
	var content []byte
	if _, err := os.Stat(gitignorePath); err == nil {
		// 既存の .gitignore を読み取り
		content, err = os.ReadFile(gitignorePath)
		if err != nil {
			return err
		}
	}

	// queue/ が既に記載されているかチェック
	contentStr := string(content)
	if strings.Contains(contentStr, "queue/") {
		// 既に記載されている
		return nil
	}

	// queue/ を追加
	newContent := contentStr
	if len(newContent) > 0 && newContent[len(newContent)-1] != '\n' {
		newContent += "\n"
	}
	newContent += "# Bastion runtime directories\n"
	newContent += "queue/\n"

	return os.WriteFile(gitignorePath, []byte(newContent), 0644)
}
