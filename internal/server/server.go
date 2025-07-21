package server

import (
	"crypto/subtle"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	rooms        = make(map[string]*Room)
	roomsMu      sync.RWMutex
	upgrader     = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	rateLimiter  = make(map[string]time.Time)
	rateMu       sync.RWMutex
	clientSecret string
)

func init() {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Get client secret from environment or use default for development
	clientSecret = os.Getenv("CLIENT_SECRET")
	if clientSecret == "" {
		log.Println("Warning: CLIENT_SECRET not set, using default (not secure for production)")
		clientSecret = "planning-poker-secure-key-2025"
	}
}

func getSecretKey() string {
	return clientSecret
}

func authenticateClient(r *http.Request) bool {
	clientSecret := r.Header.Get("X-Client-Secret")
	expectedSecret := getSecretKey()

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(clientSecret), []byte(expectedSecret)) == 1
}

func rateLimitCheck(ip string) bool {
	rateMu.Lock()
	defer rateMu.Unlock()

	if lastReq, exists := rateLimiter[ip]; exists {
		if time.Since(lastReq) < time.Second { // 1 request per second per IP
			return false
		}
	}
	rateLimiter[ip] = time.Now()
	return true
}

func cleanupRateLimit() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rateMu.Lock()
		for ip, lastReq := range rateLimiter {
			if time.Since(lastReq) > 5*time.Minute {
				delete(rateLimiter, ip)
			}
		}
		rateMu.Unlock()
	}
}

func createRoomHandler(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	if !authenticateClient(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Rate limiting
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
	// Check authentication first
	if !authenticateClient(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Rate limiting
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

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Initialize router
	router := mux.NewRouter()
	router.HandleFunc("/create", createRoomHandler).Methods("POST")
	router.HandleFunc("/ws", wsHandler)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Handle the request
	router.ServeHTTP(w, r)
}

func Run(addr string) {
	// Start the cleanup routines
	go cleanupInactiveRooms()
	go cleanupRateLimit()

	r := mux.NewRouter()
	r.HandleFunc("/create", createRoomHandler).Methods("POST")
	r.HandleFunc("/ws", wsHandler)

	// Health check endpoint for Vercel
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal("Server run error")
	}
}
