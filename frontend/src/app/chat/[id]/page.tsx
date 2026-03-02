'use client';

import { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { api, Message, Conversation, ReactionSummary } from '@/lib/api';
import { wsClient } from '@/lib/websocket';
import { useAuth } from '@/context/AuthContext';
import { cache, CACHE_KEYS } from '@/lib/cache';
import dynamic from 'next/dynamic';

// Lazy load voice components
const VoiceRecorder = dynamic(() => import('@/components/VoiceRecorder'), { ssr: false });
const VoicePlayer = dynamic(() => import('@/components/VoicePlayer'), { ssr: false });

export default function ConversationPage() {
  const { id } = useParams();
  const { user } = useAuth();
  const router = useRouter();
  
  const [conversation, setConversation] = useState<Conversation | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [newMessage, setNewMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const [onlineUsers, setOnlineUsers] = useState<number[]>([]);
  const [typing, setTyping] = useState<number[]>([]);
  const [showMenu, setShowMenu] = useState(false);
  const [showJumpMenu, setShowJumpMenu] = useState(false);
  const [showScrollDownBtn, setShowScrollDownBtn] = useState(false);
  const [copiedId, setCopiedId] = useState<number | null>(null);
  const [editingMsg, setEditingMsg] = useState<Message | null>(null);
  const [editContent, setEditContent] = useState('');
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; msg: Message } | null>(null);
  const [showVoiceRecorder, setShowVoiceRecorder] = useState(false);
  
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const lastTypingSent = useRef<number>(0);
  const wasAtBottomRef = useRef<boolean>(true);
  const editInputRef = useRef<HTMLInputElement>(null);

  const convId = parseInt(id as string);

  const scrollToBottom = useCallback((smooth = true) => {
    messagesEndRef.current?.scrollIntoView({ behavior: smooth ? 'smooth' : 'auto' });
  }, []);

  const formatName = useCallback((name?: string, email?: string) => {
    if (name) return name.length > 15 ? name.slice(0, 15) + '...' : name;
    if (email) {
      const local = email.split('@')[0];
      return local.length > 12 ? local.slice(0, 12) + '...' : local;
    }
    return 'Unknown';
  }, []);

  const getOtherUser = useCallback(() => {
    if (!conversation || conversation.type === 'group') return null;
    return conversation.members?.find(m => m.user_id !== user?.id)?.user;
  }, [conversation, user?.id]);

  const isOnline = useMemo(() => {
    const other = getOtherUser();
    return other ? onlineUsers.includes(other.id) : false;
  }, [getOtherUser, onlineUsers]);

  useEffect(() => {
    if (!convId || isNaN(convId)) return;
    let mounted = true;
    
    async function fetchData() {
      // Don't show loading if we have cached data
      const cachedMessages = cache.get<Message[]>(CACHE_KEYS.messages(convId));
      if (!cachedMessages) {
        setLoading(true);
      }
      
      const [convRes, msgRes, onlineRes] = await Promise.all([
        api.getConversations(),
        api.getMessages(convId),
        api.getOnlineUsers(),
      ]);
      
      if (!mounted) return;
      
      if (convRes.success && convRes.data) {
        const conv = convRes.data.find(c => c.id === convId);
        setConversation(conv || null);
      }
      if (msgRes.success && msgRes.data) setMessages(msgRes.data);
      if (onlineRes.success && onlineRes.data) setOnlineUsers(onlineRes.data.online_users || []);
      setLoading(false);
      api.markAsRead(convId);
      setTimeout(() => scrollToBottom(false), 10); // Reduced from 100ms
    }
    
    fetchData();
    
    // Reduced polling interval - only for online status
    const interval = setInterval(async () => {
      if (document.hidden) return;
      const onlineRes = await api.getOnlineUsers();
      if (mounted && onlineRes.success && onlineRes.data) {
        setOnlineUsers(onlineRes.data.online_users || []);
      }
    }, 15000); // 15 seconds

    return () => { mounted = false; clearInterval(interval); };
  }, [convId, scrollToBottom]);

  useEffect(() => {
    const unsubMessage = wsClient.on('message', (event, data) => {
      if (event.conversation_id === convId && data) {
        setMessages(prev => {
          // If this is our own message from WebSocket, check if we already have it
          if (data.sender_id === user?.id) {
            // Check if we already have this exact message (by ID or by recent content match)
            const alreadyExists = prev.some(m => m.id === data.id);
            if (alreadyExists) {
              return prev; // Skip duplicate
            }
            // Check if we have a temp message with same content (within last 2 seconds)
            const tempMsgIndex = prev.findIndex(m => 
              m.id < 0 && // Temp messages have negative IDs
              m.content === data.content &&
              Math.abs(new Date(m.created_at).getTime() - new Date(data.created_at).getTime()) < 2000
            );
            if (tempMsgIndex !== -1) {
              // Replace temp message with real one
              const newMessages = [...prev];
              newMessages[tempMsgIndex] = data;
              return newMessages;
            }
          }
          
          // Check if message already exists (by ID)
          if (prev.some(m => m.id === data.id)) {
            return prev;
          }
          
          // New message from someone else
          return [...prev, data];
        });
        setTyping(prev => prev.filter(uid => uid !== data.sender_id));
        if (data.sender_id !== user?.id) {
          api.markAsRead(convId);
        }
      }
    });
    
    const unsubMessageEdit = wsClient.on('message_edit', (event, data) => {
      if (event.conversation_id === convId && data) {
        setMessages(prev => prev.map(m => m.id === data.id ? data : m));
      }
    });
    
    const unsubMessageDelete = wsClient.on('message_delete', (event, data) => {
      if (event.conversation_id === convId && data) {
        const deleteData = data as unknown as { id: number; deleted_at: string; content: string };
        setMessages(prev => prev.map(m => 
          m.id === deleteData.id 
            ? { ...m, content: 'Message deleted', deleted_at: deleteData.deleted_at } 
            : m
        ));
      }
    });
    
    const unsubTyping = wsClient.on('typing', (event) => {
      if (event.conversation_id === convId && event.user_id && event.user_id !== user?.id) {
        const uid = event.user_id;
        setTyping(prev => prev.includes(uid) ? prev : [...prev, uid]);
        setTimeout(() => setTyping(prev => prev.filter(i => i !== uid)), 3000);
      }
    });
    const unsubOnline = wsClient.on('online', (event) => {
      if (event.user_id) setOnlineUsers(prev => prev.includes(event.user_id!) ? prev : [...prev, event.user_id!]);
    });
    const unsubOffline = wsClient.on('offline', (event) => {
      if (event.user_id) setOnlineUsers(prev => prev.filter(i => i !== event.user_id));
    });
    
    const unsubReaction = wsClient.on('reaction', (event, data) => {
      if (event.conversation_id === convId && data) {
        const reactionData = data as unknown as { message_id: number; reactions: ReactionSummary[] };
        setMessages(prev => prev.map(m => 
          m.id === reactionData.message_id 
            ? { ...m, reactions: reactionData.reactions } 
            : m
        ));
      }
    });
    
    return () => { 
      unsubMessage(); 
      unsubMessageEdit(); 
      unsubMessageDelete(); 
      unsubTyping(); 
      unsubOnline(); 
      unsubOffline();
      unsubReaction();
    };
  }, [convId, user?.id]);

  const handleScroll = () => {
    if (!scrollContainerRef.current) return;
    const { scrollTop, scrollHeight, clientHeight } = scrollContainerRef.current;
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 100;
    wasAtBottomRef.current = isAtBottom;
    setShowScrollDownBtn(!isAtBottom);
  };

  const scrollToDate = (dateKey: string) => {
    document.getElementById(`date-${dateKey}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
    setShowJumpMenu(false);
  };

  useEffect(() => {
    if (messages.length === 0) return;
    const lastMsg = messages[messages.length - 1];
    if (lastMsg.sender_id === user?.id || wasAtBottomRef.current) {
      setTimeout(() => scrollToBottom(true), 50);
    }
  }, [messages, scrollToBottom, user?.id]);

  async function handleSend(e: React.FormEvent) {
    e.preventDefault();
    if (!newMessage.trim()) return;
    const content = newMessage;
    setNewMessage('');
    
    // Optimistic update - show message immediately with negative temp ID
    const tempId = -Date.now(); // Negative ID to distinguish from real messages
    const tempMsg: Message = {
      id: tempId,
      conversation_id: convId,
      sender_id: user!.id,
      sender: user!,
      type: 'text',
      content: content,
      read_by: [user!.id],
      created_at: new Date().toISOString(),
    };
    setMessages(prev => [...prev, tempMsg]);
    setTimeout(() => scrollToBottom(true), 10);
    
    // Send to server
    const res = await api.sendMessage(convId, content);
    if (res.success && res.data) {
      // Replace temp message with real one from server
      setMessages(prev => prev.map(m => m.id === tempId ? res.data! : m));
    } else {
      // Remove temp message on error
      setMessages(prev => prev.filter(m => m.id !== tempId));
      alert('Failed to send message');
    }
  }

  async function handleVoiceSend(audioBlob: Blob, duration: number, waveform: number[]) {
    // Close modal immediately - don't make user wait for upload
    setShowVoiceRecorder(false);

    // Optimistic update - show voice message placeholder immediately
    const tempId = -Date.now();
    const tempMsg: Message = {
      id: tempId,
      conversation_id: convId,
      sender_id: user!.id,
      sender: user!,
      type: 'voice',
      content: 'Voice message',
      voice_duration: duration,
      voice_waveform: waveform,
      read_by: [user!.id],
      created_at: new Date().toISOString(),
    };
    setMessages(prev => [...prev, tempMsg]);
    setTimeout(() => scrollToBottom(true), 10);

    try {
      const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

      // Determine correct filename from blob MIME type
      const ext = audioBlob.type.includes('mp4') ? '.m4a' : audioBlob.type.includes('mpeg') ? '.mp3' : '.webm';

      // Upload voice file
      const formData = new FormData();
      formData.append('audio', audioBlob, `voice${ext}`);
      formData.append('duration', duration.toString());
      formData.append('waveform', JSON.stringify(waveform));

      const uploadRes = await fetch(`${API_URL}/api/voice/upload`, {
        method: 'POST',
        credentials: 'include',
        body: formData,
      });

      if (!uploadRes.ok) {
        throw new Error(`Upload failed: ${uploadRes.status}`);
      }

      const uploadData = await uploadRes.json();

      if (!uploadData.success) {
        throw new Error('Upload unsuccessful');
      }

      // Send message with voice data
      const res = await api.sendMessage(convId, 'Voice message', 'voice', undefined, {
        voice_file_path: uploadData.data.file_path,
        voice_duration: uploadData.data.duration,
        voice_waveform: uploadData.data.waveform,
        voice_file_size: uploadData.data.file_size,
      });

      if (res.success && res.data) {
        // Replace temp message with real one from server
        setMessages(prev => prev.map(m => m.id === tempId ? res.data! : m));
      } else {
        throw new Error('Failed to send voice message');
      }
    } catch (error) {
      console.error('Error sending voice message:', error);
      // Remove optimistic message on failure
      setMessages(prev => prev.filter(m => m.id !== tempId));
      alert(`Failed to send voice message: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  function handleInputChange(e: React.ChangeEvent<HTMLInputElement>) {
    setNewMessage(e.target.value);
    const now = Date.now();
    if (now - lastTypingSent.current > 1000) {
      lastTypingSent.current = now;
      api.sendTyping(convId);
    }
  }

  async function handleCopy(msg: Message) {
    await navigator.clipboard.writeText(msg.content);
    setCopiedId(msg.id);
    setContextMenu(null);
    setTimeout(() => setCopiedId(null), 2000);
  }

  function startEdit(msg: Message) {
    setEditingMsg(msg);
    setEditContent(msg.content);
    setContextMenu(null);
    setTimeout(() => editInputRef.current?.focus(), 50);
  }

  async function handleEditSave() {
    if (!editingMsg || !editContent.trim()) return;
    const res = await api.editMessage(convId, editingMsg.id, editContent);
    if (res.success && res.data) {
      // WebSocket will handle the update, but update locally for instant feedback
      setMessages(prev => prev.map(m => m.id === editingMsg.id ? res.data! : m));
    }
    setEditingMsg(null);
    setEditContent('');
  }

  async function handleDelete(msg: Message) {
    if (!confirm('Delete this message?')) return;
    const res = await api.deleteMessage(msg.id);
    if (res.success) {
      // WebSocket will handle the update, but update locally for instant feedback
      setMessages(prev => prev.map(m => m.id === msg.id ? { ...m, content: 'Message deleted', deleted_at: new Date().toISOString() } : m));
    }
    setContextMenu(null);
  }

  async function handleQuickReact(msg: Message, emoji: string) {
    const res = await api.reactToMessage(msg.id, emoji);
    if (res.success && res.data) {
      const { reactions } = res.data;
      setMessages(prev => prev.map(m =>
        m.id === msg.id ? { ...m, reactions } : m
      ));
    }
  }

  function handleContextMenu(e: React.MouseEvent, msg: Message) {
    e.preventDefault();
    setContextMenu({ x: e.clientX, y: e.clientY, msg });
  }

  function getConversationName() {
    if (!conversation) return 'TARGET_OFFLINE';
    if (conversation.type === 'group') return conversation.name || 'GROUP_NET';
    const other = getOtherUser();
    return formatName(other?.name, other?.email);
  }

  function formatTime(dateStr: string) {
    // Server sends UTC timestamps with 'Z' suffix - browser converts to local time
    return new Date(dateStr).toLocaleString([], { 
      hour: '2-digit', 
      minute: '2-digit',
      hour12: true 
    });
  }

  const groupedMessages = useMemo(() => {
    const groups: { [key: string]: Message[] } = {};
    messages.forEach(msg => {
      // Server sends UTC timestamps - browser converts to local time
      const dateObj = new Date(msg.created_at);
      const today = new Date();
      const yesterday = new Date(); yesterday.setDate(yesterday.getDate() - 1);
      let dateKey: string;
      if (dateObj.toDateString() === today.toDateString()) dateKey = "Today";
      else if (dateObj.toDateString() === yesterday.toDateString()) dateKey = "Yesterday";
      else dateKey = dateObj.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
      if (!groups[dateKey]) groups[dateKey] = [];
      groups[dateKey].push(msg);
    });
    return groups;
  }, [messages]);

  async function handleDeleteConversation() {
    if (!confirm('WARNING: TERMINATE CONNECTION?')) return;
    const res = await api.deleteConversation(convId);
    if (res.success) router.push('/chat');
  }

  useEffect(() => {
    const handleClick = () => setContextMenu(null);
    if (contextMenu) window.addEventListener('click', handleClick);
    return () => window.removeEventListener('click', handleClick);
  }, [contextMenu]);

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#050505] text-cyan-500 font-mono">
        <div className="flex flex-col items-center gap-4">
          <div className="w-16 h-16 border-4 border-cyan-500/30 border-t-cyan-500 rounded-full animate-spin"></div>
          <span className="animate-pulse tracking-[0.2em]">INITIALIZING...</span>
        </div>
      </div>
    );
  }

  const dateKeys = Object.keys(groupedMessages);

  return (
    <div className="flex-1 flex flex-col bg-[#050505] text-gray-200 relative overflow-hidden font-sans">
      <div className="scifi-grid"></div>
      <div className="absolute inset-0 z-10 pointer-events-none opacity-5 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] bg-repeat"></div>

      {/* HEADER */}
      <header className="relative z-20 px-3 sm:px-6 py-3 bg-[#0a0a0f]/95 backdrop-blur-md border-b border-cyan-900/30 flex justify-between items-center shadow-lg shadow-cyan-900/10">
        <div className="flex items-center gap-2 sm:gap-4 min-w-0">
          <div className="relative shrink-0">
            <div className="w-9 h-9 sm:w-10 sm:h-10 border border-cyan-500/50 flex items-center justify-center text-cyan-400 font-bold bg-cyan-950/20 text-sm sm:text-base" style={{ clipPath: 'polygon(20% 0, 100% 0, 100% 80%, 80% 100%, 0 100%, 0 20%)' }}>
              {getConversationName().charAt(0).toUpperCase()}
            </div>
            {conversation?.type === 'direct' && (
              <div className={`absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-[#0a0a0f] ${isOnline ? 'bg-emerald-500' : 'bg-gray-500'}`}></div>
            )}
          </div>
          <div className="min-w-0">
            <h2 className="text-sm sm:text-lg font-bold text-cyan-100 uppercase tracking-wide truncate">{getConversationName()}</h2>
            {typing.length > 0 ? (
              <p className="text-[9px] sm:text-[10px] text-cyan-400 font-mono animate-pulse">Incoming Data Stream...</p>
            ) : (
              <p className={`text-[9px] sm:text-[10px] font-mono tracking-widest ${isOnline ? 'text-emerald-500' : 'text-gray-500'}`}>
                {conversation?.type === 'direct' ? (isOnline ? 'ONLINE' : 'OFFLINE') : 'SECURE_LINK_ACTIVE'}
              </p>
            )}
          </div>
        </div>
        <div className="relative shrink-0">
          <button onClick={() => setShowMenu(!showMenu)} className="p-2 text-cyan-600 hover:text-cyan-400 hover:bg-cyan-900/20 rounded">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
            </svg>
          </button>
          {showMenu && (
            <>
              <div className="fixed inset-0 z-30" onClick={() => setShowMenu(false)}></div>
              <div className="absolute right-0 top-full mt-2 w-48 bg-[#0c0c14] border border-red-900/50 z-40 shadow-xl">
                <button onClick={handleDeleteConversation} className="w-full text-left px-4 py-3 text-red-500 text-xs font-mono hover:bg-red-950/30 uppercase">Terminate Connection</button>
              </div>
            </>
          )}
        </div>
      </header>

      {/* MESSAGES AREA */}
      <div ref={scrollContainerRef} onScroll={handleScroll} className="flex-1 overflow-y-auto p-3 sm:p-4 md:p-6 space-y-6 sm:space-y-8 relative z-10 scroll-smooth no-scrollbar">
        {messages.length === 0 && <div className="text-center py-20 font-mono text-cyan-500/30 text-sm">NO DATA LOGS FOUND</div>}
        {dateKeys.map((dateLabel) => (
          <div key={dateLabel} id={`date-${dateLabel}`}>
            <div className="flex items-center justify-center my-4 sm:my-6 opacity-70">
              <div className="h-px w-8 sm:w-12 bg-gradient-to-r from-transparent to-cyan-500/50"></div>
              <span className="mx-2 sm:mx-4 text-[9px] sm:text-[10px] font-mono text-cyan-500 bg-[#0a0a0f] px-2 border border-cyan-500/30 rounded-sm uppercase tracking-widest shadow-[0_0_10px_rgba(6,182,212,0.2)]">{dateLabel}</span>
              <div className="h-px w-8 sm:w-12 bg-gradient-to-l from-transparent to-cyan-500/50"></div>
            </div>
            <div className="space-y-3 sm:space-y-4">
              {groupedMessages[dateLabel].map((msg) => {
                const isOwn = msg.sender_id === user?.id;
                const isEdited = msg.updated_at && msg.created_at !== msg.updated_at;
                const isBeingEdited = editingMsg?.id === msg.id;
                const isDeleted = !!msg.deleted_at;
                return (
                  <div key={msg.id} className={`flex ${isOwn ? 'justify-end' : 'justify-start'} msg-enter`} onContextMenu={(e) => handleContextMenu(e, msg)}>
                    <div className={`max-w-[90%] sm:max-w-[85%] md:max-w-[70%] group flex flex-col ${isOwn ? 'items-end' : 'items-start'}`}>
                      {!isOwn && <span className="text-[8px] sm:text-[9px] font-mono text-cyan-600/70 mb-1 ml-1 uppercase">{formatName(msg.sender?.name, msg.sender?.email)}</span>}
                      <div className={`relative px-3 sm:px-5 py-2 sm:py-3 text-xs sm:text-sm md:text-base backdrop-blur-sm transition-all duration-200 ${isDeleted ? 'opacity-50' : ''} ${isOwn ? 'bg-cyan-950/60 border border-cyan-500/40 text-cyan-50 scifi-clip-sender' : 'bg-[#15151a] border-l-2 border-l-orange-500/50 text-gray-300 scifi-clip-receiver'}`}>
                        {isBeingEdited ? (
                          <div className="flex items-center gap-2">
                            <input ref={editInputRef} type="text" value={editContent} onChange={(e) => setEditContent(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter') handleEditSave(); if (e.key === 'Escape') { setEditingMsg(null); setEditContent(''); } }} className="bg-transparent border-b border-cyan-500 outline-none text-cyan-50 min-w-[100px] sm:min-w-[150px]" />
                            <button onClick={handleEditSave} className="text-emerald-400 hover:text-emerald-300"><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg></button>
                            <button onClick={() => { setEditingMsg(null); setEditContent(''); }} className="text-red-400 hover:text-red-300"><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg></button>
                          </div>
                        ) : msg.type === 'voice' && msg.voice_file_path ? (
                          <>
                            <VoicePlayer
                              audioUrl={`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/voice/download?path=${encodeURIComponent(msg.voice_file_path)}`}
                              duration={msg.voice_duration || 0}
                              waveform={msg.voice_waveform}
                              senderName={formatName(msg.sender?.name, msg.sender?.email)}
                              isOwn={isOwn}
                            />
                          </>
                        ) : msg.type === 'voice' && !msg.voice_file_path ? (
                          <div className="flex items-center gap-2 p-2 sm:p-3 rounded bg-cyan-950/40 min-w-[200px] sm:min-w-[280px]">
                            <div className="w-8 h-8 sm:w-10 sm:h-10 rounded-full bg-cyan-500/30 flex items-center justify-center shrink-0">
                              <div className="w-4 h-4 border-2 border-cyan-400/30 border-t-cyan-400 rounded-full animate-spin" />
                            </div>
                            <div className="flex-1 min-w-0">
                              <div className="h-8 sm:h-10 flex items-center gap-0.5 px-1">
                                {(msg.voice_waveform || []).slice(-50).map((amplitude, i) => (
                                  <div key={i} className="flex-1 bg-cyan-500/30 rounded-full" style={{ height: `${Math.max(10, amplitude * 100)}%` }} />
                                ))}
                              </div>
                              <div className="flex items-center justify-between mt-1 px-1">
                                <span className="text-[9px] sm:text-[10px] font-mono text-cyan-400/60 animate-pulse">Sending...</span>
                                <span className="text-[9px] sm:text-[10px] font-mono text-gray-500">
                                  {Math.floor((msg.voice_duration || 0) / 60)}:{String(Math.floor((msg.voice_duration || 0) % 60)).padStart(2, '0')}
                                </span>
                              </div>
                            </div>
                          </div>
                        ) : (
                          <p className={`whitespace-pre-wrap leading-relaxed break-words ${isDeleted ? 'italic text-gray-500' : ''}`}>{msg.content}</p>
                        )}
                        
                        {/* Reactions */}
                        {msg.reactions && msg.reactions.length > 0 && (
                          <div className="flex flex-wrap gap-1 mt-2">
                            {msg.reactions.map((reaction) => {
                              const hasReacted = reaction.user_ids.includes(user?.id || 0);
                              return (
                                <button
                                  key={reaction.emoji}
                                  onClick={() => api.reactToMessage(msg.id, reaction.emoji)}
                                  className={`flex items-center gap-1 px-2 py-0.5 rounded-full text-xs transition-all ${
                                    hasReacted 
                                      ? 'bg-cyan-500/30 border border-cyan-500/50 text-cyan-300' 
                                      : 'bg-gray-800/50 border border-gray-700/50 text-gray-400 hover:bg-gray-700/50'
                                  }`}
                                  title={`Reacted by ${reaction.user_ids.length} user(s)`}
                                >
                                  {reaction.is_custom && reaction.custom_url ? (
                                    <img src={reaction.custom_url} alt={reaction.emoji} className="w-4 h-4" />
                                  ) : (
                                    <span>{reaction.emoji}</span>
                                  )}
                                  <span className="text-[10px]">{reaction.count}</span>
                                </button>
                              );
                            })}
                          </div>
                        )}
                        
                        {copiedId === msg.id && <span className="absolute -top-6 left-1/2 -translate-x-1/2 text-[10px] text-emerald-400 font-mono bg-[#0a0a0f] px-2 py-0.5 border border-emerald-500/30 rounded">COPIED</span>}
                      </div>
                      <div className="flex items-center gap-2 mt-1 px-1">
                        <span className="text-[8px] sm:text-[9px] font-mono text-gray-500 uppercase">{formatTime(msg.created_at)}</span>
                        {isEdited && <span className="text-[7px] sm:text-[8px] font-mono text-yellow-500/70 italic tracking-wider">[MODIFIED]</span>}
                        {isDeleted && <span className="text-[7px] sm:text-[8px] font-mono text-red-500/70 italic tracking-wider">[DELETED]</span>}
                        {!isDeleted && (
                          <div className="hidden sm:flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                            <button onClick={() => handleQuickReact(msg, '👍')} className="p-1 text-gray-500 hover:text-cyan-400" title="React">😊</button>
                            <button onClick={() => handleCopy(msg)} className="p-1 text-gray-500 hover:text-cyan-400" title="Copy"><svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg></button>
                            {isOwn && <button onClick={() => startEdit(msg)} className="p-1 text-gray-500 hover:text-cyan-400" title="Edit"><svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg></button>}
                            {isOwn && <button onClick={() => handleDelete(msg)} className="p-1 text-gray-500 hover:text-red-400" title="Delete"><svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg></button>}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* Context Menu */}
      {contextMenu && (
        <div className="fixed z-50 bg-[#0c0c14] border border-cyan-500/30 rounded shadow-2xl py-1 min-w-[120px]" style={{ top: contextMenu.y, left: Math.min(contextMenu.x, window.innerWidth - 140) }}>
          <div className="px-2 py-1 flex gap-1 border-b border-cyan-900/30">
            {['👍', '❤️', '😂', '😮', '😢', '🎉'].map(emoji => (
              <button
                key={emoji}
                onClick={() => { handleQuickReact(contextMenu.msg, emoji); setContextMenu(null); }}
                className="text-lg hover:scale-125 transition-transform"
              >
                {emoji}
              </button>
            ))}
          </div>
          <button onClick={() => handleCopy(contextMenu.msg)} className="w-full text-left px-4 py-2 text-xs text-cyan-300 hover:bg-cyan-900/30 font-mono flex items-center gap-2">
            <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>Copy
          </button>
          {contextMenu.msg.sender_id === user?.id && !contextMenu.msg.deleted_at && (
            <>
              <button onClick={() => startEdit(contextMenu.msg)} className="w-full text-left px-4 py-2 text-xs text-cyan-300 hover:bg-cyan-900/30 font-mono flex items-center gap-2">
                <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>Edit
              </button>
              <button onClick={() => handleDelete(contextMenu.msg)} className="w-full text-left px-4 py-2 text-xs text-red-400 hover:bg-red-900/30 font-mono flex items-center gap-2">
                <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>Delete
              </button>
            </>
          )}
        </div>
      )}

      {/* Floating Navigation */}
      <div className="absolute right-3 sm:right-6 bottom-20 sm:bottom-24 z-30 flex flex-col gap-2 sm:gap-3 items-end pointer-events-none">
        {showScrollDownBtn && (
          <button onClick={() => scrollToBottom(true)} className="pointer-events-auto w-9 h-9 sm:w-10 sm:h-10 bg-cyan-500 text-black rounded-full shadow-[0_0_15px_rgba(6,182,212,0.6)] flex items-center justify-center hover:scale-110 transition-transform animate-bounce" title="Jump to latest">
            <svg className="w-4 h-4 sm:w-5 sm:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" /></svg>
          </button>
        )}
        {dateKeys.length > 1 && (
          <div className="pointer-events-auto relative flex flex-col items-end gap-2">
            {showJumpMenu && (
              <div className="bg-[#0c0c14] border border-cyan-500/30 rounded p-2 shadow-2xl space-y-1 mb-2 min-w-[100px] sm:min-w-[120px] backdrop-blur-xl origin-bottom-right transition-all max-h-48 overflow-y-auto no-scrollbar">
                <div className="text-[9px] sm:text-[10px] text-cyan-500/50 uppercase font-mono border-b border-cyan-900/50 pb-1 mb-1 text-center">Time Jump</div>
                {dateKeys.map(date => (
                  <button key={date} onClick={() => scrollToDate(date)} className="w-full text-right px-2 py-1 sm:py-1.5 text-[10px] sm:text-xs text-cyan-300 hover:bg-cyan-900/30 rounded font-mono block hover:text-white transition-colors">{date}</button>
                ))}
              </div>
            )}
            <button onClick={() => setShowJumpMenu(!showJumpMenu)} className={`w-9 h-9 sm:w-10 sm:h-10 bg-[#1a1a20] border border-cyan-500/30 text-cyan-400 rounded-full flex items-center justify-center hover:bg-cyan-900/30 hover:border-cyan-400 transition-all ${showJumpMenu ? 'rotate-180 bg-cyan-900/50' : ''}`} title="History Navigation">
              {showJumpMenu ? <svg className="w-4 h-4 sm:w-5 sm:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg> : <svg className="w-4 h-4 sm:w-5 sm:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" /></svg>}
            </button>
          </div>
        )}
      </div>

      {/* INPUT AREA */}
      <div className="p-3 sm:p-4 md:p-6 bg-[#0a0a0f] border-t border-cyan-900/30 relative z-20 shadow-[0_-5px_20px_rgba(0,0,0,0.5)]">
        <form onSubmit={handleSend} className="max-w-4xl mx-auto flex gap-2 sm:gap-3 items-stretch">
          {/* Voice Button */}
          <button
            type="button"
            onClick={() => setShowVoiceRecorder(true)}
            className="px-3 sm:px-4 py-2.5 sm:py-3 bg-cyan-900/30 border border-cyan-500/50 text-cyan-400 hover:bg-cyan-500 hover:text-black hover:shadow-[0_0_20px_rgba(6,182,212,0.4)] transition-all duration-300"
            style={{ clipPath: 'polygon(10px 0, 100% 0, 100% 100%, 0 100%, 0 10px)' }}
            title="Voice message"
          >
            <svg className="w-4 h-4 sm:w-5 sm:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
            </svg>
          </button>

          <div className="flex-1 relative">
            <input type="text" value={newMessage} onChange={handleInputChange} placeholder="TRANSMIT DATA..." className="w-full h-full bg-[#050508] text-cyan-100 border border-cyan-800/50 rounded-sm px-3 sm:px-4 py-2.5 sm:py-3 focus:outline-none focus:border-cyan-500 focus:shadow-[0_0_15px_rgba(6,182,212,0.1)] font-mono text-xs sm:text-sm placeholder-cyan-900 transition-all" />
          </div>
          <button type="submit" disabled={!newMessage.trim()} className="px-4 sm:px-6 py-2.5 sm:py-3 bg-cyan-900/30 border border-cyan-500/50 text-cyan-400 font-mono text-xs sm:text-sm tracking-widest hover:bg-cyan-500 hover:text-black hover:shadow-[0_0_20px_rgba(6,182,212,0.4)] transition-all duration-300 disabled:opacity-30 disabled:cursor-not-allowed uppercase" style={{ clipPath: 'polygon(10px 0, 100% 0, 100% 100%, 0 100%, 0 10px)' }}>
            <span className="hidden sm:inline">Send</span>
            <svg className="w-4 h-4 sm:hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" /></svg>
          </button>
        </form>
      </div>

      {/* Voice Recorder Modal */}
      {showVoiceRecorder && (
        <VoiceRecorder
          onSend={handleVoiceSend}
          onCancel={() => setShowVoiceRecorder(false)}
        />
      )}
    </div>
  );
}
