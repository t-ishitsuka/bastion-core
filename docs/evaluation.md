# 評価・改善システム

## 概要

Bastion の自己強化システムは、評価→知識抽出→プロンプト改善の循環で継続的に改善されます。

```
タスク完了
    │
    ▼
┌────────────┐
│ 評価実行   │ ← Marshall が correctness/quality/efficiency を採点
└────────────┘
    │
    ▼
┌────────────┐
│ 知識抽出   │ ← パターン・教訓を knowledge/ に保存
└────────────┘
    │
    ▼
┌────────────┐
│ プロンプト │ ← スキル候補を収集、承認後 .claude/skills/ へ
│ 改善       │
└────────────┘
```

## 評価システム

### 評価タイミング

Marshall は各 Specialist のレポート受信後、完了時評価を実行します。

### 評価フォーマット

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

### 評価基準（Bloom's Taxonomy 応用）

| レベル | 内容       | 評価ポイント         |
| ------ | ---------- | -------------------- |
| L1-L2  | 記憶・理解 | 仕様通りか           |
| L3-L4  | 応用・分析 | 既存コードとの整合性 |
| L5-L6  | 評価・創造 | 改善提案の質         |

### 評価項目

| 項目           | スコア範囲 | 説明                           |
| -------------- | ---------- | ------------------------------ |
| `correctness`  | 1-5        | 要件充足度                     |
| `code_quality` | 1-5        | コードの品質（可読性、保守性） |
| `efficiency`   | 1-5        | 効率性（時間、トークン消費）   |

## 知識共有システム

### 知識の種類

| 種類      | 説明                   | 例                       |
| --------- | ---------------------- | ------------------------ |
| `pattern` | 繰り返し使えるパターン | JWT 実装パターン         |
| `lesson`  | 学んだ教訓             | テスト方法の統一         |
| `pitfall` | 避けるべき落とし穴     | 特定ライブラリの既知バグ |

### 知識の保存場所

```
knowledge/
├── patterns/
│   ├── go-jwt-middleware.md
│   └── react-form-validation.md
├── lessons/
│   ├── 2026-02-08-auth-implementation.md
│   └── ...
└── index.yaml  # 検索用インデックス
```

### パターンフォーマット

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

### 教訓フォーマット

```yaml
# knowledge/lessons/2026-02-08-auth-implementation.yaml
id: lesson_001
type: lesson
source_task: subtask_001
timestamp: "2026-02-08T11:30:00"

title: "ミドルウェアテストの統一"
content: |
  httptest.NewRecorder を使用すると、
  実際のHTTPサーバーを立てずにテスト可能。
  ResponseRecorder でヘッダーやボディを検証。
context: "JWT認証実装時に発見"
tags: [go, testing, middleware]
```

### Memory MCP 連携

重要な知識は Memory MCP にも保存し、セッション跨ぎで参照可能に:

```bash
# 知識の永続化
mcp_add_observations entity="jwt-middleware" observations="[\"github.com/golang-jwt/jwt/v5 を使用\"]"

# 既存知識の確認
mcp_read_graph
```

## プロンプト改善

### スキル候補の発見

Specialist は繰り返しパターンを発見した場合、レポートに `skill_candidate` を報告:

```yaml
skill_candidate:
  found: true
  name: "jwt-middleware"
  description: "Go JWT認証ミドルウェアの雛形生成"
  reason: "認証実装パターンが繰り返し発生"
```

### スキル昇格フロー

```
Specialist報告 → Marshall収集 → Envoy承認 → .claude/skills/ に登録
                     ↓
              knowledge/ に記録
```

### スキルファイル生成

```markdown
## <!-- .claude/skills/SKILL-jwt-auth.md -->

name: jwt-auth
description: Go JWT認証の実装

---

# JWT認証実装スキル

## 使用方法

`/jwt-auth <エンドポイントパス>`

## 実行内容

1. middleware/auth.go 生成
2. テストファイル生成
3. 既存ルーターへの組み込み提案
```

### コンテキスト注入

Marshall は新しいタスク割当時、関連する知識を自動注入:

```yaml
# タスク YAML への自動追加
context_injection:
  patterns:
    - "go-jwt-middleware" # 関連パターンを自動参照
  lessons:
    - "2026-02-08-auth-implementation"
```

## 改善適用ルール

| 信頼度                | 適用方法       |
| --------------------- | -------------- |
| 高（3回以上同じ提案） | 自動適用可     |
| 中（1-2回）           | 人間の承認必要 |

## 関連ファイル

```
internal/evaluation/
├── evaluator.go       # 評価ロジック
└── knowledge.go       # 知識抽出

knowledge/
├── patterns/          # 抽出されたパターン
├── lessons/           # 教訓
├── evaluations/       # 評価履歴
└── index.yaml         # 検索インデックス
```
