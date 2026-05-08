# PipesHub Go SDK — Conversation & Search Sample

End-to-end Go sample for the PipesHub SDK. Authenticates once, then walks
through the full conversation surface (user + agent), the conversation
lifecycle (archive / unarchive / delete), KB-filtered variants of both, and
semantic search — with and without a KB filter.

The demo code is split into two domain packages — `conversation/` and
`search/` — and a single `main.go` imports both. The SDK itself lives
locally under `pipeshub/` so the sample is self-contained; once the
upstream `github.com/pipeshub-ai/pipeshub-sdk-go` exposes the same surface
(notably the KB-filter methods), swap the imports and delete the local
copy.

## What this demonstrates

1. **Authenticate** — two-step initAuth → authenticate flow, returns a JWT.
2. **User conversation lifecycle** — stream a new conversation, send a
   follow-up, then archive / unarchive / delete with count checks.
3. **Agent conversation lifecycle** — same surface against
   `/agents/{agentKey}/conversations/*`.
4. **KB-filtered conversations (user + agent)** — same as above, but every
   call carries `searchFilters: { kb: [kbID] }` so retrieval is scoped to
   one knowledge base.
5. **Semantic search with KB filter** — `POST /search` scoped to a KB.
6. **Unscoped semantic search** — `POST /search` across the whole workspace.

## Prerequisites

- Go 1.21 or newer
- A running PipesHub backend (default: `http://localhost:3000`)
- A PipesHub user account
- (Optional) an agent key to exercise the agent scenarios
- (Optional) a knowledge base ID to exercise the KB-filtered scenarios

## Setup

```bash
cp .env.example .env
# edit .env and fill in your credentials
set -a; source .env; set +a
```

The sample reads env vars directly via `os.Getenv` — no `.env` loader is
imported. `set -a; source .env; set +a` exports every line in the file to
the current shell.

## Run

```bash
go run .
```

The output is sectioned with colored headers. Each section either prints
a streamed answer, a JSON response, or a verification line. Scenarios are
independent — if one fails, a red message is printed and the next scenario
still runs.

## Project layout

```
conversation_and_search/
├── pipeshub/              # Local copy of the Go SDK (swap to upstream once published).
│   └── client.go
├── conversation/          # Conversation scenarios — user + agent + KB-filtered.
│   └── conversation.go
├── search/                # Semantic search scenarios — plain + KB-filtered.
│   └── search.go
├── internal/ui/           # Console helpers: colors, sections, SSE parser.
│   └── ui.go
├── main.go                # Auth + orchestration. Imports conversation and search.
├── go.mod
├── .env.example
└── README.md
```

## Adding your own scenario

Each scenario is a plain function that takes a `*pipeshub.Client` and
returns an `error`. To add one, drop a function in either
`conversation/conversation.go` or `search/search.go`, then call it from
`main.go`:

```go
func RunMyThing(client *pipeshub.Client, query string) error {
    ui.Section("My Thing")
    // ... use client.<SDKMethod>(...) ...
    return nil
}
```

The presentation helpers (`ui.Section`, `ui.PrintSSE`, `ui.PrintJSON`,
`ui.Success`, `ui.Errorf`, `ui.Warn`, `ui.Info`) live in
`internal/ui/ui.go`. Use them so your section blends in with the rest of
the demo.

## Switching to the published SDK

Once `github.com/pipeshub-ai/pipeshub-sdk-go` exposes the same surface
(particularly the `*WithFilters` methods and `SearchFilters` type),
replace every:

```go
import "conversation_and_search/pipeshub"
```

with:

```go
import pipeshub "github.com/pipeshub-ai/pipeshub-sdk-go"
```

Add the dependency to `go.mod` (`go get github.com/pipeshub-ai/pipeshub-sdk-go`)
and delete the local `pipeshub/` directory.
