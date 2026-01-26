'use client';

import { useState, useEffect } from 'react';
import { api, Task } from '@/lib/api';
import { useLanguage } from '@/context/LanguageContext';

interface TasksPanelProps {
  onClose: () => void;
  conversationId?: number; // Optional - if provided, shows conversation-specific tasks
}

export default function TasksPanel({ onClose, conversationId }: TasksPanelProps) {
  const { t } = useLanguage();
  const [tasks, setTasks] = useState<Task[]>([]);
  const [newTask, setNewTask] = useState('');
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'active' | 'completed'>('all');

  async function loadTasks() {
    setLoading(true);
    const res = await api.getTasks(conversationId);
    if (res.success && res.data) {
      setTasks(res.data);
    }
    setLoading(false);
  }

  async function handleAddTask(e: React.FormEvent) {
    e.preventDefault();
    if (!newTask.trim()) return;
    const res = await api.createTask(newTask, conversationId);
    if (res.success) {
      setNewTask('');
      loadTasks();
    }
  }
  
  useEffect(() => {
    let mounted = true;
    
    async function fetchTasks() {
      setLoading(true);
      const res = await api.getTasks(conversationId);
      if (mounted && res.success && res.data) {
        setTasks(res.data);
      }
      if (mounted) {
        setLoading(false);
      }
    }
    
    fetchTasks();
    
    return () => {
      mounted = false;
    };
  }, [conversationId]); // Reload when conversation changes
  async function handleToggle(id: number) {
    await api.toggleTask(id);
    setTasks(prev => prev.map(t => t.id === id ? { ...t, completed: !t.completed } : t));
  }

  async function handleDelete(id: number) {
    await api.deleteTask(id);
    setTasks(prev => prev.filter(t => t.id !== id));
  }

  const filteredTasks = tasks.filter(task => {
    if (filter === 'active') return !task.completed;
    if (filter === 'completed') return task.completed;
    return true;
  });

  const activeCount = tasks.filter(t => !t.completed).length;

  if (loading) {
    return (
      <div className="fixed inset-y-0 right-0 w-full sm:w-96 bg-[#0a0a0f] border-l border-cyan-900/30 z-50 flex items-center justify-center">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 border-4 border-cyan-500/30 border-t-cyan-500 rounded-full animate-spin"></div>
          <span className="text-cyan-500 font-mono text-xs animate-pulse tracking-[0.2em]">
            {t('LOADING_TASKS...', 'Loading...')}
          </span>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-y-0 right-0 w-full sm:w-96 bg-[#0a0a0f] border-l border-cyan-900/30 z-50 flex flex-col shadow-2xl">
      {/* Header */}
      <div className="p-4 border-b border-cyan-900/30 bg-[#050508] flex items-center justify-between">
        <h2 className="text-lg font-mono font-bold text-cyan-400 uppercase tracking-widest">
          {t('TASK_QUEUE', 'Tasks')}
        </h2>
        <button
          onClick={onClose}
          className="p-2 text-gray-500 hover:text-cyan-400 hover:bg-cyan-900/20 rounded transition-colors"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      {/* Add Task Form */}
      <div className="p-4 border-b border-cyan-900/30 bg-[#050508]">
        <form onSubmit={handleAddTask} className="flex gap-2">
          <input
            type="text"
            value={newTask}
            onChange={(e) => setNewTask(e.target.value)}
            placeholder={t('NEW_TASK_ENTRY...', 'What needs to be done?')}
            className="flex-1 px-3 py-2 bg-[#0a0a0f] border border-cyan-800/50 text-cyan-100 font-mono text-xs focus:outline-none focus:border-cyan-500 placeholder-cyan-900"
          />
          <button
            type="submit"
            className="px-4 py-2 bg-cyan-900/30 border border-cyan-500/50 text-cyan-400 font-mono text-xs uppercase hover:bg-cyan-500 hover:text-black transition-all"
            style={{ clipPath: 'polygon(6px 0, 100% 0, 100% calc(100% - 6px), calc(100% - 6px) 100%, 0 100%, 0 6px)' }}
          >
            +
          </button>
        </form>
      </div>

      {/* Filter Tabs */}
      <div className="flex border-b border-cyan-900/30 bg-[#050508]">
        {(['all', 'active', 'completed'] as const).map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`flex-1 py-2 text-xs font-mono uppercase tracking-wider transition-colors ${
              filter === f
                ? 'text-cyan-400 border-b-2 border-cyan-500 bg-cyan-950/20'
                : 'text-gray-500 hover:text-cyan-400'
            }`}
          >
            {t(
              f === 'all' ? 'ALL' : f === 'active' ? 'ACTIVE' : 'COMPLETE',
              f.charAt(0).toUpperCase() + f.slice(1)
            )}
          </button>
        ))}
      </div>

      {/* Tasks List */}
      <div className="flex-1 overflow-y-auto">
        {filteredTasks.length === 0 ? (
          <div className="p-8 text-center">
            <div className="w-16 h-16 mx-auto mb-4 border border-cyan-500/30 flex items-center justify-center bg-cyan-950/20" 
              style={{ clipPath: 'polygon(20% 0, 100% 0, 100% 80%, 80% 100%, 0 100%, 0 20%)' }}>
              <svg className="w-8 h-8 text-cyan-500/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            </div>
            <p className="text-cyan-500/50 font-mono text-xs uppercase tracking-wider">
              {t(
                filter === 'all' ? 'NO_TASKS_IN_QUEUE' : `NO_${filter.toUpperCase()}_TASKS`,
                filter === 'all' ? 'No tasks yet' : `No ${filter} tasks`
              )}
            </p>
          </div>
        ) : (
          <ul className="divide-y divide-cyan-900/20">
            {filteredTasks.map((task) => (
              <li
                key={task.id}
                className="flex items-center gap-3 p-4 hover:bg-cyan-950/20 transition-colors group"
              >
                <button
                  onClick={() => handleToggle(task.id)}
                  className={`shrink-0 w-5 h-5 border-2 flex items-center justify-center transition-all ${
                    task.completed
                      ? 'bg-emerald-500 border-emerald-500 text-black'
                      : 'border-cyan-500/50 hover:border-cyan-500'
                  }`}
                  style={{ clipPath: 'polygon(15% 0, 100% 0, 100% 85%, 85% 100%, 0 100%, 0 15%)' }}
                >
                  {task.completed && (
                    <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                    </svg>
                  )}
                </button>
                <span
                  className={`flex-1 font-mono text-sm ${
                    task.completed
                      ? 'line-through text-gray-500'
                      : 'text-cyan-100'
                  }`}
                >
                  {task.text}
                </span>
                <button
                  onClick={() => handleDelete(task.id)}
                  className="shrink-0 opacity-0 group-hover:opacity-100 text-gray-500 hover:text-red-500 transition-all p-1"
                  title={t('DELETE_TASK', 'Delete')}
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Footer Stats */}
      {tasks.length > 0 && (
        <div className="p-3 border-t border-cyan-900/30 bg-[#050508]">
          <p className="text-xs font-mono text-gray-500 text-center">
            {t(
              `${activeCount} TASK${activeCount !== 1 ? 'S' : ''} PENDING`,
              `${activeCount} task${activeCount !== 1 ? 's' : ''} remaining`
            )}
          </p>
        </div>
      )}
    </div>
  );
}
