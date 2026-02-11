package communication

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// ファイル変更イベント
type FileEvent struct {
	Path      string
	Operation string
}

// ファイル変更を監視
type Watcher struct {
	watcher *fsnotify.Watcher
	events  chan FileEvent
	errors  chan error
	done    chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	stopped bool
}

// 新しい Watcher を作成
func NewWatcher() (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return &Watcher{
		watcher: fsWatcher,
		events:  make(chan FileEvent, 100),
		errors:  make(chan error, 10),
		done:    make(chan struct{}),
	}, nil
}

// ディレクトリの監視を開始
func (w *Watcher) Watch(dir string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// ディレクトリを監視対象に追加
	if err := w.watcher.Add(dir); err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	return nil
}

// イベント処理を開始
func (w *Watcher) Start() {
	w.wg.Add(1)
	go w.eventLoop()
}

// イベントループ
func (w *Watcher) eventLoop() {
	defer w.wg.Done()

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// WRITE, CREATE, RENAME イベントを処理
			// WSL2 環境では、エディタが rename 方式でファイルを保存することがある
			if event.Has(fsnotify.Write) {
				w.events <- FileEvent{
					Path:      event.Name,
					Operation: "write",
				}
			} else if event.Has(fsnotify.Create) {
				w.events <- FileEvent{
					Path:      event.Name,
					Operation: "create",
				}
			} else if event.Has(fsnotify.Rename) {
				w.events <- FileEvent{
					Path:      event.Name,
					Operation: "rename",
				}
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.errors <- err

		case <-w.done:
			return
		}
	}
}

// イベントチャンネルを取得
func (w *Watcher) Events() <-chan FileEvent {
	return w.events
}

// エラーチャンネルを取得
func (w *Watcher) Errors() <-chan error {
	return w.errors
}

// 監視を停止
func (w *Watcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 既に停止している場合は何もしない
	if w.stopped {
		return nil
	}

	// done チャンネルをクローズしてイベントループを停止
	close(w.done)

	// watcher をクローズ
	if err := w.watcher.Close(); err != nil {
		return fmt.Errorf("failed to close watcher: %w", err)
	}

	// イベントループの終了を待つ
	w.wg.Wait()

	// チャンネルをクローズ
	close(w.events)
	close(w.errors)

	w.stopped = true

	return nil
}

// inbox ディレクトリを監視
func WatchInbox(queueDir string) (*Watcher, error) {
	watcher, err := NewWatcher()
	if err != nil {
		return nil, err
	}

	inboxDir := filepath.Join(queueDir, "inbox")
	if err := watcher.Watch(inboxDir); err != nil {
		_ = watcher.Stop()
		return nil, err
	}

	watcher.Start()
	return watcher, nil
}
