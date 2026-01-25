'use client';

import { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { api, Message, Conversation } from '@/lib/api';
import { wsClient } from '@/lib/websocket';
import { useAuth } from '@/context/AuthContext';

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
      setLoading(true);
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
      setTimeout(() => scrollToBottom(false), 100);
    }
    
    fetchData();
    
    const interval = setInterval(async () => {
      if (document.hidden) return;
      const [msgRes, onlineRes] = await Promise.all([api.getMessages(convId), api.getOnlineUsers()]);
      if (mounted && msgRes.success && msgRes.data) {
        setMessages(prev => JSON.stringify(prev) !== JSON.stringify(msgRes.data) ? msgRes.data! : prev);
      }
      if (mounted && onlineRes.success && onlineRes.data) setOnlineUsers(onlineRes.data.online_users || []);
    }, 3000);

    return () => { mounted = false; clearInterval(interval); };
  }, [convId, scrollToBottom]);

  useEffect(() => {
    const unsubMessage = wsClient.on('message', (event, data) => {
      if (event.conversation_id === convId && data) {
        setMessages(prev => prev.some(m => m.id === data.id) ? prev : [...prev, data]);
        setTyping(prev => prev.filter(uid => uid !== data.sender_id));
        api.markAsRead(convId);
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
    return () => { unsubMessage(); unsubTyping(); unsubOnline(); unsubOffline(); };
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
    const res = await api.sendMessage(convId, content);
    if (res.success && res.data) {
      setMessages(prev => prev.some(m => m.id === res.data!.id) ? prev : [...prev, res.data!]);
      setTimeout(() => scrollToBottom(true), 50);
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
      setMessages(prev => prev.map(m => m.id === editingMsg.id ? res.data! : m));
    }
    setEditingMsg(null);
    setEditContent('');
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
    console.log(dateStr, 'dateStr')
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
      <style jsx global>{`
        @keyframes gridMove { 0% { transform: perspective(500px) rotateX(60deg) translateY(0); } 100% { transform: perspective(500px) rotateX(60deg) translateY(50px); } }
        @keyframes fadePop { 0% { opacity: 0; transform: scale(0.98); } 100% { opacity: 1; transform: scale(1); } }
        .scifi-grid { position: absolute; inset: -100% 0 0 0; background-image: linear-gradient(to right, rgba(6, 182, 212, 0.05) 1px, transparent 1px), linear-gradient(to bottom, rgba(6, 182, 212, 0.05) 1px, transparent 1px); background-size: 50px 50px; animation: gridMove 20s linear infinite; opacity: 0.3; z-index: 0; pointer-events: none; }
        .msg-enter { animation: fadePop 0.2s ease-out forwards; }
        .scifi-clip-sender { clip-path: polygon(10px 0, 100% 0, 100% calc(100% - 10px), calc(100% - 10px) 100%, 0 100%, 0 10px); }
        .scifi-clip-receiver { clip-path: polygon(0 0, calc(100% - 10px) 0, 100% 10px, 100% 100%, 10px 100%, 0 calc(100% - 10px)); }
        .no-scrollbar::-webkit-scrollbar { display: none; }
        .no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
      `}</style>

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
                return (
                  <div key={msg.id} className={`flex ${isOwn ? 'justify-end' : 'justify-start'} msg-enter`} onContextMenu={(e) => handleContextMenu(e, msg)}>
                    <div className={`max-w-[90%] sm:max-w-[85%] md:max-w-[70%] group flex flex-col ${isOwn ? 'items-end' : 'items-start'}`}>
                      {!isOwn && <span className="text-[8px] sm:text-[9px] font-mono text-cyan-600/70 mb-1 ml-1 uppercase">{formatName(msg.sender?.name, msg.sender?.email)}</span>}
                      <div className={`relative px-3 sm:px-5 py-2 sm:py-3 text-xs sm:text-sm md:text-base backdrop-blur-sm transition-all duration-200 ${isOwn ? 'bg-cyan-950/60 border border-cyan-500/40 text-cyan-50 scifi-clip-sender' : 'bg-[#15151a] border-l-2 border-l-orange-500/50 text-gray-300 scifi-clip-receiver'}`}>
                        {isBeingEdited ? (
                          <div className="flex items-center gap-2">
                            <input ref={editInputRef} type="text" value={editContent} onChange={(e) => setEditContent(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter') handleEditSave(); if (e.key === 'Escape') { setEditingMsg(null); setEditContent(''); } }} className="bg-transparent border-b border-cyan-500 outline-none text-cyan-50 min-w-[100px] sm:min-w-[150px]" />
                            <button onClick={handleEditSave} className="text-emerald-400 hover:text-emerald-300"><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg></button>
                            <button onClick={() => { setEditingMsg(null); setEditContent(''); }} className="text-red-400 hover:text-red-300"><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg></button>
                          </div>
                        ) : <p className="whitespace-pre-wrap leading-relaxed break-words">{msg.content}</p>}
                        {copiedId === msg.id && <span className="absolute -top-6 left-1/2 -translate-x-1/2 text-[10px] text-emerald-400 font-mono bg-[#0a0a0f] px-2 py-0.5 border border-emerald-500/30 rounded">COPIED</span>}
                      </div>
                      <div className="flex items-center gap-2 mt-1 px-1">
                        <span className="text-[8px] sm:text-[9px] font-mono text-gray-500 uppercase">{formatTime(msg.created_at)}</span>
                        {isEdited && <span className="text-[7px] sm:text-[8px] font-mono text-yellow-500/70 italic tracking-wider">[MODIFIED]</span>}
                        <div className="hidden sm:flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                          <button onClick={() => handleCopy(msg)} className="p-1 text-gray-500 hover:text-cyan-400" title="Copy"><svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg></button>
                          {isOwn && <button onClick={() => startEdit(msg)} className="p-1 text-gray-500 hover:text-cyan-400" title="Edit"><svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg></button>}
                        </div>
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
          <button onClick={() => handleCopy(contextMenu.msg)} className="w-full text-left px-4 py-2 text-xs text-cyan-300 hover:bg-cyan-900/30 font-mono flex items-center gap-2">
            <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>Copy
          </button>
          {contextMenu.msg.sender_id === user?.id && (
            <button onClick={() => startEdit(contextMenu.msg)} className="w-full text-left px-4 py-2 text-xs text-cyan-300 hover:bg-cyan-900/30 font-mono flex items-center gap-2">
              <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>Edit
            </button>
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
          <div className="flex-1 relative">
            <input type="text" value={newMessage} onChange={handleInputChange} placeholder="TRANSMIT DATA..." className="w-full h-full bg-[#050508] text-cyan-100 border border-cyan-800/50 rounded-sm px-3 sm:px-4 py-2.5 sm:py-3 focus:outline-none focus:border-cyan-500 focus:shadow-[0_0_15px_rgba(6,182,212,0.1)] font-mono text-xs sm:text-sm placeholder-cyan-900 transition-all" />
          </div>
          <button type="submit" disabled={!newMessage.trim()} className="px-4 sm:px-6 py-2.5 sm:py-3 bg-cyan-900/30 border border-cyan-500/50 text-cyan-400 font-mono text-xs sm:text-sm tracking-widest hover:bg-cyan-500 hover:text-black hover:shadow-[0_0_20px_rgba(6,182,212,0.4)] transition-all duration-300 disabled:opacity-30 disabled:cursor-not-allowed uppercase" style={{ clipPath: 'polygon(10px 0, 100% 0, 100% 100%, 0 100%, 0 10px)' }}>
            <span className="hidden sm:inline">Send</span>
            <svg className="w-4 h-4 sm:hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" /></svg>
          </button>
        </form>
      </div>
    </div>
  );
}
