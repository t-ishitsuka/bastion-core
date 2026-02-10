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
	WindowEnvoy      = "envoy"
	WindowMarshall   = "marshall"
	WindowSpecialists = "specialists"
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
	args := []string{"send-keys", "-t", fullTarget, keys}
	if enter {
		args = append(args, "Enter")
	}

	cmd := exec.Command("tmux", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send keys: %w", err)
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

// Bastion セッションをセットアップ
func SetupBastionSession() error {
	sm := NewSessionManager()

	// 既存セッションがあれば削除
	if exists, _ := sm.SessionExists(); exists {
		if err := sm.KillSession(); err != nil {
			return err
		}
	}

	// セッション作成（envoy ウィンドウが自動作成される）
	if err := sm.CreateSession(); err != nil {
		return err
	}

	// marshall ウィンドウを作成
	if err := sm.CreateWindow(WindowMarshall); err != nil {
		return err
	}

	// specialists ウィンドウを作成
	if err := sm.CreateWindow(WindowSpecialists); err != nil {
		return err
	}

	return nil
}
