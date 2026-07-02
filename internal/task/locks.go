package task

import (
	"sync"

	"github.com/google/uuid"
)

type RunLocks struct {
	taskMu    sync.Mutex
	taskRuns  map[uuid.UUID]struct{}
	volumeMu  sync.Mutex
	volumeRuns map[uuid.UUID]struct{}
}

func NewRunLocks() *RunLocks {
	return &RunLocks{
		taskRuns:   make(map[uuid.UUID]struct{}),
		volumeRuns: make(map[uuid.UUID]struct{}),
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
