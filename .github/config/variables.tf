variable "ses_2541_hostname" {
  description = "The hostname of the SES 25.4.1 Keyfactor Command instance"
  type        = string
  default     = "int25-4-1.kftestlab.com"
}

variable "ses_2541_auth_token_url" {
  description = "The OAuth token URL for the SES 25.4.1 Keyfactor Command instance"
  type        = string
  default     = "https://auth.kftestlab.com/oauth2/token"
}

variable "ses_2541_client_id" {
  description = "The OAuth client ID for the SES 25.4.1 Keyfactor Command instance"
  type        = string
}

variable "ses_2541_client_secret" {
  description = "The OAuth client secret for the SES 25.4.1 Keyfactor Command instance"
  type        = string
  sensitive   = true
}
