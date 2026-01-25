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
  register: (email: string, password: string, name?: string) =>
    request('/api/auth/register', { method: 'POST', body: JSON.stringify({ email, password, name }) }),
  logout: () => request('/api/auth/logout', { method: 'POST' }),
  me: () => request<User>('/api/auth/me'),
  forgotPassword: (email: string) =>
    request('/api/auth/forgot-password', { method: 'POST', body: JSON.stringify({ email }) }),
  resetPassword: (token: string, password: string) =>
    request('/api/auth/reset-password', { method: 'POST', body: JSON.stringify({ token, password }) }),

  // Conversations
  getConversations: () => request<Conversation[]>('/api/conversations'),
  createDirectChat: (data: { user_id?: number; phone?: string; email?: string }) =>
    request<Conversation>('/api/conversations/direct', { method: 'POST', body: JSON.stringify(data) }),
  createGroup: (name: string, member_ids: number[]) =>
    request<Conversation>('/api/conversations/group', { method: 'POST', body: JSON.stringify({ name, member_ids }) }),
  deleteConversation: (convId: number) =>
    request(`/api/conversations/delete?id=${convId}`, { method: 'DELETE' }),

  // Messages
  getMessages: (convId: number, limit = 50, offset = 0) =>
    request<Message[]>(`/api/conversations/messages?id=${convId}&limit=${limit}&offset=${offset}`),
  sendMessage: (convId: number, content: string, type = 'text', reply_to_id?: number) =>
    request<Message>(`/api/conversations/messages?id=${convId}`, {
      method: 'POST',
      body: JSON.stringify({ content, type, reply_to_id }),
    }),
  markAsRead: (convId: number) =>
    request(`/api/conversations/read?id=${convId}`, { method: 'POST' }),
  sendTyping: (convId: number) =>
    request(`/api/conversations/typing?id=${convId}`, { method: 'POST' }),
  editMessage: (convId: number, msgId: number, content: string) =>
    request<Message>(`/api/conversations/messages/edit?id=${convId}&msg_id=${msgId}`, {
      method: 'PUT',
      body: JSON.stringify({ content }),
    }),

  // Users
  getOnlineUsers: () => request<{ online_users: number[] }>('/api/users/online'),
  searchUsers: (query: string) => request<User[]>(`/api/users/search?q=${encodeURIComponent(query)}`),

  // Tasks
  getTasks: () => request<Task[]>('/api/tasks'),
  createTask: (text: string) =>
    request('/api/tasks', { method: 'POST', body: JSON.stringify({ text }) }),
  toggleTask: (id: number) =>
    request(`/api/tasks/toggle?id=${id}`, { method: 'POST' }),
  deleteTask: (id: number) =>
    request(`/api/tasks/delete?id=${id}`, { method: 'POST' }),

  // Notes
  getNotes: () => request<Note[]>('/api/notes'),
  createNote: (title: string, content: string) =>
    request<Note>('/api/notes', { method: 'POST', body: JSON.stringify({ title, content }) }),
  updateNote: (id: number, title: string, content: string) =>
    request<Note>(`/api/notes/${id}`, { method: 'PUT', body: JSON.stringify({ title, content }) }),
  deleteNote: (id: number) => request(`/api/notes/${id}`, { method: 'DELETE' }),
  togglePinNote: (id: number) => request(`/api/notes/pin?id=${id}`, { method: 'POST' }),
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
  created_at: string;
  updated_at?: string;
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
