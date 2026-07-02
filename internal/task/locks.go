package task

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type RunLocks struct {
	taskMu     sync.Mutex
	taskRuns   map[uuid.UUID]struct{}
	volumeMu   sync.Mutex
	volumeRuns map[uuid.UUID]struct{}
	cancelMu   sync.Mutex
	cancels    map[uuid.UUID]context.CancelFunc
}

func NewRunLocks() *RunLocks {
	return &RunLocks{
		taskRuns:   make(map[uuid.UUID]struct{}),
		volumeRuns: make(map[uuid.UUID]struct{}),
		cancels:    make(map[uuid.UUID]context.CancelFunc),
	}
}

func (l *RunLocks) TryAcquireTask(id uuid.UUID) bool {
	l.taskMu.Lock()
	defer l.taskMu.Unlock()
	if _, ok := l.taskRuns[id]; ok {
		return false
	}
	l.taskRuns[id] = struct{}{}
	return true
}

func (l *RunLocks) ReleaseTask(id uuid.UUID) {
	l.taskMu.Lock()
	defer l.taskMu.Unlock()
	delete(l.taskRuns, id)
}

func (l *RunLocks) TryAcquireVolume(id uuid.UUID) bool {
	l.volumeMu.Lock()
	defer l.volumeMu.Unlock()
	if _, ok := l.volumeRuns[id]; ok {
		return false
	}
	l.volumeRuns[id] = struct{}{}
	return true
}

func (l *RunLocks) ReleaseVolume(id uuid.UUID) {
	l.volumeMu.Lock()
	defer l.volumeMu.Unlock()
	delete(l.volumeRuns, id)
}

func (l *RunLocks) RegisterRunCancel(id uuid.UUID, cancel context.CancelFunc) {
	l.cancelMu.Lock()
	defer l.cancelMu.Unlock()
	l.cancels[id] = cancel
}

func (l *RunLocks) UnregisterRunCancel(id uuid.UUID) {
	l.cancelMu.Lock()
	defer l.cancelMu.Unlock()
	delete(l.cancels, id)
}

func (l *RunLocks) CancelRun(id uuid.UUID) {
	l.cancelMu.Lock()
	defer l.cancelMu.Unlock()
	if cancel, ok := l.cancels[id]; ok {
		cancel()
	}
}
