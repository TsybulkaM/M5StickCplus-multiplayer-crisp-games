resource "azurerm_log_analytics_workspace" "main" {
  name                = "${var.project_name}-${var.environment}-logs"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "PerGB2018"
  retention_in_days   = var.environment == "prod" ? 90 : 30
  
  tags = local.common_tags
}

resource "azurerm_application_insights" "main" {
  name                = "${var.project_name}-${var.environment}-insights"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  workspace_id        = azurerm_log_analytics_workspace.main.id
  application_type    = "web"
  
  tags = local.common_tags
}

resource "azurerm_monitor_metric_alert" "storage_capacity" {
  name                = "${var.project_name}-${var.environment}-storage-capacity-alert"
  resource_group_name = azurerm_resource_group.main.name
  scopes              = [azurerm_storage_account.firmware.id]
  description         = "Alert when storage capacity exceeds 80%"
  severity            = 2
  frequency           = "PT1H"
  window_size         = "PT1H"
  
  criteria {
    metric_namespace = "Microsoft.Storage/storageAccounts"
    metric_name      = "UsedCapacity"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = 80000000000 # 80 GB
  }
  
  action {
    action_group_id = azurerm_monitor_action_group.main.id
  }
  
  tags = local.common_tags
}

resource "azurerm_monitor_metric_alert" "postgres_cpu" {
  name                = "${var.project_name}-${var.environment}-postgres-cpu-alert"
  resource_group_name = azurerm_resource_group.main.name
  scopes              = [azurerm_postgresql_flexible_server.main.id]
  description         = "Alert when PostgreSQL CPU exceeds 80%"
  severity            = 2
  frequency           = "PT5M"
  window_size         = "PT15M"
  
  criteria {
    metric_namespace = "Microsoft.DBforPostgreSQL/flexibleServers"
    metric_name      = "cpu_percent"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = 80
  }
  
  action {
    action_group_id = azurerm_monitor_action_group.main.id
  }
  
  tags = local.common_tags
}

resource "azurerm_monitor_action_group" "main" {
  name                = "${var.project_name}-${var.environment}-action-group"
  resource_group_name = azurerm_resource_group.main.name
  short_name          = substr(var.project_name, 0, 12)
  
  webhook_receiver {
    name        = "webhook"
    service_uri = "https://example.com/webhook"
  }
  
  tags = local.common_tags
}
