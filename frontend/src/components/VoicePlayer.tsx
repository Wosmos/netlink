'use client';

import { useState, useRef, useEffect } from 'react';

interface VoicePlayerProps {
  audioUrl: string;
  duration: number;
  waveform?: number[];
  senderName?: string;
  isOwn?: boolean;
}

export default function VoicePlayer({ audioUrl, duration, waveform = [], senderName, isOwn = false }: VoicePlayerProps) {
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [playbackRate, setPlaybackRate] = useState(1);
  const [isLoading, setIsLoading] = useState(false);

  const audioRef = useRef<HTMLAudioElement | null>(null);
  const progressIntervalRef = useRef<NodeJS.Timeout>();

  useEffect(() => {
    audioRef.current = new Audio(audioUrl);
    audioRef.current.playbackRate = playbackRate;

    audioRef.current.addEventListener('ended', handleEnded);
    audioRef.current.addEventListener('loadstart', () => setIsLoading(true));
    audioRef.current.addEventListener('canplay', () => setIsLoading(false));

    return () => {
      if (audioRef.current) {
        audioRef.current.pause();
        audioRef.current.removeEventListener('ended', handleEnded);
      }
      if (progressIntervalRef.current) {
        clearInterval(progressIntervalRef.current);
      }
    };
  }, [audioUrl]);

  useEffect(() => {
    if (audioRef.current) {
      audioRef.current.playbackRate = playbackRate;
    }
  }, [playbackRate]);

  const handleEnded = () => {
    setIsPlaying(false);
    setCurrentTime(0);
    if (progressIntervalRef.current) {
      clearInterval(progressIntervalRef.current);
    }
  };

  const togglePlayPause = async () => {
    if (!audioRef.current) return;

    if (isPlaying) {
      audioRef.current.pause();
      setIsPlaying(false);
      if (progressIntervalRef.current) {
        clearInterval(progressIntervalRef.current);
      }
    } else {
      try {
        await audioRef.current.play();
        setIsPlaying(true);
        
        // Update progress
        progressIntervalRef.current = setInterval(() => {
          if (audioRef.current) {
            setCurrentTime(audioRef.current.currentTime);
          }
        }, 100);
      } catch (error) {
        console.error('Error playing audio:', error);
      }
    }
  };

  const handleSeek = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!audioRef.current) return;

    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const percentage = x / rect.width;
    const newTime = percentage * duration;

    audioRef.current.currentTime = newTime;
    setCurrentTime(newTime);
  };

  const cyclePlaybackRate = () => {
    const rates = [1, 1.5, 2, 0.5];
    const currentIndex = rates.indexOf(playbackRate);
    const nextRate = rates[(currentIndex + 1) % rates.length];
    setPlaybackRate(nextRate);
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const progress = duration > 0 ? (currentTime / duration) * 100 : 0;

  return (
    <div className={`flex items-center gap-2 sm:gap-3 p-2 sm:p-3 rounded ${isOwn ? 'bg-cyan-950/40' : 'bg-[#1a1a20]'} min-w-[200px] sm:min-w-[280px]`}>
      {/* Play/Pause Button */}
      <button
        onClick={togglePlayPause}
        disabled={isLoading}
        className={`w-8 h-8 sm:w-10 sm:h-10 rounded-full flex items-center justify-center transition-all shrink-0 ${
          isOwn
            ? 'bg-cyan-500 text-black hover:bg-cyan-400'
            : 'bg-orange-500 text-black hover:bg-orange-400'
        } disabled:opacity-50 disabled:cursor-not-allowed`}
      >
        {isLoading ? (
          <div className="w-4 h-4 border-2 border-black/30 border-t-black rounded-full animate-spin" />
        ) : isPlaying ? (
          <svg className="w-4 h-4 sm:w-5 sm:h-5" fill="currentColor" viewBox="0 0 24 24">
            <path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z" />
          </svg>
        ) : (
          <svg className="w-4 h-4 sm:w-5 sm:h-5 ml-0.5" fill="currentColor" viewBox="0 0 24 24">
            <path d="M8 5v14l11-7z" />
          </svg>
        )}
      </button>

      {/* Waveform & Progress */}
      <div className="flex-1 min-w-0">
        <div
          onClick={handleSeek}
          className="h-8 sm:h-10 cursor-pointer relative flex items-center gap-0.5 px-1"
        >
          {waveform.length > 0 ? (
            waveform.map((amplitude, i) => {
              const barProgress = (i / waveform.length) * 100;
              const isPlayed = barProgress <= progress;
              return (
                <div
                  key={i}
                  className={`flex-1 rounded-full transition-all ${
                    isPlayed
                      ? isOwn
                        ? 'bg-cyan-400'
                        : 'bg-orange-400'
                      : 'bg-gray-600'
                  }`}
                  style={{
                    height: `${Math.max(10, amplitude * 100)}%`,
                    opacity: isPlayed ? 1 : 0.4,
                  }}
                />
              );
            })
          ) : (
            // Fallback progress bar if no waveform
            <div className="w-full h-1 bg-gray-700 rounded-full overflow-hidden">
              <div
                className={`h-full transition-all ${isOwn ? 'bg-cyan-400' : 'bg-orange-400'}`}
                style={{ width: `${progress}%` }}
              />
            </div>
          )}
        </div>

        {/* Time Display */}
        <div className="flex items-center justify-between mt-1 px-1">
          <span className="text-[9px] sm:text-[10px] font-mono text-gray-400">
            {formatTime(currentTime)}
          </span>
          <span className="text-[9px] sm:text-[10px] font-mono text-gray-500">
            {formatTime(duration)}
          </span>
        </div>
      </div>

      {/* Playback Speed */}
      <button
        onClick={cyclePlaybackRate}
        className={`w-8 h-8 sm:w-9 sm:h-9 rounded flex items-center justify-center text-[10px] sm:text-xs font-mono transition-all shrink-0 ${
          isOwn
            ? 'bg-cyan-900/30 text-cyan-400 hover:bg-cyan-900/50'
            : 'bg-orange-900/30 text-orange-400 hover:bg-orange-900/50'
        }`}
        title="Playback speed"
      >
        {playbackRate}x
      </button>
    </div>
  );
}
