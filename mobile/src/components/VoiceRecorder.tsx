import { useState, useEffect, useRef } from "react";
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  Modal,
  Alert,
} from "react-native";
import { Audio } from "expo-av";
import { Ionicons } from "@expo/vector-icons";
import * as Haptics from "expo-haptics";
import { Colors } from "@/constants/Colors";
import { LIMITS } from "@/constants/Config";

interface VoiceRecorderProps {
  onSend: (audioBlob: Blob, duration: number, waveform: number[]) => void;
  onCancel: () => void;
}

export default function VoiceRecorder({
  onSend,
  onCancel,
}: VoiceRecorderProps) {
  const [isRecording, setIsRecording] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [duration, setDuration] = useState(0);
  const [waveform, setWaveform] = useState<number[]>([]);
  const [volume, setVolume] = useState(0);

  const recordingRef = useRef<Audio.Recording | null>(null);
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const startTimeRef = useRef<number>(0);

  useEffect(() => {
    startRecording();
    return () => {
      stopRecording();
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    };
  }, []);

  const startRecording = async () => {
    try {
      // Request permissions
      const { status } = await Audio.requestPermissionsAsync();
      if (status !== "granted") {
        Alert.alert(
          "Permission Required",
          "Please grant microphone permission to record voice messages.",
        );
        onCancel();
        return;
      }

      // Configure audio mode
      await Audio.setAudioModeAsync({
        allowsRecordingIOS: true,
        playsInSilentModeIOS: true,
      });

      // Create recording with HIGH_QUALITY preset
      const recording = new Audio.Recording();

      // Use HIGH_QUALITY preset as base, then customize for voice
      await recording.prepareToRecordAsync({
        ...Audio.RecordingOptionsPresets.HIGH_QUALITY,
        android: {
          extension: ".m4a",
          outputFormat: Audio.AndroidOutputFormat.MPEG_4,
          audioEncoder: Audio.AndroidAudioEncoder.AAC,
          sampleRate: 48000, // Professional quality
          numberOfChannels: 1, // Mono for voice
          bitRate: 192000, // High bitrate for clarity
        },
        ios: {
          extension: ".m4a",
          outputFormat: Audio.IOSOutputFormat.MPEG4AAC,
          audioQuality: Audio.IOSAudioQuality.MAX,
          sampleRate: 48000, // Professional quality
          numberOfChannels: 1, // Mono for voice
          bitRate: 192000, // High bitrate for clarity
          linearPCMBitDepth: 16,
          linearPCMIsBigEndian: false,
          linearPCMIsFloat: false,
        },
        web: {
          mimeType: "audio/webm;codecs=opus",
          bitsPerSecond: 192000,
        },
      });

      recordingRef.current = recording;
      await recording.startAsync();
      setIsRecording(true);
      startTimeRef.current = Date.now();

      // Start timer
      timerRef.current = setInterval(() => {
        const elapsed = Math.floor((Date.now() - startTimeRef.current) / 1000);
        setDuration(elapsed);

        // Auto-stop at max duration
        if (elapsed >= LIMITS.MAX_VOICE_DURATION) {
          handleSend();
        }
      }, 1000);

      // Start waveform visualization (simplified)
      startWaveformVisualization();

      Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
    } catch (error) {
      console.error("Failed to start recording:", error);
      Alert.alert("Error", "Failed to start recording. Please try again.");
      onCancel();
    }
  };

  const startWaveformVisualization = () => {
    // Simplified waveform - generate random values for demo
    // In a real app, you'd analyze the actual audio data
    const interval = setInterval(() => {
      if (!isRecording || isPaused) return;

      const amplitude = Math.random() * 0.8 + 0.2; // Random amplitude between 0.2 and 1.0
      setVolume(amplitude);
      setWaveform((prev) => [...prev.slice(-50), amplitude]); // Keep last 50 samples
    }, 100);

    return () => clearInterval(interval);
  };

  const stopRecording = async () => {
    if (recordingRef.current && isRecording) {
      try {
        await recordingRef.current.stopAndUnloadAsync();
        setIsRecording(false);

        if (timerRef.current) {
          clearInterval(timerRef.current);
        }
      } catch (error) {
        console.error("Error stopping recording:", error);
      }
    }
  };

  const handlePauseResume = async () => {
    if (!recordingRef.current) return;

    try {
      if (isPaused) {
        // Resume recording
        await recordingRef.current.startAsync();
        setIsPaused(false);
        startTimeRef.current = Date.now() - duration * 1000;

        // Restart timer
        timerRef.current = setInterval(() => {
          const elapsed = Math.floor(
            (Date.now() - startTimeRef.current) / 1000,
          );
          setDuration(elapsed);

          if (elapsed >= LIMITS.MAX_VOICE_DURATION) {
            handleSend();
          }
        }, 1000);
      } else {
        // Pause recording
        await recordingRef.current.pauseAsync();
        setIsPaused(true);

        if (timerRef.current) {
          clearInterval(timerRef.current);
        }
      }

      Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
    } catch (error) {
      console.error("Error pausing/resuming recording:", error);
    }
  };

  const handleSend = async () => {
    if (!recordingRef.current) return;

    try {
      await stopRecording();

      const uri = recordingRef.current.getURI();
      if (uri) {
        // Convert to blob (simplified for demo)
        const response = await fetch(uri);
        const blob = await response.blob();

        onSend(blob, duration, waveform);
        Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
      }
    } catch (error) {
      console.error("Error sending voice message:", error);
      Alert.alert("Error", "Failed to send voice message. Please try again.");
    }
  };

  const handleCancel = async () => {
    await stopRecording();
    onCancel();
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  return (
    <Modal
      visible={true}
      animationType="fade"
      transparent={true}
      onRequestClose={handleCancel}
    >
      <View style={styles.overlay}>
        <View style={styles.container}>
          {/* Header */}
          <View style={styles.header}>
            <Text style={styles.title}>VOICE MESSAGE</Text>
            <Text style={styles.timer}>{formatTime(duration)}</Text>
          </View>

          {/* Waveform Visualization */}
          <View style={styles.waveformContainer}>
            {waveform.length === 0 ? (
              <Text style={styles.recordingText}>Recording...</Text>
            ) : (
              <View style={styles.waveform}>
                {waveform.slice(-30).map((amplitude, i) => (
                  <View
                    key={i}
                    style={[
                      styles.waveformBar,
                      {
                        height: Math.max(4, amplitude * 60),
                        opacity: i === waveform.length - 1 ? 1 : 0.6,
                      },
                    ]}
                  />
                ))}
              </View>
            )}
          </View>

          {/* Volume Indicator */}
          <View style={styles.volumeContainer}>
            <View
              style={[
                styles.volumeIndicator,
                {
                  backgroundColor:
                    isRecording && !isPaused ? Colors.error : Colors.textMuted,
                },
              ]}
            />
            <Text style={styles.statusText}>
              {isPaused ? "PAUSED" : isRecording ? "RECORDING" : "STOPPED"}
            </Text>
          </View>

          <View style={styles.volumeBar}>
            <View style={[styles.volumeFill, { width: `${volume * 100}%` }]} />
          </View>

          {/* Controls */}
          <View style={styles.controls}>
            <TouchableOpacity
              style={styles.cancelButton}
              onPress={handleCancel}
            >
              <Ionicons name="close" size={24} color={Colors.error} />
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.pauseButton}
              onPress={handlePauseResume}
            >
              <Ionicons
                name={isPaused ? "play" : "pause"}
                size={28}
                color={Colors.primary}
              />
            </TouchableOpacity>

            <TouchableOpacity
              style={[
                styles.sendButton,
                duration < 1 && styles.sendButtonDisabled,
              ]}
              onPress={handleSend}
              disabled={duration < 1}
            >
              <Ionicons
                name="send"
                size={24}
                color={duration < 1 ? Colors.textMuted : Colors.success}
              />
            </TouchableOpacity>
          </View>

          {/* Info */}
          <Text style={styles.infoText}>
            Max duration: {formatTime(LIMITS.MAX_VOICE_DURATION)} •
            {duration < 1 ? " Record at least 1 second" : " Ready to send"}
          </Text>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: Colors.overlay,
    justifyContent: "center",
    alignItems: "center",
  },
  container: {
    backgroundColor: Colors.surface,
    borderRadius: 16,
    padding: 24,
    margin: 20,
    minWidth: 300,
    alignItems: "center",
  },
  header: {
    alignItems: "center",
    marginBottom: 24,
  },
  title: {
    fontSize: 16,
    fontFamily: "SpaceMono",
    fontWeight: "bold",
    color: Colors.primary,
    letterSpacing: 2,
    marginBottom: 8,
  },
  timer: {
    fontSize: 24,
    fontFamily: "SpaceMono",
    fontWeight: "bold",
    color: Colors.text,
  },
  waveformContainer: {
    height: 80,
    width: "100%",
    backgroundColor: Colors.background,
    borderRadius: 8,
    marginBottom: 16,
    alignItems: "center",
    justifyContent: "center",
  },
  recordingText: {
    fontSize: 12,
    fontFamily: "SpaceMono",
    color: Colors.textMuted,
  },
  waveform: {
    flexDirection: "row",
    alignItems: "center",
    height: "100%",
    gap: 2,
    paddingHorizontal: 8,
  },
  waveformBar: {
    flex: 1,
    backgroundColor: Colors.primary,
    borderRadius: 1,
    minHeight: 4,
  },
  volumeContainer: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: 8,
  },
  volumeIndicator: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: 8,
  },
  statusText: {
    fontSize: 12,
    fontFamily: "SpaceMono",
    color: Colors.textMuted,
    letterSpacing: 1,
  },
  volumeBar: {
    width: "100%",
    height: 4,
    backgroundColor: Colors.background,
    borderRadius: 2,
    marginBottom: 24,
    overflow: "hidden",
  },
  volumeFill: {
    height: "100%",
    backgroundColor: Colors.primary,
    borderRadius: 2,
  },
  controls: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    gap: 24,
    marginBottom: 16,
  },
  cancelButton: {
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: Colors.background,
    borderWidth: 2,
    borderColor: Colors.error,
    alignItems: "center",
    justifyContent: "center",
  },
  pauseButton: {
    width: 64,
    height: 64,
    borderRadius: 32,
    backgroundColor: Colors.background,
    borderWidth: 2,
    borderColor: Colors.primary,
    alignItems: "center",
    justifyContent: "center",
  },
  sendButton: {
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: Colors.background,
    borderWidth: 2,
    borderColor: Colors.success,
    alignItems: "center",
    justifyContent: "center",
  },
  sendButtonDisabled: {
    borderColor: Colors.textMuted,
    opacity: 0.5,
  },
  infoText: {
    fontSize: 10,
    fontFamily: "SpaceMono",
    color: Colors.textMuted,
    textAlign: "center",
    lineHeight: 14,
  },
});
