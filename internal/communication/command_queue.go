package communication

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// 指令キューの読み書きを管理する
type CommandQueueManager struct {
	queueDir string
	mu       sync.Mutex
}

// 新しい指令キューマネージャーを作成
func NewCommandQueueManager(queueDir string) *CommandQueueManager {
	return &CommandQueueManager{
		queueDir: queueDir,
	}
}

// 指令を queue に書き込む
func (m *CommandQueueManager) Write(cmd Command) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	commandPath := filepath.Join(m.queueDir, "envoy_to_marshall.yaml")

	// 指令ファイルを読み込む（存在しない場合は空の CommandQueue を作成）
	queue, err := m.readCommandFile(commandPath)
	if err != nil {
		if os.IsNotExist(err) {
			queue = &CommandQueue{Commands: []Command{}}
		} else {
			return fmt.Errorf("failed to read command queue: %w", err)
		}
	}

	// 新しい指令を追加
	queue.Commands = append(queue.Commands, cmd)

	// 指令ファイルに書き込む
	if err := m.writeCommandFile(commandPath, queue); err != nil {
		return fmt.Errorf("failed to write command queue: %w", err)
	}

	return nil
}

// 指令を読み込む
func (m *CommandQueueManager) Read() ([]Command, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	commandPath := filepath.Join(m.queueDir, "envoy_to_marshall.yaml")

	queue, err := m.readCommandFile(commandPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Command{}, nil
		}
		return nil, fmt.Errorf("failed to read command queue: %w", err)
	}

	return queue.Commands, nil
}

// 指令ファイルを読み込む
func (m *CommandQueueManager) readCommandFile(path string) (*CommandQueue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var queue CommandQueue
	if err := yaml.Unmarshal(data, &queue); err != nil {
		return nil, fmt.Errorf("failed to unmarshal command queue: %w", err)
	}

	return &queue, nil
}

// 指令ファイルに書き込む
func (m *CommandQueueManager) writeCommandFile(path string, queue *CommandQueue) error {
	// ディレクトリが存在しない場合は作成
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := yaml.Marshal(queue)
	if err != nil {
		return fmt.Errorf("failed to marshal command queue: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
