# ChatGPT Terminal Monitor MCP App (Development)

This optional integration adds a read-only terminal monitor to the existing Terminal MCP server. It keeps all existing terminal tools and routes unchanged.

## What it provides

- `terminal_monitor` MCP tool with an MCP Apps HTML resource at `ui://terminal-mcp/monitor-v1.html`.
- A session selector, session count, refresh button, and terminal output with automatic polling.
- OpenAI conversation-thread entrypoint metadata to allow opening the monitor in a chat-side tab when supported.
- English-language panel labels and status messages.
- No terminal input controls in the embedded panel. The original Terminal MCP tool permissions and identity checks still apply.

## Development configuration

Build the Go server on an isolated branch, run it on a separate development port, and connect that endpoint through a distinct secure tunnel and ChatGPT development plugin. Do not replace or restart production just to evaluate the panel.

In the ChatGPT development plugin, call `terminal_monitor` in a conversation, then use **Open app in tab** to move the viewer to the conversation-side panel. The embedded viewer uses `tools/call` for `terminal_list` and `terminal_output` and polls for new terminal output.

## Tests performed

- `go test ./...` and `go test ./mcpserver -run TestPanelMCPDiscovery -v`.
- MCP `tools/list`, `resources/list`, and `resources/read`, including the conversation-thread entrypoint metadata.
- Development ChatGPT plugin connected through its own tunnel.
- Verified iframe rendering of the live terminal session in ChatGPT.
- Verified browser-side panel and automatic receipt of new output from the Ubuntu development terminal.

Production deployment remains a separate manual decision. Changes are not merged into the existing production branch.
