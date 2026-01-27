'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { api, Conversation, User } from '@/lib/api';
import { wsClient } from '@/lib/websocket';
import { useLanguage } from '@/context/LanguageContext';
import NotesPanel from './NotesPanel';
import TasksPanel from './TasksPanel';

interface SidebarProps {
  user: User;
  onLogout: () => void;
}

export default function Sidebar({ user, onLogout }: SidebarProps) {
  const { t, mode, setMode } = useLanguage();
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [onlineUsers, setOnlineUsers] = useState<number[]>([]);
  const [showNewChat, setShowNewChat] = useState(false);
  const [chatType, setChatType] = useState<'direct' | 'group'>('direct');
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [groupName, setGroupName] = useState('');
  const [selectedMembers, setSelectedMembers] = useState<User[]>([]);
  const [collapsed, setCollapsed] = useState(true);
  const [showNotesPanel, setShowNotesPanel] = useState(false);
  const [showTasksPanel, setShowTasksPanel] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const pathname = usePathname();
  const router = useRouter();

  // Derive conversation ID from pathname (no state needed)
  const currentConversationId = useMemo(() => {
    const match = pathname?.match(/^\/chat\/(\d+)$/);
    return match ? parseInt(match[1]) : undefined;
  }, [pathname]);

  const loadConversations = useCallback(async () => {
    const res = await api.getConversations();
    if (res.success && res.data) {
      setConversations(res.data);
    } else if (Array.isArray(res)) {
      setConversations(res as unknown as Conversation[]);
    }
  }, []);

  const loadOnlineUsers = useCallback(async () => {
    const res = await api.getOnlineUsers();
    if (res.success && res.data) setOnlineUsers(res.data.online_users || []);
  }, []);

  useEffect(() => {
    let mounted = true;
    
    const initLoad = async () => {
      if (mounted) {
        await loadConversations();
        await loadOnlineUsers();
      }
    };
    initLoad();
    
    const unsubMessage = wsClient.on('message', () => loadConversations());
    const unsubConv = wsClient.on('conversation', () => loadConversations());
    const unsubOnline = wsClient.on('online', (event) => {
      if (event.user_id) setOnlineUsers(prev => prev.includes(event.user_id!) ? prev : [...prev, event.user_id!]);
    });
    const unsubOffline = wsClient.on('offline', (event) => {
      if (event.user_id) setOnlineUsers(prev => prev.filter(id => id !== event.user_id));
    });
    
    // Reduced polling - only as backup since WebSocket handles real-time updates
    const interval = setInterval(() => {
      loadOnlineUsers();
    }, 15000); // Increased to 15 seconds
    
    return () => { mounted = false; unsubMessage(); unsubConv(); unsubOnline(); unsubOffline(); clearInterval(interval); };
  }, [loadConversations, loadOnlineUsers]);

  useEffect(() => {
    if (pathname?.startsWith('/chat')) {
      const refresh = async () => { await loadConversations(); };
      refresh();
    }
  }, [pathname, loadConversations]);

  async function handleSearch(query: string) {
    setSearchQuery(query);
    if (query.length >= 2) {
      const res = await api.searchUsers(query);
      if (res.success && res.data) setSearchResults(res.data.filter(u => u.id !== user.id));
    } else {
      setSearchResults([]);
    }
  }

  async function startDirectChat(targetUser: User) {
    const res = await api.createDirectChat({ user_id: targetUser.id });
    if (res.success && res.data) {
      closeModal();
      loadConversations();
      router.push(`/chat/${res.data.id}`);
    } else {
      alert(res.error || 'Failed to create conversation');
    }
  }

  async function createGroup() {
    if (!groupName.trim() || selectedMembers.length === 0) return;
    const res = await api.createGroup(groupName, selectedMembers.map(m => m.id));
    if (res.success && res.data) {
      closeModal();
      loadConversations();
      router.push(`/chat/${res.data.id}`);
    } else {
      alert(res.error || 'Failed to create group');
    }
  }

  function toggleMember(u: User) {
    if (selectedMembers.find(m => m.id === u.id)) {
      setSelectedMembers(prev => prev.filter(m => m.id !== u.id));
    } else {
      setSelectedMembers(prev => [...prev, u]);
    }
  }

  function closeModal() {
    setShowNewChat(false);
    setSearchQuery('');
    setSearchResults([]);
    setGroupName('');
    setSelectedMembers([]);
    setChatType('direct');
  }

  function formatName(name?: string, email?: string) {
    if (name) return name.length > 14 ? name.slice(0, 14) + '...' : name;
    if (email) {
      const local = email.split('@')[0];
      return local.length > 12 ? local.slice(0, 12) + '...' : local;
    }
    return 'Unknown';
  }

  function getConversationName(conv: Conversation) {
    if (conv.type === 'group') return conv.name || 'Group';
    const other = conv.members?.find(m => m.user_id !== user.id);
    return formatName(other?.user?.name, other?.user?.email);
  }

  function getOtherUserId(conv: Conversation) {
    if (conv.type === 'group') return null;
    return conv.members?.find(m => m.user_id !== user.id)?.user_id;
  }

  const isActive = (path: string) => pathname === path || (path === '/chat' && pathname?.match(/^\/chat\/\d+$/));

  function openNotesPanel() {
    setShowNotesPanel(true);
    setShowTasksPanel(false);
    setCollapsed(true);
  }

  function openTasksPanel() {
    setShowTasksPanel(true);
    setShowNotesPanel(false);
    setCollapsed(true);
  }


  return (
    <>
      {/* Mobile Toggle Button */}
      <button 
        onClick={() => setCollapsed(!collapsed)}
        className="lg:hidden fixed top-3 left-3 z-50 p-2 bg-[#0a0a0f] border border-cyan-500/30 text-cyan-400 rounded"
      >
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          {collapsed ? (
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
          ) : (
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          )}
        </svg>
      </button>

      {/* Overlay for mobile */}
      {!collapsed && (
        <div className="lg:hidden fixed inset-0 bg-black/50 z-30" onClick={() => setCollapsed(true)}></div>
      )}

      {/* Sidebar */}
      <div className={`
        fixed lg:relative inset-y-0 left-0 z-40
        w-72 sm:w-80 bg-[#0a0a0f] border-r border-cyan-900/30 flex flex-col
        transform transition-transform duration-300 ease-in-out
        ${collapsed ? '-translate-x-full lg:translate-x-0' : 'translate-x-0'}
      `}>
        {/* Header */}
        <div className="p-4 border-b border-cyan-900/30">
          <div className="flex items-center justify-between">
            <h1 className="text-xl font-mono font-bold text-cyan-400 tracking-widest uppercase">NetLink</h1>
            <button onClick={() => setShowNewChat(true)} className="p-2 text-cyan-500 hover:text-cyan-300 hover:bg-cyan-900/20 rounded" title="New Chat">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
            </button>
          </div>
        </div>

        {/* Navigation */}
        <div className="flex border-b border-cyan-900/30">
          <button 
            onClick={() => router.push('/chat')}
            className={`flex-1 py-3 text-center text-xs font-mono uppercase tracking-wider ${
              isActive('/chat') ? 'text-cyan-400 border-b-2 border-cyan-500 bg-cyan-950/20' : 'text-gray-500 hover:text-cyan-400'
            }`}>
            {t('CHATS', 'Chats')}
          </button>
          <button 
            onClick={openNotesPanel}
            className={`flex-1 py-3 text-center text-xs font-mono uppercase tracking-wider ${
              showNotesPanel ? 'text-cyan-400 border-b-2 border-cyan-500 bg-cyan-950/20' : 'text-gray-500 hover:text-cyan-400'
            }`}>
            {t('LOGS', 'Notes')}
          </button>
          <button 
            onClick={openTasksPanel}
            className={`flex-1 py-3 text-center text-xs font-mono uppercase tracking-wider ${
              showTasksPanel ? 'text-cyan-400 border-b-2 border-cyan-500 bg-cyan-950/20' : 'text-gray-500 hover:text-cyan-400'
            }`}>
            {t('QUEUE', 'Tasks')}
          </button>
        </div>

        {/* Conversations List */}
        <div className="flex-1 overflow-y-auto">
          {conversations.length === 0 ? (
            <div className="p-4 text-center">
              <p className="text-cyan-500/50 font-mono text-xs">{t('NO_CONNECTIONS', 'NO CONNECTIONS')}</p>
              <button onClick={() => setShowNewChat(true)} className="mt-2 text-cyan-400 hover:text-cyan-300 text-xs font-mono">
                {t('+ Initialize New Link', '+ Start New Chat')}
              </button>
            </div>
          ) : (
            conversations.map((conv) => {
              const otherId = getOtherUserId(conv);
              const isOnline = otherId ? onlineUsers.includes(otherId) : false;
              
              return (
                <Link 
                  key={conv.id} 
                  href={`/chat/${conv.id}`}
                  onClick={() => setCollapsed(true)}
                  onMouseEnter={() => api.prefetchMessages(conv.id)} // Prefetch on hover
                  className={`block p-3 sm:p-4 hover:bg-cyan-950/20 border-b border-cyan-900/20 transition-colors ${
                    pathname === `/chat/${conv.id}` ? 'bg-cyan-950/30 border-l-2 border-l-cyan-500' : ''
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <div className="relative shrink-0">
                      <div className={`w-10 h-10 flex items-center justify-center text-sm font-bold ${
                        conv.type === 'group' 
                          ? 'bg-orange-950/50 border border-orange-500/50 text-orange-400' 
                          : 'bg-cyan-950/50 border border-cyan-500/50 text-cyan-400'
                      }`} style={{ clipPath: 'polygon(20% 0, 100% 0, 100% 80%, 80% 100%, 0 100%, 0 20%)' }}>
                        {conv.type === 'group' ? (
                          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                          </svg>
                        ) : getConversationName(conv).charAt(0).toUpperCase()}
                      </div>
                      {conv.type === 'direct' && (
                        <div className={`absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-[#0a0a0f] ${isOnline ? 'bg-emerald-500' : 'bg-gray-600'}`}></div>
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-2">
                        <p className="font-mono text-sm text-cyan-100 truncate">{getConversationName(conv)}</p>
                        {conv.unread_count > 0 && (
                          <span className="shrink-0 bg-cyan-500 text-black text-[10px] px-1.5 py-0.5 font-mono font-bold">{conv.unread_count}</span>
                        )}
                      </div>
                      {conv.last_message && (
                        <p className="text-xs text-gray-500 truncate font-mono">{conv.last_message.content}</p>
                      )}
                    </div>
                  </div>
                </Link>
              );
            })
          )}
        </div>

        {/* User Info */}
        <div className="p-3 sm:p-4 border-t border-cyan-900/30 bg-[#050508]">
          <div className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2 min-w-0">
              <div className="w-8 h-8 shrink-0 bg-cyan-950/50 border border-cyan-500/30 flex items-center justify-center text-xs font-mono text-cyan-400">
                {user.email.charAt(0).toUpperCase()}
              </div>
              <span className="text-xs text-cyan-300/70 font-mono truncate">{formatName(user.name, user.email)}</span>
            </div>
            <div className="flex items-center gap-1 shrink-0">
              <button 
                onClick={() => setShowSettings(!showSettings)} 
                className="text-gray-500 hover:text-cyan-400 p-1 transition-colors" 
                title={t('SETTINGS', 'Settings')}
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </button>
              <button onClick={onLogout} className="shrink-0 text-gray-500 hover:text-red-500 p-1 transition-colors" title={t('LOGOUT', 'Logout')}>
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                </svg>
              </button>
            </div>
          </div>
          
          {/* Settings Dropdown */}
          {showSettings && (
            <>
              <div className="fixed inset-0 z-30" onClick={() => setShowSettings(false)}></div>
              <div className="absolute bottom-full left-3 right-3 mb-2 bg-[#0c0c14] border border-cyan-500/30 z-40 shadow-xl p-3">
                <div className="text-[10px] text-cyan-500/50 uppercase font-mono mb-2">{t('INTERFACE_MODE', 'Language Mode')}</div>
                <div className="flex gap-2">
                  <button
                    onClick={() => { setMode('techy'); setShowSettings(false); }}
                    className={`flex-1 py-2 px-3 text-xs font-mono uppercase border transition-all ${
                      mode === 'techy'
                        ? 'bg-cyan-950/50 border-cyan-500 text-cyan-400'
                        : 'border-gray-700 text-gray-500 hover:border-cyan-500/50'
                    }`}
                  >
                    Techy
                  </button>
                  <button
                    onClick={() => { setMode('normie'); setShowSettings(false); }}
                    className={`flex-1 py-2 px-3 text-xs font-mono uppercase border transition-all ${
                      mode === 'normie'
                        ? 'bg-orange-950/50 border-orange-500 text-orange-400'
                        : 'border-gray-700 text-gray-500 hover:border-orange-500/50'
                    }`}
                  >
                    Normie
                  </button>
                </div>
              </div>
            </>
          )}
        </div>

        {/* New Chat Modal */}
        {showNewChat && (
          <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4" onClick={closeModal}>
            <div className="bg-[#0a0a0f] border border-cyan-500/30 rounded p-4 sm:p-6 w-full max-w-sm max-h-[80vh] overflow-hidden flex flex-col" onClick={e => e.stopPropagation()}>
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-sm font-mono font-bold text-cyan-400 uppercase tracking-wider">{t('NEW_CONNECTION', 'New Connection')}</h2>
                <button onClick={closeModal} className="text-gray-500 hover:text-cyan-400">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              {/* Chat Type Toggle */}
              <div className="flex gap-2 mb-4">
                <button onClick={() => setChatType('direct')}
                  className={`flex-1 py-2 text-xs font-mono uppercase tracking-wider border ${
                    chatType === 'direct' 
                      ? 'bg-cyan-950/50 border-cyan-500 text-cyan-400' 
                      : 'border-gray-700 text-gray-500 hover:border-cyan-500/50'
                  }`}>
                  Direct
                </button>
                <button onClick={() => setChatType('group')}
                  className={`flex-1 py-2 text-xs font-mono uppercase tracking-wider border ${
                    chatType === 'group' 
                      ? 'bg-cyan-950/50 border-cyan-500 text-cyan-400' 
                      : 'border-gray-700 text-gray-500 hover:border-cyan-500/50'
                  }`}>
                  Group
                </button>
              </div>

              {chatType === 'group' && (
                <input 
                  type="text" 
                  placeholder="GROUP_NAME..." 
                  value={groupName} 
                  onChange={(e) => setGroupName(e.target.value)}
                  className="w-full px-3 py-2 bg-[#050508] border border-cyan-800/50 text-cyan-100 font-mono text-sm mb-2 focus:outline-none focus:border-cyan-500 placeholder-cyan-900" 
                />
              )}

              <input 
                type="text" 
                placeholder="SEARCH_USERS..." 
                value={searchQuery} 
                onChange={(e) => handleSearch(e.target.value)}
                className="w-full px-3 py-2 bg-[#050508] border border-cyan-800/50 text-cyan-100 font-mono text-sm mb-4 focus:outline-none focus:border-cyan-500 placeholder-cyan-900" 
              />

              {/* Selected Members for Group */}
              {chatType === 'group' && selectedMembers.length > 0 && (
                <div className="flex flex-wrap gap-2 mb-4">
                  {selectedMembers.map(m => (
                    <span key={m.id} className="bg-cyan-950/50 border border-cyan-500/30 text-cyan-300 px-2 py-1 text-xs font-mono flex items-center gap-1">
                      {formatName(m.name, m.email)}
                      <button onClick={() => toggleMember(m)} className="hover:text-red-400">×</button>
                    </span>
                  ))}
                </div>
              )}

              <div className="flex-1 overflow-y-auto max-h-60 space-y-1">
                {searchResults.length === 0 && searchQuery.length >= 2 && (
                  <p className="text-center text-cyan-500/50 py-4 font-mono text-xs">NO TARGETS FOUND</p>
                )}
                {searchResults.map((u) => (
                  <button 
                    key={u.id}
                    onClick={() => chatType === 'direct' ? startDirectChat(u) : toggleMember(u)}
                    className={`w-full p-3 flex items-center gap-3 hover:bg-cyan-950/30 border border-transparent ${
                      chatType === 'group' && selectedMembers.find(m => m.id === u.id) ? 'bg-cyan-950/20 border-cyan-500/30' : ''
                    }`}
                  >
                    <div className="w-9 h-9 bg-cyan-950/50 border border-cyan-500/30 flex items-center justify-center text-cyan-400 text-sm font-mono">
                      {(u.name || u.email).charAt(0).toUpperCase()}
                    </div>
                    <div className="text-left flex-1 min-w-0">
                      <p className="font-mono text-sm text-cyan-100 truncate">{formatName(u.name, u.email)}</p>
                      {u.name && <p className="text-xs text-gray-500 font-mono truncate">{u.email}</p>}
                    </div>
                    {chatType === 'group' && selectedMembers.find(m => m.id === u.id) && (
                      <svg className="w-4 h-4 text-cyan-400 shrink-0" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                      </svg>
                    )}
                  </button>
                ))}
              </div>

              {chatType === 'group' && (
                <button 
                  onClick={createGroup} 
                  disabled={!groupName.trim() || selectedMembers.length === 0}
                  className="mt-4 w-full py-2 bg-cyan-900/30 border border-cyan-500/50 text-cyan-400 font-mono text-xs uppercase tracking-wider hover:bg-cyan-500 hover:text-black transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
                >
                  Create Group ({selectedMembers.length})
                </button>
              )}
            </div>
          </div>
        )}
      </div>
      
      {/* Notes Panel */}
      {showNotesPanel && <NotesPanel onClose={() => setShowNotesPanel(false)} conversationId={currentConversationId} />}
      
      {/* Tasks Panel */}
      {showTasksPanel && <TasksPanel onClose={() => setShowTasksPanel(false)} conversationId={currentConversationId} />}
    </>
  );
}
