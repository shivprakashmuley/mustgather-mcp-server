package cmd

import (
	"errors"
	"fmt"
	omcServer "github.com/shivprakashmuley/omc-mcp-server/pkg/mcp/server"
	"github.com/shivprakashmuley/omc-mcp-server/pkg/mcp/version"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var rootCmd = &cobra.Command{
	Use:   "omc-mcp-server [command] [options]",
	Short: "OpenshiftMustGatherClient Model Context Protocol (MCP) server",
	Long: `
OpenshiftMustGatherClient Model Context Protocol (MCP) server

  # show this help
  omc-mcp-server -h

  # shows version information
  omc-mcp-server --version

  # start STDIO server
  omc-mcp-server

  # start a SSE server on port 8080
  omc-mcp-server --sse-port 8080

  # start a SSE server on port 8443 with a public HTTPS host of example.com
  omc-mcp-server --sse-port 8443 --sse-base-url https://example.com:8443

  # TODO: add more examples`,
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
