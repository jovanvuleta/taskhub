#!/bin/bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Environment parameter (defaults to production)
ENVIRONMENT=${1:-"production"}

echo -e "${GREEN}🚀 Bootstrapping Flux v2 for TaskHub (${ENVIRONMENT})${NC}"

# Check if flux CLI is installed
if ! command -v flux &> /dev/null; then
    echo -e "${RED}❌ Flux CLI not found. Installing...${NC}"
    curl -s https://fluxcd.io/install.sh | sudo bash
fi

# Check if kubectl is configured
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ kubectl not configured or cluster not accessible${NC}"
    exit 1
fi

# Set GitHub repository details
GITHUB_USER=${GITHUB_USER:-"jovanvuleta"}
GITHUB_REPO=${GITHUB_REPO:-"taskhub"}
GITHUB_TOKEN=${GITHUB_TOKEN:-""}

if [ -z "$GITHUB_TOKEN" ]; then
    echo -e "${RED}❌ GITHUB_TOKEN environment variable is required${NC}"
    echo -e "${YELLOW}Please export GITHUB_TOKEN=<your-github-token>${NC}"
    exit 1
fi

echo -e "${YELLOW}📋 Configuration:${NC}"
echo "  Environment: $ENVIRONMENT"
echo "  GitHub User: $GITHUB_USER"
echo "  GitHub Repo: $GITHUB_REPO"
echo "  Kubernetes Context: $(kubectl config current-context)"

# Create namespaces
echo -e "${GREEN}📦 Creating namespaces...${NC}"
kubectl create namespace flux-system --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace production --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace staging --dry-run=client -o yaml | kubectl apply -f -

# Bootstrap Flux
echo -e "${GREEN}🔧 Bootstrapping Flux...${NC}"
flux bootstrap github \
  --owner=$GITHUB_USER \
  --repository=$GITHUB_REPO \
  --branch=main \
  --path=./flux/clusters/${ENVIRONMENT} \
  --personal \
  --components-extra=image-reflector-controller,image-automation-controller

# Apply Flux configuration
echo -e "${GREEN}⚙️  Applying Flux configuration...${NC}"
kubectl apply -f flux/infrastructure/sources/
kubectl apply -f flux/infrastructure/controllers/
kubectl apply -f flux/apps/taskhub/

# Wait for Flux to be ready
echo -e "${GREEN}⏳ Waiting for Flux to be ready...${NC}"
flux get sources git
flux get kustomizations

echo -e "${GREEN}✅ Flux bootstrap completed!${NC}"
echo -e "${YELLOW}📖 Next steps:${NC}"
echo "1. Push changes to trigger image automation"
echo "2. Monitor deployments: flux get kustomizations --watch"
echo "3. Check logs: flux logs --follow --tail=10"