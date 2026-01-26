'use client';

import { useLanguage } from '@/context/LanguageContext';

export default function ChatPage() {
  const { t } = useLanguage();
  
  return (
    <div className="flex-1 flex items-center justify-center bg-[#050505] relative overflow-hidden">
      {/* Animated Grid Background */}
      <div className="absolute inset-0 opacity-20">
        <div className="absolute inset-0" style={{
          backgroundImage: `
            linear-gradient(to right, rgba(6, 182, 212, 0.1) 1px, transparent 1px),
            linear-gradient(to bottom, rgba(6, 182, 212, 0.1) 1px, transparent 1px)
          `,
          backgroundSize: '50px 50px'
        }}></div>
      </div>
      
      {/* Content */}
      <div className="text-center relative z-10">
        <div className="w-20 h-20 mx-auto mb-6 border border-cyan-500/50 flex items-center justify-center bg-cyan-950/20" 
          style={{ clipPath: 'polygon(20% 0, 100% 0, 100% 80%, 80% 100%, 0 100%, 0 20%)' }}>
          <svg className="w-10 h-10 text-cyan-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
        </div>
        <h2 className="text-xl font-mono text-cyan-100 uppercase tracking-widest mb-2">
          {t('SELECT_TARGET', 'Select a Chat')}
        </h2>
        <p className="text-sm text-cyan-500/50 font-mono">
          {t('INITIALIZE_CONNECTION_FROM_SIDEBAR', 'Choose a conversation from the sidebar')}
        </p>
      </div>
    </div>
  );
}
