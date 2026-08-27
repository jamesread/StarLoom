// Package mcp exposes StarApp API operations as MCP tools on /mcp.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/auth"
	"github.com/jamesread/starapp/service/internal/buildinfo"
	srvpkg "github.com/jamesread/starapp/service/internal/server"
)

// NewHandler returns an http.Handler for the MCP Streamable HTTP endpoint at /mcp.
// Run auth.Middleware before this handler so MCP uses the same auth as the Connect API.
func NewHandler(srv *srvpkg.Server) http.Handler {
	mcpServer := server.NewMCPServer(
		"StarApp",
		buildinfo.Version,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	mcpServer.AddTool(mcp.NewTool("starapp_ping",
		mcp.WithDescription("Health check — verifies connectivity and returns Init metadata."),
	), func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleInit(ctx, srv)
	})

	mcpServer.AddTool(mcp.NewTool("starapp_init",
		mcp.WithDescription("Return SPA bootstrap metadata (title, version, features, webhook events)."),
	), func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleInit(ctx, srv)
	})

	mcpServer.AddTool(mcp.NewTool("starapp_list_cvars",
		mcp.WithDescription("List runtime configuration variables (cvars)."),
	), func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := srv.ListCvars(ctx, connect.NewRequest(&apiv1.ListCvarsRequest{}))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(map[string]any{"cvars": res.Msg.GetCvars()})
	})

	mcpServer.AddTool(mcp.NewTool("starapp_update_cvar",
		mcp.WithDescription("Update a configuration variable by key."),
		mcp.WithString("key", mcp.Required(), mcp.Description("Cvar key (e.g. site_title, show_footer)")),
		mcp.WithNumber("value_int", mcp.Description("Integer/boolean value (0 or 1 for booleans)")),
		mcp.WithString("value_string", mcp.Description("String value when the cvar type is string")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := requireWrite(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		key, err := req.RequireString("key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := srv.UpdateCvar(ctx, connect.NewRequest(&apiv1.UpdateCvarRequest{
			Key:         key,
			ValueInt:    int32(req.GetFloat("value_int", 0)),
			ValueString: req.GetString("value_string", ""),
		}))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(res.Msg)
	})

	mcpServer.AddTool(mcp.NewTool("starapp_list_webhooks",
		mcp.WithDescription("List outbound webhook targets and supported event names."),
	), func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := srv.ListWebhooks(ctx, connect.NewRequest(&apiv1.ListWebhooksRequest{}))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(map[string]any{
			"webhooks": res.Msg.GetWebhooks(),
			"events":   res.Msg.GetEvents(),
		})
	})

	mcpServer.AddTool(mcp.NewTool("starapp_create_webhook",
		mcp.WithDescription("Create an outbound webhook target."),
		mcp.WithString("url", mcp.Required(), mcp.Description("HTTPS callback URL")),
		mcp.WithString("secret", mcp.Required(), mcp.Description("HMAC signing secret")),
		mcp.WithString("events", mcp.Required(), mcp.Description("Comma-separated event names (e.g. stars.awarded,redemption.requested)")),
		mcp.WithBoolean("enabled", mcp.Description("Whether delivery is enabled (default true)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := requireWrite(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		eventsStr, err := req.RequireString("events")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		url, err := req.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		secret, err := req.RequireString("secret")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := srv.CreateWebhook(ctx, connect.NewRequest(&apiv1.CreateWebhookRequest{
			Url:     url,
			Secret:  secret,
			Events:  splitCSV(eventsStr),
			Enabled: req.GetBool("enabled", true),
		}))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(res.Msg.GetWebhook())
	})

	mcpServer.AddTool(mcp.NewTool("starapp_update_webhook",
		mcp.WithDescription("Update an outbound webhook target."),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("Webhook target ID")),
		mcp.WithString("url", mcp.Description("HTTPS callback URL")),
		mcp.WithString("secret", mcp.Description("New HMAC secret (omit to keep existing)")),
		mcp.WithString("events", mcp.Description("Comma-separated event names")),
		mcp.WithBoolean("enabled", mcp.Description("Whether delivery is enabled")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := requireWrite(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id := int32(req.GetFloat("id", 0))
		if id < 1 {
			return mcp.NewToolResultError("id is required"), nil
		}
		res, err := srv.UpdateWebhook(ctx, connect.NewRequest(&apiv1.UpdateWebhookRequest{
			Id:      id,
			Url:     req.GetString("url", ""),
			Secret:  req.GetString("secret", ""),
			Events:  splitCSV(req.GetString("events", "")),
			Enabled: req.GetBool("enabled", true),
		}))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(res.Msg)
	})

	mcpServer.AddTool(mcp.NewTool("starapp_delete_webhook",
		mcp.WithDescription("Delete an outbound webhook target."),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("Webhook target ID")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := requireWrite(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id := int32(req.GetFloat("id", 0))
		if id < 1 {
			return mcp.NewToolResultError("id is required"), nil
		}
		_, err := srv.DeleteWebhook(ctx, connect.NewRequest(&apiv1.DeleteWebhookRequest{Id: id}))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(map[string]any{"deleted": true, "id": id})
	})

	return server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath("/mcp"))
}

func handleInit(ctx context.Context, srv *srvpkg.Server) (*mcp.CallToolResult, error) {
	res, err := srv.Init(ctx, connect.NewRequest(&apiv1.InitRequest{}))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return jsonResult(res.Msg)
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func requireWrite(ctx context.Context) error {
	if au := auth.UserFromContext(ctx); au != nil && au.ReadOnly {
		return fmt.Errorf("read-only API key cannot perform this action")
	}
	return nil
}
