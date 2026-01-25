'use client';

import { useState, useEffect } from 'react';
import { api, Note } from '@/lib/api';

export default function NotesPage() {
  const [notes, setNotes] = useState<Note[]>([]);
  const [selectedNote, setSelectedNote] = useState<Note | null>(null);
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [isNew, setIsNew] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadNotes();
  }, []);

  async function loadNotes() {
    const res = await api.getNotes();
    if (res.success && res.data) {
      setNotes(res.data);
    }
    setLoading(false);
  }

  function handleNewNote() {
    setSelectedNote(null);
    setTitle('');
    setContent('');
    setIsNew(true);
  }

  function handleSelectNote(note: Note) {
    setSelectedNote(note);
    setTitle(note.title);
    setContent(note.content);
    setIsNew(false);
  }

  async function handleSave() {
    if (!title.trim()) return;

    if (isNew) {
      const res = await api.createNote(title, content);
      if (res.success && res.data) {
        setNotes(prev => [res.data!, ...prev]);
        setSelectedNote(res.data);
        setIsNew(false);
      }
    } else if (selectedNote) {
      const res = await api.updateNote(selectedNote.id, title, content);
      if (res.success && res.data) {
        setNotes(prev => prev.map(n => n.id === selectedNote.id ? res.data! : n));
        setSelectedNote(res.data);
      }
    }
  }

  async function handleDelete() {
    if (!selectedNote) return;
    if (!confirm('Delete this note?')) return;

    const res = await api.deleteNote(selectedNote.id);
    if (res.success) {
      setNotes(prev => prev.filter(n => n.id !== selectedNote.id));
      setSelectedNote(null);
      setTitle('');
      setContent('');
    }
  }

  async function handleTogglePin() {
    if (!selectedNote) return;
    const res = await api.togglePinNote(selectedNote.id);
    if (res.success) {
      loadNotes();
    }
  }

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex">
      {/* Notes List */}
      <div className="w-64 border-r border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
        <div className="p-4 border-b border-gray-200 dark:border-gray-700">
          <button
            onClick={handleNewNote}
            className="w-full py-2 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            New Note
          </button>
        </div>
        <div className="overflow-y-auto">
          {notes.map((note) => (
            <button
              key={note.id}
              onClick={() => handleSelectNote(note)}
              className={`w-full p-4 text-left border-b border-gray-100 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-700 ${
                selectedNote?.id === note.id ? 'bg-blue-50 dark:bg-blue-900/30' : ''
              }`}
            >
              <div className="flex items-center gap-2">
                {note.is_pinned && (
                  <svg className="w-4 h-4 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                  </svg>
                )}
                <p className="font-medium text-gray-900 dark:text-white truncate">{note.title || 'Untitled'}</p>
              </div>
              <p className="text-sm text-gray-500 dark:text-gray-400 truncate mt-1">{note.content}</p>
            </button>
          ))}
        </div>
      </div>

      {/* Note Editor */}
      <div className="flex-1 flex flex-col bg-white dark:bg-gray-800">
        {(selectedNote || isNew) ? (
          <>
            <div className="p-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Note title..."
                className="text-xl font-semibold bg-transparent border-none focus:outline-none text-gray-900 dark:text-white flex-1"
              />
              <div className="flex items-center gap-2">
                {selectedNote && (
                  <>
                    <button
                      onClick={handleTogglePin}
                      className={`p-2 rounded-lg ${selectedNote.is_pinned ? 'text-yellow-500' : 'text-gray-400 hover:text-yellow-500'}`}
                    >
                      <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                      </svg>
                    </button>
                    <button
                      onClick={handleDelete}
                      className="p-2 text-gray-400 hover:text-red-500 rounded-lg"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </>
                )}
                <button
                  onClick={handleSave}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                >
                  Save
                </button>
              </div>
            </div>
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Start writing..."
              className="flex-1 p-4 bg-transparent border-none focus:outline-none resize-none text-gray-700 dark:text-gray-300"
            />
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <svg className="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <p className="mt-4 text-gray-500 dark:text-gray-400">Select a note or create a new one</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
