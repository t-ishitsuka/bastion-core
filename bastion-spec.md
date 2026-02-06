# Bastion 仕様書 v1.0

## 1. 概要

| 項目 | 内容 |
|------|------|
| プロジェクト名 | **Bastion** |
| 言語 | Go |
| 目的 | Claude Code のマルチエージェントオーケストレーター |
| 設計思想 | サードパーティ最小限、ローカル完結、シングルバイナリ |

---

## 2. 依存 CLI

### 必須

| CLI | 用途 | インストール |
|-----|------|-------------|
| `claude` | Claude Code 本体 | `npm install -g @anthropic-ai/claude-code` |
| `tmux` | セッション管理・並列実行 | `brew install tmux` / `apt install tmux` |
| `git` | バージョン管理、worktree | 通常プリインストール |
| `gh` | GitHub CLI（Issue/PR操作） | `brew install gh` / `apt install gh` |

### オプション

| CLI | 用途 | インストール |
|-----|------|-------------|
| `jq` | JSON パース（デバッグ用） | `brew install jq` |

---

## 3. アーキテクチャ

```
User
  │
  ▼
┌─────────────────────────────────────────────────────────────┐
│                         Bastion                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌───────────┐                                              │
│  │ Dashboard │◄────── TUI (Bubble Tea)                      │
│  └─────┬─────┘                                              │
│        │                                                    │
│  ┌─────▼─────┐                                              │
│  │   Envoy   │◄────── User Input / GitHub Issues            │
│  └─────┬─────┘        Issue 作成（ユーザー要望から）          │
│        │                                                    │
│        ▼                                                    │
│  ┌──────────┐                                               │
│  │ Marshall │──── タスク分解・サブ Issue 作成                │
│  └────┬─────┘                                               │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────────────────────────────┐                │
│  │            Specialists                  │                │
│  ├────────────┬────────────┬──────────────┤                │
│  │ Implementer│ Implementer│   Tester     │ (tmux panes)   │
│  │  (auth)    │  (api)     │              │                │
│  └────────────┴────────────┴──────────────┘                │
│       │                                                     │
│       ▼                                                     │
│  ┌──────────┐                                               │
│  │ Evaluator│──── 相互評価・改善提案                         │
│  └──────────┘                                               │
│       │                                                     │
│       ▼                                                     │
│  PR / Issue 更新 ─────────► GitHub                          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 4. 役割定義

### 4.1 Envoy（使節）

| 項目 | 内容 |
|------|------|
| 責務 | ユーザーとの対話、GitHub Issue との窓口 |
| 入力 | ユーザー入力、GitHub Issue |
| 出力 | Issue 作成、進捗報告、結果報告 |
| Issue 操作 | 読み ○ / 書き ○（ユーザー要望から作成） |

### 4.2 Marshall（元帥）

| 項目 | 内容 |
|------|------|
| 責務 | 要件分析、タスク分解、Specialist への指揮 |
| 入力 | Issue、Envoy からの指令 |
| 出力 | タスク定義、サブ Issue、スケジュール |
| Issue 操作 | 読み ○ / 書き ○（サブタスク分解） |
| モード | Claude Code Plan Mode 使用 |

### 4.3 Specialist（専門家）

| 項目 | 内容 |
|------|------|
| 責務 | タスク実行（実装、テスト、レビュー等） |
| 入力 | タスク定義 |
| 出力 | コード、テスト、レポート |
| Issue 操作 | 読み △（自分のタスクのみ）/ 書き ○（問題発見時） |
| 実行環境 | 各 worktree で tmux pane として並列実行 |

#### Specialist の種類

| 名前 | 役割 |
|------|------|
| Implementer | コード実装 |
| Tester | テスト作成 |
| Reviewer | コードレビュー |
| (カスタム) | 設定ファイルで定義可能 |

---

## 5. ディレクトリ構成

```
bastion/
├── cmd/
│   └── bastion/
│       └── main.go
├── internal/
│   ├── doctor/
│   │   └── doctor.go              # 依存チェック
│   ├── dashboard/
│   │   ├── dashboard.go           # メインモデル
│   │   ├── views.go               # 各タブのレンダリング
│   │   ├── events.go              # イベント定義
│   │   ├── styles.go              # スタイル定義
│   │   ├── commands.go            # コマンドパレット
│   │   ├── envoy_view.go          # Envoy対話画面
│   │   └── keybindings.go         # キーバインド
│   ├── envoy/
│   │   ├── envoy.go               # ユーザー対話
│   │   ├── prompt.go              # プロンプト処理
│   │   ├── reporter.go            # 進捗報告
│   │   └── issue_writer.go        # Issue 作成
│   ├── marshall/
│   │   ├── marshall.go            # タスク分解・指揮
│   │   ├── planner.go             # 計画立案
│   │   ├── scheduler.go           # スケジューリング
│   │   ├── decomposer.go          # サブ Issue 分解
│   │   └── evaluator.go           # Specialist 評価
│   ├── specialist/
│   │   ├── specialist.go          # 基底インターフェース
│   │   ├── implementer.go         # 実装特化
│   │   ├── tester.go              # テスト特化
│   │   ├── reviewer.go            # レビュー特化
│   │   ├── registry.go            # 登録・管理
│   │   ├── peer_review.go         # ピアレビュー
│   │   ├── committer.go           # コミット管理
│   │   └── warmup.go              # ウォームアップ処理
│   ├── evaluation/
│   │   ├── evaluator.go           # 評価インターフェース
│   │   ├── aggregator.go          # 提案集約
│   │   ├── improver.go            # 改善適用
│   │   └── history.go             # 履歴管理
│   ├── task/
│   │   ├── task.go                # タスク定義
│   │   ├── queue.go               # タスクキュー
│   │   └── parser.go              # YAML パース
│   ├── tmux/
│   │   ├── session.go             # セッション管理
│   │   └── pane.go                # pane 操作
│   ├── git/
│   │   ├── worktree.go            # worktree 管理
│   │   ├── operations.go          # commit, push 等
│   │   └── committer.go           # コミットポリシー実行
│   ├── github/
│   │   ├── client.go              # 共通処理
│   │   ├── issues.go              # Issue CRUD
│   │   └── pr.go                  # PR 操作
│   ├── knowledge/
│   │   ├── store.go               # 知識ストア
│   │   └── patterns.go            # パターン管理
│   ├── estimation/
│   │   ├── estimator.go           # 見積もりロジック
│   │   ├── calibrator.go          # 補正係数計算
│   │   ├── tracker.go             # 精度追跡
│   │   └── cost.go                # コスト計算
│   ├── observability/
│   │   ├── logger.go              # 構造化ログ
│   │   ├── metrics.go             # メトリクス収集
│   │   └── trace.go               # タスクトレース
│   └── comm/
│       ├── channel.go             # 通信抽象化
│       ├── inbox.go               # メッセージ受信
│       └── outbox.go              # メッセージ送信
├── configs/
│   ├── default.yaml               # デフォルト設定
│   ├── simple.yaml                # 軽量スタート用
│   ├── commit_policy.yaml         # コミットポリシー
│   ├── context_policy.yaml        # コンテキスト管理
│   ├── intervention.yaml          # 人間介入設定
│   ├── fallback.yaml              # フォールバック設定
│   ├── estimation.yaml            # 見積もり設定
│   ├── specialists/               # Specialist 定義
│   │   ├── implementer.yaml
│   │   ├── tester.yaml
│   │   └── reviewer.yaml
│   └── improvements/
│       └── history.yaml           # 改善履歴
├── comms/                         # 実行時通信ディレクトリ
│   ├── tasks/
│   ├── reports/
│   ├── evaluations/
│   └── knowledge/
├── go.mod
└── go.sum
```

---

## 6. 設定ファイル

### 6.1 デフォルト設定

```yaml
# configs/default.yaml
bastion:
  name: "bastion"
  version: "1.0.0"

