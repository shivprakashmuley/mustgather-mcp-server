package server

import (
	"slices"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	server *server.MCPServer
}

func NewSever() (*Server, error) {
	s := &Server{
		server: server.NewMCPServer(
			"mustgather-mcp-server",
			"0.0.1",
			server.WithInstructions("kubectl like functionality but read only access to a openshift kubernetes cluster snapshot"),
			server.WithResourceCapabilities(true, true),
			server.WithPromptCapabilities(true),
			server.WithToolCapabilities(true),
			server.WithLogging(),
		),
	}
	s.server.AddTools(slices.Concat(
		s.initOMC(),
	)...)
	return s, nil
}

func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.server)
}

func (s *Server) ServeSse(baseUrl string) *server.SSEServer {
	options := make([]server.SSEOption, 0)
	if baseUrl != "" {
		options = append(options, server.WithBaseURL(baseUrl))
	}
	return server.NewSSEServer(s.server, options...)
}

func NewTextResult(content string, err error) *mcp.CallToolResult {
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: err.Error(),
				},
			},
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: content,
			},
		},
	}
}
