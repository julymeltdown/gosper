#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

if [ -f scripts/k3s/.env ]; then source scripts/k3s/.env; fi

NAMESPACE=${NAMESPACE:-handy}
DOMAIN=${DOMAIN:-handy.local}
BE_IMAGE=${BE_IMAGE:-gosper/server:local}
FE_IMAGE=${FE_IMAGE:-gosper/fe:local}
BE_NODEPORT=${BE_NODEPORT:-30080}
FE_NODEPORT=${FE_NODEPORT:-30081}

echo "[deploy] Namespace=$NAMESPACE Domain=$DOMAIN"

# Optional: build and push images if images point to local names
if [[ "$BE_IMAGE" == *":local" ]]; then
  echo "[build] Building backend image $BE_IMAGE"
  docker build -f Dockerfile.server -t "$BE_IMAGE" .
fi
if [[ "$FE_IMAGE" == *":local" ]]; then
  echo "[build] Building frontend image $FE_IMAGE"
  docker build -f Dockerfile.frontend -t "$FE_IMAGE" .
fi

if [[ -n "${REGISTRY_USER:-}" && -n "${REGISTRY_PASS:-}" ]]; then
  echo "[login] Logging into registry"
  echo "$REGISTRY_PASS" | docker login --username "$REGISTRY_USER" --password-stdin
fi

if [[ "$BE_IMAGE" == *":local" ]]; then echo "[warn] Using local tag for backend; ensure your k3s can pull it (e.g., via k3d registry)"; fi
if [[ "$FE_IMAGE" == *":local" ]]; then echo "[warn] Using local tag for frontend; ensure your k3s can pull it (e.g., via k3d registry)"; fi

export NAMESPACE DOMAIN BE_IMAGE FE_IMAGE BE_NODEPORT FE_NODEPORT

RENDER_DIR="/tmp/gosper-k8s"
rm -rf "$RENDER_DIR" && mkdir -p "$RENDER_DIR"
for f in deploy/k8s/base/*.yaml; do
  envsubst < "$f" > "$RENDER_DIR/$(basename "$f")"
done

echo "[kubectl] Applying manifests to namespace $NAMESPACE"
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f "$RENDER_DIR" -n "$NAMESPACE"

echo "[kubectl] Waiting for rollout"
kubectl rollout status deployment/gosper-be -n "$NAMESPACE" --timeout=120s || true
kubectl rollout status deployment/gosper-fe -n "$NAMESPACE" --timeout=120s || true

echo "[done] Visit http://$DOMAIN/"

