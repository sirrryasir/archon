package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirrryasir/archon/ai"
)

// SessionManager handles saving and loading history from a JSONL file.
type SessionManager struct {
	SessionID   string
	ProjectPath string
	File        *os.File
}

// NewSessionManager initializes a new or existing session.
func NewSessionManager(sessionID, projectPath string) (*SessionManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sessionDir := filepath.Join(home, ".archon", "sessions")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(sessionDir, fmt.Sprintf("%s.jsonl", sessionID))
	
	// Open in append mode, create if not exists
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &SessionManager{
		SessionID:   sessionID,
		ProjectPath: projectPath,
		File:        file,
	}, nil
}

// AppendMessage appends a single message to the JSONL file.
func (sm *SessionManager) AppendMessage(msg ai.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	
	_, err = sm.File.Write(append(data, '\n'))
	return err
}

// LoadHistory reads the JSONL file and returns the list of messages.
func (sm *SessionManager) LoadHistory() ([]ai.Message, error) {
	var messages []ai.Message
	
	// Rewind to beginning
	_, err := sm.File.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	
	scanner := bufio.NewScanner(sm.File)
	for scanner.Scan() {
		var msg ai.Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err == nil {
			messages = append(messages, msg)
		}
	}
	
	// Seek back to end for future appends
	_, _ = sm.File.Seek(0, 2)
	
	return messages, scanner.Err()
}

// OverwriteHistory replaces the entire history (used during compaction).
func (sm *SessionManager) OverwriteHistory(messages []ai.Message) error {
	// Truncate file
	err := sm.File.Truncate(0)
	if err != nil {
		return err
	}
	_, err = sm.File.Seek(0, 0)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		_, err = sm.File.Write(append(data, '\n'))
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes the underlying file handle.
func (sm *SessionManager) Close() error {
	if sm.File != nil {
		return sm.File.Close()
	}
	return nil
}
