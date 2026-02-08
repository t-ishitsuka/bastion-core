# 通信フォーマット

## 概要

Bastion は Mailbox System を採用しています。

```
queue/
├── envoy_to_marshall.yaml   # Envoy → Marshall 指令
├── inbox/                   # メッセージボックス
│   ├── marshall.yaml
│   └── specialist_*.yaml
├── tasks/                   # タスク詳細
│   └── specialist_*.yaml
└── reports/                 # 完了報告
    └── specialist_*_report.yaml
```

## Mailbox System

すべて Go で実装。shell script は使用しない。

### 通信フロー

```
1. Sender: inbox.Write(target, message, msgType)
2. System: queue/inbox/<target>.yaml に追記（sync.Mutex 排他）
3. Watcher: fsnotify が変更検知 → tmux send-keys で nudge
4. Receiver: inbox YAML を読み込み処理
```

### Go API

```go
// メッセージ送信
err := inbox.Write("marshall", "新規タスク割当", TypeTaskAssigned, "envoy")

// 例: Envoy → Marshall
inbox.Write("marshall", "新規タスク割当", TypeTaskAssigned, "envoy")

// 例: Marshall → Specialist
inbox.Write("specialist_1", "タスクを確認", TypeTaskAssigned, "marshall")

// 例: Specialist → Marshall（完了報告）
inbox.Write("marshall", "タスク完了", TypeReportReceived, "specialist_1")
```

### inbox メッセージフォーマット

```yaml
# queue/inbox/marshall.yaml
- id: msg_001
  timestamp: "2026-02-08T10:00:00"
  from: envoy
  type: task_assigned # task_assigned | report_received | wake_up
  message: "新規タスクを割り当てた"
  status: pending # pending | processed
```

### 特徴

- **ゼロポーリング**: fsnotify（カーネルレベル）で API 消費ゼロ
- **排他制御**: Go の sync.Mutex で同時書き込み防止
- **永続化**: YAML でエージェント再起動を跨いで状態保持
- **nudge 方式**: send-keys は短い wakeup のみ、本文は YAML から読み取り
- **保証配信**: ファイル書き込み成功 = メッセージ配信保証

## 指令フォーマット（Envoy → Marshall）

```yaml
# queue/envoy_to_marshall.yaml
- id: cmd_001
  timestamp: "2026-02-08T10:00:00"
  purpose: "認証機能が JWT ベースで動作する"
  acceptance_criteria:
    - "POST /auth/login が JWT を返す"
    - "protected endpoint が JWT 検証する"
    - "テストがパスする"
  command: |
    JWT認証を実装
    - ログインエンドポイント作成
    - ミドルウェア実装
    - テスト作成
  project: api-server
  priority: high # high | medium | low
  status: pending # pending | in_progress | done
```

### 指令フィールド

| フィールド            | 説明                           |
| --------------------- | ------------------------------ |
| `id`                  | 指令の一意識別子（cmd_NNN）    |
| `timestamp`           | ISO 8601 形式                  |
| `purpose`             | 達成すべき状態（検証可能な文） |
| `acceptance_criteria` | 完了条件のリスト（テスト可能） |
| `command`             | Marshall への詳細指示          |
| `project`             | プロジェクト ID                |
| `priority`            | 優先度                         |
| `status`              | 状態                           |

## タスクフォーマット（Marshall → Specialist）

```yaml
# queue/tasks/specialist_1.yaml
task_id: subtask_001
parent_cmd: cmd_001
assigned_to: specialist_1
persona: "Senior Backend Engineer"
timestamp: "2026-02-08T10:05:00"
status: assigned # assigned | in_progress | done | blocked

objective: "JWT認証ミドルウェアの実装"
deliverables:
  - "middleware/auth.go"
  - "middleware/auth_test.go"
context:
  - "既存の middleware/logger.go を参考"
blocks: [] # このタスク完了まで開始不可のタスク
blocked_by: [] # 完了を待つタスク
```

### タスクフィールド

| フィールド     | 説明                                 |
| -------------- | ------------------------------------ |
| `task_id`      | タスクの一意識別子                   |
| `parent_cmd`   | 親指令の ID                          |
| `assigned_to`  | 担当 Specialist                      |
| `persona`      | 専門ペルソナ                         |
| `objective`    | タスクの目的                         |
| `deliverables` | 成果物リスト                         |
| `context`      | 参考情報                             |
| `blocks`       | 依存関係（このタスクがブロックする） |
| `blocked_by`   | 依存関係（このタスクをブロックする） |

## レポートフォーマット（Specialist → Marshall）

```yaml
# queue/reports/specialist_1_report.yaml
worker_id: specialist_1
task_id: subtask_001
parent_cmd: cmd_001
timestamp: "2026-02-08T11:00:00"
status: done # done | failed | blocked

result:
  summary: "JWT認証ミドルウェア実装完了"
  files_modified:
    - "middleware/auth.go"
    - "middleware/auth_test.go"
  test_results: "4/4 passed"
  notes: "リフレッシュトークンは次フェーズで実装推奨"

purpose_validation:
  meets_criteria: true
  gaps: []

skill_candidate:
  found: true # 必須フィールド
  name: "jwt-middleware"
  description: "Go JWT認証ミドルウェアの雛形生成"
  reason: "認証実装パターンが繰り返し発生"
```

### レポートフィールド

| フィールド           | 説明                                     |
| -------------------- | ---------------------------------------- |
| `worker_id`          | 報告者の Specialist ID                   |
| `task_id`            | タスク ID                                |
| `parent_cmd`         | 親指令の ID                              |
| `status`             | 完了状態                                 |
| `result`             | 結果の詳細                               |
| `purpose_validation` | 目的検証（parent_cmd の purpose と照合） |
| `skill_candidate`    | スキル候補（必須、found: true/false）    |

## 評価フォーマット（Marshall 内部）

```yaml
# knowledge/evaluations/eval_subtask_001.yaml
evaluation:
  task_id: subtask_001
  evaluator: marshall
  timestamp: "2026-02-08T11:30:00"

  scores:
    correctness: 5 # 1-5: 要件充足度
    code_quality: 4 # 1-5: コード品質
    efficiency: 4 # 1-5: 実行効率

  issues_found: []

  knowledge_extracted:
    - type: pattern
      content: "Go JWT実装では github.com/golang-jwt/jwt/v5 を使用"
    - type: lesson
      content: "ミドルウェアテストは httptest.NewRecorder で統一"
```

## 知識フォーマット

```yaml
# knowledge/patterns/go-jwt-middleware.yaml
id: pattern_001
type: pattern
source_task: subtask_001
timestamp: "2026-02-08T11:30:00"

title: "Go JWT認証ミドルウェア"
content: |
  - github.com/golang-jwt/jwt/v5 を使用
  - Claims は独自構造体で定義
  - ミドルウェアは net/http.Handler をラップ
files:
  - middleware/auth.go
tags: [go, jwt, auth, middleware]
```

## タイムスタンプルール

常に `date` コマンドを使用。推測禁止。

```bash
date "+%Y-%m-%dT%H:%M:%S"
```
