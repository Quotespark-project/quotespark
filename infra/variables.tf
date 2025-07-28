variable "resource_group_name" {}
variable "location" {
  default = "canadacentral"
}

variable "storage_account_name" {}
variable "subscription_id" {}

variable "acr_login_server" {}
variable "acr_username" {}
variable "acr_password" {}
variable "container_image_url" {}

variable "groq_api_key" {}