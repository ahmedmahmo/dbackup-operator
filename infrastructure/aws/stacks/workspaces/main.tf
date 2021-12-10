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

# --- set provide to authenticate terraform to perform operations on behalf ---
provider "tfe" {
  token = var.org_token
}

# --- Set local variables ---
locals {
  organization = "ahmedmahmoud"
  workspaces = ["thesis"]
}
# --- Configure Terraform ---
terraform {
  backend "remote" {
    organization = "ahmedmahmoud"
    workspaces {
      name = "workspaces"
    }
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
    tfe = {
      source = "hashicorp/tfe"
      version = "0.25.3"
    }
  }
}

# --- Create workspaces using the local module workspace > ${root}/modules/workspace ---
resource "tfe_workspace" "main" {
  for_each = toset(local.workspaces)
  name = each.key
  organization = local.organization
  working_directory = "/stacks/${each.key}"
  auto_apply = false
  trigger_prefixes = ["/modules"]
}

resource "tfe_variable" "access_key" {
  key          = "access_key"
  value        = var.access_key
  category     = "terraform"
  for_each = toset(local.workspaces)
  workspace_id = tfe_workspace.main[each.key].id
  sensitive = false
}
resource "tfe_variable" "secret_key" {
  key          = "secret_key"
  value        = var.secret_key
  category     = "terraform"
  for_each = toset(local.workspaces)
  workspace_id = tfe_workspace.main[each.key].id
  sensitive = true
}