variable "kfc1230_oauth_hostname" {
  description = "The hostname of the Keyfactor instance"
  type        = string
  default     = "int1230-oauth.eastus2.cloudapp.azure.com"
}

variable "kfc1230_oauth_token_url" {
  description = "The hostname of the Keyfactor instance"
  type        = string
  default     = "https://int1230-oauth.eastus2.cloudapp.azure.com:8444/realms/Keyfactor/protocol/openid-connect/token"
}


variable "kfc1230_client_id" {
  description = "The client ID to authenticate with the Keyfactor instance using oauth2 client credentials"
  type        = string

}
variable "kfc1230_client_secret" {
  description = "The client secret to authenticate with the Keyfactor instance using oauth2 client credentials"
  type        = string
}

module "keyfactor_github_test_environment_12_3_0_OAUTH" {
  source = "git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git?ref=main"

  gh_environment_name       = "KFC_12_3_0_OAUTH"
  gh_repo_name              = data.github_repository.repo.name
  keyfactor_hostname        = var.kfc1230_oauth_hostname
  keyfactor_auth_token_url  = var.kfc1230_oauth_token_url
  keyfactor_client_id       = var.kfc1230_client_id
  keyfactor_client_secret   = var.kfc1230_client_secret
  keyfactor_tls_skip_verify = true
  keyfactor_config_file = base64encode(file("${path.module}/int1230_oauth_command_config.json"))
}