# M5StickC+ Multiplayer Crisp Games

Cloud-connected multiplayer gaming platform for ESP32-based IoT devices with FOTA, leaderboards, and real-time telemetry.

## Overview

M5StickC+ Multiplayer Crisp Games is a production-ready IoT gaming ecosystem that enables ESP32 devices (M5StickC Plus) to run distributed games with cloud synchronization. The platform provides firmware-over-the-air updates, real-time telemetry ingestion, global leaderboards, and a web-based management portal.

The system is designed for high availability and horizontal scalability, running on Azure Kubernetes Service with managed PostgreSQL database, containerized microservices, and blob storage for firmware distribution.

### Core Capabilities

- **Native ESP32 Games**: Built on crisp-game-lib framework with FreeRTOS task management
- **FOTA (Firmware Over-The-Air)**: Secure HTTPS-based firmware distribution and versioning
- **Real-time Telemetry**: MQTT-based score tracking and device metrics collection
- **Cloud Backend**: Go microservices architecture deployed on Azure Kubernetes Service
- **Web Portal**: Administrative interface for firmware management and analytics
- **Security**: Token-based API authentication, TLS encryption, Azure Managed Identity integration

## System Architecture

### High-Level Design

```
┌─────────────────┐      HTTPS      ┌──────────────────┐
│  ESP32 Device   │ ────────────────▶│   Engine API     │
│  (M5StickC+)    │                  │  (Go + Gin)      │
│                 │◀─────────────────│  Port: 8081      │
│  - Game Logic   │   Firmware .bin  └──────────────────┘
│  - FOTA Client  │                           │
│  - MQTT Client  │                           │
└─────────────────┘                           ▼
         │                          ┌──────────────────┐
         │ MQTT                     │  PostgreSQL DB   │
         │ Scores/Telemetry         │  (Flexible)      │
         ▼                          └──────────────────┘
┌─────────────────┐                           │
│  EMQX Broker    │                           │
│  Port: 1883     │                           │
└─────────────────┘                           ▼
         │                          ┌──────────────────┐
         └─────────────────────────▶│   Portal Web     │
                                    │  (Go Templates)  │
                                    │  Port: 80        │
                                    └──────────────────┘
                                             │
                                             ▼
                                    ┌──────────────────┐
                                    │  Azure Blob      │
                                    │  (Firmware)      │
                                    └──────────────────┘
```

### Component Overview

**Edge Layer (ESP32)**
- ESP-IDF based firmware with FreeRTOS dual-core task management
- Core 1: Game execution loop
- Core 0: Network operations (MQTT, HTTPS)
- Secure boot and flash encryption support

**API Layer (Engine Service)**
- RESTful API built with Go and Gin framework
- FOTA endpoint for version checking and firmware streaming
- Token-based authentication for administrative operations
- Horizontal scaling via Kubernetes HPA (2-10 replicas)

**Data Layer**
- PostgreSQL Flexible Server for relational data (leaderboards, user profiles, telemetry)
- Azure Blob Storage (LRS) for binary firmware files
- EMQX MQTT broker for real-time message ingestion

**Presentation Layer (Portal Service)**
- Server-rendered Go templates
- Administrative dashboard for firmware management
- Analytics views for device metrics and game statistics

**Infrastructure**
- Azure Kubernetes Service (AKS) with managed node pools
- Azure Container Registry (ACR) for Docker images
- Application Insights for distributed tracing and metrics
- Log Analytics Workspace for centralized logging

## Deployment

### Prerequisites

**Hardware Requirements**
- M5StickC Plus or ESP32-PICO-D4 development board
- USB-C cable for programming

**Software Requirements**
- PlatformIO Core 6.0+ or ESP-IDF 4.4+
- Docker 20.10+ and Docker Compose 2.0+
- Azure CLI 2.40+
- Terraform 1.3+
- kubectl 1.24+
- Go 1.20+ (for local development)

**Azure Resources**
- Active Azure subscription with Owner or Contributor role
- Sufficient quota for: 2-5 vCPU, 3 Public IPs, 50GB storage

