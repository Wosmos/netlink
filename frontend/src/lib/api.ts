import { cache, CACHE_KEYS, getOrFetch, prefetch } from './cache';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface APIResponse<T = unknown> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
}

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<APIResponse<T>> {
  try {
    const res = await fetch(`${API_URL}${endpoint}`, {
      ...options,
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });
    
    // Handle empty responses (like 200 OK with no body)
    const text = await res.text();
    if (!text) {
      return { success: res.ok };
    }
    
    try {
      return JSON.parse(text);
    } catch {
      // If response isn't JSON, wrap it
      return { success: res.ok, message: text };
    }
  } catch (error) {
    console.error('API request failed:', error);
    return { success: false, error: 'Network error - is the backend running?' };
  }
}

export const api = {
  // Auth
  login: (email: string, password: string) =>
    request<User>('/api/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }),
  register: (email: string, password: string, name?: string, phone?: string) =>
    request('/api/auth/register', { 
      method: 'POST', 
      body: JSON.stringify({ email, password, name, phone }) 
    }),
  logout: () => {
    cache.clear(); // Clear all cache on logout
    return request('/api/auth/logout', { method: 'POST' });
  },
  me: () => request<User>('/api/auth/me'),
  forgotPassword: (email: string) =>
    request('/api/auth/forgot-password', { method: 'POST', body: JSON.stringify({ email }) }),
  resetPassword: (token: string, password: string) =>
    request('/api/auth/reset-password', { method: 'POST', body: JSON.stringify({ token, password }) }),

  // Conversations (with caching)
  getConversations: async () => {
    const key = CACHE_KEYS.conversations();
    
    // Return cached if available
    const cached = cache.get<Conversation[]>(key);
    if (cached) {
      return { success: true, data: cached };
    }
    
    // Fetch from server
    const res = await request<Conversation[]>('/api/conversations');
    if (res.success && res.data) {
      cache.set(key, res.data); // Cache indefinitely, updated via WebSocket
    }
    return res;
  },
  
  // Update conversations cache (called by WebSocket)
  updateConversationsCache: (conversations: Conversation[]) => {
    cache.set(CACHE_KEYS.conversations(), conversations);
  },
  
  createDirectChat: async (data: { user_id?: number; phone?: string; email?: string }) => {
    const res = await request<Conversation>('/api/conversations/direct', { 
      method: 'POST', 
      body: JSON.stringify(data) 
    });
    if (res.success) {
      cache.invalidate(CACHE_KEYS.conversations()); // Invalidate to force refresh
    }
    return res;
  },
  
  createGroup: async (name: string, member_ids: number[]) => {
    const res = await request<Conversation>('/api/conversations/group', { 
      method: 'POST', 
      body: JSON.stringify({ name, member_ids }) 
    });
    if (res.success) {
      cache.invalidate(CACHE_KEYS.conversations());
    }
    return res;
  },
  
  deleteConversation: async (convId: number) => {
    const res = await request(`/api/conversations/delete?id=${convId}`, { method: 'DELETE' });
    if (res.success) {
      cache.invalidate(CACHE_KEYS.conversations());
      cache.invalidate(CACHE_KEYS.messages(convId));
    }
    return res;
  },

  // Messages (with caching and prefetching)
  getMessages: async (convId: number, limit = 50, offset = 0) => {
    const key = CACHE_KEYS.messages(convId);
    
    // Return cached if available (only for first page)
    if (offset === 0) {
      const cached = cache.get<Message[]>(key);
      if (cached) {
        return { success: true, data: cached };
      }
    }
    
    // Fetch from server
    const res = await request<Message[]>(
      `/api/conversations/messages?id=${convId}&limit=${limit}&offset=${offset}`
    );
    
    if (res.success && res.data && offset === 0) {
      cache.set(key, res.data); // Cache indefinitely, updated via WebSocket
    }
    
    return res;
  },
  
  // Prefetch messages for a conversation (called when hovering over conversation)
  prefetchMessages: (convId: number) => {
    prefetch(
      CACHE_KEYS.messages(convId),
      () => request<Message[]>(`/api/conversations/messages?id=${convId}&limit=50&offset=0`)
        .then(res => res.data || [])
    );
  },
  
  // Update messages cache (called by WebSocket)
  updateMessagesCache: (convId: number, messages: Message[]) => {
    cache.set(CACHE_KEYS.messages(convId), messages);
  },
  
  // Add message to cache (optimistic update)
  addMessageToCache: (convId: number, message: Message) => {
    const key = CACHE_KEYS.messages(convId);
    const cached = cache.get<Message[]>(key);
    if (cached) {
      cache.set(key, [...cached, message]);
    }
  },
  
  sendMessage: async (convId: number, content: string, type = 'text', reply_to_id?: number, voiceData?: {
    voice_file_path: string;
    voice_duration: number;
    voice_waveform: number[];
    voice_file_size: number;
  }) => {
    const res = await request<Message>(`/api/conversations/messages?id=${convId}`, {
      method: 'POST',
      body: JSON.stringify({ content, type, reply_to_id, ...voiceData }),
    });
    // Don't update cache here - WebSocket will handle it
    return res;
  },
  
  markAsRead: (convId: number) =>
    request(`/api/conversations/read?id=${convId}`, { method: 'POST' }),
  sendTyping: (convId: number) =>
    request(`/api/conversations/typing?id=${convId}`, { method: 'POST' }),
  editMessage: async (convId: number, msgId: number, content: string) => {
    const res = await request<Message>(`/api/conversations/messages/edit?id=${convId}&msg_id=${msgId}`, {
      method: 'PUT',
      body: JSON.stringify({ content }),
    });
    // Don't invalidate cache - WebSocket will handle real-time updates
    return res;
  },
  
  // Message reactions
  reactToMessage: async (msgId: number, emoji: string, isCustom = false, customUrl = '', toggle = true) => {
    return request<{ reactions: ReactionSummary[] }>(`/api/messages/react?msg_id=${msgId}`, {
      method: 'POST',
      body: JSON.stringify({ emoji, is_custom: isCustom, custom_url: customUrl, toggle }),
    });
  },
  
  removeReaction: async (msgId: number, emoji: string) => {
    return request(`/api/messages/react?msg_id=${msgId}&emoji=${encodeURIComponent(emoji)}`, {
      method: 'DELETE',
    });
  },
  
  getMessageReactions: async (msgId: number) => {
    return request(`/api/messages/reactions?msg_id=${msgId}`, {
      method: 'GET',
    });
  },
  
  // Forward message
  forwardMessage: async (messageId: number, targetConvIds: number[]) => {
    return request(`/api/messages/forward`, {
      method: 'POST',
      body: JSON.stringify({ message_id: messageId, target_conversation_ids: targetConvIds }),
    });
  },
  
  // Delete message
  deleteMessage: async (msgId: number) => {
    return request(`/api/messages/delete?msg_id=${msgId}`, {
      method: 'DELETE',
    });
  },

  // Users (with caching)
  getOnlineUsers: async () => {
    return getOrFetch(
      CACHE_KEYS.onlineUsers(),
      () => request<{ online_users: number[] }>('/api/users/online')
        .then(res => res.data || { online_users: [] }),
      10000 // Cache for 10 seconds
    ).then(data => ({ success: true, data }));
  },
  
  searchUsers: async (query: string) => {
    return getOrFetch(
      CACHE_KEYS.searchUsers(query),
      () => request<User[]>(`/api/users/search?q=${encodeURIComponent(query)}`)
        .then(res => res.data || []),
      30000 // Cache for 30 seconds
    ).then(data => ({ success: true, data }));
  },

  // Tasks (with caching)
  getTasks: async (conversationId?: number) => {
    const key = CACHE_KEYS.tasks(conversationId);
    
    // Return cached if available
    const cached = cache.get<Task[]>(key);
    if (cached) {
      return { success: true, data: cached };
    }
    
    const url = conversationId 
      ? `/api/tasks?conversation_id=${conversationId}`
      : '/api/tasks';
    const res = await request<Task[]>(url);
    
    if (res.success && res.data) {
      cache.set(key, res.data);
    }
    
    return res;
  },
  
  createTask: async (text: string, conversationId?: number) => {
    const res = await request('/api/tasks', { 
      method: 'POST', 
      body: JSON.stringify({ text, conversation_id: conversationId }) 
    });
    if (res.success) {
      cache.invalidate(CACHE_KEYS.tasks(conversationId));
    }
    return res;
  },
  
  toggleTask: async (id: number) => {
    const res = await request(`/api/tasks/toggle?id=${id}`, { method: 'POST' });
    if (res.success) {
      cache.invalidatePattern(/^tasks:/); // Invalidate all task caches
    }
    return res;
  },
  
  deleteTask: async (id: number) => {
    const res = await request(`/api/tasks/delete?id=${id}`, { method: 'POST' });
    if (res.success) {
      cache.invalidatePattern(/^tasks:/);
    }
    return res;
  },

  // Notes (with caching)
  getNotes: async (conversationId?: number) => {
    const key = CACHE_KEYS.notes(conversationId);
    
    // Return cached if available
    const cached = cache.get<Note[]>(key);
    if (cached) {
      return { success: true, data: cached };
    }
    
    const url = conversationId 
      ? `/api/notes?conversation_id=${conversationId}`
      : '/api/notes';
    const res = await request<Note[]>(url);
    
    if (res.success && res.data) {
      cache.set(key, res.data);
    }
    
    return res;
  },
  
  createNote: async (title: string, content: string, conversationId?: number) => {
    const res = await request<Note>('/api/notes', { 
      method: 'POST', 
      body: JSON.stringify({ title, content, conversation_id: conversationId }) 
    });
    if (res.success) {
      cache.invalidate(CACHE_KEYS.notes(conversationId));
    }
    return res;
  },
  
  updateNote: async (id: number, title: string, content: string) => {
    const res = await request<Note>(`/api/notes/${id}`, { 
      method: 'PUT', 
      body: JSON.stringify({ title, content }) 
    });
    if (res.success) {
      cache.invalidatePattern(/^notes:/);
    }
    return res;
  },
  
  deleteNote: async (id: number) => {
    const res = await request(`/api/notes/${id}`, { method: 'DELETE' });
    if (res.success) {
      cache.invalidatePattern(/^notes:/);
    }
    return res;
  },
  
  togglePinNote: async (id: number) => {
    const res = await request(`/api/notes/pin?id=${id}`, { method: 'POST' });
    if (res.success) {
      cache.invalidatePattern(/^notes:/);
    }
    return res;
  },
};

// Types
export interface User {
  id: number;
  email: string;
  name: string;
  phone?: string;
  avatar?: string;
  last_seen_at?: string;
  is_verified: boolean;
}

export interface Conversation {
  id: number;
  type: 'direct' | 'group';
  name?: string;
  members: ConversationMember[];
  last_message?: Message;
  unread_count: number;
  created_at: string;
}

export interface ConversationMember {
  user_id: number;
  user?: User;
  role: string;
  joined_at: string;
}

export interface ReactionSummary {
  emoji: string;
  is_custom: boolean;
  custom_url?: string;
  count: number;
  user_ids: number[];
}

export interface Message {
  id: number;
  conversation_id: number;
  sender_id: number;
  sender?: User;
  type: string;
  content: string;
  reply_to_id?: number;
  reply_to?: Message;
  read_by: number[];
  reactions?: ReactionSummary[];
  created_at: string;
  updated_at?: string;
  deleted_at?: string;
  voice_file_path?: string;
  voice_duration?: number;
  voice_waveform?: number[];
  voice_file_size?: number;
}

export interface Task {
  id: number;
  text: string;
  completed: boolean;
  created_at: string;
}

export interface Note {
  id: number;
  title: string;
  content: string;
  is_pinned: boolean;
  created_at: string;
  updated_at: string;
}
