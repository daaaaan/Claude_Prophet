# Stack Research

**Domain:** Real-time trading dashboard served from Go backend
**Researched:** 2026-02-11
**Confidence:** HIGH

## Recommended Stack

### Architecture Decision: Server-Rendered with htmx + Alpine.js

**Not a SPA.** This is a server-rendered dashboard using Go templates, htmx for dynamic updates, Alpine.js for client-side interactivity, and TradingView Lightweight Charts for financial charting. The Go backend renders HTML, pushes updates over WebSocket, and htmx swaps DOM fragments. No Node.js build step. No React/Vue/Svelte. The entire dashboard embeds into the Go binary.

**Why this over a SPA framework:**
- The project already has a Go/Gin backend with full REST API -- serve HTML from the same process
- Single user, no SEO, no complex client-side routing -- SPA is overkill
- htmx + Alpine.js covers 95% of dashboard interactivity (tables, feeds, controls, modals)
- The 5% that needs real JS (financial charts) uses Lightweight Charts directly
- Zero build toolchain: no webpack, no npm build, no bundler. Just Go templates + CDN scripts
- Single binary deployment: `go build` produces one artifact with all assets embedded

### Core Technologies

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| Go `html/template` | stdlib | Server-side HTML templating | Already in Go stdlib, no build step, sufficient for dashboard layouts. Templ (v0.3.977) is better DX but adds a build tool and file format for a dashboard addon -- not worth the complexity. | HIGH |
| Go `embed` | stdlib (Go 1.22+) | Embed static assets in binary | Bake HTML, CSS, JS into the Go binary. Single artifact deployment. Use `//go:embed` directive on a `static/` directory. | HIGH |
| htmx | 2.0.7 | Dynamic HTML without JavaScript | Replaces AJAX/fetch calls with HTML attributes. Server returns HTML fragments, htmx swaps them into the DOM. Perfect for updating position tables, activity feeds, order lists. WebSocket extension handles real-time push. | HIGH |
| htmx ws extension | 2.0.x | WebSocket integration for htmx | Enables `ws-connect` attribute on elements. Server pushes HTML fragments over WebSocket, htmx morphs the DOM. Auto-reconnect with exponential backoff built in. | HIGH |
| Alpine.js | 3.15.x | Client-side reactivity | Lightweight (15KB) reactive framework for things htmx cannot do: toggles, modals, tabs, dropdown menus, client-side filtering, confirmation dialogs for emergency controls. Complements htmx perfectly. | HIGH |
| TradingView Lightweight Charts | 5.1.0 | Financial charting | The standard for web-based financial charts. 35KB gzipped, Apache 2.0 license, multi-pane support in v5, `series.update()` for real-time streaming. No viable alternative at this quality/size. | HIGH |
| gorilla/websocket | 1.5.3 | WebSocket server (Go side) | Battle-tested, 24.5K GitHub stars, stable API, works cleanly with Gin via `Upgrader`. The project uses Gin which integrates trivially. coder/websocket (v1.8.14) is more idiomatic but gorilla has more examples, docs, and community patterns for the hub/broadcast pattern this dashboard needs. | HIGH |
| Tailwind CSS | 4.1.x | Utility-first CSS | No custom CSS files to manage. Tailwind v4 is 5x faster builds, zero-config with `@import "tailwindcss"`. Use the standalone CLI to generate a single CSS file at build time, embed it in the binary. Dashboard styling without writing CSS. | MEDIUM |

### Supporting Libraries

| Library | Version | Purpose | When to Use | Confidence |
|---------|---------|---------|-------------|------------|
| gin-contrib/static | latest | Serve embedded static files | Serve the embedded `static/` directory at `/` with `static.EmbedFolder()`. Handles SPA-like fallback via Gin `NoRoute` handler. | HIGH |
| Alpaca Go SDK `marketdata/stream` | v3.5.0+ (upgrade to v3.9.1) | WebSocket market data streaming | Already in the project (v3.5.0). The SDK provides `StreamTradeUpdatesInBackground()` and the `marketdata/stream` package for real-time quotes/trades/bars via Alpaca's WebSocket. Upgrade to v3.9.1 for latest fixes. | HIGH |
| heroicons or lucide | latest | SVG icon set | Dashboard needs icons for navigation, status indicators, controls. Use inline SVGs from heroicons (by Tailwind team) or lucide. No icon font, no extra requests. | LOW |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| Tailwind standalone CLI | Generate CSS from templates | Download the binary, run `tailwindcss -i input.css -o static/css/output.css --minify`. No Node.js needed. Add to a Makefile target. |
| Air | Hot reload during development | Watches `.go` and `.html` files, rebuilds and restarts. Config in `.air.toml`. Avoid watching generated output to prevent loops. |
| Browser DevTools | WebSocket debugging | Chrome/Firefox DevTools Network tab shows WebSocket frames. Essential for debugging real-time updates. |

