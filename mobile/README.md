# NetLink Mobile - React Native Chat App

A performant, secure mobile chat application built with React Native and Expo, featuring real-time messaging, voice messages, tasks, and notes.

## Features

### Core Chat Features
- **Real-time messaging** with WebSocket support
- **Voice messages** with waveform visualization
- **Message reactions** and replies
- **Group chats** and direct messaging
- **Typing indicators** and read receipts
- **Message editing** and deletion
- **File sharing** (images, documents)

### Additional Features
- **Tasks management** with completion tracking
- **Notes system** with pinning capability
- **User search** and contact management
- **Online status** indicators
- **Push notifications** (configurable)
- **Dark theme** with cyberpunk aesthetic

### Technical Features
- **Offline support** with intelligent caching
- **Optimistic updates** for smooth UX
- **Auto-reconnecting WebSocket**
- **Secure authentication** with token storage
- **Performance optimized** with lazy loading
- **Cross-platform** (iOS, Android, Web)

## Tech Stack

- **Framework**: React Native with Expo (SDK 54)
- **State Management**: Zustand
- **Navigation**: Expo Router v3
- **Styling**: StyleSheet with custom design system
- **Audio**: Expo AV for voice messages
- **Storage**: AsyncStorage + SecureStore
- **WebSocket**: Native WebSocket with auto-reconnection
- **Icons**: Expo Vector Icons
- **Animations**: React Native Reanimated

## Project Structure

```
mobile/
├── app/                    # App Router pages
│   ├── (tabs)/            # Tab navigation
│   ├── auth/              # Authentication screens
│   └── chat/              # Chat screens
├── src/
│   ├── components/        # Reusable components
│   ├── constants/         # App constants and config
│   ├── lib/              # Utilities and API client
│   ├── store/            # Zustand stores
│   └── types/            # TypeScript definitions
└── assets/               # Static assets
```

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn
- Expo CLI
- iOS Simulator (for iOS development)
- Android Studio (for Android development)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd mobile
   ```

2. **Install dependencies**
   ```bash
   npm install
   ```

3. **Configure environment**

   Update `src/constants/Config.ts` with your backend URL:
   ```typescript
   const PRODUCTION_API = "https://your-actual-backend.up.railway.app";
   ```

4. **Start the development server**
   ```bash
   npm start
   ```

5. **Run on device/simulator**
   ```bash
   # iOS
   npm run ios

   # Android
   npm run android

   # Web
   npm run web
   ```

## Configuration

### Backend Integration

The app is designed to work with the Go backend in the `backend/` directory. Make sure to:

1. Start the backend server
2. Update API URLs in `Config.ts`
3. Ensure CORS is configured for mobile requests

### Push Notifications

To enable push notifications:

1. Configure Expo notifications in `app.json`
2. Set up push notification credentials
3. Update notification settings in the app

## Building for Production

### EAS Build (Recommended)
```bash
npm install -g @expo/eas-cli
eas build:configure
eas build --platform android
eas build --platform ios
```

## Design System

The app uses a custom design system with:

- **Colors**: Cyberpunk-inspired dark theme
- **Typography**: SpaceMono monospace font
- **Components**: Consistent styling across all screens
- **Animations**: Smooth transitions and micro-interactions

### Color Palette

| Token | Value | Usage |
|---|---|---|
| `background` | `#0a0a0f` | Dark background |
| `surface` | `#0c0c14` | Card backgrounds |
| `primary` | `#00ffff` | Cyan accent |
| `secondary` | `#ff6b35` | Orange accent |
| `text` | `#ffffff` | Primary text |
| `textMuted` | `#888888` | Secondary text |

---

## Known Issues & Bug Report

### Critical

| # | Issue | Location | Description |
|---|---|---|---|
| 1 | Voice message data never sent to server | `chatStore.ts:273-280` | `voiceData` is built from the upload response but never passed to `sendMessage`. The file uploads successfully but only plain text "Voice message" is stored. |
| 2 | `Alert.prompt` crashes on Android | `MessageItem.tsx:51` | `Alert.prompt()` is iOS-only. Tapping "Edit" on a message will crash the app on Android. |
| 3 | Waveform interval never cleared (memory leak) | `VoiceRecorder.tsx:112-133` | `startWaveformVisualization()` creates a `setInterval` but the returned cleanup function is never captured or called. The interval leaks indefinitely. |
| 4 | Missing `SafeAreaProvider` | `_layout.tsx` | Chat screen uses `SafeAreaView` from `react-native-safe-area-context`, but no `SafeAreaProvider` exists in the app tree. Expo Router v3 does not provide one. Safe area insets default to zero — content renders under notches/status bars. |
| 5 | `WebSocketEvent.type` missing event types | `types/index.ts:108-116` | The union type doesn't include `"message_edit"` or `"message_delete"`, but `_layout.tsx` subscribes to both. These events won't be properly typed at runtime. |

### Bugs

