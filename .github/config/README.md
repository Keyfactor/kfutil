# GitHub Test Environment Setup

This code sets up GitHub environments for testing against the SES 25.4.1 Keyfactor Command lab.

## Requirements

1. Terraform >= 1.0
2. GitHub Provider >= 6.2
3. SES 25.4.1 Keyfactor Command lab access
4. OAuth credentials for authenticating to the Keyfactor Command instance
5. A GitHub token with access and permissions to the repository where the environments will be created

## Adding a new environment

Modify the `environments.tf` file to include the new environment module. The module should be named appropriately.
Example:

### SES 25.4.1 Environment

```hcl
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
```

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0 |
| <a name="requirement_github"></a> [github](#requirement\_github) | >=6.2 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_github"></a> [github](#provider\_github) | 6.3.1 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_keyfactor_github_test_environment_ses_2541"></a> [keyfactor\_github\_test\_environment\_ses\_2541](#module\_keyfactor\_github\_test\_environment\_ses\_2541) | git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git | main |

## Resources

| Name | Type |
|------|------|
| [github_repository.repo](https://registry.terraform.io/providers/integrations/github/latest/docs/data-sources/repository) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_ses_2541_auth_token_url"></a> [ses\_2541\_auth\_token\_url](#input\_ses\_2541\_auth\_token\_url) | The OAuth token URL for the SES 25.4.1 Keyfactor Command instance | `string` | `"https://auth.kftestlab.com/oauth2/token"` | no |
| <a name="input_ses_2541_client_id"></a> [ses\_2541\_client\_id](#input\_ses\_2541\_client\_id) | The OAuth client ID for the SES 25.4.1 Keyfactor Command instance | `string` | n/a | yes |
| <a name="input_ses_2541_client_secret"></a> [ses\_2541\_client\_secret](#input\_ses\_2541\_client\_secret) | The OAuth client secret for the SES 25.4.1 Keyfactor Command instance | `string` | n/a | yes |
| <a name="input_ses_2541_hostname"></a> [ses\_2541\_hostname](#input\_ses\_2541\_hostname) | The hostname of the SES 25.4.1 Keyfactor Command instance | `string` | `"int25-4-1.kftestlab.com"` | no |

## Outputs

No outputs.
<!-- END_TF_DOCS -->