specialists:
  max_count: 8
  default_model: sonnet

github:
  labels_prefix: "bastion"
  auto_create_labels: true

dashboard:
  refresh_interval: 1s
  default_tab: overview
```

### 6.2 コンテキスト管理

```yaml
# configs/context_policy.yaml
context:
  # Specialist に渡すコンテキストの上限
  max_tokens_per_task: 50000
  
  # 含めるもの
  include:
    - task_definition
    - related_files
    - project_conventions
  
  # 含めないもの
  exclude:
    - other_tasks
    - full_history
  
  # コンテキスト肥大化時の対応
  overflow_strategy: summarize  # summarize | truncate | fail
```

### 6.3 コミットポリシー

```yaml
# configs/commit_policy.yaml
policy:
  # コミット粒度
  granularity: step  # step | file | task
  
  # コミットするタイミング
  triggers:
    - type: file_complete
    - type: test_pass
    - type: milestone
    - type: before_risky_change
  
  # コミットしないタイミング
  skip:
    - syntax_error_exists
    - test_failing
  
  # コミットメッセージ形式
  format: conventional
  
  # 自動メッセージ生成
  auto_message:
    enabled: true
    use_claude: true
  
  # WIP コミット
  wip:
    enabled: true
    prefix: "wip"
    squash_on_complete: false
```

### 6.4 人間介入設定

```yaml
# configs/intervention.yaml
intervention:
  # 人間の承認が必要なケース
  require_approval:
    - pr_create
    - destructive_change
    - confidence_low
    - cost_threshold: 10000
  
  # 通知するケース
  notify:
    - task_complete
    - task_failed
    - evaluation_ready
  
  # 通知方法
  notification:
    method: terminal_bell  # terminal_bell | desktop | slack
```

### 6.5 フォールバック設定

```yaml
# configs/fallback.yaml
fallback:
  specialist_failure:
    retry_count: 2
    retry_delay: 30s
    on_max_retry:
      - action: reassign
      - action: simplify_task
      - action: escalate_to_human
  
  context_overflow:
    action: compact_and_retry
  
  rate_limit:
    action: queue_and_wait
    max_wait: 30m
  
  unknown_error:
    action: pause_and_notify
```

### 6.6 Specialist 設定

```yaml
# configs/specialists/implementer.yaml
name: implementer
description: コード実装を担当
prompt: |
  あなたは実装担当の Specialist です。
  与えられたタスク定義に従って、コードを実装してください。
  
  ルール:
  - 既存のコードスタイルに従う
  - テストは別の Specialist が担当するため、実装に集中
  - 完了したら reports/ にレポートを出力
  - ステップごとにコミットする
capabilities:
  - code_write
  - file_create
  - file_edit
