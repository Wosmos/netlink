import { useState, useEffect } from 'react';
import {
  View,
  Text,
  Modal,
  TextInput,
  TouchableOpacity,
  FlatList,
  StyleSheet,
  Alert,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Colors } from '@/constants/Colors';
import { User, Conversation } from '@/types';
import { api } from '@/lib/api';
import { useChatStore } from '@/store/chatStore';
import { LIMITS } from '@/constants/Config';

interface NewChatModalProps {
  visible: boolean;
  onClose: () => void;
  onConversationCreated: (conversation: Conversation) => void;
}

export default function NewChatModal({ visible, onClose, onConversationCreated }: NewChatModalProps) {
  const [chatType, setChatType] = useState<'direct' | 'group'>('direct');
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [selectedUsers, setSelectedUsers] = useState<User[]>([]);
  const [groupName, setGroupName] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [searchTimeout, setSearchTimeout] = useState<NodeJS.Timeout | null>(null);

  const { createDirectChat, createGroup } = useChatStore();

  useEffect(() => {
    if (searchQuery.length >= 2) {
      // Debounce search
      if (searchTimeout) {
        clearTimeout(searchTimeout);
      }
      
      const timeout = setTimeout(async () => {
        try {
          const response = await api.searchUsers(searchQuery);
          if (response.success && response.data) {
            setSearchResults(response.data);
          }
        } catch (error) {
          console.error('Search error:', error);
        }
      }, LIMITS.SEARCH_DEBOUNCE);
      
      setSearchTimeout(timeout);
    } else {
      setSearchResults([]);
    }

    return () => {
      if (searchTimeout) {
        clearTimeout(searchTimeout);
      }
    };
  }, [searchQuery]);

  const handleClose = () => {
    setSearchQuery('');
    setSearchResults([]);
    setSelectedUsers([]);
    setGroupName('');
    setChatType('direct');
    onClose();
  };

  const handleUserSelect = async (user: User) => {
    if (chatType === 'direct') {
      setIsLoading(true);
      try {
        const conversation = await createDirectChat(user.id);
        if (conversation) {
          onConversationCreated(conversation);
        } else {
          Alert.alert('Error', 'Failed to create conversation');
        }
      } catch (error) {
        Alert.alert('Error', 'Failed to create conversation');
      } finally {
        setIsLoading(false);
      }
    } else {
      // Toggle user selection for group
      const isSelected = selectedUsers.some(u => u.id === user.id);
      if (isSelected) {
        setSelectedUsers(prev => prev.filter(u => u.id !== user.id));
      } else {
        setSelectedUsers(prev => [...prev, user]);
      }
    }
  };

  const handleCreateGroup = async () => {
    if (!groupName.trim()) {
      Alert.alert('Error', 'Please enter a group name');
      return;
    }

    if (selectedUsers.length === 0) {
      Alert.alert('Error', 'Please select at least one member');
      return;
    }

    setIsLoading(true);
    try {
      const conversation = await createGroup(groupName.trim(), selectedUsers.map(u => u.id));
      if (conversation) {
        onConversationCreated(conversation);
      } else {
        Alert.alert('Error', 'Failed to create group');
      }
    } catch (error) {
      Alert.alert('Error', 'Failed to create group');
    } finally {
      setIsLoading(false);
    }
  };

  const renderUser = ({ item }: { item: User }) => {
    const isSelected = selectedUsers.some(u => u.id === item.id);
    
    return (
      <TouchableOpacity
        style={[styles.userItem, isSelected && styles.userItemSelected]}
        onPress={() => handleUserSelect(item)}
        disabled={isLoading}
      >
        <View style={styles.userAvatar}>
          <Text style={styles.userAvatarText}>
            {(item.name || item.email).charAt(0).toUpperCase()}
          </Text>
        </View>
        
        <View style={styles.userInfo}>
          <Text style={styles.userName} numberOfLines={1}>
            {item.name || item.email.split('@')[0]}
          </Text>
          {item.name && (
            <Text style={styles.userEmail} numberOfLines={1}>
              {item.email}
            </Text>
          )}
        </View>
        
        {chatType === 'group' && isSelected && (
          <Ionicons name="checkmark-circle" size={24} color={Colors.primary} />
        )}
      </TouchableOpacity>
    );
  };

  return (
    <Modal
      visible={visible}
      animationType="slide"
      presentationStyle="pageSheet"
      onRequestClose={handleClose}
    >
      <KeyboardAvoidingView 
        style={styles.container}
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      >
        {/* Header */}
        <View style={styles.header}>
          <TouchableOpacity onPress={handleClose} style={styles.closeButton}>
            <Ionicons name="close" size={24} color={Colors.text} />
          </TouchableOpacity>
          <Text style={styles.headerTitle}>NEW CONNECTION</Text>
          <View style={styles.placeholder} />
        </View>

        {/* Chat Type Toggle */}
        <View style={styles.toggleContainer}>
          <TouchableOpacity
            style={[styles.toggleButton, chatType === 'direct' && styles.toggleButtonActive]}
            onPress={() => setChatType('direct')}
          >
            <Text style={[styles.toggleText, chatType === 'direct' && styles.toggleTextActive]}>
              DIRECT
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[styles.toggleButton, chatType === 'group' && styles.toggleButtonActive]}
            onPress={() => setChatType('group')}
          >
            <Text style={[styles.toggleText, chatType === 'group' && styles.toggleTextActive]}>
              GROUP
            </Text>
          </TouchableOpacity>
        </View>

        {/* Group Name Input */}
        {chatType === 'group' && (
          <View style={styles.inputContainer}>
            <TextInput
              style={styles.input}
              placeholder="GROUP NAME"
              placeholderTextColor={Colors.textMuted}
              value={groupName}
              onChangeText={setGroupName}
              maxLength={50}
            />
          </View>
        )}

        {/* Search Input */}
        <View style={styles.inputContainer}>
          <Ionicons name="search" size={20} color={Colors.primary} style={styles.searchIcon} />
          <TextInput
            style={styles.searchInput}
            placeholder="SEARCH USERS..."
            placeholderTextColor={Colors.textMuted}
            value={searchQuery}
            onChangeText={setSearchQuery}
            autoCapitalize="none"
            autoCorrect={false}
          />
        </View>

        {/* Selected Users (Group only) */}
        {chatType === 'group' && selectedUsers.length > 0 && (
          <View style={styles.selectedContainer}>
            <Text style={styles.selectedTitle}>SELECTED ({selectedUsers.length})</Text>
            <FlatList
              data={selectedUsers}
              horizontal
              showsHorizontalScrollIndicator={false}
              keyExtractor={(item) => item.id.toString()}
              renderItem={({ item }) => (
                <View style={styles.selectedUser}>
                  <Text style={styles.selectedUserText} numberOfLines={1}>
                    {item.name || item.email.split('@')[0]}
                  </Text>
                  <TouchableOpacity
                    onPress={() => setSelectedUsers(prev => prev.filter(u => u.id !== item.id))}
                    style={styles.removeButton}
                  >
                    <Ionicons name="close" size={16} color={Colors.text} />
                  </TouchableOpacity>
                </View>
              )}
            />
          </View>
        )}

        {/* Search Results */}
        <FlatList
          data={searchResults}
          keyExtractor={(item) => item.id.toString()}
          renderItem={renderUser}
          style={styles.userList}
          ListEmptyComponent={
            searchQuery.length >= 2 ? (
              <View style={styles.emptyState}>
                <Text style={styles.emptyText}>NO USERS FOUND</Text>
              </View>
            ) : (
              <View style={styles.emptyState}>
                <Text style={styles.emptyText}>ENTER NAME OR EMAIL TO SEARCH</Text>
              </View>
            )
          }
        />

        {/* Create Group Button */}
        {chatType === 'group' && (
          <TouchableOpacity
            style={[styles.createButton, (!groupName.trim() || selectedUsers.length === 0 || isLoading) && styles.createButtonDisabled]}
            onPress={handleCreateGroup}
            disabled={!groupName.trim() || selectedUsers.length === 0 || isLoading}
          >
            <Text style={styles.createButtonText}>
              {isLoading ? 'CREATING...' : `CREATE GROUP (${selectedUsers.length})`}
            </Text>
          </TouchableOpacity>
        )}
      </KeyboardAvoidingView>
    </Modal>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.background,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 20,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: Colors.border,
  },
  closeButton: {
    width: 40,
    height: 40,
    alignItems: 'center',
    justifyContent: 'center',
  },
  headerTitle: {
    fontSize: 16,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.primary,
    letterSpacing: 2,
  },
  placeholder: {
    width: 40,
  },
  toggleContainer: {
    flexDirection: 'row',
    margin: 20,
    backgroundColor: Colors.surfaceVariant,
    borderRadius: 8,
    padding: 4,
  },
  toggleButton: {
    flex: 1,
    paddingVertical: 12,
    alignItems: 'center',
    borderRadius: 6,
  },
  toggleButtonActive: {
    backgroundColor: Colors.primary,
  },
  toggleText: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.textMuted,
    letterSpacing: 2,
  },
  toggleTextActive: {
    color: Colors.background,
  },
  inputContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: Colors.surfaceVariant,
    borderWidth: 1,
    borderColor: Colors.border,
    borderRadius: 8,
    marginHorizontal: 20,
    marginBottom: 16,
    paddingHorizontal: 16,
    height: 48,
  },
  input: {
    flex: 1,
    fontSize: 14,
    fontFamily: 'SpaceMono',
    color: Colors.text,
    letterSpacing: 1,
  },
  searchIcon: {
    marginRight: 12,
  },
  searchInput: {
    flex: 1,
    fontSize: 14,
    fontFamily: 'SpaceMono',
    color: Colors.text,
    letterSpacing: 1,
  },
  selectedContainer: {
    paddingHorizontal: 20,
    marginBottom: 16,
  },
  selectedTitle: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.textMuted,
    letterSpacing: 1,
    marginBottom: 8,
  },
  selectedUser: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: Colors.surfaceVariant,
    borderWidth: 1,
    borderColor: Colors.borderPrimary,
    borderRadius: 16,
    paddingHorizontal: 12,
    paddingVertical: 6,
    marginRight: 8,
    maxWidth: 120,
  },
  selectedUserText: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    color: Colors.text,
    marginRight: 6,
    flex: 1,
  },
  removeButton: {
    width: 20,
    height: 20,
    alignItems: 'center',
    justifyContent: 'center',
  },
  userList: {
    flex: 1,
  },
  userItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: Colors.border,
  },
  userItemSelected: {
    backgroundColor: Colors.surfaceVariant,
  },
  userAvatar: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: Colors.surfaceVariant,
    borderWidth: 1,
    borderColor: Colors.borderPrimary,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 12,
  },
  userAvatarText: {
    fontSize: 16,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.primary,
  },
  userInfo: {
    flex: 1,
  },
  userName: {
    fontSize: 14,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.text,
    marginBottom: 2,
  },
  userEmail: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
  },
  emptyState: {
    alignItems: 'center',
    paddingVertical: 40,
  },
  emptyText: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
    letterSpacing: 1,
  },
  createButton: {
    backgroundColor: Colors.primary,
    marginHorizontal: 20,
    marginVertical: 16,
    paddingVertical: 16,
    borderRadius: 8,
    alignItems: 'center',
  },
  createButtonDisabled: {
    backgroundColor: Colors.textMuted,
    opacity: 0.5,
  },
  createButtonText: {
    fontSize: 14,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.background,
    letterSpacing: 2,
  },
});