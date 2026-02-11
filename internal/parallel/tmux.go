package parallel

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	// セッション名
	SessionName = "bastion"

	// ウィンドウ名
	WindowEnvoy       = "envoy"
	WindowMarshall    = "marshall"
	WindowSpecialists = "specialists"
	WindowWatcher     = "watcher"
)

// tmux セッションを管理
type SessionManager struct {
	sessionName string
}

// 新しい SessionManager を作成
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessionName: SessionName,
	}
}

// セッションが存在するか確認
func (sm *SessionManager) SessionExists() (bool, error) {
	cmd := exec.Command("tmux", "has-session", "-t", sm.sessionName)
	err := cmd.Run()
	if err != nil {
		// 終了コード 1 はセッションが存在しない
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check session: %w", err)
	}
	return true, nil
}

// セッションを作成
func (sm *SessionManager) CreateSession() error {
	exists, err := sm.SessionExists()
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("session already exists: %s", sm.sessionName)
	}

	// デタッチモードでセッションを作成
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sm.sessionName, "-n", WindowEnvoy)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// ウィンドウを作成
func (sm *SessionManager) CreateWindow(name string) error {
	target := fmt.Sprintf("%s:", sm.sessionName)
	cmd := exec.Command("tmux", "new-window", "-t", target, "-n", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create window: %w", err)
	}
	return nil
}

// ペインを分割（水平）
func (sm *SessionManager) SplitPaneHorizontal(window string) error {
	target := fmt.Sprintf("%s:%s", sm.sessionName, window)
	cmd := exec.Command("tmux", "split-window", "-h", "-t", target)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to split pane horizontally: %w", err)
	}
	return nil
}

// ペインを分割（垂直）
func (sm *SessionManager) SplitPaneVertical(window string) error {
	target := fmt.Sprintf("%s:%s", sm.sessionName, window)
	cmd := exec.Command("tmux", "split-window", "-v", "-t", target)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to split pane vertically: %w", err)
	}
	return nil
}

// コマンドを送信
func (sm *SessionManager) SendKeys(target, keys string, enter bool) error {
	fullTarget := fmt.Sprintf("%s:%s", sm.sessionName, target)

	// exec.Command は引数を適切にエスケープするため、-l フラグは不要
	// すべてのテキスト（シェルコマンドと単純なテキストの両方）を同じ方法で送信
	args := []string{"send-keys", "-t", fullTarget, keys}

	cmd := exec.Command("tmux", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys: %w", err)
	}

	// Enter キーが必要な場合は別途送信
	if enter {
		enterArgs := []string{"send-keys", "-t", fullTarget, "Enter"}
		enterCmd := exec.Command("tmux", enterArgs...)
		if err := enterCmd.Run(); err != nil {
			return fmt.Errorf("failed to send enter key: %w", err)
		}
	}

	return nil
}

// ウィンドウ一覧を取得
func (sm *SessionManager) ListWindows() ([]string, error) {
	cmd := exec.Command("tmux", "list-windows", "-t", sm.sessionName, "-F", "#{window_name}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list windows: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return lines, nil
}

// ペイン一覧を取得
func (sm *SessionManager) ListPanes(window string) ([]string, error) {
	target := fmt.Sprintf("%s:%s", sm.sessionName, window)
	cmd := exec.Command("tmux", "list-panes", "-t", target, "-F", "#{pane_index}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list panes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return lines, nil
}

// ペインタイトルを設定（カスタム属性を使用して上書き防止）
func (sm *SessionManager) SetPaneTitle(target, title string) error {
	fullTarget := fmt.Sprintf("%s:%s", sm.sessionName, target)

	// カスタムペイン属性を設定（アプリケーションに上書きされない）
	cmd := exec.Command("tmux", "set-option", "-p", "-t", fullTarget, "@pane_label", title)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set pane label: %w", err)
	}

	return nil
}

// ペインボーダーの表示設定を有効化
func (sm *SessionManager) EnablePaneBorders() error {
	// ペインボーダーにタイトルを表示
	cmd1 := exec.Command("tmux", "set-option", "-g", "pane-border-status", "top")
	if err := cmd1.Run(); err != nil {
		return fmt.Errorf("failed to enable pane border status: %w", err)
	}

	// ペインボーダーのフォーマットを設定（カスタム属性を使用）
	// #{@pane_label} はアプリケーションに上書きされないカスタム属性
	cmd2 := exec.Command("tmux", "set-option", "-g", "pane-border-format", " #{@pane_label} ")
	if err := cmd2.Run(); err != nil {
		return fmt.Errorf("failed to set pane border format: %w", err)
	}

	return nil
}

// カスタムキーバインドを設定
func (sm *SessionManager) SetupKeyBindings(bastionCmd string) error {
	// Ctrl+b q で確認付き停止
	// confirm-before で "Stop Bastion session? (y/n)" を表示し、y で bastion stop を実行
	cmd := exec.Command("tmux", "bind-key", "-T", "prefix", "q",
		"confirm-before", "-p", "Stop Bastion session? (y/n)",
		fmt.Sprintf("run-shell '%s stop'", bastionCmd))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set key binding: %w", err)
	}

	return nil
}

