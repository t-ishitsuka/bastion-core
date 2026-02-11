package orchestrator

import (
	"fmt"
	"log"
	"os/exec"
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
	// Envoy を起動（メインウィンドウの左ペイン）
	if err := o.StartAgent(AgentEnvoy, "main.0", 0); err != nil {
		return fmt.Errorf("failed to start envoy: %w", err)
	}

	// Marshall を起動（メインウィンドウの右下ペイン）
	if err := o.StartAgent(AgentMarshall, "main.2", 0); err != nil {
		return fmt.Errorf("failed to start marshall: %w", err)
	}

	// Specialists ウィンドウにグリッドレイアウトをセットアップ
	if err := parallel.SetupSpecialistsGrid(o.specialistCount); err != nil {
		return fmt.Errorf("failed to setup specialists grid: %w", err)
	}

	// Specialists を起動
	for i := 1; i <= o.specialistCount; i++ {
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
	// ファイルアクセス権限は .claude/settings.local.json で管理
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
	// Envoy を wakeup（メインウィンドウの左ペイン）
	if err := o.Wakeup(AgentEnvoy, "main.0"); err != nil {
		return fmt.Errorf("failed to wakeup envoy: %w", err)
	}

	// Marshall を wakeup（メインウィンドウの右下ペイン）
	if err := o.Wakeup(AgentMarshall, "main.2"); err != nil {
		return fmt.Errorf("failed to wakeup marshall: %w", err)
	}

	return nil
}

// 個別のエージェントを wakeup
func (o *Orchestrator) Wakeup(agentType, target string) error {
	// inbox チェックを促す具体的なメッセージを送信
	// エージェントは "inbox" というメッセージを受け取ったら inbox をチェックする
	if err := o.sm.SendKeys(target, "inbox", true); err != nil {
		return fmt.Errorf("failed to wakeup %s: %w", agentType, err)
	}
	return nil
}

// エスカレーション付きで wakeup（応答しないエージェント向け）
func (o *Orchestrator) WakeupWithEscalation(agentType, target string, phase int) error {
	switch phase {
	case 1:
		// Phase 1: 通常の inbox チェック要求
		return o.Wakeup(agentType, target)
	case 2:
		// Phase 2: Escape×2 でカーソル位置をリセット + inbox チェック
		if err := o.sm.SendKeys(target, "Escape", false); err != nil {
			return fmt.Errorf("failed to send escape: %w", err)
		}
		if err := o.sm.SendKeys(target, "Escape", false); err != nil {
			return fmt.Errorf("failed to send escape: %w", err)
		}
		return o.Wakeup(agentType, target)
	case 3:
		// Phase 3: /clear でセッションを強制リセット
		if err := o.sm.SendKeys(target, "/clear", true); err != nil {
			return fmt.Errorf("failed to send /clear: %w", err)
		}
		return nil
	default:
		return o.Wakeup(agentType, target)
	}
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
	log.Println("[watcher] イベント処理を開始しました")
	for {
		select {
		case event, ok := <-o.watcher.Events():
			if !ok {
				log.Println("[watcher] イベントチャンネルがクローズされました")
				return
			}

			log.Printf("[watcher] イベント受信: %s (%s)", event.Path, event.Operation)

			// inbox ファイルが変更された場合、該当エージェントに通知
			if err := o.handleInboxChange(event.Path); err != nil {
				log.Printf("[watcher] inbox 変更処理エラー: %v", err)
			}

		case err, ok := <-o.watcher.Errors():
			if !ok {
				log.Println("[watcher] エラーチャンネルがクローズされました")
				return
			}
			log.Printf("[watcher] エラー: %v", err)
		}
	}
}

// inbox 変更を処理
func (o *Orchestrator) handleInboxChange(path string) error {
	// ファイル名から対象エージェントを特定
	// 例: queue/inbox/marshall.yaml -> marshall
	base := filepath.Base(path)
	target := base[:len(base)-len(filepath.Ext(base))]

	log.Printf("[watcher] inbox 変更検知: %s -> エージェント: %s", path, target)

	// エージェントタイプに応じて wakeup
	switch target {
	case AgentEnvoy:
		log.Printf("[watcher] Envoy に wakeup を送信")
		return o.Wakeup(AgentEnvoy, "main.0")
	case AgentMarshall:
		log.Printf("[watcher] Marshall に wakeup を送信")
		return o.Wakeup(AgentMarshall, "main.2")
	default:
		// Specialist の場合はスキップ（将来実装）
		log.Printf("[watcher] %s はスキップ（未実装）", target)
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

// watcher ペインで bastion watch コマンドを起動
func (o *Orchestrator) StartWatcherWindow() error {
	// bastion コマンドのパスを決定
	// 1. パスに通っている bastion コマンドを確認
	// 2. なければプロジェクトルートの ./bastion を使用
	bastionCmd := "bastion"
	if _, err := exec.LookPath("bastion"); err != nil {
		// コマンドが見つからない場合、プロジェクトルートのバイナリを使用
		bastionCmd = filepath.Join(o.projectRoot, "bastion")
	}

	// bastion watch コマンドを構築
	cmd := fmt.Sprintf("cd %s && %s watch", o.projectRoot, bastionCmd)

	// tmux send-keys でコマンドを送信（メインウィンドウの右上ペイン）
	if err := o.sm.SendKeys("main.1", cmd, true); err != nil {
		return fmt.Errorf("failed to send watch command: %w", err)
	}

	return nil
}
