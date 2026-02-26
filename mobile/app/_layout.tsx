import { useEffect } from "react";
import { Stack } from "expo-router";
import { StatusBar } from "expo-status-bar";
import { GestureHandlerRootView } from "react-native-gesture-handler";
import * as SplashScreen from "expo-splash-screen";
import * as Font from "expo-font";
import { Colors } from "@/constants/Colors";
import { useAuthStore } from "@/store/authStore";
import { useChatStore } from "@/store/chatStore";
import { wsClient } from "@/lib/websocket";
import { cache } from "@/lib/cache";

// Keep the splash screen visible while we fetch resources
SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const { checkAuth, isAuthenticated, user } = useAuthStore();
  const chatStore = useChatStore();

  useEffect(() => {
    async function prepare() {
      try {
        // Load fonts
        await Font.loadAsync({
          SpaceMono: require("../assets/fonts/SpaceMono-Regular.ttf"),
        });

        // Check authentication
        await checkAuth();

        // Cleanup cache periodically
        await cache.cleanup();
      } catch (e) {
        console.warn(e);
      } finally {
        // Hide splash screen
        await SplashScreen.hideAsync();
      }
    }

    prepare();
  }, []);

  useEffect(() => {
    let unsubscribe: (() => void)[] = [];

    if (isAuthenticated && user) {
      // Message events
      unsubscribe.push(
        wsClient.on("message", (event) => {
          if (event.payload) {
            chatStore.handleNewMessage(event.payload);
          }
        }),
      );

      // Typing events
      unsubscribe.push(
        wsClient.on("typing", (event) => {
          if (event.conversation_id && event.user_id) {
            chatStore.handleTyping(
              event.conversation_id,
              event.user_id,
              event.payload?.user_name || "Someone",
            );
          }
        }),
      );

      unsubscribe.push(
        wsClient.on("stop_typing", (event) => {
          if (event.conversation_id && event.user_id) {
            chatStore.handleStopTyping(event.conversation_id, event.user_id);
          }
        }),
      );

      // Online status events
      unsubscribe.push(
        wsClient.on("online", (event) => {
          if (event.user_id) {
            chatStore.handleUserOnline(event.user_id);
          }
        }),
      );

      unsubscribe.push(
        wsClient.on("offline", (event) => {
          if (event.user_id) {
            chatStore.handleUserOffline(event.user_id);
          }
        }),
      );

      // Read events
      unsubscribe.push(
        wsClient.on("read", (event) => {
          if (event.conversation_id && event.user_id && event.message_id) {
            chatStore.handleMessageRead(
              event.conversation_id,
              event.user_id,
              event.message_id,
            );
          }
        }),
      );

      // Edit events
      unsubscribe.push(
        wsClient.on("message_edit", (event) => {
          if (event.payload) {
            chatStore.handleMessageEdit(event.payload);
          }
        }),
      );

      // Delete events
      unsubscribe.push(
        wsClient.on("message_delete", (event) => {
          if (event.conversation_id && event.message_id) {
            chatStore.handleMessageDelete(
              event.conversation_id,
              event.message_id,
            );
          }
        }),
      );

      // Reaction events
      unsubscribe.push(
        wsClient.on("reaction", (event) => {
          if (event.conversation_id && event.payload) {
            const { emoji, added, reactions, message_id } = event.payload;
            const targetMessageId = event.message_id || message_id;

            if (targetMessageId) {
              chatStore.handleReaction(
                event.conversation_id,
                targetMessageId,
                event.user_id || 0,
                emoji,
                added,
                reactions,
              );
            }
          }
        }),
      );

      // Conversation events
      unsubscribe.push(
        wsClient.on("conversation", () => {
          // Reload conversations when new ones are created
          chatStore.loadConversations();
        }),
      );
    }

    return () => {
      unsubscribe.forEach((fn) => fn());
    };
  }, [isAuthenticated, user]);

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <StatusBar style="light" backgroundColor={Colors.background} />
      <Stack
        screenOptions={{
          headerStyle: {
            backgroundColor: Colors.surface,
          },
          headerTintColor: Colors.primary,
          headerTitleStyle: {
            fontFamily: "SpaceMono",
            fontSize: 16,
            fontWeight: "bold",
          },
          contentStyle: {
            backgroundColor: Colors.background,
          },
        }}
      >
        <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
        <Stack.Screen name="auth/login" options={{ headerShown: false }} />
        <Stack.Screen name="auth/register" options={{ headerShown: false }} />
        <Stack.Screen
          name="auth/forgot-password"
          options={{ headerShown: false }}
        />
        <Stack.Screen
          name="chat/[id]"
          options={{
            headerShown: false,
            presentation: "card",
          }}
        />
      </Stack>
    </GestureHandlerRootView>
  );
}
