# Next.js & React Native Integration Guide

Complete integration guide for using the Net Link API with Next.js (web) and React Native (mobile) applications.

---

## Table of Contents
- [Shared TypeScript Types](#shared-typescript-types)
- [Next.js Integration](#nextjs-integration)
- [React Native Integration](#react-native-integration)

---

## Shared TypeScript Types

Create a shared types file for both platforms:

```typescript
// types/chat.ts

export interface User {
  id: number;
  email: string;
  phone?: string;
  name: string;
  avatar?: string;
  is_verified: boolean;
  last_seen_at?: string;
  created_at: string;
}

export interface ConversationMember {
  conversation_id: number;
  user_id: number;
  role: 'admin' | 'member';
  joined_at: string;
  last_read_at: string;
  user?: User;
}

export interface Message {
  id: number;
  conversation_id: number;
  sender_id: number;
  type: 'text' | 'image' | 'file' | 'system';
  content: string;
  reply_to_id?: number;
  sender?: User;
  reply_to?: Message;
  created_at: string;
  updated_at: string;
}

export interface Conversation {
  id: number;
  type: 'direct' | 'group';
  name: string;
  avatar: string;
  created_by: number;
  members: ConversationMember[];
  last_message?: Message;
  unread_count: number;
  created_at: string;
  updated_at: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export type WebSocketEventType = 
  | 'message' 
  | 'typing' 
  | 'stop_typing' 
  | 'online' 
  | 'offline' 
  | 'read' 
  | 'conversation';

export interface WebSocketEvent {
  type: WebSocketEventType;
  conversation_id?: number;
  user_id?: number;
  payload?: any;
  timestamp: string;
}
```

---

## Next.js Integration

### Project Structure
```
src/
├── lib/
│   └── chat-api.ts        # API client
├── hooks/
│   ├── useChat.ts         # Chat hook
│   └── useWebSocket.ts    # WebSocket hook
├── context/
│   └── ChatContext.tsx    # Chat context provider
└── components/
    └── chat/
        ├── ChatList.tsx
        ├── ChatWindow.tsx
        └── MessageInput.tsx
```

### Environment Setup
```env
# .env.local
NEXT_PUBLIC_CHAT_API_URL=http://localhost:8080
NEXT_PUBLIC_CHAT_WS_URL=ws://localhost:8080
```

### API Client

```typescript
// lib/chat-api.ts

const API_URL = process.env.NEXT_PUBLIC_CHAT_API_URL || 'http://localhost:8080';

class ChatAPI {
  private baseUrl: string;

  constructor(baseUrl: string = API_URL) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      credentials: 'include', // Important for cookies
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });
    return response.json();
  }

  // Auth
  async login(email: string, password: string) {
    return this.request<User>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  }

  async register(email: string, password: string, name: string) {
    return this.request<void>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, name }),
    });
  }

  async logout() {
    return this.request<void>('/api/auth/logout', { method: 'POST' });
  }

  async getCurrentUser() {
    return this.request<User>('/api/auth/me');
  }

  // Conversations
  async getConversations() {
    return this.request<Conversation[]>('/api/conversations');
  }

  async createDirectChat(userId: number) {
    return this.request<Conversation>('/api/conversations/direct', {
      method: 'POST',
      body: JSON.stringify({ user_id: userId }),
    });
  }

  async createGroup(name: string, memberIds: number[]) {
    return this.request<Conversation>('/api/conversations/group', {
      method: 'POST',
      body: JSON.stringify({ name, member_ids: memberIds }),
    });
  }

  async deleteConversation(conversationId: number) {
    return this.request<void>(`/api/conversations/delete?id=${conversationId}`, {
      method: 'DELETE',
    });
  }

  // Messages
  async getMessages(conversationId: number, limit = 50, offset = 0) {
    return this.request<Message[]>(
      `/api/conversations/messages?id=${conversationId}&limit=${limit}&offset=${offset}`
    );
  }

  async sendMessage(conversationId: number, content: string, type = 'text', replyToId?: number) {
    return this.request<Message>(`/api/conversations/messages?id=${conversationId}`, {
      method: 'POST',
      body: JSON.stringify({ content, type, reply_to_id: replyToId }),
    });
  }

  async editMessage(conversationId: number, messageId: number, content: string) {
    return this.request<Message>(
      `/api/conversations/messages/edit?id=${conversationId}&msg_id=${messageId}`,
      { method: 'PUT', body: JSON.stringify({ content }) }
    );
  }

  async markAsRead(conversationId: number) {
    return this.request<void>(`/api/conversations/read?id=${conversationId}`, {
      method: 'POST',
    });
  }

  async sendTyping(conversationId: number) {
    return this.request<void>(`/api/conversations/typing?id=${conversationId}`, {
      method: 'POST',
    });
  }

  // Users
  async searchUsers(query: string) {
    return this.request<User[]>(`/api/users/search?q=${encodeURIComponent(query)}`);
  }

  async getOnlineUsers() {
    return this.request<{ online_users: number[] }>('/api/users/online');
  }
}

export const chatApi = new ChatAPI();
export default chatApi;
```


### WebSocket Hook

```typescript
// hooks/useWebSocket.ts
'use client';

import { useEffect, useRef, useCallback, useState } from 'react';
import { WebSocketEvent, WebSocketEventType } from '@/types/chat';

const WS_URL = process.env.NEXT_PUBLIC_CHAT_WS_URL || 'ws://localhost:8080';

type EventHandler = (event: WebSocketEvent) => void;

export function useWebSocket() {
  const ws = useRef<WebSocket | null>(null);
  const handlers = useRef<Map<WebSocketEventType, EventHandler[]>>(new Map());
  const [isConnected, setIsConnected] = useState(false);
  const reconnectTimeout = useRef<NodeJS.Timeout>();

  const connect = useCallback(() => {
    if (ws.current?.readyState === WebSocket.OPEN) return;

    ws.current = new WebSocket(`${WS_URL}/ws`);

    ws.current.onopen = () => {
      setIsConnected(true);
      console.log('WebSocket connected');
    };

    ws.current.onclose = () => {
      setIsConnected(false);
      console.log('WebSocket disconnected, reconnecting...');
      reconnectTimeout.current = setTimeout(connect, 3000);
    };

    ws.current.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    ws.current.onmessage = (event) => {
      try {
        const data: WebSocketEvent = JSON.parse(event.data);
        const eventHandlers = handlers.current.get(data.type);
        eventHandlers?.forEach((handler) => handler(data));
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e);
      }
    };
  }, []);

  const disconnect = useCallback(() => {
    if (reconnectTimeout.current) {
      clearTimeout(reconnectTimeout.current);
    }
    ws.current?.close();
  }, []);

  const on = useCallback((eventType: WebSocketEventType, handler: EventHandler) => {
    const existing = handlers.current.get(eventType) || [];
    handlers.current.set(eventType, [...existing, handler]);

    // Return cleanup function
    return () => {
      const current = handlers.current.get(eventType) || [];
      handlers.current.set(eventType, current.filter((h) => h !== handler));
    };
  }, []);

  const send = useCallback((data: Partial<WebSocketEvent>) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(data));
    }
  }, []);

  useEffect(() => {
    return () => disconnect();
  }, [disconnect]);

  return { connect, disconnect, on, send, isConnected };
}
```


### Chat Context Provider

```typescript
// context/ChatContext.tsx
'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { User, Conversation, Message, WebSocketEvent } from '@/types/chat';
import { chatApi } from '@/lib/chat-api';
import { useWebSocket } from '@/hooks/useWebSocket';

interface ChatContextType {
  user: User | null;
  conversations: Conversation[];
  activeConversation: Conversation | null;
  messages: Message[];
  onlineUsers: number[];
  typingUsers: Map<number, number[]>; // conversationId -> userIds
  isLoading: boolean;
  
  // Actions
  login: (email: string, password: string) => Promise<boolean>;
  logout: () => Promise<void>;
  selectConversation: (conversation: Conversation) => void;
  sendMessage: (content: string, replyToId?: number) => Promise<void>;
  createDirectChat: (userId: number) => Promise<Conversation | null>;
  createGroup: (name: string, memberIds: number[]) => Promise<Conversation | null>;
  sendTypingIndicator: () => void;
  markAsRead: () => void;
  loadMoreMessages: () => Promise<void>;
}

const ChatContext = createContext<ChatContextType | null>(null);

export function ChatProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [activeConversation, setActiveConversation] = useState<Conversation | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [onlineUsers, setOnlineUsers] = useState<number[]>([]);
  const [typingUsers, setTypingUsers] = useState<Map<number, number[]>>(new Map());
  const [isLoading, setIsLoading] = useState(true);
  const [messageOffset, setMessageOffset] = useState(0);

  const { connect, disconnect, on, isConnected } = useWebSocket();

  // Initialize - check auth and load data
  useEffect(() => {
    const init = async () => {
      try {
        const res = await chatApi.getCurrentUser();
        if (res.success && res.data) {
          setUser(res.data);
          const convRes = await chatApi.getConversations();
          if (convRes.success && convRes.data) {
            setConversations(convRes.data);
          }
          connect();
        }
      } catch (e) {
        console.error('Init error:', e);
      } finally {
        setIsLoading(false);
      }
    };
    init();
  }, [connect]);

  // WebSocket event handlers
  useEffect(() => {
    if (!isConnected) return;

    const cleanups = [
      on('message', (event: WebSocketEvent) => {
        const msg = event.payload as Message;
        if (activeConversation?.id === event.conversation_id) {
          setMessages((prev) => [...prev, msg]);
        }
        // Update conversation's last message
        setConversations((prev) =>
          prev.map((c) =>
            c.id === event.conversation_id
              ? { ...c, last_message: msg, unread_count: c.unread_count + 1 }
              : c
          )
        );
      }),

      on('typing', (event: WebSocketEvent) => {
        if (event.user_id === user?.id) return;
        setTypingUsers((prev) => {
          const newMap = new Map(prev);
          const convTyping = newMap.get(event.conversation_id!) || [];
          if (!convTyping.includes(event.user_id!)) {
            newMap.set(event.conversation_id!, [...convTyping, event.user_id!]);
          }
          return newMap;
        });
        // Clear typing after 3 seconds
        setTimeout(() => {
          setTypingUsers((prev) => {
            const newMap = new Map(prev);
            const convTyping = newMap.get(event.conversation_id!) || [];
            newMap.set(event.conversation_id!, convTyping.filter((id) => id !== event.user_id));
            return newMap;
          });
        }, 3000);
      }),

      on('online', (event: WebSocketEvent) => {
        setOnlineUsers((prev) => [...new Set([...prev, event.user_id!])]);
      }),

      on('offline', (event: WebSocketEvent) => {
        setOnlineUsers((prev) => prev.filter((id) => id !== event.user_id));
      }),

      on('conversation', (event: WebSocketEvent) => {
        const conv = event.payload as Conversation;
        setConversations((prev) => {
          const exists = prev.find((c) => c.id === conv.id);
          return exists ? prev : [conv, ...prev];
        });
      }),
    ];

    return () => cleanups.forEach((cleanup) => cleanup());
  }, [isConnected, on, activeConversation, user]);

  const login = async (email: string, password: string): Promise<boolean> => {
    const res = await chatApi.login(email, password);
    if (res.success && res.data) {
      setUser(res.data);
      const convRes = await chatApi.getConversations();
      if (convRes.success && convRes.data) {
        setConversations(convRes.data);
      }
      connect();
      return true;
    }
    return false;
  };

  const logout = async () => {
    await chatApi.logout();
    disconnect();
    setUser(null);
    setConversations([]);
    setActiveConversation(null);
    setMessages([]);
  };

  const selectConversation = async (conversation: Conversation) => {
    setActiveConversation(conversation);
    setMessageOffset(0);
    const res = await chatApi.getMessages(conversation.id);
    if (res.success && res.data) {
      setMessages(res.data);
      setMessageOffset(res.data.length);
    }
    await chatApi.markAsRead(conversation.id);
    setConversations((prev) =>
      prev.map((c) => (c.id === conversation.id ? { ...c, unread_count: 0 } : c))
    );
  };

  const sendMessage = async (content: string, replyToId?: number) => {
    if (!activeConversation) return;
    await chatApi.sendMessage(activeConversation.id, content, 'text', replyToId);
  };

  const createDirectChat = async (userId: number) => {
    const res = await chatApi.createDirectChat(userId);
    if (res.success && res.data) {
      setConversations((prev) => [res.data!, ...prev]);
      return res.data;
    }
    return null;
  };

  const createGroup = async (name: string, memberIds: number[]) => {
    const res = await chatApi.createGroup(name, memberIds);
    if (res.success && res.data) {
      setConversations((prev) => [res.data!, ...prev]);
      return res.data;
    }
    return null;
  };

  const sendTypingIndicator = useCallback(() => {
    if (activeConversation) {
      chatApi.sendTyping(activeConversation.id);
    }
  }, [activeConversation]);

  const markAsRead = useCallback(() => {
    if (activeConversation) {
      chatApi.markAsRead(activeConversation.id);
    }
  }, [activeConversation]);

  const loadMoreMessages = async () => {
    if (!activeConversation) return;
    const res = await chatApi.getMessages(activeConversation.id, 50, messageOffset);
    if (res.success && res.data && res.data.length > 0) {
      setMessages((prev) => [...res.data!, ...prev]);
      setMessageOffset((prev) => prev + res.data!.length);
    }
  };

  return (
    <ChatContext.Provider
      value={{
        user, conversations, activeConversation, messages, onlineUsers, typingUsers, isLoading,
        login, logout, selectConversation, sendMessage, createDirectChat, createGroup,
        sendTypingIndicator, markAsRead, loadMoreMessages,
      }}
    >
      {children}
    </ChatContext.Provider>
  );
}

export const useChat = () => {
  const context = useContext(ChatContext);
  if (!context) throw new Error('useChat must be used within ChatProvider');
  return context;
};
```


### Example Components

```tsx
// components/chat/ChatList.tsx
'use client';

import { useChat } from '@/context/ChatContext';
import { formatDistanceToNow } from 'date-fns';

export function ChatList() {
  const { conversations, activeConversation, selectConversation, onlineUsers, user } = useChat();

  const getOtherUser = (conv: Conversation) => {
    return conv.members.find((m) => m.user_id !== user?.id)?.user;
  };

  const getDisplayName = (conv: Conversation) => {
    if (conv.type === 'group') return conv.name;
    return getOtherUser(conv)?.name || getOtherUser(conv)?.email || 'Unknown';
  };

  return (
    <div className="w-80 border-r h-full overflow-y-auto">
      <div className="p-4 border-b">
        <h2 className="font-semibold text-lg">Messages</h2>
      </div>
      {conversations.map((conv) => {
        const otherUser = getOtherUser(conv);
        const isOnline = otherUser && onlineUsers.includes(otherUser.id);
        
        return (
          <div
            key={conv.id}
            onClick={() => selectConversation(conv)}
            className={`p-4 cursor-pointer hover:bg-gray-50 border-b ${
              activeConversation?.id === conv.id ? 'bg-blue-50' : ''
            }`}
          >
            <div className="flex items-center gap-3">
              <div className="relative">
                <div className="w-12 h-12 rounded-full bg-gray-300 flex items-center justify-center">
                  {getDisplayName(conv)[0]?.toUpperCase()}
                </div>
                {isOnline && (
                  <div className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 rounded-full border-2 border-white" />
                )}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex justify-between items-center">
                  <span className="font-medium truncate">{getDisplayName(conv)}</span>
                  {conv.last_message && (
                    <span className="text-xs text-gray-500">
                      {formatDistanceToNow(new Date(conv.last_message.created_at), { addSuffix: true })}
                    </span>
                  )}
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-500 truncate">
                    {conv.last_message?.content || 'No messages yet'}
                  </span>
                  {conv.unread_count > 0 && (
                    <span className="bg-blue-500 text-white text-xs rounded-full px-2 py-0.5">
                      {conv.unread_count}
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
```

```tsx
// components/chat/ChatWindow.tsx
'use client';

import { useChat } from '@/context/ChatContext';
import { useEffect, useRef } from 'react';
import { MessageInput } from './MessageInput';

export function ChatWindow() {
  const { activeConversation, messages, user, typingUsers, onlineUsers } = useChat();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  if (!activeConversation) {
    return (
      <div className="flex-1 flex items-center justify-center text-gray-500">
        Select a conversation to start chatting
      </div>
    );
  }

  const otherUser = activeConversation.members.find((m) => m.user_id !== user?.id)?.user;
  const isTyping = typingUsers.get(activeConversation.id)?.length > 0;

  return (
    <div className="flex-1 flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b flex items-center gap-3">
        <div className="relative">
          <div className="w-10 h-10 rounded-full bg-gray-300 flex items-center justify-center">
            {(activeConversation.name || otherUser?.name)?.[0]?.toUpperCase()}
          </div>
          {otherUser && onlineUsers.includes(otherUser.id) && (
            <div className="absolute bottom-0 right-0 w-2.5 h-2.5 bg-green-500 rounded-full border-2 border-white" />
          )}
        </div>
        <div>
          <div className="font-medium">
            {activeConversation.type === 'group' ? activeConversation.name : otherUser?.name}
          </div>
          <div className="text-xs text-gray-500">
            {otherUser && onlineUsers.includes(otherUser.id) ? 'Online' : 'Offline'}
          </div>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((msg) => {
          const isOwn = msg.sender_id === user?.id;
          return (
            <div key={msg.id} className={`flex ${isOwn ? 'justify-end' : 'justify-start'}`}>
              <div
                className={`max-w-[70%] rounded-2xl px-4 py-2 ${
                  isOwn ? 'bg-blue-500 text-white' : 'bg-gray-100'
                }`}
              >
                {!isOwn && activeConversation.type === 'group' && (
                  <div className="text-xs font-medium mb-1">{msg.sender?.name}</div>
                )}
                <div>{msg.content}</div>
                <div className={`text-xs mt-1 ${isOwn ? 'text-blue-100' : 'text-gray-500'}`}>
                  {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </div>
              </div>
            </div>
          );
        })}
        {isTyping && (
          <div className="flex justify-start">
            <div className="bg-gray-100 rounded-2xl px-4 py-2 text-gray-500">
              typing...
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      <MessageInput />
    </div>
  );
}
```

```tsx
// components/chat/MessageInput.tsx
'use client';

import { useState, useRef } from 'react';
import { useChat } from '@/context/ChatContext';

export function MessageInput() {
  const [message, setMessage] = useState('');
  const { sendMessage, sendTypingIndicator, activeConversation } = useChat();
  const typingTimeout = useRef<NodeJS.Timeout>();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!message.trim() || !activeConversation) return;
    
    await sendMessage(message.trim());
    setMessage('');
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setMessage(e.target.value);
    
    // Debounced typing indicator
    if (typingTimeout.current) clearTimeout(typingTimeout.current);
    sendTypingIndicator();
    typingTimeout.current = setTimeout(() => {}, 2000);
  };

  return (
    <form onSubmit={handleSubmit} className="p-4 border-t">
      <div className="flex gap-2">
        <input
          type="text"
          value={message}
          onChange={handleChange}
          placeholder="Type a message..."
          className="flex-1 px-4 py-2 border rounded-full focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <button
          type="submit"
          disabled={!message.trim()}
          className="px-6 py-2 bg-blue-500 text-white rounded-full hover:bg-blue-600 disabled:opacity-50"
        >
          Send
        </button>
      </div>
    </form>
  );
}
```


### App Layout Setup

```tsx
// app/layout.tsx
import { ChatProvider } from '@/context/ChatContext';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <ChatProvider>{children}</ChatProvider>
      </body>
    </html>
  );
}
```

```tsx
// app/chat/page.tsx
'use client';

import { ChatList } from '@/components/chat/ChatList';
import { ChatWindow } from '@/components/chat/ChatWindow';
import { useChat } from '@/context/ChatContext';
import { redirect } from 'next/navigation';

export default function ChatPage() {
  const { user, isLoading } = useChat();

  if (isLoading) return <div>Loading...</div>;
  if (!user) redirect('/login');

  return (
    <div className="flex h-screen">
      <ChatList />
      <ChatWindow />
    </div>
  );
}
```

---

## React Native Integration

### Project Setup

```bash
npx react-native init MyChatApp --template react-native-template-typescript
cd MyChatApp
npm install @react-native-async-storage/async-storage
```

### Project Structure
```
src/
├── api/
│   └── chatApi.ts
├── hooks/
│   ├── useChat.ts
│   └── useWebSocket.ts
├── context/
│   └── ChatContext.tsx
├── screens/
│   ├── LoginScreen.tsx
│   ├── ConversationsScreen.tsx
│   └── ChatScreen.tsx
├── components/
│   ├── ConversationItem.tsx
│   ├── MessageBubble.tsx
│   └── MessageInput.tsx
└── types/
    └── chat.ts
```

### API Client (React Native)

```typescript
// src/api/chatApi.ts

import AsyncStorage from '@react-native-async-storage/async-storage';
import { User, Conversation, Message, ApiResponse } from '../types/chat';

const API_URL = 'http://10.0.2.2:8080'; // Android emulator
// const API_URL = 'http://localhost:8080'; // iOS simulator

class ChatAPI {
  private sessionCookie: string | null = null;

  async init() {
    this.sessionCookie = await AsyncStorage.getItem('session_cookie');
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<ApiResponse<T>> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.sessionCookie) {
      headers['Cookie'] = this.sessionCookie;
    }

    const response = await fetch(`${API_URL}${endpoint}`, {
      ...options,
      headers,
    });

    // Extract and save session cookie
    const setCookie = response.headers.get('set-cookie');
    if (setCookie) {
      const sessionMatch = setCookie.match(/session_id=([^;]+)/);
      if (sessionMatch) {
        this.sessionCookie = `session_id=${sessionMatch[1]}`;
        await AsyncStorage.setItem('session_cookie', this.sessionCookie);
      }
    }

    return response.json();
  }

  async login(email: string, password: string) {
    return this.request<User>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  }

  async register(email: string, password: string, name: string) {
    return this.request<void>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, name }),
    });
  }

  async logout() {
    const result = await this.request<void>('/api/auth/logout', { method: 'POST' });
    this.sessionCookie = null;
    await AsyncStorage.removeItem('session_cookie');
    return result;
  }

  async getCurrentUser() {
    return this.request<User>('/api/auth/me');
  }

  async getConversations() {
    return this.request<Conversation[]>('/api/conversations');
  }

  async createDirectChat(userId: number) {
    return this.request<Conversation>('/api/conversations/direct', {
      method: 'POST',
      body: JSON.stringify({ user_id: userId }),
    });
  }

  async createGroup(name: string, memberIds: number[]) {
    return this.request<Conversation>('/api/conversations/group', {
      method: 'POST',
      body: JSON.stringify({ name, member_ids: memberIds }),
    });
  }

  async getMessages(conversationId: number, limit = 50, offset = 0) {
    return this.request<Message[]>(
      `/api/conversations/messages?id=${conversationId}&limit=${limit}&offset=${offset}`
    );
  }

  async sendMessage(conversationId: number, content: string, type = 'text', replyToId?: number) {
    return this.request<Message>(`/api/conversations/messages?id=${conversationId}`, {
      method: 'POST',
      body: JSON.stringify({ content, type, reply_to_id: replyToId }),
    });
  }

  async markAsRead(conversationId: number) {
    return this.request<void>(`/api/conversations/read?id=${conversationId}`, {
      method: 'POST',
    });
  }

  async sendTyping(conversationId: number) {
    return this.request<void>(`/api/conversations/typing?id=${conversationId}`, {
      method: 'POST',
    });
  }

  async searchUsers(query: string) {
    return this.request<User[]>(`/api/users/search?q=${encodeURIComponent(query)}`);
  }

  getSessionCookie() {
    return this.sessionCookie;
  }
}

export const chatApi = new ChatAPI();
export default chatApi;
```


### WebSocket Hook (React Native)

```typescript
// src/hooks/useWebSocket.ts

import { useEffect, useRef, useCallback, useState } from 'react';
import { AppState, AppStateStatus } from 'react-native';
import { WebSocketEvent, WebSocketEventType } from '../types/chat';
import { chatApi } from '../api/chatApi';

const WS_URL = 'ws://10.0.2.2:8080'; // Android emulator
// const WS_URL = 'ws://localhost:8080'; // iOS simulator

type EventHandler = (event: WebSocketEvent) => void;

export function useWebSocket() {
  const ws = useRef<WebSocket | null>(null);
  const handlers = useRef<Map<WebSocketEventType, EventHandler[]>>(new Map());
  const [isConnected, setIsConnected] = useState(false);
  const reconnectTimeout = useRef<NodeJS.Timeout>();
  const appState = useRef(AppState.currentState);

  const connect = useCallback(() => {
    if (ws.current?.readyState === WebSocket.OPEN) return;

    const cookie = chatApi.getSessionCookie();
    if (!cookie) {
      console.log('No session cookie, skipping WebSocket connection');
      return;
    }

    ws.current = new WebSocket(`${WS_URL}/ws`, undefined, {
      headers: { Cookie: cookie },
    });

    ws.current.onopen = () => {
      setIsConnected(true);
      console.log('WebSocket connected');
    };

    ws.current.onclose = () => {
      setIsConnected(false);
      console.log('WebSocket disconnected');
      // Only reconnect if app is active
      if (appState.current === 'active') {
        reconnectTimeout.current = setTimeout(connect, 3000);
      }
    };

    ws.current.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    ws.current.onmessage = (event) => {
      try {
        const data: WebSocketEvent = JSON.parse(event.data);
        const eventHandlers = handlers.current.get(data.type);
        eventHandlers?.forEach((handler) => handler(data));
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e);
      }
    };
  }, []);

  const disconnect = useCallback(() => {
    if (reconnectTimeout.current) {
      clearTimeout(reconnectTimeout.current);
    }
    ws.current?.close();
    ws.current = null;
  }, []);

  const on = useCallback((eventType: WebSocketEventType, handler: EventHandler) => {
    const existing = handlers.current.get(eventType) || [];
    handlers.current.set(eventType, [...existing, handler]);

    return () => {
      const current = handlers.current.get(eventType) || [];
      handlers.current.set(eventType, current.filter((h) => h !== handler));
    };
  }, []);

  const send = useCallback((data: Partial<WebSocketEvent>) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(data));
    }
  }, []);

  // Handle app state changes (background/foreground)
  useEffect(() => {
    const subscription = AppState.addEventListener('change', (nextAppState: AppStateStatus) => {
      if (appState.current.match(/inactive|background/) && nextAppState === 'active') {
        // App came to foreground - reconnect
        connect();
      } else if (nextAppState.match(/inactive|background/)) {
        // App went to background - disconnect
        disconnect();
      }
      appState.current = nextAppState;
    });

    return () => subscription.remove();
  }, [connect, disconnect]);

  useEffect(() => {
    return () => disconnect();
  }, [disconnect]);

  return { connect, disconnect, on, send, isConnected };
}
```


### Chat Context (React Native)

```typescript
// src/context/ChatContext.tsx

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { User, Conversation, Message, WebSocketEvent } from '../types/chat';
import { chatApi } from '../api/chatApi';
import { useWebSocket } from '../hooks/useWebSocket';

interface ChatContextType {
  user: User | null;
  conversations: Conversation[];
  activeConversation: Conversation | null;
  messages: Message[];
  onlineUsers: number[];
  typingUsers: Map<number, number[]>;
  isLoading: boolean;
  isAuthenticated: boolean;

  login: (email: string, password: string) => Promise<boolean>;
  logout: () => Promise<void>;
  selectConversation: (conversation: Conversation) => Promise<void>;
  clearActiveConversation: () => void;
  sendMessage: (content: string, replyToId?: number) => Promise<void>;
  createDirectChat: (userId: number) => Promise<Conversation | null>;
  sendTypingIndicator: () => void;
  refreshConversations: () => Promise<void>;
}

const ChatContext = createContext<ChatContextType | null>(null);

export function ChatProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [activeConversation, setActiveConversation] = useState<Conversation | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [onlineUsers, setOnlineUsers] = useState<number[]>([]);
  const [typingUsers, setTypingUsers] = useState<Map<number, number[]>>(new Map());
  const [isLoading, setIsLoading] = useState(true);

  const { connect, disconnect, on, isConnected } = useWebSocket();

  // Initialize
  useEffect(() => {
    const init = async () => {
      try {
        await chatApi.init();
        const res = await chatApi.getCurrentUser();
        if (res.success && res.data) {
          setUser(res.data);
          const convRes = await chatApi.getConversations();
          if (convRes.success && convRes.data) {
            setConversations(convRes.data);
          }
          connect();
        }
      } catch (e) {
        console.error('Init error:', e);
      } finally {
        setIsLoading(false);
      }
    };
    init();
  }, []);

  // WebSocket handlers
  useEffect(() => {
    if (!isConnected) return;

    const cleanups = [
      on('message', (event: WebSocketEvent) => {
        const msg = event.payload as Message;
        if (activeConversation?.id === event.conversation_id) {
          setMessages((prev) => [...prev, msg]);
        }
        setConversations((prev) =>
          prev.map((c) =>
            c.id === event.conversation_id
              ? { ...c, last_message: msg, unread_count: c.unread_count + 1 }
              : c
          )
        );
      }),

      on('typing', (event: WebSocketEvent) => {
        if (event.user_id === user?.id) return;
        setTypingUsers((prev) => {
          const newMap = new Map(prev);
          const convTyping = newMap.get(event.conversation_id!) || [];
          if (!convTyping.includes(event.user_id!)) {
            newMap.set(event.conversation_id!, [...convTyping, event.user_id!]);
          }
          return newMap;
        });
        setTimeout(() => {
          setTypingUsers((prev) => {
            const newMap = new Map(prev);
            const convTyping = newMap.get(event.conversation_id!) || [];
            newMap.set(event.conversation_id!, convTyping.filter((id) => id !== event.user_id));
            return newMap;
          });
        }, 3000);
      }),

      on('online', (event: WebSocketEvent) => {
        setOnlineUsers((prev) => [...new Set([...prev, event.user_id!])]);
      }),

      on('offline', (event: WebSocketEvent) => {
        setOnlineUsers((prev) => prev.filter((id) => id !== event.user_id));
      }),

      on('conversation', (event: WebSocketEvent) => {
        const conv = event.payload as Conversation;
        setConversations((prev) => {
          const exists = prev.find((c) => c.id === conv.id);
          return exists ? prev : [conv, ...prev];
        });
      }),
    ];

    return () => cleanups.forEach((cleanup) => cleanup());
  }, [isConnected, on, activeConversation, user]);

  const login = async (email: string, password: string): Promise<boolean> => {
    const res = await chatApi.login(email, password);
    if (res.success && res.data) {
      setUser(res.data);
      const convRes = await chatApi.getConversations();
      if (convRes.success && convRes.data) {
        setConversations(convRes.data);
      }
      connect();
      return true;
    }
    return false;
  };

  const logout = async () => {
    await chatApi.logout();
    disconnect();
    setUser(null);
    setConversations([]);
    setActiveConversation(null);
    setMessages([]);
  };

  const selectConversation = async (conversation: Conversation) => {
    setActiveConversation(conversation);
    const res = await chatApi.getMessages(conversation.id);
    if (res.success && res.data) {
      setMessages(res.data);
    }
    await chatApi.markAsRead(conversation.id);
    setConversations((prev) =>
      prev.map((c) => (c.id === conversation.id ? { ...c, unread_count: 0 } : c))
    );
  };

  const clearActiveConversation = () => {
    setActiveConversation(null);
    setMessages([]);
  };

  const sendMessage = async (content: string, replyToId?: number) => {
    if (!activeConversation) return;
    await chatApi.sendMessage(activeConversation.id, content, 'text', replyToId);
  };

  const createDirectChat = async (userId: number) => {
    const res = await chatApi.createDirectChat(userId);
    if (res.success && res.data) {
      setConversations((prev) => [res.data!, ...prev]);
      return res.data;
    }
    return null;
  };

  const sendTypingIndicator = useCallback(() => {
    if (activeConversation) {
      chatApi.sendTyping(activeConversation.id);
    }
  }, [activeConversation]);

  const refreshConversations = async () => {
    const res = await chatApi.getConversations();
    if (res.success && res.data) {
      setConversations(res.data);
    }
  };

  return (
    <ChatContext.Provider
      value={{
        user, conversations, activeConversation, messages, onlineUsers, typingUsers,
        isLoading, isAuthenticated: !!user,
        login, logout, selectConversation, clearActiveConversation, sendMessage,
        createDirectChat, sendTypingIndicator, refreshConversations,
      }}
    >
      {children}
    </ChatContext.Provider>
  );
}

export const useChat = () => {
  const context = useContext(ChatContext);
  if (!context) throw new Error('useChat must be used within ChatProvider');
  return context;
};
```
