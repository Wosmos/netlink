# NetLink Mobile - React Native Chat App

A performant, secure mobile chat application built with React Native and Expo, featuring real-time messaging, voice messages, tasks, and notes.

## 🚀 Features

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

## 🛠 Tech Stack

- **Framework**: React Native with Expo
- **State Management**: Zustand
- **Navigation**: Expo Router
- **Styling**: StyleSheet with custom design system
- **Audio**: Expo AV for voice messages
- **Storage**: AsyncStorage + SecureStore
- **WebSocket**: Native WebSocket with auto-reconnection
- **Icons**: Expo Vector Icons
- **Animations**: React Native Reanimated

## 📱 Screenshots

*Screenshots would go here in a real project*

## 🏗 Project Structure

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

## 🚀 Getting Started

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
   # or
   yarn install
   ```

3. **Configure environment**
   
   Update `src/constants/Config.ts` with your backend URL:
   ```typescript
   export const API_CONFIG = {
     BASE_URL: 'http://your-backend-url:8080',
     WS_URL: 'ws://your-backend-url:8080/ws',
   };
   ```

4. **Start the development server**
   ```bash
   npm start
   # or
   yarn start
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

## 🔧 Configuration

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

### Fonts

The app uses SpaceMono font. To add the actual font file:

1. Download SpaceMono from Google Fonts
2. Place the font file in `assets/fonts/`
3. Update the font loading in `_layout.tsx`

## 📦 Building for Production

### Development Build
```bash
npx expo build:android
npx expo build:ios
```

### EAS Build (Recommended)
```bash
# Install EAS CLI
npm install -g @expo/eas-cli

# Configure EAS
eas build:configure

# Build for Android
eas build --platform android

# Build for iOS
eas build --platform ios
```

## 🎨 Design System

The app uses a custom design system with:

- **Colors**: Cyberpunk-inspired dark theme
- **Typography**: SpaceMono monospace font
- **Components**: Consistent styling across all screens
- **Animations**: Smooth transitions and micro-interactions

### Color Palette

```typescript
Colors = {
  background: '#0a0a0f',    // Dark background
  surface: '#0c0c14',       // Card backgrounds
  primary: '#00ffff',       // Cyan accent
  secondary: '#ff6b35',     // Orange accent
  text: '#ffffff',          // Primary text
  textMuted: '#888888',     // Secondary text
  // ... more colors
}
```

## 🔐 Security Features

- **Secure token storage** using Expo SecureStore
- **HTTPS/WSS** for all network communication
- **Input validation** and sanitization
- **Secure file handling** for uploads
- **Session management** with automatic refresh

## 📊 Performance Optimizations

- **Lazy loading** of screens and components
- **Image optimization** with Expo Image
- **Efficient list rendering** with FlatList
- **Memory management** for voice messages
- **Caching strategy** for API responses
- **Bundle splitting** for faster startup

## 🧪 Testing

```bash
# Run tests (when implemented)
npm test

# Type checking
npx tsc --noEmit

# Linting
npx eslint .
```

## 📱 Platform-Specific Features

### iOS
- Native haptic feedback
- iOS-style navigation
- Background app refresh
- Push notifications

### Android
- Material Design elements
- Android-specific permissions
- Background processing
- Notification channels

## 🚀 Deployment

### App Store (iOS)
1. Build with EAS
2. Upload to App Store Connect
3. Submit for review

### Google Play (Android)
1. Build signed APK/AAB
2. Upload to Google Play Console
3. Submit for review

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## 🆘 Troubleshooting

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

### Performance Issues

- Enable Flipper for debugging
- Use React DevTools for component analysis
- Monitor memory usage with Xcode/Android Studio

## 📞 Support

For support and questions:
- Create an issue on GitHub
- Check the documentation
- Review the troubleshooting guide

---

Built with ❤️ using React Native and Expo