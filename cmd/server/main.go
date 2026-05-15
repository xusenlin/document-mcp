package main

import (
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xusenlin/document-mcp/internal/server"
)

func main() {
	s := mcp.NewServer(
		&mcp.Implementation{Name: "document-mcp", Version: "v0.1.0"},
		&mcp.ServerOptions{
			Capabilities: &mcp.ServerCapabilities{
				Tools: &mcp.ToolCapabilities{},
			},
		},
	)

	server.RegisterTools(s)

	addr := os.Getenv("MCP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return s },
		&mcp.StreamableHTTPOptions{
			Stateless:    true,
			JSONResponse: true,
		},
	)

	log.Printf("document-mcp v0.1.0 starting on %s (stateless)", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
