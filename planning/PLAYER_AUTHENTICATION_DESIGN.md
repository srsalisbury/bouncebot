# Player Authentication Design

## Problem Statement

Currently, `player_id` serves as both the public identifier and the authentication credential. Since player IDs are broadcast to all room members (via `GetRoom` responses and WebSocket events), any user can impersonate any other user by simply using their player_id in API requests.

### Current Vulnerabilities

1. **Direct impersonation**: Anyone in a room can extract another player's ID and submit actions as them
2. **Room enumeration**: Room IDs are 4 characters (32^4 = ~1M possibilities), easily brute-forced to discover active rooms
3. **Passive snooping**: `GetRoom` can be called by anyone, exposing all player IDs without joining
4. **Host takeover**: Room creator's ID is public, allowing attackers to boot players or change settings

### Threat Model

Protect against:
- Casual attackers (mischievous players in the same room)
- Determined attackers (scripts enumerating rooms and extracting player IDs)
- Application-level attacks (not network-level MITM, since HTTPS is not guaranteed)

### Constraints

- User should never see or handle secrets manually
- No cookies (localStorage only)
- Reconnection from same browser must continue working
- No need for cross-device session transfer
- Per-game-session protection is sufficient
- Design can be replaced later when adding a login system

---

## Proposed Solution: Session Tokens

Introduce a secret `session_token` that is:
- Generated server-side alongside `player_id`
- Returned **only** to the owning player (never broadcast)
- Required for all authenticated API requests
- Stored in localStorage on the client

### Key Principle

Separate **identity** (public, for display) from **authentication** (secret, for actions):

| Field | Visibility | Purpose |
|-------|------------|---------|
| `player_id` | Public (broadcast to room) | Identify who did what, leaderboards, display |
| `session_token` | Private (only owner sees) | Prove you are who you claim to be |

---

## Implementation Details

### 1. Server Changes

#### 1.1 Player struct (room/room.go)

```go
type Player struct {
    ID           string  // Public identifier (existing)
    SessionToken string  // Secret authentication token (new)
    Name         string
    ColorIndex   int
    // ...
}
```

#### 1.2 Token generation (room/repository.go)

```go
func generateSessionToken() string {
    // Use crypto/rand for security-sensitive tokens
    b := make([]byte, 32)
    if _, err := crypto_rand.Read(b); err != nil {
        panic(err) // Should never happen
    }
    return base64.URLEncoding.EncodeToString(b)
}
```

#### 1.3 CreateRoom / JoinRoom responses

Return `session_token` only in the response to the creating/joining player:

```proto
message CreateRoomResponse {
    string room_id = 1;
    string player_id = 2;
    string session_token = 3;  // NEW - only returned to room creator
    Room room = 4;
}

message JoinRoomResponse {
    string player_id = 1;
    string session_token = 2;  // NEW - only returned to joining player
    Room room = 3;
}
```

#### 1.4 Room/Player messages (broadcast to everyone)

**Do NOT include session_token** in the `Player` message that's broadcast:

```proto
message Player {
    string id = 1;
    string name = 2;
    int32 color_index = 3;
    // NO session_token here - it's secret
}
```

#### 1.5 Authenticated requests

All action requests should use `session_token` instead of `player_id` for auth:

```proto
message SubmitSolutionRequest {
    string room_id = 1;
    string session_token = 2;  // CHANGED from player_id
    repeated BotPos moves = 3;
}

message MarkFinishedSolvingRequest {
    string room_id = 1;
    string session_token = 2;  // CHANGED from player_id
}

message MarkReadyForNextRequest {
    string room_id = 1;
    string session_token = 2;  // CHANGED from player_id
}

message UpdateRoomSettingsRequest {
    string room_id = 1;
    string session_token = 2;  // CHANGED from player_id
    // ...settings fields
}

message BootPlayerRequest {
    string room_id = 1;
    string session_token = 2;      // CHANGED from host_player_id
    string target_player_id = 3;   // Keep as player_id (public identifier)
}
```

