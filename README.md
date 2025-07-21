# Go For PlanningPoker
🃏 PlanningPoker is a collaborative tool for agile teams to estimate tasks using the Planning Poker technique. It consists of a server and a terminal-based client UI written in Go.

## Features
- Create or Join Rooms: Host can create a new room; participants can join with a code.
- Real-time Voting: Participants select their estimates; votes are revealed once everyone has voted.
- Live Participant List: See who is in the room and who has voted.
- Room Cleanup: Inactive/empty rooms are cleaned up automatically.
- Terminal UI: Client uses a modern TUI for smooth navigation.
## Getting Started
### Prerequisites
- Go 1.22 or above
- Windows, Linux, or macOS
- Git
### Build

```
git clone https://github.com/Heidarie/PlanningPoker.git
cd PlanningPoker
make build
```
This will generate server.exe and cli_planning_poker.exe in the root directory.

Alternatively run
```
go build /src/server -o /<your_name>.exe
go build /src/client -o /<your_name>.exe
```
### Run the Server
```
./server.exe
```
- The server listens on http://localhost:8080.

### Run the Client
```
./cli_planning_poker.exe
```
- Navigate the menu with ↑/↓, Enter to select, and q/Ctrl+C to quit.
- Create a new room or join an existing room by entering the room code and your name.
## Usage flow
1. Connect to the server on a host machine.
2. Clients connect via the terminal application.
3. Host creates a room and shares the code.
4. Participants join using the code and enter their name.
5. Host starts the voting round.
6. Each participant selects an estimate.
7. Results are displayed after all votes are in.

## Download

You can always find the latest Windows executable here:

[![Download](https://img.shields.io/github/v/release/Heidarie/PlanningPoker)](https://github.com/Heidarie/PlanningPoker/releases/latest)

- [Download PlanningPoker.exe](https://github.com/Heidarie/PlanningPoker/releases/latest/download/PlanningPoker.exe)

Automated Releases
The repository uses GitHub Actions to automatically build and publish .exe files for each release. Tagging a commit (e.g., v1.0.0) will trigger the workflow, build the binaries, and attach them to the release.

### License
No license specified yet.

### Contributing
Pull requests and suggestions are welcome!
