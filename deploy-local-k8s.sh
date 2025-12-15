#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "=== Local Kubernetes Deployment Script ==="
echo ""

check_command() {
    if ! command -v $1 &> /dev/null; then
        echo "Error: $1 is not installed"
        echo "Please install $1 first"
        exit 1
    fi
}

echo "Checking required tools..."
check_command terraform
check_command kubectl
check_command docker
check_command az

echo "Checking Azure login..."
if ! az account show &> /dev/null; then
    echo "Error: Not logged in to Azure"
    echo "Please run: az login"
    exit 1
fi

echo "Getting Azure subscription..."
SUBSCRIPTION_ID=$(az account show --query id -o tsv)
SUBSCRIPTION_NAME=$(az account show --query name -o tsv)
echo "Using subscription: $SUBSCRIPTION_NAME ($SUBSCRIPTION_ID)"
echo ""

cd terraform

echo "=== Step 1: Checking Terraform state ==="
if [ ! -d ".terraform" ]; then
    echo "Initializing Terraform..."
    terraform init
else
    echo "Terraform already initialized"
fi

if [ ! -f "terraform.tfstate" ] || [ ! -s "terraform.tfstate" ]; then
    echo ""
    echo "No existing infrastructure found."
    echo "This will create:"
    echo "  - Azure Container Registry (ACR)"
    echo "  - Azure Kubernetes Service (AKS)"
    echo "  - PostgreSQL Flexible Server"
    echo "  - Azure Storage Account"
    echo "  - Application Insights"
    echo ""
    read -p "Deploy infrastructure? This takes 10-15 minutes (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Deployment cancelled"
        exit 1
    fi
    
    echo ""
    echo "Deploying infrastructure..."
    terraform apply -auto-approve
else
    echo "Infrastructure already exists"
    read -p "Run terraform apply to update? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        terraform apply -auto-approve
    fi
fi

echo ""
echo "=== Step 2: Getting infrastructure details ==="
ACR_NAME=$(terraform output -raw acr_login_server | cut -d'.' -f1)
ACR_LOGIN_SERVER=$(terraform output -raw acr_login_server)
RESOURCE_GROUP=$(terraform output -raw resource_group_name)
AKS_CLUSTER=$(terraform output -raw aks_cluster_name)
DB_USER=$(terraform output -raw postgres_username)
DB_PASSWORD=$(terraform output -raw postgres_password)
DB_HOST=$(terraform output -raw postgres_fqdn)
STORAGE_ACCOUNT=$(terraform output -raw storage_account_name)
STORAGE_KEY=$(terraform output -raw storage_account_key)
APPINSIGHTS_KEY=$(terraform output -raw appinsights_instrumentation_key)

echo "ACR: $ACR_LOGIN_SERVER"
echo "AKS: $AKS_CLUSTER"
echo "Resource Group: $RESOURCE_GROUP"
echo ""

cd ..

echo "=== Step 3: Checking Docker images ==="
echo "Logging in to ACR..."
az acr login --name $ACR_NAME

IMAGE_EXISTS=false
if az acr repository show --name $ACR_NAME --repository engine &> /dev/null && \
   az acr repository show --name $ACR_NAME --repository portal &> /dev/null; then
    echo "Images found in ACR"
    read -p "Rebuild and push images? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        IMAGE_EXISTS=false
    else
        IMAGE_EXISTS=true
    fi
else
    echo "Images not found in ACR, will build them"
fi

if [ "$IMAGE_EXISTS" = false ]; then
    echo ""
    echo "Building Docker images..."
    
    echo "Building Engine..."
    docker build -f services/Dockerfile.engine -t ${ACR_LOGIN_SERVER}/engine:latest services/
    
    echo "Building Portal..."
    docker build -f services/Dockerfile.portal -t ${ACR_LOGIN_SERVER}/portal:latest services/
    
    echo ""
    echo "Pushing images to ACR..."
    docker push ${ACR_LOGIN_SERVER}/engine:latest
    docker push ${ACR_LOGIN_SERVER}/portal:latest
    
    echo "✓ Images pushed successfully"
fi

echo ""
echo "=== Step 4: Configuring kubectl ==="
echo "Getting AKS credentials..."
az aks get-credentials --resource-group $RESOURCE_GROUP --name $AKS_CLUSTER --overwrite-existing

# Export KUBECONFIG to use AKS config instead of k3s
export KUBECONFIG=~/.kube/config

echo "Checking cluster status..."
if ! kubectl get nodes &> /dev/null; then
    echo "Error: Cannot connect to AKS cluster"
    echo "Cluster might be stopped. Starting cluster..."
    az aks start --resource-group $RESOURCE_GROUP --name $AKS_CLUSTER
    echo "Waiting for cluster to start (this may take a few minutes)..."
    sleep 60
    az aks get-credentials --resource-group $RESOURCE_GROUP --name $AKS_CLUSTER --overwrite-existing
fi

