package browser

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// CDPClient is a minimal Chrome DevTools Protocol client over a WebSocket.
// It supports the handful of commands needed to read open tabs: target
// enumeration, attachment, and JavaScript evaluation.
type CDPClient struct {
	conn    *websocket.Conn
	mu      sync.Mutex
	nextID  int
	pending map[int]chan *cdpMessage
	readErr error
	once    sync.Once
	closed  chan struct{}
}

type cdpMessage struct {
	ID        int             `json:"id,omitempty"`
	SessionID string          `json:"sessionId,omitempty"`
	Method    string          `json:"method,omitempty"`
	Params    json.RawMessage `json:"params,omitempty"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     *cdpError       `json:"error,omitempty"`
}

type cdpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Target describes a single browser tab / page.
type Target struct {
	TargetID         string `json:"targetId"`
	Type             string `json:"type"`
	Title            string `json:"title"`
	URL              string `json:"url"`
	Attached         bool   `json:"attached"`
	BrowserContextID string `json:"browserContextId"`
}

// Dial connects to the DevTools HTTP endpoint and opens the debugger socket.
func Dial(endpoint string) (*CDPClient, error) {
	v, err := Version(endpoint)
	if err != nil {
		return nil, err
	}
	if v.WebSocketDebuggerURL == "" {
		return nil, fmt.Errorf("devtools websocket url yok")
	}
	conn, _, err := websocket.DefaultDialer.Dial(v.WebSocketDebuggerURL, nil)
	if err != nil {
		return nil, fmt.Errorf("devtools websocket connection: %w", err)
	}
	c := &CDPClient{
		conn:    conn,
		pending: make(map[int]chan *cdpMessage),
		closed:  make(chan struct{}),
	}
	go c.readLoop()
	return c, nil
}

// Close terminates the WebSocket connection.
func (c *CDPClient) Close() {
	c.once.Do(func() {
		close(c.closed)
		c.conn.Close()
	})
}

func (c *CDPClient) readLoop() {
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			c.mu.Lock()
			c.readErr = err
			pending := make([]chan *cdpMessage, 0, len(c.pending))
			for _, ch := range c.pending {
				pending = append(pending, ch)
			}
			c.mu.Unlock()
			for _, ch := range pending {
				ch <- &cdpMessage{Error: &cdpError{Code: -1, Message: "websocket closed"}}
			}
			return
		}
		var msg cdpMessage
		if json.Unmarshal(data, &msg) != nil {
			continue
		}
		if msg.ID == 0 {
			continue // ignore events
		}
		c.mu.Lock()
		ch, ok := c.pending[msg.ID]
		c.mu.Unlock()
		if ok {
			ch <- &msg
		}
	}
}

func (c *CDPClient) send(method, sessionID string, params any) (json.RawMessage, error) {
	c.mu.Lock()
	id := c.nextID
	c.nextID++
	ch := make(chan *cdpMessage, 1)
	c.pending[id] = ch
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	req := cdpMessage{ID: id, Method: method, SessionID: sessionID}
	if params != nil {
		p, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		req.Params = p
	}
	if err := c.conn.WriteJSON(req); err != nil {
		return nil, err
	}

	select {
	case msg := <-ch:
		if msg.Error != nil {
			return nil, fmt.Errorf("cdp hata %d: %s", msg.Error.Code, msg.Error.Message)
		}
		return msg.Result, nil
	case <-c.closed:
		return nil, fmt.Errorf("cdp client closed")
	case <-time.After(20 * time.Second):
		return nil, fmt.Errorf("cdp timeout: %s", method)
	}
}

// Send issues a command on the default (browser) session.
func (c *CDPClient) Send(method string, params any) (json.RawMessage, error) {
	return c.send(method, "", params)
}

// SendInSession issues a command on an attached target's session.
func (c *CDPClient) SendInSession(sessionID, method string, params any) (json.RawMessage, error) {
	return c.send(method, sessionID, params)
}

// GetTargets returns all open browser targets (tabs, pages, workers, …).
func (c *CDPClient) GetTargets() ([]Target, error) {
	res, err := c.Send("Target.getTargets", nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Targets []Target `json:"targets"`
	}
	if err := json.Unmarshal(res, &out); err != nil {
		return nil, err
	}
	return out.Targets, nil
}

// Attach attaches to a target and returns the session id used for commands.
func (c *CDPClient) Attach(targetID string) (string, error) {
	res, err := c.Send("Target.attachToTarget", map[string]any{
		"targetId": targetID,
		"flatten":  true,
	})
	if err != nil {
		return "", err
	}
	var out struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.Unmarshal(res, &out); err != nil {
		return "", err
	}
	return out.SessionID, nil
}

// Evaluate runs JavaScript in the given target session and returns the
// JSON-encoded value (returnByValue). A runtime error in the page is returned
// as a Go error.
func (c *CDPClient) Evaluate(sessionID, expr string) (json.RawMessage, error) {
	if _, err := c.SendInSession(sessionID, "Runtime.enable", nil); err != nil {
		return nil, err
	}
	res, err := c.SendInSession(sessionID, "Runtime.evaluate", map[string]any{
		"expression":    expr,
		"returnByValue": true,
		"awaitPromise":  true,
	})
	if err != nil {
		return nil, err
	}
	var out struct {
		Result           json.RawMessage `json:"result"`
		ExceptionDetails *struct {
			Text string `json:"text"`
		} `json:"exceptionDetails"`
	}
	if err := json.Unmarshal(res, &out); err != nil {
		return nil, err
	}
	if out.ExceptionDetails != nil {
		return nil, fmt.Errorf("page JS error: %s", out.ExceptionDetails.Text)
	}
	var robj struct {
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(out.Result, &robj); err != nil {
		return nil, err
	}
	return robj.Value, nil
}