// セッションを停止
func (sm *SessionManager) KillSession() error {
	exists, err := sm.SessionExists()
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	cmd := exec.Command("tmux", "kill-session", "-t", sm.sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill session: %w", err)
	}
	return nil
}

// ペインのサイズを変更
func (sm *SessionManager) ResizePane(target string, size int) error {
	fullTarget := fmt.Sprintf("%s:%s", sm.sessionName, target)
	cmd := exec.Command("tmux", "resize-pane", "-t", fullTarget, "-x", fmt.Sprintf("%d%%", size))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to resize pane: %w", err)
	}
	return nil
}

// ペインを選択
func (sm *SessionManager) SelectPane(target string) error {
	fullTarget := fmt.Sprintf("%s:%s", sm.sessionName, target)
	cmd := exec.Command("tmux", "select-pane", "-t", fullTarget)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to select pane: %w", err)
	}
	return nil
}

// tiled レイアウトを適用
func (sm *SessionManager) SetTiledLayout(window string) error {
	target := fmt.Sprintf("%s:%s", sm.sessionName, window)
	cmd := exec.Command("tmux", "select-layout", "-t", target, "tiled")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set tiled layout: %w", err)
	}
	return nil
}

// Bastion セッションをセットアップ
func SetupBastionSession() error {
	sm := NewSessionManager()

	// 既存セッションがあれば削除
	if exists, _ := sm.SessionExists(); exists {
		if err := sm.KillSession(); err != nil {
			return err
		}
	}

	// セッション作成（メインウィンドウが自動作成される）
	// ウィンドウ名を "main" に設定
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sm.sessionName, "-n", "main")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// ウィンドウ1（メイン）のレイアウトを作成
	// 左: Envoy (50%)、右上: Watcher、右下: Marshall

	// 水平分割で右側を作成（左右に分割）
	if err := sm.SplitPaneHorizontal("main"); err != nil {
		return err
	}

	// 右側のペイン（pane 1）を選択して垂直分割（上下に分割）
	if err := sm.SelectPane("main.1"); err != nil {
		return err
	}

	if err := sm.SplitPaneVertical("main"); err != nil {
		return err
	}

	// specialists ウィンドウを作成
	if err := sm.CreateWindow(WindowSpecialists); err != nil {
		return err
	}

	// 初期表示は Envoy ペイン（main.0）を選択
	if err := sm.SelectPane("main.0"); err != nil {
		return err
	}

	// main ウィンドウを選択（specialists ウィンドウではなく）
	selectCmd := exec.Command("tmux", "select-window", "-t", fmt.Sprintf("%s:main", sm.sessionName))
	if err := selectCmd.Run(); err != nil {
		return fmt.Errorf("failed to select main window: %w", err)
	}

	return nil
}

// Specialists ウィンドウをグリッドレイアウトでセットアップ
func SetupSpecialistsGrid(numSpecialists int) error {
	sm := NewSessionManager()

	if numSpecialists <= 1 {
		// 1つだけの場合は分割不要
		return nil
	}

	// 最初のペインは既に存在するので、残りを作成
	for i := 1; i < numSpecialists; i++ {
		if err := sm.SplitPaneHorizontal(WindowSpecialists); err != nil {
			return err
		}
	}

	// tiled レイアウトを適用してグリッド状に配置
	if err := sm.SetTiledLayout(WindowSpecialists); err != nil {
		return err
	}

	return nil
}
