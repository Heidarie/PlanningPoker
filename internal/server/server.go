package server

import (
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	rooms    = make(map[string]*Room)
	roomsMu  sync.RWMutex
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
)

func createRoomHandler(w http.ResponseWriter, r *http.Request) {
	code := randCode(4)
	room := NewRoom(code)
	roomsMu.Lock()
	rooms[code] = room
	roomsMu.Unlock()
	w.Write([]byte(code))
	log.Printf("Create a room %v", code)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	name := r.URL.Query().Get("name")
	mode := r.URL.Query().Get("mode")
	if code == "" || (mode == "join" && name == "") {
		http.Error(w, "code and name required", http.StatusBadRequest)
		return
	}

	// Check if room exists BEFORE upgrading to WebSocket
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

func randCode(n int) string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func cleanupInactiveRooms() {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		roomsMu.Lock()
		for code, room := range rooms {
			room.mu.Lock()
			inactive := time.Since(room.lastActivity) > 10*time.Minute
			empty := len(room.clients) == 0
			room.mu.Unlock()

			if inactive || empty {
				log.Printf("Cleaning up room: %s (inactive: %v, empty: %v)", code, inactive, empty)
				// Close all client connections
				for client := range room.clients {
					close(client.Send)
					client.Conn.Close()
				}
				delete(rooms, code)
			}
		}
		roomsMu.Unlock()
	}
}

func Run(addr string) {
	// Start the cleanup routine
	go cleanupInactiveRooms()

	r := mux.NewRouter()
	r.HandleFunc("/create", createRoomHandler).Methods("POST")
	r.HandleFunc("/ws", wsHandler)
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal("Server run error")
	}
}
