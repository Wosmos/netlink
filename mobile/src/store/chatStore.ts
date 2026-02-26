import { create } from "zustand";
import { Conversation, Message, TypingUser } from "@/types";
import { api } from "@/lib/api";
import { wsClient } from "@/lib/websocket";
import { cache, CACHE_KEYS } from "@/lib/cache";
import { LIMITS } from "@/constants/Config";
import { useAuthStore } from "@/store/authStore";

interface ChatState {
  conversations: Conversation[];
  messages: Record<number, Message[]>;
  onlineUsers: number[];
  typingUsers: Record<number, TypingUser[]>; // conversationId -> typing users
  isLoading: boolean;
  error: string | null;

  // Actions
  loadConversations: () => Promise<void>;
  loadMessages: (convId: number, offset?: number) => Promise<void>;
  sendMessage: (
    convId: number,
    content: string,
    type?: string,
    replyToId?: number,
  ) => Promise<void>;
  sendVoiceMessage: (
    convId: number,
    audioBlob: Blob,
    duration: number,
    waveform: number[],
  ) => Promise<void>;
  editMessage: (
    convId: number,
    msgId: number,
    content: string,
  ) => Promise<void>;
  deleteMessage: (msgId: number) => Promise<void>;
  markAsRead: (convId: number) => Promise<void>;
  sendTyping: (convId: number) => void;
  createDirectChat: (userId: number) => Promise<Conversation | null>;
  createGroup: (
    name: string,
    memberIds: number[],
  ) => Promise<Conversation | null>;
  deleteConversation: (convId: number) => Promise<void>;
  reactToMessage: (msgId: number, emoji: string) => Promise<void>;
  forwardMessage: (msgId: number, targetConvIds: number[]) => Promise<void>;
  clearError: () => void;

  // WebSocket handlers
  handleNewMessage: (message: Message) => void;
  handleTyping: (convId: number, userId: number, userName: string) => void;
  handleStopTyping: (convId: number, userId: number) => void;
  handleUserOnline: (userId: number) => void;
  handleUserOffline: (userId: number) => void;
  handleMessageRead: (
    convId: number,
    userId: number,
    messageId: number,
  ) => void;
  handleMessageEdit: (message: Message) => void;
  handleMessageDelete: (convId: number, messageId: number) => void;
  handleReaction: (
    convId: number,
    messageId: number,
    userId: number,
    emoji: string,
    added: boolean,
    reactions: any,
  ) => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  conversations: [],
  messages: {},
  onlineUsers: [],
  typingUsers: {},
  isLoading: false,
  error: null,

