import { API_CONFIG, LIMITS } from '@/constants/Config';
import { APIResponse, User, Conversation, Message, Task, Note, LoginCredentials, RegisterCredentials } from '@/types';
import { getAuthToken } from './auth';

class APIClient {
  private baseURL: string;
  private timeout: number;

  constructor() {
    this.baseURL = API_CONFIG.BASE_URL;
    this.timeout = API_CONFIG.TIMEOUT;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<APIResponse<T>> {
    try {
      const token = await getAuthToken();
      
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.timeout);

      console.log(`Making API request to: ${this.baseURL}${endpoint}`);

      const response = await fetch(`${this.baseURL}${endpoint}`, {
        ...options,
        signal: controller.signal,
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          ...(token && { Authorization: `Bearer ${token}` }),
          ...options.headers,
        },
      });

      clearTimeout(timeoutId);

      console.log(`API response status: ${response.status}`);

      // Handle empty responses
      const text = await response.text();
      if (!text) {
        return { success: response.ok };
      }

      try {
        const data = JSON.parse(text);
        return data;
      } catch {
        return { success: response.ok, message: text };
      }
    } catch (error: any) {
      console.error('API request failed:', error);
      
      if (error.name === 'AbortError') {
        return { success: false, error: 'Request timeout - check your connection' };
      }
      
      if (error.message?.includes('Network request failed')) {
        return { 
          success: false, 
          error: `Cannot connect to server at ${this.baseURL}. Make sure the backend is running and accessible.` 
        };
      }
      
      return { 
        success: false, 
        error: error.message || 'Network error - check your connection' 
      };
    }
  }

  // Auth endpoints
  async login(credentials: LoginCredentials): Promise<APIResponse<User>> {
    return this.request<User>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
  }

  async register(credentials: RegisterCredentials): Promise<APIResponse> {
    return this.request('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
  }

  async logout(): Promise<APIResponse> {
    return this.request('/api/auth/logout', { method: 'POST' });
  }

  async getCurrentUser(): Promise<APIResponse<User>> {
    return this.request<User>('/api/auth/me');
  }

  async forgotPassword(email: string): Promise<APIResponse> {
    return this.request('/api/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email }),
    });
  }

  async resetPassword(token: string, password: string): Promise<APIResponse> {
    return this.request('/api/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ token, password }),
    });
  }

  // Conversation endpoints
  async getConversations(): Promise<APIResponse<Conversation[]>> {
    return this.request<Conversation[]>('/api/conversations');
  }

  async createDirectChat(data: { 
    user_id?: number; 
    phone?: string; 
    email?: string; 
  }): Promise<APIResponse<Conversation>> {
    return this.request<Conversation>('/api/conversations/direct', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async createGroup(name: string, member_ids: number[]): Promise<APIResponse<Conversation>> {
    return this.request<Conversation>('/api/conversations/group', {
      method: 'POST',
      body: JSON.stringify({ name, member_ids }),
    });
  }

  async deleteConversation(convId: number): Promise<APIResponse> {
    return this.request(`/api/conversations/delete?id=${convId}`, {
      method: 'DELETE',
    });
  }

  // Message endpoints
  async getMessages(
    convId: number,
    limit = LIMITS.MESSAGES_PER_PAGE,
    offset = 0
  ): Promise<APIResponse<Message[]>> {
    return this.request<Message[]>(
      `/api/conversations/messages?id=${convId}&limit=${limit}&offset=${offset}`
    );
  }

  async sendMessage(
    convId: number,
    content: string,
    type = 'text',
    reply_to_id?: number,
    voiceData?: {
      voice_file_path: string;
      voice_duration: number;
      voice_waveform: number[];
      voice_file_size: number;
    }
  ): Promise<APIResponse<Message>> {
    return this.request<Message>(`/api/conversations/messages?id=${convId}`, {
      method: 'POST',
      body: JSON.stringify({ content, type, reply_to_id, ...voiceData }),
    });
  }

  async editMessage(
    convId: number,
    msgId: number,
    content: string
  ): Promise<APIResponse<Message>> {
    return this.request<Message>(
      `/api/conversations/messages/edit?id=${convId}&msg_id=${msgId}`,
      {
        method: 'PUT',
        body: JSON.stringify({ content }),
      }
    );
  }

  async deleteMessage(msgId: number): Promise<APIResponse> {
    return this.request(`/api/messages/delete?msg_id=${msgId}`, {
      method: 'DELETE',
    });
  }

  async markAsRead(convId: number): Promise<APIResponse> {
    return this.request(`/api/conversations/read?id=${convId}`, {
      method: 'POST',
    });
  }

  async sendTyping(convId: number): Promise<APIResponse> {
    return this.request(`/api/conversations/typing?id=${convId}`, {
      method: 'POST',
    });
  }

  // Reaction endpoints
  async reactToMessage(
    msgId: number,
    emoji: string,
    isCustom = false,
    customUrl = '',
    toggle = true
  ): Promise<APIResponse> {
    return this.request(`/api/messages/react?msg_id=${msgId}`, {
      method: 'POST',
      body: JSON.stringify({
        emoji,
        is_custom: isCustom,
        custom_url: customUrl,
        toggle,
      }),
    });
  }

  async removeReaction(msgId: number, emoji: string): Promise<APIResponse> {
    return this.request(
      `/api/messages/react?msg_id=${msgId}&emoji=${encodeURIComponent(emoji)}`,
      { method: 'DELETE' }
    );
  }

  async getMessageReactions(msgId: number): Promise<APIResponse> {
    return this.request(`/api/messages/reactions?msg_id=${msgId}`);
  }

  // Forward message
  async forwardMessage(
    messageId: number,
    targetConvIds: number[]
  ): Promise<APIResponse> {
    return this.request('/api/messages/forward', {
      method: 'POST',
      body: JSON.stringify({
        message_id: messageId,
        target_conversation_ids: targetConvIds,
      }),
    });
  }

  // User endpoints
  async getOnlineUsers(): Promise<APIResponse<{ online_users: number[] }>> {
    return this.request<{ online_users: number[] }>('/api/users/online');
  }

  async searchUsers(query: string): Promise<APIResponse<User[]>> {
    return this.request<User[]>(
      `/api/users/search?q=${encodeURIComponent(query)}`
    );
  }

  // Task endpoints
  async getTasks(conversationId?: number): Promise<APIResponse<Task[]>> {
    const url = conversationId
      ? `/api/tasks?conversation_id=${conversationId}`
      : '/api/tasks';
    return this.request<Task[]>(url);
  }

  async createTask(text: string, conversationId?: number): Promise<APIResponse<Task>> {
    return this.request<Task>('/api/tasks', {
      method: 'POST',
      body: JSON.stringify({ text, conversation_id: conversationId }),
    });
  }

  async toggleTask(id: number): Promise<APIResponse> {
    return this.request(`/api/tasks/toggle?id=${id}`, { method: 'POST' });
  }

  async deleteTask(id: number): Promise<APIResponse> {
    return this.request(`/api/tasks/delete?id=${id}`, { method: 'POST' });
  }

  // Note endpoints
  async getNotes(conversationId?: number): Promise<APIResponse<Note[]>> {
    const url = conversationId
      ? `/api/notes?conversation_id=${conversationId}`
      : '/api/notes';
    return this.request<Note[]>(url);
  }

  async createNote(
    title: string,
    content: string,
    conversationId?: number
  ): Promise<APIResponse<Note>> {
    return this.request<Note>('/api/notes', {
      method: 'POST',
      body: JSON.stringify({ title, content, conversation_id: conversationId }),
    });
  }

  async updateNote(
    id: number,
    title: string,
    content: string
  ): Promise<APIResponse<Note>> {
    return this.request<Note>(`/api/notes/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ title, content }),
    });
  }

  async deleteNote(id: number): Promise<APIResponse> {
    return this.request(`/api/notes/${id}`, { method: 'DELETE' });
  }

  async togglePinNote(id: number): Promise<APIResponse> {
    return this.request(`/api/notes/pin?id=${id}`, { method: 'POST' });
  }

  // Voice upload endpoint
  async uploadVoice(
    audioBlob: Blob,
    duration: number,
    waveform: number[],
    onProgress?: (progress: number) => void
  ): Promise<APIResponse<{ file_path: string; file_size: number; duration: number; waveform: number[] }>> {
    try {
      const token = await getAuthToken();

      const formData = new FormData();
      // Use correct extension based on blob type
      const ext = audioBlob.type.includes('mp4') || audioBlob.type.includes('m4a') || audioBlob.type.includes('aac')
        ? '.m4a' : audioBlob.type.includes('mpeg') ? '.mp3' : '.webm';
      formData.append('audio', audioBlob as any, `voice${ext}`);
      formData.append('duration', duration.toString());
      formData.append('waveform', JSON.stringify(waveform));

      return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();

        xhr.upload.addEventListener('progress', (event) => {
          if (event.lengthComputable && onProgress) {
            const progress = (event.loaded / event.total) * 100;
            onProgress(progress);
          }
        });

        xhr.addEventListener('load', () => {
          try {
            const response = JSON.parse(xhr.responseText);
            resolve(response);
          } catch {
            resolve({ success: xhr.status === 200, message: xhr.responseText });
          }
        });

        xhr.addEventListener('error', () => {
          reject({ success: false, error: 'Voice upload failed' });
        });

        xhr.open('POST', `${this.baseURL}/api/voice/upload`);
        if (token) {
          xhr.setRequestHeader('Authorization', `Bearer ${token}`);
        }
        xhr.send(formData);
      });
    } catch (error: any) {
      return { success: false, error: error.message };
    }
  }

  // File upload endpoint
  async uploadFile(
    file: FormData,
    onProgress?: (progress: number) => void
  ): Promise<APIResponse<{ file_path: string; file_name: string; file_size: number }>> {
    try {
      const token = await getAuthToken();

      return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();

        xhr.upload.addEventListener('progress', (event) => {
          if (event.lengthComputable && onProgress) {
            const progress = (event.loaded / event.total) * 100;
            onProgress(progress);
          }
        });

        xhr.addEventListener('load', () => {
          try {
            const response = JSON.parse(xhr.responseText);
            resolve(response);
          } catch {
            resolve({ success: xhr.status === 200, message: xhr.responseText });
          }
        });

        xhr.addEventListener('error', () => {
          reject({ success: false, error: 'Upload failed' });
        });

        xhr.open('POST', `${this.baseURL}/api/upload`);
        if (token) {
          xhr.setRequestHeader('Authorization', `Bearer ${token}`);
        }
        xhr.send(file);
      });
    } catch (error: any) {
      return { success: false, error: error.message };
    }
  }
}

export const api = new APIClient();