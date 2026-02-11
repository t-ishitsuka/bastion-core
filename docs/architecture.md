# アーキテクチャ

## システム全体像

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

## ディレクトリ構成

```
bastion/
├── cmd/
│   └── bastion/
│       └── main.go              # CLI エントリーポイント
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
│   │   ├── tmux.go              # tmux セッション管理
│   │   └── worktree.go          # git worktree 管理
│   └── config/
│       └── loader.go            # 設定読み込み
├── agents/queue/                       # 通信ディレクトリ（実行時生成）
│   ├── inbox/                   # メッセージボックス
│   │   ├── envoy.yaml
│   │   ├── marshall.yaml
│   │   └── specialist_*.yaml
│   ├── tasks/                   # タスク定義（1タスク = 1ファイル）
│   │   ├── <id>.yaml            # Envoy からの指令
│   │   └── specialist_*.yaml    # Specialist へのタスク
│   └── reports/                 # 完了報告
│       └── specialist_*_report.yaml
├── knowledge/                   # 抽出された知識
│   ├── patterns/
│   ├── lessons/
│   └── index.yaml
└── agents/dashboard.md                 # Marshall が更新
```

## コンポーネント間の関係

### データフロー

1. **ユーザー入力** → Envoy が受け取り
2. **指令記録** → Envoy が `agents/queue/tasks/<id>.yaml` に書き込み（個別ファイル）
3. **通知送信** → Envoy が `agents/queue/inbox/marshall.yaml` に通知メッセージを送信
4. **タスク分解** → Marshall が `agents/queue/tasks/specialist_*.yaml` に書き込み
5. **inbox_write（並列）** → 各 Specialist に通知
6. **タスク実行** → Specialist が作業
7. **レポート** → `agents/queue/reports/` に書き込み + inbox_write
8. **評価・知識抽出** → Marshall が実施
9. **ダッシュボード更新** → Marshall が `agents/dashboard.md` 更新
10. **結果報告** → Envoy がユーザーに報告

### 通信方式（Mailbox System）

すべて Go で実装。shell script は使用しない。

```go
// メッセージ送信
inbox.Write(target, message, msgType, from)

// 仕組み
1. sync.Mutex で排他ロック取得
2. agents/queue/inbox/<target>.yaml に追記
3. ロック解放
4. fsnotify が変更検知
5. tmux send-keys で target pane に nudge
6. target が inbox YAML を読み込み
```

**特徴:**

- **ゼロポーリング**: fsnotify（カーネルレベル）で API 消費ゼロ
- **排他制御**: Go の sync.Mutex で同時書き込み防止
- **永続化**: YAML ファイルでエージェント再起動を跨いで状態保持
- **nudge 方式**: send-keys は短い wakeup のみ、本文は YAML から読み取り

### 並列実行（tmux + git worktree）

```
Session: bastion
├── Window 0: envoy (main branch)
├── Window 1: marshall (main branch)
└── Window 2: specialists
    ├── Pane 0.0: marshall control
    ├── Pane 0.1: specialist_1 (worktree: .worktrees/sp1)
    ├── Pane 0.2: specialist_2 (worktree: .worktrees/sp2)
    └── ...
```

**worktree 戦略:**

- 各 Specialist は独立した worktree で作業
- ファイル競合を物理的に回避
- 完了後、Marshall がマージ調整

### 依存関係管理

```yaml
# タスク YAML 内
blocks: ["subtask_003"] # このタスク完了まで 003 は開始不可
blocked_by: ["subtask_001"] # 001 の完了を待つ
```

Marshall が依存グラフを管理し、並列可能なタスクを同時 dispatch。

## 設計原則

### 採用したパターン

1. **即時委譲** - Envoy は指示を書いたら即座にターン終了
2. **単一書き込み者** - agents/dashboard.md は Marshall のみ更新
3. **nudge 方式** - send-keys は短い wakeup のみ
4. **YAML 権威** - YAML が正、dashboard は副次情報

### Bastion 独自

1. **Go 堅牢性** - シェルスクリプトの脆弱性を Go で解消
2. **自己強化** - 評価→知識→プロンプト改善の循環
3. **外部注入** - Specialist の動的追加
4. **Claude Code 統合** - Skills/Hooks/Agents との連携
