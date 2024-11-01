# GitHub Test Environment Setup

This code sets up GitHub environments for testing against Keyfactor Command instances that are configured to use
Active Directory or Keycloak for authentication.

## Requirements

1. Terraform >= 1.0
2. GitHub Provider >= 6.2
3. Keyfactor Command instance(s) configured to use Active Directory or Keycloak for authentication
4. AD or Keycloak credentials for authenticating to the Keyfactor Command instance(s)
5. A GitHub token with access and permissions to the repository where the environments will be created

## Adding a new environment

Modify the `environments.tf` file to include the new environment module. The module should be named appropriately.
Example:

### Active Directory Environment

```hcl
module "keyfactor_github_test_environment_ad_10_5_0" {
  source = "git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git?ref=main"

  gh_environment_name = "KFC_10_5_0" # Keyfactor Command 10.5.0 environment using Active Directory(/Basic Auth)
  gh_repo_name       = data.github_repository.repo.name
  keyfactor_hostname = var.keyfactor_hostname_10_5_0
  keyfactor_username = var.keyfactor_username_AD
  keyfactor_password = var.keyfactor_password_AD
}
```

### oAuth Client Environment

```hcl
module "keyfactor_github_test_environment_12_3_0_kc" {
  source = "git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-kc.git?ref=main"

  gh_environment_name = "KFC_12_3_0_KC" # Keyfactor Command 12.3.0 environment using Keycloak
  gh_repo_name              = data.github_repository.repo.name
  keyfactor_hostname        = var.keyfactor_hostname_12_3_0_OAUTH
  keyfactor_auth_token_url  = var.keyfactor_auth_token_url
  keyfactor_client_id       = var.keyfactor_client_id
  keyfactor_client_secret   = var.keyfactor_client_secret
  keyfactor_tls_skip_verify = true
}
```

<!-- BEGIN_TF_DOCS -->

## Requirements

