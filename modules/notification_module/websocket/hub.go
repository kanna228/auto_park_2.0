package websocket

import "sync"

type Hub struct {
	mu      sync.RWMutex
	clients map[int64]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int64]map[*Client]struct{}),
	}
}

func (h *Hub) Register(client *Client) {
	if client == nil || client.UserID <= 0 {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.UserID]; !ok {
		h.clients[client.UserID] = make(map[*Client]struct{})
	}
	h.clients[client.UserID][client] = struct{}{}
}

func (h *Hub) Unregister(client *Client) {
	if client == nil || client.UserID <= 0 {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	userClients, ok := h.clients[client.UserID]
	if !ok {
		return
	}

	if _, exists := userClients[client]; exists {
		delete(userClients, client)
		close(client.Send)
	}

	if len(userClients) == 0 {
		delete(h.clients, client.UserID)
	}
}

func (h *Hub) SendToUser(userID int64, payload []byte) {
	if userID <= 0 || len(payload) == 0 {
		return
	}

	h.mu.RLock()
	userClients := h.clients[userID]
	clients := make([]*Client, 0, len(userClients))
	for client := range userClients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- payload:
		default:
			h.Unregister(client)
		}
	}
}
