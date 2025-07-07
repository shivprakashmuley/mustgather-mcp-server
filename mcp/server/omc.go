package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	//"net/http"
	"os"
	"os/exec"
	//"strings"
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
func (s *Server) DownloadMustGather(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	// get the url from the request
	collectionURL := req.Params.Arguments["url"].(string)

	//mustGatherURL, _ := utils.GetGatherFolderPath(prowJobURL.(string))
	//mustGatherURL := "gs://test-platform-results/logs/periodic-ci-openshift-osde2e-main-nightly-4.18-osd-aws/1937008867888074752/artifacts/osd-aws/gather-must-gather/artifacts/must-gather.tar"
	mustGatherURL := collectionURL + "/gather-must-gather/artifacts/must-gather.tar"
	destDir, err := os.MkdirTemp("", "must-gather-extract-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	//defer os.RemoveAll(destDir) // Ensure temporary directory is removed on function exit

	//auditURL := "gs://test-platform-results/logs/periodic-ci-openshift-osde2e-main-nightly-4.18-osd-aws/1937008867888074752/artifacts/osd-aws/gather-audit-logs/artifacts/audit-logs.tar"
	auditURL := collectionURL + "/gather-audit-logs/artifacts/audit-logs.tar"
	fullAuditDir := destDir + "/audit-logs.tar"
	cmd := exec.Command("gsutil", "-m", "cp", auditURL, destDir)

	err = cmd.Run()
	if err != nil {
		// cmd.Run() returns an error if the command fails (non-zero exit code)
		return nil, fmt.Errorf("gsutil cp command failed: %w", err)
	}
	fmt.Println("Download successful. Starting extraction and analysis...")

	fmt.Printf("Extracting contents from %s to: %s \n", fullAuditDir, destDir)
	cmd = exec.Command("tar", "-xvf", fullAuditDir, "-C", destDir)
	err = cmd.Run()
	if err != nil {
		// cmd.Run() returns an error if the command fails (non-zero exit code)
		return nil, fmt.Errorf("tar command failed: %w", err)
	}
	// 1. Download the must-gather.tar.gz file using an HTTP GET request.
	fmt.Printf("MustGatherURL: %s \n", mustGatherURL)
	fullDir := destDir + "/must-gather.tar"

	cmd = exec.Command("gsutil", "-m", "cp", mustGatherURL, destDir)

	err = cmd.Run()
	if err != nil {
		// cmd.Run() returns an error if the command fails (non-zero exit code)
		return nil, fmt.Errorf("gsutil cp command failed: %w", err)
	}

	cmd = exec.Command("tar", "-xvf", fullDir, "-C", destDir)
	err = cmd.Run()
	if err != nil {
		// cmd.Run() returns an error if the command fails (non-zero exit code)
		return nil, fmt.Errorf("tar command failed: %w", err)
	}

	// collect inspect.local
	//inspectLocalURL := "gs://test-platform-results/logs/periodic-ci-openshift-release-master-nightly-4.20-e2e-aws-ovn-single-node-techpreview-serial/1939310646927560704/artifacts/e2e-aws-ovn-single-node-techpreview-serial/gather-must-gather/artifacts/must-gather/inspect.local.4821810590815119360/"
	//inspectLocalURL := mustGatherURL
	/*cmd := exec.Command("gsutil", "-m", "cp", "-r", mustGatherURL, destDir)

	err = cmd.Run()
	if err != nil {
		// cmd.Run() returns an error if the command fails (non-zero exit code)
		return nil, fmt.Errorf("gsutil cp command failed: %w", err)
	}
	fmt.Println("Extraction successful.")
	*/
	// set "omc use" to destDir
	//cmd := exec.Command("omc", "use", "gs://test-platform-results/logs/periodic-ci-openshift-release-master-okd-scos-4.20-upgrade-from-okd-scos-4.19-e2e-aws-ovn-upgrade/1937465859354136576/artifacts/e2e-aws-ovn-upgrade/gather-must-gather/artifacts/must-gather")
	cmd = exec.Command("omc", "use", destDir)
	err = cmd.Run()
	if err != nil {
		// cmd.Run() returns an error if the command fails (non-zero exit code)
		return nil, fmt.Errorf("omc use command failed: %w", err)
	}
	fmt.Println("omc use command successful.", destDir)
	//return nil, err
	return NewTextResult("Extraction successful.", err), nil
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
