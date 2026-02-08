# 役割定義

## 概要

Bastion は 3 層の階層構造で構成されています。

```
Envoy → Marshall → Specialists
```

## Envoy

| 項目   | 内容                                             |
| ------ | ------------------------------------------------ |
| 責務   | ユーザーとの唯一の対話窓口                       |
| 入力   | ユーザー入力                                     |
| 出力   | 指令（`queue/envoy_to_marshall.yaml`）、結果報告 |
| 決定権 | **what（目的）** と **acceptance_criteria**      |

### 主な機能

- ユーザーからの要望を受け取り、目的と完了条件を定義
- `queue/envoy_to_marshall.yaml` に指令を書き込み
- Marshall への即時委譲（ターン終了）
- `dashboard.md` を読んでユーザーに報告

### 禁止事項

- 自らファイル操作を行う（F001）
- Specialist に直接指示する（F002）
- ポーリングループ（F004）

### 指令フォーマット

```yaml
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
  priority: high
  status: pending
```

## Marshall

| 項目   | 内容                                        |
| ------ | ------------------------------------------- |
| 責務   | タスク分解、Specialist への割当・管理       |
| 入力   | Envoy からの指令、Specialist からのレポート |
| 出力   | タスク定義、評価、知識抽出                  |
| 決定権 | **how（実行方法）**                         |
| 特権   | `dashboard.md` の更新（単一書き込み者）     |

### 主な機能

- 指令をタスクに分解
- 依存関係を把握しスケジューリング
- Specialist への並列割当
- 完了時の品質評価
- 知識抽出・共有
- `dashboard.md` の更新

### 禁止事項

- ユーザーに直接報告（Envoy 経由）
- ポーリングループ

### タスク分解フォーマット

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
blocks: []
blocked_by: []
```

### 評価フォーマット

```yaml
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

## Specialist

| 項目     | 内容                                                 |
| -------- | ---------------------------------------------------- |
| 責務     | 割り当てられたタスクの実行                           |
| 入力     | タスク定義（`queue/tasks/specialist_N.yaml`）        |
| 出力     | レポート（`queue/reports/specialist_N_report.yaml`） |
| 実行環境 | 各 worktree で tmux pane として並列実行              |

### 主な機能

- 専門ペルソナでの高品質な成果物作成
- 完了報告（Marshall 宛）
- スキル候補の報告
- 目的検証（`parent_cmd` の purpose と照合）

### 禁止事項

- Envoy に直接報告（F001）
- ユーザーに直接連絡（F002）
- 割り当て外の作業（F003）
- 他 Specialist のファイルを読み書き

### ペルソナ例

| カテゴリ      | ペルソナ                                                             |
| ------------- | -------------------------------------------------------------------- |
| development   | Senior Software Engineer, QA Engineer, SRE/DevOps, Database Engineer |
| documentation | Technical Writer, Senior Consultant, Presentation Designer           |
| analysis      | Data Analyst, Market Researcher, Strategy Analyst                    |

### レポートフォーマット

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
  found: true
  name: "jwt-middleware"
  description: "Go JWT認証ミドルウェアの雛形生成"
  reason: "認証実装パターンが繰り返し発生"
```

## 外部 Specialist 注入

### Specialist 定義

```yaml
# .claude/agents/specialist-security.yaml
name: security-auditor
description: "セキュリティ監査専門"
persona: "Senior Security Engineer"
capabilities:
  - "OWASP Top 10 チェック"
  - "依存関係脆弱性スキャン"
  - "認証・認可レビュー"
trigger_patterns:
  - "セキュリティ"
  - "脆弱性"
  - "監査"
```

### 注入タイミング

- **起動時**: `.claude/agents/` をスキャン
- **実行時**: `bastion specialist add <path>` コマンド
- **タスク時**: Marshall がタスク内容から trigger_patterns マッチ
