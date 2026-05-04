module "keyfactor_github_test_environment_ses_2541" {
  source = "git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git?ref=main"

  gh_environment_name       = "SES_2541"
  gh_repo_name              = data.github_repository.repo.name
  keyfactor_hostname        = var.ses_2541_hostname
  keyfactor_auth_token_url  = var.ses_2541_auth_token_url
  keyfactor_client_id       = var.ses_2541_client_id
  keyfactor_client_secret   = var.ses_2541_client_secret
  keyfactor_tls_skip_verify = true
  keyfactor_config_file     = base64encode(file("${path.module}/ses2541_command_config.json"))
}
