import { useEffect, useState } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  Alert,
  RefreshControl,
  TextInput,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Colors } from '@/constants/Colors';
import { Task } from '@/types';
import { api } from '@/lib/api';

export default function TasksScreen() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [newTaskText, setNewTaskText] = useState('');
  const [showCompleted, setShowCompleted] = useState(true);

  useEffect(() => {
    loadTasks();
  }, []);

  const loadTasks = async () => {
    setIsLoading(true);
    try {
      const response = await api.getTasks();
      if (response.success && response.data) {
        setTasks(response.data);
      }
    } catch (error) {
      console.error('Failed to load tasks:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const onRefresh = async () => {
    setRefreshing(true);
    await loadTasks();
    setRefreshing(false);
  };

  const handleCreateTask = async () => {
    if (!newTaskText.trim()) return;

    try {
      const response = await api.createTask(newTaskText.trim());
      if (response.success && response.data) {
        setTasks(prev => [response.data!, ...prev]);
        setNewTaskText('');
      } else {
        Alert.alert('Error', 'Failed to create task');
      }
    } catch (error) {
      Alert.alert('Error', 'Failed to create task');
    }
  };

  const handleToggleTask = async (task: Task) => {
    try {
      const response = await api.toggleTask(task.id);
      if (response.success) {
        setTasks(prev => prev.map(t => 
          t.id === task.id ? { ...t, completed: !t.completed } : t
        ));
      } else {
        Alert.alert('Error', 'Failed to update task');
      }
    } catch (error) {
      Alert.alert('Error', 'Failed to update task');
    }
  };

  const handleDeleteTask = (task: Task) => {
    Alert.alert(
      'Delete Task',
      'Are you sure you want to delete this task?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: async () => {
            try {
              const response = await api.deleteTask(task.id);
              if (response.success) {
                setTasks(prev => prev.filter(t => t.id !== task.id));
              } else {
                Alert.alert('Error', 'Failed to delete task');
              }
            } catch (error) {
              Alert.alert('Error', 'Failed to delete task');
            }
          },
        },
      ]
    );
  };

  const formatDate = (dateString: string) => {
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

  const renderTask = ({ item }: { item: Task }) => (
    <TouchableOpacity
      style={[styles.taskItem, item.completed && styles.completedTask]}
      onPress={() => handleToggleTask(item)}
      onLongPress={() => handleDeleteTask(item)}
    >
      <TouchableOpacity
        style={styles.checkbox}
        onPress={() => handleToggleTask(item)}
      >
        <Ionicons
          name={item.completed ? "checkbox" : "square-outline"}
          size={24}
          color={item.completed ? Colors.success : Colors.primary}
        />
      </TouchableOpacity>
      
      <View style={styles.taskContent}>
        <Text style={[
          styles.taskText,
          item.completed && styles.completedTaskText
        ]}>
          {item.text}
        </Text>
        <Text style={styles.taskDate}>
          {formatDate(item.created_at)}
        </Text>
      </View>
      
      <TouchableOpacity
        style={styles.deleteButton}
        onPress={() => handleDeleteTask(item)}
      >
        <Ionicons name="trash-outline" size={18} color={Colors.error} />
      </TouchableOpacity>
    </TouchableOpacity>
  );

  const renderEmptyState = () => (
    <View style={styles.emptyState}>
      <Ionicons name="checkbox-outline" size={64} color={Colors.textMuted} />
      <Text style={styles.emptyTitle}>NO TASKS FOUND</Text>
      <Text style={styles.emptySubtitle}>
        Create your first task to start organizing your work
      </Text>
    </View>
  );

  // Filter and sort tasks
  const filteredTasks = showCompleted 
    ? tasks 
    : tasks.filter(task => !task.completed);
  
  const sortedTasks = [...filteredTasks].sort((a, b) => {
    // Incomplete tasks first, then by creation date
    if (a.completed !== b.completed) {
      return a.completed ? 1 : -1;
    }
    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
  });

  const completedCount = tasks.filter(t => t.completed).length;
  const totalCount = tasks.length;

  return (
    <View style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>TASK QUEUE</Text>
        <TouchableOpacity
          style={styles.filterButton}
          onPress={() => setShowCompleted(!showCompleted)}
        >
          <Ionicons 
            name={showCompleted ? "eye" : "eye-off"} 
            size={20} 
            color={Colors.primary} 
          />
        </TouchableOpacity>
      </View>

      {/* Stats */}
      {totalCount > 0 && (
        <View style={styles.stats}>
          <Text style={styles.statsText}>
            {completedCount}/{totalCount} completed
          </Text>
          <View style={styles.progressBar}>
            <View 
              style={[
                styles.progressFill, 
                { width: `${totalCount > 0 ? (completedCount / totalCount) * 100 : 0}%` }
              ]} 
            />
          </View>
        </View>
      )}

      {/* New Task Input */}
      <View style={styles.inputContainer}>
        <TextInput
          style={styles.input}
          placeholder="ADD NEW TASK..."
          placeholderTextColor={Colors.textMuted}
          value={newTaskText}
          onChangeText={setNewTaskText}
          onSubmitEditing={handleCreateTask}
          returnKeyType="done"
          maxLength={200}
        />
        <TouchableOpacity
          style={[styles.addButton, !newTaskText.trim() && styles.addButtonDisabled]}
          onPress={handleCreateTask}
          disabled={!newTaskText.trim()}
        >
          <Ionicons name="add" size={24} color={Colors.background} />
        </TouchableOpacity>
      </View>

      {/* Tasks List */}
      <FlatList
        data={sortedTasks}
        renderItem={renderTask}
        keyExtractor={(item) => item.id.toString()}
        style={styles.list}
        contentContainerStyle={sortedTasks.length === 0 ? styles.emptyContainer : styles.listContent}
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
  filterButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: Colors.surfaceVariant,
    borderWidth: 1,
    borderColor: Colors.borderPrimary,
    alignItems: 'center',
    justifyContent: 'center',
  },
  stats: {
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: Colors.border,
  },
  statsText: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
    marginBottom: 8,
  },
  progressBar: {
    height: 4,
    backgroundColor: Colors.surfaceVariant,
    borderRadius: 2,
    overflow: 'hidden',
  },
  progressFill: {
    height: '100%',
    backgroundColor: Colors.success,
    borderRadius: 2,
  },
  inputContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: Colors.border,
  },
  input: {
    flex: 1,
    backgroundColor: Colors.surfaceVariant,
    borderWidth: 1,
    borderColor: Colors.border,
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    fontSize: 14,
    fontFamily: 'SpaceMono',
    color: Colors.text,
    marginRight: 12,
  },
  addButton: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: Colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
  },
  addButtonDisabled: {
    backgroundColor: Colors.textMuted,
    opacity: 0.5,
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
  taskItem: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: Colors.surfaceVariant,
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: Colors.border,
  },
  completedTask: {
    opacity: 0.6,
  },
  checkbox: {
    marginRight: 12,
  },
  taskContent: {
    flex: 1,
  },
  taskText: {
    fontSize: 14,
    fontFamily: 'SpaceMono',
    color: Colors.text,
    lineHeight: 20,
    marginBottom: 4,
  },
  completedTaskText: {
    textDecorationLine: 'line-through',
    color: Colors.textMuted,
  },
  taskDate: {
    fontSize: 10,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
  },
  deleteButton: {
    width: 32,
    height: 32,
    alignItems: 'center',
    justifyContent: 'center',
    marginLeft: 8,
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
  },
});