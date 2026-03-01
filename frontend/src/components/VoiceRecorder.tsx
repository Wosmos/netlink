'use client';

import { useState, useRef, useEffect } from 'react';

interface VoiceRecorderProps {
  onSend: (audioBlob: Blob, duration: number, waveform: number[]) => void;
  onCancel: () => void;
}

export default function VoiceRecorder({ onSend, onCancel }: VoiceRecorderProps) {
  const [isRecording, setIsRecording] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [duration, setDuration] = useState(0);
  const [waveform, setWaveform] = useState<number[]>([]);
  const [volume, setVolume] = useState(0);

  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);
  const mimeTypeRef = useRef<string>('audio/webm;codecs=opus');
  const audioContextRef = useRef<AudioContext | null>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const animationFrameRef = useRef<number | undefined>(undefined);
  const startTimeRef = useRef<number>(0);
  const timerRef = useRef<NodeJS.Timeout | undefined>(undefined);

  const MAX_DURATION = 600; // 10 minutes

  useEffect(() => {
    startRecording();
    return () => {
      stopRecording();
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    };
  }, []);

  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true,
          channelCount: 1,
        },
      });

      // Setup audio context for waveform (use default sample rate to match device)
      audioContextRef.current = new AudioContext();
      const source = audioContextRef.current.createMediaStreamSource(stream);
      analyserRef.current = audioContextRef.current.createAnalyser();
      analyserRef.current.fftSize = 2048;
      analyserRef.current.smoothingTimeConstant = 0.8;
      source.connect(analyserRef.current);

      // Setup MediaRecorder with best available codec
      let mimeType = 'audio/webm;codecs=opus';
      if (!MediaRecorder.isTypeSupported(mimeType)) {
        mimeType = 'audio/webm';
      }
      if (!MediaRecorder.isTypeSupported(mimeType)) {
        mimeType = 'audio/mp4';
      }

      mimeTypeRef.current = mimeType;

      const options: MediaRecorderOptions = {
        mimeType,
        audioBitsPerSecond: 128000,
      };

      mediaRecorderRef.current = new MediaRecorder(stream, options);
      audioChunksRef.current = [];

      mediaRecorderRef.current.ondataavailable = (event) => {
        if (event.data.size > 0) {
          audioChunksRef.current.push(event.data);
        }
      };

      mediaRecorderRef.current.start(); // Single clean blob on stop (no timeslice fragmentation)
      setIsRecording(true);
      startTimeRef.current = Date.now();

      // Start timer
      timerRef.current = setInterval(() => {
        const elapsed = Math.floor((Date.now() - startTimeRef.current) / 1000);
        setDuration(elapsed);
        if (elapsed >= MAX_DURATION) {
          handleSend();
        }
      }, 1000);

      // Start waveform visualization
      visualizeWaveform();
    } catch (error) {
      console.error('Error accessing microphone:', error);
      alert('Could not access microphone. Please check permissions.');
      onCancel();
    }
  };

  const visualizeWaveform = () => {
    if (!analyserRef.current) return;

    const bufferLength = analyserRef.current.frequencyBinCount;
    const dataArray = new Uint8Array(bufferLength);

    const draw = () => {
      if (!analyserRef.current) return;

      analyserRef.current.getByteTimeDomainData(dataArray);

      // Calculate RMS (Root Mean Square) for volume
      let sum = 0;
      for (let i = 0; i < bufferLength; i++) {
        const normalized = (dataArray[i] - 128) / 128;
        sum += normalized * normalized;
      }
      const rms = Math.sqrt(sum / bufferLength);
      setVolume(rms);

      // Add to waveform every 100ms
      if (waveform.length < duration * 10) {
        setWaveform((prev) => [...prev, rms]);
      }

      animationFrameRef.current = requestAnimationFrame(draw);
    };

    draw();
  };

  const stopRecording = () => {
    if (mediaRecorderRef.current && isRecording) {
      mediaRecorderRef.current.stop();
      mediaRecorderRef.current.stream.getTracks().forEach((track) => track.stop());
    }
    if (audioContextRef.current && audioContextRef.current.state !== 'closed') {
      audioContextRef.current.close();
    }
    if (timerRef.current) {
      clearInterval(timerRef.current);
    }
    setIsRecording(false);
  };

  const handlePauseResume = () => {
    if (!mediaRecorderRef.current) return;

    if (isPaused) {
      mediaRecorderRef.current.resume();
      startTimeRef.current = Date.now() - duration * 1000;
      timerRef.current = setInterval(() => {
        const elapsed = Math.floor((Date.now() - startTimeRef.current) / 1000);
        setDuration(elapsed);
        if (elapsed >= MAX_DURATION) {
          handleSend();
        }
      }, 1000);
    } else {
      mediaRecorderRef.current.pause();
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    }
    setIsPaused(!isPaused);
  };

  const handleSend = () => {
    if (!mediaRecorderRef.current || mediaRecorderRef.current.state === 'inactive') return;

    // Capture values before async stop
    const finalDuration = duration;
    const finalWaveform = [...waveform];

    // Wait for 'stop' event (fires after final ondataavailable) to ensure all data is collected
    mediaRecorderRef.current.addEventListener('stop', () => {
      const audioBlob = new Blob(audioChunksRef.current, { type: mimeTypeRef.current });
      onSend(audioBlob, finalDuration, finalWaveform);
    }, { once: true });

    stopRecording();
  };

  const handleCancel = () => {
    stopRecording();
    onCancel();
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-[#0c0c14] border border-cyan-500/30 rounded-lg p-6 max-w-md w-full">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-cyan-400 font-mono text-sm uppercase tracking-wider">Voice Message</h3>
          <div className="text-cyan-300 font-mono text-lg">{formatTime(duration)}</div>
        </div>

        {/* Waveform Visualization */}
        <div className="h-24 bg-[#050508] border border-cyan-900/30 rounded mb-6 p-2 flex items-center gap-0.5 overflow-hidden">
          {waveform.length === 0 ? (
            <div className="w-full text-center text-cyan-500/30 text-xs font-mono">Recording...</div>
          ) : (
            waveform.slice(-100).map((amplitude, i) => (
              <div
                key={i}
                className="flex-1 bg-cyan-500 rounded-full transition-all"
                style={{
                  height: `${Math.max(2, amplitude * 100)}%`,
                  opacity: i === waveform.length - 1 ? 1 : 0.6,
                }}
              />
            ))
          )}
        </div>

        {/* Volume Indicator */}
        <div className="mb-6">
          <div className="flex items-center gap-2 mb-2">
            <div className={`w-2 h-2 rounded-full ${isRecording && !isPaused ? 'bg-red-500 animate-pulse' : 'bg-gray-500'}`} />
            <span className="text-xs text-gray-400 font-mono">
              {isPaused ? 'PAUSED' : isRecording ? 'RECORDING' : 'STOPPED'}
            </span>
          </div>
          <div className="h-1 bg-[#050508] rounded-full overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-cyan-500 to-emerald-500 transition-all duration-100"
              style={{ width: `${volume * 100}%` }}
            />
          </div>
        </div>

        {/* Controls */}
        <div className="flex items-center justify-center gap-4">
          <button
            onClick={handleCancel}
            className="w-12 h-12 rounded-full bg-red-900/30 border border-red-500/50 text-red-400 hover:bg-red-900/50 transition-all flex items-center justify-center"
            title="Cancel"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>

          <button
            onClick={handlePauseResume}
            className="w-14 h-14 rounded-full bg-cyan-900/30 border border-cyan-500/50 text-cyan-400 hover:bg-cyan-900/50 transition-all flex items-center justify-center"
            title={isPaused ? 'Resume' : 'Pause'}
          >
            {isPaused ? (
              <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z" />
              </svg>
            ) : (
              <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
                <path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z" />
              </svg>
            )}
          </button>

          <button
            onClick={handleSend}
            disabled={duration < 1}
            className="w-12 h-12 rounded-full bg-emerald-900/30 border border-emerald-500/50 text-emerald-400 hover:bg-emerald-900/50 transition-all flex items-center justify-center disabled:opacity-30 disabled:cursor-not-allowed"
            title="Send"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
            </svg>
          </button>
        </div>

        {/* Info */}
        <div className="mt-4 text-center text-xs text-gray-500 font-mono">
          Max duration: {formatTime(MAX_DURATION)} • {duration < 1 ? 'Record at least 1 second' : 'Ready to send'}
        </div>
      </div>
    </div>
  );
}
