output "container_url" {
  value = "http://${azurerm_container_group.quotespark.fqdn}:8080"
}

output "storage_account_name" {
  value = azurerm_storage_account.storage.name
  description = "The name of the storage account"
}

output "storage_account_primary_key" {
  value = azurerm_storage_account.storage.primary_access_key
  description = "The primary access key for the storage account"
  sensitive = true
}

output "container_group_fqdn" {
  value = azurerm_container_group.quotespark.fqdn
  description = "The FQDN of the container group"
}

output "resource_group_name" {
  value = azurerm_resource_group.rg.name
  description = "The name of the resource group"
}
