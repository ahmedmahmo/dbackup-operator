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

# AWS Virtual Private Cloud
resource "aws_vpc" "main" {
  cidr_block       = "192.168.0.0/16"
  instance_tenancy = "default"
  enable_dns_hostnames = "true"
  tags = {
    Name = "tf-vpc"
  }
}

# Create a private subnet with CIDR 24
resource "aws_subnet" "private_subnet" {
  vpc_id     = aws_key_pair.main.id
  cidr_block = "192.168.0.0/24"
  availability_zone = "${var.region}a"

  tags = {
    Name = "tf-private-subnet-1"
  }
}

# Create public subnet with CIDR 24
resource "aws_subnet" "public_subnet" {
  vpc_id     = aws_key_pair.main.id
  cidr_block = "192.168.1.0/24"
  availability_zone = "${var.region}b"
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-public-subnet-1"
  }
}

// Create public facing internet gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_key_pair.main.id

  tags = {
    Name = "tf-internet-gateway"
  }
}

# Create a routing table for Internet gateway
resource "aws_route_table" "main" {
  vpc_id = aws_key_pair.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "tf-route-table"
  }
}

# Associating Public subnet to this route table
resource "aws_route_table_association" "associate" {
  subnet_id      = aws_subnet.public_subnet.id
  route_table_id = aws_route_table.main.id
}

# Kubernetes security group
resource "aws_security_group" "kube" {
  name        = "tf-allow-kube"
  description = "Allow SSh"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "Kube API server"
    protocol    = "tcp"
    from_port   = 6443
    to_port     = 6443
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "Ingress"
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "platform agent"
    protocol    = "tcp"
    from_port   = 8090
    to_port     = 8090
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "operator node"
    protocol    = "tcp"
    from_port   = 8091
    to_port     = 8091
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "flannel"
    protocol    = "udp"
    from_port   = 8472
    to_port     = 8472
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "kubelet"
    protocol    = "tcp"
    from_port   = 10250
    to_port     = 10250
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "kubelet with no Auth"
    protocol    = "tcp"
    from_port   = 10255
    to_port     = 10255
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "tf-allow-kube",
    Description = "tf-kubernetes-security-group"
  }
}