| # | Issue | Location | Description |
|---|---|---|---|
| 6 | `parseInt(undefined)` produces `NaN` | `chat/[id].tsx:31` | `id` from route params can be `undefined`. `parseInt(undefined)` returns `NaN`, rendering an empty chat with no error. |
| 7 | FlatList not inverted — chat UX broken | `chat/[id].tsx:261-272` | Without `inverted={true}`, newest messages aren't auto-visible at bottom. `onEndReached` fires at the wrong end (bottom instead of top), breaking pagination. |
| 8 | Unread count always increments | `chatStore.ts:480` | `handleNewMessage` increments `unread_count` unconditionally — even for the current user's own messages or when the conversation is actively open. |
| 9 | Reply button missing from action menu | `MessageItem.tsx:242-270` | The modal renders React/Edit/Delete but no Reply button. `handleAction("reply")` case exists but is unreachable from the UI. |
| 10 | Logout doesn't clear `token` on success | `authStore.ts:128-133` | The success-path `set()` call is missing `token: null`. The error path (line 143) clears it. Stale token persists in state after successful logout. |
| 11 | Online status is hardcoded | `chat/[id].tsx:236` | Chat header always shows "Online" regardless of actual user status. Should check `onlineUsers` from the store. |
| 12 | Stale state in `loadMessages` pagination | `chatStore.ts:110-157` | `state` is captured once at function entry. By the time the API resolves, `state.messages` may be stale. Should use `get().messages` at update time. |
| 13 | Optimistic message matching is fragile | `chatStore.ts:460-461` | Matches by `msg.sending && msg.content === message.content`. If the user sends identical text twice quickly, the wrong message gets replaced. |
| 14 | Double `checkAuth()` on startup | `index.tsx:11` + `_layout.tsx:29` | Both call `checkAuth()` on mount, causing duplicate API calls to `/api/auth/me`. |

### Logic / Runtime Issues

| # | Issue | Location | Description |
|---|---|---|---|
| 15 | Stale closure in waveform visualization | `VoiceRecorder.tsx:126-127` | `isRecording` and `isPaused` inside the interval capture the initial values and never update. |
| 16 | `handleSend` called from stale closure in timer | `VoiceRecorder.tsx:106-108` | `handleSend` inside `setInterval` captures the initial render's version and never picks up re-renders. |
| 17 | `uploadFile` XHR rejection produces wrong error shape | `api.ts:344-345` | Rejects with `{ success: false, error: 'Upload failed' }`. Caller reads `error.message` which is `undefined`. |
| 18 | `uploadFile` XHR has no timeout | `api.ts:326-353` | Unlike `request()` (30s timeout), the XHR upload has no timeout. A hanging server blocks forever. |
| 19 | Stale timeout in NewChatModal search debounce | `NewChatModal.tsx:38-66` | `searchTimeout` in the effect cleanup captures the stale value from the previous render. |
| 20 | `handleTyping` useCallback recreates on every call | `chat/[id].tsx:124-138` | `typingTimeout` in deps causes constant recreation. The empty `setTimeout` callback serves no purpose. |
| 21 | Settings state not persisted | `settings.tsx:18-20` | Notification/sound/vibration toggles use local `useState` only. Values reset on every app launch. |

### Type Issues

| # | Issue | Location | Description |
|---|---|---|---|
| 22 | `type: type as any` hides invalid types | `chatStore.ts:187` | Casts away type safety on `Message.type`. |
| 23 | `Blob` type in React Native | `chatStore.ts:28-29` | `sendVoiceMessage` uses `Blob` (Web API). React Native's Blob polyfill is inconsistent. Should use file URI. |
| 24 | Unused imports | `chat/[id].tsx:12-13` | `Keyboard` and `Modal` imported but never used. |

### Config / Build Issues

| # | Issue | Location | Description |
|---|---|---|---|
| 25 | TypeScript version mismatch | `package.json` | Has `typescript: ~5.3.3` but Expo SDK 54 base tsconfig uses `module: "preserve"` (requires TS 5.4+). Works only because `skipLibCheck: true` hides the incompatibility. |
| 26 | Hardcoded placeholder production URL | `Config.ts:26` | `PRODUCTION_API` is `"https://your-backend.up.railway.app"`. Production builds will point to a non-existent server. |
| 27 | `app.json` missing icon and splash config | `app.json` | No `icon`, `splash`, or `adaptiveIcon` fields defined. Production builds will use Expo defaults or fail store review. |
| 28 | `credentials: 'include'` is a no-op | `api.ts:29` | Browser-specific fetch option. Unnecessary in React Native where auth uses Bearer tokens. |
| 29 | `console.log` left in API client | `api.ts:24,39` | Logs every request URL and response status. Will spam device logs in production. |

### Recently Fixed

- **Distorted voice playback** — Voice messages previously played back with audio distortion. This has been resolved.

---

## Troubleshooting

### Common Issues

**Metro bundler issues**
```bash
npx expo start --clear
```

**iOS build issues**
```bash
cd ios && pod install
```

**Android build issues**
```bash
cd android && ./gradlew clean
```

**WebSocket connection issues**
- Check backend URL configuration
- Verify network connectivity
- Check firewall settings

---

Built with React Native and Expo
