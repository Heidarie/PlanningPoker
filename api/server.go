package handler

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Types
type Message struct {
	Type    string          `json:"type"`
	Sender  string          `json:"sender,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Client struct {
	Conn   *websocket.Conn
	Send   chan Message
	Name   string
	Room   *Room
	IsHost bool
}

type Room struct {
	code         string
	clients      map[*Client]bool
	broadcast    chan Message
	register     chan *Client
	unregister   chan *Client
	votes        map[string]string
	mu           sync.Mutex
	lastActivity time.Time
}

// Global variables
var (
	rooms        = make(map[string]*Room)
	roomsMu      sync.RWMutex
	upgrader     = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	rateLimiter  = make(map[string]time.Time)
	rateMu       sync.RWMutex
	clientSecret string
)

func init() {
	// Get client secret from environment or use default for development
	clientSecret = os.Getenv("CLIENT_SECRET")
	if clientSecret == "" {
		log.Println("Warning: CLIENT_SECRET not set, using default (not secure for production)")
		clientSecret = "planning-poker-secure-key-2025"
	}
}

// Client methods
func (c *Client) readPump() {
	defer func() { c.Room.unregister <- c; c.Conn.Close() }()
	for {
		var msg Message
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}
		msg.Sender = c.Name
		c.Room.broadcast <- msg
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		c.Conn.WriteJSON(msg)
	}
}

// Room methods
func NewRoom(code string) *Room {
	r := &Room{
		code:         code,
		clients:      make(map[*Client]bool),
		broadcast:    make(chan Message),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		votes:        make(map[string]string),
		lastActivity: time.Now(),
	}
	go r.run()
	return r
}

func (r *Room) updateActivity() {
	r.mu.Lock()
	r.lastActivity = time.Now()
	r.mu.Unlock()
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.register:
			r.clients[client] = true
			r.updateActivity()
			r.sendParticipantList(client)
			if !client.IsHost {
				r.broadcastParticipantUpdate("participant_join", client.Name)
			}

		case client := <-r.unregister:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.Send)
				r.updateActivity()
				if !client.IsHost {
					r.broadcastParticipantUpdate("participant_leave", client.Name)
				}
				if len(r.clients) == 0 {
					log.Printf("Room %s is now empty, will be cleaned up by periodic cleanup", r.code)
					return
				}
			}

		case msg := <-r.broadcast:
			r.updateActivity()
			switch msg.Type {
			case "vote":
				var p struct{ Value string }
				json.Unmarshal(msg.Payload, &p)

				var senderClient *Client
				for c := range r.clients {
					if c.Name == msg.Sender {
						senderClient = c
						break
					}
				}

				if senderClient != nil && senderClient.IsHost {
					senderClient.Send <- Message{Type: "vote_rejected", Payload: []byte(`{"reason":"host_cannot_vote"}`)}
					return
				}

				r.mu.Lock()
				if _, alreadyVoted := r.votes[msg.Sender]; alreadyVoted {
					r.mu.Unlock()
					if senderClient != nil {
						senderClient.Send <- Message{Type: "vote_rejected", Payload: []byte(`{"reason":"already_voted"}`)}
					}
				} else {
					r.votes[msg.Sender] = p.Value

					nonHostCount := 0
					for c := range r.clients {
						if !c.IsHost {
							nonHostCount++
						}
					}

					all := len(r.votes) == nonHostCount
					r.mu.Unlock()

					if senderClient != nil {
						senderClient.Send <- Message{Type: "vote_confirmed"}
					}

					progressPayload, _ := json.Marshal(map[string]int{
						"current": len(r.votes),
						"total":   nonHostCount,
					})
					progressMsg := Message{Type: "vote_progress", Payload: progressPayload}
					for c := range r.clients {
						c.Send <- progressMsg
					}

					if all {
						out, _ := json.Marshal(r.votes)
						reveal := Message{Type: "reveal", Payload: out}
						for c := range r.clients {
							c.Send <- reveal
						}
					}
				}

			case "reset":
				var senderClient *Client
				for c := range r.clients {
					if c.Name == msg.Sender {
						senderClient = c
						break
					}
				}

				if senderClient == nil || !senderClient.IsHost {
					if senderClient != nil {
						senderClient.Send <- Message{Type: "error", Payload: []byte(`{"reason":"only_host_can_reset"}`)}
					}
					return
				}

				r.mu.Lock()
				r.votes = make(map[string]string)
				r.mu.Unlock()

				for c := range r.clients {
					c.Send <- Message{Type: "reset"}
				}

			case "start":
				var senderClient *Client
				for c := range r.clients {
					if c.Name == msg.Sender {
						senderClient = c
						break
					}
				}

				if senderClient == nil || !senderClient.IsHost {
					if senderClient != nil {
						senderClient.Send <- Message{Type: "error", Payload: []byte(`{"reason":"only_host_can_start"}`)}
					}
					return
				}

				r.mu.Lock()
				r.votes = make(map[string]string)
				r.mu.Unlock()

				for c := range r.clients {
					c.Send <- Message{Type: "start"}
				}

			default:
				for c := range r.clients {
					c.Send <- msg
				}
			}
		}
	}
}

func (r *Room) broadcastParticipantUpdate(msgType, participantName string) {
	payload, _ := json.Marshal(map[string]string{"name": participantName})
	msg := Message{
		Type:    msgType,
		Payload: payload,
	}
	for c := range r.clients {
		c.Send <- msg
	}
}

func (r *Room) sendParticipantList(client *Client) {
	var participants []string
	for c := range r.clients {
		if !c.IsHost {
			participants = append(participants, c.Name)
		}
	}
	payload, _ := json.Marshal(participants)
	msg := Message{
		Type:    "participant_list",
		Payload: payload,
	}
	client.Send <- msg
}

// Utility functions
func getSecretKey() string {
	return clientSecret
}

func authenticateClient(r *http.Request) bool {
	clientSecret := r.Header.Get("X-Client-Secret")
	expectedSecret := getSecretKey()
	return subtle.ConstantTimeCompare([]byte(clientSecret), []byte(expectedSecret)) == 1
}

func rateLimitCheck(ip string) bool {
	rateMu.Lock()
	defer rateMu.Unlock()

	if lastReq, exists := rateLimiter[ip]; exists {
		if time.Since(lastReq) < time.Second {
			return false
		}
	}
	rateLimiter[ip] = time.Now()
	return true
}

func randCode(n int) string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Handlers
func createRoomHandler(w http.ResponseWriter, r *http.Request) {
	if !authenticateClient(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}
	if !rateLimitCheck(clientIP) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	code := randCode(4)
	room := NewRoom(code)
	roomsMu.Lock()
	rooms[code] = room
	roomsMu.Unlock()
	w.Write([]byte(code))
	log.Printf("Create a room %v from IP %v", code, clientIP)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	if !authenticateClient(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}
	if !rateLimitCheck(clientIP) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	code := r.URL.Query().Get("code")
	name := r.URL.Query().Get("name")
	mode := r.URL.Query().Get("mode")
	if code == "" || (mode == "join" && name == "") {
		http.Error(w, "code and name required", http.StatusBadRequest)
		return
	}

	roomsMu.RLock()
	room, ok := rooms[code]
	roomsMu.RUnlock()
	if !ok {
		http.Error(w, "room does not exist", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "could not upgrade", http.StatusInternalServerError)
		return
	}

	client := &Client{
		Conn:   conn,
		Send:   make(chan Message),
		Name:   name,
		Room:   room,
		IsHost: mode == "create",
	}
	room.register <- client

	go client.writePump()
	go client.readPump()
}

// Main handler
func Handler(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()
	router.HandleFunc("/create", createRoomHandler).Methods("POST")
	router.HandleFunc("/ws", wsHandler)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	router.ServeHTTP(w, r)
}
