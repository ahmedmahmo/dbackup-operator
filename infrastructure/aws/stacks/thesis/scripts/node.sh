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
hostname="node"
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
    lsb-release -y
echo "base dependencies installation is done."

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
