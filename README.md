# InspectNodePodCount

## Description:

    InspectNodePodCount helps to determine if the SDN CNI configured for OCP works as expected for managing IP addresses assigned to pods created in the worker nodes for cluster default network by validating the IPv4 files maintained in the SDN CNI directory are inline with the actual pods we see in the worker node


## Steps to Run scripts

Step 1 : Login to the OCP cluster in your Bash terminal

Step 2 : Clone this repository in to your local system directory

Step 3 : Apply the Machine Config Yaml file using the command "oc apply -f mc_deploy_node_pod_info.yaml"

Step 4 : Login to one of the worker node using the command "oc debug node/<nodeName>" and check if node-pod-info.sh file is created in the directory "/usr/local/bin"

Step 5 : Check if node-pod-info service is created with the command "systemctl status node-pod-info.service"

Step 6 : From your local machine and within the previously cloned git, Run the inspect service logs golang script using "go run inspect_node_pod_info_service_logs.go"

Step 7 : If the script finds any Error log with in a worker node,
        i.   Login to the node
        ii.  Collect the output of the command "ls -lrt /var/lib/cni/networks/openshift-SDN"
        iii. Generate must-gather and sos reports. Here are the links help creating must-gather and sos reports
             https://docs.redhat.com/en/documentation/openshift_container_platform/4.17/html/support/gathering-cluster-data#support_gathering_data_gathering-cluster-data 
             https://access.redhat.com/solutions/5065411

## Note

The scripts are written to run on the IPv4 cluster only
