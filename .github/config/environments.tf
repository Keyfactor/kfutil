module "keyfactor_github_test_environment_10_5_0" {
  source = "git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git?ref=main"

  gh_environment_name = "KFC_10_5_0"
  gh_repo_name        = data.github_repository.repo.name
  keyfactor_hostname  = var.keyfactor_hostname_10_5_0
  keyfactor_username  = var.keyfactor_username_AD
  keyfactor_password  = var.keyfactor_password_AD
  keyfactor_config_file = base64encode(file("${path.module}/command_config.json"))
}

module "keyfactor_github_test_environment_10_5_0_CLEAN" {
  source = "git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git?ref=main"

  gh_environment_name = "KFC_10_5_0_CLEAN"
  gh_repo_name        = data.github_repository.repo.name
  keyfactor_hostname = var.keyfactor_hostname_10_5_0_CLEAN
  keyfactor_username  = var.keyfactor_username_AD
  keyfactor_password  = var.keyfactor_password_AD
  keyfactor_config_file = base64encode(file("${path.module}/command_config.json"))
}

module "keyfactor_github_test_environment_11_5_0" {
  source = "git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git?ref=main"

  gh_environment_name = "KFC_11_5_0"
  gh_repo_name        = data.github_repository.repo.name
  keyfactor_hostname  = var.keyfactor_hostname_11_5_0
  keyfactor_username  = var.keyfactor_username_AD
  keyfactor_password  = var.keyfactor_password_AD
  keyfactor_config_file = base64encode(file("${path.module}/command_config.json"))
}

module "keyfactor_github_test_environment_11_5_0_CLEAN" {
  source = "git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git?ref=main"

  gh_environment_name = "KFC_11_5_0_CLEAN"
  gh_repo_name        = data.github_repository.repo.name
  keyfactor_hostname = var.keyfactor_hostname_11_5_0_CLEAN
  keyfactor_username  = var.keyfactor_username_AD
  keyfactor_password  = var.keyfactor_password_AD
  keyfactor_config_file = base64encode(file("${path.module}/command_config.json"))
}

module "keyfactor_github_test_environment_11_5_0_OAUTH" {
  source = "git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git?ref=main"

  gh_environment_name       = "KFC_11_5_0_OAUTH"
  gh_repo_name              = data.github_repository.repo.name
  keyfactor_hostname        = var.keyfactor_hostname_11_5_0_OAUTH
  keyfactor_auth_token_url  = var.keyfactor_auth_token_url
  keyfactor_client_id       = var.keyfactor_client_id
  keyfactor_client_secret   = var.keyfactor_client_secret
  keyfactor_tls_skip_verify = true
  keyfactor_config_file = base64encode(file("${path.module}/command_config.json"))
}

module "keyfactor_github_test_environment_11_5_0_OAUTH_CLEAN" {
  source = "git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git?ref=main"

  gh_environment_name       = "KFC_11_5_0_OAUTH_CLEAN"
  gh_repo_name              = data.github_repository.repo.name
  keyfactor_hostname        = var.keyfactor_hostname_11_5_0_OAUTH_CLEAN
  keyfactor_auth_token_url  = var.keyfactor_auth_token_url
  keyfactor_client_id       = var.keyfactor_client_id
  keyfactor_client_secret   = var.keyfactor_client_secret
  keyfactor_tls_skip_verify = true
  keyfactor_config_file = base64encode(file("${path.module}/command_config.json"))
}

module "keyfactor_github_test_environment_12_3_0_AD" {
  source                    = "git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git?ref=main"
  gh_environment_name       = "KFC_12_3_0_AD"
  gh_repo_name              = data.github_repository.repo.name
  keyfactor_hostname        = var.keyfactor_hostname_12_3_0
  keyfactor_username        = var.keyfactor_username_AD
  keyfactor_password        = var.keyfactor_password_AD
  keyfactor_tls_skip_verify = true
  keyfactor_config_file = base64encode(file("${path.module}/command_config.json"))
}



