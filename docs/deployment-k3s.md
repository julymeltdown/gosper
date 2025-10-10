# Deployment (k3s)

Artifacts
- Backend image: `Dockerfile.server` (Go + whisper)
- Frontend image: `Dockerfile.frontend` (nginx serving `web/`)
- Manifests: `deploy/k8s/base` (Namespace, Deployments/Services, Ingress)
- Scripts: `scripts/k3s/deploy.sh`, `scripts/k3s/uninstall.sh`, `scripts/k3s/env.example`

Prereqs
- k3s cluster (or k3d)
- kubectl on PATH
- Docker for building images

Steps
1) Build images (or set BE_IMAGE/FE_IMAGE to registry tags in `.env`):
   - `docker build -f Dockerfile.server -t handy/server:local .`
   - `docker build -f Dockerfile.frontend -t handy/fe:local .`
2) Configure:
   - `cp scripts/k3s/env.example scripts/k3s/.env`
   - Edit NAMESPACE, DOMAIN, BE_IMAGE, FE_IMAGE
3) Deploy:
   - `bash scripts/k3s/deploy.sh`
4) Open `http://$DOMAIN/` and use the UI to upload a WAV

Ingress
- Traefik entrypoint `web` used by default; set DNS or `/etc/hosts` for `${DOMAIN}`

Models
- Server downloads models on demand from `MODEL_BASE_URL` unless a local path is provided via request or env

Cleanup
- `bash scripts/k3s/uninstall.sh`

## Mermaid Diagrams

Traffic Flow (User → Ingress → Services → Pods)

```mermaid
flowchart LR
  U[User Browser] --> DNS[(DNS/Hosts)]
  DNS --> T[Traefik Ingress (k3s)]
  T -->|/| SFE[Service handy-fe]
  T -->|/api| SBE[Service handy-be]
  SFE --> PFE[Deployment handy-fe → Pod/nginx]
  SBE --> PBE[Deployment handy-be → Pod/server]
  PBE --> SRV[Go server /api/transcribe]
  SRV --> WH[Whisper adapter]
  WH --> CACHE[(Model Cache)]
  SRV -. optional .-> HF[Hugging Face Models]
```

Kubernetes Object Relationships

```mermaid
flowchart TB
  NS[Namespace ${NAMESPACE}] --- ING[Ingress handy-ingress]
  NS --- FEDEP[Deployment handy-fe]
  NS --- BEDEP[Deployment handy-be]
  FEDEP --> FEPO[ReplicaSet/Pods]
  BEDEP --> BEPO[ReplicaSet/Pods]
  NS --- FESVC[Service handy-fe]
  NS --- BESVC[Service handy-be]
  FESVC -->|selector app=handy-fe| FEPO
  BESVC -->|selector app=handy-be| BEPO
  ING -->|/ → handy-fe:80| FESVC
  ING -->|/api → handy-be:80| BESVC
```

Deploy Script Sequence (envsubst + kubectl apply)

```mermaid
sequenceDiagram
  participant Dev as Developer
  participant Sh as scripts/k3s/deploy.sh
  participant DO as Docker
  participant K as kubectl
  participant K3 as k3s Cluster

  Dev->>Sh: set .env (NAMESPACE, DOMAIN, BE_IMAGE, FE_IMAGE)
  Sh->>DO: docker build -f Dockerfile.server (optional)
  Sh->>DO: docker build -f Dockerfile.frontend (optional)
  Sh->>Sh: envsubst deploy/k8s/base/*.yaml → /tmp/handy-k8s
  Sh->>K: kubectl create ns | apply ns
  Sh->>K: kubectl apply -f /tmp/handy-k8s
  K->>K3: Create/Update Ingress, Services, Deployments
  Sh->>K: kubectl rollout status deployments
  K3-->>Sh: Deployed and ready
```
