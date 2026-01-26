'use client';

import { useState, useEffect } from 'react';
import { api, Note } from '@/lib/api';
import { useLanguage } from '@/context/LanguageContext';

interface NotesPanelProps {
  onClose: () => void;
  conversationId?: number; // Optional - if provided, shows conversation-specific notes
}

export default function NotesPanel({ onClose, conversationId }: NotesPanelProps) {
  const { t } = useLanguage();
  const [notes, setNotes] = useState<Note[]>([]);
  const [selectedNote, setSelectedNote] = useState<Note | null>(null);
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [isNew, setIsNew] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;
    
    async function fetchNotes() {
      setLoading(true);
      const res = await api.getNotes(conversationId);
      if (mounted && res.success && res.data) {
        setNotes(res.data);
      }
      if (mounted) {
        setLoading(false);
      }
    }
    
    fetchNotes();
    
    return () => {
      mounted = false;
    };
  }, [conversationId]); // Reload when conversation changes

  async function loadNotes() {
    const res = await api.getNotes(conversationId);
    if (res.success && res.data) {
      setNotes(res.data);
    }
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
      const res = await api.createNote(title, content, conversationId);
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
    if (!confirm(t('TERMINATE DATA LOG?', 'Delete this note?'))) return;

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
      <div className="fixed inset-y-0 right-0 w-full sm:w-96 bg-[#0a0a0f] border-l border-cyan-900/30 z-50 flex items-center justify-center">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 border-4 border-cyan-500/30 border-t-cyan-500 rounded-full animate-spin"></div>
          <span className="text-cyan-500 font-mono text-xs animate-pulse tracking-[0.2em]">
            {t('LOADING_DATA...', 'Loading...')}
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
          {t('DATA_LOGS', 'Notes')}
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

      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Notes List */}
        <div className="border-b border-cyan-900/30 bg-[#050508]">
          <div className="p-3">
            <button
              onClick={handleNewNote}
              className="w-full py-2 px-4 bg-cyan-900/30 border border-cyan-500/50 text-cyan-400 font-mono text-xs uppercase tracking-wider hover:bg-cyan-500 hover:text-black transition-all"
              style={{ clipPath: 'polygon(8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%, 0 8px)' }}
            >
              + {t('NEW_LOG', 'New Note')}
            </button>
          </div>
          <div className="max-h-48 overflow-y-auto">
            {notes.map((note) => (
              <button
                key={note.id}
                onClick={() => handleSelectNote(note)}
                className={`w-full p-3 text-left border-b border-cyan-900/20 hover:bg-cyan-950/20 transition-colors ${
                  selectedNote?.id === note.id ? 'bg-cyan-950/30 border-l-2 border-l-cyan-500' : ''
                }`}
              >
                <div className="flex items-center gap-2 mb-1">
                  {note.is_pinned && (
                    <svg className="w-3 h-3 text-yellow-500 shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                    </svg>
                  )}
                  <p className="font-mono text-sm text-cyan-100 truncate flex-1">
                    {note.title || t('UNTITLED_LOG', 'Untitled')}
                  </p>
                </div>
                <p className="text-xs text-gray-500 truncate font-mono">{note.content}</p>
              </button>
            ))}
          </div>
        </div>

        {/* Note Editor */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {(selectedNote || isNew) ? (
            <>
              <div className="p-3 border-b border-cyan-900/30 flex items-center justify-between gap-2 bg-[#050508]">
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder={t('LOG_TITLE...', 'Note title...')}
                  className="flex-1 text-sm font-mono bg-transparent border-b border-cyan-800/50 focus:border-cyan-500 focus:outline-none text-cyan-100 py-1 placeholder-cyan-900"
                />
                <div className="flex items-center gap-1 shrink-0">
                  {selectedNote && (
                    <>
                      <button
                        onClick={handleTogglePin}
                        className={`p-1.5 rounded transition-colors ${
                          selectedNote.is_pinned 
                            ? 'text-yellow-500 hover:text-yellow-400' 
                            : 'text-gray-500 hover:text-yellow-500'
                        }`}
                        title={t('PIN_LOG', 'Pin note')}
                      >
                        <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                        </svg>
                      </button>
                      <button
                        onClick={handleDelete}
                        className="p-1.5 text-gray-500 hover:text-red-500 rounded transition-colors"
                        title={t('DELETE_LOG', 'Delete')}
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>
                    </>
                  )}
                  <button
                    onClick={handleSave}
                    disabled={!title.trim()}
                    className="px-3 py-1.5 bg-cyan-900/30 border border-cyan-500/50 text-cyan-400 font-mono text-xs uppercase hover:bg-cyan-500 hover:text-black transition-all disabled:opacity-30 disabled:cursor-not-allowed"
                  >
                    {t('SAVE', 'Save')}
                  </button>
                </div>
              </div>
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder={t('ENTER_DATA...', 'Start writing...')}
                className="flex-1 p-4 bg-[#050508] border-none focus:outline-none resize-none text-cyan-100 font-mono text-sm placeholder-cyan-900/50"
              />
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center p-8">
              <div className="text-center">
                <div className="w-16 h-16 mx-auto mb-4 border border-cyan-500/30 flex items-center justify-center bg-cyan-950/20" 
                  style={{ clipPath: 'polygon(20% 0, 100% 0, 100% 80%, 80% 100%, 0 100%, 0 20%)' }}>
                  <svg className="w-8 h-8 text-cyan-500/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <p className="text-cyan-500/50 font-mono text-xs uppercase tracking-wider">
                  {t('SELECT_OR_CREATE_LOG', 'Select or create a note')}
                </p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
