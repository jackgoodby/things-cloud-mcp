package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	thingscloud "github.com/arthursoares/things-cloud-sdk"
)

// historyEntry is one raw history record for an item, as stored in Things Cloud.
type historyEntry struct {
	UUID   string          `json:"uuid"`
	Kind   string          `json:"e"`
	Action int             `json:"t"`
	P      json.RawMessage `json:"p,omitempty"`
}

// handleDebugHistory returns every raw history record for the given item UUIDs,
// in server order. It reads the full history and changes nothing.
func (t *ThingsMCP) handleDebugHistory(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := req.RequireString("uuids")
	if err != nil {
		return errResult("uuids is required"), nil
	}
	wanted := map[string]bool{}
	for _, u := range strings.Split(raw, ",") {
		if u = strings.TrimSpace(u); u != "" {
			wanted[u] = true
		}
	}
	if len(wanted) == 0 {
		return errResult("uuids is required"), nil
	}

	history := t.client.HistoryWithID(t.history.ID)
	startIndex := 0
	var entries []historyEntry
	for {
		items, hasMore, err := history.Items(thingscloud.ItemsOptions{StartIndex: startIndex})
		if err != nil {
			return errResult(fmt.Sprintf("fetch items: %v", err)), nil
		}
		for _, item := range items {
			if wanted[item.UUID] {
				entries = append(entries, historyEntry{UUID: item.UUID, Kind: string(item.Kind), Action: int(item.Action), P: item.P})
			}
		}
		if len(items) == 0 || !hasMore {
			break
		}
		startIndex = history.LoadedServerIndex
	}
	return jsonResult(map[string]any{"count": len(entries), "entries": entries}), nil
}

func debugHistoryTool() mcp.Tool {
	return mcp.NewTool("things_debug_history",
		mcp.WithDescription("Show every raw Things Cloud history record (kind, action, payload) for one or more item UUIDs, in server order. Read-only; for debugging how the Things apps write items."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("uuids", mcp.Required(), mcp.Description("Comma-separated full item UUIDs")),
	)
}
