package mcpadapter

import (
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"sshub-mcp/internal/domain/ports"
)

func NewHandler(tools *ToolSet, resolver ports.TokenResolver) http.Handler {
	srv := mcp.NewServer(&mcp.Implementation{Name: "sshub-mcp", Version: "0.1.0"}, nil)
	RegisterTools(srv, tools)
	inner := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv }, nil)
	return authMiddleware(inner, AuthConfig{Resolver: resolver})
}
