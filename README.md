# InspectNodePodCount

## Description:

    InspectNodePodCount helps to determine if the SDN CNI configured for OCP works as expected for managing IP addresses assigned to pods created in the worker nodes for cluster default network by validating the IPv4 files maintained in the SDN CNI directory are inline with the actual pods we see in the worker node

## Steps to Run scripts

Step 1 : Clone this repository in to your local system directory

Step 2 : Login to the OCP cluster in your Bash terminal 

Step 3 : Run below comands to deploy the scripts to all nodes of a cluster
         i. Build the go script using cmd "go build deploy_scripts_to_cluster_nodes.go"
        ii. Deploy scripts using cmd "./deploy_scripts_to_cluster_nodes.go "copy""

Step 4 : Login to one of the node using the command "oc debug node/<nodeName>" and check if node-pod-info.sh file is copied to the directory "/usr/local/bin"          and also check if "node-pod-info.service" is copied to "/etc/systemd/system" directory

Step 5 : Check if node-pod-info service is created and running with the command "systemctl status node-pod-info.service"

Step 6 : From your local machine and within the previously cloned git path, Run the inspect service logs golang script using "go run inspect_node_pod_info_service_logs.go" (script would outputs the node names where ERROR logs have detected)

Step 7 : Execute the below commands
        i.  Login to the node using "oc debug node/<node-name>"
       ii.  Run a command "journalctl -u node-pod-info.service" to check logs of deployed service       
      iii.  Zip the whole directory "/var/lib/cni/" and ensure Zip file includes files and directories, from all subdirectories within "/var/lib/cni/"
       iv.  Generate must-gather and sos reports. Here are the links help creating must-gather and sos reports
             https://docs.redhat.com/en/documentation/openshift_container_platform/4.17/html/support/gathering-cluster-data#support_gathering_data_gathering-cluster-data 
             https://access.redhat.com/solutions/5065411

Step 8: Once we are done with the tasks. we can clean up all the nodes with the below command
        "./deploy_scripts_to_cluster_nodes.go "remove"

## Note

The scripts are written to run on the IPv4 cluster only