| Name                                                                      | Version |
|---------------------------------------------------------------------------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0  |
| <a name="requirement_github"></a> [github](#requirement\_github)          | >=6.2   |

## Providers

| Name                                                       | Version |
|------------------------------------------------------------|---------|
| <a name="provider_github"></a> [github](#provider\_github) | 6.3.1   |

## Modules

| Name                                                                                                                                                                                                             | Source                                                                                        | Version |
|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------|---------|
| <a name="module_keyfactor_github_test_environment_10_5_0"></a> [keyfactor\_github\_test\_environment\_10\_5\_0](#module\_keyfactor\_github\_test\_environment\_10\_5\_0)                                         | git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git | main    |
| <a name="module_keyfactor_github_test_environment_10_5_0_CLEAN"></a> [keyfactor\_github\_test\_environment\_10\_5\_0\_CLEAN](#module\_keyfactor\_github\_test\_environment\_10\_5\_0\_CLEAN)                     | git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git | main    |
| <a name="module_keyfactor_github_test_environment_11_5_0"></a> [keyfactor\_github\_test\_environment\_11\_5\_0](#module\_keyfactor\_github\_test\_environment\_11\_5\_0)                                         | git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git | main    |
| <a name="module_keyfactor_github_test_environment_11_5_0_CLEAN"></a> [keyfactor\_github\_test\_environment\_11\_5\_0\_CLEAN](#module\_keyfactor\_github\_test\_environment\_11\_5\_0\_CLEAN)                     | git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git | main    |
| <a name="module_keyfactor_github_test_environment_11_5_0_OAUTH"></a> [keyfactor\_github\_test\_environment\_11\_5\_0\_OAUTH](#module\_keyfactor\_github\_test\_environment\_11\_5\_0\_OAUTH)                     | git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git | main    |
| <a name="module_keyfactor_github_test_environment_11_5_0_OAUTH_CLEAN"></a> [keyfactor\_github\_test\_environment\_11\_5\_0\_OAUTH\_CLEAN](#module\_keyfactor\_github\_test\_environment\_11\_5\_0\_OAUTH\_CLEAN) | git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git | main    |
| <a name="module_keyfactor_github_test_environment_12_3_0_AD"></a> [keyfactor\_github\_test\_environment\_12\_3\_0\_AD](#module\_keyfactor\_github\_test\_environment\_12\_3\_0\_AD)                              | git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git | main    |
| <a name="module_keyfactor_github_test_environment_12_3_0_AD_CLEAN"></a> [keyfactor\_github\_test\_environment\_12\_3\_0\_AD\_CLEAN](#module\_keyfactor\_github\_test\_environment\_12\_3\_0\_AD\_CLEAN)          | git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git | main    |
| <a name="module_keyfactor_github_test_environment_12_3_0_OAUTH"></a> [keyfactor\_github\_test\_environment\_12\_3\_0\_OAUTH](#module\_keyfactor\_github\_test\_environment\_12\_3\_0\_OAUTH)                     | git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git | main    |
| <a name="module_keyfactor_github_test_environment_12_3_0_OAUTH_CLEAN"></a> [keyfactor\_github\_test\_environment\_12\_3\_0\_OAUTH\_CLEAN](#module\_keyfactor\_github\_test\_environment\_12\_3\_0\_OAUTH\_CLEAN) | git::ssh://git@github.com/Keyfactor/terraform-module-keyfactor-github-test-environment-ad.git | main    |

## Resources

| Name                                                                                                                      | Type        |
|---------------------------------------------------------------------------------------------------------------------------|-------------|
| [github_repository.repo](https://registry.terraform.io/providers/integrations/github/latest/docs/data-sources/repository) | data source |

## Inputs

| Name                                                                                                                                                          | Description                                                                                                        | Type     | Default                                                                                                 | Required |
|---------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------|----------|---------------------------------------------------------------------------------------------------------|:--------:|
| <a name="input_keyfactor_auth_token_url"></a> [keyfactor\_auth\_token\_url](#input\_keyfactor\_auth\_token\_url)                                              | The token URL to authenticate with the Keyfactor instance using oauth2 client credentials                          | `string` | `"https://int-oidc-lab.eastus2.cloudapp.azure.com:8444/realms/Keyfactor/protocol/openid-connect/token"` |    no    |
| <a name="input_keyfactor_client_id"></a> [keyfactor\_client\_id](#input\_keyfactor\_client\_id)                                                               | The client ID to authenticate with the Keyfactor instance using oauth2 client credentials                          | `string` | n/a                                                                                                     |   yes    |
| <a name="input_keyfactor_client_secret"></a> [keyfactor\_client\_secret](#input\_keyfactor\_client\_secret)                                                   | The client secret to authenticate with the Keyfactor instance using oauth2 client credentials                      | `string` | n/a                                                                                                     |   yes    |
| <a name="input_keyfactor_hostname_10_5_0"></a> [keyfactor\_hostname\_10\_5\_0](#input\_keyfactor\_hostname\_10\_5\_0)                                         | The hostname of the Keyfactor v10.5.x instance                                                                     | `string` | `"integrations1050-lab.kfdelivery.com"`                                                                 |    no    |
| <a name="input_keyfactor_hostname_10_5_0_CLEAN"></a> [keyfactor\_hostname\_10\_5\_0\_CLEAN](#input\_keyfactor\_hostname\_10\_5\_0\_CLEAN)                     | The hostname of the Keyfactor v10.5.x instance with no stores or orchestrators. This is used for store-type tests. | `string` | `"int1050-test-clean.kfdelivery.com"`                                                                   |    no    |
| <a name="input_keyfactor_hostname_11_5_0"></a> [keyfactor\_hostname\_11\_5\_0](#input\_keyfactor\_hostname\_11\_5\_0)                                         | The hostname of the Keyfactor v11.5.x instance                                                                     | `string` | `"integrations1150-lab.kfdelivery.com"`                                                                 |    no    |
| <a name="input_keyfactor_hostname_11_5_0_CLEAN"></a> [keyfactor\_hostname\_11\_5\_0\_CLEAN](#input\_keyfactor\_hostname\_11\_5\_0\_CLEAN)                     | The hostname of the Keyfactor v11.5.x instance with no stores or orchestrators. This is used for store-type tests. | `string` | `"int1150-test-clean.kfdelivery.com"`                                                                   |    no    |
| <a name="input_keyfactor_hostname_11_5_0_OAUTH"></a> [keyfactor\_hostname\_11\_5\_0\_OAUTH](#input\_keyfactor\_hostname\_11\_5\_0\_OAUTH)                     | The hostname of the Keyfactor instance                                                                             | `string` | `"int-oidc-lab.eastus2.cloudapp.azure.com"`                                                             |    no    |
| <a name="input_keyfactor_hostname_11_5_0_OAUTH_CLEAN"></a> [keyfactor\_hostname\_11\_5\_0\_OAUTH\_CLEAN](#input\_keyfactor\_hostname\_11\_5\_0\_OAUTH\_CLEAN) | The hostname of the Keyfactor instance                                                                             | `string` | `"int1150-oauth-test-clean.eastus2.cloudapp.azure.com"`                                                 |    no    |
| <a name="input_keyfactor_hostname_12_3_0"></a> [keyfactor\_hostname\_12\_3\_0](#input\_keyfactor\_hostname\_12\_3\_0)                                         | The hostname of the Keyfactor v12.3.x instance                                                                     | `string` | `"integrations1230-lab.kfdelivery.com"`                                                                 |    no    |
| <a name="input_keyfactor_hostname_12_3_0_CLEAN"></a> [keyfactor\_hostname\_12\_3\_0\_CLEAN](#input\_keyfactor\_hostname\_12\_3\_0\_CLEAN)                     | The hostname of the Keyfactor v12.3.x instance with no stores or orchestrators. This is used for store-type tests. | `string` | `"int1230-test-clean.kfdelivery.com"`                                                                   |    no    |
| <a name="input_keyfactor_hostname_12_3_0_OAUTH"></a> [keyfactor\_hostname\_12\_3\_0\_OAUTH](#input\_keyfactor\_hostname\_12\_3\_0\_OAUTH)                     | The hostname of the Keyfactor instance                                                                             | `string` | `"int-oidc-lab.eastus2.cloudapp.azure.com"`                                                             |    no    |
| <a name="input_keyfactor_hostname_12_3_0_OAUTH_CLEAN"></a> [keyfactor\_hostname\_12\_3\_0\_OAUTH\_CLEAN](#input\_keyfactor\_hostname\_12\_3\_0\_OAUTH\_CLEAN) | The hostname of the Keyfactor instance                                                                             | `string` | `"int1230-oauth-test-clean.eastus2.cloudapp.azure.com"`                                                 |    no    |
| <a name="input_keyfactor_password_AD"></a> [keyfactor\_password\_AD](#input\_keyfactor\_password\_AD)                                                         | The password to authenticate with Keyfactor instance that uses AD authentication                                   | `string` | n/a                                                                                                     |   yes    |
| <a name="input_keyfactor_username_AD"></a> [keyfactor\_username\_AD](#input\_keyfactor\_username\_AD)                                                         | The username to authenticate with a Keyfactor instance that uses AD authentication                                 | `string` | n/a                                                                                                     |   yes    |

## Outputs

No outputs.
<!-- END_TF_DOCS -->