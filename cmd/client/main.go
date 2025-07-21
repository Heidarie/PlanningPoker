package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
	"github.com/heidarie/cli_planning_poker/internal/client"
)

// Build-time variables (injected via -ldflags)
var (
	BuildServerURL    string
	BuildClientSecret string
	BuildVersion      = "dev"
	BuildCommit       = "unknown"
	BuildDate         = "unknown"
)

// Menu states
const (
	menuMain = iota
	menuJoinCode
	menuJoinName
	menuConnecting
)

// Menu item for the main menu
type menuItem string

func (e menuItem) Title() string       { return fmt.Sprintf("%v", e) }
func (e menuItem) Description() string { return "" }
func (e menuItem) FilterValue() string { return fmt.Sprintf("%v", e) }

// Menu model
type menuModel struct {
	state      int
	list       list.Model
	textInput  textinput.Model
	roomCode   string
	playerName string
	isHost     bool
	mode       string
	serverAddr string
	err        error
}

// Messages
type connectionSuccessMsg struct {
	conn     *websocket.Conn
	roomCode string
	name     string
	isHost   bool
}

type connectionErrorMsg struct {
	err error
}

type roomCreatedMsg struct {
	roomCode string
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))
)

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case menuMain:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				// Clear any previous errors when making a selection
				m.err = nil
				selected := m.list.SelectedItem()
				if selected != nil {
					switch selected.(menuItem) {
					case "Create a new room (be the host)":
						return m, createRoom(m.serverAddr)
					case "Join an existing room":
						m.state = menuJoinCode
						m.textInput.Focus()
						m.textInput.Placeholder = "Enter room code..."
					}
				}
			}
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case menuJoinCode:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.roomCode = strings.TrimSpace(m.textInput.Value())
				if m.roomCode != "" {
					m.state = menuJoinName
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Enter your name..."
				}
			case "esc":
				m.state = menuMain
				m.textInput.Blur()
				m.textInput.SetValue("")
			}
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd

		case menuJoinName:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.playerName = strings.TrimSpace(m.textInput.Value())
				if m.playerName != "" {
					m.state = menuConnecting
					m.mode = "join"
					m.isHost = false
					return m, connectToRoom(m.serverAddr, m.roomCode, m.playerName, m.mode)
				}
			case "esc":
				m.state = menuJoinCode
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Enter room code..."
			}
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

	case roomCreatedMsg:
		m.roomCode = msg.roomCode
		m.mode = "create"
		m.isHost = true
		m.state = menuConnecting
		return m, connectToRoom(m.serverAddr, m.roomCode, "", m.mode)

	case connectionSuccessMsg:
		// Switch to the main game TUI
		gameModel := client.NewModel(msg.roomCode, msg.name, msg.conn, msg.isHost)
		return gameModel, gameModel.Init()

	case connectionErrorMsg:
		m.err = msg.err
		m.state = menuMain
		// Reset the text input when going back to main menu
		m.textInput.SetValue("")
		m.textInput.Blur()
		return m, nil
	}

	return m, nil
}

func (m menuModel) View() string {
	switch m.state {
	case menuMain:
		title := titleStyle.Render("🃏 Planning Poker")

		content := ""
		if m.err != nil {
			content = errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n"
		}

		content += m.list.View()
		content += "\n\n" + lipgloss.NewStyle().Faint(true).Render("Use ↑/↓ to navigate, Enter to select, q/Ctrl+C to quit")

		return fmt.Sprintf("%s\n\n%s", title, content)

	case menuJoinCode:
		title := titleStyle.Render("🃏 Join Room")
		content := "Enter the room code:\n\n"
		content += m.textInput.View()
		content += "\n\n" + lipgloss.NewStyle().Faint(true).Render("Press Enter to continue, Esc to go back")
		return fmt.Sprintf("%s\n\n%s", title, content)

	case menuJoinName:
		title := titleStyle.Render("🃏 Join Room")
		content := fmt.Sprintf("Room: %s\n\n", infoStyle.Render(m.roomCode))
		content += "Enter your name:\n\n"
		content += m.textInput.View()
		content += "\n\n" + lipgloss.NewStyle().Faint(true).Render("Press Enter to join, Esc to go back")
		return fmt.Sprintf("%s\n\n%s", title, content)

	case menuConnecting:
		title := titleStyle.Render("🃏 Planning Poker")
		var content string
		if m.isHost {
			content = fmt.Sprintf("Creating room %s...", infoStyle.Render(m.roomCode))
		} else {
			content = fmt.Sprintf("Joining room %s as %s...", infoStyle.Render(m.roomCode), infoStyle.Render(m.playerName))
		}
		return fmt.Sprintf("%s\n\n%s", title, content)
	}

	return ""
}

