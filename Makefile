build:
	go build -o server.exe ./cmd/server
	go build -o cli_planning_poker.exe ./cmd/client

clean:
	del server.exe cli_planning_poker.exe

.PHONY: build clean