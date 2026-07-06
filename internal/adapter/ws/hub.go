package ws

import "context"

// NewHub builds the WebSocket hub that fans out realtime device metrics.
// TODO: implement per TECH-STACK-DAN-PERSIAPAN.md §9 (github.com/coder/websocket
// — the renamed, actively-maintained successor to nhooyr.io/websocket;
// do not import the nhooyr.io path, it is deprecated).
func NewHub(ctx context.Context) error {
	return nil
}
