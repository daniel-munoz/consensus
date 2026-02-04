# Decision Memo: UI Integration for Consensus CLI

**Date:** February 2026
**Status:** Approved
**Author:** Claude (with Daniel)

---

## Context

The Consensus CLI tool sends prompts to multiple AI providers (OpenAI, Anthropic, Gemini) concurrently and saves responses for comparison. Three UI prototypes were created to provide a visual alternative to the CLI:

1. **Web Simple** - Static HTML/CSS/JS
2. **React App** - React SPA with Tailwind CSS
3. **Electron App** - Desktop application

The goal is to add an optional UI while keeping the tool primarily CLI-based.

---

## Decision

**Selected: Web Simple + HTTP API Server**

---

## UI Prototype Analysis

### Web Simple
| Aspect | Assessment |
|--------|------------|
| Files | 3 files (~19KB total) |
| Dependencies | None |
| Build step | None required |
| Launch | `open index.html` |
| Modification effort | Low - vanilla JS |
| Mobile ready | Yes (responsive CSS) |

### React App
| Aspect | Assessment |
|--------|------------|
| Files | Full React project |
| Dependencies | React, Tailwind, Headless UI, Heroicons |
| Build step | `npm install && npm start` |
| Launch | Dev server on port 3000 |
| Modification effort | Medium - requires React knowledge |
| Mobile ready | Yes (Tailwind responsive) |

### Electron App
| Aspect | Assessment |
|--------|------------|
| Files | Electron project with main/renderer |
| Dependencies | Electron, electron-builder |
| Build step | `npm install && npm start` |
| Launch | Opens desktop window |
| Modification effort | High - IPC complexity |
| Mobile ready | No (desktop only) |

---

## Integration Approaches Considered

### Option A: HTTP API Server (Selected)
Add `--serve` flag that starts an HTTP server. UI calls REST endpoints.

**Pros:**
- CLI remains primary interface
- Works with any UI (not locked to one)
- Standard REST patterns
- Clean separation of concerns

**Cons:**
- Server runs separately from UI
- Requires polling for updates

### Option B: Electron Child Process
Electron spawns Go binary, communicates via stdin/stdout.

**Pros:**
- Single executable distribution
- No separate server

**Cons:**
- Only works with Electron
- Complex IPC and bundling
- Overkill for CLI-first philosophy

### Option C: JSON Output Mode
Add `--json` flag for machine-readable output.

**Pros:**
- Minimal Go changes

**Cons:**
- Only works with Electron
- No real-time updates
- Poor UX

---

## Response Update Mechanism

### Polling (Selected)
UI polls `/api/session/{id}` every 2 seconds.

**Pros:**
- Simple to implement
- No additional dependencies
- Works everywhere

**Cons:**
- Slight delay in updates
- More HTTP requests

### WebSocket Streaming
Real-time character-by-character updates.

**Pros:**
- Best UX
- Efficient for long responses

**Cons:**
- More complex implementation
- Requires WebSocket handling in Go and JS

**Decision:** Start with polling. Can add WebSocket later if needed.

---

## Why Web Simple Wins

1. **Zero build step** - Just open the HTML file
2. **Self-contained** - No npm, no node_modules
3. **Easy to modify** - Vanilla JS, no framework
4. **Embeddable** - Could later embed in Go binary using `embed`
5. **Aligns with CLI-first** - Simple, optional, not the main focus

---

## Existing Architecture Benefits

The current Go codebase has several abstractions that make UI integration straightforward:

| Component | Benefit for UI |
|-----------|----------------|
| `Provider` interface | Polymorphic - call any provider the same way |
| `Hub` abstraction | Already designed for async response consumption |
| `DelayedResponse` | Handles timeouts gracefully |
| `Writer` interface | Can add new output types easily |
| Config system | Provider management already solved |

The `Hub` interface (`core/response/hub.go`) is particularly well-suited - it abstracts away async complexity and provides clean methods for getting responses.

---

## Trade-offs Accepted

1. **Polling vs real-time** - Acceptable 2-second delay for simplicity
2. **In-memory sessions** - No persistence across restarts (acceptable for local tool)
3. **CORS wildcard** - Security acceptable for local development tool
4. **No authentication** - Local tool doesn't need it

---

## Future Considerations

- **WebSocket streaming** - Add if users want real-time character updates
- **Embed UI in binary** - Use Go 1.16+ `embed` to ship single executable
- **Session persistence** - Add if users want to resume sessions
- **Remote access** - Add auth if tool needs to be accessed remotely

---

## Summary

The combination of **Web Simple UI + HTTP API Server + Polling** provides:
- Minimal changes to existing CLI
- Optional UI that doesn't complicate the core tool
- Clean separation allowing future UI changes
- Reuse of existing Provider, Hub, and Writer abstractions
