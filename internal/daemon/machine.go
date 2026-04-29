package daemon

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/peterfox/claude2-d2/internal/r2"
)

type state int

const (
	stateIdle state = iota
	stateThinking
	stateAwaitingResponse
)

var idleAnimations = []string{"idle_1", "idle_2", "idle_3"}

var stopAnimations = []string{
	"emote_excited",
	"emote_laugh",
	"emote_yes",
	"emote_spin",
	"wwm_bow",
	"wwm_happy",
	"wwm_excited",
	"wwm_relieved",
}

var impatientAnimations = []string{
	"emote_angry",
	"emote_annoyed",
	"wwm_angry",
	"wwm_frustrated",
	"wwm_fiery",
	"wwm_jittery",
}

var failureAnimations = []string{
	"emote_angry",
	"emote_alarm",
	"emote_annoyed",
	"emote_no",
	"emote_sad",
	"emote_ion_blast",
	"wwm_angry",
	"wwm_frustrated",
	"wwm_fiery",
	"wwm_ominous",
	"wwm_no",
	"wwm_yelling",
}

const (
	idleInterval        = 30 * time.Second
	angryDelay          = 60 * time.Second
	happyDuration       = 10 * time.Second
	inactivityThreshold = 5 * time.Minute
	inactivityCheckRate = 1 * time.Minute
	permissionDebounce  = 3 * time.Second
)

type Machine struct {
	mu              sync.Mutex
	current         state
	idleTicker      *time.Ticker
	idleStop        chan struct{}
	animCancel      chan struct{}
	angryTimer      *time.Timer
	permissionTimer *time.Timer
	client          *r2.Client
	lastEventAt     time.Time
	reconnect       func() (*r2.Client, error)
	stopKeepAlive   chan struct{}
}

func NewMachine(client *r2.Client, reconnect func() (*r2.Client, error)) *Machine {
	m := &Machine{
		client:        client,
		reconnect:     reconnect,
		lastEventAt:   time.Now(),
		stopKeepAlive: make(chan struct{}),
	}
	go m.keepAliveLoop()
	return m
}

func (m *Machine) Stop() {
	close(m.stopKeepAlive)
}

func (m *Machine) keepAliveLoop() {
	ticker := time.NewTicker(inactivityCheckRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkInactivity()
		case <-m.stopKeepAlive:
			return
		}
	}
}

func (m *Machine) checkInactivity() {
	m.mu.Lock()
	sinceLastEvent := time.Since(m.lastEventAt)
	m.mu.Unlock()

	if sinceLastEvent < inactivityThreshold {
		return
	}

	if err := m.ping(); err != nil {
		fmt.Fprintf(os.Stderr, "claude2-d2: keepalive ping failed (%v), reconnecting...\n", err)
		if err := m.attemptReconnect(); err != nil {
			fmt.Fprintf(os.Stderr, "claude2-d2: reconnect failed: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "claude2-d2: reconnected\n")
		}
	}
}

func (m *Machine) ping() error {
	m.mu.Lock()
	client := m.client
	m.mu.Unlock()
	return client.StopAnimation()
}

func (m *Machine) attemptReconnect() error {
	newClient, err := m.reconnect()
	if err != nil {
		return err
	}
	m.mu.Lock()
	oldClient := m.client
	m.client = newClient
	m.mu.Unlock()
	_ = oldClient.Disconnect()
	return nil
}

func (m *Machine) Dispatch(event string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastEventAt = time.Now()
	m.cancelPermissionLocked()

	switch event {
	case "prompt":
		m.cancelTimedAnimLocked()
		m.stopIdleLocked()
		m.cancelAngryLocked()
		m.current = stateAwaitingResponse
		m.angryTimer = time.AfterFunc(angryDelay, m.onAngryTimeout)

	case "thinking":
		m.cancelTimedAnimLocked()
		m.cancelAngryLocked()
		if m.current != stateThinking {
			m.current = stateThinking
			m.startIdleLoopLocked()
		}

	case "stop":
		m.cancelAngryLocked()
		m.stopIdleLocked()
		m.current = stateIdle
		m.playTimedAnimLocked(r2.Animations[stopAnimations[rand.Intn(len(stopAnimations))]])

	case "session_start":
		m.cancelAngryLocked()
		m.stopIdleLocked()
		m.current = stateIdle
		m.playTimedAnimLocked(r2.Animations["wwm_yoohoo"])

	case "permission_request":
		m.permissionTimer = time.AfterFunc(permissionDebounce, func() {
			m.mu.Lock()
			m.permissionTimer = nil
			m.playTimedAnimLocked(r2.Animations["wwm_anxious"])
			m.mu.Unlock()
		})

	case "stop_failure":
		m.cancelAngryLocked()
		m.stopIdleLocked()
		m.current = stateIdle
		m.playTimedAnimLocked(r2.Animations[failureAnimations[rand.Intn(len(failureAnimations))]])
	}
}

// playTimedAnimLocked plays an animation for happyDuration then stops.
// Cancels any previously running timed animation first.
// Must be called with m.mu held.
func (m *Machine) playTimedAnimLocked(animID byte) {
	m.cancelTimedAnimLocked()

	cancel := make(chan struct{})
	m.animCancel = cancel
	client := m.client

	go func() {
		_ = client.Animate(animID)
		select {
		case <-time.After(happyDuration):
			_ = client.StopAnimation()
		case <-cancel:
		}
	}()
}

func (m *Machine) cancelTimedAnimLocked() {
	if m.animCancel != nil {
		close(m.animCancel)
		m.animCancel = nil
	}
}

func (m *Machine) startIdleLoopLocked() {
	stop := make(chan struct{})
	m.idleStop = stop
	m.idleTicker = time.NewTicker(idleInterval)
	ticker := m.idleTicker

	go func() {
		_ = m.client.Animate(r2.Animations[idleAnimations[rand.Intn(len(idleAnimations))]])
		for {
			select {
			case <-ticker.C:
				_ = m.client.Animate(r2.Animations[idleAnimations[rand.Intn(len(idleAnimations))]])
			case <-stop:
				return
			}
		}
	}()
}

func (m *Machine) stopIdleLocked() {
	if m.idleTicker != nil {
		m.idleTicker.Stop()
		m.idleTicker = nil
	}
	if m.idleStop != nil {
		close(m.idleStop)
		m.idleStop = nil
	}
}

func (m *Machine) cancelAngryLocked() {
	if m.angryTimer != nil {
		m.angryTimer.Stop()
		m.angryTimer = nil
	}
}

func (m *Machine) cancelPermissionLocked() {
	if m.permissionTimer != nil {
		m.permissionTimer.Stop()
		m.permissionTimer = nil
	}
}

func (m *Machine) onAngryTimeout() {
	m.mu.Lock()
	m.angryTimer = nil
	m.current = stateIdle
	m.playTimedAnimLocked(r2.Animations[impatientAnimations[rand.Intn(len(impatientAnimations))]])
	m.mu.Unlock()
}