model: sonnet

warmup:
  enabled: true
  steps:
    - read_claude_md
    - read_related_files
    - understand_conventions
  max_tokens: 10000
```

```yaml
# configs/specialists/tester.yaml
name: tester
description: テスト作成を担当
prompt: |
  あなたはテスト担当の Specialist です。
  実装されたコードに対してテストを作成してください。
  
  ルール:
  - ユニットテストを優先
  - エッジケースを網羅
  - 既存のテストフレームワークに従う
capabilities:
  - code_write
  - test_run
model: sonnet
```

```yaml
# configs/specialists/reviewer.yaml
name: reviewer
description: コードレビューを担当
prompt: |
  あなたはレビュー担当の Specialist です。
  実装されたコードをレビューしてください。
  
  観点:
  - バグの可能性
  - パフォーマンス問題
  - セキュリティリスク
  - より良い実装方法
capabilities:
  - code_read
  - comment_create
model: sonnet
```

### 6.7 軽量スタート用設定

```yaml
# configs/simple.yaml
mode: simple

specialists:
  count: 1
  type: implementer

evaluation:
  enabled: false

commit_policy:
  granularity: task

intervention:
  require_approval:
    - all_changes
```

---

## 7. 通信フォーマット

### 7.1 タスク定義

```yaml
# comms/tasks/task-001.yaml
id: task-001
source: marshall
target: implementer-1
type: implement
priority: high
status: pending

# 見積もり情報
estimate:
  time: 45m
  tokens: 25000
  complexity: L           # S/M/L/XL
  risk: medium            # low/medium/high
  confidence: 0.8         # Marshall の自信度
  cost:
    model: sonnet
    input_tokens: 20000
    output_tokens: 5000
    estimated_cost: 0.135  # USD

context:
  issue_number: 42
  branch: bastion/issue-42

spec:
  title: "ユーザー認証のログイン機能"
  description: |
    JWT を使用したログイン機能を実装する
  files:
    - src/auth/login.ts
    - src/auth/types.ts
  acceptance:
    - ログインAPIが動作する
    - エラーハンドリングが適切

steps:
  - id: step-1
    description: "型定義の作成"
    files: ["src/auth/types.ts"]
  - id: step-2
    description: "バリデーション作成"
    files: ["src/auth/validation.ts"]
  - id: step-3
    description: "ハンドラー実装"
    files: ["src/auth/login.ts"]

dependencies: []
created_at: 2025-02-05T10:00:00Z
```

### 7.2 完了レポート

```yaml
# comms/reports/task-001-done.yaml
id: task-001
source: implementer-1
target: marshall
type: report
status: completed

result:
  success: true
  summary: "ログイン機能を実装完了"
  
  changed_files:
    - src/auth/login.ts
    - src/auth/types.ts
    - src/auth/validation.ts
  
  commits:
    - hash: "abc123"
      message: "feat(auth): define login request/response types"
      files: ["src/auth/types.ts"]
    - hash: "def456"
      message: "feat(auth): add login input validation"
      files: ["src/auth/validation.ts"]
    - hash: "ghi789"
      message: "wip(auth): login handler in progress"
      files: ["src/auth/login.ts"]
      is_wip: true
    - hash: "jkl012"
      message: "feat(auth): implement login handler"
      files: ["src/auth/login.ts"]
  
  rollback_points:
    - hash: "def456"
      description: "validation まで戻す"
    - hash: "abc123"
      description: "types だけ残す"

# 見積もり vs 実績
metrics:
  estimate:
    time: 45m
    tokens: 25000
    cost: 0.135
  actual:
    time: 38m
    tokens: 21500
    cost: 0.116
  variance:
    time: -16%
    tokens: -14%
    cost: -14%

completed_at: 2025-02-05T10:30:00Z
```

### 7.3 評価レポート

```yaml
# comms/evaluations/eval-001.yaml
id: eval-001
evaluator: marshall
target: implementer-1
task_id: task-001
timestamp: 2025-02-05T12:00:00Z

scores:
  code_quality: 4
  spec_compliance: 5
  efficiency: 3
  documentation: 2

feedback:
  positive:
    - "エラーハンドリングが適切"
    - "命名規則に準拠"
  improvements:
    - "コメントが少ない"
    - "早期リターンを使うとより読みやすい"

suggestions:
  - type: prompt_update
    description: "ドキュメント生成を強化"
    diff: |
      + - 関数には必ずコメントを付ける
      + - 複雑なロジックには説明を追加
```

### 7.4 知識共有

```yaml
# comms/knowledge/auth-patterns.yaml
id: knowledge-001
source: implementer-1
task_id: task-001
type: pattern

content:
  title: "このプロジェクトの認証パターン"
  description: |
    - JWT は jose ライブラリを使用
    - エラーは AppError クラスでラップ
    - バリデーションは zod を使用
  files:
    - src/lib/auth.ts
    - src/lib/errors.ts