### 1. Local Development Setup

```bash
# Clone repository
git clone --recursive https://github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games.git
cd M5StickCplus-multiplayer-crisp-games

# Start local services (PostgreSQL + EMQX + Engine + Portal)
docker-compose up -d

# Check services
docker-compose ps
```

**Services will be available at:**
- Portal: http://localhost:8080
- Engine API: http://localhost:8081
- MQTT: localhost:1883
- PostgreSQL: localhost:5432

### 2. Build & Flash ESP32 Firmware

```bash
cd esp-crisp

# Build and flash
pio run -t upload -t monitor

# Or with ESP-IDF
idf.py build flash monitor
```

### 3. Upload First Game Firmware

```bash
# Get admin token
export ADMIN_TOKEN="dev-token-12345"  # from docker-compose.yml

# Upload firmware
curl -X POST http://localhost:8081/api/fota/upload \
  -H "X-API-Token: $ADMIN_TOKEN" \
  -F "file=@.pio/build/m5stick-c-plus/firmware.bin" \
  -F "version=1.0.0"
```

## Production Deployment to Azure

### Infrastructure Provisioning

The system uses Terraform to provision a complete Azure environment:

```bash
# Authenticate to Azure
az login

# Initialize Terraform
cd terraform
terraform init

# Review planned changes
terraform plan

# Deploy infrastructure
terraform apply
```

**Provisioned Resources:**
- Azure Kubernetes Service (AKS) cluster with 1-2 Standard_B2as_v2 nodes
- Azure Container Registry (ACR) for Docker image storage
- PostgreSQL Flexible Server (B_Standard_B1ms, 32GB storage)
- Azure Storage Account (LRS) with firmware and backup containers
- Application Insights workspace for monitoring
- Log Analytics workspace for log aggregation
- Azure Load Balancers with public IP addresses
- Metric alert rules for CPU and storage monitoring

### Service Deployment

Deploy containerized services to Kubernetes:

```bash
# Build and push Docker images to ACR
./deploy-local-k8s.sh

# This script performs:
# 1. ACR authentication
# 2. Docker image build (engine, portal)
# 3. Image push to registry
# 4. Kubernetes namespace creation
# 5. ConfigMap and Secret creation
# 6. Service deployment (engine, portal, mqtt)
# 7. HPA configuration
# 8. Database migration execution
```

### Service Discovery

After deployment, retrieve service endpoints:

```bash
kubectl get svc -n crisp-game

# Output example:
# NAME     TYPE           EXTERNAL-IP       PORT(S)
# engine   LoadBalancer   134.112.162.52    8081:30518/TCP
# mqtt     LoadBalancer   134.112.132.197   1883:30490/TCP
# portal   LoadBalancer   134.112.162.52    80:30496/TCP
```

### Authentication Configuration

Retrieve the generated admin API token:

```bash
kubectl get secret app-secrets -n crisp-game \
  -o jsonpath='{.data.ADMIN_API_TOKEN}' | base64 -d
```

This token is required for all firmware upload operations.

## 📦 Project Structure

```
.
├── esp-crisp/              # ESP32 firmware (PlatformIO)
│   ├── include/
│   │   └── config.h        # WiFi & server configuration
│   ├── src/
│   │   ├── main.cpp        # Entry point
│   │   ├── fota_client.cpp # FOTA update logic
│   │   └── mqtt_client.cpp # Telemetry sender
│   └── platformio.ini
├── services/               # Go backend services
│   ├── cmd/
│   │   ├── engine/         # FOTA & game API
│   │   └── portal/         # Web UI
│   ├── internal/
│   │   ├── core/           # Shared config
│   │   ├── engine/         # Engine business logic
│   │   └── storage/        # Azure Blob client
│   └── migrations/         # SQL schema
├── crisp-game-lib/         # Game framework (JavaScript → C++)
├── k8s/                    # Kubernetes manifests
│   ├── engine.yaml
│   ├── portal.yaml
│   ├── mqtt.yaml
│   └── hpa.yaml            # Auto-scaling
├── terraform/              # Azure infrastructure
│   ├── aks.tf              # Kubernetes cluster
│   ├── database.tf         # PostgreSQL
│   ├── storage.tf          # Blob storage
│   └── monitoring.tf       # Application Insights
├── docker-compose.yml      # Local development
└── deploy-local-k8s.sh     # One-click deployment
```


