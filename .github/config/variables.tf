// Hosts
variable "keyfactor_hostname_10_5_0" {
  description = "The hostname of the Keyfactor v10.5.x instance"
  type        = string
  default     = "integrations1050-lab.kfdelivery.com"
}

variable "keyfactor_hostname_10_5_0_CLEAN" {
  description = "The hostname of the Keyfactor v10.5.x instance with no stores or orchestrators. This is used for store-type tests."
  type        = string
  default     = "int1050-test-clean.kfdelivery.com"
}


variable "keyfactor_hostname_11_5_0" {
  description = "The hostname of the Keyfactor v11.5.x instance"
  type        = string
  default     = "integrations1150-lab.kfdelivery.com"
}

variable "keyfactor_hostname_11_5_0_CLEAN" {
  description = "The hostname of the Keyfactor v11.5.x instance with no stores or orchestrators. This is used for store-type tests."
  type        = string
  default     = "int1150-test-clean.kfdelivery.com"
}

variable "keyfactor_hostname_11_5_0_OAUTH" {
  description = "The hostname of the Keyfactor instance"
  type        = string
  default     = "int-oidc-lab.eastus2.cloudapp.azure.com"
}

variable "keyfactor_hostname_11_5_0_OAUTH_CLEAN" {
  description = "The hostname of the Keyfactor instance"
  type        = string
  default     = "int1150-oauth-test-clean.eastus2.cloudapp.azure.com"
}


variable "keyfactor_hostname_12_3_0" {
  description = "The hostname of the Keyfactor v12.3.x instance"
  type        = string
  default     = "integrations1230-lab.kfdelivery.com"
}

variable "keyfactor_hostname_12_3_0_CLEAN" {
  description = "The hostname of the Keyfactor v12.3.x instance with no stores or orchestrators. This is used for store-type tests."
  type        = string
  default     = "int1230-test-clean.kfdelivery.com"
}

variable "keyfactor_hostname_12_3_0_OAUTH" {
  description = "The hostname of the Keyfactor instance"
  type        = string
  default     = "int-oidc-lab.eastus2.cloudapp.azure.com"
}

variable "keyfactor_hostname_12_3_0_OAUTH_CLEAN" {
  description = "The hostname of the Keyfactor instance"
  type        = string
  default = "int-oidc-lab.eastus2.cloudapp.azure.com"
}

// Authentication
variable "keyfactor_username_AD" {
  description = "The username to authenticate with a Keyfactor instance that uses AD authentication"
  type        = string
}

variable "keyfactor_password_AD" {
  description = "The password to authenticate with Keyfactor instance that uses AD authentication"
  type        = string
}

variable "keyfactor_client_id" {
  description = "The client ID to authenticate with the Keyfactor instance using oauth2 client credentials"
  type        = string
}

variable "keyfactor_client_secret" {
  description = "The client secret to authenticate with the Keyfactor instance using oauth2 client credentials"
  type        = string
}

variable "keyfactor_auth_token_url" {
  description = "The token URL to authenticate with the Keyfactor instance using oauth2 client credentials"
  type        = string
  default     = "https://int-oidc-lab.eastus2.cloudapp.azure.com:8444/realms/Keyfactor/protocol/openid-connect/token"
}

