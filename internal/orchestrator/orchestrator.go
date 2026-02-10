package orchestrator

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/t-ishitsuka/bastion-core/internal/communication"
	"github.com/t-ishitsuka/bastion-core/internal/parallel"
)

// エージェントの種類
const (
	AgentEnvoy      = "envoy"
	AgentMarshall   = "marshall"
	AgentSpecialist = "specialist"
)

// オーケストレーター
type Orchestrator struct {
	sm              *parallel.SessionManager
	projectRoot     string
	agentsDir       string
	queueDir        string
	specialistCount int
	watcher         *communication.Watcher
}

// 新しい Orchestrator を作成
func NewOrchestrator(projectRoot string, specialistCount int) *Orchestrator {
	return &Orchestrator{
		sm:              parallel.NewSessionManager(),
		projectRoot:     projectRoot,
		agentsDir:       filepath.Join(projectRoot, "agents"),
		queueDir:        filepath.Join(projectRoot, "queue"),
		specialistCount: specialistCount,
	}
}

// すべてのエージェントを起動
func (o *Orchestrator) StartAll() error {
	// Envoy を起動
	if err := o.StartAgent(AgentEnvoy, parallel.WindowEnvoy, 0); err != nil {
		return fmt.Errorf("failed to start envoy: %w", err)
	}

	// Marshall を起動
	if err := o.StartAgent(AgentMarshall, parallel.WindowMarshall, 0); err != nil {
		return fmt.Errorf("failed to start marshall: %w", err)
	}

	// Specialists を起動
	for i := 1; i <= o.specialistCount; i++ {
		if i > 1 {
			// 2つ目以降はペインを分割
			if err := o.sm.SplitPaneHorizontal(parallel.WindowSpecialists); err != nil {
				return fmt.Errorf("failed to split pane for specialist %d: %w", i, err)
			}
		}

		// Specialist を起動
		target := fmt.Sprintf("%s.%d", parallel.WindowSpecialists, i-1)
		if err := o.StartAgent(AgentSpecialist, target, i); err != nil {
			return fmt.Errorf("failed to start specialist %d: %w", i, err)
		}
	}

	return nil
}

// 個別のエージェントを起動
func (o *Orchestrator) StartAgent(agentType, target string, index int) error {
	// エージェントディレクトリのパスを構築
	agentDir := filepath.Join(o.agentsDir, agentType)

	// claude コマンドを構築
	// エージェントディレクトリに移動してから claude を起動
	// claude は現在のディレクトリの CLAUDE.md を自動的に読み込む
	// --add-dir でプロジェクトルートへのアクセスを許可
	cmd := fmt.Sprintf("cd %s && claude --add-dir %s",
		agentDir,
		o.projectRoot,
	)

	// tmux send-keys でコマンドを送信
	if err := o.sm.SendKeys(target, cmd, true); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	return nil
}

// すべてのエージェントに wakeup を送信
func (o *Orchestrator) WakeupAll() error {
	// Envoy を wakeup
	if err := o.Wakeup(AgentEnvoy, parallel.WindowEnvoy); err != nil {
		return fmt.Errorf("failed to wakeup envoy: %w", err)
	}

	// Marshall を wakeup
	if err := o.Wakeup(AgentMarshall, parallel.WindowMarshall); err != nil {
		return fmt.Errorf("failed to wakeup marshall: %w", err)
	}

	return nil
}

// 個別のエージェントを wakeup
func (o *Orchestrator) Wakeup(agentType, target string) error {
	// 簡単な wakeup メッセージを送信（Enter キーを押すだけ）
	if err := o.sm.SendKeys(target, "", true); err != nil {
		return fmt.Errorf("failed to wakeup %s: %w", agentType, err)
	}
	return nil
}

// inbox 監視を開始
func (o *Orchestrator) StartWatcher() error {
	// watcher を作成して inbox ディレクトリを監視
	watcher, err := communication.WatchInbox(o.queueDir)
	if err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	o.watcher = watcher

	// バックグラウンドでイベントを処理
	go o.processWatcherEvents()

	return nil
}

// watcher イベントを処理
func (o *Orchestrator) processWatcherEvents() {
	for {
		select {
		case event, ok := <-o.watcher.Events():
			if !ok {
				return
			}

			// inbox ファイルが変更された場合、該当エージェントに通知
			if err := o.handleInboxChange(event.Path); err != nil {
				log.Printf("failed to handle inbox change: %v", err)
			}

		case err, ok := <-o.watcher.Errors():
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

// inbox 変更を処理
func (o *Orchestrator) handleInboxChange(path string) error {
	// ファイル名から対象エージェントを特定
	// 例: queue/inbox/marshall.yaml -> marshall
	base := filepath.Base(path)
	target := base[:len(base)-len(filepath.Ext(base))]

	// エージェントタイプに応じて wakeup
	switch target {
	case AgentEnvoy:
		return o.Wakeup(AgentEnvoy, parallel.WindowEnvoy)
	case AgentMarshall:
		return o.Wakeup(AgentMarshall, parallel.WindowMarshall)
	default:
		// Specialist の場合はスキップ（将来実装）
		return nil
	}
}

// watcher を停止
func (o *Orchestrator) StopWatcher() error {
	if o.watcher == nil {
		return nil
	}
	return o.watcher.Stop()
}