// Commands
func createRoom(serverAddr string) tea.Cmd {
	return func() tea.Msg {
		httpClient := &http.Client{}
		req, err := http.NewRequest("POST", serverAddr+"/create", nil)
		if err != nil {
			return connectionErrorMsg{err}
		}

		// Add authentication header
		req.Header.Set("X-Client-Secret", client.CLIENT_SECRET)

		resp, err := httpClient.Do(req)
		if err != nil {
			return connectionErrorMsg{err}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return connectionErrorMsg{fmt.Errorf("server error: %s", string(bodyBytes))}
		}

		body, _ := io.ReadAll(resp.Body)
		return roomCreatedMsg{roomCode: string(body)}
	}
}

func connectToRoom(serverAddr, roomCode, playerName, mode string) tea.Cmd {
	return func() tea.Msg {
		// Convert HTTP(S) to WS(S)
		wsURL := strings.Replace(serverAddr, "https://", "wss://", 1)
		wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
		wsURL += "/ws?code=" + url.QueryEscape(roomCode) + "&name=" + url.QueryEscape(playerName) + "&mode=" + url.QueryEscape(mode)

		// WebSocket doesn't support custom headers during handshake easily with gorilla/websocket
		// So we need to use a custom dialer with headers
		headers := http.Header{}
		headers.Set("X-Client-Secret", client.CLIENT_SECRET)

		wsConn, response, err := websocket.DefaultDialer.Dial(wsURL, headers)
		if err != nil {
			if response != nil && response.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(response.Body)
				return connectionErrorMsg{fmt.Errorf("%s", string(bodyBytes))}
			}
			return connectionErrorMsg{err}
		}

		return connectionSuccessMsg{
			conn:     wsConn,
			roomCode: roomCode,
			name:     playerName,
			isHost:   mode == "create",
		}
	}
}

func main() {
	// Set build-time configuration if available
	if BuildServerURL != "" || BuildClientSecret != "" {
		client.SetBuildTimeConfig(BuildServerURL, BuildClientSecret)
	}

	// Show version info in dev mode
	if client.DEV_MODE {
		fmt.Printf("Planning Poker Client %s\n", BuildVersion)
		fmt.Printf("Build: %s (%s)\n", BuildCommit, BuildDate)
		fmt.Printf("Server: %s\n\n", client.SERVER_URL)
	}

	// Validate configuration
	if err := client.ValidateConfig(); err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		fmt.Println("Please set CLIENT_SECRET environment variable or use a pre-built binary.")
		os.Exit(1)
	}

	// Create menu items
	items := []list.Item{
		menuItem("Create a new room (be the host)"),
		menuItem("Join an existing room"),
	}

	// Create list
	l := list.New(items, list.NewDefaultDelegate(), 50, 10)
	l.Title = "What would you like to do?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Bold(true)

	// Create text input
	ti := textinput.New()
	ti.CharLimit = 50
	ti.Width = 20

	// Create initial model
	m := menuModel{
		state:      menuMain,
		list:       l,
		textInput:  ti,
		serverAddr: client.SERVER_URL,
	}

	// Create program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Handle Ctrl+C for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		p.Quit()
	}()

	// Start the program
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
