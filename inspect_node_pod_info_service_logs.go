// inspect_node_pod_info_service_logs.go
// In an Openshift cluster, this script would loop through a list of nodes, check the node-pod-info service logs installed on 
// those nodes for ERROR logs, and output the node name if detects any ERROR logs.
package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Node represents oc node
type Node struct {
	Name string
}

func main() {
	// Get the list of all nodes in the cluster
	nodes, err := getAllNodes()
	if err != nil {
		fmt.Printf("Error fetching nodes: %v\n", err)
		return
	}

	// Inspect logs of node-pod-info service on each node
	myService := "node-pod-info.service"
	for _, node := range nodes {
		errorsFound, err := checkServiceLogsForErrors(node.Name, myService)
		if err != nil {
			fmt.Printf("Error checking logs on node %s: %v\n", node.Name, err)
			continue
		}
		if errorsFound {
			fmt.Printf("ERROR logs are detected for %s\n", node.Name)
		} else {
			fmt.Printf("No Errors found in logs on node: %s\n", node.Name)
		}
	}
}

// getAllNodes fetches the list of all nodes in the cluster using kubectl
func getAllNodes() ([]Node, error) {
	cmd := exec.Command("oc", "get", "nodes", "-o", "name")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run kubectl command: %v, output: %s", err, out.String())
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	nodes := []Node{}
	for _, line := range lines {
		parts := strings.Split(line, "/")
		if len(parts) == 2 {
			nodes = append(nodes, Node{Name: parts[1]})
		}
	}

	return nodes, nil
}

// checkServiceLogsForErrors checks the logs of node-pod-info service on a node for ERRORS
func checkServiceLogsForErrors(nodeName, serviceName string) (bool, error) {
	var out bytes.Buffer

	// Command to execute on the node
	remoteCommand := "chroot /host bash -c 'journalctl -u node-pod-info.service --since \"1 hour ago\" --no-pager'"

	cmd := exec.Command("oc", "debug", fmt.Sprintf("node/%s", nodeName), "--", "bash", "-c", remoteCommand)
	cmd.Stdout = &out
        cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to fetch logs: %v, output: %s", err, out.String())
	}

	logs := out.String()
	// Searching the logs produced by the node-pod-info script for a log containing the string "ERROR"
	if strings.Contains(logs, "ERROR") {
		return true, nil
	}

	return false, nil
}
