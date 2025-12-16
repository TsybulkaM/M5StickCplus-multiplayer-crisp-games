resource "random_string" "storage_suffix" {
  length  = 6
  special = false
  upper   = false
}

resource "azurerm_storage_account" "firmware" {
  name                     = "crisp${var.environment}${random_string.storage_suffix.result}"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = var.storage_account_replication
  account_kind             = "StorageV2"
  
  allow_nested_items_to_be_public = false
  min_tls_version                 = "TLS1_2"
  https_traffic_only_enabled      = true
  
  blob_properties {
    versioning_enabled = true
    
    delete_retention_policy {
      days = 7
    }
    
    container_delete_retention_policy {
      days = 7
    }
  }
  
  tags = local.common_tags
}

resource "azurerm_storage_container" "firmware" {
  name                  = "firmware"
  storage_account_name  = azurerm_storage_account.firmware.name
  container_access_type = "private"
}

resource "azurerm_storage_container" "backups" {
  name                  = "backups"
  storage_account_name  = azurerm_storage_account.firmware.name
  container_access_type = "private"
}

resource "azurerm_storage_management_policy" "firmware" {
  storage_account_id = azurerm_storage_account.firmware.id

  rule {
    name    = "delete-old-versions"
    enabled = true
    
    filters {
      blob_types = ["blockBlob"]
    }
    
    actions {
      base_blob {
        tier_to_cool_after_days_since_modification_greater_than = 30
        delete_after_days_since_modification_greater_than = 365
      }
      
      version {
        delete_after_days_since_creation = 90
      }
    }
  }
}
