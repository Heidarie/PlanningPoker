package client

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
	"github.com/heidarie/cli_planning_poker/internal/server"
)

const (
	stateLobby = iota
	stateHost
	stateVoting
	stateResults
)

type Model struct {
	roomCode     string
	name         string
	Conn         *websocket.Conn
	state        int
	votes        map[string]string
	list         list.Model
	participants []string
	hasVoted     bool
	isHost       bool
	voteProgress struct {
		current int
		total   int
	}
}

type estimateItem int

func (e estimateItem) Title() string       { return fmt.Sprintf("%d", int(e)) }
func (e estimateItem) Description() string { return fmt.Sprintf("Story point: %d", int(e)) }
func (e estimateItem) FilterValue() string { return fmt.Sprintf("%d", int(e)) }

func NewModel(code, name string, conn *websocket.Conn, isHost bool) Model {
	choices := []list.Item{
		estimateItem(1),
		estimateItem(2),
		estimateItem(3),
		estimateItem(5),
		estimateItem(8),
		estimateItem(13),
		estimateItem(21),
	}
	lst := list.New(choices, list.NewDefaultDelegate(), 20, 10)
	lst.Title = "Select your estimate"
	lst.SetShowStatusBar(false)
	lst.SetFilteringEnabled(false)
	lst.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Bold(true)

	state := stateLobby

	if isHost {
		state = stateHost
	}

	return Model{
		roomCode:     code,
		name:         name,
		Conn:         conn,
		state:        state,
		votes:        make(map[string]string),
		list:         lst,
		participants: make([]string, 0),
		hasVoted:     false,
		isHost:       isHost,
	}
}

func (m Model) Init() tea.Cmd {
	return listenForMessages(m.Conn)
}

func listenForMessages(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		var msg server.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			return nil
		}
		return wsMsg{Msg: msg}
	}
}
