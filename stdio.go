package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/server"
)

// stdioRequested reports whether the server should talk MCP over stdin/stdout
// instead of listening on HTTP. A client such as Claude Desktop then starts
// and stops the process itself, so nothing runs in the background.
func stdioRequested(args []string) bool {
	for _, a := range args {
		if a == "-stdio" || a == "--stdio" {
			return true
		}
	}
	return false
}

// stdioCredentials reads the Things Cloud login for stdio mode from
// THINGS_AUTH ("Basic <base64 email:password>", as stored for the HTTP
// launcher) or from THINGS_EMAIL and THINGS_PASSWORD.
func stdioCredentials(getenv func(string) string) (string, string, error) {
	if auth := strings.TrimSpace(getenv("THINGS_AUTH")); auth != "" {
		raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(strings.TrimPrefix(auth, "Basic ")))
		if err != nil {
			return "", "", fmt.Errorf("THINGS_AUTH is not a Basic credential: %w", err)
		}
		email, password, ok := strings.Cut(string(raw), ":")
		if !ok || email == "" || password == "" {
			return "", "", fmt.Errorf("THINGS_AUTH does not contain email:password")
		}
		return email, password, nil
	}
	email, password := getenv("THINGS_EMAIL"), getenv("THINGS_PASSWORD")
	if email == "" || password == "" {
		return "", "", fmt.Errorf("set THINGS_AUTH, or THINGS_EMAIL and THINGS_PASSWORD")
	}
	return email, password, nil
}

// serveStdio runs the MCP server over stdin/stdout for a single user.
// Logs go to stderr, which keeps stdout free for the protocol.
func serveStdio(mcpServer *server.MCPServer) error {
	email, password, err := stdioCredentials(os.Getenv)
	if err != nil {
		return err
	}
	user := &UserInfo{Email: email, Password: password}
	return server.ServeStdio(mcpServer, server.WithStdioContextFunc(func(ctx context.Context) context.Context {
		return context.WithValue(ctx, userContextKey, user)
	}))
}
