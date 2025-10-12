# HTTP API Reference

Complete reference for Gosper's HTTP API endpoints.

## Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Endpoints](#endpoints)
  - [POST /api/transcribe](#post-apitranscribe)
  - [GET /health](#get-health)
- [Request Format](#request-format)
- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Client Examples](#client-examples)

## Overview

**Base URL**: `http://your-server:8080`

**Content Type**: `multipart/form-data` for file uploads

**Response Format**: `application/json`

**Max Request Size**: 200 MB (for MP3 files), unlimited for WAV

## Authentication

Currently, Gosper API does not require authentication. For production deployments:

**Recommended Approaches**:
1. **API Gateway** - Add auth layer (Kong, Traefik, nginx)
2. **VPN/Private Network** - Restrict network access
3. **Cloudflare Tunnel** - Use Cloudflare Access for authentication
4. **Custom Middleware** - Add API key validation

**Example with nginx** (API key validation):
```nginx
location /api/ {
    if ($http_x_api_key != "your-secret-key") {
        return 401;
    }
    proxy_pass http://gosper-backend:8080;
}
```

## Endpoints

### POST /api/transcribe

Transcribe an audio file to text.

**Request**

```http
POST /api/transcribe HTTP/1.1
Host: localhost:8080
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary
```

**Form Data Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `audio` | file | ✅ Yes | Audio file (WAV or MP3) |
| `model` | string | ❌ No | Model name (default: `ggml-tiny.en.bin`) |
| `lang` | string | ❌ No | Language code or `auto` (default: `auto`) |

**Supported Audio Formats**:
- **WAV**: `.wav`, `.Wave`, `.WAV`
- **MP3**: `.mp3`, `.MP3` (max 200 MB)

**Supported Languages** (multilingual models):
- `auto` - Automatic detection
- `en` - English
- `es` - Spanish
- `fr` - French
- `de` - German
- `ja` - Japanese
- `zh` - Chinese
- ...and 90+ more

**Example Request** (curl):
```bash
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@recording.mp3" \
  -F "lang=auto"
```

**Example Request** (with specific model):
```bash
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@meeting.wav" \
  -F "model=ggml-base.en.bin" \
  -F "lang=en"
```

**Response** (Success - 200 OK):
```json
{
  "text": "This is the complete transcribed text from your audio file.",
  "language": "en",
  "duration_ms": 5420,
  "segments": [
    {
      "start_ms": 0,
      "end_ms": 2800,
      "text": "This is the complete transcribed text"
    },
    {
      "start_ms": 2800,
      "end_ms": 5420,
      "text": " from your audio file."
    }
  ]
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `text` | string | Complete transcription text |
| `language` | string | Detected or specified language code |
| `duration_ms` | int | Processing time in milliseconds |
| `segments` | array | Individual speech segments with timestamps |
| `segments[].start_ms` | int | Segment start time (milliseconds) |
| `segments[].end_ms` | int | Segment end time (milliseconds) |
| `segments[].text` | string | Segment text |

**Error Response** (400 Bad Request):
```json
{
  "error": "audio file is required"
}
```

**Error Response** (500 Internal Server Error):
```json
{
  "error": "transcription failed: model not found"
}
```

### GET /health

Health check endpoint for monitoring and load balancers.

**Request**:
```http
GET /health HTTP/1.1
Host: localhost:8080
```

**Response** (200 OK):
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

**Example**:
```bash
curl http://localhost:8080/health
```

**Use Cases**:
- Kubernetes liveness/readiness probes
- Load balancer health checks
- Monitoring systems

## Request Format

### Multipart Form Data

Audio files must be sent as `multipart/form-data`:

**Structure**:
```
POST /api/transcribe
Content-Type: multipart/form-data; boundary=----Boundary

------Boundary
Content-Disposition: form-data; name="audio"; filename="recording.mp3"
Content-Type: audio/mpeg

[binary audio data]
------Boundary
Content-Disposition: form-data; name="lang"

auto
------Boundary--
```

### File Size Limits

| Format | Maximum Size | Reason |
|--------|--------------|--------|
| **WAV** | Unlimited | Efficient streaming decode |
| **MP3** | 200 MB | Memory protection (~600 MB decoded) |

**For Large Files**:
```bash
# Convert MP3 > 200MB to WAV
ffmpeg -i large-audio.mp3 large-audio.wav
curl -F "audio=@large-audio.wav" http://localhost:8080/api/transcribe
```

## Response Format

### Success Response

**Structure**:
```json
{
  "text": string,
  "language": string,
  "duration_ms": integer,
  "segments": [
    {
      "start_ms": integer,
      "end_ms": integer,
      "text": string
    }
  ]
}
```

**Example** (Short Audio):
```json
{
  "text": "Hello world.",
  "language": "en",
  "duration_ms": 856,
  "segments": [
    {
      "start_ms": 0,
      "end_ms": 856,
      "text": "Hello world."
    }
  ]
}
```

**Example** (Long Audio with Multiple Segments):
```json
{
  "text": "This is a longer transcription. It contains multiple sentences. Each sentence may be a separate segment.",
  "language": "en",
  "duration_ms": 12340,
  "segments": [
    {
      "start_ms": 0,
      "end_ms": 3200,
      "text": "This is a longer transcription."
    },
    {
      "start_ms": 3200,
      "end_ms": 6800,
      "text": " It contains multiple sentences."
    },
    {
      "start_ms": 6800,
      "end_ms": 12340,
      "text": " Each sentence may be a separate segment."
    }
  ]
}
```

### Error Response

**Structure**:
```json
{
  "error": string
}
```

**Common Error Messages**:

| HTTP Status | Error Message | Cause | Solution |
|-------------|---------------|-------|----------|
| 400 | `audio file is required` | Missing `audio` form field | Include audio file in request |
| 400 | `unsupported audio format: .m4a` | Unsupported file extension | Convert to WAV or MP3 |
| 400 | `mp3: file too large (250 MB, max 200 MB)` | MP3 exceeds 200 MB | Convert to WAV or compress |
| 400 | `mp3: invalid format` | Corrupted or invalid MP3 | Verify file integrity |
| 500 | `model not found: ggml-xyz.bin` | Invalid model name | Use valid model name |
| 500 | `transcription failed` | Internal processing error | Check server logs |

## Error Handling

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 400 | Bad Request | Invalid request (missing file, unsupported format) |
| 413 | Payload Too Large | File exceeds server limits |
| 500 | Internal Server Error | Server-side processing error |
| 503 | Service Unavailable | Server overloaded or starting up |

### Retry Strategy

**Recommended**:
- **400 errors**: Do not retry (client error)
- **500 errors**: Retry with exponential backoff (server error)
- **503 errors**: Retry after delay (service temporarily unavailable)

**Example** (Python with retries):
```python
import requests
from requests.adapters import HTTPAdapter
from requests.packages.urllib3.util.retry import Retry

session = requests.Session()
retry = Retry(
    total=3,
    backoff_factor=1,
    status_forcelist=[500, 502, 503, 504]
)
adapter = HTTPAdapter(max_retries=retry)
session.mount('http://', adapter)

response = session.post(
    'http://localhost:8080/api/transcribe',
    files={'audio': open('recording.mp3', 'rb')},
    data={'lang': 'auto'}
)
```

## Rate Limiting

**Current Behavior**: No rate limiting implemented.

**For Production**:
- Use API gateway (Kong, Traefik) with rate limiting
- Use nginx `limit_req` module
- Implement application-level throttling

**Example** (nginx rate limiting):
```nginx
http {
    limit_req_zone $binary_remote_addr zone=transcribe:10m rate=10r/m;

    server {
        location /api/transcribe {
            limit_req zone=transcribe burst=5;
            proxy_pass http://gosper-backend:8080;
        }
    }
}
```

## Client Examples

### cURL

**Basic Transcription**:
```bash
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@recording.mp3" \
  -F "lang=auto"
```

**With Custom Model**:
```bash
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@meeting.wav" \
  -F "model=ggml-medium.en.bin" \
  -F "lang=en"
```

**Save Response to File**:
```bash
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@audio.mp3" \
  -F "lang=auto" \
  -o transcript.json
```

**Parse with jq**:
```bash
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@audio.mp3" \
  -F "lang=auto" | jq -r '.text'
```

### Python

**Using requests**:
```python
import requests

# Basic transcription
with open('recording.mp3', 'rb') as audio_file:
    files = {'audio': audio_file}
    data = {'lang': 'auto'}

    response = requests.post(
        'http://localhost:8080/api/transcribe',
        files=files,
        data=data
    )

    if response.status_code == 200:
        result = response.json()
        print(f"Transcription: {result['text']}")
        print(f"Language: {result['language']}")
        print(f"Duration: {result['duration_ms']}ms")
    else:
        print(f"Error: {response.json()['error']}")
```

**With Custom Model**:
```python
import requests

files = {'audio': open('meeting.wav', 'rb')}
data = {
    'model': 'ggml-base.en.bin',
    'lang': 'en'
}

response = requests.post(
    'http://localhost:8080/api/transcribe',
    files=files,
    data=data
)

result = response.json()
for segment in result['segments']:
    start = segment['start_ms'] / 1000
    end = segment['end_ms'] / 1000
    text = segment['text']
    print(f"[{start:.2f}s - {end:.2f}s] {text}")
```

### JavaScript (Node.js)

**Using axios**:
```javascript
const axios = require('axios');
const FormData = require('form-data');
const fs = require('fs');

async function transcribe(audioPath) {
    const form = new FormData();
    form.append('audio', fs.createReadStream(audioPath));
    form.append('lang', 'auto');

    try {
        const response = await axios.post(
            'http://localhost:8080/api/transcribe',
            form,
            { headers: form.getHeaders() }
        );

        console.log('Transcription:', response.data.text);
        console.log('Language:', response.data.language);
        console.log('Duration:', response.data.duration_ms, 'ms');

        return response.data;
    } catch (error) {
        console.error('Error:', error.response?.data?.error || error.message);
        throw error;
    }
}

transcribe('recording.mp3');
```

### JavaScript (Browser)

**Using Fetch API**:
```javascript
async function transcribeAudio(audioFile) {
    const formData = new FormData();
    formData.append('audio', audioFile);
    formData.append('lang', 'auto');

    try {
        const response = await fetch('http://localhost:8080/api/transcribe', {
            method: 'POST',
            body: formData
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error);
        }

        const result = await response.json();
        console.log('Transcription:', result.text);

        return result;
    } catch (error) {
        console.error('Transcription failed:', error.message);
        throw error;
    }
}

// Usage with file input
document.getElementById('audioInput').addEventListener('change', async (e) => {
    const file = e.target.files[0];
    if (file) {
        const result = await transcribeAudio(file);
        document.getElementById('output').textContent = result.text;
    }
});
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
)

type TranscriptResponse struct {
    Text       string    `json:"text"`
    Language   string    `json:"language"`
    DurationMs int       `json:"duration_ms"`
    Segments   []Segment `json:"segments"`
}

type Segment struct {
    StartMs int    `json:"start_ms"`
    EndMs   int    `json:"end_ms"`
    Text    string `json:"text"`
}

func transcribe(audioPath string) (*TranscriptResponse, error) {
    // Open audio file
    file, err := os.Open(audioPath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Create multipart form
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // Add audio file
    part, err := writer.CreateFormFile("audio", audioPath)
    if err != nil {
        return nil, err
    }
    io.Copy(part, file)

    // Add language parameter
    writer.WriteField("lang", "auto")
    writer.Close()

    // Send request
    resp, err := http.Post(
        "http://localhost:8080/api/transcribe",
        writer.FormDataContentType(),
        body,
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Parse response
    var result TranscriptResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

func main() {
    result, err := transcribe("recording.mp3")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Transcription:", result.Text)
    fmt.Println("Language:", result.Language)
    fmt.Printf("Duration: %dms\n", result.DurationMs)
}
```

### Ruby

```ruby
require 'net/http'
require 'json'

def transcribe(audio_path)
  uri = URI('http://localhost:8080/api/transcribe')

  request = Net::HTTP::Post.new(uri)
  form_data = [
    ['audio', File.open(audio_path)],
    ['lang', 'auto']
  ]
  request.set_form(form_data, 'multipart/form-data')

  response = Net::HTTP.start(uri.hostname, uri.port) do |http|
    http.request(request)
  end

  if response.code == '200'
    result = JSON.parse(response.body)
    puts "Transcription: #{result['text']}"
    puts "Language: #{result['language']}"
    puts "Duration: #{result['duration_ms']}ms"
    result
  else
    error = JSON.parse(response.body)
    puts "Error: #{error['error']}"
    nil
  end
end

transcribe('recording.mp3')
```

## Performance Considerations

### Request Duration

Transcription time depends on:
- **Audio duration** - Longer audio = longer processing
- **Model size** - Larger models are slower but more accurate
- **Server resources** - CPU/memory availability
- **Thread count** - Set via `GOSPER_THREADS`

**Typical Performance** (with `ggml-base.en.bin` on 4-core CPU):
- 1 minute audio → ~20 seconds processing (~3x real-time)
- 10 minutes audio → ~3 minutes processing
- 1 hour audio → ~18 minutes processing

### Concurrent Requests

Gosper processes requests sequentially (no built-in queuing).

**For High Concurrency**:
- Deploy multiple Gosper instances behind load balancer
- Use Kubernetes HPA (Horizontal Pod Autoscaler)
- Implement request queue (Redis, RabbitMQ)

**Example** (Kubernetes HPA):
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gosper-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gosper-be
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Next Steps

- **[Quick Start](QUICKSTART.md)** - Get started quickly
- **[Configuration](CONFIGURATION.md)** - Environment variables and settings
- **[Deployment](DEPLOYMENT.md)** - Production deployment guide
- **[Troubleshooting](TROUBLESHOOTING.md)** - Common issues and solutions
