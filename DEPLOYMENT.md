# QuoteSpark Deployment Guide

This document provides comprehensive instructions for deploying the QuoteSpark application to Azure using Terraform and Docker.

## 📋 Prerequisites

### Architecture Overview

**Two-Resource-Group Design**:
- **ACR Resource Group** (`quotespark-acr-rg`): Manually created, contains only ACR
- **Application Resource Group** (`quotespark-rg`): Terraform-managed, contains application infrastructure

This design prevents ACR deletion when destroying application infrastructure.

### Required Software
- **Azure CLI** (version 2.0+)
- **Terraform** (version 1.0+)
- **Docker** (Desktop or Engine)
- **Go** (version 1.23.2+ for local development)

### Required Accounts & Services
- **Azure Subscription** with billing enabled
- **Groq API Key** for AI question generation
- **GitHub Account** (for code repository)

## 🚀 Deployment Process

### Phase 1: Azure Setup

#### Step 1: Azure CLI Authentication
```bash
# Login to Azure
az login

# Set your subscription (if you have multiple)
az account set --subscription "your-subscription-id"

# Verify current subscription
az account show
```

#### Step 2: Create Resource Group
```bash
# Create resource group
az group create --name quotespark-rg --location "East US"

# Verify creation
az group show --name quotespark-rg
```

### Phase 2: Azure Container Registry (ACR)

#### Step 1: Create Separate Resource Group for ACR
```bash
# Create a separate resource group for ACR
az group create --name quotespark-acr-rg --location "East US"
```

#### Step 2: Create ACR
```bash
# Create ACR with admin access enabled
az acr create \
  --resource-group quotespark-acr-rg \
  --name quotesparkacr \
  --sku Basic \
  --admin-enabled true
```

#### Step 2: Get ACR Credentials
```bash
# Get ACR login server
az acr show --name quotesparkacr --query loginServer --output tsv

# Get ACR credentials
az acr credential show --name quotesparkacr
```

**Note**: Save the username and password for Terraform configuration.

### Phase 3: Application Containerization

#### Step 1: Build Docker Image
```bash
# Navigate to app directory
cd app

# Build the Docker image
docker build -t quotesparkacr.azurecr.io/quotespark:latest .

# Verify image creation
docker images | grep quotespark
```

#### Step 2: Push to ACR
```bash
# Login to ACR
az acr login --name quotesparkacr

# Push the image
docker push quotesparkacr.azurecr.io/quotespark:latest

# Verify push
az acr repository list --name quotesparkacr
```

### Phase 4: Terraform Infrastructure Deployment

#### Step 1: Configure Terraform Variables
Create `infra/terraform.tfvars`:

```hcl
subscription_id     = "your-subscription-id"
resource_group_name = "quotespark-rg"
location           = "East US"
storage_account_name = "quotesparkstorage2211"
groq_api_key       = "your-groq-api-key"
acr_login_server   = "quotesparkacr.azurecr.io"
acr_username       = "quotesparkacr"
acr_password       = "your-acr-password"
container_image_url = "quotesparkacr.azurecr.io/quotespark:latest"
```

#### Step 2: Initialize Terraform
```bash
# Navigate to infra directory
cd infra

# Initialize Terraform
terraform init

# Verify initialization
terraform version
```

#### Step 3: Plan Deployment
```bash
# Create execution plan
terraform plan -out=tfplan

# Review the plan
terraform show tfplan
```

#### Step 4: Apply Configuration
```bash
# Apply the configuration
terraform apply tfplan

# Monitor deployment
terraform show
```

#### Step 5: Get Application URL
```bash
# Get the application URL
terraform output container_url

# Get other useful information
terraform output storage_account_name
terraform output container_group_fqdn
```

## 🔧 Configuration Details

### Environment Variables

The application requires these environment variables:

| Variable | Description | Source |
|----------|-------------|---------|
| `GROQ_API_KEY` | Groq API key for AI integration | Groq Console |
| `AZURE_STORAGE_ACCOUNT` | Azure Storage account name | Terraform output |
| `AZURE_STORAGE_KEY` | Azure Storage account key | Terraform output |

### Azure Resources Created

Terraform creates the following resources:

