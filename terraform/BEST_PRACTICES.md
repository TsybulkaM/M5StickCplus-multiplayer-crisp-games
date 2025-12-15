# Terraform Best Practices для Azure

## 📚 Содержание

- [Структура проекта](#структура-проекта)
- [Именование ресурсов](#именование-ресурсов)
- [Управление состоянием](#управление-состоянием)
- [Переменные и outputs](#переменные-и-outputs)
- [Безопасность](#безопасность)
- [Модули](#модули)
- [Testing](#testing)

---

## Структура проекта

### ✅ Рекомендуемая структура

```
terraform/
├── main.tf                 # Provider и основная конфигурация
├── variables.tf            # Входные переменные
├── outputs.tf              # Выходные значения
├── terraform.tfvars        # Значения переменных (в .gitignore)
├── terraform.tfvars.example # Пример конфигурации
├── .gitignore              # Исключения из git
├── README.md               # Документация
├── resource_group.tf       # Resource Group
├── storage.tf              # Storage Account и Blob Storage
├── database.tf             # PostgreSQL
├── container_instances.tf  # Container Instances
├── monitoring.tf           # Log Analytics, App Insights, Alerts
├── networking.tf           # VNet, Subnets, NSG (если нужно)
└── modules/                # Переиспользуемые модули
    ├── storage/
    ├── database/
    └── monitoring/
```

### ❌ Избегайте

- Один огромный `main.tf` со всеми ресурсами
- Хардкод значений вместо переменных
- Отсутствие документации

---

## Именование ресурсов

### Конвенция именования

```hcl
# Формат: {project}-{environment}-{resource-type}
# Пример: m5stick-dev-postgres

resource "azurerm_resource_group" "main" {
  name     = "${var.project_name}-${var.environment}-rg"
  location = var.location
}

resource "azurerm_storage_account" "firmware" {
  # Storage Account имена: только lowercase и цифры, 3-24 символа
  name = "${replace(var.project_name, "-", "")}${var.environment}${random_string.suffix.result}"
}
```

### Правила

1. **Используйте переменные**: `var.project_name` вместо хардкода
2. **Добавляйте окружение**: `-dev`, `-staging`, `-prod`
3. **Уникальность**: Storage Account требует глобально уникальное имя
4. **Ограничения Azure**: соблюдайте лимиты символов и разрешенные символы

---

## Управление состоянием

### ✅ Remote State (для команды)

```hcl
# main.tf
terraform {
  backend "azurerm" {
    resource_group_name  = "terraform-state-rg"
    storage_account_name = "tfstate${random_string.suffix.result}"
    container_name       = "tfstate"
    key                  = "m5stick-multiplayer.terraform.tfstate"
  }
}
```

**Преимущества:**
- Командная работа
- State locking (предотвращает конфликты)
- Версионирование
- Бэкапы

### Создание backend storage

```bash
#!/bin/bash
RESOURCE_GROUP="terraform-state-rg"
STORAGE_ACCOUNT="tfstate$(openssl rand -hex 4)"
CONTAINER="tfstate"
LOCATION="eastus"

# Создание resource group
az group create --name $RESOURCE_GROUP --location $LOCATION

# Создание storage account
az storage account create \
  --name $STORAGE_ACCOUNT \
  --resource-group $RESOURCE_GROUP \
  --location $LOCATION \
  --sku Standard_LRS \
  --encryption-services blob

# Получение ключа
ACCOUNT_KEY=$(az storage account keys list \
  --resource-group $RESOURCE_GROUP \
  --account-name $STORAGE_ACCOUNT \
  --query '[0].value' -o tsv)

# Создание контейнера
az storage container create \
  --name $CONTAINER \
  --account-name $STORAGE_ACCOUNT \
  --account-key $ACCOUNT_KEY

echo "Backend Storage Account: $STORAGE_ACCOUNT"
```

### State locking

```hcl
# Автоматически включен при использовании azurerm backend
# Azure использует blob leases для лока
```

---

## Переменные и outputs

### ✅ Хорошие практики для variables

```hcl
# variables.tf

variable "environment" {
  description = "Окружение развертывания (dev, staging, prod)"
  type        = string
  
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment должен быть: dev, staging или prod"
  }
}

variable "postgres_admin_password" {
  description = "Пароль администратора PostgreSQL"
  type        = string
  sensitive   = true
  
  validation {
    condition     = length(var.postgres_admin_password) >= 8
    error_message = "Пароль должен содержать минимум 8 символов"
  }
}

variable "allowed_ip_ranges" {
  description = "Разрешенные IP диапазоны для доступа"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Теги для всех ресурсов"
  type        = map(string)
  default = {
    ManagedBy = "Terraform"
  }
}
```

### ✅ Хорошие практики для outputs

```hcl
# outputs.tf

output "postgres_fqdn" {
  description = "FQDN PostgreSQL сервера"
  value       = azurerm_postgresql_flexible_server.main.fqdn
}

output "storage_account_key" {
  description = "Ключ доступа к Storage Account"
  value       = azurerm_storage_account.firmware.primary_access_key
  sensitive   = true  # Скрывает значение в логах
}

output "connection_strings" {
  description = "Connection strings для всех сервисов"
  value = {
    postgres = "postgres://${var.postgres_admin_username}@${azurerm_postgresql_flexible_server.main.fqdn}"
    mqtt     = "tcp://${azurerm_container_group.mqtt_broker.fqdn}:1883"
    portal   = "http://${azurerm_container_group.portal.fqdn}:8080"
  }
  sensitive = true
}
```

---

## Безопасность

### 1. Секреты и sensitive данные

```hcl
# ❌ НЕ ДЕЛАЙТЕ ТАК
variable "password" {
  default = "MyPassword123!"  # Хардкод в коде
}

# ✅ ПРАВИЛЬНО
variable "password" {
  type      = string
  sensitive = true
  # Передается через terraform.tfvars (в .gitignore)
  # Или через environment variable: TF_VAR_password
}
```

### 2. Используйте Azure Key Vault

```hcl
# Получение секретов из Key Vault
data "azurerm_key_vault_secret" "db_password" {
  name         = "postgres-admin-password"
  key_vault_id = azurerm_key_vault.main.id
}

resource "azurerm_postgresql_flexible_server" "main" {
  administrator_password = data.azurerm_key_vault_secret.db_password.value
}
```

### 3. Network Security

```hcl
# Ограничение доступа к Storage Account
resource "azurerm_storage_account_network_rules" "main" {
  storage_account_id = azurerm_storage_account.firmware.id
  
  default_action             = "Deny"
  bypass                     = ["AzureServices"]
  ip_rules                   = var.allowed_ip_ranges
  virtual_network_subnet_ids = [azurerm_subnet.main.id]
}

# PostgreSQL firewall
resource "azurerm_postgresql_flexible_server_firewall_rule" "office" {
  name             = "OfficeNetwork"
  server_id        = azurerm_postgresql_flexible_server.main.id
  start_ip_address = var.office_ip
  end_ip_address   = var.office_ip
}
```

### 4. TLS/SSL

```hcl
resource "azurerm_storage_account" "main" {
  enable_https_traffic_only = true
  min_tls_version          = "TLS1_2"
}

resource "azurerm_postgresql_flexible_server" "main" {
  # SSL по умолчанию включен
  # Connection string должен содержать sslmode=require
}
```

---

## Модули

### Когда создавать модули

1. Код повторяется в разных проектах
2. Логически связанные ресурсы (например, storage + backup policy)
3. Разные окружения с одинаковой инфраструктурой

### Структура модуля

```
modules/
└── storage/
    ├── main.tf       # Ресурсы
    ├── variables.tf  # Входные переменные
    ├── outputs.tf    # Выходные значения
    └── README.md     # Документация
```

### Пример модуля

```hcl
# modules/storage/main.tf
resource "azurerm_storage_account" "main" {
  name                     = var.storage_account_name
  resource_group_name      = var.resource_group_name
  location                 = var.location
  account_tier             = var.account_tier
  account_replication_type = var.replication_type
  
  enable_https_traffic_only = true
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_container" "main" {
  for_each = toset(var.containers)
  
  name                  = each.value
  storage_account_name  = azurerm_storage_account.main.name
  container_access_type = "private"
}
```

### Использование модуля

```hcl
# main.tf
module "firmware_storage" {
  source = "./modules/storage"
  
  storage_account_name = "m5stickdev"
  resource_group_name  = azurerm_resource_group.main.name
  location             = azurerm_resource_group.main.location
  
  containers = ["firmware", "backups", "logs"]
  
  account_tier     = "Standard"
  replication_type = "LRS"
}

output "firmware_storage_url" {
  value = module.firmware_storage.primary_blob_endpoint
}
```

---

## Testing

### 1. Terraform validate

```bash
terraform validate
```

Проверяет синтаксис и конфигурацию.

### 2. Terraform plan

```bash
terraform plan -out=tfplan

# Для конкретного окружения
terraform plan -var-file=environments/dev.tfvars
```

### 3. TFLint

```bash
# Установка
brew install tflint  # macOS
# или
curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | bash

# Использование
tflint --init
tflint
```

### 4. Checkov (Security scanning)

```bash
# Установка
pip install checkov

# Сканирование
checkov -d terraform/

# В CI/CD
checkov -d terraform/ --output json > checkov-report.json
```

### 5. Terratest (Go tests)

```go
// test/storage_test.go
package test

import (
    "testing"
    "github.com/gruntwork-io/terratest/modules/terraform"
    "github.com/stretchr/testify/assert"
)

func TestStorageAccount(t *testing.T) {
    terraformOptions := &terraform.Options{
        TerraformDir: "../terraform",
        Vars: map[string]interface{}{
            "environment": "test",
        },
    }
    
    defer terraform.Destroy(t, terraformOptions)
    terraform.InitAndApply(t, terraformOptions)
    
    storageAccountName := terraform.Output(t, terraformOptions, "storage_account_name")
    assert.NotEmpty(t, storageAccountName)
}
```

### 6. Pre-commit hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/antonbabenko/pre-commit-terraform
    rev: v1.77.0
    hooks:
      - id: terraform_fmt
      - id: terraform_validate
      - id: terraform_tflint
      - id: terraform_checkov
```

---

## Дополнительные best practices

### 1. Используйте locals для вычисляемых значений

```hcl
locals {
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
    CostCenter  = var.cost_center
  }
  
  # Conditional logic
  enable_ha = var.environment == "prod" ? true : false
  
  # String manipulation
  storage_account_name = "${replace(var.project_name, "-", "")}${var.environment}"
}
```

### 2. Dynamic blocks

```hcl
resource "azurerm_postgresql_flexible_server" "main" {
  # High availability только для prod
  dynamic "high_availability" {
    for_each = var.environment == "prod" ? [1] : []
    content {
      mode = "ZoneRedundant"
    }
  }
}
```

### 3. Count vs for_each

```hcl
# ✅ Используйте for_each для коллекций
resource "azurerm_storage_container" "containers" {
  for_each = toset(["firmware", "backups", "logs"])
  
  name                  = each.value
  storage_account_name  = azurerm_storage_account.main.name
}

# ❌ Избегайте count для списков (проблемы с удалением элементов из середины)
resource "azurerm_storage_container" "containers" {
  count = length(var.containers)
  name  = var.containers[count.index]
}
```

### 4. Data sources

```hcl
# Получение существующих ресурсов
data "azurerm_client_config" "current" {}

data "azurerm_resource_group" "existing" {
  name = "existing-rg"
}

# Использование
resource "azurerm_key_vault" "main" {
  tenant_id = data.azurerm_client_config.current.tenant_id
  location  = data.azurerm_resource_group.existing.location
}
```

### 5. Версионирование providers

```hcl
terraform {
  required_version = ">= 1.0"
  
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"  # Используйте patch updates автоматически
    }
  }
}
```

### 6. Комментарии и документация

```hcl
# Объясняйте WHY, а не WHAT
# ❌ Плохо
# Create storage account
resource "azurerm_storage_account" "main" {

# ✅ Хорошо
# LRS replication для dev окружения, чтобы снизить стоимость.
# В prod используется GRS для disaster recovery.
resource "azurerm_storage_account" "main" {
  account_replication_type = var.environment == "prod" ? "GRS" : "LRS"
}
```

---

## Чек-лист перед production deployment

- [ ] Remote state настроен
- [ ] Все секреты в Azure Key Vault или переменных
- [ ] Firewall rules ограничены (не 0.0.0.0/0)
- [ ] Включено логирование и мониторинг
- [ ] Настроены alerts для критических метрик
- [ ] Включены автоматические бэкапы
- [ ] High Availability для критических сервисов
- [ ] Disaster recovery процедуры протестированы
- [ ] Документация актуальна
- [ ] Tags для cost tracking настроены
- [ ] Security scanning (Checkov) пройден
- [ ] Peer review выполнен

---

## Полезные ссылки

- [Terraform Best Practices](https://www.terraform-best-practices.com/)
- [Azure Naming Conventions](https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-naming)
- [Terraform Registry](https://registry.terraform.io/)
- [TFLint Rules](https://github.com/terraform-linters/tflint/tree/master/docs/rules)
