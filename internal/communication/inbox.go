package communication

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// inbox の読み書きを管理する
type InboxManager struct {
	queueDir string
	mu       sync.Mutex
}

// 新しい inbox マネージャーを作成
func NewInboxManager(queueDir string) *InboxManager {
	return &InboxManager{
		queueDir: queueDir,
	}
}

// メッセージを inbox に書き込む
func (m *InboxManager) Write(target, message string, msgType MessageType, from string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inboxPath := filepath.Join(m.queueDir, "inbox", target+".yaml")

	// inbox ファイルを読み込む（存在しない場合は空の Inbox を作成）
	inbox, err := m.readInboxFile(inboxPath)
	if err != nil {
		if os.IsNotExist(err) {
			inbox = &Inbox{Messages: []Message{}}
		} else {
			return fmt.Errorf("failed to read inbox: %w", err)
		}
	}

	// 新しいメッセージを追加
	msg := Message{
		ID:        fmt.Sprintf("msg_%d", time.Now().Unix()),
		Timestamp: time.Now(),
		From:      from,
		Type:      msgType,
		Message:   message,
		Status:    MessageStatusPending,
	}
	inbox.Messages = append(inbox.Messages, msg)

	// inbox ファイルに書き込む
	if err := m.writeInboxFile(inboxPath, inbox); err != nil {
		return fmt.Errorf("failed to write inbox: %w", err)
	}

	return nil
}

// inbox からメッセージを読み込む
func (m *InboxManager) Read(target string) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	inboxPath := filepath.Join(m.queueDir, "inbox", target+".yaml")

	inbox, err := m.readInboxFile(inboxPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Message{}, nil
		}
		return nil, fmt.Errorf("failed to read inbox: %w", err)
	}

	return inbox.Messages, nil
}

// メッセージを処理済みにする
func (m *InboxManager) MarkAsProcessed(target, messageID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inboxPath := filepath.Join(m.queueDir, "inbox", target+".yaml")

	inbox, err := m.readInboxFile(inboxPath)
	if err != nil {
		return fmt.Errorf("failed to read inbox: %w", err)
	}

	// メッセージを検索して処理済みにする
	found := false
	for i := range inbox.Messages {
		if inbox.Messages[i].ID == messageID {
			inbox.Messages[i].Status = MessageStatusProcessed
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("message not found: %s", messageID)
	}

	// inbox ファイルに書き込む
	if err := m.writeInboxFile(inboxPath, inbox); err != nil {
		return fmt.Errorf("failed to write inbox: %w", err)
	}

	return nil
}

// 未処理のメッセージを取得
func (m *InboxManager) GetPendingMessages(target string) ([]Message, error) {
	messages, err := m.Read(target)
	if err != nil {
		return nil, err
	}

	pending := []Message{}
	for _, msg := range messages {
		if msg.Status == MessageStatusPending {
			pending = append(pending, msg)
		}
	}

	return pending, nil
}

// inbox ファイルを読み込む
func (m *InboxManager) readInboxFile(path string) (*Inbox, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var inbox Inbox
	if err := yaml.Unmarshal(data, &inbox); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inbox: %w", err)
	}

	return &inbox, nil
}

// inbox ファイルに書き込む
func (m *InboxManager) writeInboxFile(path string, inbox *Inbox) error {
	// ディレクトリが存在しない場合は作成
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := yaml.Marshal(inbox)
	if err != nil {
		return fmt.Errorf("failed to marshal inbox: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
