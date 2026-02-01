// Package inbox provides a file-based message queue for session IPC.
// Each session has an inbox directory where messages can be posted.
// Messages are processed in creation order (oldest first).
package inbox

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Message represents a message in the inbox.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Created time.Time       `json:"created"`
}

// Message types
const (
	TypeFocus = "focus" // Request to focus the session's terminal window
)

// inboxDir returns the inbox directory for a session.
func inboxDir(sessionID string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".tofu", "claude-sessions", sessionID, "inbox")
}

// EnsureInbox creates the inbox directory if it doesn't exist.
func EnsureInbox(sessionID string) error {
	dir := inboxDir(sessionID)
	if dir == "" {
		return fmt.Errorf("could not determine inbox directory")
	}
	return os.MkdirAll(dir, 0755)
}

// Post sends a message to a session's inbox.
func Post(sessionID string, msgType string, payload any) error {
	if err := EnsureInbox(sessionID); err != nil {
		return err
	}

	var payloadBytes json.RawMessage
	if payload != nil {
		var err error
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	msg := Message{
		Type:    msgType,
		Payload: payloadBytes,
		Created: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Use timestamp + random suffix for unique filename
	filename := fmt.Sprintf("%d-%d.msg", time.Now().UnixNano(), os.Getpid())
	msgPath := filepath.Join(inboxDir(sessionID), filename)

	return os.WriteFile(msgPath, data, 0644)
}

// ReadAll reads all messages from a session's inbox in creation order.
// Messages are deleted after reading.
func ReadAll(sessionID string) ([]Message, error) {
	dir := inboxDir(sessionID)
	if dir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// Filter and sort by name (which includes timestamp)
	var msgFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".msg" {
			msgFiles = append(msgFiles, entry.Name())
		}
	}
	sort.Strings(msgFiles) // Natural order = creation order due to timestamp prefix

	var messages []Message
	for _, name := range msgFiles {
		msgPath := filepath.Join(dir, name)
		data, err := os.ReadFile(msgPath)
		if err != nil {
			continue
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			// Invalid message, delete it
			os.Remove(msgPath)
			continue
		}

		messages = append(messages, msg)

		// Delete after reading
		os.Remove(msgPath)
	}

	return messages, nil
}

// Watcher watches an inbox directory for new messages.
type Watcher struct {
	sessionID string
	watcher   *fsnotify.Watcher
	handler   func(Message)
	done      chan struct{}
}

// NewWatcher creates a new inbox watcher for a session.
func NewWatcher(sessionID string, handler func(Message)) (*Watcher, error) {
	if err := EnsureInbox(sessionID); err != nil {
		return nil, err
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	dir := inboxDir(sessionID)
	if err := fsw.Add(dir); err != nil {
		fsw.Close()
		return nil, fmt.Errorf("failed to watch inbox: %w", err)
	}

	w := &Watcher{
		sessionID: sessionID,
		watcher:   fsw,
		handler:   handler,
		done:      make(chan struct{}),
	}

	return w, nil
}

// Start begins watching for messages. This blocks until Stop is called.
func (w *Watcher) Start() {
	// First, process any existing messages
	w.processAll()

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				// Small delay to ensure file is fully written
				time.Sleep(10 * time.Millisecond)
				w.processAll()
			}
		case <-w.watcher.Errors:
			// Ignore errors, keep watching
		case <-w.done:
			return
		}
	}
}

// StartAsync starts watching in a background goroutine.
func (w *Watcher) StartAsync() {
	go w.Start()
}

// Stop stops the watcher.
func (w *Watcher) Stop() {
	close(w.done)
	w.watcher.Close()
}

// processAll reads and handles all messages in the inbox.
func (w *Watcher) processAll() {
	messages, err := ReadAll(w.sessionID)
	if err != nil {
		return
	}
	for _, msg := range messages {
		w.handler(msg)
	}
}

// Cleanup removes the inbox directory for a session.
func Cleanup(sessionID string) error {
	dir := inboxDir(sessionID)
	if dir == "" {
		return nil
	}
	// Remove the inbox directory and its contents
	return os.RemoveAll(dir)
}
