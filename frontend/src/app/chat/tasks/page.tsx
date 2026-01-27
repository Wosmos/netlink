'use client';

import { useState, useEffect } from 'react';
import { api, Task } from '@/lib/api';

export default function TasksPage() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [newTask, setNewTask] = useState('');
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'active' | 'completed'>('all');

  useEffect(() => {
    loadTasks();
  }, []);

  async function loadTasks() {
    const res = await api.getTasks();
    if (res.success && res.data) {
      setTasks(res.data);
    }
    setLoading(false);
  }

  async function handleAddTask(e: React.FormEvent) {
    e.preventDefault();
    if (!newTask.trim()) return;
    const res = await api.createTask(newTask);
    if (res.success) {
      setNewTask('');
      loadTasks();
    }
  }

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

  return (
    <div className="flex-1 bg-gray-50 dark:bg-gray-900 p-8">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-8">Tasks</h1>
        <form onSubmit={handleAddTask} className="mb-6">
          <div className="flex gap-3">
            <input type="text" value={newTask} onChange={(e) => setNewTask(e.target.value)}
              placeholder="What needs to be done?"
              className="flex-1 px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg dark:bg-gray-800 dark:text-white" />
            <button type="submit" className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700">Add</button>
          </div>
        </form>
        <div className="flex gap-2 mb-4">
          {(['all', 'active', 'completed'] as const).map((f) => (
            <button key={f} onClick={() => setFilter(f)}
              className={`px-4 py-2 rounded-lg text-sm font-medium ${filter === f ? 'bg-blue-600 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'}`}>
              {f.charAt(0).toUpperCase() + f.slice(1)}
            </button>
          ))}
        </div>
        {loading ? (
          <div className="flex justify-center py-8"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div></div>
        ) : (
          <div className="bg-white dark:bg-gray-800 rounded-xl shadow overflow-hidden">
            {filteredTasks.length === 0 ? (
              <div className="p-8 text-center text-gray-500">{filter === 'all' ? 'No tasks yet' : `No ${filter} tasks`}</div>
            ) : (
              <ul className="divide-y divide-gray-200 dark:divide-gray-700">
                {filteredTasks.map((task) => (
                  <li key={task.id} className="flex items-center gap-4 p-4 hover:bg-gray-50 dark:hover:bg-gray-700/50">
                    <button onClick={() => handleToggle(task.id)}
                      className={`w-6 h-6 rounded-full border-2 flex items-center justify-center ${task.completed ? 'bg-green-500 border-green-500 text-white' : 'border-gray-300'}`}>
                      {task.completed && <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg>}
                    </button>
                    <span className={`flex-1 ${task.completed ? 'line-through text-gray-400' : 'text-gray-900 dark:text-white'}`}>{task.text}</span>
                    <button onClick={() => handleDelete(task.id)} className="text-gray-400 hover:text-red-500">
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}
        {tasks.length > 0 && <div className="mt-4 text-sm text-gray-500">{activeCount} task{activeCount !== 1 ? 's' : ''} remaining</div>}
      </div>
    </div>
  );
}
