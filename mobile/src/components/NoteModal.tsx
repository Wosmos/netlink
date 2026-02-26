import { useState, useEffect } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  Modal,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  Alert,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Colors } from '@/constants/Colors';
import { Note } from '@/types';
import { api } from '@/lib/api';

interface NoteModalProps {
  visible: boolean;
  note?: Note | null;
  onClose: () => void;
  onSave: (note: Note) => void;
}

export default function NoteModal({ visible, note, onClose, onSave }: NoteModalProps) {
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (note) {
      setTitle(note.title);
      setContent(note.content);
    } else {
      setTitle('');
      setContent('');
    }
  }, [note, visible]);

  const handleSave = async () => {
    if (!title.trim()) {
      Alert.alert('Error', 'Please enter a title for the note');
      return;
    }

    setIsLoading(true);
    try {
      let response;
      
      if (note) {
        // Update existing note
        response = await api.updateNote(note.id, title.trim(), content.trim());
      } else {
        // Create new note
        response = await api.createNote(title.trim(), content.trim());
      }

      if (response.success && response.data) {
        onSave(response.data);
      } else {
        Alert.alert('Error', response.error || 'Failed to save note');
      }
    } catch (error) {
      Alert.alert('Error', 'Failed to save note');
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    if ((title.trim() || content.trim()) && !note) {
      Alert.alert(
        'Discard Changes',
        'You have unsaved changes. Are you sure you want to discard them?',
        [
          { text: 'Keep Editing', style: 'cancel' },
          { text: 'Discard', style: 'destructive', onPress: onClose },
        ]
      );
    } else {
      onClose();
    }
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
          
          <Text style={styles.headerTitle}>
            {note ? 'EDIT NOTE' : 'NEW NOTE'}
          </Text>
          
          <TouchableOpacity
            style={[styles.saveButton, (!title.trim() || isLoading) && styles.saveButtonDisabled]}
            onPress={handleSave}
            disabled={!title.trim() || isLoading}
          >
            <Text style={[styles.saveButtonText, (!title.trim() || isLoading) && styles.saveButtonTextDisabled]}>
              {isLoading ? 'SAVING...' : 'SAVE'}
            </Text>
          </TouchableOpacity>
        </View>

        <ScrollView style={styles.content} keyboardShouldPersistTaps="handled">
          {/* Title Input */}
          <View style={styles.inputContainer}>
            <TextInput
              style={styles.titleInput}
              placeholder="NOTE TITLE"
              placeholderTextColor={Colors.textMuted}
              value={title}
              onChangeText={setTitle}
              maxLength={100}
              autoFocus={!note}
            />
          </View>

          {/* Content Input */}
          <View style={[styles.inputContainer, styles.contentContainer]}>
            <TextInput
              style={styles.contentInput}
              placeholder="Write your note here..."
              placeholderTextColor={Colors.textMuted}
              value={content}
              onChangeText={setContent}
              multiline
              textAlignVertical="top"
              maxLength={5000}
            />
          </View>

          {/* Character Count */}
          <View style={styles.footer}>
            <Text style={styles.characterCount}>
              {content.length}/5000 characters
            </Text>
            {note && (
              <Text style={styles.lastModified}>
                Last modified: {new Date(note.updated_at).toLocaleDateString()}
              </Text>
            )}
          </View>
        </ScrollView>
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
  saveButton: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 6,
    backgroundColor: Colors.primary,
  },
  saveButtonDisabled: {
    backgroundColor: Colors.textMuted,
    opacity: 0.5,
  },
  saveButtonText: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.background,
    letterSpacing: 1,
  },
  saveButtonTextDisabled: {
    color: Colors.textMuted,
  },
  content: {
    flex: 1,
    padding: 20,
  },
  inputContainer: {
    backgroundColor: Colors.surfaceVariant,
    borderWidth: 1,
    borderColor: Colors.border,
    borderRadius: 8,
    marginBottom: 16,
  },
  contentContainer: {
    flex: 1,
    minHeight: 200,
  },
  titleInput: {
    fontSize: 18,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.text,
    paddingHorizontal: 16,
    paddingVertical: 16,
  },
  contentInput: {
    fontSize: 14,
    fontFamily: 'SpaceMono',
    color: Colors.text,
    paddingHorizontal: 16,
    paddingVertical: 16,
    minHeight: 200,
    lineHeight: 20,
  },
  footer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingTop: 16,
  },
  characterCount: {
    fontSize: 10,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
  },
  lastModified: {
    fontSize: 10,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
  },
});