## Installation

```bash
# Go dependencies (add to existing go.mod)
go get github.com/gorilla/websocket@v1.5.3
go get github.com/gin-contrib/static@latest

# Upgrade Alpaca SDK to latest
go get github.com/alpacahq/alpaca-trade-api-go/v3@v3.9.1

# Frontend (CDN, no install -- loaded via script tags in templates)
# htmx:               https://unpkg.com/htmx.org@2.0.7
# htmx ws extension:  https://unpkg.com/htmx-ext-ws@2.0.7
# Alpine.js:          https://cdn.jsdelivr.net/npm/alpinejs@3.15.8/dist/cdn.min.js
# Lightweight Charts: https://unpkg.com/lightweight-charts@5.1.0/dist/lightweight-charts.standalone.production.mjs

# OR download and embed (preferred for single-binary deployment):
mkdir -p static/js static/css
curl -o static/js/htmx.min.js https://unpkg.com/htmx.org@2.0.7/dist/htmx.min.js
curl -o static/js/htmx-ws.js https://unpkg.com/htmx-ext-ws@2.0.7/ext/ws.js
curl -o static/js/alpine.min.js https://cdn.jsdelivr.net/npm/alpinejs@3.15.8/dist/cdn.min.js
curl -o static/js/lightweight-charts.js https://unpkg.com/lightweight-charts@5.1.0/dist/lightweight-charts.standalone.production.mjs

# Tailwind CSS standalone CLI (Linux x64)
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
chmod +x tailwindcss-linux-x64
mv tailwindcss-linux-x64 ./tailwindcss

# Dev tools
go install github.com/air-verse/air@latest
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| htmx + Alpine.js | React/Vue/Svelte SPA | If the dashboard grows to need complex client-side state (multi-step wizards, drag-and-drop portfolio rebalancing, offline mode). For this single-user monitoring dashboard, SPA is overkill. |
| htmx + Alpine.js | Templ + htmx | If the project were a full web application with dozens of pages and forms, Templ's type-safe components and IDE support would justify the build step. For a dashboard addon, `html/template` is sufficient. |
| gorilla/websocket | coder/websocket (v1.8.14) | If you want idiomatic context.Context support, zero-alloc reads, and built-in concurrent writes. gorilla is chosen here because the Gin + gorilla combo has the most documentation and examples for the hub/broadcast pattern. |
| Tailwind CSS v4 | Plain CSS / Pico CSS | If you want zero build tools at all. Pico CSS provides classless semantic styling. Tailwind is chosen because the dashboard has many components (tables, cards, badges, grids) where utility classes are significantly faster to iterate on. |
| Go `html/template` | Templ (v0.3.977) | If starting a greenfield Go web app. Templ's type-safe components, LSP support, and compiled performance are superior. But it adds a CLI tool (`templ generate`), a new file format (`.templ`), and a build step to a project that currently has none for the frontend. |
| Downloaded + embedded JS | CDN script tags | If you do not care about single-binary deployment or offline operation. CDN is simpler for development but means the dashboard requires internet access to load. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| React / Next.js / Vue / Svelte | Massive overkill for a single-user dashboard. Adds Node.js build toolchain, npm dependencies, separate dev server, API serialization layer. The Go backend already serves the data. | htmx + Alpine.js |
| Server-Sent Events (SSE) | One-directional (server-to-client only). The dashboard needs bidirectional communication for emergency controls (close position, cancel order) that should flow through the same WebSocket connection. | WebSocket via gorilla/websocket |
| GraphQL | Adds a query layer between the Go backend and frontend for no benefit. REST endpoints already exist. Single consumer (the dashboard). | Existing REST API + WebSocket |
| WebSocket-only architecture | Tempting to put everything over WebSocket, but REST is better for request-response patterns (fetch orders, get historical data). Use WebSocket for streaming only. | REST for queries, WebSocket for streaming |
| Sass / Less / CSS Modules | Additional build tooling for CSS. Tailwind utility classes eliminate the need for custom CSS authoring in most cases. | Tailwind CSS |
| jQuery | Legacy. htmx and Alpine.js cover all the use cases jQuery would serve, with cleaner APIs. | htmx + Alpine.js |
| templ (for this project) | Adds build complexity (CLI, code generation, new file format) to a project that is a dashboard addon, not a full web application. The productivity gain does not justify the toolchain cost here. | Go `html/template` |

## Stack Patterns by Variant

**For real-time position/price updates (streaming data):**
- WebSocket connection from browser via htmx `ws-connect`
- Go backend runs a WebSocket hub (gorilla/websocket) with broadcast to connected clients
- Hub receives updates from Alpaca SDK's `marketdata/stream` and internal position manager
- Server renders HTML fragments and pushes them over WebSocket
- htmx swaps the fragments into the DOM (e.g., update a position row, price badge)

**For financial charts (Lightweight Charts):**
- Charts are initialized with vanilla JS using Lightweight Charts API
- Historical data loaded via REST fetch on page load
- Real-time updates come over the same WebSocket as JSON (not HTML)
- A small Alpine.js component or vanilla JS handler calls `series.update()` with new bar/tick data
- This is the one area where raw JS is needed -- htmx cannot drive canvas chart updates

**For emergency controls (close position, cancel order):**
- Alpine.js handles the confirmation dialog UI (`x-data`, `x-show`, `x-on:click`)
- On confirm, htmx sends the action via `hx-delete` or `hx-post` to the REST API
- Response swaps a status indicator (e.g., position row shows "Closing...")
- WebSocket push updates the final state when Alpaca confirms

**For initial page load / navigation:**
- Full HTML page rendered by Go `html/template` on first load
- Subsequent panel switches use htmx `hx-get` to swap panel contents
- Browser URL updated via htmx `hx-push-url` for back-button support

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| Go 1.22 | gorilla/websocket 1.5.3 | Confirmed compatible. Project already on Go 1.22. |
| Go 1.22 | gin 1.10.0 | Already in use in the project. |
| Go 1.22 | gin-contrib/static | Uses Go `embed` (requires Go 1.16+). |
| htmx 2.0.7 | htmx-ext-ws 2.0.x | Extensions moved to separate repo in 2.0. Must use matching major version. |
| htmx 2.0.7 | Alpine.js 3.15.x | No conflicts. htmx operates on attributes, Alpine on `x-` attributes. They coexist cleanly. |
| Lightweight Charts 5.1.0 | Any (standalone) | No framework dependency. Pure JS/Canvas. Works alongside htmx and Alpine.js. |
| Tailwind CSS 4.1.x | Go html/template | Tailwind scans any file for class names. Configure `content` paths to include `.html` template files. |
| Alpaca SDK v3.9.1 | Go 1.22 | Upgrade from current v3.5.0. Check release notes for breaking changes in minor versions. |

## Sources

- [gorilla/websocket GitHub](https://github.com/gorilla/websocket) -- v1.5.3, stable API, 24.5K stars
- [coder/websocket GitHub](https://github.com/coder/websocket) -- v1.8.14, idiomatic alternative
- [TradingView Lightweight Charts](https://tradingview.github.io/lightweight-charts/) -- v5.1.0, real-time updates docs
- [TradingView Lightweight Charts Releases](https://github.com/tradingview/lightweight-charts/releases) -- v5.1.0 Dec 2024
- [htmx.org](https://htmx.org/) -- v2.0.7, WebSocket extension docs
- [htmx WebSocket Extension](https://htmx.org/extensions/ws/) -- ws-connect, auto-reconnect
- [Alpine.js](https://alpinejs.dev/) -- v3.15.8
- [Tailwind CSS v4](https://tailwindcss.com/blog/tailwindcss-v4) -- v4.1.18, standalone CLI
- [gin-contrib/static](https://github.com/gin-contrib/static) -- EmbedFolder() for Go embed
- [Alpaca WebSocket Streaming Docs](https://docs.alpaca.markets/docs/streaming-market-data) -- market data WebSocket
- [Alpaca Go SDK](https://pkg.go.dev/github.com/alpacahq/alpaca-trade-api-go/v3) -- v3.9.1, marketdata/stream package
- [Go embed package](https://pkg.go.dev/embed) -- stdlib file embedding
- [Ersin's Go Web Stack 2025](https://www.ersin.nz/articles/a-great-web-stack-for-go) -- templ + htmx pattern reference
- [Gin + Gorilla WebSocket integration](https://arlimus.github.io/articles/gin.and.gorilla/) -- hub/broadcast pattern
- [templ](https://templ.guide/) -- v0.3.977, evaluated but not recommended for this project

---
*Stack research for: Prophet Trader real-time web dashboard*
*Researched: 2026-02-11*
