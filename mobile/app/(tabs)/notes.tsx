import { useEffect, useState } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  Alert,
  RefreshControl,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Colors } from '@/constants/Colors';
import { Note } from '@/types';
import { api } from '@/lib/api';
import NoteModal from '@/components/NoteModal';

export default function NotesScreen() {
  const [notes, setNotes] = useState<Note[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [showNoteModal, setShowNoteModal] = useState(false);
  const [editingNote, setEditingNote] = useState<Note | null>(null);

  useEffect(() => {
    loadNotes();
  }, []);

  const loadNotes = async () => {
    setIsLoading(true);
    try {
      const response = await api.getNotes();
      if (response.success && response.data) {
        setNotes(response.data);
      }
    } catch (error) {
      console.error('Failed to load notes:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const onRefresh = async () => {
    setRefreshing(true);
    await loadNotes();
    setRefreshing(false);
  };

  const handleCreateNote = () => {
    setEditingNote(null);
    setShowNoteModal(true);
  };

  const handleEditNote = (note: Note) => {
    setEditingNote(note);
    setShowNoteModal(true);
  };

  const handleDeleteNote = (note: Note) => {
    Alert.alert(
      'Delete Note',
      'Are you sure you want to delete this note?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: async () => {
            try {
              const response = await api.deleteNote(note.id);
              if (response.success) {
                setNotes(prev => prev.filter(n => n.id !== note.id));
              } else {
                Alert.alert('Error', 'Failed to delete note');
              }
            } catch (error) {
              Alert.alert('Error', 'Failed to delete note');
            }
          },
        },
      ]
    );
  };

  const handleTogglePin = async (note: Note) => {
    try {
      const response = await api.togglePinNote(note.id);
      if (response.success) {
        setNotes(prev => prev.map(n => 
          n.id === note.id ? { ...n, is_pinned: !n.is_pinned } : n
        ));
      } else {
        Alert.alert('Error', 'Failed to update note');
      }
    } catch (error) {
      Alert.alert('Error', 'Failed to update note');
    }
  };

  const handleNoteSaved = (savedNote: Note) => {
    if (editingNote) {
      // Update existing note
      setNotes(prev => prev.map(n => n.id === savedNote.id ? savedNote : n));
    } else {
      // Add new note
      setNotes(prev => [savedNote, ...prev]);
    }
    setShowNoteModal(false);
    setEditingNote(null);
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString([], { 
      month: 'short', 
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const renderNote = ({ item }: { item: Note }) => (
    <TouchableOpacity
      style={[styles.noteItem, item.is_pinned && styles.pinnedNote]}
      onPress={() => handleEditNote(item)}
      onLongPress={() => {
        Alert.alert(
          'Note Options',
          item.title,
          [
            { text: 'Cancel', style: 'cancel' },
            {
              text: item.is_pinned ? 'Unpin' : 'Pin',
              onPress: () => handleTogglePin(item),
            },
            {
              text: 'Delete',
              style: 'destructive',
              onPress: () => handleDeleteNote(item),
            },
          ]
        );
      }}
    >
      <View style={styles.noteHeader}>
        <Text style={styles.noteTitle} numberOfLines={1}>
          {item.title}
        </Text>
        <View style={styles.noteActions}>
          {item.is_pinned && (
            <Ionicons name="pin" size={14} color={Colors.secondary} />
          )}
          <Text style={styles.noteDate}>
            {formatDate(item.updated_at)}
          </Text>
        </View>
      </View>
      <Text style={styles.noteContent} numberOfLines={3}>
        {item.content}
      </Text>
    </TouchableOpacity>
  );

  const renderEmptyState = () => (
    <View style={styles.emptyState}>
      <Ionicons name="document-text-outline" size={64} color={Colors.textMuted} />
      <Text style={styles.emptyTitle}>NO NOTES FOUND</Text>
      <Text style={styles.emptySubtitle}>
        Create your first note to keep track of important information
      </Text>
      <TouchableOpacity
        style={styles.emptyButton}
        onPress={handleCreateNote}
      >
        <Text style={styles.emptyButtonText}>CREATE NOTE</Text>
      </TouchableOpacity>
    </View>
  );

  // Sort notes: pinned first, then by updated date
  const sortedNotes = [...notes].sort((a, b) => {
    if (a.is_pinned && !b.is_pinned) return -1;
    if (!a.is_pinned && b.is_pinned) return 1;
    return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
  });

  return (
    <View style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>SECURE NOTES</Text>
        <TouchableOpacity
          style={styles.addButton}
          onPress={handleCreateNote}
        >
          <Ionicons name="add" size={24} color={Colors.primary} />
        </TouchableOpacity>
      </View>

      {/* Notes List */}
      <FlatList
        data={sortedNotes}
        renderItem={renderNote}
        keyExtractor={(item) => item.id.toString()}
        style={styles.list}
        contentContainerStyle={sortedNotes.length === 0 ? styles.emptyContainer : styles.listContent}
        ListEmptyComponent={renderEmptyState}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            tintColor={Colors.primary}
            colors={[Colors.primary]}
          />
        }
        showsVerticalScrollIndicator={false}
      />

      {/* Note Modal */}
      <NoteModal
        visible={showNoteModal}
        note={editingNote}
        onClose={() => {
          setShowNoteModal(false);
          setEditingNote(null);
        }}
        onSave={handleNoteSaved}
      />
    </View>
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
  headerTitle: {
    fontSize: 16,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.primary,
    letterSpacing: 2,
  },
  addButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: Colors.surfaceVariant,
    borderWidth: 1,
    borderColor: Colors.borderPrimary,
    alignItems: 'center',
    justifyContent: 'center',
  },
  list: {
    flex: 1,
  },
  listContent: {
    padding: 16,
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
  },
  noteItem: {
    backgroundColor: Colors.surfaceVariant,
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: Colors.border,
  },
  pinnedNote: {
    borderColor: Colors.secondary,
    backgroundColor: Colors.surfaceVariant + 'CC',
  },
  noteHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 8,
  },
  noteTitle: {
    fontSize: 16,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.text,
    flex: 1,
    marginRight: 8,
  },
  noteActions: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  noteDate: {
    fontSize: 10,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
  },
  noteContent: {
    fontSize: 14,
    fontFamily: 'SpaceMono',
    color: Colors.textSecondary,
    lineHeight: 20,
  },
  emptyState: {
    alignItems: 'center',
    paddingHorizontal: 40,
  },
  emptyTitle: {
    fontSize: 16,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.textMuted,
    letterSpacing: 2,
    marginTop: 24,
    marginBottom: 8,
  },
  emptySubtitle: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
    textAlign: 'center',
    lineHeight: 18,
    marginBottom: 32,
  },
  emptyButton: {
    backgroundColor: Colors.surfaceVariant,
    borderWidth: 1,
    borderColor: Colors.borderPrimary,
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },
  emptyButtonText: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.primary,
    letterSpacing: 2,
  },
});