## API Reference

### FOTA Endpoints

#### Check for Firmware Updates (Public)

```http
GET /api/fota/check?current_version=1.0.0
```

Response:
```json
{
  "update_available": true,
  "latest_version": "1.0.1",
  "download_url": "/api/fota/download?version=1.0.1",
  "file_size": 1284880,
  "release_notes": "Bug fixes and performance improvements"
}
```

#### Download Firmware Binary (Public)

```http
GET /api/fota/download?version=1.0.1
```

Returns: `application/octet-stream` binary data

#### Upload Firmware (Protected)

```http
POST /api/fota/upload
Content-Type: multipart/form-data
X-API-Token: <admin-token>

file: firmware.bin
version: 1.0.0
release_notes: Optional description
```

Response:
```json
{
  "success": true,
  "version": "1.0.0",
  "file_size": 1284880,
  "uploaded_at": "2025-12-16T10:30:00Z"
}
```

### Health Check Endpoints

```http
GET /health         # Returns 200 OK if service is running
GET /health/ready   # Returns 200 OK if ready to accept traffic
```

## Operations and Monitoring

### Log Access

```bash
# Engine service logs
kubectl logs -f deployment/engine -n crisp-game

# Portal service logs
kubectl logs -f deployment/portal -n crisp-game

# MQTT broker logs
kubectl logs -f statefulset/mqtt -n crisp-game

# View all pods
kubectl get pods -n crisp-game -o wide
```

### Metrics and Alerts

**Application Insights**
- Access via Azure Portal: https://portal.azure.com
- Navigate to: Resource Group > Application Insights > crisp-project-dev-insights
- View: Request rates, response times, dependency calls, exceptions

**Configured Alerts:**
- PostgreSQL CPU usage > 80%
- Storage account capacity > 90%
- Email notifications to configured action group

### Horizontal Pod Autoscaling

Engine and Portal services auto-scale based on CPU utilization:

```yaml
Min replicas: 2
Max replicas: 10
Target CPU: 70%
```

Monitor autoscaling:
```bash
kubectl get hpa -n crisp-game
kubectl describe hpa engine-hpa -n crisp-game
```

### Database Access

```bash
# Connection string from Terraform output
terraform output -raw postgres_connection_string

# Direct psql connection
psql "postgres://<user>:<password>@<host>:5432/multiplayer_db?sslmode=require"
```

### Blob Storage Management

```bash
# List firmware files
az storage blob list \
  --account-name crispdevkd1icl \
  --container-name firmware \
  --output table \
  --auth-mode login

# Download firmware
az storage blob download \
  --account-name crispdevkd1icl \
  --container-name firmware \
  --name firmware_v1.0.0.bin \
  --file ./local-firmware.bin
```

## Security

### Authentication and Authorization

**API Token System**
- All firmware upload operations require X-API-Token header
- Token generated during deployment and stored in Kubernetes secrets
- Token rotation: Update secret and restart engine pods

**Network Security**
- Azure Network Security Groups control inbound traffic
- Kubernetes Network Policies isolate namespace traffic
- TLS/SSL supported for FOTA downloads (configurable)

**Identity and Access Management**
- Azure Managed Identity for AKS-to-ACR authentication
- Role-Based Access Control (RBAC) for Kubernetes resources
- PostgreSQL authentication via username/password with SSL enforcement

**Data Protection**
- Database connections use SSL/TLS (sslmode=require)
- Firmware files stored in private blob containers
- Kubernetes secrets for credential management
- Azure Key Vault integration (optional, not currently implemented)

### Security Best Practices

1. Rotate API tokens regularly
2. Enable Azure Defender for cloud resources
3. Configure firewall rules for PostgreSQL (currently allows all IPs for development)
4. Implement network policies to restrict pod-to-pod communication
5. Use Azure Private Link for database connections in production
6. Enable container image scanning in ACR

