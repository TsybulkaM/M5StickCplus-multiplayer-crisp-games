variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "dev"
  
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be: dev, staging or prod"
  }
}

variable "location" {
  description = "Azure region for resources"
  type        = string
  default     = "eastus"
}

variable "project_name" {
  description = "Project name (used in resource names)"
  type        = string
  default     = "m5stick-multiplayer"
}

variable "postgres_admin_username" {
  description = "PostgreSQL admin username"
  type        = string
  default     = "pgadmin"
  sensitive   = true
}

variable "postgres_admin_password" {
  description = "PostgreSQL admin password"
  type        = string
  sensitive   = true
}

variable "postgres_sku_name" {
  description = "SKU for PostgreSQL Flexible Server"
  type        = string
  default     = "B_Standard_B1ms"
}

variable "postgres_storage_size_gb" {
  description = "PostgreSQL storage size in GB"
  type        = number
  default     = 32
}

variable "storage_account_replication" {
  description = "Replication type for Storage Account"
  type        = string
  default     = "LRS"
  
  validation {
    condition     = contains(["LRS", "GRS", "RAGRS", "ZRS"], var.storage_account_replication)
    error_message = "Replication must be: LRS, GRS, RAGRS or ZRS"
  }
}

variable "container_instances_cpu" {
  description = "Number of CPU cores for container"
  type        = number
  default     = 1
}

variable "container_instances_memory" {
  description = "Memory amount for container in GB"
  type        = number
  default     = 1.5
}

variable "mqtt_broker_image" {
  description = "Docker image for EMQX broker"
  type        = string
  default     = "emqx/emqx:5.8.3"
}

variable "allowed_ip_ranges" {
  description = "Allowed IP ranges for resource access"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "aks_node_count" {
  description = "Initial number of nodes in AKS cluster"
  type        = number
  default     = 1
}

variable "aks_min_node_count" {
  description = "Minimum number of nodes for autoscaling"
  type        = number
  default     = 1
}

variable "aks_max_node_count" {
  description = "Maximum number of nodes for autoscaling"
  type        = number
  default     = 5
}

variable "aks_vm_size" {
  description = "VM size for AKS nodes"
  type        = string
  default     = "Standard_B2as_v2"
}
