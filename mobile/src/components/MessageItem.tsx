import { useState } from "react";
import {
  View,
  Text,
  TouchableOpacity,
  TouchableWithoutFeedback,
  StyleSheet,
  Alert,
  Modal,
  Pressable,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import * as Haptics from "expo-haptics";
import { Colors } from "@/constants/Colors";
import { API_CONFIG } from "@/constants/Config";
import { MessageItemProps } from "@/types";
import { useAuthStore } from "@/store/authStore";
import VoicePlayer from "./VoicePlayer";

export default function MessageItem({
  message,
  isOwn,
  showSender,
  onReply,
  onReact,
  onEdit,
  onDelete,
}: MessageItemProps) {
  const { token } = useAuthStore();
  const [showActions, setShowActions] = useState(false);

  const formatTime = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  };

  const handleLongPress = () => {
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
    setShowActions(true);
  };

  const handleAction = (action: string) => {
    setShowActions(false);

    switch (action) {
      case "reply":
        onReply?.(message);
        break;
      case "edit":
        if (isOwn) {
          Alert.prompt(
            "Edit Message",
            "Enter new message:",
            [
              { text: "Cancel", style: "cancel" },
              {
                text: "Save",
                onPress: (text) => {
                  if (text && text.trim()) {
                    onEdit?.(message.id, text.trim());
                  }
                },
              },
            ],
            "plain-text",
            message.content,
          );
        }
        break;
      case "delete":
        if (isOwn) {
          onDelete?.(message.id);
        }
        break;
      case "react":
        // Show emoji picker (simplified for now)
        const emojis = ["👍", "❤️", "😂", "😮", "😢", "😡"];
        Alert.alert("React to Message", "Choose a reaction:", [
          { text: "Cancel", style: "cancel" },
          ...emojis.map((emoji) => ({
            text: emoji,
            onPress: () => onReact?.(message.id, emoji),
          })),
        ]);
        break;
    }
  };

  const renderMessageContent = () => {
    switch (message.type) {
      case "voice":
        const voiceUrl = message.voice_file_path
          ? `${API_CONFIG.BASE_URL}/api/voice/download?path=${encodeURIComponent(message.voice_file_path)}${token ? `&token=${token}` : ""}`
          : "";
        return (
          <VoicePlayer
            audioUrl={voiceUrl}
            duration={message.voice_duration || 0}
            waveform={message.voice_waveform}
            isOwn={isOwn}
          />
        );
      case "image":
        return (
          <View style={styles.imageContainer}>
            <Ionicons name="image" size={24} color={Colors.textMuted} />
            <Text style={styles.imageText}>Image</Text>
          </View>
        );
      case "file":
        return (
          <View style={styles.fileContainer}>
            <Ionicons name="document" size={24} color={Colors.textMuted} />
            <Text style={styles.fileName}>{message.file_name || "File"}</Text>
          </View>
        );
      default:
        return (
          <Text style={[styles.messageText, isOwn && styles.ownMessageText]}>
            {message.content}
          </Text>
        );
    }
  };

  const renderReactions = () => {
    if (!message.reactions || message.reactions.length === 0) return null;

    return (
      <View style={styles.reactionsContainer}>
        {message.reactions.map((reaction, index) => (
          <TouchableOpacity
            key={index}
            style={styles.reactionBubble}
            onPress={() => onReact?.(message.id, reaction.emoji)}
          >
            <Text style={styles.reactionEmoji}>{reaction.emoji}</Text>
            <Text style={styles.reactionCount}>{reaction.count}</Text>
          </TouchableOpacity>
        ))}
      </View>
    );
  };

  return (
    <TouchableWithoutFeedback onPress={() => setShowActions(false)}>
      <View style={[styles.container, isOwn && styles.ownContainer]}>
        {/* Sender name for group chats */}
        {showSender && !isOwn && (
          <Text style={styles.senderName}>
            {message.sender?.name || message.sender?.email || "Unknown"}
          </Text>
        )}

        {/* Reply indicator */}
        {message.reply_to && (
          <View
            style={[styles.replyIndicator, isOwn && styles.ownReplyIndicator]}
          >
            <View style={styles.replyLine} />
            <Text style={styles.replyText} numberOfLines={1}>
              {message.reply_to.content}
            </Text>
          </View>
        )}

        {/* Message bubble */}
        <TouchableOpacity
          style={[
            styles.messageBubble,
            isOwn ? styles.ownMessageBubble : styles.otherMessageBubble,
            message.sending && styles.sendingBubble,
            message.failed && styles.failedBubble,
          ]}
          onLongPress={handleLongPress}
          activeOpacity={0.7}
        >
          {renderMessageContent()}

          {/* Message status and time */}
          <View style={styles.messageFooter}>
            <Text style={[styles.timeText, isOwn && styles.ownTimeText]}>
              {formatTime(message.created_at)}
            </Text>

            {isOwn && (
              <View style={styles.statusContainer}>
                {message.sending && (
                  <Ionicons
                    name="time-outline"
                    size={12}
                    color={Colors.textMuted}
                  />
                )}
                {message.failed && (
                  <Ionicons
                    name="alert-circle-outline"
                    size={12}
                    color={Colors.error}
                  />
                )}
                {!message.sending && !message.failed && (
                  <Ionicons
                    name={
                      (message.read_by?.length || 0) > 1
                        ? "checkmark-done"
                        : "checkmark"
                    }
                    size={12}
                    color={
                      (message.read_by?.length || 0) > 1
                        ? Colors.primary
                        : Colors.textMuted
                    }
                  />
                )}
              </View>
            )}
          </View>
        </TouchableOpacity>

        {/* Reactions */}
        {renderReactions()}

        {/* Action Menu Modal */}
        <Modal
          visible={showActions}
          transparent
          animationType="fade"
          onRequestClose={() => setShowActions(false)}
        >
          <Pressable
            style={styles.modalOverlay}
            onPress={() => setShowActions(false)}
          >
            <View
              style={[
                styles.actionsMenuModal,
                isOwn ? styles.ownActionsModal : styles.otherActionsModal,
              ]}
            >
              <TouchableOpacity
                style={styles.actionButton}
                onPress={() => handleAction("react")}
              >
                <Ionicons name="happy" size={20} color={Colors.text} />
                <Text style={styles.actionText}>React</Text>
              </TouchableOpacity>

              {isOwn && (
                <>
                  <TouchableOpacity
                    style={styles.actionButton}
                    onPress={() => handleAction("edit")}
                  >
                    <Ionicons name="create" size={20} color={Colors.text} />
                    <Text style={styles.actionText}>Edit</Text>
                  </TouchableOpacity>

                  <TouchableOpacity
                    style={styles.actionButton}
                    onPress={() => handleAction("delete")}
                  >
                    <Ionicons name="trash" size={20} color={Colors.error} />
                    <Text style={[styles.actionText, { color: Colors.error }]}>
                      Delete
                    </Text>
                  </TouchableOpacity>
                </>
              )}
            </View>
          </Pressable>
        </Modal>
      </View>
    </TouchableWithoutFeedback>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: 16,
    paddingVertical: 4,
    alignItems: "flex-start",
  },
  ownContainer: {
    alignItems: "flex-end",
  },
  senderName: {
    fontSize: 12,
    fontFamily: "SpaceMono",
    color: Colors.primary,
    marginBottom: 4,
    marginLeft: 8,
  },
  replyIndicator: {
    flexDirection: "row",
    alignItems: "center",
    marginBottom: 4,
    marginLeft: 8,
    maxWidth: "80%",
  },
  ownReplyIndicator: {
    marginLeft: 0,
    marginRight: 8,
  },
  replyLine: {
    width: 3,
    height: 20,
    backgroundColor: Colors.primary,
    marginRight: 8,
    borderRadius: 2,
  },
  replyText: {
    fontSize: 12,
    fontFamily: "SpaceMono",
    color: Colors.textMuted,
    fontStyle: "italic",
    flex: 1,
  },
  messageBubble: {
    maxWidth: "80%",
    borderRadius: 16,
    paddingHorizontal: 12,
    paddingVertical: 8,
    position: "relative",
  },
  ownMessageBubble: {
    backgroundColor: Colors.primary,
    borderBottomRightRadius: 4,
  },
  otherMessageBubble: {
    backgroundColor: Colors.surfaceVariant,
    borderBottomLeftRadius: 4,
  },
  sendingBubble: {
    opacity: 0.7,
  },
  failedBubble: {
    borderWidth: 1,
    borderColor: Colors.error,
  },
  messageText: {
    fontSize: 14,
    fontFamily: "SpaceMono",
    color: Colors.text,
    lineHeight: 20,
  },
  ownMessageText: {
    color: Colors.background,
  },
  imageContainer: {
    flexDirection: "row",
    alignItems: "center",
    paddingVertical: 8,
  },
  imageText: {
    fontSize: 14,
    fontFamily: "SpaceMono",
    color: Colors.textMuted,
    marginLeft: 8,
  },
  fileContainer: {
    flexDirection: "row",
    alignItems: "center",
    paddingVertical: 8,
  },
  fileName: {
    fontSize: 14,
    fontFamily: "SpaceMono",
    color: Colors.text,
    marginLeft: 8,
  },
  messageFooter: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "flex-end",
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
  statusContainer: {
    marginLeft: 4,
  },
  reactionsContainer: {
    flexDirection: "row",
    flexWrap: "wrap",
    marginTop: 4,
    marginHorizontal: 8,
  },
  reactionBubble: {
    flexDirection: "row",
    alignItems: "center",
    backgroundColor: Colors.surfaceVariant,
    borderRadius: 12,
    paddingHorizontal: 8,
    paddingVertical: 4,
    marginRight: 4,
    marginBottom: 4,
  },
  reactionEmoji: {
    fontSize: 12,
  },
  reactionCount: {
    fontSize: 10,
    fontFamily: "SpaceMono",
    color: Colors.textMuted,
    marginLeft: 4,
  },
  actionButton: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 12,
    minWidth: 120,
  },
  actionText: {
    fontSize: 14,
    fontFamily: "SpaceMono",
    color: Colors.text,
    marginLeft: 8,
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: "rgba(0, 0, 0, 0.5)",
    justifyContent: "center",
    alignItems: "center",
  },
  actionsMenuModal: {
    backgroundColor: Colors.surface,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: Colors.border,
    padding: 8,
    minWidth: 200,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 10,
    elevation: 8,
  },
  ownActionsModal: {
    borderRightWidth: 4,
    borderRightColor: Colors.primary,
  },
  otherActionsModal: {
    borderLeftWidth: 4,
    borderLeftColor: Colors.textMuted,
  },
});
