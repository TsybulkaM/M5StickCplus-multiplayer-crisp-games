output "resource_group_name" {
  description = "Resource Group name"
  value       = azurerm_resource_group.main.name
}

output "storage_account_name" {
  description = "Storage Account name for firmware"
  value       = azurerm_storage_account.firmware.name
}

output "storage_account_key" {
  description = "Storage Account access key"
  value       = azurerm_storage_account.firmware.primary_access_key
  sensitive   = true
}

output "storage_connection_string" {
  description = "Connection string for Storage Account"
  value       = azurerm_storage_account.firmware.primary_connection_string
  sensitive   = true
}

output "firmware_container_name" {
  description = "Container name for firmware"
  value       = azurerm_storage_container.firmware.name
}

output "postgres_fqdn" {
  description = "PostgreSQL server FQDN"
  value       = azurerm_postgresql_flexible_server.main.fqdn
}

output "postgres_username" {
  description = "PostgreSQL admin username"
  value       = var.postgres_admin_username
  sensitive   = true
}

output "postgres_password" {
  description = "PostgreSQL admin password"
  value       = var.postgres_admin_password
  sensitive   = true
}

output "postgres_admin_username" {
  description = "PostgreSQL admin username (alias)"
  value       = var.postgres_admin_username
  sensitive   = true
}

output "postgres_admin_password" {
  description = "PostgreSQL admin password (alias)"
  value       = var.postgres_admin_password
  sensitive   = true
}

output "postgres_connection_string" {
  description = "Connection string for PostgreSQL"
  value       = "postgres://${var.postgres_admin_username}:${var.postgres_admin_password}@${azurerm_postgresql_flexible_server.main.fqdn}:5432/${azurerm_postgresql_flexible_server_database.main.name}?sslmode=require"
  sensitive   = true
}

output "appinsights_instrumentation_key" {
  description = "Instrumentation Key for Application Insights"
  value       = azurerm_application_insights.main.instrumentation_key
  sensitive   = true
}

output "application_insights_instrumentation_key" {
  description = "Instrumentation Key for Application Insights (alias)"
  value       = azurerm_application_insights.main.instrumentation_key
  sensitive   = true
}

output "application_insights_connection_string" {
  description = "Connection String for Application Insights"
  value       = azurerm_application_insights.main.connection_string
  sensitive   = true
}

output "log_analytics_workspace_id" {
  description = "Log Analytics Workspace ID"
  value       = azurerm_log_analytics_workspace.main.id
}

output "aks_cluster_name" {
  description = "AKS cluster name"
  value       = azurerm_kubernetes_cluster.aks.name
}

output "aks_cluster_id" {
  description = "AKS cluster ID"
  value       = azurerm_kubernetes_cluster.aks.id
}

output "aks_kube_config" {
  description = "Kubeconfig for AKS connection"
  value       = azurerm_kubernetes_cluster.aks.kube_config_raw
  sensitive   = true
}

output "acr_login_server" {
  description = "Azure Container Registry URL"
  value       = azurerm_container_registry.acr.login_server
}

output "acr_admin_username" {
  description = "Admin username for ACR"
  value       = azurerm_container_registry.acr.admin_username
  sensitive   = true
}

output "acr_admin_password" {
  description = "Admin password for ACR"
  value       = azurerm_container_registry.acr.admin_password
  sensitive   = true
}
