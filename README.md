# Must Gather MCP Server

An MCP (Model Context Protocol) server that provides a wrapper over `omc` commands to allow LLMs and MCP clients in analyzing any [OpenShift mustgather](https://github.com/openshift/must-gather) bundles. OMC is like a oc or kubectl cli tool that reads kubernetes objects, events and other resources like logs from a cluster mustgather. 

## Features

The server provides the following tools:

### 1. mustgather_get
Get Kubernetes resources using `omc get` command.

**Parameters:**
- `kind` (required): Resource type (pods, nodes, services, deployments, configmaps, secrets, namespaces, ingress, pvc, pv)
- `all_namespaces` (optional): Get resources from all namespaces (-A flag)
- `namespace` (optional): Specific namespace (-n flag)
- `output` (optional): Output format (wide, yaml, json)

### 2. mustgather_describe
Describe pods or nodes using `omc describe` command.

**Parameters:**
- `kind` (required): Resource type (pods or nodes only)
- `all_namespaces` (optional): Describe resources from all namespaces (-A flag)
- `namespace` (optional): Specific namespace (-n flag)
- `output` (optional): Output format (wide, yaml)

### 3. mustgather_logs
Get logs from a specific pod and container.

**Parameters:**
- `pod_name` (required): Name of the pod
- `namespace` (required): Namespace of the pod
- `container` (required): Container name within the pod

### 4. mustgather_events
Get cluster events using `omc events` command.

**Parameters:**
- `all_namespaces` (optional): Get events from all namespaces (-A flag)
- `namespace` (optional): Specific namespace (-n flag)
- `for` (optional): Filter events for a specific resource (--for flag)
- `output` (optional): Output format (yaml, name)

### 5. mustgather_node_logs
Get node logs for a specific journalctl service.

**Parameters:**
- `service_name` (required): Journalctl service name

### 6. mustgather_haproxy_backends
Get HAProxy backends information.

**Parameters:** None

### 7. mustgather_etcd_health
Check etcd cluster health.

**Parameters:** None

### 8. mustgather_etcd_status
Get etcd cluster status.

**Parameters:** None

### 9. mustgather_projects
List available projects (namespaces) in the OpenShift cluster.

**Parameters:** None

### 10. mustgather_use
Switch to a different mustgather directory.

**Parameters:**
- `path` (required): Path to use for reading the mustgather bundle.

## Prerequisites

- go 1.23+
- A mustgather bundle from any cluster

## Installation

Clone or install the dependencies, build and run the MCP server.

```bash
git clone https://github.com/shivprakashmuley/mustgather-mcp-server.git
cd mustgather-mcp-server

cd omc-cli
go install . # Installs the omc cli to PATH (required)

cd ..
go build -o mustgather-mcp-server

./mustgather-mcp-server # runs the MCP in stdio mode
```

## Integrate with an MCP client

1. Claude desktop

Set your `claude_desktop_config.json` as follows:

```
{
  "mcpServers": {
    "mustgather": {
      "command": "/path/to/mustgather-mcp-server",
    }
  }
}
```

2. Goose

Run `./mustgather-mcp-server --sse-port 8911` and set the following in your `~/.config/goose/config.yaml`.

```
extensions:
  mustgather:
    bundled: null
    description: ''
    enabled: false
    env_keys: []
    envs: {}
    name: mustgather
    timeout: 300
    type: sse
    uri: http://localhost:8911/sse
```

3. Gemini CLI

Set the following in your ~/.gemini/settings.json

```
{
  "mcpServers": {
    "mustgather": {
      "command": "/path/to/mustgather-mcp-server",
      "cwd": "./",
      "timeout": 900,
      "trust": true
    }
  }
}
```

## Demos

1. https://asciinema.org/a/mvb4GGUfaAuUhMggBLrsmBBhe
2. https://asciinema.org/a/xqbgtCi3QIqfAmioeTZNurJ0P
