# Voice Messaging Architecture

## Overview
High-performance voice messaging system with better quality and features than WhatsApp/Telegram.

## Features Comparison

| Feature | Our System | WhatsApp | Telegram |
|---------|-----------|----------|----------|
| Codec | Opus (32kbps) | AMR-WB (23kbps) | Opus (16-32kbps) |
| Max Duration | 10 minutes | 15 minutes | Unlimited |
| Playback Speed | 0.5x, 1x, 1.5x, 2x | 1x, 1.5x, 2x | 1x, 1.5x, 2x |
| Waveform | Real-time | Static | Static |
| Transcription | Auto (optional) | No | Premium only |
| Noise Cancellation | Yes | No | No |
| Streaming Upload | Yes | No | No |
| Resume Recording | Yes | No | No |

## Technical Architecture

### 1. Frontend (React/Next.js)

#### Recording Flow:
```
User Press → Request Mic → Start Recording → Process Audio → Upload Chunks → Send Message
```

#### Components:
- `VoiceRecorder.tsx` - Recording UI with waveform
- `VoicePlayer.tsx` - Playback with speed control
- `AudioProcessor.ts` - Web Audio API processing
- `VoiceUploader.ts` - Chunked upload handler

### 2. Backend (Go)

#### Endpoints:
- `POST /api/voice/upload/start` - Initialize upload session
- `POST /api/voice/upload/chunk` - Upload audio chunk
- `POST /api/voice/upload/complete` - Finalize upload
- `GET /api/voice/download/:id` - Stream audio file
- `POST /api/voice/transcribe/:id` - Request transcription

#### Storage Structure:
```
/voice-messages/
  /{user_id}/
    /{message_id}/
      audio.opus          # Original audio
      waveform.json       # Waveform data
      transcript.txt      # Optional transcription
      metadata.json       # Duration, size, etc.
```

### 3. Audio Processing

#### Recording Settings:
```javascript
{
  mimeType: 'audio/webm;codecs=opus',
  audioBitsPerSecond: 32000,  // 32kbps - optimal quality/size
  sampleRate: 48000,          // Opus native sample rate
  channelCount: 1             // Mono for voice
}
```

#### Noise Cancellation:
- High-pass filter (80Hz) - Remove low-frequency noise
- Noise gate (-40dB) - Remove background silence
- Dynamic range compression - Normalize volume

### 4. Waveform Generation

#### Real-time (Frontend):
- Sample audio buffer every 100ms
- Calculate RMS (Root Mean Square) for amplitude
- Store normalized values (0-1 range)
- Render as bars/curve

#### Server-side (Go):
- Use FFmpeg to extract audio samples
- Calculate peak amplitudes per time window
- Generate JSON array of normalized values

### 5. Transcription

#### Options:
1. **OpenAI Whisper API** (Cloud, $0.006/min)
   - Best accuracy
   - Multiple languages
   - Fast processing

2. **Vosk** (Self-hosted, Free)
   - Good accuracy
   - Offline processing
   - Lower cost

3. **Google Speech-to-Text** (Cloud, $0.006/15s)
   - Excellent accuracy
   - Real-time streaming

### 6. Storage & CDN

#### Storage:
- MinIO (self-hosted S3-compatible)
- AWS S3 (cloud)
- Backblaze B2 (cost-effective)

#### CDN:
- CloudFlare for global distribution
- Signed URLs for security
- 24-hour expiry on links

### 7. Database Schema

```sql
CREATE TABLE voice_messages (
    id SERIAL PRIMARY KEY,
    message_id INTEGER REFERENCES messages(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    duration_seconds DECIMAL(10,2) NOT NULL,
    file_size_bytes INTEGER NOT NULL,
    waveform_data JSONB,
    transcript TEXT,
    transcription_language VARCHAR(10),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_voice_messages_message ON voice_messages(message_id);
```

## Implementation Priority

### Phase 1 (MVP):
1. Basic recording with Opus codec
2. Simple upload to local storage
3. Playback with waveform visualization
4. Duration limit (5 minutes)

### Phase 2 (Enhanced):
1. Chunked streaming upload
2. Playback speed control
3. Noise cancellation
4. S3/MinIO storage

### Phase 3 (Advanced):
1. Auto-transcription (Whisper)
2. Voice activity detection
3. Resume interrupted recordings
4. Offline recording queue

## Performance Targets

- **Recording latency**: < 50ms
- **Upload speed**: Real-time (1x speed minimum)
- **Playback start**: < 200ms
- **Transcription**: < 5 seconds for 1-minute audio
- **File size**: ~240KB per minute (32kbps)

## Security Considerations

1. **Access Control**: Signed URLs with expiry
2. **Encryption**: TLS for transfer, optional at-rest
3. **Validation**: Check file type, duration, size
4. **Rate Limiting**: Max 10 recordings per minute
5. **Virus Scanning**: Optional for uploaded files

## Cost Estimation (1000 users, 10 voice msgs/day)

### Storage:
- 10,000 messages/day × 240KB × 30 days = 72GB/month
- S3: ~$1.66/month
- MinIO (self-hosted): Free (server costs)

### Transcription (optional):
- 10,000 messages/day × 1 min avg × $0.006 = $60/day = $1,800/month
- Vosk (self-hosted): Free (server costs ~$50/month)

### Bandwidth:
- 72GB upload + 144GB download (2x playback) = 216GB/month
- CloudFlare: Free (up to unlimited)
- AWS: ~$19/month

**Total**: $20-40/month (self-hosted) or $1,840/month (cloud transcription)

## Browser Compatibility

- Chrome/Edge: Full support
- Firefox: Full support
- Safari: Requires MediaRecorder polyfill
- Mobile browsers: Native support (iOS 14.3+, Android 5+)

## Next Steps

1. Create voice message components
2. Implement audio processing utilities
3. Add backend endpoints for upload/download
4. Integrate with existing message system
5. Add transcription service (optional)
