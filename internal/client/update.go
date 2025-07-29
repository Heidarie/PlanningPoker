package client

import (
	"encoding/json"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea" // TUI framework
)

type wsMsg struct {
	Msg Message
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update list dimensions when window resizes
		if m.state == stateVoting {
			m.list.SetWidth(msg.Width)
			m.list.SetHeight(msg.Height - 3) // Leave space for title
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			// Only hosts can start voting sessions
			if m.state == stateHost {
				startMsg := Message{Type: "start"}
				if err := m.Conn.WriteJSON(startMsg); err != nil {
					log.Println("Failed to send start msg.")
				}
			}
			if m.state == stateVoting {
				// Only allow voting if user hasn't voted yet
				if !m.hasVoted {
					selected := m.list.SelectedItem()
					vote := fmt.Sprintf("%v", selected)
					payload, _ := json.Marshal(map[string]string{"value": vote})
					voteMsg := Message{
						Type:    "vote",
						Sender:  m.name,
						Payload: payload,
					}
					if err := m.Conn.WriteJSON(voteMsg); err != nil {
						log.Println("Failed to send vote msg.")
					}
				}
			}
			if m.state == stateResults {
				// Only hosts can reset the game
				if m.isHost {
					resetMsg := Message{Type: "reset"}
					if err := m.Conn.WriteJSON(resetMsg); err != nil {
						log.Println("Failed to send reset msg.")
					}
				}
			}
		default:
			// In voting state, let the list handle other key events only if not voted yet
			if m.state == stateVoting && !m.hasVoted {
				m.list, cmd = m.list.Update(msg)
			}
		}

	case wsMsg:
		event := msg.Msg
		switch event.Type {
		case "start":
			if m.isHost {
				// Host doesn't vote, just waits for others to vote
				m.state = stateHost
			} else {
				// Non-host participants go to voting state
				m.state = stateVoting
			}
			m.hasVoted = false // Reset vote status when new round starts
			// Reset vote progress
			m.voteProgress.current = 0
			m.voteProgress.total = 0

		case "reveal":
			var votes map[string]string
			if err := json.Unmarshal(event.Payload, &votes); err != nil {
				log.Println("Invalid reveal payload:", err)
			} else {
				m.votes = votes
				m.state = stateResults
			}

		case "reset":
			m.votes = make(map[string]string)
			if m.isHost {
				m.state = stateHost
			} else {
				m.state = stateLobby
			}
			m.hasVoted = false // Reset vote status when game resets
			// Reset vote progress
			m.voteProgress.current = 0
			m.voteProgress.total = 0

		case "vote_confirmed":
			// Server confirms vote was received, lock the user's vote
			m.hasVoted = true

		case "vote_rejected":
			// Vote was rejected (probably already voted), don't change state
			log.Println("Vote rejected: already voted")

		case "vote_progress":
			var progress struct {
				Current int `json:"current"`
				Total   int `json:"total"`
			}
			if err := json.Unmarshal(event.Payload, &progress); err != nil {
				log.Println("Invalid vote_progress payload:", err)
			} else {
				m.voteProgress.current = progress.Current
				m.voteProgress.total = progress.Total
			}

		case "participant_join":
			var participant struct{ Name string }
			if err := json.Unmarshal(event.Payload, &participant); err != nil {
				log.Println("Invalid participant_join payload:", err)
			} else {
				// Add participant if not already in list
				found := false
				for _, p := range m.participants {
					if p == participant.Name {
						found = true
						break
					}
				}
				if !found {
					m.participants = append(m.participants, participant.Name)
				}
			}

		case "participant_leave":
			var participant struct{ Name string }
			if err := json.Unmarshal(event.Payload, &participant); err != nil {
				log.Println("Invalid participant_leave payload:", err)
			} else {
				// Remove participant from list
				for i, p := range m.participants {
					if p == participant.Name {
						m.participants = append(m.participants[:i], m.participants[i+1:]...)
						break
					}
				}
			}

		case "participant_list":
			var participants []string
			if err := json.Unmarshal(event.Payload, &participants); err != nil {
				log.Println("Invalid participant_list payload:", err)
			} else {
				m.participants = participants
			}
		}
		// Continue listening for more messages
		return m, func() tea.Cmd {
			return func() tea.Msg {
				var msg Message
				err := m.Conn.ReadJSON(&msg)
				if err != nil {
					return nil
				}
				return wsMsg{Msg: msg}
			}
		}()
	}

	// Return any command from list updates or continue listening
	if cmd != nil {
		return m, cmd
	}
	return m, nil
}
