package mcp

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	//"net/http"
	"os"
	"os/exec"
	"strings"
)

// Register the CLI tools for the OMC
func (s *Server) initOMC() []server.ServerTool {
	return []server.ServerTool{
		{mcp.NewTool("download_must_gather",
			mcp.WithDescription("Downloads and extracts the must-gather.tar file from a given Prow URL."),
			mcp.WithString("prowurl", mcp.Description("The prow job URL"), mcp.Required()),
		), s.DownloadMustGather},
	}
}

// DownloadMustGather implements the "download_must_gather" tool.
func (s *Server) DownloadMustGather(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	prowJobURL := req.Params.Arguments["prowurl"]
	if prowJobURL == "" {
		return nil, fmt.Errorf("missing 'prowurl' in request for download_must_gather")
	}

	//mustGatherURL, _ := utils.GetGatherFolderPath(prowJobURL.(string))
	mustGatherURL := "gs://test-platform-results/logs/periodic-ci-openshift-osde2e-main-nightly-4.18-osd-aws/1937008867888074752/artifacts/osd-aws/gather-must-gather/artifacts/must-gather.tar"
	destDir, err := os.MkdirTemp("", "must-gather-extract-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	//defer os.RemoveAll(destDir) // Ensure temporary directory is removed on function exit

	
	auditURL := "gs://test-platform-results/logs/periodic-ci-openshift-osde2e-main-nightly-4.18-osd-aws/1937008867888074752/artifacts/osd-aws/gather-audit-logs/artifacts/audit-logs.tar"
	fullAuditDir := destDir + "/audit-logs.tar"
	cmd := exec.Command("gsutil", "-m", "cp", "-r", auditURL, destDir)

	err = cmd.Run()
	if err != nil {
		// cmd.Run() returns an error if the command fails (non-zero exit code)
		return nil, fmt.Errorf("wget command failed: %w", err)
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

	cmd = exec.Command("gsutil", "-m", "cp", "-r", mustGatherURL, destDir)

	err = cmd.Run()
	if err != nil {
		// cmd.Run() returns an error if the command fails (non-zero exit code)
		return nil, fmt.Errorf("wget command failed: %w", err)
	}

	cmd = exec.Command("tar", "-xvf", fullDir, "-C", destDir)
	err = cmd.Run()
	if err != nil {
		// cmd.Run() returns an error if the command fails (non-zero exit code)
		return nil, fmt.Errorf("tar command failed: %w", err)
	}

	fmt.Println("Extraction successful.")
	return nil, err
}