#### 1.6 Server-side validation

Add a method to look up player by session token:

```go
func (r *Room) GetPlayerBySessionToken(token string) *Player {
    for i := range r.Players {
        if r.Players[i].SessionToken == token {
            return &r.Players[i]
        }
    }
    return nil
}
```

Each handler validates the token before processing:

```go
func (s *Server) SubmitSolution(ctx context.Context, req *pb.SubmitSolutionRequest) (*pb.SubmitSolutionResponse, error) {
    room := s.rooms.GetRoom(req.RoomId)
    if room == nil {
        return nil, errors.New("room not found")
    }

    player := room.GetPlayerBySessionToken(req.SessionToken)
    if player == nil {
        return nil, errors.New("invalid session token")
    }

    // Continue with player.ID for the actual logic...
}
```

#### 1.7 WebSocket authentication

Update WebSocket connection to use session_token:

```go
// Current: /ws?roomId=ABC5&playerId=xxx
// New:     /ws?roomId=ABC5&sessionToken=xxx

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    roomID := r.URL.Query().Get("roomId")
    sessionToken := r.URL.Query().Get("sessionToken")

    room := h.rooms.GetRoom(roomID)
    if room == nil {
        http.Error(w, "room not found", http.StatusNotFound)
        return
    }

    player := room.GetPlayerBySessionToken(sessionToken)
    if player == nil {
        http.Error(w, "invalid session token", http.StatusUnauthorized)
        return
    }

    // Continue with player.ID for connection tracking...
}
```

### 2. Client Changes

#### 2.1 Storage (roomStore.ts)

```typescript
const STORAGE_KEY_PLAYER_ID = 'bouncebot_player_id'
const STORAGE_KEY_SESSION_TOKEN = 'bouncebot_session_token'  // NEW

// Store both on join/create
function storeSession(playerId: string, sessionToken: string) {
    localStorage.setItem(STORAGE_KEY_PLAYER_ID, playerId)
    localStorage.setItem(STORAGE_KEY_SESSION_TOKEN, sessionToken)
}

// Use session token for API calls
function getSessionToken(): string | null {
    return localStorage.getItem(STORAGE_KEY_SESSION_TOKEN)
}
```

#### 2.2 API calls (useGameActions.ts)

```typescript
async function submitSolution() {
    const sessionToken = roomStore.getSessionToken()
    if (!sessionToken) return

    await bounceBotClient.submitSolution({
        roomId: roomId.value,
        sessionToken: sessionToken,  // Use token instead of playerId
        moves,
    })
}
```

#### 2.3 WebSocket connection (websocket.ts)

```typescript
private doConnect(): void {
    if (!this.roomId || !this.sessionToken) return
    const url = `${config.wsUrl}?roomId=${this.roomId}&sessionToken=${this.sessionToken}`
    this.ws = new WebSocket(url)
}
```

### 3. Preventing Room Enumeration (Optional Enhancement)

To prevent attackers from discovering rooms by brute force:

#### Option A: Require session token for GetRoom

Only allow `GetRoom` calls from authenticated players:

```proto
message GetRoomRequest {
    string room_id = 1;
    string session_token = 2;  // Required - must be a player in this room
}
```

**Downside**: Breaks the "join by room code" flow where you check if a room exists before joining.

#### Option B: Rate limiting on GetRoom

Add rate limiting (e.g., 10 requests/minute per IP) to make enumeration impractical:

```go
// Using a simple token bucket or sliding window
if !rateLimiter.Allow(clientIP, "GetRoom", 10, time.Minute) {
    return nil, status.Error(codes.ResourceExhausted, "rate limited")
}
```

#### Option C: Longer room codes

Increase room ID length from 4 to 6 characters:
- Current: 32^4 = 1,048,576 possibilities
- Proposed: 32^6 = 1,073,741,824 possibilities (~1000x harder to enumerate)