## Cost Analysis and Optimization

### Monthly Cost Breakdown (Azure Students Subscription)

**Current Configuration:**

| Resource | SKU/Size | Monthly Cost (USD) |
|----------|----------|-------------------|
| AKS Cluster | 1x Standard_B2as_v2 | $30-35 |
| PostgreSQL | B_Standard_B1ms (2GB RAM, 32GB storage) | $15 |
| Storage Account | LRS, ~10GB | $2 |
| Container Registry | Basic | $5 |
| Load Balancers | 2-3 public IPs | $20-25 each |
| Log Analytics | ~2GB ingestion/month | $5-10 |
| Application Insights | ~1GB ingestion/month | $5 |
| **Total** | | **$145-160/month** |

### Scaling Recommendations

**For 100-200 devices (Development/Testing):**
- Current configuration is adequate
- Single AKS node with autoscaling

**For 1,000 devices (Production):**
- AKS: 3-5 nodes, Standard_D2as_v5 (2 vCPU, 8GB RAM)
- PostgreSQL: GP_Standard_D2s_v3 (2 vCore, 8GB RAM)
- Redis Cache: Standard C2 for session/leaderboard caching
- Azure IoT Hub: Standard tier for MQTT at scale
- Estimated cost: $400-600/month

**For 10,000+ devices (Enterprise):**
- AKS: 10-20 nodes with node pools
- PostgreSQL: GP_Standard_D8s_v3 (8 vCore, 32GB RAM) with read replicas
- Azure Front Door for global load balancing
- Azure CDN for firmware distribution
- Estimated cost: $2,000-3,000/month

### Cost Optimization Strategies

**Infrastructure:**
```bash
# Stop AKS cluster when not in use
az aks stop --name crisp-project-aks-dev \
  --resource-group crisp-project-dev-rg

# Start cluster
az aks start --name crisp-project-aks-dev \
  --resource-group crisp-project-dev-rg
```

**Database:**
- Use B-series burstable instances for development
- Schedule backups during off-peak hours
- Reduce backup retention period (7 days minimum)

**Storage:**
- Implement lifecycle policies for old firmware (configured: delete after 90 days)
- Use Cool tier for infrequently accessed files
- Enable blob versioning only if required

**Compute:**
- Use spot instances for non-critical workloads
- Implement cluster autoscaler
- Right-size pod resource requests/limits

**Monitoring:**
- Set Log Analytics retention to 30 days (default)
- Use sampling for Application Insights in high-traffic scenarios
- Archive logs to blob storage for long-term retention

## Development Workflow

### Local Development Setup

```bash
# Clone repository with submodules
git clone --recursive https://github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games.git
cd M5StickCplus-multiplayer-crisp-games

# Start local services (PostgreSQL, EMQX, Engine, Portal)
docker-compose up -d

# Verify services
docker-compose ps

# View logs
docker-compose logs -f engine
```

**Local Service Endpoints:**
- Portal: http://localhost:8080
- Engine API: http://localhost:8081
- MQTT Broker: localhost:1883
- PostgreSQL: localhost:5432
- EMQX Dashboard: http://localhost:18083 (admin/public)

### Running Tests

```bash
cd services

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/engine/fota/...

# Verbose output
go test -v ./...
```

### Building Docker Images

```bash
# Build engine service
docker build -f services/Dockerfile.engine \
  -t crispprojectacrdev.azurecr.io/engine:latest \
  services/

# Build portal service
docker build -f services/Dockerfile.portal \
  -t crispprojectacrdev.azurecr.io/portal:latest \
  services/

# Push to ACR
az acr login --name crispprojectacrdev
docker push crispprojectacrdev.azurecr.io/engine:latest
docker push crispprojectacrdev.azurecr.io/portal:latest
```

### Deploying Code Changes

```bash
# Option 1: Full rebuild and redeploy
./deploy-local-k8s.sh

# Option 2: Rolling update (faster)
kubectl rollout restart deployment/engine -n crisp-game
kubectl rollout restart deployment/portal -n crisp-game

# Monitor rollout
kubectl rollout status deployment/engine -n crisp-game
```

