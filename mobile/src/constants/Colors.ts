export const Colors = {
  // Dark theme colors matching web app
  background: '#0a0a0f',
  surface: '#0c0c14',
  surfaceVariant: '#1a1a20',
  
  // Primary colors
  primary: '#00ffff', // cyan
  primaryDark: '#00cccc',
  primaryLight: '#66ffff',
  
  // Secondary colors
  secondary: '#ff6b35', // orange
  secondaryDark: '#cc5529',
  secondaryLight: '#ff8c66',
  
  // Text colors
  text: '#ffffff',
  textSecondary: '#cccccc',
  textMuted: '#888888',
  textDisabled: '#555555',
  
  // Border colors
  border: '#333333',
  borderLight: '#444444',
  borderPrimary: '#00ffff33',
  
  // Status colors
  success: '#10b981',
  error: '#ef4444',
  warning: '#f59e0b',
  info: '#3b82f6',
  
  // Online status
  online: '#10b981',
  offline: '#6b7280',
  
  // Message colors
  ownMessage: '#0c4a6e',
  otherMessage: '#1f2937',
  
  // Transparent overlays
  overlay: 'rgba(0, 0, 0, 0.7)',
  overlayLight: 'rgba(0, 0, 0, 0.3)',
} as const;

export type ColorKey = keyof typeof Colors;