package main

import (
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	omcServer "github.com/shivprakashmuley/mustgather-mcp-server/mcp/server"
	"github.com/shivprakashmuley/mustgather-mcp-server/mcp/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var rootCmd = &cobra.Command{
	Use:   "mustgather-mcp-server [command] [options]",
	Short: "OpenShift must-gather Model Context Protocol (MCP) server",
	Long: `
OpenShift must-gather Model Context Protocol (MCP) server

  # show this help
  mustgather-mcp-server -h

  # shows version information
  mustgather-mcp-server --version

  # start STDIO server
  mustgather-mcp-server

  # start a SSE server on port 8080
  mustgather-mcp-server --sse-port 8080

  # start a SSE server on port 8443 with a public HTTPS host of example.com
  mustgather-mcp-server --sse-port 8443 --sse-base-url https://example.com:8443`,

	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("version") {
			fmt.Println(version.Version)
			return
		}
		mcpServer, err := omcServer.NewSever()
		if err != nil {
			panic(err)
		}

		var sseServer *server.SSEServer
		if ssePort := viper.GetInt("sse-port"); ssePort > 0 {
			sseServer = mcpServer.ServeSse(viper.GetString("sse-base-url"))
			if err := sseServer.Start(fmt.Sprintf(":%d", ssePort)); err != nil {
				panic(err)
			}
		}
		if err := mcpServer.ServeStdio(); err != nil && !errors.Is(err, context.Canceled) {
			panic(err)
		}
		if sseServer != nil {
			_ = sseServer.Shutdown(cmd.Context())
		}
	},
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	rootCmd.Flags().IntP("sse-port", "", 0, "Start a SSE server on the specified port")
	rootCmd.Flags().StringP("sse-base-url", "", "", "SSE public base URL to use when sending the endpoint message (e.g. https://example.com)")
	_ = viper.BindPFlags(rootCmd.Flags())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
