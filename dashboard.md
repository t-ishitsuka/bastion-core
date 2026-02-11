# Bastion Dashboard

## 現在の状態

- 状態: Phase 1 完了、テスト修正完了、Marshall 機能拡張完了
- 最終更新: 2026-02-10T17:15:00

## 進行中のタスク

### cmd_010: Specialistたちに「あいうえお」でお父さんスイッチ

- 優先度: low
- 目的: 動作確認（Envoy → Marshall → Specialist の通信テスト）
- 開始: 2026-02-11T12:16:00+09:00

**割り当てたサブタスク:**
- cmd_010_subtask_001: specialist_1 (Senior Backend Engineer) [割当済]
- cmd_010_subtask_002: specialist_2 (Frontend Developer) [割当済]
- cmd_010_subtask_003: specialist_3 (DevOps Engineer) [割当済]
- cmd_010_subtask_004: specialist_4 (Data Scientist) [割当済]

**期待される成果:**
各 Specialist が「あいうえお」を使った創造的な反応を返す（お父さんスイッチ風）

**テスト待ち:**
- cmd_006 & cmd_008: watcher の動作確認
  - bastion をビルド→再起動→inbox ファイル変更テスト
  - Marshall ウィンドウに "inbox" が自動送信されることを確認

## 完了したタスク

### cmd_001: Phase 1 実装のコミットと PR 作成

- タスク #1: watcher 起動機能を start コマンドに統合 [完了]
- タスク #2: Phase 1 実装をコミット [完了]
- タスク #3: PR を作成 [完了]

**成果物:**

- コミット: b37f69d - feat: Phase 1 基盤実装（MVP）完了
- PR: https://github.com/t-ishitsuka/bastion-core/pull/1

### cmd_002: GitHub Actions テスト修正

- タスク #1: tmux テストの CI 環境対応を確認 [完了]

**成果物:**

- internal/parallel/tmux_test.go:11-17 - CI 環境チェック実装済みを確認
- cmd/bastion/cmd/start_test.go:13-19 - CI 環境チェック実装済みを確認

### cmd_003: Marshall の指令認識機能実装

- タスク #1: agents/marshall/CLAUDE.md に指令キュー監視機能を追加 [完了]
- タスク #2: 起動時処理フローを追加 [完了]
- タスク #3: 指令フォーマットのドキュメント化 [完了]

**成果物:**

- agents/marshall/CLAUDE.md 更新
  - 起動時に envoy_to_marshall.yaml をチェックする機能追加
  - status が "pending" の指令を自動処理
  - 指令フォーマットのドキュメント追加

### cmd_007: Envoy CLAUDE.md の inbox 処理方法を修正

- 完了: 2026-02-10T19:05:00+09:00
- 優先度: critical

**実施した修正:**
- agents/envoy/CLAUDE.md を更新
  - line 110: read: false → status: pending に修正
  - line 120-127: inbox チェックの処理セクションを更新
  - GetPendingMessages() の使用方法を記載
  - MarkAsProcessed() の使用方法を記載
  - 実装例を追加

**成果物:**

Envoy エージェントが inbox メッセージを正しく処理できるようになりました：
- 未処理メッセージを status: pending で検出
- 処理済みメッセージを status: processed でマーク
- セッション再開時に重複処理を回避

### cmd_004 & cmd_005: エージェント間通信の自動承認機能と検証

- 完了: 2026-02-10T19:10:00+09:00
- 優先度: critical (cmd_004), high (cmd_005)

**検証結果:**
- Envoy: queue/, inbox/, dashboard.md のアクセス設定 [OK]
- Marshall: queue/, dashboard.md, knowledge/ の完全アクセス設定 [OK]
- Specialist: queue/tasks/, reports/, inbox/ のアクセス設定 [OK]

**成果物:**

各エージェントの .claude/settings.local.json が正しく配置され、以下を実現：
- Envoy → Marshall への指令送信が承認不要
- Marshall → Specialist へのタスク割当が承認不要
- dashboard.md, knowledge/ へのアクセスが承認不要
- エージェント間の自動連携が実現

### cmd_006 & cmd_008: watcher 自動起動機能と動作修正

- 完了: 2026-02-10T19:15:00+09:00
- 優先度: critical

**実装確認:**
1. start.go:66-72 - StartWatcherWindow() で watcher ウィンドウを起動
2. orchestrator.go:147-160 - StartWatcher() で inbox 監視を開始
3. orchestrator.go:186-202 - handleInboxChange() でエージェントに nudge 送信

**cmd_008 で追加修正:**
1. internal/communication/watcher.go:62-93
   - WRITE, CREATE, RENAME イベントを監視するように拡張
   - WSL2 環境でのエディタ保存方法（rename 方式）に対応
2. internal/orchestrator/orchestrator.go:163-183
   - processWatcherEvents() にデバッグログを追加
   - handleInboxChange() にデバッグログを追加
   - イベント受信・処理の各ステップでログ出力

**成果物:**

watcher の完全な実装：
- bastion start で自動起動
- inbox/ ディレクトリの変更を検知
- エージェントに自動的に nudge 送信
- デバッグログで動作状況を確認可能

**次のステップ:**
動作確認（bastion をビルド→再起動→inbox テスト）

## Phase 1 実装内容

### 通信層
- InboxManager: メッセージの読み書き管理
- Watcher: fsnotify によるファイル監視
- メッセージフォーマット定義

### tmux セッション管理
- SessionManager: tmux 操作の抽象化
- ウィンドウ・ペイン管理

### CLI コマンド
- doctor: 環境チェック
- start: セッション起動 + watcher 起動
- status: 状態確認
- stop: セッション停止

### Orchestrator
- エージェント起動・管理
- watcher 統合（inbox 変更検知 → エージェント通知）

## 知識ベース更新

- watcher 起動機能の統合パターン
  - orchestrator に watcher フィールドを追加
  - StartWatcher メソッドで inbox ディレクトリを監視
  - ファイル変更検知時にエージェントに nudge 送信

- CI 環境での tmux テスト対応パターン
  - os.Getenv("CI") で CI 環境を検出
  - PTY が利用できない環境ではテストをスキップ
  - isTmuxAvailable() 関数でテスト前にチェック

- Marshall の能動的指令認識パターン
  - 起動時に envoy_to_marshall.yaml を確認
  - status: pending の指令を能動的に検出
  - inbox 待機だけでなく、指令キューを能動的にチェック

## 次のアクション

Phase 2 の実装に進む:
- git worktree 管理
- 依存関係解決（blocks / blocked_by）
- 複数 Specialist 同時実行