1. **Resource Group**: `quotespark-rg`
2. **Storage Account**: `quotesparkstorage2211`
3. **Storage Containers**: 
   - `quotesubmissions` (for user reflections)
   - `questions` (for daily questions)
4. **Container Group**: `quotespark-container`
5. **Network Profile**: Auto-created for container group

## 🐛 Troubleshooting

### Common Issues & Solutions

#### 1. Docker Build Failures

**Error**: `read-only file system`
```bash
# Solution: Use debian:bullseye-slim instead of scratch
FROM debian:bullseye-slim
```

**Error**: `go.mod requires go >= 1.23.2`
```bash
# Solution: Update Dockerfile
FROM golang:1.23.2 AS builder
```

#### 2. Azure SDK Compatibility Issues

**Error**: `undefined: azblob.NewServiceClientWithSharedKey`
```go
// Solution: Use new API
client, err := azblob.NewClientWithSharedKeyCredential(
    blobURL, 
    credential, 
    nil,
)
```

#### 3. Terraform State Issues

**Error**: `Resource already exists`
```bash
# Solution: Import existing resources
terraform import azurerm_resource_group.rg /subscriptions/.../resourceGroups/quotespark-rg
```

#### 4. Container Registry Issues

**Error**: `unauthorized: authentication required`
```bash
# Solution: Re-login to ACR
az acr login --name quotesparkacr
```

### Debugging Commands

#### Check Container Status
```bash
# Get container group details
az container show \
  --resource-group quotespark-rg \
  --name quotespark-container

# View container logs
az container logs \
  --resource-group quotespark-rg \
  --name quotespark-container
```

#### Check Storage Account
```bash
# List storage containers
az storage container list \
  --account-name quotesparkstorage2211 \
  --account-key $(az storage account keys list --account-name quotesparkstorage2211 --query '[0].value' -o tsv)
```

#### Check ACR Images
```bash
# List repositories
az acr repository list --name quotesparkacr

# Show image tags
az acr repository show-tags --name quotesparkacr --repository quotespark
```

## 🔄 Update Process

### Updating Application Code

1. **Modify code** in the `app/` directory
2. **Rebuild Docker image**:
   ```bash
   docker build -t quotesparkacr.azurecr.io/quotespark:latest ./app
   ```
3. **Push to ACR**:
   ```bash
   docker push quotesparkacr.azurecr.io/quotespark:latest
   ```
4. **Restart container group**:
   ```bash
   az container restart \
     --resource-group quotespark-rg \
     --name quotespark-container
   ```

### Updating Infrastructure

1. **Modify Terraform files** in `infra/`
2. **Plan changes**:
   ```bash
   terraform plan -out=tfplan
   ```
3. **Apply changes**:
   ```bash
   terraform apply tfplan
   ```

## 🧹 Cleanup

### Destroy Infrastructure
```bash
# Destroy all resources
terraform destroy

# Verify cleanup
az group show --name quotespark-rg
```

### Remove Docker Images
```bash
# Remove local images
docker rmi quotesparkacr.azurecr.io/quotespark:latest

# Remove from ACR
az acr repository delete --name quotesparkacr --image quotespark:latest
```

## 📊 Monitoring & Maintenance

### Health Checks
- **Application**: `http://your-app-url/`
- **Container Status**: Azure Portal → Container Groups
- **Storage**: Azure Portal → Storage Accounts

### Logs
- **Container Logs**: Azure Portal → Container Groups → Logs
- **Application Logs**: Built into the Go application

### Cost Optimization
- **Container Instances**: Stop when not in use
- **Storage**: Use appropriate tier (Standard LRS)
- **ACR**: Basic tier for development

## 🔒 Security Best Practices

1. **Environment Variables**: Never commit sensitive data
2. **ACR Access**: Use managed identities in production
3. **Storage Security**: Enable encryption at rest
4. **Network Security**: Use private endpoints for production

## 📞 Support

For deployment issues:
1. Check the troubleshooting section above
2. Review Azure Container Instances documentation
3. Check Terraform Azure provider documentation
4. Create an issue in the GitHub repository

---

**Note**: This deployment guide assumes a development environment. For production deployments, additional security, monitoring, and scaling considerations should be implemented. 