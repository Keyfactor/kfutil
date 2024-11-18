variable "kfc1230c_ad_hostname" {
  description = "The hostname of the Keyfactor instance"
  type        = string
  default     = "int1230c-ad.eastus2.cloudapp.azure.com"
}

module "keyfactor_github_test_environment_12_3_0_AD_CLEAN" {
  source                    = "git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git?ref=main"
  gh_environment_name       = "KFC_12_3_0_AD_CLEAN"
  gh_repo_name              = data.github_repository.repo.name
  keyfactor_hostname        = var.kfc1230c_ad_hostname
  keyfactor_username        = var.keyfactor_username_AD
  keyfactor_password        = var.keyfactor_password_AD
  keyfactor_tls_skip_verify = true
  keyfactor_config_file = base64encode(file("${path.module}/command_config.json"))
}