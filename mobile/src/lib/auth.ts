import AsyncStorage from '@react-native-async-storage/async-storage';
import * as SecureStore from 'expo-secure-store';
import { STORAGE_KEYS } from '@/constants/Config';
import { User } from '@/types';

// Use SecureStore for sensitive data, AsyncStorage for non-sensitive data
const isSecureStoreAvailable = SecureStore.isAvailableAsync();

export async function setAuthToken(token: string): Promise<void> {
  try {
    if (await isSecureStoreAvailable) {
      await SecureStore.setItemAsync(STORAGE_KEYS.AUTH_TOKEN, token);
    } else {
      await AsyncStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, token);
    }
  } catch (error) {
    console.error('Failed to store auth token:', error);
    throw error;
  }
}

export async function getAuthToken(): Promise<string | null> {
  try {
    if (await isSecureStoreAvailable) {
      return await SecureStore.getItemAsync(STORAGE_KEYS.AUTH_TOKEN);
    } else {
      return await AsyncStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
    }
  } catch (error) {
    console.error('Failed to get auth token:', error);
    return null;
  }
}

export async function removeAuthToken(): Promise<void> {
  try {
    if (await isSecureStoreAvailable) {
      await SecureStore.deleteItemAsync(STORAGE_KEYS.AUTH_TOKEN);
    } else {
      await AsyncStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
    }
  } catch (error) {
    console.error('Failed to remove auth token:', error);
  }
}

export async function setUserData(user: User): Promise<void> {
  try {
    await AsyncStorage.setItem(STORAGE_KEYS.USER_DATA, JSON.stringify(user));
  } catch (error) {
    console.error('Failed to store user data:', error);
    throw error;
  }
}

export async function getUserData(): Promise<User | null> {
  try {
    const userData = await AsyncStorage.getItem(STORAGE_KEYS.USER_DATA);
    return userData ? JSON.parse(userData) : null;
  } catch (error) {
    console.error('Failed to get user data:', error);
    return null;
  }
}

export async function removeUserData(): Promise<void> {
  try {
    await AsyncStorage.removeItem(STORAGE_KEYS.USER_DATA);
  } catch (error) {
    console.error('Failed to remove user data:', error);
  }
}

export async function clearAuthData(): Promise<void> {
  await Promise.all([
    removeAuthToken(),
    removeUserData(),
  ]);
}

export async function isAuthenticated(): Promise<boolean> {
  const token = await getAuthToken();
  const user = await getUserData();
  return !!(token && user);
}