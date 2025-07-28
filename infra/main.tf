provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
}

resource "azurerm_resource_group" "rg" {
  name     = var.resource_group_name
  location = var.location
}

resource "azurerm_storage_account" "storage" {
  name                     = var.storage_account_name
  resource_group_name      = azurerm_resource_group.rg.name
  location                 = var.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

# Create the quotesubmissions container for storing reflections
resource "azurerm_storage_container" "quotesubmissions" {
  name                  = "quotesubmissions"
  storage_account_id    = azurerm_storage_account.storage.id
  container_access_type = "private"
  
  depends_on = [azurerm_storage_account.storage]
}

# Create the questions container for storing daily questions
resource "azurerm_storage_container" "questions" {
  name                  = "questions"
  storage_account_id    = azurerm_storage_account.storage.id
  container_access_type = "private"
  
  depends_on = [azurerm_storage_account.storage]
}

resource "azurerm_container_group" "quotespark" {
  name                = "quotespark-container"
  location            = var.location
  resource_group_name = azurerm_resource_group.rg.name
  os_type             = "Linux"

  container {
    name   = "quotespark"
    image  = var.container_image_url
    cpu    = "0.5"
    memory = "1.0"

    ports {
      port     = 8080
      protocol = "TCP"
    }

    environment_variables = {
      GROQ_API_KEY = var.groq_api_key
      AZURE_STORAGE_ACCOUNT = azurerm_storage_account.storage.name
      AZURE_STORAGE_KEY     = azurerm_storage_account.storage.primary_access_key
    }
  }

  image_registry_credential {
    server   = var.acr_login_server
    username = var.acr_username
    password = var.acr_password
  }

  ip_address_type = "Public"
  dns_name_label  = "quotespark-2211"
  
  depends_on = [
    azurerm_storage_container.quotesubmissions,
    azurerm_storage_container.questions
  ]
}