**Recommendation**: Implement Option B (rate limiting) as a quick win, consider Option C for new rooms.

---

## Migration Strategy

### Phase 1: Add session tokens (backward compatible)

1. Generate `session_token` for all new players
2. Return `session_token` in Create/Join responses
3. Accept **both** `player_id` and `session_token` in requests (for backward compat)
4. Update client to send `session_token` when available
5. Log warnings when `player_id` is used without `session_token`

### Phase 2: Require session tokens

1. After client is deployed and users have refreshed:
2. Require `session_token` for all authenticated requests
3. Reject requests using only `player_id`

### Phase 3: Stop exposing player_id in requests (cleanup)

1. Remove `player_id` field from authenticated request messages
2. Server derives `player_id` from `session_token` internally

---

## Security Analysis

### What this protects against

| Attack | Protected? | Notes |
|--------|------------|-------|
| Player in room impersonating another | Yes | Session token is secret |
| External attacker calling GetRoom | Yes | Can see player_ids but not tokens |
| Brute-force room discovery | Partial | Rate limiting helps; longer codes better |
| Host privilege escalation | Yes | Need host's session token |
| Replay attacks | No | Same token works until session ends |
| Network-level MITM | No | Requires HTTPS |

### What this does NOT protect against

1. **Network attackers** (without HTTPS): Anyone who can see network traffic can intercept the session token
2. **XSS attacks**: Malicious JavaScript can still read localStorage
3. **Compromised client**: If the user's browser/device is compromised, token can be stolen
4. **Server-side token leakage**: If rooms.json is exposed, tokens are compromised

### Acceptable limitations

Given the game's casual nature and the constraint that HTTPS is not guaranteed, these limitations are acceptable trade-offs.

---

## Alternative Approaches Considered

### Alternative 1: WebSocket-only actions

Move all game actions through WebSocket instead of HTTP. Server tracks which connection owns which player.

**Pros:**
- Elegant - no repeated auth needed
- Connection itself is the credential

**Cons:**
- Major refactor of action handling
- Lose HTTP semantics (request/response)
- Harder to debug

**Verdict**: Good long-term architecture, but session tokens are simpler to implement now.

### Alternative 2: HMAC signatures

Client signs each request with HMAC(secret, request_data + timestamp).

**Pros:**
- Secret never transmitted after initial setup
- Replay attacks limited by timestamp

**Cons:**
- Complex client-side crypto
- Clock synchronization issues
- Still vulnerable to MITM on initial exchange

**Verdict**: Over-engineered for this use case.

### Alternative 3: IP binding

Bind session to client IP address.

**Pros:**
- Simple

**Cons:**
- Many users share IPs (NAT, corporate networks)
- Doesn't protect against same-network attackers
- IPs change (mobile networks)

**Verdict**: Too unreliable.

---

## Files to Modify

### Server (Go)

| File | Changes |
|------|---------|
| `proto/bouncebot.proto` | Add session_token to messages |
| `server/room/room.go` | Add SessionToken to Player struct |
| `server/room/repository.go` | Generate session tokens, add lookup method |
| `server/bouncebotserver/server.go` | Validate session tokens in handlers |
| `server/websocket/hub.go` | Auth WebSocket with session token |

### Client (TypeScript/Vue)

| File | Changes |
|------|---------|
| `client/vue1/src/stores/roomStore.ts` | Store/retrieve session token |
| `client/vue1/src/composables/useGameActions.ts` | Send session token in requests |
| `client/vue1/src/services/websocket.ts` | Connect with session token |
| `client/vue1/src/gen/bouncebot_pb.ts` | Regenerated from proto |

---

## Open Questions

1. **Existing rooms**: Should we invalidate all existing sessions when deploying? (Forces everyone to rejoin)
2. **Token rotation**: Should tokens rotate periodically or on reconnect?
3. **Audit logging**: Should we log failed authentication attempts for monitoring?
