'use client';

import { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { api, User } from '@/lib/api';
import { wsClient } from '@/lib/websocket';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<{ success: boolean; error?: string }>;
  register: (email: string, password: string, name?: string, phone?: string) => Promise<{ success: boolean; error?: string }>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    checkAuth();
  }, []);

  async function checkAuth() {
    try {
      const res = await api.me();
      if (res.success && res.data) {
        setUser(res.data);
        wsClient.connect();
      }
    } catch {
      // Not authenticated
    } finally {
      setLoading(false);
    }
  }

  async function login(email: string, password: string) {
    const res = await api.login(email, password);
    if (res.success && res.data) {
      setUser(res.data);
      wsClient.connect();
      return { success: true };
    }
    return { success: false, error: res.error || 'Login failed' };
  }

  async function register(email: string, password: string, name?: string, phone?: string) {
  try {
    const res = await api.register(email, password, name, phone);
    console.log('Registration response:', res);
    
    if (res.success) {
      return { success: true };
    }
    return { success: false, error: res.error || 'Registration failed' };
  } catch (error) {
    console.error('Registration error:', error);
    return { success: false, error: 'Network error' };
  }
}

  async function logout() {
    await api.logout();
    setUser(null);
    wsClient.disconnect();
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuth must be used within AuthProvider');
  return context;
}
