# Bastion クイックスタート

Bastion マルチエージェントシステムの使い方を説明します。

## 前提条件

必須ツールがインストールされていることを確認してください：

```bash
bastion doctor
```

## セッション起動

```bash
# Bastion セッションを起動（Envoy, Marshall, Specialists が自動起動）
bastion start

# オプション: Specialist 数を指定
bastion start --specialists 4
```

## tmux セッションへの接続

```bash
# セッションにアタッチ
tmux attach -t bastion

# セッション内の操作
# - Ctrl+B → 0: Envoy ウィンドウに移動
# - Ctrl+B → 1: Marshall ウィンドウに移動
# - Ctrl+B → 2: Specialists ウィンドウに移動
# - Ctrl+B → D: セッションをデタッチ（バックグラウンド実行継続）
```

## 各エージェントの役割

### Envoy (Window 0)

- **役割**: ユーザーとの唯一の対話窓口
- **機能**:
  - ユーザー要望を受け取り
  - 目的（what）と完了条件（acceptance_criteria）を定義
  - Marshall に即座に委譲

**使用例**:

```
Envoy> ユーザーからの入力を待っています...

ユーザー: 「認証機能を実装してほしい」
Envoy:
1. 目的を定義: JWT認証が動作する
2. 完了条件:
   - POST /auth/login が JWT を返す
   - protected endpoint が JWT 検証する
   - テストがパスする
3. queue/envoy_to_marshall.yaml に指令を書き込み
4. 「Marshall に委譲しました」とユーザーに報告
```

### Marshall (Window 1)

- **役割**: タスク管理・オーケストレーター
- **機能**:
  - Envoy からの指令をタスクに分解
  - 依存関係を分析してスケジューリング
  - Specialists に並列でタスク割当
  - 完了時に品質評価・知識抽出
  - dashboard.md を更新

**ワークフロー**:

```
1. queue/inbox/marshall.yaml を監視
2. Envoy からの指令を読み取り
3. タスクに分解:
   - subtask_001: JWT ミドルウェア実装
   - subtask_002: テスト作成
4. Specialists に割当
5. 完了報告を待つ
6. 評価・知識抽出
7. dashboard.md 更新
```

### Specialists (Window 2)

- **役割**: 専門タスク実行者（複数ペイン）
- **機能**:
  - 割り当てられたペルソナで高品質な成果物作成
  - 完了報告を Marshall に提出
  - スキル候補の発見・報告

**ペルソナ例**:
- Senior Software Engineer
- QA Engineer
- Technical Writer
- Data Analyst

**ワークフロー**:

```
1. queue/tasks/specialist_N.yaml を読み取り
2. ペルソナとして実装
3. テスト実行
4. queue/reports/specialist_N_report.yaml に報告
5. Marshall に完了通知
```

## 通信フロー

```
ユーザー
  │
  ▼
Envoy (定義)
  │ queue/envoy_to_marshall.yaml
  ▼
Marshall (分解・割当)
  │ queue/tasks/*.yaml
  ├─▶ Specialist 1
  ├─▶ Specialist 2
  └─▶ Specialist N
  │ queue/reports/*.yaml
  ▼
Marshall (評価・統合)
  │ dashboard.md
  ▼
Envoy (報告)
  │
  ▼
ユーザー
```

## ファイル構造

```
queue/
├── envoy_to_marshall.yaml      # Envoy → Marshall 指令
├── inbox/
│   ├── marshall.yaml           # Marshall 宛メッセージ
│   └── specialist_*.yaml       # Specialist 宛メッセージ
├── tasks/
│   └── specialist_*.yaml       # タスク定義
└── reports/
    └── specialist_*_report.yaml # 完了報告

dashboard.md                    # 進捗ダッシュボード（Marshall が更新）
```

## セッション管理

```bash
# 状態確認
bastion status

# セッション停止
bastion stop

# 再起動
bastion start
```

## デバッグ

各ウィンドウのログを確認：

```bash
# tmux セッションにアタッチしてウィンドウを切り替え
tmux attach -t bastion

# または、ペインの内容をキャプチャ
tmux capture-pane -t bastion:0 -p  # Envoy
tmux capture-pane -t bastion:1 -p  # Marshall
tmux capture-pane -t bastion:2.0 -p  # Specialist 1
tmux capture-pane -t bastion:2.1 -p  # Specialist 2
```

## 次のステップ

Phase 1 が完了しました。次は Phase 2 で以下を実装予定：

- git worktree 管理（ファイル競合の物理的回避）
- 依存関係解決（blocks / blocked_by）
- 複数 Specialist の本格的な並列実行

## トラブルシューティング

### claude が起動しない

```bash
# claude CLI がインストールされているか確認
which claude

# 手動で起動してみる
cd agents/envoy
claude
```

### tmux セッションが残っている

```bash
# 既存セッションを削除
tmux kill-session -t bastion

# 再起動
bastion start
```

### queue ディレクトリがない

```bash
# 必要なディレクトリを作成
mkdir -p queue/inbox queue/tasks queue/reports
touch dashboard.md
```
