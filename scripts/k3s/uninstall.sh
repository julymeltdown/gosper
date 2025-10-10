#!/usr/bin/env bash
set -euo pipefail

if [ -f scripts/k3s/.env ]; then source scripts/k3s/.env; fi
NAMESPACE=${NAMESPACE:-handy}

echo "[kubectl] Deleting namespace $NAMESPACE"
kubectl delete namespace "$NAMESPACE" --ignore-not-found

