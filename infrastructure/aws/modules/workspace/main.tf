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
variable "name" {}
variable "organization" {}
variable "access_key" {}
variable "secret_key" {}

resource "tfe_workspace" "main" {
  name = var.name
  organization = var.organization
  working_directory = "/stacks/${var.name}"
  auto_apply = false
  trigger_prefixes = ["/modules"]
}
resource "tfe_variable" "access_key" {
  key          = "azure_client_id"
  value        = var.access_key
  category     = "terraform"
  workspace_id = tfe_workspace.main.id
  sensitive = false
}
resource "tfe_variable" "secret_key" {
  key          = "azure_client_secret"
  value        = var.secret_key
  category     = "terraform"
  workspace_id = tfe_workspace.main.id
  sensitive = true
}