# Voice Messaging Implementation Guide

## Quick Start

### 1. Backend Setup

Add voice routes to `main.go`:

```go
// Voice message handlers
voiceHandler := handlers.NewVoiceHandler(chatRepo, userRepo, authService)
http.HandleFunc("/api/voice/upload", corsMiddleware(voiceHandler.UploadVoice))
http.HandleFunc("/api/voice/download", corsMiddleware(voiceHandler.DownloadVoice))
http.HandleFunc("/api/voice/delete", corsMiddleware(voiceHandler.DeleteVoice))
```

Set environment variable:
```bash
export VOICE_UPLOAD_DIR="./uploads/voice"
```

### 2. Database Migration

Add voice message support to messages table:

```sql
-- Add voice message fields
ALTER TABLE messages ADD COLUMN IF NOT EXISTS voice_file_path TEXT;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS voice_duration DECIMAL(10,2);
ALTER TABLE messages ADD COLUMN IF NOT EXISTS voice_waveform JSONB;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS voice_file_size INTEGER;

-- Create index for voice messages
CREATE INDEX IF NOT EXISTS idx_messages_voice ON messages(voice_file_path) WHERE voice_file_path IS NOT NULL;
```

### 3. Frontend Integration

Update `chat/[id]/page.tsx`:

```typescript
import VoiceRecorder from '@/components/VoiceRecorder';
import VoicePlayer from '@/components/VoicePlayer';

// Add state
const [showVoiceRecorder, setShowVoiceRecorder] = useState(false);

// Add voice send handler
async function handleVoiceSend(audioBlob: Blob, duration: number, waveform: number[]) {
  const formData = new FormData();
  formData.append('audio', audioBlob, 'voice.webm');
  formData.append('duration', duration.toString());
  formData.append('waveform', JSON.stringify(waveform));

  // Upload voice file
  const uploadRes = await fetch(`${API_URL}/api/voice/upload`, {
    method: 'POST',
    credentials: 'include',
    body: formData,
  });

  const uploadData = await uploadRes.json();
  if (!uploadData.success) {
    alert('Failed to upload voice message');
    setShowVoiceRecorder(false);
    return;
  }

  // Send message with voice data
  const res = await api.sendMessage(convId, '', 'voice', undefined, {
    voice_file_path: uploadData.data.file_path,
    voice_duration: uploadData.data.duration,
    voice_waveform: uploadData.data.waveform,
    voice_file_size: uploadData.data.file_size,
  });

  if (res.success) {
    setShowVoiceRecorder(false);
  }
}

// Add voice button to input area
<button
  onClick={() => setShowVoiceRecorder(true)}
  className="p-2 text-cyan-400 hover:text-cyan-300"
  title="Voice message"
>
  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
  </svg>
</button>

// Render voice recorder
{showVoiceRecorder && (
  <VoiceRecorder
    onSend={handleVoiceSend}
    onCancel={() => setShowVoiceRecorder(false)}
  />
)}

// Render voice player in messages
{msg.type === 'voice' && msg.voice_file_path && (
  <VoicePlayer
    audioUrl={`${API_URL}/api/voice/download?path=${encodeURIComponent(msg.voice_file_path)}`}
    duration={msg.voice_duration || 0}
    waveform={msg.voice_waveform}
    senderName={msg.sender?.name}
    isOwn={isOwn}
  />
)}
```

### 4. Update API Types

Add to `lib/api.ts`:

```typescript
export interface Message {
  // ... existing fields
  voice_file_path?: string;
  voice_duration?: number;
  voice_waveform?: number[];
  voice_file_size?: number;
}

// Update sendMessage
sendMessage: async (
  convId: number, 
  content: string, 
  type = 'text', 
  reply_to_id?: number,
  voiceData?: {
    voice_file_path: string;
    voice_duration: number;
    voice_waveform: number[];
    voice_file_size: number;
  }
) => {
  const res = await request<Message>(`/api/conversations/messages?id=${convId}`, {
    method: 'POST',
    body: JSON.stringify({ 
      content: content || 'Voice message', 
      type, 
      reply_to_id,
      ...voiceData 
    }),
  });
  return res;
},
```

