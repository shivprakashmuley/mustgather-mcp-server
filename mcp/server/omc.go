package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const execTimeout = 2 * time.Minute

// Register the CLI tools for the OMC
func (s *Server) initOMC() []server.ServerTool {
	return []server.ServerTool{
		// 1. omc get
		{mcp.NewTool("mustgather_get",
			mcp.WithDescription("Get kubernetes and openshift resources using oc get command"),
			mcp.WithString("kind", mcp.Description("Resource kind"), mcp.Required()),
			mcp.WithBoolean("all_namespaces", mcp.Description("Get resources from all namespaces (-A flag)")),
			mcp.WithString("namespace", mcp.Description("Namespace to get resources from (-n flag)")),
			mcp.WithString("output", mcp.Description("Output format"), mcp.Enum("wide", "yaml", "json")),
		), func(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("mustgather_get{}")

			kind := ctr.Params.Arguments["kind"].(string)

			cmdArgs := []string{"get", kind}

			if allNamespaces, ok := ctr.Params.Arguments["all_namespaces"].(bool); ok && allNamespaces {
				cmdArgs = append(cmdArgs, "-A")
			} else if namespace, ok := ctr.Params.Arguments["namespace"].(string); ok {
				cmdArgs = append(cmdArgs, "-n", namespace)
			}

			if output, ok := ctr.Params.Arguments["output"].(string); ok {
				cmdArgs = append(cmdArgs, "-o", output)
			}

			result, err := executeOMCCommand(cmdArgs)
			return NewTextResult(result, err), nil
		}},

		// 2. omc describe
		{mcp.NewTool("mustgather_describe",
			mcp.WithDescription("Describe pods or nodes using oc describe command, other resources are not supported."),
			mcp.WithString("kind", mcp.Description("Resource kind (pods or nodes only)"), mcp.Required(), mcp.Enum("pods", "nodes")),
			mcp.WithBoolean("all_namespaces", mcp.Description("Describe resources from all namespaces (-A flag)")),
			mcp.WithString("namespace", mcp.Description("Namespace to describe resources from (-n flag)")),
			mcp.WithString("output", mcp.Description("Output format"), mcp.Enum("wide", "yaml")),
		), func(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("mustgather_describe{}")

			kind := ctr.Params.Arguments["kind"].(string)
			if kind != "pods" && kind != "nodes" {
				return NewTextResult("", fmt.Errorf("describe only supports 'pods' or 'nodes'")), nil
			}

			cmdArgs := []string{"describe", kind}

			if allNamespaces, ok := ctr.Params.Arguments["all_namespaces"].(bool); ok && allNamespaces {
				cmdArgs = append(cmdArgs, "-A")
			} else if namespace, ok := ctr.Params.Arguments["namespace"].(string); ok {
				cmdArgs = append(cmdArgs, "-n", namespace)
			}

			if output, ok := ctr.Params.Arguments["output"].(string); ok {
				cmdArgs = append(cmdArgs, "-o", output)
			}

			result, err := executeOMCCommand(cmdArgs)
			return NewTextResult(result, err), nil
		}},

		// 3. omc logs
		{mcp.NewTool("mustgather_logs",
			mcp.WithDescription("Get logs from a specific pod and container"),
			mcp.WithString("pod_name", mcp.Description("Name of the pod to get logs from"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace of the pod"), mcp.Required()),
			mcp.WithString("container", mcp.Description("Container name within the pod"), mcp.Required()),
		), func(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("mustgather_logs{}")

			podName := ctr.Params.Arguments["pod_name"].(string)
			namespace := ctr.Params.Arguments["namespace"].(string)
			container := ctr.Params.Arguments["container"].(string)

			cmdArgs := []string{"logs", podName, "-n", namespace, "-c", container}
			result, err := executeOMCCommand(cmdArgs)
			return NewTextResult(result, err), nil
		}},

		// 4. omc events
		{mcp.NewTool("mustgather_events",
			mcp.WithDescription("Get cluster events using oc events command"),
			mcp.WithBoolean("all_namespaces", mcp.Description("Get events from all namespaces (-A flag)")),
			mcp.WithString("namespace", mcp.Description("Namespace to get events from (-n flag)")),
			mcp.WithString("for", mcp.Description("Filter events for a specific resource (--for flag)")),
			mcp.WithString("output", mcp.Description("Output format"), mcp.Enum("yaml", "name")),
		), func(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("mustgather_events{}")

			cmdArgs := []string{"events"}

			if allNamespaces, ok := ctr.Params.Arguments["all_namespaces"].(bool); ok && allNamespaces {
				cmdArgs = append(cmdArgs, "-A")
			} else if namespace, ok := ctr.Params.Arguments["namespace"].(string); ok {
				cmdArgs = append(cmdArgs, "-n", namespace)
			}

			if forResource, ok := ctr.Params.Arguments["for"].(string); ok {
				cmdArgs = append(cmdArgs, "--for", forResource)
			}

			if output, ok := ctr.Params.Arguments["output"].(string); ok {
				if output != "yaml" && output != "name" {
					return NewTextResult("", fmt.Errorf("events only supports 'yaml' or 'name' output")), nil
				}
				cmdArgs = append(cmdArgs, "-o", output)
			}

			result, err := executeOMCCommand(cmdArgs)
			return NewTextResult(result, err), nil
		}},

		// 5. omc node-logs
		{mcp.NewTool("mustgather_node_logs",
			mcp.WithDescription("Get node logs for a specific journalctl service like NetworkManager, crio, kubelet, machine-config-daemon-firstboot, machine-config-daemon-host, openvswitch, ostree-finalize-staged, ovs-configuration, ovs-vswitchd, ovsdb-server, rpm-ostreed"),
			mcp.WithString("service_name", mcp.Description("Journalctl service name to get logs from"), mcp.Required()),
		), func(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("mustgather_node_logs{}")

			serviceName := ctr.Params.Arguments["service_name"].(string)

			cmdArgs := []string{"node-logs", serviceName}
			result, err := executeOMCCommand(cmdArgs)
			return NewTextResult(result, err), nil
		}},

		// 6. omc haproxy backends
		{mcp.NewTool("mustgather_haproxy_backends",
			mcp.WithDescription("Get HAProxy backends information from openshift router"),
		), func(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("mustgather_haproxy_backends{}")

			cmdArgs := []string{"haproxy", "backends"}
			result, err := executeOMCCommand(cmdArgs)
			return NewTextResult(result, err), nil
		}},

		// 7. omc etcd health
		{mcp.NewTool("mustgather_etcd_health",
			mcp.WithDescription("Check etcd cluster health"),
		), func(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("mustgather_etcd_health{}")

			cmdArgs := []string{"etcd", "health"}
			result, err := executeOMCCommand(cmdArgs)
			return NewTextResult(result, err), nil
		}},

		// 8. omc etcd status
		{mcp.NewTool("mustgather_etcd_status",
			mcp.WithDescription("Get etcd cluster status"),
		), func(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("mustgather_etcd_status{}")

			cmdArgs := []string{"etcd", "status"}
			result, err := executeOMCCommand(cmdArgs)
			return NewTextResult(result, err), nil
		}},

		// 9. omc projects
		{mcp.NewTool("mustgather_projects",
			mcp.WithDescription("List available projects / namespaces in the cluster"),
		), func(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("mustgather_projects{}")

			cmdArgs := []string{"projects"}
			result, err := executeOMCCommand(cmdArgs)
			return NewTextResult(result, err), nil
		}},

		// 10. omc use
		{mcp.NewTool("mustgather_use",
			mcp.WithDescription("Switch to a different mustgather snapshot directory: supports https, local, gcs bucket"),
			mcp.WithString("path", mcp.Description("Path to switch to"), mcp.Required()),
		), func(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("mustgather_use{}")

			path := ctr.Params.Arguments["path"].(string)

			cmdArgs := []string{"use", path}
			result, err := executeOMCCommand(cmdArgs)
			return NewTextResult(result, err), nil
		}},

		// 11. download_must_gather
		{mcp.NewTool("download_must_gather",
			mcp.WithDescription("Download must-gather from a specific URL"),
			mcp.WithString("url", mcp.Description("URL of the must-gather to download"), mcp.Required()),
		), func(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			log.Printf("download_must_gather{}")

			result, err := s.DownloadMustGather(ctx, ctr.Params.Arguments["url"].(string))
			return NewTextResult(result, err), nil
		}},
	}
}

func executeOMCCommand(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "omc", args...)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("command timed out after 30 seconds")
	}

	if err != nil {
		return string(output), fmt.Errorf("command failed: %v\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// DownloadMustGather implements the "download_must_gather" tool.
func (s *Server) DownloadMustGather(ctx context.Context, mustGatherURL string) (string, error) {
	// Create temporary directory for extraction
	destDir, err := os.MkdirTemp("", "must-gather-extract-")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	fmt.Printf("Downloading must-gather from: %s\n", mustGatherURL)

	// Check if URL contains inspect.local - if so, download directory contents directly
	if strings.Contains(mustGatherURL, "inspect.local") {
		fmt.Printf("Detected inspect.local URL - downloading directory contents directly\n")
		
		if err := downloadInspectLocal(mustGatherURL, destDir); err != nil {
			return "", fmt.Errorf("failed to download inspect.local contents: %w", err)
		}
		
		fmt.Printf("Successfully downloaded inspect.local contents to: %s\n", destDir)
	} else {
		// Standard tar file download and extraction
		mustGatherDestPath := filepath.Join(destDir, "must-gather.tar")
		if err := downloadFile(mustGatherURL, mustGatherDestPath); err != nil {
			return "", fmt.Errorf("failed to download must-gather: %w", err)
		}

		fmt.Printf("Extracting must-gather from %s to: %s\n", mustGatherDestPath, destDir)
		if err := extractTarFile(mustGatherDestPath, destDir); err != nil {
			return "", fmt.Errorf("failed to extract must-gather: %w", err)
		}
	}

	// Set omc to use the downloaded directory
	cmd := exec.Command("omc", "use", destDir)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("omc use command failed: %w", err)
	}

	fmt.Printf("Successfully set omc to use directory: %s\n", destDir)
	return "Must-gather download and extraction successful.", nil
}

// downloadInspectLocal downloads inspect.local directory contents
func downloadInspectLocal(url, destDir string) error {
	if strings.HasPrefix(url, "gs://") {
		// Use gsutil for Google Cloud Storage - recursive copy
		cmd := exec.Command("gsutil", "-m", "cp", "-r", url, destDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("gsutil recursive cp command failed: %w", err)
		}
	} else {
		// For HTTP URLs, we can't do recursive download easily
		// This would require a more complex implementation or the user to provide a tar/zip
		return fmt.Errorf("HTTP recursive directory download not supported for inspect.local - please provide gs:// URL")
	}
	return nil
}

// downloadFile downloads a file from either HTTP/HTTPS URL or GS bucket URL
func downloadFile(url, destPath string) error {
	if strings.HasPrefix(url, "gs://") {
		// Use gsutil for Google Cloud Storage
		cmd := exec.Command("gsutil", "-m", "cp", url, destPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("gsutil cp command failed: %w", err)
		}
	} else if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		// Use HTTP client for HTTP/HTTPS URLs
		if err := downloadHTTPFile(url, destPath); err != nil {
			return fmt.Errorf("HTTP download failed: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported URL scheme: %s", url)
	}
	return nil
}

// downloadHTTPFile downloads a file via HTTP/HTTPS
func downloadHTTPFile(url, destPath string) error {
	// Create the destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Minute, // Allow for large file downloads
	}

	// Get the data
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Copy the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// extractTarFile extracts a tar file to a destination directory
func extractTarFile(tarPath, destDir string) error {
	cmd := exec.Command("tar", "-xf", tarPath, "-C", destDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tar extraction failed: %w", err)
	}
	return nil
}

// AnalyzeMustGather implements the "analyze_must_gather" tool.
func (s *Server) AnalyzeMustGather(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var Output []byte
	// run omc get pods
	cmd := exec.Command("omc", "get", "pods", "--all-namespaces")

	/*err := cmd.Run()
	if err != nil {
		// cmd.Run() returns an error if the command fails (non-zero exit code)
		return nil, fmt.Errorf("omc get pods command failed: %w", err)
	}*/

	// get the output of the command
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("omc get pods command failed: %w", err)
	}

	Output = append(Output, output...)

	// get the pod which is not in running state

	cmd = exec.Command("omc", "get", "pods", "--all-namespaces", "--field-selector=status.phase!=Running")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("omc get pods command failed: %w", err)
	}
	Output = append(Output, output...)

	// get the logs of the pod which is not in running state
	cmd = exec.Command("omc", "logs", "--all-namespaces", "--field-selector=status.phase!=Running")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("omc get logs command failed: %w", err)
	}
	Output = append(Output, output...)

	// describe the pod which is not in running state
	cmd = exec.Command("omc", "describe", "pods", "--all-namespaces", "--field-selector=status.phase!=Running")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("omc describe pods command failed: %w", err)
	}
	Output = append(Output, output...)

	// get all cluster operators in yaml format
	cmd = exec.Command("omc", "get", "clusteroperators", "-o", "yaml")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("omc get clusteroperators command failed: %w", err)
	}
	Output = append(Output, output...)

	return NewTextResult(string(Output), err), nil
}

// AnalyzeNodeLogs implements the "analyze_node_logs" tool.
func (s *Server) AnalyzeNodeLogs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var Output []byte
	cmd := exec.Command("omc", "node-logs", "kubelet")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("omc get logs command failed: %w", err)
	}
	Output = append(Output, output...)

	// get the journal logs
	cmd = exec.Command("omc", "node-logs", "journal")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("omc get journal logs command failed: %w", err)
	}
	Output = append(Output, output...)

	// get Networkmanager logs
	cmd = exec.Command("omc", "node-logs", "networkmanager")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("omc get networkmanager logs command failed: %w", err)
	}
	Output = append(Output, output...)

	return NewTextResult(string(Output), err), nil
}
