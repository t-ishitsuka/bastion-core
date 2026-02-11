package templates

import "embed"

// FS はエージェントテンプレートファイルを埋め込んだファイルシステム
// all: プレフィックスで隠しファイル・隠しディレクトリも含める
//
//go:embed all:agents
var FS embed.FS
