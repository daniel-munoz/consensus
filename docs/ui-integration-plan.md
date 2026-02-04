# UI Integration Plan for Consensus CLI

## Decision Summary

| Choice | Selected |
|--------|----------|
| **UI Prototype** | Web Simple |
| **Integration** | HTTP API Server |
| **Response Updates** | Polling |
| **Auto-open Browser** | Yes |

---

## Phase 1: Add HTTP Server to Go CLI

**New Files:**
- `server.go` - HTTP server, handlers, session management

**Modify:**
- `cli.go` - Add `--serve` and `--port` flags
- `main.go` - Check for serve mode, start server instead of CLI flow

**API Endpoints:**
```
POST /api/consensus     - Submit prompt, returns session ID immediately
GET  /api/session/{id}  - Poll for status and responses
GET  /api/providers     - List available providers
GET  /api/health        - Health check
```

**Data Structures:**
```go
type ConsensusRequest struct {
    Prompt            string   `json:"prompt"`
    UseMasterPrompt   bool     `json:"useMasterPrompt"`
    MasterProvider    string   `json:"masterProvider"`
    ResponseProviders []string `json:"responseProviders"`
    EmailRecipients   []string `json:"emailRecipients"`
}

type SessionResponse struct {
    SessionID    string                     `json:"sessionId"`
    Status       string                     `json:"status"` // pending, processing, complete, error
    MasterPrompt *ProviderResponse          `json:"masterPrompt,omitempty"`
    Responses    map[string]ProviderResponse `json:"responses"`
}

type ProviderResponse struct {
    Status  string `json:"status"` // pending, complete, error
    Content string `json:"content,omitempty"`
    Error   string `json:"error,omitempty"`
}
```

**Session Management:**
- In-memory map of session ID → session state
- Background goroutines process requests (reuse existing `runProvidersConcurrently` logic)
- Sessions expire after 30 minutes

**Auto-open Browser:**
- Use `os/exec` to run `open` (macOS), `xdg-open` (Linux), or `start` (Windows)
- Serve static files from `ui-prototypes/web-simple/` at root path

---

## Phase 2: Update Web Simple UI

**Modify:** `ui-prototypes/web-simple/script.js`

**Changes:**
1. Replace `sendRequest()` mock with real fetch calls
2. Add polling loop for session status
3. Update `API_BASE_URL` constant

```javascript
const API_BASE_URL = 'http://localhost:8080';

async sendRequest() {
    // 1. Submit request
    const res = await fetch(`${API_BASE_URL}/api/consensus`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(this.collectFormData())
    });
    const { sessionId } = await res.json();

    // 2. Poll for results
    this.pollSession(sessionId);
}

async pollSession(sessionId) {
    const poll = async () => {
        const res = await fetch(`${API_BASE_URL}/api/session/${sessionId}`);
        const session = await res.json();
        this.updateResults(session);
        if (session.status !== 'complete') {
            setTimeout(poll, 2000); // Poll every 2 seconds
        }
    };
    poll();
}
```

---

## Phase 3: Add CORS Support

Add middleware in `server.go`:
```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

---

## Existing Code to Reuse

| Component | Location | How to Reuse |
|-----------|----------|--------------|
| Provider interface | `ai/*.go` | Call `Send()` from session goroutines |
| Hub/DelayedResponse | `core/response/*.go` | Manage async responses per session |
| Config loading | `config.go` | Load providers for server mode |
| Output writers | `output/*.go` | Still save files + send emails |
| CLI flags | `cli.go` | Extend with `--serve`, `--port` |

---

## Usage After Implementation

```bash
# Start server (auto-opens browser with UI)
go run main.go --serve

# Start server on custom port
go run main.go --serve --port 3000

# CLI mode still works as before
go run main.go -p "Your prompt here"
go run main.go -p "prompt" -mpp anthropic -rp "openai,gemini"
```

---

## Verification Plan

1. `go run main.go --serve` → Server starts, browser opens
2. Submit prompt in UI → See "Processing..." state
3. Wait for responses → All provider responses appear
4. Check `responses/` directory → Files created as before
5. Test with email flag → Emails sent correctly
6. `go run main.go -p "test"` → CLI still works normally
7. Test error cases → Missing API keys show errors in UI
