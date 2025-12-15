# PostgreSQL Flexible Server
resource "azurerm_postgresql_flexible_server" "main" {
  name                   = "${var.project_name}-${var.environment}-postgres"
  resource_group_name    = azurerm_resource_group.main.name
  location               = azurerm_resource_group.main.location
  
  administrator_login    = var.postgres_admin_username
  administrator_password = var.postgres_admin_password
  
  sku_name   = var.postgres_sku_name
  version    = "15"
  storage_mb = var.postgres_storage_size_gb * 1024
  
  backup_retention_days        = 7
  geo_redundant_backup_enabled = var.environment == "prod" ? true : false
  
  zone = "1"
  
  dynamic "high_availability" {
    for_each = var.environment == "prod" ? [1] : []
    content {
      mode                      = "ZoneRedundant"
      standby_availability_zone = "2"
    }
  }
  
  tags = local.common_tags
}

resource "azurerm_postgresql_flexible_server_firewall_rule" "allow_azure_services" {
  name             = "AllowAzureServices"
  server_id        = azurerm_postgresql_flexible_server.main.id
  start_ip_address = "0.0.0.0"
  end_ip_address   = "0.0.0.0"
}

resource "azurerm_postgresql_flexible_server_firewall_rule" "allowed_ips" {
  for_each = toset(var.allowed_ip_ranges)
  
  name             = "AllowIP-${md5(each.value)}"
  server_id        = azurerm_postgresql_flexible_server.main.id
  start_ip_address = split("/", each.value)[0]
  end_ip_address   = split("/", each.value)[0]
}

resource "azurerm_postgresql_flexible_server_database" "main" {
  name      = "multiplayer_db"
  server_id = azurerm_postgresql_flexible_server.main.id
  charset   = "UTF8"
  collation = "en_US.utf8"
}

resource "azurerm_postgresql_flexible_server_configuration" "timezone" {
  name      = "timezone"
  server_id = azurerm_postgresql_flexible_server.main.id
  value     = "UTC"
}

resource "azurerm_postgresql_flexible_server_configuration" "log_connections" {
  name      = "log_connections"
  server_id = azurerm_postgresql_flexible_server.main.id
  value     = "on"
}

resource "azurerm_postgresql_flexible_server_configuration" "log_disconnections" {
  name      = "log_disconnections"
  server_id = azurerm_postgresql_flexible_server.main.id
  value     = "on"
}
