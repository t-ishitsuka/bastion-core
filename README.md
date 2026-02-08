# Bastion

> 「一人開発会社」を実現する Claude Code マルチエージェントオーケストレーター

## 概要

Bastion は、実証済みのマルチエージェントパターンを Go で堅牢に実装し、自己強化機能を追加したシステムです。

| 項目     | 内容                                             |
| -------- | ------------------------------------------------ |
| 言語     | Go                                               |
| 目的     | Claude Code マルチエージェントオーケストレーター |
| 設計思想 | Claude Code First、イベント駆動、自己強化        |

## 特徴

- **階層構造**: Envoy → Marshall → Specialists
- **イベント駆動**: fsnotify + inbox でポーリングゼロ
- **並列実行**: tmux + git worktree で複数 Specialist 同時作業
- **自己強化**: 評価→知識抽出→プロンプト改善の循環
- **外部注入**: Specialist の動的追加・削除
- **Claude Code 統合**: Skills / Hooks / Agents との連携

## アーキテクチャ

```
User
  │
  ▼
┌─────────────────────────────────────────────────────────────┐
│                           Envoy                              │
│  ・ユーザーとの唯一の対話窓口                                 │
│  ・what / acceptance_criteria の定義                         │
│  ・即時委譲してターン終了                                     │
└─────────────────────────────────────────────────────────────┘
                          │ inbox_write
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                         Marshall                             │
│  ・タスク分解（how の決定）                                   │
│  ・Specialist 割当・並列実行管理                              │
│  ・品質評価・知識抽出                                         │
│  ・dashboard.md 更新（単一書き込み者）                        │
└─────────────────────────────────────────────────────────────┘
                          │ inbox_write（並列）
            ┌─────────────┼─────────────┐
            ▼             ▼             ▼
     ┌──────────┐  ┌──────────┐  ┌──────────┐
     │Specialist│  │Specialist│  │Specialist│
     │    1     │  │    2     │  │    N     │
     └──────────┘  └──────────┘  └──────────┘
            │             │             │
            └─────────────┴─────────────┘
                          │ report + inbox_write
                          ▼
                    Marshall へ報告
```

## 必要条件

### 必須 CLI

| CLI      | 用途                     | インストール                               |
| -------- | ------------------------ | ------------------------------------------ |
| `claude` | Claude Code 本体         | `npm install -g @anthropic-ai/claude-code` |
| `tmux`   | セッション管理・並列実行 | `brew install tmux` / `apt install tmux`   |
| `git`    | バージョン管理、worktree | 通常プリインストール                       |

### オプション CLI

| CLI  | 用途                        | インストール      |
| ---- | --------------------------- | ----------------- |
| `gh` | GitHub CLI（Issue/PR 操作） | `brew install gh` |
| `jq` | JSON パース（デバッグ用）   | `brew install jq` |

## インストール

```bash
# ソースからビルド
git clone https://github.com/yourname/bastion.git
cd bastion
go build -o bastion ./cmd/bastion

# パスに追加
mv bastion /usr/local/bin/
```

## 使い方

```bash
# セッション起動
bastion start

# Specialist 数を指定して起動
bastion start --specialists 4

# 状態確認
bastion status

# Specialist 追加
bastion specialist add ./my-specialist.yaml

# Specialist 一覧
bastion specialist list
```

## 通信プロトコル

Mailbox System を採用（Go で実装）:

```
1. Sender: inbox.Write(target, message, type)
2. System: queue/inbox/<target>.yaml に追記（sync.Mutex 排他）
3. Watcher: fsnotify が変更検知 → tmux send-keys で nudge
4. Receiver: inbox YAML を読み込み処理
```

**特徴:**

- ゼロポーリング（fsnotify はカーネルレベル）
- Go の sync.Mutex による排他制御
- YAML でエージェント再起動を跨いで状態保持

## 自己強化システム

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

## ディレクトリ構造

```
queue/
├── envoy_to_marshall.yaml      # Envoy → Marshall 指令
├── inbox/                      # メッセージボックス
│   ├── marshall.yaml
│   └── specialist_*.yaml
├── tasks/                      # タスク詳細
│   └── specialist_*.yaml
└── reports/                    # 完了報告
    └── specialist_*_report.yaml

knowledge/
├── patterns/                   # 抽出されたパターン
├── lessons/                    # 教訓
└── index.yaml                  # 検索インデックス
```

## ドキュメント

詳細仕様は [bastion-spec-v2.md](bastion-spec-v2.md) を参照。

- [アーキテクチャ](docs/architecture.md)
- [役割定義](docs/roles.md)
- [通信フォーマット](docs/communication.md)
- [評価・改善システム](docs/evaluation.md)

## 実装したいこと

```
ユーザー入力
    │
    ▼
┌────────────────────────────────────────────────────┐
│ Envoy: 目的と完了条件を定義 → 即座に Marshall へ委譲 │
└────────────────────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────────────────────┐
│ Marshall: タスク分解 → 複数 Specialist に並列割当    │
└────────────────────────────────────────────────────┘
    │
    ├─────────────────┬─────────────────┐
    ▼                 ▼                 ▼
┌─────────┐     ┌─────────┐      ┌─────────┐
│ Spec 1  │     │ Spec 2  │      │ Spec N  │  ← 各 worktree で並列作業
└─────────┘     └─────────┘      └─────────┘
    │                 │                 │
    └─────────────────┴─────────────────┘
                      │
                      ▼
┌────────────────────────────────────────────────────┐
│ Marshall: 評価 → 知識抽出 → スキル候補収集           │
└────────────────────────────────────────────────────┘
                      │
                      ▼
              dashboard.md 更新
                      │
                      ▼
             Envoy がユーザーに報告
```

**ポイント:**

- Envoy は即座に委譲してユーザー入力を待てる状態に戻る
- Marshall がバックグラウンドで並列処理を管理
- 完了すると知識として蓄積され、次回以降に活用

## 参考

- [multi-agent-shogun](https://github.com/yohey-w/multi-agent-shogun) - 設計の基盤となったマルチエージェントシステム

## ライセンス

MIT License
