package consolepump

import (
	"fmt"
	"sync"
	"sync/atomic"

	websocket "github.com/fasthttp/websocket"
)

// ConsolePump handles serving the current logs on the dashboard through Websocket.
type ConsolePump struct {
	ConnsMu sync.RWMutex              `json:"-"`
	Conns   map[int64]*websocket.Conn `json:"-"`

	ConnMuMu sync.RWMutex          `json:"-"`
	ConnMu   map[int64]*sync.Mutex `json:"-"`

	iter *int64
}

// NewConsolePump creates a new console pump.
func NewConsolePump() *ConsolePump {
	return &ConsolePump{
		ConnsMu: sync.RWMutex{},
		Conns:   make(map[int64]*websocket.Conn),

		ConnMuMu: sync.RWMutex{},
		ConnMu:   make(map[int64]*sync.Mutex),

		iter: new(int64),
	}
}

// Write implements io.Writer.
func (cp *ConsolePump) Write(p []byte) (n int, err error) {
	message, err := websocket.NewPreparedMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, fmt.Errorf("failed to create prepared message: %w", err)
	}

	cp.ConnsMu.RLock()
	cp.ConnMuMu.RLock()
	for id, conn := range cp.Conns {
		cp.ConnMu[id].Lock()
		conn.WritePreparedMessage(message) // nolint
		cp.ConnMu[id].Unlock()
	}
	cp.ConnMuMu.RUnlock()
	cp.ConnsMu.RUnlock()

	return len(p), nil
}

// RegisterConnection will take the already established websocket and add it to the pump.
func (cp *ConsolePump) RegisterConnection(conn *websocket.Conn) (id int64) {
	id = atomic.AddInt64(cp.iter, 1)

	cp.ConnsMu.Lock()
	defer cp.ConnsMu.Unlock()

	cp.ConnMuMu.Lock()
	defer cp.ConnMuMu.Unlock()

	cp.Conns[id] = conn
	cp.ConnMu[id] = &sync.Mutex{}

	return
}

// DeregisterConnection removes a connection from the pump. Returns boolean if it was removed.
func (cp *ConsolePump) DeregisterConnection(id int64) (ok bool) {
	cp.ConnsMu.Lock()
	defer cp.ConnsMu.Unlock()

	cp.ConnMuMu.Lock()
	defer cp.ConnMuMu.Unlock()

	_, ok = cp.Conns[id]
	if ok {
		delete(cp.Conns, id)
	}

	delete(cp.ConnMu, id)

	return
}
