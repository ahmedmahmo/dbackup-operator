/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

resource "aws_key_pair" "main" {
  key_name_prefix = "${local.cluster_name}-"
  public_key      = file("../../id_rsa.pub")
}

resource "random_string" "token_id" {
  length  = 6
  special = false
  upper   = false
}

resource "random_string" "token_secret" {
  length  = 16
  special = false
  upper   = false
}

locals {
  token = "${random_string.token_id.result}.${random_string.token_secret.result}"
}

resource "aws_instance" "kube_master" {
  ami = local.ami
  key_name = aws_key_pair.main.key_name
  instance_type = local.master
  vpc_security_group_ids = [aws_security_group.kube.id]
  subnet_id = aws_subnet.public_subnet.id
  user_data = <<-EOD
  #!/bin/bash

  set -e
  hostname="master"

  # Logging
  mkdir -p kubelog
  touch /kubelog/exec.log
  LOG_FILE="/kubelog/exec.log"
  exec 3>&1 1>>$LOG_FILE 2>&1

  # System update
  sudo apt update -y && sudo apt upgrade -y

  # Base dependencies
  sudo apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    wget \
    unzip \
    jq \
    lsb-release
  
  # AWS CLI
  curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
  unzip awscliv2.zip
  sudo ./aws/install

  # Set hostname
  sudo hostname $hostname

  # Kubernets GPG
  curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
  cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list
  deb https://apt.kubernetes.io/ kubernetes-xenial main
  EOF
  sudo apt-get update -y

  # disable swap
  swapoff -a
  sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab

  # Install Docker
  sudo apt install docker.io -y
  sudo cat <<EOF > /etc/docker/daemon.json
    {
    "exec-opts": ["native.cgroupdriver=systemd"]
    }
  EOF

  # Use docker without sudo
  sudo usermod -aG docker ubuntu
  sudo systemctl restart docker
  sudo systemctl enable docker.service
  sudo apt-get update

  # install kubeadm and kubernetes client
  sudo apt-get install -y kubeadm kubectl
  sudo systemctl daemon-reload
  echo "Machine Setup Completed!"

  # Run kubeadm
  kubeadm init \
    --token "${local.token}" \
    --token-ttl 15m \
    --apiserver-cert-extra-sans "${aws_eip.master.public_ip}" \
    --pod-network-cidr "10.244.0.0/16" \
    --node-name master

  # Install Container Network Interface CNI
  export KUBECONFIG=/etc/kubernetes/admin.conf
  kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml

  # Allow Schedule on master node
  kubectl taint node $(hostname) node-role.kubernetes.io/master:NoSchedule-

  # Install helm
  curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
  chmod 700 get_helm.sh
  ./get_helm.sh

  # Install kubed
  helm repo add appscode https://charts.appscode.com/stable/
  helm repo update
  helm install kubed appscode/kubed \
    --version v0.12.0 \
    --namespace kube-system

  # join command
  sudo echo "$(sudo kubeadm token create --print-join-command)" > /home/ubuntu/join.txt

  # Setup kube client for ubuntu user
  mkdir -p ubuntu/.kube
  sudo cp -i /etc/kubernetes/admin.conf ubuntu/.kube/config
  sudo chown 1000:1000 ubuntu/.kube/config

  touch /home/ubuntu/done
  EOD
  tags = {
    Name = "tf-master-node"
  }
}

resource "aws_instance" "kube_nodes" {
  ami = local.ami
  key_name = aws_key_pair.main.key_name
  vpc_security_group_ids = [aws_security_group.kube.id]
  subnet_id = aws_subnet.public_subnet.id
  for_each = local.workers
  instance_type = each.value
  user_data = <<-EOD
  #!/bin/bash

  set -e
  hostname=worker-${each.key}

  # Logging
  mkdir -p kubelog
  touch /kubelog/exec.log
  LOG_FILE="/kubelog/exec.log"
  exec 3>&1 1>>$LOG_FILE 2>&1

  # System update
  sudo apt update -y && sudo apt upgrade -y

  # Base dependencies
  sudo apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    wget \
    unzip \
    jq \
    lsb-release
  
  # AWS CLI
  curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
  unzip awscliv2.zip
  sudo ./aws/install

  # Set hostname
  sudo hostname $hostname

  # Kubernets GPG
  curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
  cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list
  deb https://apt.kubernetes.io/ kubernetes-xenial main
  EOF
  sudo apt-get update -y

  # disable swap
  swapoff -a
  sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab

  # Install Docker
  sudo apt install docker.io -y
  sudo cat <<EOF > /etc/docker/daemon.json
    {
    "exec-opts": ["native.cgroupdriver=systemd"]
    }
  EOF

  # Use docker without sudo
  sudo usermod -aG docker ubuntu
  sudo systemctl restart docker
  sudo systemctl enable docker.service
  sudo apt-get update

  # install kubeadm and kubernetes client
  sudo apt-get install -y kubeadm kubectl
  sudo systemctl daemon-reload
  echo "Machine Setup Completed!"

  kubeadm join ${aws_instance.kube_master.private_ip}:6443 \
    --token ${local.token} \
    --discovery-token-unsafe-skip-ca-verification \
    --node-name worker-${each.key}
  EOD
  tags = {
    Name = "tf-${each.key}"
  }
}

resource "aws_eip" "master" {
  vpc  = true
}

resource "aws_eip_association" "master" {
  allocation_id = aws_eip.master.id
  instance_id   = aws_instance.kube_master.id
}