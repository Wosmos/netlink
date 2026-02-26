import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Colors } from '@/constants/Colors';
import { Conversation, ConversationItemProps } from '@/types';
import { useChatStore } from '@/store/chatStore';

export default function ConversationItem({ 
  conversation, 
  currentUserId, 
  onPress, 
  onLongPress 
}: ConversationItemProps) {
  const { onlineUsers } = useChatStore();

  const getConversationName = () => {
    if (conversation.type === 'group') {
      return conversation.name || 'Group Chat';
    }
    
    const otherMember = conversation.members?.find(m => m.user_id !== currentUserId);
    return otherMember?.user?.name || otherMember?.user?.email || 'Unknown User';
  };

  const getOtherUserId = () => {
    if (conversation.type === 'group') return null;
    return conversation.members?.find(m => m.user_id !== currentUserId)?.user_id;
  };

  const isOnline = () => {
    const otherId = getOtherUserId();
    return otherId ? onlineUsers.includes(otherId) : false;
  };

  const formatTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60);
    
    if (diffInHours < 24) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } else if (diffInHours < 168) { // 7 days
      return date.toLocaleDateString([], { weekday: 'short' });
    } else {
      return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
    }
  };

  const truncateMessage = (content: string, maxLength = 50) => {
    if (content.length <= maxLength) return content;
    return content.substring(0, maxLength) + '...';
  };

  return (
    <TouchableOpacity
      style={styles.container}
      onPress={() => onPress(conversation)}
      onLongPress={() => onLongPress?.(conversation)}
      activeOpacity={0.7}
    >
      {/* Avatar */}
      <View style={styles.avatarContainer}>
        <View style={[
          styles.avatar,
          conversation.type === 'group' ? styles.groupAvatar : styles.directAvatar
        ]}>
          {conversation.type === 'group' ? (
            <Ionicons name="people" size={20} color={Colors.secondary} />
          ) : (
            <Text style={styles.avatarText}>
              {getConversationName().charAt(0).toUpperCase()}
            </Text>
          )}
        </View>
        
        {/* Online indicator for direct chats */}
        {conversation.type === 'direct' && (
          <View style={[
            styles.onlineIndicator,
            { backgroundColor: isOnline() ? Colors.online : Colors.offline }
          ]} />
        )}
      </View>

      {/* Content */}
      <View style={styles.content}>
        <View style={styles.header}>
          <Text style={styles.name} numberOfLines={1}>
            {getConversationName()}
          </Text>
          {conversation.last_message && (
            <Text style={styles.time}>
              {formatTime(conversation.last_message.created_at)}
            </Text>
          )}
        </View>

        <View style={styles.messageRow}>
          {conversation.last_message ? (
            <Text style={styles.lastMessage} numberOfLines={1}>
              {conversation.last_message.type === 'voice' 
                ? '🎵 Voice message'
                : conversation.last_message.type === 'image'
                ? '📷 Image'
                : conversation.last_message.type === 'file'
                ? '📎 File'
                : truncateMessage(conversation.last_message.content)
              }
            </Text>
          ) : (
            <Text style={styles.noMessages}>No messages yet</Text>
          )}
          
          {conversation.unread_count > 0 && (
            <View style={styles.unreadBadge}>
              <Text style={styles.unreadText}>
                {conversation.unread_count > 99 ? '99+' : conversation.unread_count}
              </Text>
            </View>
          )}
        </View>
      </View>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: Colors.border,
    backgroundColor: Colors.background,
  },
  avatarContainer: {
    position: 'relative',
    marginRight: 16,
  },
  avatar: {
    width: 48,
    height: 48,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 1,
  },
  directAvatar: {
    backgroundColor: Colors.surfaceVariant,
    borderColor: Colors.borderPrimary,
  },
  groupAvatar: {
    backgroundColor: Colors.surfaceVariant,
    borderColor: Colors.secondary + '50',
  },
  avatarText: {
    fontSize: 18,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.primary,
  },
  onlineIndicator: {
    position: 'absolute',
    bottom: 2,
    right: 2,
    width: 12,
    height: 12,
    borderRadius: 6,
    borderWidth: 2,
    borderColor: Colors.background,
  },
  content: {
    flex: 1,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 4,
  },
  name: {
    fontSize: 16,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.text,
    flex: 1,
    marginRight: 8,
  },
  time: {
    fontSize: 11,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
  },
  messageRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  lastMessage: {
    fontSize: 13,
    fontFamily: 'SpaceMono',
    color: Colors.textSecondary,
    flex: 1,
    marginRight: 8,
  },
  noMessages: {
    fontSize: 13,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
    fontStyle: 'italic',
    flex: 1,
  },
  unreadBadge: {
    backgroundColor: Colors.primary,
    borderRadius: 10,
    minWidth: 20,
    height: 20,
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 6,
  },
  unreadText: {
    fontSize: 10,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.background,
  },
});