### 5. Update Backend Message Handler

Modify `handlers/chat_handler.go`:

```go
type SendMessageRequest struct {
    Content         string    `json:"content"`
    Type            string    `json:"type"`
    ReplyToID       *int      `json:"reply_to_id"`
    VoiceFilePath   string    `json:"voice_file_path"`
    VoiceDuration   float64   `json:"voice_duration"`
    VoiceWaveform   []float64 `json:"voice_waveform"`
    VoiceFileSize   int64     `json:"voice_file_size"`
}

// In SendMessage handler, save voice data to database
if req.Type == "voice" && req.VoiceFilePath != "" {
    // Store voice metadata in database
    // This can be added to the messages table or a separate voice_messages table
}
```

## Advanced Features

### 1. Voice-to-Text Transcription

#### Option A: OpenAI Whisper API

```go
func (h *VoiceHandler) TranscribeVoice(filePath string) (string, error) {
    apiKey := os.Getenv("OPENAI_API_KEY")
    
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    
    part, _ := writer.CreateFormFile("file", filepath.Base(filePath))
    io.Copy(part, file)
    writer.WriteField("model", "whisper-1")
    writer.Close()

    req, _ := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", writer.FormDataContentType())

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        Text string `json:"text"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    
    return result.Text, nil
}
```

#### Option B: Vosk (Self-hosted)

```bash
# Install Vosk
go get github.com/alphacep/vosk-api/go

# Download model
wget https://alphacephei.com/vosk/models/vosk-model-small-en-us-0.15.zip
unzip vosk-model-small-en-us-0.15.zip
```

```go
import "github.com/alphacep/vosk-api/go"

func (h *VoiceHandler) TranscribeVoiceVosk(filePath string) (string, error) {
    model, _ := vosk.NewModel("vosk-model-small-en-us-0.15")
    defer model.Free()
    
    rec, _ := vosk.NewRecognizer(model, 16000.0)
    defer rec.Free()
    
    file, _ := os.Open(filePath)
    defer file.Close()
    
    buf := make([]byte, 4096)
    var transcript string
    
    for {
        n, err := file.Read(buf)
        if err == io.EOF {
            break
        }
        
        if rec.AcceptWaveform(buf[:n]) {
            var result struct {
                Text string `json:"text"`
            }
            json.Unmarshal([]byte(rec.Result()), &result)
            transcript += result.Text + " "
        }
    }
    
    return strings.TrimSpace(transcript), nil
}
```

### 2. Noise Cancellation (Frontend)

```typescript
// Add to VoiceRecorder.tsx
const applyNoiseReduction = (audioContext: AudioContext, source: MediaStreamAudioSourceNode) => {
  // High-pass filter (remove low-frequency noise)
  const highpass = audioContext.createBiquadFilter();
  highpass.type = 'highpass';
  highpass.frequency.value = 80;
  
  // Noise gate (remove quiet background noise)
  const compressor = audioContext.createDynamicsCompressor();
  compressor.threshold.value = -40;
  compressor.knee.value = 30;
  compressor.ratio.value = 12;
  compressor.attack.value = 0;
  compressor.release.value = 0.25;
  
  // Connect nodes
  source.connect(highpass);
  highpass.connect(compressor);
  
  return compressor;
};
```

### 3. Streaming Upload (Chunked)

```typescript
// For large files, upload in chunks
async function uploadVoiceChunked(audioBlob: Blob, duration: number, waveform: number[]) {
  const chunkSize = 1024 * 1024; // 1MB chunks
  const chunks = Math.ceil(audioBlob.size / chunkSize);
  
  // Initialize upload
  const initRes = await fetch(`${API_URL}/api/voice/upload/init`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ 
      total_size: audioBlob.size,
      duration,
      waveform 
    }),
  });
  
  const { upload_id } = await initRes.json();
  
  // Upload chunks
  for (let i = 0; i < chunks; i++) {
    const start = i * chunkSize;
    const end = Math.min(start + chunkSize, audioBlob.size);
    const chunk = audioBlob.slice(start, end);
    
    const formData = new FormData();
    formData.append('chunk', chunk);
    formData.append('chunk_index', i.toString());
    formData.append('upload_id', upload_id);
    
    await fetch(`${API_URL}/api/voice/upload/chunk`, {
      method: 'POST',
      credentials: 'include',
      body: formData,
    });
  }
  
  // Finalize upload
  const finalRes = await fetch(`${API_URL}/api/voice/upload/complete`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ upload_id }),
  });
  
  return finalRes.json();
}
```

### 4. Offline Recording Queue

```typescript
// Store recordings in IndexedDB when offline
import { openDB } from 'idb';

