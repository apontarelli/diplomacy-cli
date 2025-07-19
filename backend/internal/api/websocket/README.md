# WebSocket Real-Time Updates

This package provides real-time WebSocket functionality for the Diplomacy game engine, enabling live updates for game state changes, player actions, and resolution events.

## Features

- **Real-time game updates**: Broadcast game state changes to all connected clients
- **Per-game isolation**: Clients are organized by game ID for efficient message routing
- **User-specific notifications**: Send targeted messages to specific users
- **JWT Authentication**: Secure WebSocket connections using existing JWT tokens
- **Event-driven architecture**: Structured event types for different game actions
- **Connection management**: Automatic cleanup of disconnected clients

## Architecture

### Components

1. **Hub**: Central message broker that manages client connections and broadcasts
2. **Handler**: WebSocket connection handler with authentication
3. **Client**: Represents an individual WebSocket connection
4. **Events**: Structured event types for different game actions

### Connection Flow

1. Client connects to `/ws/game/{gameID}` with JWT token
2. Server authenticates the token and upgrades to WebSocket
3. Client is registered in the game-specific hub
4. Server broadcasts relevant events to connected clients
5. Client disconnection triggers automatic cleanup

## Usage

### Client Connection

Connect to a game's WebSocket endpoint:

```javascript
// Using Authorization header (preferred)
const ws = new WebSocket('ws://localhost:8080/ws/game/123', [], {
  headers: {
    'Authorization': 'Bearer your_jwt_token'
  }
});

// Using query parameter (fallback)
const ws = new WebSocket('ws://localhost:8080/ws/game/123?token=your_jwt_token');
```

### Message Format

All WebSocket messages follow this JSON structure:

```json
{
  "type": "event_type",
  "data": {
    // Event-specific data
  }
}
```

### Event Types

#### Game Events
- `game_state_update`: Game state has changed
- `game_phase_change`: Game phase transition
- `game_started`: Game has started
- `game_ended`: Game has ended

#### Player Events
- `player_joined`: A player joined the game
- `player_left`: A player left the game

#### Order Events
- `orders_submitted`: A player submitted orders
- `orders_resolved`: Orders have been resolved
- `order_deadline`: Order deadline notification

#### Resolution Events
- `resolution_started`: Resolution process started
- `resolution_completed`: Resolution completed with results
- `resolution_error`: Resolution failed

#### System Events
- `connected`: Client successfully connected
- `disconnected`: Client disconnected
- `error`: Error occurred
- `ping`/`pong`: Heartbeat messages

### Example Events

#### Player Joined Event
```json
{
  "type": "player_joined",
  "data": {
    "game_id": 123,
    "player_id": 456,
    "user_id": 789,
    "username": "alice",
    "nation": "england",
    "event_time": "2023-07-19T10:30:00Z"
  }
}
```

#### Orders Submitted Event
```json
{
  "type": "orders_submitted",
  "data": {
    "game_id": 123,
    "player_id": 456,
    "user_id": 789,
    "order_count": 3,
    "event_time": "2023-07-19T10:35:00Z",
    "orders": [
      {
        "id": 1,
        "unit_province": "london",
        "type": "move",
        "target": "wales"
      }
    ]
  }
}
```

#### Resolution Completed Event
```json
{
  "type": "resolution_completed",
  "data": {
    "game_id": 123,
    "phase": "spring_movement",
    "year": 1901,
    "event_time": "2023-07-19T10:40:00Z",
    "results": {
      // Resolution results from the game engine
    }
  }
}
```

## Server-Side Broadcasting

### Broadcasting to All Players in a Game

```go
// Broadcast game state update to all players
gameUpdate := websocket.NewGameStateUpdate(gameID, "spring", 1901, "active", gameData)
wsHandler.BroadcastGameUpdate(gameID, websocket.EventGameStateUpdate, gameUpdate)
```

### Broadcasting to a Specific User

```go
// Send notification to a specific user
notification := websocket.NewErrorEvent("INVALID_ORDER", "Your order was invalid", orderContext)
wsHandler.BroadcastUserNotification(gameID, userID, websocket.EventError, notification)
```

### Integration with Game Handlers

The WebSocket system is integrated into the existing HTTP handlers:

```go
// In game handler - player joins
playerEvent := websocket.NewPlayerJoinedEvent(gameID, playerID, userID, username, nation)
gh.wsHandler.BroadcastGameUpdate(gameID, websocket.EventPlayerJoined, playerEvent)

// In order handler - orders submitted
orderEvent := websocket.NewOrdersSubmittedEvent(gameID, playerID, userID, orderCount, orders)
oh.wsHandler.BroadcastGameUpdate(gameID, websocket.EventOrdersSubmitted, orderEvent)
```

## Security

- **JWT Authentication**: All connections must provide a valid JWT token
- **Game Isolation**: Players can only connect to games they're authorized for
- **Rate Limiting**: Connection upgrades respect existing rate limits
- **Origin Checking**: Configure `CheckOrigin` for production deployments

## Testing

Run the WebSocket tests:

```bash
go test ./internal/api/websocket/...
```

The test suite covers:
- Hub broadcasting functionality
- Game-specific message routing
- User-specific message targeting
- Connection management
- Event helper functions

## Performance Considerations

- **Connection Limits**: Monitor concurrent connections per game
- **Message Queuing**: Client send channels are buffered (256 messages)
- **Automatic Cleanup**: Disconnected clients are automatically removed
- **Heartbeat**: Ping/pong messages maintain connection health

## Future Enhancements

- **Message Persistence**: Store messages for offline players
- **Subscription Filtering**: Allow clients to subscribe to specific event types
- **Compression**: Enable WebSocket compression for large messages
- **Clustering**: Support for multiple server instances with Redis pub/sub