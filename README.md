#InspectNodePodCount

## Description:

    InspectNodePodCount helps to determine if the SDN CNI configured in the OCP works as expected in allocating the IPv4 addresses to the pods created in the worker nodes for cluster default network by validating the IPv4 files maintained in the SDN CNI directory are inline with the actual pods we see in the worker node


## Steps to Run scripts

Step 1 : Login to the OCP cluster in your Bash terminal

Step 2 : Clone this repository in to your local system directory

Step 3 : Apply the Machine Config Yaml file using the command "oc apply -f mc_deploy_node_pod_info.yaml"

Step 4 : Login to one of the worker node using the command "oc debug node/<nodeName>" and check if node-pod-info.sh file is created in the directory "/usr/local/bin"

Step 5 : Then Check if node-pod-info service is created with the command "systemctl status node-pod-info.service"

Step 6 : Come out of the node and run the inspect service logs golang script using "go run inspect_node_pod_info_service_logs.go"

Step 7 : If the script finds any Error log with in a worker node,
        i.   Login to the node
        ii.  Collect the output of the command "ls -lrt /var/lib/cni/networks/openshift-SDN"
        iii. Execute 'omc' commands and collect the logs  
