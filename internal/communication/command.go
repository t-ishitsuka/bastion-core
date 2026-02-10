package communication

import "time"

// コマンド状態
type CommandStatus string

const (
	// 未処理
	CommandStatusPending CommandStatus = "pending"
	// 処理中
	CommandStatusInProgress CommandStatus = "in_progress"
	// 完了
	CommandStatusCompleted CommandStatus = "completed"
	// 失敗
	CommandStatusFailed CommandStatus = "failed"
)

// Envoy から Marshall への指令
type Command struct {
	ID                 string        `yaml:"id"`
	Timestamp          time.Time     `yaml:"timestamp"`
	Purpose            string        `yaml:"purpose"`
	AcceptanceCriteria []string      `yaml:"acceptance_criteria"`
	Command            string        `yaml:"command"`
	Project            string        `yaml:"project"`
	Priority           string        `yaml:"priority"`
	Status             CommandStatus `yaml:"status"`
}

// 指令キュー
type CommandQueue struct {
	Commands []Command `yaml:"commands"`
}
