# ChatGPT Terminal Monitor

This integration is read-only and maintains the existing Terminal MCP routes and permissions.

## Behavior

- `terminal_monitor` opens the ChatGPT MCP Apps monitor.
- `terminal_open` advertises the same UI resource so supported ChatGPT hosts present a **new live monitor** for each newly created PTY session. Actual placement and whether the host renders a separate panel is controlled by ChatGPT; the tool cannot force an external tab or side pane.
- The active agent session is tracked per authenticated MCP owner when the agent opens, sends to, controls, checks status on, or explores a session. Viewer polling is *not* counted as agent activity.
- On opening the monitor, the agent-active session is selected automatically. When another agent session is created, the monitor detects it within the periodic refresh interval and switches to it. A user can manually select any existing session.
- **Follow: On** is the default. New output follows the vertical end of the terminal. Click Follow to turn it off, preserving the manual scroll position. Click it again to resume bottom-following.
- **Refresh** immediately reloads the session list and terminal output. Session list and output are also updated periodically. A failed refresh displays an error and can be retried.
- Terminal polling uses only `terminal_list` and `terminal_output(mode=tail)` and never sends commands, closes sessions, or enables human takeover.
- New session presentation via tool metadata requires a host that supports MCP Apps tool UI metadata; older hosts can explicitly call `terminal_monitor`.

## Verification

1. `go test ./...` verifies existing terminal behavior and UI discovery metadata.
2. `NODE_PATH=/home/virat/.local/share/playwright-mcp/node_modules node scripts/panel-browser-test.cjs` tests the exact app HTML in a Chrome iframe, including selection, new sessions, Follow, Refresh and recovery.
3. For a safe live test, build in Docker, run the binary on an isolated loopback port, and check MCP `tools/list`, `terminal_open`, `terminal_list`, `terminal_monitor`, `terminal_send`, `terminal_output`, and `terminal_close` with an authenticated test identity.
4. When productionizing, save the original binary, deploy via an independent systemd job, verify the service and connector, and keep an automatic rollback timer until checks succeed.

Production runs through `terminal-mcp.service`; do not alter existing Tailscale Serve rules, desktop/Wayland settings, or unrelated MCP services as part of this change.

## Cache invalidation

Monitor v2 has a distinct MCP Apps resource URI. Version the URI whenever the embedded HTML changes: some ChatGPT clients cache the HTML resource for an existing URI, which can otherwise display old buttons after a server upgrade.
