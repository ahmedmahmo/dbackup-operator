#!/bin/bash

#Copyright 2021.
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.

set -e
hostname="master"
user="ubuntu"

#
# Logging
#
mkdir -p kubelog
touch /kubelog/exec.log
LOG_FILE="/kubelog/exec.log"

exec 3>&1 1>>${LOG_FILE} 2>&1

#
# Update & Upgrade
#

echo "starting..."
sudo apt update -y && sudo apt upgrade -y
echo "done upgrading..."

#
# Install  all required dependencies
#
echo "installing base dependencies..."
sudo apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    wget \
    unzip \
    jq \
    lsb-release -y
echo "base dependencies installation is done."

#
# install aws cli
#
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

mkdir -p /home/$user/.aws
cat <<EOF > /home/$user/.aws/config
[default]
region = eu-central-1
EOF

cat <<EOF > /home/$user/.aws/credentials
[default]
EOF

#
# Set hostname
#
sudo hostname $hostname

#
# Add GPG Key
#
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF
sudo apt-get update -y

#
# Disable swapping
#
swapoff -a
sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab

#
# Install Docker
#
sudo apt install docker.io -y
sudo cat <<EOF > /etc/docker/daemon.json
  {
  "exec-opts": ["native.cgroupdriver=systemd"]
  }
EOF
#
# Use docker without sudo
#
sudo usermod -aG docker ubuntu
sudo systemctl restart docker
sudo systemctl enable docker.service
sudo apt-get update

#
# Intall the following packges
# Kubelet
# kubectl
# kubeadm
#
sudo apt-get install -y kubeadm kubectl
sudo systemctl daemon-reload
echo "Machine Setup Completed!"

#
# set up master node
#
public_ip=$(dig +short myip.opendns.com @resolver1.opendns.com)
sudo kubeadm init --control-plane-endpoint=$public_ip  --pod-network-cidr=10.244.0.0/16
export KUBECONFIG=/etc/kubernetes/admin.conf

#
# install CNI plugin
#
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml

#
# check all running
#
function is_all_running {
  for pod in $( kubectl get pods --all-namespaces| awk 'NR>1{print $4}'); do
    if [[ "$pod" != "Running" ]]; then
      echo "$pod is not running"
      sleep 20
    fi
  done
  echo "all running"
}
is_all_running

#
# helm
#
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh

# Allow kubernetes to schedule over master
kubectl taint node $(hostname) node-role.kubernetes.io/master:NoSchedule-


#
# kubed
#
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm install kubed appscode/kubed \
  --version v0.12.0 \
  --namespace kube-system
#
# join command
#
sudo echo "$(sudo kubeadm token create --print-join-command)" > /natix/join.txt

#
# Message
#
echo "Kubernetes is all set up!"