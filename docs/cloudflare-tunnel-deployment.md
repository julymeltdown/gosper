# Cloudflare Tunnel Deployment Guide

This guide explains how to deploy Gosper with Cloudflare Tunnel when backend and frontend are exposed via separate tunnels.

## Deployment Scenario

- **Frontend**: `gosperfe.yourdomain.com` → NodePort 31987
- **Backend**: `gosperbe.yourdomain.com` → NodePort 31209

## Problem

When frontend and backend are on different Cloudflare Tunnel hostnames, the frontend cannot use internal Kubernetes service discovery to reach the backend. A 502 Bad Gateway error occurs because Cloudflare Tunnel can only see NodePort endpoints, not internal ClusterIP services.

## Solution Options

### Option 1: Path-Based Routing in Cloudflare Tunnel (Recommended)

Configure Cloudflare Tunnel to route `/api/*` requests directly to backend:

```yaml
# ~/.cloudflared/config.yml
tunnel: <your-tunnel-id>
credentials-file: /path/to/credentials.json

ingress:
  # Route /api requests to backend
  - hostname: gosperfe.yourdomain.com
    path: ^/api(/.*)?$
    service: http://localhost:31209

  # Route everything else to frontend
  - hostname: gosperfe.yourdomain.com
    service: http://localhost:31987

  # Direct backend access (optional)
  - hostname: gosperbe.yourdomain.com
    service: http://localhost:31209

  # Catch-all
  - service: http_status:404
```

**Benefits:**
- ✅ Frontend uses relative URLs (`/api/transcribe`)
- ✅ No hardcoded backend URLs
- ✅ Fully open-source friendly

### Option 2: ConfigMap with Backend URL (Current Implementation)

Use ConfigMap to inject backend URL at deployment time:

```yaml
# deploy/k8s/base/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-config
  namespace: gosper
data:
  config.js: |
    window.GOSPER_CONFIG = {
      BACKEND_URL: 'https://gosperbe.yourdomain.com'
    };
```

**Update for your deployment:**

```bash
# Edit ConfigMap with your backend URL
kubectl edit configmap nginx-config -n gosper

# Or patch it
kubectl patch configmap nginx-config -n gosper --type merge -p '{
  "data": {
    "config.js": "window.GOSPER_CONFIG = { BACKEND_URL: '\''https://your-backend.yourdomain.com'\'' };"
  }
}'

# Restart frontend to pick up changes
kubectl rollout restart deployment/gosper-fe -n gosper
```

**Benefits:**
- ✅ Works immediately without Cloudflare config changes
- ✅ Configurable per deployment
- ⚠️ Requires updating ConfigMap for each deployment

### Option 3: Default config.js Template

For open-source distributions, include a template:

```javascript
// web/config.js (default - use same origin)
window.GOSPER_CONFIG = {
  BACKEND_URL: ''  // Empty = use same origin (Ingress/path-based routing)
};
```

Users deploying with separate domains can mount their own config:

```yaml
# user-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gosper-user-config
data:
  config.js: |
    window.GOSPER_CONFIG = {
      BACKEND_URL: 'https://my-backend.example.com'
    };
---
# Mount in deployment
volumeMounts:
  - name: user-config
    mountPath: /usr/share/nginx/html/config.js
    subPath: config.js
volumes:
  - name: user-config
    configMap:
      name: gosper-user-config
```

## Deployment Steps

### 1. Deploy to Kubernetes

```bash
bash scripts/k3s/deploy.sh
```

### 2. Configure Cloudflare Tunnel

#### Method A: Path-Based Routing

```bash
# Edit Cloudflare config
nano ~/.cloudflared/config.yml

# Add path-based ingress rules (see Option 1 above)

# Restart cloudflared
sudo systemctl restart cloudflared
```

#### Method B: ConfigMap Backend URL

```bash
# Update ConfigMap with your backend URL
kubectl patch configmap nginx-config -n gosper --type merge -p '{
  "data": {
    "config.js": "window.GOSPER_CONFIG = { BACKEND_URL: '\''https://gosperbe.yourdomain.com'\'' };"
  }
}'

# Restart frontend
kubectl rollout restart deployment/gosper-fe -n gosper
```

### 3. Test

```bash
# Visit frontend
open https://gosperfe.yourdomain.com

# Upload audio file and verify transcription works
```

## Troubleshooting

### 502 Bad Gateway

1. **Check frontend logs:**
   ```bash
   kubectl logs -n gosper deployment/gosper-fe --tail=50
   ```

2. **Check backend connectivity from frontend pod:**
   ```bash
   kubectl exec -n gosper deployment/gosper-fe -- curl -v http://gosper-be/healthz
   ```

3. **Test backend directly:**
   ```bash
   curl https://gosperbe.yourdomain.com/healthz
   ```

4. **Check Cloudflare Tunnel logs:**
   ```bash
   sudo journalctl -u cloudflared -f
   ```

### CORS Errors

CORS is already configured in the backend (`cmd/server/main.go`) to allow all origins:

```go
w.Header().Set("Access-Control-Allow-Origin", "*")
```

If you still see CORS errors, check that:
- Backend is accessible: `curl https://gosperbe.yourdomain.com/healthz`
- Response includes CORS headers: Look for `Access-Control-Allow-Origin` in response

## Best Practices

1. **For Production**: Use Option 1 (Path-Based Routing) - cleaner and more maintainable
2. **For Testing**: Use Option 2 (ConfigMap) - faster to set up
3. **For Open Source**: Default to empty `BACKEND_URL` and document both options

## Security Notes

- ConfigMaps are **not encrypted** - don't store secrets there
- Backend URL is public anyway (visible in browser DevTools)
- For authentication, implement proper API keys or JWT tokens in backend
