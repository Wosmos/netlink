// API Response Types
export interface APIResponse<T = unknown> {
  success: boolean;
  message?: string;
  data?: T;
  token?: string;
  error?: string;
}

// User Types
export interface User {
  id: number;
  email: string;
  name: string;
  phone?: string;
  avatar?: string;
  last_seen_at?: string;
  is_verified: boolean;
  created_at: string;
}

// Conversation Types
export interface Conversation {
  id: number;
  type: "direct" | "group";
  name?: string;
  avatar?: string;
  members: ConversationMember[];
  last_message?: Message;
  unread_count: number;
  created_at: string;
  updated_at: string;
}

export interface ConversationMember {
  user_id: number;
  user?: User;
  role: "admin" | "member";
  joined_at: string;
  last_read_at?: string;
}

// Message Types
export interface Message {
  id: number;
  conversation_id: number;
  sender_id: number;
  sender?: User;
  type: "text" | "image" | "file" | "voice" | "system";
  content: string;
  reply_to_id?: number;
  reply_to?: Message;
  read_by?: number[];
  reactions?: ReactionSummary[];
  created_at: string;
  updated_at?: string;
  deleted_at?: string;

  // Voice message specific
  voice_file_path?: string;
  voice_duration?: number;
  voice_waveform?: number[];
  voice_file_size?: number;

  // File/Image specific
  file_path?: string;
  file_name?: string;
  file_size?: number;
  mime_type?: string;

  // Local state for optimistic updates
  local_id?: string;
  sending?: boolean;
  failed?: boolean;
}

export interface ReactionSummary {
  emoji: string;
  is_custom: boolean;
  custom_url?: string;
  count: number;
  user_ids: number[];
}

// Task Types
export interface Task {
  id: number;
  text: string;
  completed: boolean;
  conversation_id?: number;
  created_at: string;
  updated_at?: string;
}

// Note Types
export interface Note {
  id: number;
  title: string;
  content: string;
  is_pinned: boolean;
  conversation_id?: number;
  created_at: string;
  updated_at: string;
}

// WebSocket Event Types
export interface WebSocketEvent {
  type:
    | "message"
    | "typing"
    | "stop_typing"
    | "online"
    | "offline"
    | "read"
    | "conversation"
    | "reaction";
  conversation_id?: number;
  user_id?: number;
  message_id?: number;
  payload?: any;
  timestamp: string;
}

// Auth Types
export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterCredentials {
  email: string;
  password: string;
  name?: string;
  phone?: string;
}

// App State Types
export interface AppSettings {
  notifications: boolean;
  soundEnabled: boolean;
  vibrationEnabled: boolean;
  theme: "dark" | "light";
  language: "en" | "techy";
  fontSize: "small" | "medium" | "large";
}

// Navigation Types
export type RootStackParamList = {
  "(tabs)": undefined;
  "auth/login": undefined;
  "auth/register": undefined;
  "auth/forgot-password": undefined;
  "chat/[id]": { id: string };
  "modal": undefined;
};

export type TabParamList = {
  index: undefined;
  chats: undefined;
  notes: undefined;
  tasks: undefined;
  settings: undefined;
};

// Component Props Types
export interface MessageItemProps {
  message: Message;
  isOwn: boolean;
  showSender: boolean;
  onReply?: (message: Message) => void;
  onReact?: (messageId: number, emoji: string) => void;
  onEdit?: (messageId: number, content: string) => void;
  onDelete?: (messageId: number) => void;
}

export interface ConversationItemProps {
  conversation: Conversation;
  currentUserId: number;
  onPress: (conversation: Conversation) => void;
  onLongPress?: (conversation: Conversation) => void;
}

// Cache Types
export interface CacheItem<T> {
  data: T;
  timestamp: number;
  ttl?: number;
}

// Error Types
export interface AppError {
  code: string;
  message: string;
  details?: any;
}

// Typing Indicator Types
export interface TypingUser {
  user_id: number;
  user_name: string;
  timestamp: number;
}