const db = await openDB('voice-messages', 1, {
  upgrade(db) {
    db.createObjectStore('pending', { keyPath: 'id', autoIncrement: true });
  },
});

// Save offline
async function saveOfflineRecording(audioBlob: Blob, metadata: any) {
  await db.add('pending', {
    blob: audioBlob,
    metadata,
    timestamp: Date.now(),
  });
}

// Sync when online
window.addEventListener('online', async () => {
  const pending = await db.getAll('pending');
  for (const item of pending) {
    await uploadVoice(item.blob, item.metadata);
    await db.delete('pending', item.id);
  }
});
```

## Performance Optimization

### 1. Lazy Loading
```typescript
// Load voice components only when needed
const VoiceRecorder = dynamic(() => import('@/components/VoiceRecorder'), {
  ssr: false,
  loading: () => <div>Loading recorder...</div>
});
```

### 2. Audio Preloading
```typescript
// Preload audio when hovering over message
<div onMouseEnter={() => {
  const audio = new Audio(audioUrl);
  audio.preload = 'auto';
}}>
```

### 3. Waveform Caching
```typescript
// Cache waveform data in localStorage
const cacheKey = `waveform_${messageId}`;
const cached = localStorage.getItem(cacheKey);
if (cached) {
  setWaveform(JSON.parse(cached));
} else {
  // Generate and cache
  localStorage.setItem(cacheKey, JSON.stringify(waveform));
}
```

## Testing

### 1. Browser Compatibility Test
```javascript
// Check MediaRecorder support
if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
  alert('Voice messages not supported in this browser');
}

// Check Opus codec support
const isOpusSupported = MediaRecorder.isTypeSupported('audio/webm;codecs=opus');
```

### 2. Load Testing
```bash
# Test concurrent uploads
ab -n 100 -c 10 -p voice.webm -T audio/webm http://localhost:8080/api/voice/upload
```

### 3. Audio Quality Test
```bash
# Compare original vs uploaded
ffmpeg -i original.webm -i uploaded.webm -filter_complex "[0:a][1:a]amerge=inputs=2[a]" -map "[a]" comparison.wav
```

## Deployment Checklist

- [ ] Set VOICE_UPLOAD_DIR environment variable
- [ ] Configure storage (S3/MinIO) for production
- [ ] Set up CDN for audio delivery
- [ ] Enable CORS for audio streaming
- [ ] Configure max file size limits
- [ ] Set up backup for voice files
- [ ] Enable monitoring for upload failures
- [ ] Test on mobile devices (iOS/Android)
- [ ] Verify HTTPS for microphone access
- [ ] Set up transcription service (optional)

## Troubleshooting

### Issue: No microphone access
**Solution**: Ensure HTTPS is enabled (required for getUserMedia)

### Issue: Audio not playing
**Solution**: Check CORS headers, ensure audio URL is accessible

### Issue: Poor audio quality
**Solution**: Increase bitrate to 48kbps or use higher sample rate

### Issue: Large file sizes
**Solution**: Ensure Opus codec is being used, not fallback formats

### Issue: Waveform not showing
**Solution**: Check if waveform data is being sent with message
