# Sickchat

A lightweight TCP-based chat system written in Go, featuring a multiclient server and terminal-based client.

##  Getting Started

### Build and Run

#### Server

```bash
cd server
go run main.go
````

#### Client

```bash
cd terminalClient
go run main.go
```

## Future Improvements

* Nicknames and direct messages
* Basic command support (`/quit`, `/list`, etc.)
* Persistent message history
* WebSocket front end

## License

This project is licensed under the [GPL-3.0 License](LICENSE).
