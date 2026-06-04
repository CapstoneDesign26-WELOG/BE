package notification

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	"welog/internal/auth"
)

type NotificationService struct {
	clients map[uint]map[chan string]bool
	mu      sync.RWMutex
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		clients: make(map[uint]map[chan string]bool),
	}
}

func (s *NotificationService) Subscribe(w http.ResponseWriter, r *http.Request) {
	userClaims := auth.GetUserFromContext(r.Context())
	if userClaims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	clientChan := make(chan string, 5)

	s.mu.Lock()
	if s.clients[userClaims.UserID] == nil {
		s.clients[userClaims.UserID] = make(map[chan string]bool)
	}
	s.clients[userClaims.UserID][clientChan] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		if chans, exists := s.clients[userClaims.UserID]; exists {
			delete(chans, clientChan)

			if len(chans) == 0 {
				delete(s.clients, userClaims.UserID)
			}
		}
		close(clientChan)
		s.mu.Unlock()
	}()

	fmt.Fprintf(w, "data: {\"type\": \"CONNECTED\"}\n\n")
	w.(http.Flusher).Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.(http.Flusher).Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *NotificationService) Notify(userID uint, message string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chans, ok := s.clients[userID]
	if !ok || len(chans) == 0 {
		return
	}

	for clientChan := range chans {
		select {
		case clientChan <- message:
		default:
			fmt.Printf("[SSE] 유저 %d의 특정 탭 채널이 꽉 차서 알림이 누락되었습니다.\n", userID)
		}
	}
}
