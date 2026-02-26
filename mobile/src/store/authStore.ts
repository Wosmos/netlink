import { create } from "zustand";
import { User, LoginCredentials, RegisterCredentials } from "@/types";
import { api } from "@/lib/api";
import {
  setAuthToken,
  getAuthToken,
  removeAuthToken,
  setUserData,
  removeUserData,
  getUserData,
  isAuthenticated,
} from "@/lib/auth";
import { wsClient } from "@/lib/websocket";
import { cache } from "@/lib/cache";

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  error: string | null;

  // Actions
  login: (credentials: LoginCredentials) => Promise<boolean>;
  register: (credentials: RegisterCredentials) => Promise<boolean>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
  clearError: () => void;
  updateUser: (user: Partial<User>) => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  token: null,
  isLoading: false,
  isAuthenticated: false,
  error: null,

  login: async (credentials: LoginCredentials) => {
    set({ isLoading: true, error: null });

    try {
      const response = await api.login(credentials);

      if (response.success && response.data) {
        // Store auth data
        await setUserData(response.data);
        if (response.token) {
          await setAuthToken(response.token);
        }

        // Connect WebSocket
        try {
          await wsClient.connect();
        } catch (wsError) {
          console.warn("WebSocket connection failed:", wsError);
          // Don't fail login if WebSocket fails
        }

        set({
          user: response.data,
          token: response.token || null,
          isAuthenticated: true,
          isLoading: false,
          error: null,
        });

        return true;
      } else {
        set({
          error: response.error || "Login failed",
          isLoading: false,
        });
        return false;
      }
    } catch (error: any) {
      set({
        error: error.message || "Login failed",
        isLoading: false,
      });
      return false;
    }
  },

  register: async (credentials: RegisterCredentials) => {
    set({ isLoading: true, error: null });

    try {
      const response = await api.register(credentials);

      if (response.success) {
        set({
          isLoading: false,
          error: null,
        });
        return true;
      } else {
        set({
          error: response.error || "Registration failed",
          isLoading: false,
        });
        return false;
      }
    } catch (error: any) {
      set({
        error: error.message || "Registration failed",
        isLoading: false,
      });
      return false;
    }
  },

  logout: async () => {
    set({ isLoading: true });

    try {
      // Disconnect WebSocket
      wsClient.disconnect();

      // Call logout API
      await api.logout();

      // Clear local data
      await removeAuthToken();
      await removeUserData();
      await cache.clear();

      set({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      });
    } catch (error: any) {
      console.error("Logout error:", error);

      // Force logout even if API call fails
      await removeAuthToken();
      await removeUserData();
      await cache.clear();
      wsClient.disconnect();

      set({
        user: null,
        token: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      });
    }
  },

  checkAuth: async () => {
    set({ isLoading: true });

    try {
      const authenticated = await isAuthenticated();

      if (authenticated) {
        const userData = await getUserData();

        if (userData) {
          // Verify with server
          const response = await api.getCurrentUser();

          if (response.success && response.data) {
            // Update user data if server has newer info
            await setUserData(response.data);

            // Connect WebSocket
            try {
              await wsClient.connect();
            } catch (wsError) {
              console.warn("WebSocket connection failed:", wsError);
            }

            set({
              user: response.data,
              token: (await getAuthToken()) || null,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            });
          } else {
            // Server auth failed, clear local data
            await removeAuthToken();
            await removeUserData();

            set({
              user: null,
              isAuthenticated: false,
              isLoading: false,
              error: null,
            });
          }
        } else {
          set({
            user: null,
            isAuthenticated: false,
            isLoading: false,
            error: null,
          });
        }
      } else {
        set({
          user: null,
          isAuthenticated: false,
          isLoading: false,
          error: null,
        });
      }
    } catch (error: any) {
      console.error("Auth check error:", error);

      // Clear auth data on error
      await removeAuthToken();
      await removeUserData();

      set({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      });
    }
  },

  clearError: () => {
    set({ error: null });
  },

  updateUser: (userData: Partial<User>) => {
    const currentUser = get().user;
    if (currentUser) {
      const updatedUser = { ...currentUser, ...userData };
      set({ user: updatedUser });
      setUserData(updatedUser);
    }
  },
}));