  loadConversations: async () => {
    set({ isLoading: true, error: null });

    try {
      // Try cache first
      const cached = await cache.get<Conversation[]>(
        CACHE_KEYS.conversations(),
      );
      if (cached) {
        set({ conversations: cached, isLoading: false });
      }

      // Fetch from API
      const response = await api.getConversations();

      if (response.success && response.data) {
        await cache.set(CACHE_KEYS.conversations(), response.data);
        set({ conversations: response.data, isLoading: false });
      } else {
        set({
          error: response.error || "Failed to load conversations",
          isLoading: false,
        });
      }
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  loadMessages: async (convId: number, offset = 0) => {
    const state = get();

    try {
      // Try cache first for initial load
      if (offset === 0) {
        const cached = await cache.get<Message[]>(CACHE_KEYS.messages(convId));
        if (cached) {
          set({
            messages: {
              ...state.messages,
              [convId]: cached,
            },
          });
        }
      }

      // Fetch from API
      const response = await api.getMessages(
        convId,
        LIMITS.MESSAGES_PER_PAGE,
        offset,
      );

      if (response.success && response.data) {
        const newMessages = response.data;

        if (offset === 0) {
          // Replace messages for initial load
          await cache.set(CACHE_KEYS.messages(convId), newMessages);
          set({
            messages: {
              ...state.messages,
              [convId]: newMessages,
            },
          });
        } else {
          // Prepend messages for pagination
          const existingMessages = state.messages[convId] || [];
          const combinedMessages = [...newMessages, ...existingMessages];

          set({
            messages: {
              ...state.messages,
              [convId]: combinedMessages,
            },
          });
        }
      }
    } catch (error: any) {
      set({ error: error.message });
    }
  },

  sendMessage: async (
    convId: number,
    content: string,
    type = "text",
    replyToId?: number,
  ) => {
    // Get current user ID from auth store
    const userId = useAuthStore.getState().user?.id;

    if (!userId) {
      set({ error: "User not authenticated" });
      return;
    }

    const state = get();
    const existingMessages = state.messages[convId] || [];

    // Create optimistic message with CORRECT sender_id for instant UI update
    const optimisticMessage: Message = {
      id: Date.now(), // Temporary ID
      local_id: `temp_${Date.now()}`,
      conversation_id: convId,
      sender_id: userId, // ✅ Use actual user ID so it renders in correct bubble
      type: type as any,
      content,
      reply_to_id: replyToId,
      read_by: [],
      created_at: new Date().toISOString(),
      sending: true,
    };

    // ⚡ INSTANT UI UPDATE - This happens synchronously, <10ms
    set({
      messages: {
        ...state.messages,
        [convId]: [...existingMessages, optimisticMessage],
      },
    });

    // 🔥 Background processing - API call runs async, doesn't block UI
    // Using Promise without await to make it truly fire-and-forget
    api
      .sendMessage(convId, content, type, replyToId)
      .then((response) => {
        if (response.success && response.data) {
          // Replace optimistic message with real message from server
          const currentMessages = get().messages[convId] || [];
          const updatedMessages = currentMessages.map((msg) =>
            msg.local_id === optimisticMessage.local_id ? response.data! : msg,
          );

          set({
            messages: {
              ...get().messages,
              [convId]: updatedMessages,
            },
          });
        } else {
          // Mark message as failed
          const currentMessages = get().messages[convId] || [];
          const updatedMessages = currentMessages.map((msg) =>
            msg.local_id === optimisticMessage.local_id
              ? { ...msg, sending: false, failed: true }
              : msg,
          );

          set({
            messages: {
              ...get().messages,
              [convId]: updatedMessages,
            },
            error: response.error || "Failed to send message",
          });
        }
      })
      .catch((error: any) => {
        // Mark message as failed
        const currentMessages = get().messages[convId] || [];
        const updatedMessages = currentMessages.map((msg) =>
          msg.local_id === optimisticMessage.local_id
            ? { ...msg, sending: false, failed: true }
            : msg,
        );

        set({
          messages: {
            ...get().messages,
            [convId]: updatedMessages,
          },
          error: error.message,
        });
      });
  },

  sendVoiceMessage: async (
    convId: number,
    audioBlob: Blob,
    duration: number,
    waveform: number[],
  ) => {
    try {
      // Upload audio file first
      const formData = new FormData();
      formData.append("file", audioBlob as any, "voice.webm");

      const uploadResponse = await api.uploadFile(formData);

      if (uploadResponse.success && uploadResponse.data) {
        // Send voice message
        const voiceData = {
          voice_file_path: uploadResponse.data.file_path,
          voice_duration: duration,
          voice_waveform: waveform,
          voice_file_size: uploadResponse.data.file_size,
        };

        await get().sendMessage(convId, "Voice message", "voice", undefined);
      } else {
        set({
          error: uploadResponse.error || "Failed to upload voice message",
        });
      }
    } catch (error: any) {
      set({ error: error.message });
    }
  },

  editMessage: async (convId: number, msgId: number, content: string) => {
    try {
      const response = await api.editMessage(convId, msgId, content);

      if (!response.success) {
        set({ error: response.error || "Failed to edit message" });
      }
      // WebSocket will handle the update
    } catch (error: any) {
      set({ error: error.message });
    }
  },

  deleteMessage: async (msgId: number) => {
    try {
      const response = await api.deleteMessage(msgId);

      if (!response.success) {
        set({ error: response.error || "Failed to delete message" });
      }
      // WebSocket will handle the update
    } catch (error: any) {
      set({ error: error.message });
    }
  },

  markAsRead: async (convId: number) => {
    try {
      await api.markAsRead(convId);

      // Update local conversation unread count
      const state = get();
      const updatedConversations = state.conversations.map((conv) =>
        conv.id === convId ? { ...conv, unread_count: 0 } : conv,
      );

      set({ conversations: updatedConversations });
      await cache.set(CACHE_KEYS.conversations(), updatedConversations);
    } catch (error: any) {
      console.error("Failed to mark as read:", error);
    }
  },

  sendTyping: (convId: number) => {
    wsClient.sendTyping(convId);
  },

  createDirectChat: async (userId: number) => {
    try {
      const response = await api.createDirectChat({ user_id: userId });

      if (response.success && response.data) {
        // Add to conversations
        const state = get();
        const updatedConversations = [response.data, ...state.conversations];

        set({ conversations: updatedConversations });
        await cache.set(CACHE_KEYS.conversations(), updatedConversations);

        return response.data;
      } else {
        set({ error: response.error || "Failed to create chat" });
        return null;
      }
    } catch (error: any) {
      set({ error: error.message });
      return null;
    }
  },

  createGroup: async (name: string, memberIds: number[]) => {
    try {
      const response = await api.createGroup(name, memberIds);

      if (response.success && response.data) {
        // Add to conversations
        const state = get();
        const updatedConversations = [response.data, ...state.conversations];

        set({ conversations: updatedConversations });
        await cache.set(CACHE_KEYS.conversations(), updatedConversations);

        return response.data;
      } else {
        set({ error: response.error || "Failed to create group" });
        return null;
      }
    } catch (error: any) {
      set({ error: error.message });
      return null;
    }
  },

  deleteConversation: async (convId: number) => {
    try {
      const response = await api.deleteConversation(convId);

      if (response.success) {
        const state = get();
        const updatedConversations = state.conversations.filter(
          (conv) => conv.id !== convId,
        );

        const newMessages = { ...state.messages };
        delete newMessages[convId];

        set({
          conversations: updatedConversations,
          messages: newMessages,
        });

        await cache.set(CACHE_KEYS.conversations(), updatedConversations);
        await cache.remove(CACHE_KEYS.messages(convId));
      } else {
        set({ error: response.error || "Failed to delete conversation" });
      }
    } catch (error: any) {
      set({ error: error.message });
    }
  },

  reactToMessage: async (msgId: number, emoji: string) => {
    try {
      const response = await api.reactToMessage(msgId, emoji);

      if (!response.success) {
        set({ error: response.error || "Failed to react to message" });
      }
      // WebSocket will handle the update
    } catch (error: any) {
      set({ error: error.message });
    }
  },

  forwardMessage: async (msgId: number, targetConvIds: number[]) => {
    try {
      const response = await api.forwardMessage(msgId, targetConvIds);

      if (!response.success) {
        set({ error: response.error || "Failed to forward message" });
      }
    } catch (error: any) {
      set({ error: error.message });
    }
  },

  clearError: () => {
    set({ error: null });
  },

  // WebSocket handlers
  handleNewMessage: (message: Message) => {
    const state = get();
    const convId = message.conversation_id;

    // Get existing messages
    const existingMessages = state.messages[convId] || [];

    // Check if message already exists (by ID) to prevent duplicates
    const messageExists = existingMessages.some((msg) => msg.id === message.id);

    let updatedMessages;
    if (messageExists) {
      // Update existing message (in case of edits or status changes)
      updatedMessages = existingMessages.map((msg) =>
        msg.id === message.id ? message : msg,
      );
    } else {
      // Check if this is replacing an optimistic message
      const optimisticIndex = existingMessages.findIndex(
        (msg) => msg.sending && msg.content === message.content,
      );

      if (optimisticIndex !== -1) {
        // Replace optimistic message with real one
        updatedMessages = [...existingMessages];
        updatedMessages[optimisticIndex] = message;
      } else {
        // Add new message
        updatedMessages = [...existingMessages, message];
      }
    }

    // Update conversations list
    const updatedConversations = state.conversations.map((conv) => {
      if (conv.id === convId) {
        return {
          ...conv,
          last_message: message,
          unread_count: conv.unread_count + 1,
          updated_at: message.created_at,
        };
      }
      return conv;
    });

    set({
      messages: {
        ...state.messages,
        [convId]: updatedMessages,
      },
      conversations: updatedConversations,
    });

    // Update cache
    cache.set(CACHE_KEYS.messages(convId), updatedMessages);
    cache.set(CACHE_KEYS.conversations(), updatedConversations);
  },

  handleTyping: (convId: number, userId: number, userName: string) => {
    const state = get();
    const existingTyping = state.typingUsers[convId] || [];

    // Remove existing typing indicator for this user
    const filteredTyping = existingTyping.filter((t) => t.user_id !== userId);

    // Add new typing indicator
    const newTyping: TypingUser = {
      user_id: userId,
      user_name: userName,
      timestamp: Date.now(),
    };

    set({
      typingUsers: {
        ...state.typingUsers,
        [convId]: [...filteredTyping, newTyping],
      },
    });

    // Auto-remove after 3 seconds
    setTimeout(() => {
      get().handleStopTyping(convId, userId);
    }, 3000);
  },

  handleStopTyping: (convId: number, userId: number) => {
    const state = get();
    const existingTyping = state.typingUsers[convId] || [];
    const filteredTyping = existingTyping.filter((t) => t.user_id !== userId);

    set({
      typingUsers: {
        ...state.typingUsers,
        [convId]: filteredTyping,
      },
    });
  },

  handleUserOnline: (userId: number) => {
    const state = get();
    if (!state.onlineUsers.includes(userId)) {
      set({ onlineUsers: [...state.onlineUsers, userId] });
    }
  },

  handleUserOffline: (userId: number) => {
    const state = get();
    set({ onlineUsers: state.onlineUsers.filter((id) => id !== userId) });
  },

  handleMessageRead: (convId: number, userId: number, messageId: number) => {
    const state = get();
    const messages = state.messages[convId];

    if (messages) {
      const updatedMessages = messages.map((msg) => {
        const readBy = msg.read_by || [];
        if (msg.id <= messageId && !readBy.includes(userId)) {
          return {
            ...msg,
            read_by: [...readBy, userId],
          };
        }
        return msg;
      });

      set({
        messages: {
          ...state.messages,
          [convId]: updatedMessages,
        },
      });

      cache.set(CACHE_KEYS.messages(convId), updatedMessages);
    }
  },

  handleMessageEdit: (message: Message) => {
    const state = get();
    const convId = message.conversation_id;
    const messages = state.messages[convId];

    if (messages) {
      const updatedMessages = messages.map((msg) =>
        msg.id === message.id ? message : msg,
      );

      set({
        messages: {
          ...state.messages,
          [convId]: updatedMessages,
        },
      });

      cache.set(CACHE_KEYS.messages(convId), updatedMessages);
    }
  },

  handleMessageDelete: (convId: number, messageId: number) => {
    const state = get();
    const messages = state.messages[convId];

    if (messages) {
      const updatedMessages = messages.filter((msg) => msg.id !== messageId);

      set({
        messages: {
          ...state.messages,
          [convId]: updatedMessages,
        },
      });

      cache.set(CACHE_KEYS.messages(convId), updatedMessages);
    }
  },

  handleReaction: (
    convId: number,
    messageId: number,
    userId: number,
    emoji: string,
    added: boolean,
    reactions: any,
  ) => {
    const state = get();
    const messages = state.messages[convId];

    if (messages) {
      const updatedMessages = messages.map((msg) => {
        if (msg.id === messageId) {
          return {
            ...msg,
            reactions: reactions,
          };
        }
        return msg;
      });

      set({
        messages: {
          ...state.messages,
          [convId]: updatedMessages,
        },
      });

      cache.set(CACHE_KEYS.messages(convId), updatedMessages);
    }
  },
}));
