package server

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	//"net/http"
	"os"
	"os/exec"
	//"strings"
)

// Register the CLI tools for the OMC
func (s *Server) initOMC() []server.ServerTool {
	return []server.ServerTool{
		{mcp.NewTool("download_must_gather",
			mcp.WithDescription("Downloads and extracts the must-gather.tar file from a given Prow URL."),
			mcp.WithString("url", mcp.Description("URL of any log collection folder, it could be a prow job URL or artificats url or gs bucket url or inspect.local url"), mcp.Required()),
		), s.DownloadMustGather},
		{mcp.NewTool("analyze_must_gather",
			mcp.WithDescription("Analyzes the downloaded and extracted must-gather.tar file which is set in omc use. \n" +
			"This will run omc get pods command. Get the status of all pods. \n" +
			"Check how many of them are in CrashLoopBackOff, Pending, Running, Terminating, Unknown, Completed, Failed, Unknown state. \n" +
			"Gets the logs of the pod which is not in running state. Describes the pod which is not in running state. \n" +
			"Returns the summary of the analysis."),
		), s.AnalyzeMustGather},
		{mcp.NewTool("analyze_node_logs",
			mcp.WithDescription("Analyzes the node logs in case from above analysis, if any node is not in running state. \n" +
			"Gets and analyzes system level logs like kubelet logs, journal logs, bootkube logs. \n" +
			"Returns the summary of the analysis."),
		), s.AnalyzeNodeLogs},
	}
}

// DownloadMustGather implements the "download_must_gather" tool.
func (s *Server) DownloadMustGather(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	// get the url from the request
	mgArgs := req.GetArguments()
	if mgArgs["url"] == nil {
		return nil, fmt.Errorf("missing 'url' in request for download_must_gather")
	}
	collectionURL := mgArgs["url"].(string)
	
	

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