### Database Migrations

```bash
# Create new migration
cd services/migrations
# Add new .sql file with timestamp: 003_add_achievements_table.sql

# Apply migrations (automatic on deployment)
# Or manually:
psql $DATABASE_URL -f migrations/003_add_achievements_table.sql
```

### Hot Reload for Local Development

```bash
cd services
./run.sh  # Uses Air for auto-restart on file changes
```

## Troubleshooting

### LoadBalancer Service Stuck in Pending State

**Symptom:** `kubectl get svc` shows `<pending>` for EXTERNAL-IP

**Root Cause:** Azure subscription has reached Public IP address quota limit (typically 3-4 for Azure Students)

**Solutions:**

1. Identify unused Public IPs:
```bash
az network public-ip list \
  --resource-group MC_crisp-project-dev-rg_crisp-project-aks-dev_polandcentral \
  --query "[].{Name:name, IP:ipAddress}" -o table
```

2. Convert service to NodePort:
```yaml
spec:
  type: NodePort
  ports:
    - port: 8081
      nodePort: 30518
```

3. Use Ingress Controller with single public IP for multiple services

4. Request quota increase via Azure Portal

### FOTA Update Fails on ESP32

**Check 1: Network Connectivity**
```bash
# From ESP32 serial monitor, verify WiFi connection
# Expected: "WiFi connected, IP: 192.168.x.x"
```

**Check 2: Engine Service Health**
```bash
kubectl logs deployment/engine -n crisp-game --tail=50
curl http://<engine-ip>:8081/health
```

**Check 3: Firmware Availability**
```bash
az storage blob list \
  --account-name crispdevkd1icl \
  --container-name firmware
```

**Check 4: Engine API Response**
```bash
curl "http://<engine-ip>:8081/api/fota/check?current_version=1.0.0"
```

### Database Connection Issues

**Symptom:** Engine pods crash with "connection refused" or timeout errors

**Diagnosis:**
```bash
# Check PostgreSQL firewall rules
az postgres flexible-server firewall-rule list \
  --resource-group crisp-project-dev-rg \
  --name crisp-project-dev-postgres

# Test connection from AKS node
kubectl run psql-test --image=postgres:15 -it --rm -- \
  psql "postgres://user:pass@host:5432/db?sslmode=require"
```

**Resolution:**
- Verify firewall rules allow AKS egress IPs
- Check PostgreSQL server status in Azure Portal
- Confirm connection string in Kubernetes secret

### Pod Crashes with OOMKilled

**Symptom:** Pods restart frequently, describe shows "OOMKilled"

```bash
kubectl describe pod <pod-name> -n crisp-game
```

**Resolution:**
```yaml
# Increase memory limits in deployment
resources:
  limits:
    memory: "512Mi"  # Increase from 256Mi
  requests:
    memory: "256Mi"
```

## Contributing

Contributions are welcome. Please follow these guidelines:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/feature-name`
3. Implement changes with appropriate tests
4. Ensure all tests pass: `go test ./...`
5. Update documentation if needed
6. Commit with descriptive messages following conventional commits
7. Push to your fork: `git push origin feature/feature-name`
8. Submit a pull request with detailed description

### Code Style

- Go: Follow official Go style guide, use `gofmt` and `golint`
- C/C++ (ESP32): Follow ESP-IDF style conventions
- Commit messages: Use conventional commits format

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) file for details.

## Acknowledgments

- crisp-game-lib by ABA Games - Game framework foundation
- M5Stack community - Hardware platform and ecosystem
- EMQX - MQTT broker implementation
- Espressif Systems - ESP-IDF framework and ESP32 chipset
- Microsoft Azure - Cloud infrastructure platform

## Support and Contact

- GitHub Issues: [Project Issues](https://github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games/issues)
- Repository: [M5StickCplus-multiplayer-crisp-games](https://github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games)
- Maintainer: [@TsybulkaM](https://github.com/TsybulkaM)
