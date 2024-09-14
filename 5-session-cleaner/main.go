//////////////////////////////////////////////////////////////////////
//
// Given is a SessionManager that stores session information in
// memory. The SessionManager itself is working, however, since we
// keep on adding new sessions to the manager our program will
// eventually run out of memory.
//
// Your task is to implement a session cleaner routine that runs
// concurrently in the background and cleans every session that
// hasn't been updated for more than 5 seconds (of course usually
// session times are much longer).
//
// Note that we expect the session to be removed anytime between 5 and
// 7 seconds after the last update. Also, note that you have to be
// very careful in order to prevent race conditions.
//

package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"sync"
	"time"
)

// SessionManager keeps track of all sessions from creation, updating
// to destroying.
type SessionManager struct {
	sessions map[string]Session
	sync     sync.Mutex
	done     chan bool
}

// Session stores the session's data and the last update timestamp.
type Session struct {
	Data       map[string]interface{}
	LastUpdate time.Time
}

// NewSessionManager creates a new sessionManager and starts the session cleaner.
func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessions: make(map[string]Session),
		done:     make(chan bool),
	}

	// Start the session cleaner in the background.
	go m.MangeSession()

	return m
}

// CreateSession creates a new session and returns the sessionID.
func (m *SessionManager) CreateSession() (string, error) {
	m.sync.Lock()
	defer m.sync.Unlock()

	sessionID, err := MakeSessionID()
	if err != nil {
		return "", err
	}

	m.sessions[sessionID] = Session{
		Data:       make(map[string]interface{}),
		LastUpdate: time.Now(), // Set the initial LastUpdate time to now.
	}

	return sessionID, nil
}

// ErrSessionNotFound is returned when sessionID is not listed in the SessionManager.
var ErrSessionNotFound = errors.New("SessionID does not exist")

// GetSessionData returns data related to a session if the sessionID is found.
func (m *SessionManager) GetSessionData(sessionID string) (map[string]interface{}, error) {
	m.sync.Lock()
	defer m.sync.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	return session.Data, nil
}

// UpdateSessionData overwrites the old session data with the new one.
func (m *SessionManager) UpdateSessionData(sessionID string, data map[string]interface{}) error {
	m.sync.Lock()
	defer m.sync.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}

	// Update session data and renew LastUpdate.
	session.Data = data
	session.LastUpdate = time.Now() // Update LastUpdate time.

	m.sessions[sessionID] = session
	return nil
}

// MangeSession runs the session cleaner in the background and removes sessions older than 5 seconds.
func (m *SessionManager) MangeSession() {
	for {
		select {
		case <-m.done:
			log.Println("Stopping session cleaner.")
			return
		default:
			time.Sleep(5 * time.Second)

			m.sync.Lock()

			// Check if there are any sessions to clean up.
			if len(m.sessions) == 0 {
				m.sync.Unlock()
				continue
			}

			var deletedKeys []string

			for id, sess := range m.sessions {
				// Check if the session is older than 5 seconds.
				if time.Since(sess.LastUpdate) >= 5*time.Second {
					deletedKeys = append(deletedKeys, id)
				}
			}

			// Delete expired sessions.
			for _, key := range deletedKeys {
				delete(m.sessions, key)
			}

			if len(deletedKeys) > 0 {
				log.Printf("Deleted session IDs: %v", deletedKeys)
			}

			m.sync.Unlock()

		}

	}
}

func (m *SessionManager) StopSession() {
	time.Sleep(6 * time.Microsecond)
	m.done <- true
	close(m.done) // Signal the session cleaner to stop
}

// MakeSessionID generates a random session ID.
func MakeSessionID() (string, error) {
	buf := make([]byte, 26)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

func main() {
	// Create new sessionManager and new session
	m := NewSessionManager()
	sID, err := m.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Created new session with ID", sID)

	// Update session data
	data := make(map[string]interface{})
	data["website"] = "longhoang.de"

	err = m.UpdateSessionData(sID, data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Update session data, set website to longhoang.de")

	// Retrieve data from manager again
	updatedData, err := m.GetSessionData(sID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Get session data:", updatedData)
	m.StopSession()
}
