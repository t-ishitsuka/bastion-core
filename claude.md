# Bastion プロジェクト概要

Bastion は「一人開発会社」を実現する Claude Code マルチエージェントオーケストレーターです。
実証済みのマルチエージェントパターンを Go で堅牢に実装しています。

## 設計思想

1. **Claude Code First** - 独自実装より公式機能（Skills, Agents, Hooks）を優先
2. **イベント駆動** - fsnotify + inbox でポーリングゼロ
3. **自己強化** - 評価→知識共有→プロンプト改善の循環
4. **外部注入** - Specialist の動的追加・削除

## プロジェクト構成

すべて Go で実装。shell script は使用しない。

```
bastion/
├── cmd/bastion/main.go          # CLI エントリーポイント
├── internal/
│   ├── orchestrator/            # メイン制御
│   │   ├── orchestrator.go
│   │   ├── envoy.go             # Envoy ロジック
│   │   ├── marshall.go          # Marshall ロジック
│   │   └── specialist.go        # Specialist 管理
│   ├── communication/           # 通信
│   │   ├── inbox.go             # inbox 読み書き
│   │   ├── watcher.go           # fsnotify でファイル監視
│   │   └── yaml.go              # YAML 操作
│   ├── evaluation/              # 評価・知識抽出
│   │   ├── evaluator.go
│   │   └── knowledge.go
│   ├── parallel/                # 並列実行
│   │   ├── tmux.go
│   │   └── worktree.go
│   └── config/
│       └── loader.go
├── queue/                       # 通信ディレクトリ（実行時生成）
│   ├── envoy_to_marshall.yaml
│   ├── inbox/
│   ├── tasks/
│   └── reports/
└── knowledge/                   # 抽出された知識
    ├── patterns/
    └── lessons/
```

## 詳細ドキュメント

**仕様書**: `bastion-spec-v2.md` を読み込む

### 補足ドキュメント

- 役割定義（Envoy, Marshall, Specialist）: `docs/roles.md`
- 通信フォーマット: `docs/communication.md`
- 評価・改善システム: `docs/evaluation.md`

## 必須 CLI ツール

| CLI      | 用途                     |
| -------- | ------------------------ |
| `claude` | Claude Code 本体         |
| `tmux`   | セッション管理・並列実行 |
| `git`    | バージョン管理、worktree |

## 階層構造

```
User
  │
  ▼
Envoy
  ・ユーザーとの唯一の対話窓口
  ・what / acceptance_criteria を定義
  ・即時委譲してターン終了
  │
  ▼
Marshall
  ・タスク分解（how の決定）
  ・Specialist 割当・並列実行管理
  ・品質評価・知識抽出
  ・dashboard.md 更新（単一書き込み者）
  │
  ▼
Specialists x N
  ・割り当てられたタスクの実行
  ・専門ペルソナでの高品質な成果物
  ・完了報告（Marshall 宛）
```

## 通信プロトコル（Mailbox System）

```go
// メッセージ送信
inbox.Write("marshall", "新規タスク割当", TypeTaskAssigned, "envoy")
```

**特徴:**

- ゼロポーリング（fsnotify でカーネルレベル監視）
- Go の sync.Mutex による排他制御
- YAML でエージェント再起動を跨いで状態保持

## 自己強化システム

```
タスク完了 → 評価実行 → 知識抽出 → プロンプト改善
              │           │           │
              ▼           ▼           ▼
          スコア付け    patterns/   .claude/skills/
                        lessons/
```

## コーディング規約

### Go コード

- Go 1.22 以上を使用
- 標準的な Go のコーディングスタイルに従う
- エラーは適切にハンドリングし、ラップして返す
- internal パッケージ内で機能を分離

### YAML ファイル

- 通信ファイルは `queue/` ディレクトリに配置
- 知識ファイルは `knowledge/` ディレクトリに配置
- タイムスタンプは ISO 8601 形式（`time.Now().Format(time.RFC3339)`）

## 並列実行

```
Session: bastion
├── Window 0: envoy (main branch)
├── Window 1: marshall (main branch)
└── Window 2: specialists
    ├── Pane 0.1: specialist_1 (worktree: .worktrees/sp1)
    ├── Pane 0.2: specialist_2 (worktree: .worktrees/sp2)
    └── ...
```

各 Specialist は独立した git worktree で作業し、ファイル競合を物理的に回避。

## 参考

- [multi-agent-shogun](https://github.com/yohey-w/multi-agent-shogun) - 設計の基盤
