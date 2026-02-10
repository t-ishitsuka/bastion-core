package communication

import "time"

// メッセージ種類
type MessageType string

const (
	// タスクが割り当てられた
	MessageTypeTaskAssigned MessageType = "task_assigned"
	// レポートを受信した
	MessageTypeReportReceived MessageType = "report_received"
	// 起床通知
	MessageTypeWakeUp MessageType = "wake_up"
)

// メッセージ処理状態
type MessageStatus string

const (
	// 未処理
	MessageStatusPending MessageStatus = "pending"
	// 処理済み
	MessageStatusProcessed MessageStatus = "processed"
)

// inbox エントリ
type Message struct {
	ID        string        `yaml:"id"`
	Timestamp time.Time     `yaml:"timestamp"`
	From      string        `yaml:"from"`
	Type      MessageType   `yaml:"type"`
	Message   string        `yaml:"message"`
	Status    MessageStatus `yaml:"status"`
}

// inbox ファイルの内容
type Inbox struct {
	Messages []Message `yaml:"messages"`
}
