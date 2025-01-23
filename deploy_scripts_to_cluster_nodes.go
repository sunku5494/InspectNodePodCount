// deploy_scripts_to_cluster_nodes.go
// This file can copy the files defined, start service defined and it can also perform the cleanup of services and files copied in all the cluster nodes.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Node represents oc worker node
type Node struct {
	Name string
}

// executeCommand executes a shell command and returns the output or an error.
func executeCommand(cmd string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	command := exec.Command(cmd, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	if err != nil {
		return "", fmt.Errorf("error: %v, stderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

// runCmdsOnNodes runs a list of commands on a specific node using `oc debug`.
func runCmdsOnNodes(nodeName string, cmds []string) error {
	combinedCommands := strings.Join(cmds, " && ")
	ocDebugCommand := fmt.Sprintf("chroot /host /bin/bash -c '%s'", combinedCommands)

	output, err := executeCommand("oc", "debug", "node/"+nodeName, "--", "/bin/bash", "-c", ocDebugCommand)
	if err != nil {
		return fmt.Errorf("failed to run commands on node %s: %v", nodeName, err)
	}

	fmt.Printf("Commands executed successfully on node %s: %s\n", nodeName, output)
	return nil
}

// getMCPDaemonPods fetches the list of MCD pods in all the nodes of a cluster
func getMCPDaemonPods() []string {
	// Get the list of pods with the label
	cmdGetPods := exec.Command("oc", "get", "pod", "-n", "openshift-machine-config-operator", "-l", "k8s-app=machine-config-daemon", "-o", "name")
	var out bytes.Buffer
	cmdGetPods.Stdout = &out
	cmdGetPods.Stderr = os.Stderr

	if err := cmdGetPods.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error getting pods: %v\n", err)
		os.Exit(1)
	}
	// Split the output into pod names
	pods := strings.Split(strings.TrimSpace(out.String()), "\n")
	return pods
}

// getAllNodes fetches the list of all nodes in the cluster using oc command
func getAllNodes() ([]Node, error) {
	cmd := exec.Command("oc", "get", "nodes", "-o", "name")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run oc command: %v, output: %s", err, out.String())
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

// copyFilesToAllNodes copies files to respective paths in all the nodes
func copyFilesToAllNodes() {
	pods := getMCPDaemonPods()

	var wg sync.WaitGroup
	// Iterate over the pods and copy the file
	for _, pod := range pods {
		wg.Add(1)
		// Remove the "pod/" prefix from the pod name
		go func(pod string) {
			defer wg.Done()
			podName := strings.TrimPrefix(pod, "pod/")

			filesToCopy := []struct {
				source      string
				destination string
			}{
				{
					source:      "node-pod-info.sh",
					destination: fmt.Sprintf("openshift-machine-config-operator/%s:/rootfs/%s", podName, "usr/local/bin"),
				},
				{
					source:      "node-pod-info.service",
					destination: fmt.Sprintf("openshift-machine-config-operator/%s:/rootfs/%s", podName, "etc/systemd/system/"),
				},
			}

			for _, file := range filesToCopy {
				fmt.Printf("Copying %s into POD %s:%s\n", file.source, podName, file.destination)
				cmdCopy := exec.Command("oc", "cp", "-c", "machine-config-daemon", file.source, file.destination)
				cmdCopy.Stdout = os.Stdout
				cmdCopy.Stderr = os.Stderr

				if err := cmdCopy.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Error copying file to pod %s: %v\n", podName, err)
				}
				//fmt.Printf("Done copying file %s to pod %s\n", file.source, podName)
			}
		}(pod)
	}
	wg.Wait()
	//fmt.Println("Came out of wait!!!")
}

// stopServiceAndRmFilesFromAllNodes runs a list of commands in all the nodes to stop a service and delete all the files related to the service
func stopServiceAndRmFilesFromAllNodes() {
	nodes, err := getAllNodes()
	if err != nil {
		fmt.Printf("Error fetching nodes: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	commands := []string{
		"systemctl stop node-pod-info.service",
		"systemctl disable node-pod-info.service",
		"rm /usr/local/bin/node-pod-info.sh",
		"rm /etc/systemd/system/node-pod-info.service",
		"systemctl daemon-reload",
	}

	for _, node := range nodes {
		wg.Add(1)
		go func(node Node) {
			defer wg.Done()
			err := runCmdsOnNodes(node.Name, commands)
			if err != nil {
				fmt.Errorf("Error executing commands on node %s: %v", node.Name, err)
			}
		}(node)
	}
	wg.Wait()
}

// startServiceInAllNodes runs a list of commands to start a service in all the nodes of a cluster
func startServiceInAllNodes() {
	fmt.Println("Starting Service in all nodes")
	nodes, err := getAllNodes()

	if err != nil {
		fmt.Printf("Error fetching nodes: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	commands := []string{
		"chmod +x /usr/local/bin/node-pod-info.sh",
		"systemctl daemon-reload",
		"systemctl start node-pod-info.service",
		"systemctl enable node-pod-info.service",
	}

	for _, node := range nodes {
		wg.Add(1)
		go func(node Node) {
			defer wg.Done()
			err := runCmdsOnNodes(node.Name, commands)
			if err != nil {
				fmt.Errorf("Error executing commands on node %s: %v", node.Name, err)
			}
		}(node)
	}
	wg.Wait()
	fmt.Println("Done Starting Service in all nodes!")
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("No arguments provided!")
		return
	}

	action := os.Args[1]

	switch {
	case action == "copy":
		copyFilesToAllNodes()
		startServiceInAllNodes()
	case action == "remove":
		stopServiceAndRmFilesFromAllNodes()
	default:
		fmt.Println("Supported args: 1. copy and 2. remove")
	}
	return
}
