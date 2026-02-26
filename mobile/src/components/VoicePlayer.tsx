import { useState, useRef, useEffect } from "react";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { Audio } from "expo-av";
import { Ionicons } from "@expo/vector-icons";
import { Colors } from "@/constants/Colors";

interface VoicePlayerProps {
  audioUrl: string;
  duration: number;
  waveform?: number[];
  isOwn?: boolean;
}

export default function VoicePlayer({
  audioUrl,
  duration,
  waveform = [],
  isOwn = false,
}: VoicePlayerProps) {
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [playbackRate, setPlaybackRate] = useState(1);
  const [isLoading, setIsLoading] = useState(false);

  const soundRef = useRef<Audio.Sound | null>(null);
  const progressIntervalRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    return () => {
      if (soundRef.current) {
        soundRef.current.unloadAsync();
        soundRef.current = null;
      }
      if (progressIntervalRef.current) {
        clearInterval(progressIntervalRef.current);
      }
    };
  }, [audioUrl]);

  const loadSound = async () => {
    if (soundRef.current) return soundRef.current;

    try {
      setIsLoading(true);
      const { sound } = await Audio.Sound.createAsync(
        { uri: audioUrl },
        { shouldPlay: false, rate: playbackRate },
      );

      sound.setOnPlaybackStatusUpdate((status) => {
        if (status.isLoaded) {
          if (status.didJustFinish) {
            setIsPlaying(false);
            setCurrentTime(0);
            if (progressIntervalRef.current) {
              clearInterval(progressIntervalRef.current);
            }
          }
        }
      });

      soundRef.current = sound;
      setIsLoading(false);
      return sound;
    } catch (error) {
      console.error("Error loading sound:", error);
      setIsLoading(false);
      return null;
    }
  };

  const togglePlayPause = async () => {
    try {
      const sound = await loadSound();
      if (!sound) return;

      if (isPlaying) {
        await sound.pauseAsync();
        setIsPlaying(false);
        if (progressIntervalRef.current) {
          clearInterval(progressIntervalRef.current);
        }
      } else {
        await sound.setRateAsync(playbackRate, true);
        await sound.playAsync();
        setIsPlaying(true);

        // Update progress
        progressIntervalRef.current = setInterval(async () => {
          const status = await sound.getStatusAsync();
          if (status.isLoaded && status.positionMillis !== undefined) {
            setCurrentTime(status.positionMillis / 1000);
          }
        }, 100);
      }
    } catch (error) {
      console.error("Error playing sound:", error);
    }
  };

  const handleSeek = async (percentage: number) => {
    try {
      const sound = await loadSound();
      if (!sound) return;

      const newPosition = percentage * duration * 1000; // Convert to milliseconds
      await sound.setPositionAsync(newPosition);
      setCurrentTime(newPosition / 1000);
    } catch (error) {
      console.error("Error seeking:", error);
    }
  };

  const cyclePlaybackRate = async () => {
    const rates = [1, 1.5, 2, 0.5];
    const currentIndex = rates.indexOf(playbackRate);
    const nextRate = rates[(currentIndex + 1) % rates.length];

    setPlaybackRate(nextRate);

    if (soundRef.current) {
      try {
        await soundRef.current.setRateAsync(nextRate, true);
      } catch (error) {
        console.error("Error setting playback rate:", error);
      }
    }
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  const progress = duration > 0 ? (currentTime / duration) * 100 : 0;

  return (
    <View style={[styles.container, isOwn && styles.ownContainer]}>
      {/* Play/Pause Button */}
      <TouchableOpacity
        style={[styles.playButton, isOwn && styles.ownPlayButton]}
        onPress={togglePlayPause}
        disabled={isLoading}
      >
        {isLoading ? (
          <View style={styles.loadingSpinner} />
        ) : isPlaying ? (
          <Ionicons
            name="pause"
            size={16}
            color={isOwn ? Colors.background : Colors.text}
          />
        ) : (
          <Ionicons
            name="play"
            size={16}
            color={isOwn ? Colors.background : Colors.text}
          />
        )}
      </TouchableOpacity>

      {/* Waveform & Progress */}
      <View style={styles.waveformContainer}>
        <TouchableOpacity
          style={styles.waveform}
          onPress={(e) => {
            const { locationX } = e.nativeEvent;
            const percentage = locationX / 200; // Approximate width
            handleSeek(Math.max(0, Math.min(1, percentage)));
          }}
          activeOpacity={0.7}
        >
          {waveform.length > 0 ? (
            waveform.slice(0, 40).map((amplitude, i) => {
              const barProgress = (i / waveform.length) * 100;
              const isPlayed = barProgress <= progress;
              return (
                <View
                  key={i}
                  style={[
                    styles.waveformBar,
                    {
                      height: Math.max(4, amplitude * 24),
                      backgroundColor: isPlayed
                        ? isOwn
                          ? Colors.background
                          : Colors.primary
                        : isOwn
                          ? Colors.background + "60"
                          : Colors.textMuted,
                    },
                  ]}
                />
              );
            })
          ) : (
            // Fallback progress bar
            <View style={styles.progressBar}>
              <View
                style={[
                  styles.progressFill,
                  {
                    width: `${progress}%`,
                    backgroundColor: isOwn ? Colors.background : Colors.primary,
                  },
                ]}
              />
            </View>
          )}
        </TouchableOpacity>

        {/* Time Display */}
        <View style={styles.timeContainer}>
          <Text style={[styles.timeText, isOwn && styles.ownTimeText]}>
            {formatTime(currentTime)}
          </Text>
          <Text style={[styles.timeText, isOwn && styles.ownTimeText]}>
            {formatTime(duration)}
          </Text>
        </View>
      </View>

      {/* Playback Speed */}
      <TouchableOpacity
        style={[styles.speedButton, isOwn && styles.ownSpeedButton]}
        onPress={cyclePlaybackRate}
      >
        <Text style={[styles.speedText, isOwn && styles.ownSpeedText]}>
          {playbackRate}x
        </Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: "row",
    alignItems: "center",
    paddingVertical: 8,
    minWidth: 200,
  },
  ownContainer: {
    // Styles for own messages
  },
  playButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: Colors.primary,
    alignItems: "center",
    justifyContent: "center",
    marginRight: 12,
  },
  ownPlayButton: {
    backgroundColor: Colors.background,
  },
  loadingSpinner: {
    width: 16,
    height: 16,
    borderRadius: 8,
    borderWidth: 2,
    borderColor: Colors.background,
    borderTopColor: "transparent",
  },
  waveformContainer: {
    flex: 1,
    marginRight: 12,
  },
  waveform: {
    flexDirection: "row",
    alignItems: "center",
    height: 32,
    gap: 2,
  },
  waveformBar: {
    flex: 1,
    borderRadius: 1,
    minHeight: 4,
  },
  progressBar: {
    height: 4,
    backgroundColor: Colors.textMuted,
    borderRadius: 2,
    overflow: "hidden",
    flex: 1,
  },
  progressFill: {
    height: "100%",
    borderRadius: 2,
  },
  timeContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    marginTop: 4,
  },
  timeText: {
    fontSize: 10,
    fontFamily: "SpaceMono",
    color: Colors.textMuted,
  },
  ownTimeText: {
    color: Colors.background + "CC",
  },
  speedButton: {
    width: 28,
    height: 28,
    borderRadius: 14,
    backgroundColor: Colors.surfaceVariant,
    alignItems: "center",
    justifyContent: "center",
  },
  ownSpeedButton: {
    backgroundColor: Colors.background + "40",
  },
  speedText: {
    fontSize: 10,
    fontFamily: "SpaceMono",
    fontWeight: "bold",
    color: Colors.text,
  },
  ownSpeedText: {
    color: Colors.background,
  },
});