tags: [auth, jwt, validation]
```

---

## 8. コミット戦略

### 8.1 コミットタイミング

| トリガー | 説明 | 例 |
|----------|------|-----|
| `file_complete` | ファイル単位の実装完了 | types.ts 作成完了 |
| `test_pass` | テストが通った | ユニットテスト全パス |
| `milestone` | マイルストーン到達 | API エンドポイント完成 |
| `before_risky_change` | 危険な変更の前 | 大規模リファクタ前 |

### 8.2 スキップ条件

| 条件 | 説明 |
|------|------|
| `syntax_error_exists` | 構文エラーがある状態ではコミットしない |
| `test_failing` | テストが落ちている状態ではコミットしない（WIP除く） |

### 8.3 コミットフロー例

```
Task: ログイン機能の実装 (issue #42)
Branch: bastion/issue-42

Timeline:
─────────────────────────────────────────────────────────
10:00  タスク開始
       └── worktree 作成、branch checkout

10:05  types.ts 作成完了
       └── ✓ コミット "feat(auth): define login types"

10:15  validation.ts 作成完了
       └── ✓ コミット "feat(auth): add input validation"

10:25  login.ts 作成中... 複雑なので WIP
       └── ✓ コミット "wip(auth): login handler in progress"

10:35  login.ts 完成、でもテスト未実行
       └── ✗ コミットしない（テスト通ってない）

10:40  テスト実行 → 失敗
       └── ✗ コミットしない

10:45  バグ修正、テスト通過
       └── ✓ コミット "feat(auth): implement login handler"

10:50  タスク完了
       └── レポート生成、Marshall に報告
─────────────────────────────────────────────────────────
```

### 8.4 失敗時のリカバリー

```
Specialist がステップ3で失敗した場合:

コミット履歴:
abc123 feat(auth): define types        ← 残る
def456 feat(auth): add validation      ← 残る
ghi789 wip(auth): checkpoint           ← 残る

Marshall の選択肢:
1. 同じ Specialist に再試行（ghi789 から継続）
2. 別の Specialist に引き継ぐ
3. ghi789 を revert して別アプローチ
4. 人間に判断を委ねる
```

---

## 9. GitHub 連携

### 9.1 操作一覧

| 操作 | コマンド | 用途 |
|------|----------|------|
| Issue 一覧 | `gh issue list` | タスクソース取得 |
| Issue 詳細 | `gh issue view` | 要件取得 |
| Issue 作成 | `gh issue create` | Envoy/Marshall が作成 |
| Issue コメント | `gh issue comment` | 進捗報告 |
| Issue クローズ | `gh issue close` | 完了時 |
| PR 作成 | `gh pr create` | 実装完了時 |
| PR ステータス | `gh pr view` | CI 結果確認 |

### 9.2 ラベル設計

| ラベル | 用途 |
|--------|------|
| `bastion` | Bastion 管理下 |
| `bastion:parent` | 親 Issue |
| `bastion:sub-task` | サブタスク |
| `bastion:pending` | 未着手 |
| `bastion:in-progress` | 作業中 |
| `bastion:blocked` | 依存待ち |
| `bastion:done` | 完了 |

### 9.3 Issue テンプレート

```markdown
<!-- .github/ISSUE_TEMPLATE/bastion-task.md -->
---
name: Bastion Task
about: Bastion が自動実装するタスク
labels: bastion
---

## 概要
<!-- 何を実装したいか -->

## 詳細要件
<!-- 具体的な仕様 -->

## 受け入れ条件
- [ ] ...
- [ ] ...

## 参考
<!-- 関連ファイル、ドキュメント等 -->
```

---

## 10. 評価・改善システム

### 10.1 評価フロー

```
Task完了 → 評価収集 → 集約・分析 → 改善提案 → 適用（承認制 or 自動）
```

### 10.2 評価マトリクス

| 対象 | 評価者 | 評価項目 |
|------|--------|----------|
| Specialist | Marshall | コード品質、仕様準拠、効率、ドキュメント |
| Specialist | 他Specialist | ピアレビュー |
| Marshall | Envoy | タスク分解の適切さ、見積もり精度 |
| Envoy | Marshall | 要件の明確さ |

### 10.3 改善適用ルール

| 信頼度 | 適用方法 |
|--------|----------|
| 高（3回以上同じ提案） | 自動適用可 |
| 中（1-2回） | 人間の承認必要 |

---

## 11. ダッシュボード

### 11.1 全体レイアウト

```
┌─ Bastion ─────────────────────────────────────────────────────────────────┐
│ [F1]Overview [F2]Specialists [F3]Tasks [F4]Logs [F5]Envoy [F6]Settings    │
├───────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─ Status ──────────────────────┐ ┌─ Metrics ─────────────────────────┐ │
│  │ ● Running     Issue: #42      │ │ Tokens: ████████░░ 65,231/88,000  │ │
│  │ Specialists: 3/4 active       │ │ Tasks:  ██████░░░░ 6/10 complete  │ │
│  │ Uptime: 1h 23m                │ │ Commits: 12     PRs: 0           │ │
│  └───────────────────────────────┘ └────────────────────────────────────┘ │
│                                                                           │
│  ┌─ Specialists ─────────────────────────────────────────────────────────┐│
│  │ NAME           STATUS      TASK              PROGRESS    TOKENS       ││
│  │ implementer-1  ● working   task-001 login    ████████░░  12,450       ││
│  │ implementer-2  ● working   task-002 api      ███░░░░░░░   4,200       ││
│  │ tester-1       ○ waiting   (waiting task-001)             0          ││
│  │ reviewer-1     ○ idle      -                              0          ││
│  └───────────────────────────────────────────────────────────────────────┘│
│                                                                           │
│  ┌─ Live Output (implementer-1) ─────────────────────────────────────────┐│
│  │ > Reading src/auth/types.ts...                                        ││
│  │ > Creating login handler with JWT validation                          ││
│  │ > Writing src/auth/login.ts                                           ││
│  │ > ✓ Committed: "feat(auth): implement login handler"                  ││
│  │ █                                                                     ││
│  └───────────────────────────────────────────────────────────────────────┘│
│                                                                           │
│  ┌─ Command ─────────────────────────────────────────────────────────────┐│
│  │ > _                                                                   ││
│  └───────────────────────────────────────────────────────────────────────┘│
│                                                                           │
│ [p]Pause [r]Resume [c]Cancel [m]Message [?]Help              ESC:Menu    │
└───────────────────────────────────────────────────────────────────────────┘
```

### 11.2 タブ一覧

| タブ | キー | 内容 |
|------|------|------|
| Overview | F1 | 全体状況、Specialist一覧、ライブ出力 |
| Specialists | F2 | Specialist詳細、個別操作 |
| Tasks | F3 | タスクパイプライン、依存関係可視化 |
| Logs | F4 | 構造化ログ、フィルタ、検索 |
| Envoy | F5 | 対話インターフェース |
| Settings | F6 | 設定変更 |

### 11.3 キーバインド

| キー | 動作 |
|------|------|
| `F1-F6` | タブ切り替え |
| `Ctrl+P` | コマンドパレット |
| `Ctrl+C` | 選択中の Specialist をキャンセル |
| `p` | 一時停止（全体 or 選択中） |
| `r` | 再開 |
| `m` | メッセージ送信 |
| `?` | ヘルプ |
| `q` / `Esc` | 終了 / 戻る |

### 11.4 Envoy 対話画面

```
┌─ Envoy ───────────────────────────────────────────────────────────────────┐
│                                                                           │
│  ┌─ Conversation ────────────────────────────────────────────────────────┐│
│  │                                                                       ││
│  │  You: ログイン機能を実装して                                           ││
│  │                                                                       ││
│  │  Envoy: Issue #42 を作成しました。                                     ││
│  │         以下のタスクに分解しました：                                    ││
│  │         1. 型定義の作成                                                ││
│  │         2. バリデーションの実装                                        ││
│  │         3. ログインハンドラーの実装                                     ││
│  │                                                                       ││
│  │         実行しますか？ [Y/n]                                           ││
│  │                                                                       ││
│  │  You: Y                                                               ││
│  │                                                                       ││
│  │  Envoy: 実行を開始しました。                                           ││
│  │                                                                       ││
│  └───────────────────────────────────────────────────────────────────────┘│
│                                                                           │
│  ┌─ Input ───────────────────────────────────────────────────────────────┐│
│  │ > _                                                                   ││
│  └───────────────────────────────────────────────────────────────────────┘│
│                                                                           │
│ Quick: [1]Status [2]Pause all [3]Resume [4]New issue [5]Show tasks       │
└───────────────────────────────────────────────────────────────────────────┘
```

### 11.5 Specialist へのメッセージ送信

```
┌─ Send Message to implementer-1 ───────────────────────────────────────────┐
│                                                                           │
│  Current task: task-001 (ログインハンドラーの実装)                          │
│  Status: Step 3/4 - ハンドラー実装中                                        │
│                                                                           │
│  ┌─ Message ─────────────────────────────────────────────────────────────┐│
│  │ _                                                                     ││
│  └───────────────────────────────────────────────────────────────────────┘│
│                                                                           │
│  Templates:                                                               │
│  [1] もっと詳しくコメントを追加して                                         │
│  [2] テストを先に書いて                                                    │
│  [3] 既存の実装を参考にして                                                │
│  [4] 一旦止めて、ここまでをコミットして                                     │
│                                                                           │
│                                          [Enter]Send  [Esc]Cancel        │
└───────────────────────────────────────────────────────────────────────────┘
```

---

## 12. CLI コマンド

```bash
# ダッシュボード起動（デフォルト）
bastion
bastion dashboard

# ヘッドレスモード
bastion start --headless

# 依存チェック
bastion doctor

# Envoy と対話（CLI）
bastion envoy

# Issue から直接タスク化
bastion dispatch --issue 42

# Issue 分解
bastion decompose --issue 42 --dry-run
bastion decompose --issue 42

# Specialist 状態確認
bastion status

# Specialist 起動
bastion spawn --type implementer --count 3

# 計画のみ作成
bastion plan --issue 42

# 評価確認
bastion eval show --target implementer-1

# 改善提案確認・適用
bastion eval suggest
bastion eval auto-improve --confidence 0.7

# 見積もり関連
bastion estimate --issue 42              # Issue の見積もり確認
bastion estimate accuracy                # 見積もり精度レポート
bastion estimate history --task task-001 # 見積もり履歴
bastion estimate calibrate               # 補正係数の確認・更新
bastion estimate cost --issue 42 --model opus  # モデル別コスト比較

# 軽量モードで起動
bastion --config simple

# 介入
bastion intervene
```

---

## 13. Go 依存ライブラリ

```go
// go.mod
module github.com/yourname/bastion

go 1.22

require (
    github.com/fsnotify/fsnotify v1.7.0     // ファイル監視
    github.com/spf13/cobra v1.8.0           // CLI
    gopkg.in/yaml.v3 v3.0.1                 // YAML
    github.com/charmbracelet/bubbletea v0.25.0  // TUI
    github.com/charmbracelet/lipgloss v0.9.0    // TUI スタイル
    github.com/charmbracelet/bubbles v0.18.0    // TUI コンポーネント
)
```

---

## 14. トークン・料金情報

### 14.1 プラン別比較

| プラン | 月額 | 5時間あたりトークン | Opus |
|--------|------|---------------------|------|
| Free | $0 | 少量 | × |
| Pro | $20 | 約 44,000 | × |
| Max 5x | $100 | 約 88,000 | ○ |
| Max 20x | $200 | 約 220,000 | ○ |

### 14.2 API 料金（従量課金時）

| モデル | Input | Output |
|--------|-------|--------|
| Sonnet 4.5 | $3 / 1M | $15 / 1M |
| Opus 4.5 | $15 / 1M | $75 / 1M |
| Haiku 4.5 | $0.80 / 1M | $4 / 1M |

### 14.3 Bastion でのトークン消費見積もり

| シナリオ | エージェント数 | 1日の見積もり |
|----------|---------------|--------------|
| 小規模 | 2-3 | 100,000-200,000 |
| 中規模 | 4-6 | 300,000-500,000 |
| 大規模 | 8+ | 800,000-2,000,000 |

### 14.4 推奨プラン

| 用途 | 推奨 |
|------|------|
| 開発・テスト | Max 5x ($100) |
| 本格運用 | Max 20x ($200) or API 従量課金 |

---

## 15. 実装フェーズ

### 15.1 段階的ロールアウト

```
Phase 0: 手動オーケストレーション（1週間）
├── tmux + worktree のスクリプト化だけ
├── タスク定義は手動で YAML 書く
├── Claude Code は直接操作
└── 目的: 基本フローの検証

Phase 1: シングルエージェント自動化（2週間）
├── doctor 実装
├── tmux/git パッケージ
├── 1つの Specialist だけ自動化
├── Marshall は人間が代行
└── 目的: Specialist の動作検証

Phase 2: Marshall 追加（2週間）
├── Marshall 実装
├── タスク分解の自動化
├── まだ並列化しない
└── 目的: 計画精度の検証

Phase 3: 並列化 + ダッシュボード（2週間）
├── 複数 Specialist
├── 競合防止
├── ダッシュボード基本実装
└── 目的: オーケストレーション検証

Phase 4: Envoy + GitHub 連携（2週間）
├── Envoy 実装
├── Issue/PR 連携
├── ダッシュボード内対話
└── 目的: E2E ワークフロー検証

Phase 5: 評価・改善システム（2週間）
├── 相互評価
├── プロンプト改善
├── 知識共有
└── 目的: 自己改善の検証

Phase 6: 安定化・UX 改善（1-2週間）
├── エラーハンドリング
├── ログ・メトリクス
├── ドキュメント
└── 目的: 本番運用準備
```

### 15.2 期間目安

| フェーズ | 期間 | 累計 |
|----------|------|------|
| Phase 0 | 1週間 | 1週間 |
| Phase 1 | 2週間 | 3週間 |
| Phase 2 | 2週間 | 5週間 |
| Phase 3 | 2週間 | 7週間 |
| Phase 4 | 2週間 | 9週間 |
| Phase 5 | 2週間 | 11週間 |
| Phase 6 | 1-2週間 | 12-13週間 |

**MVP（Phase 0-3）: 約7週間**
**実用レベル（Phase 0-6）: 約3ヶ月**

---

## 16. GitHub Actions（補助・オプション）

テストカバレッジ漏れ検出時のみ使用：

```yaml
# .github/workflows/coverage-check.yml
name: Coverage Check

on:
  pull_request:
    branches: [main]

jobs:
  coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run tests with coverage
        run: go test -coverprofile=coverage.out ./...
      
      - name: Check coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            gh issue create \
              --title "⚠️ Test coverage dropped below 80%" \
              --body "Current coverage: $COVERAGE%" \
              --label "bastion,test"
          fi
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 17. 見積もり機能

### 17.1 見積もり項目

| 項目 | 説明 | 用途 |
|------|------|------|
| **時間** | 推定所要時間 | スケジュール、進捗表示、ETA |
| **トークン** | 推定消費トークン | コスト管理、リミット対策 |
| **コスト** | 推定費用（$） | 予算管理 |
| **複雑度** | S/M/L/XL | リソース配分 |
| **リスク** | Low/Medium/High | 人間介入の判断 |
| **信頼度** | 0.0-1.0 | 見積もりの確からしさ |

### 17.2 複雑度の基準

| 複雑度 | 時間目安 | トークン目安 | 説明 |
|--------|----------|--------------|------|
| **S** | ~15分 | ~5,000 | 単純な型定義、設定変更 |
| **M** | 15-45分 | ~15,000 | 中程度の実装、テスト追加 |
| **L** | 45-90分 | ~30,000 | 複雑な実装、複数ファイル |
| **XL** | 90分以上 | ~50,000+ | 大規模リファクタ、新機能 |

### 17.3 リスク判定基準

| リスク | 基準 |
|--------|------|
| **Low** | 既存パターンの踏襲、影響範囲が限定的 |
| **Medium** | 新しいパターン、中程度の影響範囲 |
| **High** | 未知の技術、広範囲への影響、外部依存 |

### 17.4 タスク定義フォーマット（見積もり付き）

```yaml
# comms/tasks/task-001.yaml
id: task-001
source: marshall
target: implementer-1
type: implement
priority: high
status: pending

# 見積もり情報
estimate:
  time: 45m
  tokens: 25000
  complexity: L           # S/M/L/XL
  risk: medium            # low/medium/high
  confidence: 0.8         # Marshall の自信度
  cost:
    model: sonnet
    input_tokens: 20000
    output_tokens: 5000
    estimated_cost: 0.135  # USD

context:
  issue_number: 42
  branch: bastion/issue-42

spec:
  title: "ログインハンドラーの実装"
  # ...
```

### 17.5 完了レポートフォーマット（実績付き）

```yaml
# comms/reports/task-001-done.yaml
id: task-001
source: implementer-1
target: marshall
type: report
status: completed

result:
  success: true
  summary: "ログイン機能を実装完了"

# 見積もり vs 実績
metrics:
  estimate:
    time: 45m
    tokens: 25000
    cost: 0.135
  actual:
    time: 38m
    tokens: 21500
    cost: 0.116
  variance:
    time: -16%        # 予定より早い
    tokens: -14%      # 予定より少ない
    cost: -14%

# ...
```

### 17.6 見積もり精度追跡

```yaml
# comms/evaluations/estimate-accuracy.yaml
period: 2025-02-01 ~ 2025-02-05
total_tasks: 23

accuracy:
  time:
    avg_variance: +12%      # 平均で12%超過
    within_20%: 78%         # 20%以内に収まった割合
    trend: improving        # improving/stable/declining
  
  tokens:
    avg_variance: -5%
    within_20%: 91%
    trend: stable
  
  cost:
    avg_variance: -3%
    within_20%: 89%
    trend: stable

by_complexity:
  S:
    count: 8
    avg_time_variance: +5%
    avg_token_variance: -2%
  M:
    count: 10
    avg_time_variance: +10%
    avg_token_variance: -8%
  L:
    count: 4
    avg_time_variance: +25%   # Lタスクの見積もりが甘い
    avg_token_variance: +15%
  XL:
    count: 1
    avg_time_variance: +40%

suggestions:
  - "Lタスクの時間見積もりを1.3倍に補正することを推奨"
  - "トークン見積もりは概ね正確"
```

### 17.7 自動補正設定

```yaml
# configs/estimation.yaml
estimation:
  # 複雑度別の基準値
  baseline:
    S:
      time: 15m
      tokens: 5000
    M:
      time: 30m
      tokens: 15000
    L:
      time: 60m
      tokens: 30000
    XL:
      time: 120m
      tokens: 50000
  
  # 自動補正
  auto_correction:
    enabled: true
    min_samples: 10           # 10タスク以上で補正開始
    max_correction: 0.5       # 最大50%の補正
    
  # 手動補正係数（過去の精度から調整）
  manual_correction:
    L:
      time: 1.3               # Lタスクは1.3倍
    XL:
      time: 1.4
      tokens: 1.2
  
  # コスト計算
  cost:
    default_model: sonnet
    rates:
      sonnet:
        input: 3.0            # $ per 1M tokens
        output: 15.0
      opus:
        input: 15.0
        output: 75.0
      haiku:
        input: 0.8
        output: 4.0
```

### 17.8 Marshall の見積もりプロンプト

```yaml
# configs/specialists/marshall.yaml に追加
estimation_prompt: |
  タスクの見積もりを行ってください。
  
  ## 見積もり基準
  
  ### 複雑度と時間
  - S: 15分以内（単純な型定義、設定変更）
  - M: 15-45分（中程度の実装、テスト追加）
  - L: 45-90分（複雑な実装、複数ファイル）
  - XL: 90分以上（大規模リファクタ、新機能）
  
  ### トークン目安
  - S: ~5,000
  - M: ~15,000
  - L: ~30,000
  - XL: ~50,000+
  
  ### リスク判定
  - Low: 既存パターンの踏襲、影響範囲が限定的
  - Medium: 新しいパターン、中程度の影響範囲
  - High: 未知の技術、広範囲への影響、外部依存
  
  ## 過去の精度データ
  $ESTIMATION_HISTORY
  
  ## 適用する補正係数
  $CORRECTION_FACTORS
  
  ## 出力形式
  各タスクについて以下を出力:
  - time: 推定時間
  - tokens: 推定トークン数
  - complexity: S/M/L/XL
  - risk: low/medium/high
  - confidence: 0.0-1.0
  - reasoning: 見積もりの根拠
```

### 17.9 ダッシュボード表示

```
┌─ Bastion ─────────────────────────────────────────────────────────────────┐
│ [F1]Overview [F2]Specialists [F3]Tasks [F4]Logs [F5]Envoy [F6]Settings    │
├───────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─ Status ──────────────────────┐ ┌─ Estimates ────────────────────────┐│
│  │ ● Running     Issue: #42      │ │ Time:   ██████░░░░ 1h02m / 2h30m   ││
│  │ Specialists: 3/4 active       │ │ Tokens: ████░░░░░░ 35K / 85K       ││
│  │ Uptime: 1h 02m                │ │ Cost:   ████░░░░░░ $0.52 / $1.50   ││
│  └───────────────────────────────┘ │ ETA: ~1h 28m                       ││
│                                    └────────────────────────────────────┘│
│                                                                           │
│  ┌─ Tasks ───────────────────────────────────────────────────────────────┐│
│  │ TASK        COMP  STATUS     EST      ACTUAL   VARIANCE   RISK       ││
│  │ task-001    S     ✓ done     15m      12m      -20% ✓     Low        ││
│  │ task-002    M     ✓ done     30m      35m      +17%       Low        ││
│  │ task-003    L     ● working  45m      32m...   (71%)      Medium ⚠️   ││
│  │ task-004    L     ○ pending  40m      -        -          High ⚠️     ││
│  │ task-005    M     ○ blocked  20m      -        -          Low        ││
│  └───────────────────────────────────────────────────────────────────────┘│
│                                                                           │
│  ┌─ Live Output (implementer-1) ─────────────────────────────────────────┐│
│  │ > Reading src/auth/types.ts...                                        ││
│  │ > Creating login handler with JWT validation                          ││
│  │ █                                                                     ││
│  └───────────────────────────────────────────────────────────────────────┘│
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘
```

### 17.10 Envoy での見積もり表示

```
┌─ Envoy ───────────────────────────────────────────────────────────────────┐
│                                                                           │
│  You: ログイン機能を実装して                                                │
│                                                                           │
│  Envoy: Issue #42 を作成し、分析しました。                                  │
│                                                                           │
│         ┌─ 見積もり ─────────────────────────────────────┐                │
│         │ 時間:     約 2時間30分                          │                │
│         │ トークン: 約 85,000                             │                │
│         │ コスト:   約 $1.50 (Sonnet)                     │                │
│         │ タスク数: 5                                     │                │
│         │ リスク:   Medium（OAuth部分に注意）             │                │
│         └────────────────────────────────────────────────┘                │
│                                                                           │
│         タスク一覧:                                                        │
│           #  複雑度  タスク               時間   リスク                    │
│           1. [S]     型定義               15m    Low                      │
│           2. [M]     バリデーション        30m    Low                      │
│           3. [L]     ログインハンドラー    45m    Medium ⚠️                 │
│           4. [L]     OAuth               40m    High ⚠️                   │
│           5. [M]     テスト              20m    Low                       │
│                                                                           │
│         実行しますか？                                                     │
│         [Y] 実行  [D] 詳細  [E] 見積もり調整  [N] キャンセル                │
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘
```

### 17.11 CLI コマンド

```bash
# Issue の見積もり確認
bastion estimate --issue 42
# Issue #42 の見積もり:
#   全体: 2h30m / ~85,000 tokens / ~$1.50
#   タスク数: 5
#   リスク: Medium
#   
#   詳細を見る？ [y/N]

# 見積もり精度レポート
bastion estimate accuracy
# 過去30日の見積もり精度:
#   時間: 平均 +12% 超過 (78% が20%以内)
#   トークン: 平均 -5% (91% が20%以内)
#   コスト: 平均 -3%
#
#   複雑度別:
#     S: 時間 +5%, トークン -2%
#     M: 時間 +10%, トークン -8%
#     L: 時間 +25% ⚠️, トークン +15%
#     XL: 時間 +40% ⚠️
#
#   改善提案:
#   - Lタスクの時間見積もりを1.3倍に補正

# 見積もり履歴
bastion estimate history --task task-001
# task-001:
#   見積もり: 15m / 5,000 tokens / $0.03
#   実績: 12m / 4,200 tokens / $0.025
#   差異: -20% / -16% / -17%

# 補正係数の確認・更新
bastion estimate calibrate
# 現在の補正係数:
#   L: time x1.3
#   XL: time x1.4, tokens x1.2
#
# 過去の精度から再計算しますか？ [y/N]

# コスト見積もり（モデル別比較）
bastion estimate cost --issue 42 --model opus
# Issue #42 を Opus で実行した場合:
#   推定コスト: $6.25 (Sonnet比 4.2x)
```

### 17.12 ディレクトリ構成（追加）

```
bastion/
├── internal/
│   ├── estimation/              # ← 追加
│   │   ├── estimator.go         # 見積もりロジック
│   │   ├── calibrator.go        # 補正係数計算
│   │   ├── tracker.go           # 精度追跡
│   │   └── cost.go              # コスト計算
│   └── ...
├── configs/
│   ├── estimation.yaml          # ← 追加
│   └── ...
```

---

## 18. 将来の拡張案

- [ ] Slack 連携（通知、コマンド実行）
- [ ] Web ダッシュボード
- [ ] 複数プロジェクト同時管理
- [ ] カスタム Specialist のマーケットプレイス
- [ ] AI モデルの切り替え（Gemini, GPT 等）
- [ ] チーム機能（複数人での利用）
- [ ] 見積もり機械学習モデル（精度向上）

---

## 19. 参考リンク

- [Claude Code 公式ドキュメント](https://docs.anthropic.com/claude-code)
- [Claude Code GitHub Actions](https://github.com/anthropics/claude-code-action)
- [Bubble Tea (TUI フレームワーク)](https://github.com/charmbracelet/bubbletea)
- [Cobra (CLI フレームワーク)](https://github.com/spf13/cobra)

---

*Last Updated: 2025-02-05 v1.1 - 見積もり機能追加*
