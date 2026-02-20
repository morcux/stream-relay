# Stream Relay

Stream Relay is a simple WebRTC Selective Forwarding Unit (SFU) built with Go and the [Pion WebRTC](https://github.com/pion/webrtc) library. It acts as a central server that receives media streams (video/audio) from peers and relays them to other connected participants in a room. 

## Features

- **WebRTC SFU Architecture**: Efficiently routes video streams between peers without requiring a full mesh topology.
- **WebSocket Signaling**: Handles SDP negotiation and message exchange directly over WebSockets.
- **Simple Web Frontend**: Includes a Vanilla JavaScript frontend client to test media streaming directly from the browser.
- **Concurrent Connections**: Leverages Go's concurrency model ensuring stable multi-client routing with internal locking mechanisms.

## Project Structure

- `cmd/server/main.go`: Main application entrypoint setting up the HTTP and signaling server.
- `internal/sfu/`: Core WebRTC logic managing rooms, routing RTP packets, tracks, and remote peers.
- `internal/signal/`: Signaling protocol implementation handling HTTP endpoints and WebSockets for WebRTC handshakes.
- `web/public/`: Static frontend web application for testing WebRTC publishing and subscribing interactions.

## Prerequisites

- **Go**: 1.25.5 or higher.
- A modern web browser that supports WebRTC (Chrome, Firefox, Safari, Edge).

## Getting Started

### 1. Install Dependencies

Ensure that Go modules are downloaded:

```bash
cd stream-relay
go mod download
```

### 2. Run the Server

Start the WebRTC SFU server. It will listen on port `8080` by default.

```bash
go run cmd/server/main.go
```

### 3. Usage

1. Open `http://localhost:8080` in your web browser.
2. Grant camera/microphone permissions when prompted.
3. Click the start button (per `app.js`) to initiate the WebRTC connection.
4. Open the same URL in another tab or a different browser.
5. The local video stream should appear on your screen, and the remote participant's stream will be relayed and playback dynamically.

## License

Check the `LICENSE` file for details.