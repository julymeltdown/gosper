# Troubleshooting

Build
- CGO errors: run `make deps`; ensure `C_INCLUDE_PATH` and `LIBRARY_PATH` point at `whisper.cpp` headers and `bindings/go/build`
- Go module downloads blocked: set proxy or run `go env -w GOPROXY=direct`

Audio
- macOS mic permissions: approve in System Settings → Privacy & Security → Microphone
- Linux: ensure Pulse/PipeWire/ALSA; list inputs and try different device ids

Inference
- Integration test fails: ensure `GOSPER_MODEL_PATH` points to an existing model file
- Slow speed: increase `--threads`; try larger model only if needed

Kubernetes
- Ingress 404: check k3s Traefik entrypoints; verify host DNS/hosts mapping
- Image pull errors: push to a registry reachable by the cluster; fix imagePullSecrets if private

