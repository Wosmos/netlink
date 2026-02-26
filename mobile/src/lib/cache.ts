import AsyncStorage from '@react-native-async-storage/async-storage';
import { STORAGE_KEYS } from '@/constants/Config';
import { CacheItem } from '@/types';

class CacheManager {
  private memoryCache = new Map<string, CacheItem<any>>();
  private readonly DEFAULT_TTL = 5 * 60 * 1000; // 5 minutes

  async set<T>(key: string, data: T, ttl?: number): Promise<void> {
    const item: CacheItem<T> = {
      data,
      timestamp: Date.now(),
      ttl: ttl || this.DEFAULT_TTL,
    };

    // Store in memory cache
    this.memoryCache.set(key, item);

    // Store in persistent storage for important data
    if (this.shouldPersist(key)) {
      try {
        await AsyncStorage.setItem(
          `${STORAGE_KEYS.CACHE_PREFIX}${key}`,
          JSON.stringify(item)
        );
      } catch (error) {
        console.error('Failed to persist cache item:', error);
      }
    }
  }

  async get<T>(key: string): Promise<T | null> {
    // Check memory cache first
    const memoryItem = this.memoryCache.get(key);
    if (memoryItem && this.isValid(memoryItem)) {
      return memoryItem.data;
    }

    // Check persistent storage
    if (this.shouldPersist(key)) {
      try {
        const stored = await AsyncStorage.getItem(`${STORAGE_KEYS.CACHE_PREFIX}${key}`);
        if (stored) {
          const item: CacheItem<T> = JSON.parse(stored);
          if (this.isValid(item)) {
            // Restore to memory cache
            this.memoryCache.set(key, item);
            return item.data;
          } else {
            // Remove expired item
            await this.remove(key);
          }
        }
      } catch (error) {
        console.error('Failed to get cached item:', error);
      }
    }

    return null;
  }

  async remove(key: string): Promise<void> {
    this.memoryCache.delete(key);
    
    if (this.shouldPersist(key)) {
      try {
        await AsyncStorage.removeItem(`${STORAGE_KEYS.CACHE_PREFIX}${key}`);
      } catch (error) {
        console.error('Failed to remove cached item:', error);
      }
    }
  }

  async clear(): Promise<void> {
    this.memoryCache.clear();
    
    try {
      const keys = await AsyncStorage.getAllKeys();
      const cacheKeys = keys.filter(key => key.startsWith(STORAGE_KEYS.CACHE_PREFIX));
      await AsyncStorage.multiRemove(cacheKeys);
    } catch (error) {
      console.error('Failed to clear cache:', error);
    }
  }

  async invalidatePattern(pattern: RegExp): Promise<void> {
    // Clear from memory cache
    for (const key of this.memoryCache.keys()) {
      if (pattern.test(key)) {
        this.memoryCache.delete(key);
      }
    }

    // Clear from persistent storage
    try {
      const keys = await AsyncStorage.getAllKeys();
      const cacheKeys = keys.filter(key => {
        if (!key.startsWith(STORAGE_KEYS.CACHE_PREFIX)) return false;
        const cacheKey = key.replace(STORAGE_KEYS.CACHE_PREFIX, '');
        return pattern.test(cacheKey);
      });
      await AsyncStorage.multiRemove(cacheKeys);
    } catch (error) {
      console.error('Failed to invalidate cache pattern:', error);
    }
  }

  private isValid<T>(item: CacheItem<T>): boolean {
    if (!item.ttl) return true; // No expiration
    return Date.now() - item.timestamp < item.ttl;
  }

  private shouldPersist(key: string): boolean {
    // Persist important data like conversations and user data
    return key.includes('conversations') || 
           key.includes('user') || 
           key.includes('settings');
  }

  // Cleanup expired items
  async cleanup(): Promise<void> {
    const now = Date.now();
    
    // Cleanup memory cache
    for (const [key, item] of this.memoryCache.entries()) {
      if (!this.isValid(item)) {
        this.memoryCache.delete(key);
      }
    }

    // Cleanup persistent storage
    try {
      const keys = await AsyncStorage.getAllKeys();
      const cacheKeys = keys.filter(key => key.startsWith(STORAGE_KEYS.CACHE_PREFIX));
      
      for (const key of cacheKeys) {
        try {
          const stored = await AsyncStorage.getItem(key);
          if (stored) {
            const item: CacheItem<any> = JSON.parse(stored);
            if (!this.isValid(item)) {
              await AsyncStorage.removeItem(key);
            }
          }
        } catch (error) {
          // Remove corrupted items
          await AsyncStorage.removeItem(key);
        }
      }
    } catch (error) {
      console.error('Failed to cleanup cache:', error);
    }
  }
}

export const cache = new CacheManager();

// Cache key generators
export const CACHE_KEYS = {
  conversations: () => 'conversations',
  messages: (convId: number) => `messages:${convId}`,
  user: (userId: number) => `user:${userId}`,
  onlineUsers: () => 'online_users',
  searchUsers: (query: string) => `search_users:${query}`,
  tasks: (convId?: number) => convId ? `tasks:${convId}` : 'tasks',
  notes: (convId?: number) => convId ? `notes:${convId}` : 'notes',
} as const;