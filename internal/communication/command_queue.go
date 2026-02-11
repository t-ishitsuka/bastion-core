package communication

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"gopkg.in/yaml.v3"
)

// 指令キューの読み書きを管理する
type CommandQueueManager struct {
	queueDir string
	tasksDir string
	mu       sync.Mutex
}

// 新しい指令キューマネージャーを作成
func NewCommandQueueManager(queueDir string) *CommandQueueManager {
	return &CommandQueueManager{
		queueDir: queueDir,
		tasksDir: filepath.Join(queueDir, "tasks"),
	}
}

// 指令を個別ファイルとして書き込む
func (m *CommandQueueManager) Write(cmd Command) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// tasks ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(m.tasksDir, 0755); err != nil {
		return fmt.Errorf("failed to create tasks directory: %w", err)
	}

	// タスクファイルのパス: queue/tasks/<id>.yaml
	taskPath := filepath.Join(m.tasksDir, fmt.Sprintf("%s.yaml", cmd.ID))

	// タスクを YAML にシリアライズ
	data, err := yaml.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	// ファイルに書き込む
	if err := os.WriteFile(taskPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}

// すべての指令を読み込む
func (m *CommandQueueManager) Read() ([]Command, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// tasks ディレクトリが存在しない場合は空のスライスを返す
	if _, err := os.Stat(m.tasksDir); os.IsNotExist(err) {
		return []Command{}, nil
	}

	// tasks ディレクトリ内のすべての .yaml ファイルを取得
	entries, err := os.ReadDir(m.tasksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks directory: %w", err)
	}

	var commands []Command
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		taskPath := filepath.Join(m.tasksDir, entry.Name())
		cmd, err := m.readTaskFile(taskPath)
		if err != nil {
			// エラーをログに記録するが、処理は継続
			fmt.Fprintf(os.Stderr, "Warning: failed to read task file %s: %v\n", taskPath, err)
			continue
		}

		commands = append(commands, *cmd)
	}

	// タイムスタンプでソート（古い順）
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Timestamp.Before(commands[j].Timestamp)
	})

	return commands, nil
}

// 特定の指令を読み込む
func (m *CommandQueueManager) ReadByID(id string) (*Command, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	taskPath := filepath.Join(m.tasksDir, fmt.Sprintf("%s.yaml", id))
	return m.readTaskFile(taskPath)
}

// タスクファイルを読み込む
func (m *CommandQueueManager) readTaskFile(path string) (*Command, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cmd Command
	if err := yaml.Unmarshal(data, &cmd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal command: %w", err)
	}

	return &cmd, nil
}

// 指令の状態を更新
func (m *CommandQueueManager) UpdateStatus(id string, status CommandStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	taskPath := filepath.Join(m.tasksDir, fmt.Sprintf("%s.yaml", id))

	// 既存のタスクを読み込む
	cmd, err := m.readTaskFile(taskPath)
	if err != nil {
		return fmt.Errorf("failed to read task: %w", err)
	}

	// 状態を更新
	cmd.Status = status

	// タスクを YAML にシリアライズ
	data, err := yaml.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	// ファイルに書き込む
	if err := os.WriteFile(taskPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}

// 指令を削除
func (m *CommandQueueManager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	taskPath := filepath.Join(m.tasksDir, fmt.Sprintf("%s.yaml", id))

	if err := os.Remove(taskPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 既に削除されている場合はエラーとしない
		}
		return fmt.Errorf("failed to delete task file: %w", err)
	}

	return nil
}
