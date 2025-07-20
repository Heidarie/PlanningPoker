package client

import (
	"fmt"
)

// View renders UI based on current state
func (m Model) View() string {
	switch m.state {
	case stateLobby:
		return fmt.Sprintf("Room %s - waiting for host to start voting...", m.roomCode)

	case stateHost:
		out := fmt.Sprintf("Room %s created. Waiting for others to join.\n\n", m.roomCode)
		out += "Participants:\n"
		if len(m.participants) == 0 {
			out += "  (none yet)\n"
		} else {
			for _, participant := range m.participants {
				out += fmt.Sprintf("  • %s\n", participant)
			}
		}

		// Check if voting is in progress by looking at vote progress
		if m.voteProgress.total > 0 {
			// Show voting progress
			out += fmt.Sprintf("\nVoting in progress... (%d/%d votes received)", m.voteProgress.current, m.voteProgress.total)
			out += "\nWaiting for all participants to vote..."
		} else {
			out += "\nPress Enter to start voting when ready."
		}
		return out

	case stateVoting:
		// show list of cards for user to choose
		if m.hasVoted {
			return m.list.View() + "\n\n✓ Vote submitted and locked! Waiting for other players..."
		} else {
			return m.list.View() + "\n\nUse ↑/↓ to navigate, Enter to select, q/Ctrl+C to quit"
		}

	case stateResults:
		// build results table
		out := "Votes:\n"
		for user, val := range m.votes {
			out += fmt.Sprintf("%s: %s\n", user, val)
		}
		out += "\nPress Enter to reset or quit."
		return out
	}
	return ""
}
