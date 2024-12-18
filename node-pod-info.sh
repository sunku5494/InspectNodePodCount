#!/bin/sh

################
#Description:
#This script is to troubleshoot https://issues.redhat.com/browse/OCPBUGS-35070 to gather the pods count attached
#to cluster default network and actively running in a worker node from the control plane, as well as the pods IP
#count in the SDN CNI directory in a worker node.
#
#Example:
#Let's say if the cidr assigned for a node is /24 and we see pods IP count listed in SDN directory is 254(this is
#the max possible IP count we can see for /24) and if the pods count attched to cluster default network is less
#than to it,Then we can say SDN CNI directory have some stale IPS and possibly they failed to get removed from th
#e directory when the correspoding pods have removed.This way we can prove that SDN CNI has some unused IP files
#and that is causing to not finding an IP to allocate from an assigned CIDR to a worker node when a new pod is cr
#eated in it.
#################

# Function to handle cleanup on termination
cleanup() {
	echo "ERROR: Received SIGTERM. Exiting gracefully at $(date '+%Y-%m-%d %H:%M:%S')."
	exit 0
}

# Trap SIGTERM and call the cleanup function
trap cleanup TERM

while :; do
	#collect the count of ipv4 files created under CNI_NETWORK_DIR
	CNI_NETWORK_DIR="/var/lib/cni/networks/openshift-sdn"
	ALLOCATED_IPS=$(find $CNI_NETWORK_DIR -type f -regextype posix-extended -regex '.*/([0-9]{1,3}\.){3}[0-9]{1,3}$' | wc -l)

	#collect the count of pods or containers attached to cluster default network created with in a node
	#filter out the pods that dont have Ip address assigned
	POD_COUNT=$(oc --kubeconfig=/var/lib/kubelet/kubeconfig get pods -A \
		--field-selector spec.nodeName="$(hostname)",status.podIP!=null,status.phase="Running" \
		-o json | jq '[.items[] | select(.metadata.annotations["k8s.v1.cni.cncf.io/network-status"] | fromjson? | any(.default))] | length')

	if [ "$POD_COUNT" -lt "$ALLOCATED_IPS" ]; then #validate if count types are string or integer
		echo "FATAL: $(date '+%Y-%m-%d %H:%M:%S') - POD COUNT is less than the POD IPs Allocated"
	elif [ "$POD_COUNT" -gt "$ALLOCATED_IPS" ]; then
		echo "ERROR: $(date '+%Y-%m-%d %H:%M:%S') - POD COUNT is greater than the POD IPs Allocated"
	else
		echo "INFO: $(date '+%Y-%m-%d %H:%M:%S') - POD and POD IP Count are same"
	fi
	echo "POD COUNT: ${POD_COUNT} and PODIPs COUNT: ${ALLOCATED_IPS}"
	sleep 3600
done
