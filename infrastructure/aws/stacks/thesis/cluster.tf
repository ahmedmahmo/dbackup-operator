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


resource "aws_instance" "kube_master" {
  ami = local.ami
  key_name = "ahmed-private-key"
  instance_type = local.master
  vpc_security_group_ids = [aws_security_group.kube.id]
  subnet_id = aws_subnet.public_subnet.id
  user_data = file("${path.module}/scripts/master.sh")
  tags = {
    Name = "tf-master-node"
  }
}

resource "aws_instance" "kube_nodes" {
  ami = local.ami
  key_name = "ahmed-private-key"
  vpc_security_group_ids = [aws_security_group.kube.id]
  subnet_id = aws_subnet.public_subnet.id
  for_each = local.workers
  instance_type = each.value
  user_data = file("${path.module}/scripts/node.sh")
  tags = {
    Name = "tf-${each.key}"
  }
}

resource "aws_eip" "master" {
  vpc  = true
}

resource "aws_eip_association" "master" {
  allocation_id = aws_eip.kube_master.id
  instance_id   = aws_instance.kube_master.id
}
