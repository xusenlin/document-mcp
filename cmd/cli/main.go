package main

import (
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xusenlin/document-mcp/internal/cli"
	"github.com/xusenlin/document-mcp/internal/server"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		cli.Run(os.Args[1:])
		return
	}

	s := mcp.NewServer(
		&mcp.Implementation{Name: "document-mcp", Version: version},
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

	log.Printf("document-mcp %s starting on %s (stateless)", version, addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
