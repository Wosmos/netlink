import Constants from "expo-constants";
import { Platform } from "react-native";

// API Configuration
// For Android Emulator, the host machine is at 10.0.2.2
// For iOS Simulator, the host machine is at localhost
// For Physical Devices, we use the IP address where the Metro bundler is running
const getDevelopmentIP = () => {
  if (Platform.OS === "android") {
    return "10.0.2.2";
  }

  // Try to get the IP from Expo's host URI (most reliable for physical devices + Expo Go)
  const hostUri = Constants.expoConfig?.hostUri;
  if (hostUri) {
    return hostUri.split(":")[0];
  }

  // Fallback for iOS Simulator
  return "localhost";
};

const DEVELOPMENT_IP = getDevelopmentIP();

export const API_CONFIG = {
  BASE_URL: __DEV__
    ? `http://${DEVELOPMENT_IP}:8080`
    : "https://your-api-domain.com", // Production
  WS_URL: __DEV__
    ? `ws://${DEVELOPMENT_IP}:8080/ws`
    : "wss://your-api-domain.com/ws", // Production WebSocket
  TIMEOUT: 30000, // 30 seconds
} as const;

// App Configuration
export const APP_CONFIG = {
  NAME: "NetLink",
  VERSION: Constants.expoConfig?.version || "1.0.0",
  BUILD_NUMBER: Constants.expoConfig?.ios?.buildNumber || "1",
  BUNDLE_ID:
    Constants.expoConfig?.ios?.bundleIdentifier || "com.netlink.mobile",
} as const;

// Feature Flags
export const FEATURES = {
  VOICE_MESSAGES: true,
  IMAGE_MESSAGES: true,
  FILE_SHARING: true,
  PUSH_NOTIFICATIONS: true,
  BIOMETRIC_AUTH: true,
  DARK_MODE_ONLY: true,
} as const;

// Limits
export const LIMITS = {
  MAX_MESSAGE_LENGTH: 4000,
  MAX_VOICE_DURATION: 600, // 10 minutes in seconds
  MAX_FILE_SIZE: 50 * 1024 * 1024, // 50MB
  MAX_IMAGE_SIZE: 10 * 1024 * 1024, // 10MB
  MESSAGES_PER_PAGE: 50,
  SEARCH_DEBOUNCE: 300, // milliseconds
} as const;

// Storage Keys
export const STORAGE_KEYS = {
  AUTH_TOKEN: "auth_token",
  USER_DATA: "user_data",
  SETTINGS: "app_settings",
  CACHE_PREFIX: "cache_",
  CONVERSATIONS: "conversations",
  MESSAGES_PREFIX: "messages_",
} as const;
