# Bastion

[![Test](https://github.com/t-ishitsuka/bastion-core/actions/workflows/test.yml/badge.svg)](https://github.com/t-ishitsuka/bastion-core/actions/workflows/test.yml)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)

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
│  ・agents/dashboard.md 更新（単一書き込み者）                        │
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
# リポジトリをクローン
git clone https://github.com/t-ishitsuka/bastion-core.git
cd bastion-core

# 環境チェック
go run ./cmd/bastion doctor

# ビルド
go build -o bastion ./cmd/bastion

# パスに追加（オプション）
mv bastion /usr/local/bin/
```

## 実装状況

### Phase 0: 環境検証 [完了]

- bastion doctor コマンド実装
- プロジェクト構造確立（cmd/bastion, internal/）
- カラー出力対応（terminal パッケージ）
- テストコード完備

### Phase 1: 基盤（MVP）[完了]

- [x] 通信層実装
  - InboxManager（メッセージ読み書き）
  - CommandQueueManager（指令管理）
  - Watcher（fsnotify によるファイル監視）
  - カバレッジ: 84.4%
- [x] tmux セッション管理
  - SessionManager（セッション・ウィンドウ・ペイン管理）
  - セッション作成・停止・状態確認
  - send-keys によるコマンド送信
  - 全 10 テストパス、カバレッジ: 69.9%
- [x] 基本 CLI（start, status, stop）
  - `bastion start`: セッション起動
  - `bastion status`: 状態確認
  - `bastion stop`: セッション停止
  - 全 7 テストパス、カバレッジ: 43.0%
- [x] Envoy → Marshall → Specialist 通信
  - Orchestrator（エージェント管理）
  - 各エージェント用 CLAUDE.md（envoy, marshall, specialist）
  - 自動起動: `bastion start` で claude が各ウィンドウで起動
  - 役割分担: Envoy（対話）→ Marshall（管理）→ Specialists（実行）

**全体カバレッジ: 72.3%** （目標 50% 超え）

## 使い方

```bash
# Bastion セッションを起動（自動的にセッションにアタッチされます）
$ bastion start

# セッション内の操作:
# - Window 0 (main): メインウィンドウ
#   - 左ペイン: Envoy (ユーザー対話窓口)
#   - 右上ペイン: Watcher (inbox 監視)
#   - 右下ペイン: Marshall (タスク管理)
# - Window 1 (specialists): Specialist がグリッド配置

# ウィンドウ切り替え: Ctrl+B → 0 or 1
# セッションをデタッチ: Ctrl+B → D

# デタッチ後に再接続
$ bastion attach

# セッション状態確認
$ bastion status

# セッション停止
$ bastion stop
```

### Phase 2-4: 実装予定

詳細は [bastion-spec-v2.md](bastion-spec-v2.md) を参照してください。

## 使い方

### 環境チェック

```bash
# 必須 CLI ツールの確認
bastion doctor
```

doctor コマンドは以下をチェックします:

- claude CLI のインストール状況
- tmux のインストール状況とバージョン
- git のインストール状況とバージョン
- Go のバージョン（1.22 以上が必要）

### セッション管理（未実装）

以下のコマンドは今後実装予定です:

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
2. System: agents/queue/inbox/<target>.yaml に追記（sync.Mutex 排他）
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
agents/queue/
├── inbox/                      # メッセージボックス
│   ├── envoy.yaml
│   ├── marshall.yaml
│   └── specialist_*.yaml
├── tasks/                      # タスク定義（1タスク = 1ファイル）
│   ├── <id>.yaml              # Envoy からの指令
│   └── specialist_*.yaml      # Specialist へのタスク
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
              agents/dashboard.md 更新
                      │
                      ▼
             Envoy がユーザーに報告
```

**ポイント:**

- Envoy は即座に委譲してユーザー入力を待てる状態に戻る
- Marshall がバックグラウンドで並列処理を管理
- 完了すると知識として蓄積され、次回以降に活用

## 開発

### テスト実行

```bash
# すべてのテストを実行
go test ./...

# カバレッジ付きテスト
go test -v -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### CI/CD

GitHub Actions で自動的に以下が実行されます:

- テスト（カバレッジ閾値: 25% 以上必須、50% 以上推奨）
- golangci-lint による静的解析
- ビルド検証

ワークフロー: `.github/workflows/test.yml`

## 参考

- [multi-agent-shogun](https://github.com/yohey-w/multi-agent-shogun) - 設計の基盤となったマルチエージェントシステム

## ライセンス

MIT License
