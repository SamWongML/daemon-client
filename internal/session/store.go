package session

import "sync"

type Store struct {
	mu    sync.RWMutex
	order []ID
	byID  map[ID]*Session
}

func NewStore() *Store {
	return &Store{byID: map[ID]*Session{}}
}

func (s *Store) Add(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byID[sess.ID]; !ok {
		s.order = append(s.order, sess.ID)
	}
	s.byID[sess.ID] = sess
}

func (s *Store) Get(id ID) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.byID[id]
}

func (s *Store) All() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Session, 0, len(s.order))
	for _, id := range s.order {
		if sess, ok := s.byID[id]; ok {
			out = append(out, sess)
		}
	}
	return out
}

func (s *Store) SetStatus(id ID, st Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.byID[id]; ok {
		sess.Status = st
	}
}

func (s *Store) Append(id ID, chunk string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.byID[id]; ok {
		sess.Transcript += chunk
	}
}
