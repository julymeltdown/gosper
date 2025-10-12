# Troubleshooting Guide

Common issues and solutions for Gosper.

## Table of Contents

- [Build Issues](#build-issues)
- [Runtime Issues](#runtime-issues)
- [Audio Issues](#audio-issues)
- [Model Issues](#model-issues)
- [Deployment Issues](#deployment-issues)
- [Platform-Specific Issues](#platform-specific-issues)
- [Performance Issues](#performance-issues)

## Build Issues

### CGO Errors

#### Error: `gcc: command not found`

**Cause**: C compiler not installed.

**Solution**:

**Linux**:
```bash
# Ubuntu/Debian
sudo apt-get install build-essential

# Fedora/RHEL
sudo dnf install gcc gcc-c++
```

**macOS**:
```bash
xcode-select --install
```

**Windows**:
- Install [MSYS2](https://www.msys2.org/)
- Add GCC to PATH

---

#### Error: `undefined reference to whisper_*`

**Cause**: whisper.cpp not built or linker can't find library.

**Solution**:
```bash
# Build whisper.cpp first
make deps

# If still failing, set library path
export LIBRARY_PATH="$(pwd)/whisper.cpp/bindings/go/build"
export LD_LIBRARY_PATH="$(pwd)/whisper.cpp/bindings/go/build"  # Linux
export DYLD_LIBRARY_PATH="$(pwd)/whisper.cpp/bindings/go/build"  # macOS

# Then build
make build
```

---

#### Error: `cannot find whisper.h`

**Cause**: Include path not set for CGO.

**Solution**:
```bash
export C_INCLUDE_PATH="$(pwd)/whisper.cpp:$(pwd)/whisper.cpp/bindings/go/build"
export CPLUS_INCLUDE_PATH="$(pwd)/whisper.cpp:$(pwd)/whisper.cpp/bindings/go/build"

make build
```

---

#### Error: `missing go.sum entry for module`

**Cause**: Go dependencies not synced.

**Solution**:
```bash
go mod tidy
go mod download
make build
```

### Build Tag Errors

#### Error: `undefined: cmd` when running binary

**Cause**: Missing `cli` build tag.

**Solution**:
```bash
# Include cli tag
go build -tags "cli whisper" -o gosper ./cmd/gosper
```

#### Error: `undefined: NewWhisper`

**Cause**: Missing `whisper` build tag.

**Solution**:
```bash
# Include whisper tag
go build -tags "cli whisper" -o gosper ./cmd/gosper
```

## Runtime Issues

### Model Issues

#### Error: `model not found: ggml-tiny.en.bin`

**Cause**: Model doesn't exist in cache or specified path.

**Solution**:

**Option 1**: Let Gosper download automatically:
```bash
# Will download to cache on first use
gosper transcribe audio.mp3 --model ggml-tiny.en.bin
```

**Option 2**: Download manually:
```bash
mkdir -p ~/.cache/gosper
curl -L -o ~/.cache/gosper/ggml-tiny.en.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin

gosper transcribe audio.mp3 --model ggml-tiny.en.bin
```

**Option 3**: Use absolute path:
```bash
gosper transcribe audio.mp3 --model /path/to/ggml-tiny.en.bin
```

---

#### Error: `failed to load model`

**Cause**: Corrupted or incompatible model file.

**Solution**:
```bash
# Remove corrupted model
rm ~/.cache/gosper/ggml-tiny.en.bin

# Re-download
curl -L -o ~/.cache/gosper/ggml-tiny.en.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin

# Verify checksum
sha256sum ~/.cache/gosper/ggml-tiny.en.bin
```

### Audio Format Issues

#### Error: `unsupported audio format: .m4a`

**Cause**: Only WAV and MP3 are supported.

**Solution**: Convert to supported format:
```bash
# M4A → MP3
ffmpeg -i audio.m4a audio.mp3

# FLAC → WAV
ffmpeg -i audio.flac audio.wav

# OGG → MP3
ffmpeg -i audio.ogg audio.mp3

# Then transcribe
gosper transcribe audio.mp3
```

---

#### Error: `mp3: file too large (250 MB, max 200 MB)`

**Cause**: MP3 files are limited to 200 MB for memory protection.

**Solution**:

**Option 1**: Convert to WAV (no limit):
```bash
ffmpeg -i large-audio.mp3 large-audio.wav
gosper transcribe large-audio.wav
```

**Option 2**: Compress MP3:
```bash
ffmpeg -i large-audio.mp3 -ab 128k compressed-audio.mp3
gosper transcribe compressed-audio.mp3
```

**Option 3**: Split into chunks:
```bash
# Split into 10-minute chunks
ffmpeg -i large-audio.mp3 -f segment -segment_time 600 -c copy chunk%03d.mp3

# Transcribe each
for chunk in chunk*.mp3; do
    gosper transcribe "$chunk" -o "${chunk%.mp3}.txt"
done
```

---

#### Error: `mp3: invalid format`

**Cause**: Corrupted or non-MP3 file with `.mp3` extension.

**Solution**:
```bash
# Verify file is actually MP3
file audio.mp3
# Should show: "MPEG ADTS, layer III"

# Re-encode if corrupted
ffmpeg -i audio.mp3 -codec:a libmp3lame audio-fixed.mp3
gosper transcribe audio-fixed.mp3
```

---

#### Error: `wav: unsupported format code`

**Cause**: WAV file uses unsupported codec (only PCM16 and Float32 supported).

**Solution**:
```bash
# Convert to PCM16
ffmpeg -i audio.wav -acodec pcm_s16le audio-pcm.wav
gosper transcribe audio-pcm.wav
```

### Transcription Issues

#### Error: `transcription failed: context deadline exceeded`

**Cause**: Transcription taking too long (timeout).

**Solution**:
- Use smaller model (e.g., `ggml-tiny.en.bin`)
- Increase thread count: `--threads 8`
- Process shorter audio clips
- Add more CPU resources

---

#### Poor Transcription Quality

**Symptoms**: Inaccurate or nonsensical transcripts.

**Solutions**:

1. **Use larger model**:
```bash
# Base (better than tiny)
gosper transcribe audio.mp3 --model ggml-base.en.bin

# Medium (high accuracy)
gosper transcribe audio.mp3 --model ggml-medium.en.bin
```

2. **Specify language explicitly**:
```bash
# Don't use 'auto' if you know the language
gosper transcribe audio.mp3 --lang en
```

3. **Improve audio quality**:
```bash
# Remove noise with ffmpeg
ffmpeg -i noisy.mp3 -af "highpass=f=200, lowpass=f=3000" clean.mp3
gosper transcribe clean.mp3
```

4. **Check audio sample rate**:
```bash
# Verify sample rate is reasonable (8kHz - 48kHz)
ffmpeg -i audio.mp3 2>&1 | grep "Audio:"
```

## Audio Issues

### Microphone Not Found (CLI)

#### macOS

**Error**: `audio device not found` or no prompt for microphone access.

**Solution**:
1. System Settings → Privacy & Security → Microphone
2. Enable microphone for Terminal or iTerm
3. Restart terminal
4. Run again: `gosper record --duration 10s`

**If permission prompt doesn't appear**:
```bash
# Reset permissions database
tccutil reset Microphone
# Then run gosper again
```

---

#### Linux

**Error**: `audio device not found` or `failed to initialize device`.

**Solution**:

**Check PulseAudio**:
```bash
# List devices
pactl list sources short

# Test recording
arecord -l
```

**Add user to audio group**:
```bash
sudo usermod -a -G audio $USER
# Log out and back in
```

**Install audio libraries**:
```bash
# Ubuntu/Debian
sudo apt-get install libasound2-dev libpulse-dev

# Fedora
sudo dnf install alsa-lib-devel pulseaudio-libs-devel
```

---

#### Windows

**Error**: Microphone not detected.

**Solution**:
1. Settings → Privacy → Microphone
2. Allow apps to access microphone
3. Allow desktop apps to access microphone
4. Restart terminal
5. Run as administrator if needed

### Device Selection Issues

#### Error: `device "USB Mic" not found`

**Cause**: Device name doesn't match exactly.

**Solution**:
```bash
# List all devices
gosper devices list

# Use exact device ID from list
gosper record --device "hw:1,0" --duration 10s

# Or use fuzzy match (partial name works)
gosper record --device "USB" --duration 10s
```

## Model Issues

### Download Failures

#### Error: `failed to download model: connection timeout`

**Cause**: Network issue or Hugging Face unavailable.

**Solution**:

**Option 1**: Retry with manual download:
```bash
# Download with curl (supports resume)
curl -C - -L -o ggml-tiny.en.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin

# Move to cache
mkdir -p ~/.cache/gosper
mv ggml-tiny.en.bin ~/.cache/gosper/
```

**Option 2**: Use custom mirror:
```bash
export MODEL_BASE_URL=https://your-mirror.com/models/
gosper transcribe audio.mp3
```

### Model Compatibility

#### Error: `incompatible model version`

**Cause**: Model built for different whisper.cpp version.

**Solution**:
```bash
# Rebuild whisper.cpp
cd whisper.cpp
git pull origin master
cd ..
make deps

# Re-download models
rm -rf ~/.cache/gosper
gosper transcribe audio.mp3  # Will re-download
```

## Deployment Issues

### Docker Issues

#### Error: `model download failed` in container

**Cause**: Container has no internet or can't write to cache.

**Solution**:

**Pre-download models**:
```bash
# Build custom image with models
FROM gosper/server:latest
RUN mkdir -p /root/.cache/gosper && \
    curl -L -o /root/.cache/gosper/ggml-tiny.en.bin \
    https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin
```

**Or mount volume**:
```bash
# Download locally first
mkdir -p ./models
curl -L -o ./models/ggml-tiny.en.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin

# Mount in container
docker run -v $(pwd)/models:/root/.cache/gosper \
  -p 8080:8080 gosper/server:latest
```

---

#### Error: `connection refused` to API

**Cause**: Port not exposed or server not started.

**Solution**:
```bash
# Check container is running
docker ps

# Check logs
docker logs <container-id>

# Verify port mapping
docker port <container-id>

# Should show: 8080/tcp -> 0.0.0.0:8080
```

### Kubernetes Issues

#### Pod CrashLoopBackOff

**Symptom**: Backend pod keeps restarting.

**Diagnosis**:
```bash
kubectl get pods -n gosper
kubectl logs -f deployment/gosper-be -n gosper
kubectl describe pod <pod-name> -n gosper
```

**Common Causes**:

**1. Image pull error**:
```bash
# Check events
kubectl describe pod <pod-name> -n gosper
# Look for "ErrImagePull" or "ImagePullBackOff"

# Solution: Re-import image to k3s
docker save gosper/server:local | sudo k3s ctr images import -
```

**2. OOM (Out of Memory)**:
```bash
# Check if killed by OOM
kubectl describe pod <pod-name> -n gosper | grep -i oom

# Solution: Increase memory limits
# Edit deploy/k8s/base/backend-deployment.yaml
resources:
  limits:
    memory: "4Gi"  # Increase from 2Gi
```

**3. Model download failed**:
```bash
# Check logs
kubectl logs <pod-name> -n gosper

# Solution: Pre-load models in image or use PVC
```

---

#### Service Not Accessible

**Error**: Can't reach `http://<NODE_IP>:<NODEPORT>/api/transcribe`

**Diagnosis**:
```bash
# Check service
kubectl get svc -n gosper

# Should show TYPE=NodePort with ports like 31209:8080

# Check pod is running
kubectl get pods -n gosper

# Test from inside cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://gosper-be-svc.gosper.svc.cluster.local:8080/health
```

**Solutions**:

**1. NodePort range issue**:
```bash
# Verify NodePort is in allowed range (30000-32767)
# Edit scripts/k3s/.env
export BE_NODEPORT=31209  # Must be 30000-32767
```

**2. Firewall blocking**:
```bash
# Allow NodePort through firewall
sudo ufw allow 31209/tcp
```

**3. Wrong IP**:
```bash
# Get node IP
kubectl get nodes -o wide
# Use INTERNAL-IP or EXTERNAL-IP, not pod IP
```

## Platform-Specific Issues

### macOS

#### Apple Silicon (M1/M2/M3) Issues

**Issue**: Slow performance on Apple Silicon.

**Cause**: whisper.cpp not using Metal acceleration.

**Solution**:
```bash
# Build with Metal support
cd whisper.cpp
make clean
WHISPER_METAL=1 make
cd ..
make build
```

---

#### Gatekeeper Blocking Binary

**Error**: "gosper cannot be opened because it is from an unidentified developer"

**Solution**:
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine ./dist/gosper

# Or allow in System Settings
# System Settings → Privacy & Security → Allow anyway
```

### Linux

#### GLIBC Version Issues

**Error**: `version GLIBC_2.32 not found`

**Cause**: Binary built on newer Linux than runtime.

**Solution**:

**Option 1**: Build from source on target system:
```bash
git clone https://github.com/yourusername/gosper.git
cd gosper
make deps
make build
```

**Option 2**: Use Docker:
```bash
docker run -p 8080:8080 gosper/server:latest
```

---

#### SELinux Issues

**Error**: Permission denied errors on Fedora/RHEL.

**Solution**:
```bash
# Check SELinux status
getenforce

# Temporarily disable
sudo setenforce 0

# Or create policy
sudo ausearch -c 'gosper' --raw | audit2allow -M gosper-policy
sudo semodule -i gosper-policy.pp
```

### Windows

#### Path Length Issues

**Error**: "file name too long" when building.

**Cause**: Windows MAX_PATH limit (260 characters).

**Solution**:
```powershell
# Enable long paths in Windows 10/11
New-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\FileSystem" `
  -Name "LongPathsEnabled" -Value 1 -PropertyType DWORD -Force

# Or build in shorter path
cd C:\gosper
make build
```

## Performance Issues

### Slow Transcription

**Symptoms**: Transcription takes much longer than expected.

**Solutions**:

**1. Use more threads**:
```bash
# Set to number of CPU cores
gosper transcribe audio.mp3 --threads 8
```

**2. Use smaller model**:
```bash
# Tiny (5x faster than medium)
gosper transcribe audio.mp3 --model ggml-tiny.en.bin
```

**3. Check CPU usage**:
```bash
# Monitor during transcription
top
# Should see high CPU usage on gosper process
```

**4. Disable other processes**:
```bash
# Free up CPU/memory
# Close unnecessary applications
```

### High Memory Usage

**Symptoms**: Server using excessive RAM or OOM killed.

**Diagnosis**:
```bash
# Check memory usage
free -h
docker stats  # For containers
kubectl top pods  # For k8s
```

**Solutions**:

**1. Use smaller model**:
```bash
export GOSPER_MODEL=ggml-tiny.en.bin  # 500 MB RAM
# Instead of ggml-large-v3.bin  # 6 GB RAM
```

**2. Limit concurrent requests**:
```bash
# Add nginx reverse proxy with queue
# Or use k8s resource limits
```

**3. Process shorter audio**:
```bash
# Split large files into chunks
ffmpeg -i large.mp3 -f segment -segment_time 600 chunk%03d.mp3
```

## Integration Test Issues

### Tests Failing

#### Error: `model not found` in integration tests

**Cause**: `GOSPER_INTEGRATION=1` set but no model available.

**Solution**:
```bash
# Download model
curl -L -o ggml-tiny.en.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin

# Set path
export GOSPER_MODEL_PATH=$(pwd)/ggml-tiny.en.bin
export GOSPER_INTEGRATION=1

# Run tests
make itest
```

---

#### Tests Timeout

**Cause**: Transcription taking too long in CI.

**Solution**:
```bash
# Use tiny model for tests
export GOSPER_MODEL=ggml-tiny.en.bin

# Or increase timeout in test code
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
```

## Getting Further Help

If you're still experiencing issues:

1. **Check logs** with `GOSPER_LOG=debug`:
```bash
export GOSPER_LOG=debug
gosper transcribe audio.mp3
```

2. **Search existing issues**:
   - [GitHub Issues](https://github.com/yourusername/gosper/issues)

3. **Open a new issue** with:
   - Operating system and version
   - Go version (`go version`)
   - Gosper version (`gosper --version`)
   - Full error message
   - Steps to reproduce
   - Logs with `GOSPER_LOG=debug`

4. **Check whisper.cpp issues**:
   - Many issues are upstream: https://github.com/ggerganov/whisper.cpp/issues

## Next Steps

- **[Configuration Guide](CONFIGURATION.md)** - Advanced settings
- **[API Reference](API.md)** - HTTP API documentation
- **[Build Guide](BUILD.md)** - Building from source
- **[Contributing](CONTRIBUTING.md)** - Development guidelines
