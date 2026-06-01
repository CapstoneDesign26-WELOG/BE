package notification

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	"welog/internal/auth"
)

type NotificationService struct {
	clients map[uint]chan string
	mu      sync.RWMutex
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		clients: make(map[uint]chan string),
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

	clientChan := make(chan string, 1)
	s.mu.Lock()
	s.clients[userClaims.UserID] = clientChan
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		if s.clients[userClaims.UserID] == clientChan {
			delete(s.clients, userClaims.UserID)
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
	if clientChan, ok := s.clients[userID]; ok {
		select {
		case clientChan <- message:
		default:
			fmt.Printf("[SSE] 유저 %d의 채널이 꽉 차서 알림이 누락되었습니다.\n", userID)
		}
	}
}
