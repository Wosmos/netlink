import { useEffect, useState, useRef, useCallback } from "react";
import {
  View,
  Text,
  FlatList,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
  Alert,
  Keyboard,
  Modal,
} from "react-native";
import { useLocalSearchParams, router } from "expo-router";
import { Ionicons } from "@expo/vector-icons";
import { StatusBar } from "expo-status-bar";
import * as Haptics from "expo-haptics";
import { SafeAreaView } from "react-native-safe-area-context";
import { Colors } from "@/constants/Colors";
import { useAuthStore } from "@/store/authStore";
import { useChatStore } from "@/store/chatStore";
import { Message } from "@/types";
import MessageItem from "@/components/MessageItem";
import VoiceRecorder from "@/components/VoiceRecorder";
import TypingIndicator from "@/components/TypingIndicator";
import { LIMITS } from "@/constants/Config";

export default function ChatScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const conversationId = parseInt(id);

  const { user } = useAuthStore();
  const {
    conversations,
    messages,
    typingUsers,
    loadMessages,
    sendMessage,
    sendVoiceMessage,
    markAsRead,
    sendTyping,
    editMessage,
    deleteMessage,
    reactToMessage,
  } = useChatStore();

  const [messageText, setMessageText] = useState("");
  const [replyToMessage, setReplyToMessage] = useState<Message | null>(null);
  const [showVoiceRecorder, setShowVoiceRecorder] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [typingTimeout, setTypingTimeout] = useState<NodeJS.Timeout | null>(
    null,
  );

  const flatListRef = useRef<FlatList>(null);
  const textInputRef = useRef<TextInput>(null);

  const conversation = conversations.find((c) => c.id === conversationId);
  const conversationMessages = messages[conversationId] || [];
  const currentTypingUsers = typingUsers[conversationId] || [];

  useEffect(() => {
    if (conversationId) {
      loadMessages(conversationId);
      markAsRead(conversationId);
    }
  }, [conversationId]);

  useEffect(() => {
    // Mark as read when new messages arrive
    if (conversationMessages.length > 0) {
      markAsRead(conversationId);
    }
  }, [conversationMessages.length]);

  const getConversationTitle = () => {
    if (!conversation) return "Chat";

    if (conversation.type === "group") {
      return conversation.name || "Group Chat";
    }

    const otherMember = conversation.members?.find(
      (m) => m.user_id !== user?.id,
    );
    return (
      otherMember?.user?.name || otherMember?.user?.email || "Unknown User"
    );
  };

  const handleSendMessage = async () => {
    if (!messageText.trim()) return;

    const content = messageText.trim();
    setMessageText("");
    setReplyToMessage(null);

    // Haptic feedback
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);

    try {
      await sendMessage(conversationId, content, "text", replyToMessage?.id);
    } catch (error) {
      Alert.alert("Error", "Failed to send message");
    }
  };

  const handleVoiceMessage = async (
    audioBlob: Blob,
    duration: number,
    waveform: number[],
  ) => {
    setShowVoiceRecorder(false);

    try {
      await sendVoiceMessage(conversationId, audioBlob, duration, waveform);
      Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
    } catch (error) {
      Alert.alert("Error", "Failed to send voice message");
    }
  };

  const handleTyping = useCallback(() => {
    sendTyping(conversationId);

    // Clear existing timeout
    if (typingTimeout) {
      clearTimeout(typingTimeout);
    }

    // Set new timeout to stop typing indicator
    const timeout = setTimeout(() => {
      // Typing indicator will auto-expire on the receiving end
    }, 1000);

    setTypingTimeout(timeout);
  }, [conversationId, typingTimeout]);

  const handleTextChange = (text: string) => {
    setMessageText(text);

    if (text.length > 0) {
      handleTyping();
    }
  };

  const handleReply = (message: Message) => {
    setReplyToMessage(message);
    textInputRef.current?.focus();
  };

  const handleEdit = async (messageId: number, content: string) => {
    try {
      await editMessage(conversationId, messageId, content);
    } catch (error) {
      Alert.alert("Error", "Failed to edit message");
    }
  };

  const handleDelete = async (messageId: number) => {
    Alert.alert(
      "Delete Message",
      "Are you sure you want to delete this message?",
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Delete",
          style: "destructive",
          onPress: async () => {
            try {
              await deleteMessage(messageId);
            } catch (error) {
              Alert.alert("Error", "Failed to delete message");
            }
          },
        },
      ],
    );
  };

  const handleReact = async (messageId: number, emoji: string) => {
    try {
      await reactToMessage(messageId, emoji);
      Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
    } catch (error) {
      Alert.alert("Error", "Failed to react to message");
    }
  };

  const handleLoadMore = async () => {
    if (isLoadingMore || conversationMessages.length < LIMITS.MESSAGES_PER_PAGE)
      return;

    setIsLoadingMore(true);
    try {
      await loadMessages(conversationId, conversationMessages.length);
    } catch (error) {
      console.error("Failed to load more messages:", error);
    } finally {
      setIsLoadingMore(false);
    }
  };

  const renderMessage = ({ item, index }: { item: Message; index: number }) => {
    const isOwn = item.sender_id === user?.id;
    const nextMessage = conversationMessages[index + 1];
    const showSender = !nextMessage || nextMessage.sender_id !== item.sender_id;

    return (
      <MessageItem
        message={item}
        isOwn={isOwn}
        showSender={showSender}
        onReply={handleReply}
        onReact={handleReact}
        onEdit={handleEdit}
        onDelete={handleDelete}
      />
    );
  };

  const renderHeader = () => (
    <View style={styles.header}>
      <TouchableOpacity style={styles.backButton} onPress={() => router.back()}>
        <Ionicons name="arrow-back" size={24} color={Colors.primary} />
      </TouchableOpacity>

      <View style={styles.headerInfo}>
        <Text style={styles.headerTitle} numberOfLines={1}>
          {getConversationTitle()}
        </Text>
        {conversation?.type === "direct" && (
          <Text style={styles.headerSubtitle}>
            {/* Online status or last seen */}
            Online
          </Text>
        )}
      </View>

      <TouchableOpacity style={styles.headerButton}>
        <Ionicons name="ellipsis-vertical" size={20} color={Colors.primary} />
      </TouchableOpacity>
    </View>
  );

  return (
    <View style={styles.container}>
      <StatusBar style="light" />

      <SafeAreaView edges={["top"]} style={{ backgroundColor: Colors.surface }}>
        {renderHeader()}
      </SafeAreaView>

      <KeyboardAvoidingView
        style={{ flex: 1 }}
        behavior={Platform.OS === "ios" ? "padding" : "height"}
        keyboardVerticalOffset={Platform.OS === "ios" ? 0 : 20}
      >
        {/* Messages */}
        <FlatList
          ref={flatListRef}
          data={conversationMessages}
          renderItem={renderMessage}
          keyExtractor={(item) => item.id.toString()}
          style={styles.messagesList}
          contentContainerStyle={styles.messagesContent}
          onEndReached={handleLoadMore}
          onEndReachedThreshold={0.1}
          showsVerticalScrollIndicator={false}
          keyboardShouldPersistTaps="handled"
        />

        {/* Typing Indicator */}
        {currentTypingUsers.length > 0 && (
          <TypingIndicator users={currentTypingUsers} />
        )}

        {/* Reply Preview */}
        {replyToMessage && (
          <View style={styles.replyPreview}>
            <View style={styles.replyContent}>
              <Text style={styles.replyLabel}>
                Replying to {replyToMessage.sender?.name || "Unknown"}
              </Text>
              <Text style={styles.replyText} numberOfLines={1}>
                {replyToMessage.content}
              </Text>
            </View>
            <TouchableOpacity
              onPress={() => setReplyToMessage(null)}
              style={styles.replyClose}
            >
              <Ionicons name="close" size={20} color={Colors.textMuted} />
            </TouchableOpacity>
          </View>
        )}

        {/* Input Area */}
        <View style={styles.inputContainer}>
          <View style={styles.inputRow}>
            <TextInput
              ref={textInputRef}
              style={styles.textInput}
              placeholder="Type a message..."
              placeholderTextColor={Colors.textMuted}
              value={messageText}
              onChangeText={handleTextChange}
              multiline
              maxLength={LIMITS.MAX_MESSAGE_LENGTH}
              blurOnSubmit={false}
              onSubmitEditing={handleSendMessage}
            />

            {messageText.trim() ? (
              <TouchableOpacity
                style={styles.sendButton}
                onPress={handleSendMessage}
              >
                <Ionicons name="send" size={20} color={Colors.background} />
              </TouchableOpacity>
            ) : (
              <TouchableOpacity
                style={styles.voiceButton}
                onPress={() => setShowVoiceRecorder(true)}
              >
                <Ionicons name="mic" size={20} color={Colors.primary} />
              </TouchableOpacity>
            )}
          </View>
        </View>

        {/* Voice Recorder Modal */}
        {showVoiceRecorder && (
          <VoiceRecorder
            onSend={handleVoiceMessage}
            onCancel={() => setShowVoiceRecorder(false)}
          />
        )}
      </KeyboardAvoidingView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.background,
  },
  header: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: Colors.border,
    backgroundColor: Colors.surface,
  },
  backButton: {
    width: 40,
    height: 40,
    alignItems: "center",
    justifyContent: "center",
    marginRight: 8,
  },
  headerInfo: {
    flex: 1,
  },
  headerTitle: {
    fontSize: 16,
    fontFamily: "SpaceMono",
    fontWeight: "bold",
    color: Colors.text,
  },
  headerSubtitle: {
    fontSize: 12,
    fontFamily: "SpaceMono",
    color: Colors.textMuted,
    marginTop: 2,
  },
  headerButton: {
    width: 40,
    height: 40,
    alignItems: "center",
    justifyContent: "center",
  },
  messagesList: {
    flex: 1,
  },
  messagesContent: {
    paddingVertical: 16,
  },
  replyPreview: {
    flexDirection: "row",
    alignItems: "center",
    backgroundColor: Colors.surfaceVariant,
    borderTopWidth: 1,
    borderTopColor: Colors.border,
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  replyContent: {
    flex: 1,
  },
  replyLabel: {
    fontSize: 12,
    fontFamily: "SpaceMono",
    color: Colors.primary,
    marginBottom: 2,
  },
  replyText: {
    fontSize: 14,
    fontFamily: "SpaceMono",
    color: Colors.textSecondary,
  },
  replyClose: {
    width: 32,
    height: 32,
    alignItems: "center",
    justifyContent: "center",
  },
  inputContainer: {
    backgroundColor: Colors.surface,
    borderTopWidth: 1,
    borderTopColor: Colors.border,
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  inputRow: {
    flexDirection: "row",
    alignItems: "flex-end",
  },
  textInput: {
    flex: 1,
    backgroundColor: Colors.surfaceVariant,
    borderWidth: 1,
    borderColor: Colors.border,
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 14,
    fontFamily: "SpaceMono",
    color: Colors.text,
    maxHeight: 100,
    marginRight: 8,
  },
  sendButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: Colors.primary,
    alignItems: "center",
    justifyContent: "center",
  },
  voiceButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: Colors.surfaceVariant,
    borderWidth: 1,
    borderColor: Colors.borderPrimary,
    alignItems: "center",
    justifyContent: "center",
  },
});
