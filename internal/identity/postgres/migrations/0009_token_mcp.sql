-- Whether a script token may use the MCP server (docs/DESIGN.md D-210, D-213). The owner chooses it when the token is
-- created, and the Access page shows it. It is stored only: nothing serves MCP yet, so the flag changes no access.
-- Tokens that exist now were made without it, so they are not allowed.
ALTER TABLE identity.script_tokens ADD COLUMN mcp_allowed boolean NOT NULL DEFAULT false;
