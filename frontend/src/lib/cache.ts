// Smart caching system for instant conversation switching

interface CacheEntry<T> {
  data: T;
  timestamp: number;
  expiresAt?: number;
}

class SmartCache {
  private cache = new Map<string, CacheEntry<any>>();
  private subscribers = new Map<string, Set<(data: any) => void>>();

  // Get cached data
  get<T>(key: string): T | null {
    const entry = this.cache.get(key);
    if (!entry) return null;
    
    // Check if expired
    if (entry.expiresAt && Date.now() > entry.expiresAt) {
      this.cache.delete(key);
      return null;
    }
    
    return entry.data;
  }

  // Set cache with optional TTL (time to live in ms)
  set<T>(key: string, data: T, ttl?: number): void {
    const entry: CacheEntry<T> = {
      data,
      timestamp: Date.now(),
      expiresAt: ttl ? Date.now() + ttl : undefined,
    };
    
    this.cache.set(key, entry);
    
    // Notify subscribers
    this.notifySubscribers(key, data);
  }

  // Update cache without notifying (for background updates)
  silentUpdate<T>(key: string, data: T, ttl?: number): void {
    const entry: CacheEntry<T> = {
      data,
      timestamp: Date.now(),
      expiresAt: ttl ? Date.now() + ttl : undefined,
    };
    
    this.cache.set(key, entry);
  }

  // Subscribe to cache updates
  subscribe(key: string, callback: (data: any) => void): () => void {
    if (!this.subscribers.has(key)) {
      this.subscribers.set(key, new Set());
    }
    
    this.subscribers.get(key)!.add(callback);
    
    // Return unsubscribe function
    return () => {
      const subs = this.subscribers.get(key);
      if (subs) {
        subs.delete(callback);
        if (subs.size === 0) {
          this.subscribers.delete(key);
        }
      }
    };
  }

  // Notify all subscribers
  private notifySubscribers(key: string, data: any): void {
    const subs = this.subscribers.get(key);
    if (subs) {
      subs.forEach(callback => callback(data));
    }
  }

  // Invalidate cache
  invalidate(key: string): void {
    this.cache.delete(key);
  }

  // Invalidate by pattern
  invalidatePattern(pattern: RegExp): void {
    const keys = Array.from(this.cache.keys());
    keys.forEach(key => {
      if (pattern.test(key)) {
        this.cache.delete(key);
      }
    });
  }

  // Clear all cache
  clear(): void {
    this.cache.clear();
  }

  // Get cache size
  size(): number {
    return this.cache.size;
  }

  // Check if key exists and is valid
  has(key: string): boolean {
    const entry = this.cache.get(key);
    if (!entry) return false;
    
    if (entry.expiresAt && Date.now() > entry.expiresAt) {
      this.cache.delete(key);
      return false;
    }
    
    return true;
  }
}

// Global cache instance
export const cache = new SmartCache();

// Cache keys
export const CACHE_KEYS = {
  // Conversations (never expire, updated via WebSocket)
  conversations: () => 'conversations',
  
  // Messages per conversation (never expire, updated via WebSocket)
  messages: (convId: number) => `messages:${convId}`,
  
  // User data (cache for 5 minutes)
  user: (userId: number) => `user:${userId}`,
  
  // Online users (cache for 10 seconds)
  onlineUsers: () => 'online_users',
  
  // Tasks per conversation (never expire, updated on change)
  tasks: (convId?: number) => convId ? `tasks:${convId}` : 'tasks:personal',
  
  // Notes per conversation (never expire, updated on change)
  notes: (convId?: number) => convId ? `notes:${convId}` : 'notes:personal',
  
  // Search results (cache for 30 seconds)
  searchUsers: (query: string) => `search:${query}`,
};

// Helper to get or fetch data
export async function getOrFetch<T>(
  key: string,
  fetcher: () => Promise<T>,
  ttl?: number
): Promise<T> {
  // Check cache first
  const cached = cache.get<T>(key);
  if (cached !== null) {
    return cached;
  }
  
  // Fetch and cache
  const data = await fetcher();
  cache.set(key, data, ttl);
  return data;
}

// Prefetch data in background
export function prefetch<T>(
  key: string,
  fetcher: () => Promise<T>,
  ttl?: number
): void {
  // Don't prefetch if already cached
  if (cache.has(key)) return;
  
  // Fetch in background
  fetcher().then(data => {
    cache.silentUpdate(key, data, ttl);
  }).catch(err => {
    console.warn('Prefetch failed:', key, err);
  });
}