echo "Cluster nodes:"
kubectl get nodes

echo ""
echo "=== Step 5: Creating namespace ==="
kubectl apply -f k8s/namespace.yaml

echo ""
echo "=== Step 6: Creating ConfigMap ==="
kubectl apply -f k8s/configmap.yaml

echo ""
echo "=== Step 7: Creating secrets ==="
DB_NAME="crisp_game"
DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:5432/$DB_NAME?sslmode=require"

kubectl delete secret app-secrets -n crisp-game --ignore-not-found=true

kubectl create secret generic app-secrets -n crisp-game \
  --from-literal=DATABASE_URL="$DATABASE_URL" \
  --from-literal=AZURE_STORAGE_ACCOUNT="$STORAGE_ACCOUNT" \
  --from-literal=AZURE_STORAGE_KEY="$STORAGE_KEY" \
  --from-literal=APPINSIGHTS_INSTRUMENTATIONKEY="$APPINSIGHTS_KEY"

echo "✓ Secrets created"

echo ""
echo "=== Step 8: Deploying services to Kubernetes ==="

echo "Deploying MQTT broker..."
sed "s|\${ACR_LOGIN_SERVER}|$ACR_LOGIN_SERVER|g" k8s/mqtt.yaml | kubectl apply -f -

echo "Waiting forMQTT to be ready..."
kubectl wait --for=condition=ready pod -l app=mqtt -n crisp-game --timeout=300s || true

echo "Running database migrations..."
sed "s|\${ACR_LOGIN_SERVER}|$ACR_LOGIN_SERVER|g" k8s/db-migration-job.yaml | kubectl apply -f -
kubectl wait --for=condition=complete --timeout=300s job/db-migration -n crisp-game || echo "Migration job may need manual check"

echo "Deploying Engine service..."
sed "s|\${ACR_LOGIN_SERVER}|$ACR_LOGIN_SERVER|g" k8s/engine.yaml | kubectl apply -f -

echo "Deploying Portal service..."
sed "s|\${ACR_LOGIN_SERVER}|$ACR_LOGIN_SERVER|g" k8s/portal.yaml | kubectl apply -f -

echo "Deploying HPA..."
kubectl apply -f k8s/hpa.yaml

echo "Deploying Ingress (if using)..."
kubectl apply -f k8s/ingress.yaml || echo "Note: Ingress deployment failed (may need ingress controller)"

echo ""
echo "=== Step 9: Waiting for deployments ==="
echo "This may take a few minutes..."
kubectl wait --for=condition=available deployment/engine -n crisp-game --timeout=300s || true
kubectl wait --for=condition=available deployment/portal -n crisp-game --timeout=300s || true

echo ""
echo "=== Deployment Status ==="
kubectl get pods -n crisp-game
echo ""
kubectl get svc -n crisp-game

echo ""
echo "=== Getting Service URLs ==="
echo "Waiting for LoadBalancer IPs (this may take 2-3 minutes)..."
sleep 30

PORTAL_IP=""
MQTT_IP=""
ENGINE_IP=""
for i in {1..20}; do
    PORTAL_IP=$(kubectl get svc portal -n crisp-game -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    MQTT_IP=$(kubectl get svc mqtt -n crisp-game -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    ENGINE_IP=$(kubectl get svc engine -n crisp-game -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    
    if [ -n "$PORTAL_IP" ] && [ -n "$MQTT_IP" ] && [ -n "$ENGINE_IP" ]; then
        break
    fi
    echo "Waiting for IPs... (attempt $i/20)"
    sleep 15
done

echo ""
echo "=========================================="
echo "✓ DEPLOYMENT COMPLETE!"
echo "=========================================="
echo ""
echo "Portal URL: http://$PORTAL_IP"
echo "Engine API: http://$ENGINE_IP:8080"
echo ""
echo "FOTA Update URL for ESP32:"
echo "  http://$ENGINE_IP:8080/api/fota/check"
echo ""
echo "MQTT Broker: $MQTT_IP:1883"
echo "MQTT Dashboard: http://$MQTT_IP:18083"
echo "  Default credentials: admin / public"
echo ""
echo "PostgreSQL: $DB_HOST:5432"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo ""
echo "=========================================="
echo ""
echo "Useful commands:"
echo "  View pods: kubectl get pods -n crisp-game"
echo "  View logs: kubectl logs -f -l app=engine -n crisp-game"
echo "  View logs: kubectl logs -f -l app=portal -n crisp-game"
echo "  Shell into pod: kubectl exec -it <pod-name> -n crisp-game -- /bin/sh"
echo ""
echo "To update application:"
echo "  1. Rebuild images: docker build & docker push"
echo "  2. Restart: kubectl rollout restart deployment/engine -n crisp-game"
echo "  3. Restart: kubectl rollout restart deployment/portal -n crisp-game"
echo ""
echo "To destroy everything:"
echo "  cd terraform && terraform destroy"
echo ""
