package server

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

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
			// Send current participant list to the new client
			r.sendParticipantList(client)
			// Notify all clients about the new participant (but not if it's the host)
			if !client.IsHost {
				r.broadcastParticipantUpdate("participant_join", client.Name)
			}

		case client := <-r.unregister:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.Send)
				r.updateActivity()
				// Send participant leave notification (but not if it's the host)
				if !client.IsHost {
					r.broadcastParticipantUpdate("participant_leave", client.Name)
				}

				// If no clients left, the room should be cleaned up
				if len(r.clients) == 0 {
					log.Printf("Room %s is now empty, will be cleaned up by periodic cleanup", r.code)
					return // Exit the goroutine
				}
			}

		case msg := <-r.broadcast:
			r.updateActivity()
			switch msg.Type {
			case "vote":
				var p struct{ Value string }
				json.Unmarshal(msg.Payload, &p)

				// Find the sender client to check if they're the host
				var senderClient *Client
				for c := range r.clients {
					if c.Name == msg.Sender {
						senderClient = c
						break
					}
				}

				// Reject vote if sender is the host
				if senderClient != nil && senderClient.IsHost {
					senderClient.Send <- Message{Type: "vote_rejected", Payload: []byte(`{"reason":"host_cannot_vote"}`)}
					return
				}

				r.mu.Lock()
				// Check if user already voted
				if _, alreadyVoted := r.votes[msg.Sender]; alreadyVoted {
					r.mu.Unlock()
					// Send vote rejection to the sender
					if senderClient != nil {
						senderClient.Send <- Message{Type: "vote_rejected", Payload: []byte(`{"reason":"already_voted"}`)}
					}
				} else {
					// Accept the vote
					r.votes[msg.Sender] = p.Value

					// Count non-host participants for vote completion check
					nonHostCount := 0
					for c := range r.clients {
						if !c.IsHost {
							nonHostCount++
						}
					}

					all := len(r.votes) == nonHostCount
					r.mu.Unlock()

					// Send vote confirmation to the sender
					if senderClient != nil {
						senderClient.Send <- Message{Type: "vote_confirmed"}
					}

					// Send vote progress update to all clients
					progressPayload, _ := json.Marshal(map[string]int{
						"current": len(r.votes),
						"total":   nonHostCount,
					})
					progressMsg := Message{Type: "vote_progress", Payload: progressPayload}
					for c := range r.clients {
						c.Send <- progressMsg
					}

					// Check if all non-host participants have voted
					if all {
						out, _ := json.Marshal(r.votes)
						reveal := Message{Type: "reveal", Payload: out}
						for c := range r.clients {
							c.Send <- reveal
						}
					}
				}

			case "reset":
				// Find the sender client to check if they're the host
				var senderClient *Client
				for c := range r.clients {
					if c.Name == msg.Sender {
						senderClient = c
						break
					}
				}

				// Only allow hosts to reset
				if senderClient == nil || !senderClient.IsHost {
					if senderClient != nil {
						senderClient.Send <- Message{Type: "error", Payload: []byte(`{"reason":"only_host_can_reset"}`)}
					}
					return
				}

				r.mu.Lock()
				r.votes = make(map[string]string)
				r.mu.Unlock()

				// Reset vote progress for all clients
				for c := range r.clients {
					c.Send <- Message{Type: "reset"}
				}

			case "start":
				// Find the sender client to check if they're the host
				var senderClient *Client
				for c := range r.clients {
					if c.Name == msg.Sender {
						senderClient = c
						break
					}
				}

				// Only allow hosts to start voting
				if senderClient == nil || !senderClient.IsHost {
					if senderClient != nil {
						senderClient.Send <- Message{Type: "error", Payload: []byte(`{"reason":"only_host_can_start"}`)}
					}
					return
				}

				// Clear previous votes and send start message to all clients
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
		// Only include non-host participants in the list
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
