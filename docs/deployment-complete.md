# Complete System Deployment Guide

Comprehensive deployment guide for Gosper speech-to-text system across different environments.

## Table of Contents

1. [Overview](#overview)
2. [Deployment Scenarios](#deployment-scenarios)
3. [Architecture](#architecture)
4. [Local Development](#local-development)
5. [Docker Compose](#docker-compose)
6. [Kubernetes (K3s/K8s)](#kubernetes-deployment)
7. [Production Cloud](#production-cloud)
8. [CI/CD Pipeline](#cicd-pipeline)
9. [Monitoring & Logging](#monitoring--logging)
10. [Security](#security)
11. [Troubleshooting](#troubleshooting)

---

## Overview

Gosper is a distributed speech-to-text application consisting of:

- **CLI Tool**: Local binary for file/mic transcription
- **Backend Server**: HTTP API (`/api/transcribe`) running Go + whisper.cpp
- **Frontend**: Static HTML/JS UI served via nginx
- **Model Repository**: On-demand model download from Hugging Face or local cache

### Deployment Options

| Environment | Use Case | Components | Complexity |
|------------|----------|------------|-----------|
| **Local CLI** | Development, local transcription | CLI binary only | Low |
| **Docker Compose** | Testing, small deployments | Backend + Frontend | Medium |
| **K3s/K8s** | Production, scalability | Full stack + ingress | High |
| **Cloud (AWS/GCP)** | Enterprise, multi-region | Managed K8s + CDN | Very High |

---

## Architecture

### System Architecture (Full Stack)

```mermaid
flowchart TB
    subgraph Client["Client Layer"]
        CLI[CLI Tool<br/>gosper]
        WEB[Web Browser]
    end

    subgraph Ingress["Ingress Layer"]
        LB[Load Balancer/<br/>Traefik/Nginx]
    end

    subgraph Application["Application Layer"]
        FE[Frontend<br/>nginx:alpine<br/>Port 80]
        BE[Backend Server<br/>Go + whisper.cpp<br/>Port 8080]
    end

    subgraph Storage["Storage Layer"]
        CACHE[(Model Cache<br/>Persistent Volume)]
        TMP[(Temp Upload<br/>ephemeral)]
    end

    subgraph External["External Services"]
        HF[Hugging Face<br/>Model Repository]
    end

    CLI -->|Direct| BE
    WEB --> LB
    LB -->|/| FE
    LB -->|/api| BE
    FE -->|fetch /api/transcribe| BE
    BE --> CACHE
    BE --> TMP
    BE -.->|model download| HF

    style CLI fill:#e1f5ff
    style WEB fill:#e1f5ff
    style BE fill:#ffe1e1
    style FE fill:#ffe1cc
    style CACHE fill:#f0f0f0
    style HF fill:#d4edda
```

### Component Interaction

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant Backend
    participant Whisper
    participant Cache
    participant HuggingFace

    User->>Frontend: Upload WAV file
    Frontend->>Backend: POST /api/transcribe (multipart)
    Backend->>Cache: Check model exists
    alt Model not cached
        Cache-->>Backend: Not found
        Backend->>HuggingFace: Download model
        HuggingFace-->>Backend: Return model binary
        Backend->>Cache: Store model
    end
    Cache-->>Backend: Model ready
    Backend->>Whisper: Process audio
    Whisper-->>Backend: Return segments
    Backend-->>Frontend: JSON response
    Frontend-->>User: Display transcript
```

### Deployment Topology (Kubernetes)

```mermaid
graph TB
    subgraph Internet
        U[Users]
    end

    subgraph K8s_Cluster["Kubernetes Cluster"]
        subgraph Ingress_NS["Ingress Controller"]
            ING[Traefik/Nginx Ingress]
        end

        subgraph App_NS["Namespace: gosper"]
            FE_SVC[Service<br/>gosper-fe<br/>ClusterIP:80]
            BE_SVC[Service<br/>gosper-be<br/>ClusterIP:80]

            FE_DEP[Deployment<br/>gosper-fe<br/>replicas: 2]
            BE_DEP[Deployment<br/>gosper-be<br/>replicas: 3]

            FE_POD1[Pod: fe-xxx-1<br/>nginx:alpine]
            FE_POD2[Pod: fe-xxx-2<br/>nginx:alpine]

            BE_POD1[Pod: be-xxx-1<br/>server:latest]
            BE_POD2[Pod: be-xxx-2<br/>server:latest]
            BE_POD3[Pod: be-xxx-3<br/>server:latest]

            PVC[PersistentVolumeClaim<br/>model-cache<br/>10Gi]
        end
    end

    U -->|HTTPS| ING
    ING -->|/ route| FE_SVC
    ING -->|/api route| BE_SVC

    FE_SVC --> FE_DEP
    BE_SVC --> BE_DEP

    FE_DEP --> FE_POD1
    FE_DEP --> FE_POD2

    BE_DEP --> BE_POD1
    BE_DEP --> BE_POD2
    BE_DEP --> BE_POD3

    BE_POD1 -.mount.- PVC
    BE_POD2 -.mount.- PVC
    BE_POD3 -.mount.- PVC

    style U fill:#e1f5ff
    style ING fill:#fff3cd
    style FE_SVC fill:#d1ecf1
    style BE_SVC fill:#d1ecf1
    style PVC fill:#f8d7da
```

---

## Local Development

### Prerequisites

- Go 1.22+
- Make
- C compiler (gcc/clang)
- whisper.cpp submodule

### Build & Run CLI

```bash
# 1. Clone repository
git clone https://github.com/julymeltdown/go-whispher.git
cd go-whispher

# 2. Initialize whisper.cpp submodule
git submodule update --init --recursive

# 3. Build whisper static library
make deps

# 4. Build CLI binary
make build
# Output: dist/gosper

# 5. Run transcription
./dist/gosper transcribe audio.wav \
  --model ggml-tiny.en.bin \
  --lang en \
  -o transcript.txt
```

### Environment Variables

```bash
export GOSPER_MODEL=ggml-base.en.bin
export GOSPER_LANG=auto
export GOSPER_THREADS=4
export GOSPER_CACHE=$HOME/.cache/gosper/models
export GOSPER_LOG=debug
```

### Development Workflow

```mermaid
flowchart LR
    A[Edit Code] --> B[make build]
    B --> C[Run Tests]
    C --> D{Tests Pass?}
    D -->|Yes| E[Test CLI]
    D -->|No| A
    E --> F{Works?}
    F -->|Yes| G[Commit]
    F -->|No| A
    G --> H[Push]
```

---

## Docker Compose

### Quick Start

```yaml
# docker-compose.yml
version: '3.8'

services:
  backend:
    build:
      context: .
      dockerfile: Dockerfile.server
    ports:
      - "8080:8080"
    environment:
      - GOSPER_MODEL=ggml-tiny.en.bin
      - GOSPER_LANG=en
      - MODEL_BASE_URL=https://huggingface.co/ggerganov/whisper.cpp/resolve/main
    volumes:
      - model-cache:/root/.cache/gosper/models
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    build:
      context: .
      dockerfile: Dockerfile.frontend
    ports:
      - "80:80"
    depends_on:
      - backend

volumes:
  model-cache:
```

### Deploy

```bash
# Build and start
docker-compose up -d --build

# View logs
docker-compose logs -f backend

# Scale backend
docker-compose up -d --scale backend=3

# Stop
docker-compose down

# Clean volumes
docker-compose down -v
```

### Docker Compose Architecture

```mermaid
flowchart LR
    subgraph Host
        subgraph Docker_Network["Bridge Network: gosper_default"]
            FE[frontend:80<br/>nginx]
            BE1[backend-1:8080]
            BE2[backend-2:8080]
            BE3[backend-3:8080]
        end
        VOL[(Volume<br/>model-cache)]
    end

    U[User] -->|http://localhost:80| FE
    FE -->|proxy /api| BE1
    FE -->|proxy /api| BE2
    FE -->|proxy /api| BE3

    BE1 -.mount.- VOL
    BE2 -.mount.- VOL
    BE3 -.mount.- VOL
```

---

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (K3s, K8s, EKS, GKE, AKS)
- kubectl configured
- Docker registry access (Docker Hub, GHCR, ECR)

### K3s Local Setup

```bash
# Install K3s (Linux/macOS)
curl -sfL https://get.k3s.io | sh -

# Verify
kubectl get nodes

# Access config
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```

### Build & Push Images

```bash
# 1. Build images
docker build -f Dockerfile.server -t ghcr.io/yourusername/gosper-server:v1.0.0 .
docker build -f Dockerfile.frontend -t ghcr.io/yourusername/gosper-fe:v1.0.0 .

# 2. Login to registry
echo $GITHUB_TOKEN | docker login ghcr.io -u yourusername --password-stdin

# 3. Push images
docker push ghcr.io/yourusername/gosper-server:v1.0.0
docker push ghcr.io/yourusername/gosper-fe:v1.0.0
```

### Configure Deployment

```bash
# 1. Copy environment template
cp scripts/k3s/env.example scripts/k3s/.env

# 2. Edit configuration
cat > scripts/k3s/.env <<EOF
export NAMESPACE=gosper-prod
export DOMAIN=gosper.example.com

export BE_IMAGE=ghcr.io/yourusername/gosper-server:v1.0.0
export FE_IMAGE=ghcr.io/yourusername/gosper-fe:v1.0.0

# Optional: registry credentials
export REGISTRY_USER=yourusername
export REGISTRY_PASS=ghp_xxxxxxxxxxxxx
EOF
```

### Deploy to Kubernetes

```bash
# Deploy using script
bash scripts/k3s/deploy.sh

# Manual deployment
source scripts/k3s/.env
export NAMESPACE DOMAIN BE_IMAGE FE_IMAGE

# Create namespace
kubectl create namespace $NAMESPACE

# Apply manifests
for f in deploy/k8s/base/*.yaml; do
  envsubst < "$f" | kubectl apply -n $NAMESPACE -f -
done

# Wait for rollout
kubectl rollout status deployment/gosper-be -n $NAMESPACE
kubectl rollout status deployment/gosper-fe -n $NAMESPACE

# Check status
kubectl get all -n $NAMESPACE
```

### Access Application

```bash
# Get ingress IP
kubectl get ingress -n gosper-prod

# Add to /etc/hosts (if using local domain)
echo "192.168.1.100 gosper.example.com" | sudo tee -a /etc/hosts

# Visit
open http://gosper.example.com
```

### Kubernetes Deployment Flow

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant Docker as Docker
    participant Registry as Container Registry
    participant Script as deploy.sh
    participant Kubectl as kubectl
    participant K8s as Kubernetes API
    participant Pods as Pods

    Dev->>Docker: docker build (backend & frontend)
    Docker-->>Dev: Images built
    Dev->>Registry: docker push
    Registry-->>Dev: Images stored

    Dev->>Script: bash scripts/k3s/deploy.sh
    Script->>Script: Load .env
    Script->>Script: envsubst manifests
    Script->>Kubectl: kubectl create namespace
    Script->>Kubectl: kubectl apply -f manifests
    Kubectl->>K8s: Create/Update resources

    K8s->>Pods: Schedule Deployment pods
    K8s->>Pods: Create Services
    K8s->>Pods: Configure Ingress

    Pods->>Registry: Pull images
    Registry-->>Pods: Image layers

    Pods->>Pods: Container startup
    Pods-->>K8s: Ready
    K8s-->>Script: Rollout complete
    Script-->>Dev: Deployment successful
```

### Scale & Update

```bash
# Scale backend
kubectl scale deployment/gosper-be --replicas=5 -n gosper-prod

# Update image
kubectl set image deployment/gosper-be \
  server=ghcr.io/yourusername/gosper-server:v1.0.1 \
  -n gosper-prod

# Rollback
kubectl rollout undo deployment/gosper-be -n gosper-prod

# Check rollout history
kubectl rollout history deployment/gosper-be -n gosper-prod
```

### Persistent Storage

```yaml
# deploy/k8s/base/pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: model-cache
  namespace: ${NAMESPACE}
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
  storageClassName: local-path  # or your storage class
```

Mount in backend deployment:

```yaml
# deploy/k8s/base/backend.yaml (add to spec.template.spec)
volumes:
  - name: model-cache
    persistentVolumeClaim:
      claimName: model-cache

# Add to container spec
volumeMounts:
  - name: model-cache
    mountPath: /root/.cache/gosper/models
```

---

## Production Cloud

### AWS Deployment (EKS)

#### Architecture

```mermaid
flowchart TB
    subgraph Internet
        Users[Users Worldwide]
    end

    subgraph AWS["AWS Cloud"]
        subgraph Route53["Route 53"]
            DNS[DNS: gosper.com]
        end

        subgraph CloudFront["CloudFront CDN"]
            CDN[Edge Locations]
        end

        subgraph VPC["VPC: 10.0.0.0/16"]
            subgraph PublicSubnet["Public Subnet"]
                ALB[Application Load Balancer]
            end

            subgraph PrivateSubnet["Private Subnet"]
                subgraph EKS["EKS Cluster"]
                    NG[Node Group<br/>t3.large x 3]

                    subgraph Pods["Pods"]
                        FE[Frontend Pods x2]
                        BE[Backend Pods x5]
                    end
                end
            end

            subgraph Storage["Storage"]
                EFS[(EFS<br/>Model Cache)]
                S3[(S3<br/>Audio Archive)]
            end
        end

        subgraph Monitoring["Monitoring"]
            CW[CloudWatch<br/>Logs & Metrics]
            XR[X-Ray<br/>Tracing]
        end
    end

    Users --> DNS
    DNS --> CDN
    CDN --> ALB
    ALB --> FE
    ALB --> BE
    BE --> EFS
    BE --> S3
    BE --> CW
    BE --> XR

    style Users fill:#e1f5ff
    style CDN fill:#fff3cd
    style ALB fill:#d1ecf1
    style EFS fill:#f8d7da
    style S3 fill:#f8d7da
```

#### Setup Steps

```bash
# 1. Create EKS cluster
eksctl create cluster \
  --name gosper-prod \
  --region us-west-2 \
  --nodegroup-name standard-workers \
  --node-type t3.large \
  --nodes 3 \
  --nodes-min 2 \
  --nodes-max 10 \
  --managed

# 2. Configure kubectl
aws eks update-kubeconfig --region us-west-2 --name gosper-prod

# 3. Install AWS Load Balancer Controller
helm repo add eks https://aws.github.io/eks-charts
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=gosper-prod

# 4. Create EFS for model cache
aws efs create-file-system \
  --region us-west-2 \
  --performance-mode generalPurpose \
  --tags Key=Name,Value=gosper-model-cache

# 5. Install EFS CSI driver
kubectl apply -k "github.com/kubernetes-sigs/aws-efs-csi-driver/deploy/kubernetes/overlays/stable/?ref=release-1.5"

# 6. Deploy application
bash scripts/k3s/deploy.sh
```

#### Ingress Configuration (ALB)

```yaml
# deploy/k8s/aws/ingress-alb.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gosper-ingress
  namespace: gosper-prod
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
    alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-west-2:123456789:certificate/xxx
    alb.ingress.kubernetes.io/ssl-redirect: '443'
spec:
  rules:
    - host: gosper.example.com
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: gosper-be
                port:
                  number: 80
          - path: /
            pathType: Prefix
            backend:
              service:
                name: gosper-fe
                port:
                  number: 80
```

### GCP Deployment (GKE)

```bash
# 1. Create GKE cluster
gcloud container clusters create gosper-prod \
  --region us-central1 \
  --num-nodes 3 \
  --machine-type n1-standard-2 \
  --enable-autoscaling \
  --min-nodes 2 \
  --max-nodes 10

# 2. Get credentials
gcloud container clusters get-credentials gosper-prod --region us-central1

# 3. Deploy
bash scripts/k3s/deploy.sh
```

### Azure Deployment (AKS)

```bash
# 1. Create resource group
az group create --name gosper-rg --location eastus

# 2. Create AKS cluster
az aks create \
  --resource-group gosper-rg \
  --name gosper-prod \
  --node-count 3 \
  --node-vm-size Standard_D2s_v3 \
  --enable-managed-identity \
  --generate-ssh-keys

# 3. Get credentials
az aks get-credentials --resource-group gosper-rg --name gosper-prod

# 4. Deploy
bash scripts/k3s/deploy.sh
```

---

## CI/CD Pipeline

### GitHub Actions Workflow

```yaml
# .github/workflows/deploy.yml
name: Build and Deploy

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_PREFIX: ${{ github.repository_owner }}/gosper

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          submodules: recursive

      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Run tests
        run: |
          make test

  build-and-push:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v3
        with:
          submodules: recursive

      - name: Login to GitHub Container Registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v4
        with:
          images: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-server
            ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-frontend
          tags: |
            type=ref,event=branch
            type=sha,prefix={{branch}}-
            type=semver,pattern={{version}}

      - name: Build and push backend
        uses: docker/build-push-action@v4
        with:
          context: .
          file: Dockerfile.server
          push: true
          tags: ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-server:${{ github.sha }}

      - name: Build and push frontend
        uses: docker/build-push-action@v4
        with:
          context: .
          file: Dockerfile.frontend
          push: true
          tags: ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-frontend:${{ github.sha }}

  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v3

      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          method: kubeconfig
          kubeconfig: ${{ secrets.KUBE_CONFIG }}

      - name: Deploy to Kubernetes
        env:
          NAMESPACE: gosper-prod
          DOMAIN: gosper.example.com
          BE_IMAGE: ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-server:${{ github.sha }}
          FE_IMAGE: ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-frontend:${{ github.sha }}
        run: |
          bash scripts/k3s/deploy.sh
```

### CI/CD Flow

```mermaid
flowchart LR
    A[Git Push] --> B[GitHub Actions]
    B --> C{Tests Pass?}
    C -->|No| D[Fail Build]
    C -->|Yes| E[Build Docker Images]
    E --> F[Push to Registry]
    F --> G{Main Branch?}
    G -->|No| H[Skip Deploy]
    G -->|Yes| I[Deploy to K8s]
    I --> J[Health Check]
    J --> K{Healthy?}
    K -->|No| L[Rollback]
    K -->|Yes| M[Complete]

    style A fill:#e1f5ff
    style D fill:#f8d7da
    style L fill:#f8d7da
    style M fill:#d4edda
```

---

## Monitoring & Logging

### Prometheus + Grafana

```yaml
# deploy/k8s/monitoring/prometheus.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    scrape_configs:
      - job_name: 'gosper-backend'
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_label_app]
            action: keep
            regex: gosper-be
```

### Metrics Endpoints

Add to backend server:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

func main() {
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    // ... existing handlers
}
```

### Log Aggregation

```mermaid
flowchart LR
    BE[Backend Pods] --> FB[Fluent Bit]
    FE[Frontend Pods] --> FB
    FB --> ES[Elasticsearch]
    ES --> KB[Kibana]
    KB --> USER[DevOps Team]

    style BE fill:#ffe1e1
    style FE fill:#ffe1cc
    style ES fill:#f0f0f0
    style KB fill:#d1ecf1
```

---

## Security

### Best Practices

1. **Image Scanning**
   ```bash
   # Scan images for vulnerabilities
   trivy image ghcr.io/yourusername/gosper-server:v1.0.0
   ```

2. **Network Policies**
   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: backend-policy
   spec:
     podSelector:
       matchLabels:
         app: gosper-be
     policyTypes:
       - Ingress
     ingress:
       - from:
         - podSelector:
             matchLabels:
               app: gosper-fe
         ports:
         - protocol: TCP
           port: 8080
   ```

3. **Secrets Management**
   ```bash
   # Create secret for model download credentials
   kubectl create secret generic huggingface-token \
     --from-literal=token=hf_xxxxxxxxxxxxx \
     -n gosper-prod
   ```

4. **RBAC**
   ```yaml
   apiVersion: rbac.authorization.k8s.io/v1
   kind: Role
   metadata:
     name: gosper-pod-reader
   rules:
   - apiGroups: [""]
     resources: ["pods", "pods/log"]
     verbs: ["get", "list"]
   ```

---

## Troubleshooting

### Common Issues

#### Pods Not Starting

```bash
# Check pod status
kubectl get pods -n gosper-prod

# Describe pod
kubectl describe pod gosper-be-xxx -n gosper-prod

# View logs
kubectl logs gosper-be-xxx -n gosper-prod --tail=100

# Check events
kubectl get events -n gosper-prod --sort-by='.lastTimestamp'
```

#### Image Pull Errors

```bash
# Verify image exists
docker pull ghcr.io/yourusername/gosper-server:v1.0.0

# Create image pull secret
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=yourusername \
  --docker-password=$GITHUB_TOKEN \
  -n gosper-prod

# Add to deployment
spec:
  template:
    spec:
      imagePullSecrets:
        - name: ghcr-secret
```

#### Ingress Not Working

```bash
# Check ingress
kubectl get ingress -n gosper-prod
kubectl describe ingress gosper-ingress -n gosper-prod

# Check ingress controller logs
kubectl logs -n kube-system -l app=traefik

# Test service directly
kubectl port-forward svc/gosper-be 8080:80 -n gosper-prod
curl http://localhost:8080/healthz
```

#### Model Download Failures

```bash
# Check network policies
kubectl get networkpolicy -n gosper-prod

# Test connectivity from pod
kubectl exec -it gosper-be-xxx -n gosper-prod -- \
  curl -I https://huggingface.co

# Check environment variables
kubectl exec -it gosper-be-xxx -n gosper-prod -- env | grep GOSPER
```

### Debug Flow

```mermaid
flowchart TD
    A[Issue Detected] --> B{Pod Running?}
    B -->|No| C[Check Events]
    B -->|Yes| D{Service Accessible?}

    C --> E[Check Image Pull]
    C --> F[Check Resources]
    C --> G[Check Config]

    D -->|No| H[Check Service]
    D -->|Yes| I{Ingress Working?}

    H --> J[Check Selectors]
    H --> K[Check Endpoints]

    I -->|No| L[Check Ingress Config]
    I -->|Yes| M[Check Application Logs]

    M --> N[Fix Application Bug]
    L --> O[Fix Ingress]
    K --> P[Fix Service]
    G --> Q[Fix Deployment]

    style A fill:#f8d7da
    style N fill:#d4edda
    style O fill:#d4edda
    style P fill:#d4edda
    style Q fill:#d4edda
```

---

## Summary

### Deployment Checklist

- [ ] Choose deployment environment
- [ ] Build and test locally
- [ ] Build Docker images
- [ ] Push to container registry
- [ ] Configure environment variables
- [ ] Deploy to target environment
- [ ] Verify health checks
- [ ] Configure monitoring
- [ ] Set up logging
- [ ] Configure backups
- [ ] Document access credentials
- [ ] Test disaster recovery

### Quick Reference

| Task | Command |
|------|---------|
| Build CLI | `make build` |
| Run tests | `make test` |
| Build backend image | `docker build -f Dockerfile.server -t gosper/server .` |
| Deploy to K8s | `bash scripts/k3s/deploy.sh` |
| Scale deployment | `kubectl scale deployment/gosper-be --replicas=5` |
| View logs | `kubectl logs -f deployment/gosper-be` |
| Update image | `kubectl set image deployment/gosper-be server=new-image:tag` |
| Rollback | `kubectl rollout undo deployment/gosper-be` |

---

For more details, see:
- [K3s Deployment](./deployment-k3s.md)
- [Configuration](./configuration.md)
- [Troubleshooting](./troubleshooting.md)
- [Architecture](./architecture.md)
