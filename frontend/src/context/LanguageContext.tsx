'use client';

import { createContext, useContext, useEffect, useState, ReactNode } from 'react';

type LanguageMode = 'techy' | 'normie' | null;

interface LanguageContextType {
  mode: LanguageMode;
  setMode: (mode: LanguageMode) => void;
  t: (techy: string, normie: string) => string;
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

const STORAGE_KEY = 'netlink_language_mode';

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [mode, setModeState] = useState<LanguageMode>(null);
  const [showModal, setShowModal] = useState(false);

  useEffect(() => {
    const stored = localStorage.getItem(STORAGE_KEY) as LanguageMode;
    if (stored === 'techy' || stored === 'normie') {
      setModeState(stored);
    } else {
      setShowModal(true);
    }
  }, []);

  const setMode = (newMode: LanguageMode) => {
    setModeState(newMode);
    if (newMode) {
      localStorage.setItem(STORAGE_KEY, newMode);
    }
    setShowModal(false);
  };

  const t = (techy: string, normie: string) => {
    return mode === 'normie' ? normie : techy;
  };

  return (
    <LanguageContext.Provider value={{ mode, setMode, t }}>
      {children}
      
      {/* Language Selection Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/90 flex items-center justify-center z-[9999] p-4 backdrop-blur-sm">
          <div className="bg-[#0a0a0f] border-2 border-cyan-500/50 max-w-lg w-full p-6 sm:p-8 relative overflow-hidden">
            {/* Animated background */}
            <div className="absolute inset-0 opacity-10">
              <div className="absolute inset-0" style={{
                backgroundImage: `
                  linear-gradient(to right, rgba(6, 182, 212, 0.1) 1px, transparent 1px),
                  linear-gradient(to bottom, rgba(6, 182, 212, 0.1) 1px, transparent 1px)
                `,
                backgroundSize: '30px 30px'
              }}></div>
            </div>

            <div className="relative z-10">
              <div className="text-center mb-8">
                <div className="w-16 h-16 mx-auto mb-4 border-2 border-cyan-500/50 flex items-center justify-center bg-cyan-950/20" 
                  style={{ clipPath: 'polygon(20% 0, 100% 0, 100% 80%, 80% 100%, 0 100%, 0 20%)' }}>
                  <svg className="w-8 h-8 text-cyan-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                </div>
                <h2 className="text-2xl font-mono font-bold text-cyan-400 uppercase tracking-widest mb-2">
                  Interface Mode
                </h2>
                <p className="text-sm text-gray-400 font-mono">
                  Choose your preferred communication style
                </p>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {/* Techy Mode */}
                <button
                  onClick={() => setMode('techy')}
                  className="group p-6 bg-[#050508] border-2 border-cyan-500/30 hover:border-cyan-500 hover:bg-cyan-950/20 transition-all duration-300 relative overflow-hidden"
                  style={{ clipPath: 'polygon(15px 0, 100% 0, 100% calc(100% - 15px), calc(100% - 15px) 100%, 0 100%, 0 15px)' }}
                >
                  <div className="absolute inset-0 bg-gradient-to-br from-cyan-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
                  <div className="relative z-10">
                    <div className="w-12 h-12 mx-auto mb-3 border border-cyan-500/50 flex items-center justify-center bg-cyan-950/30 group-hover:scale-110 transition-transform">
                      <svg className="w-6 h-6 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                      </svg>
                    </div>
                    <h3 className="text-lg font-mono font-bold text-cyan-400 uppercase tracking-wider mb-2">
                      Techy
                    </h3>
                    <p className="text-xs text-gray-400 font-mono leading-relaxed">
                      TERMINAL_STYLE<br/>
                      SYSTEM_MESSAGES<br/>
                      TECH_JARGON_ENABLED
                    </p>
                  </div>
                </button>

                {/* Normie Mode */}
                <button
                  onClick={() => setMode('normie')}
                  className="group p-6 bg-[#050508] border-2 border-orange-500/30 hover:border-orange-500 hover:bg-orange-950/20 transition-all duration-300 relative overflow-hidden"
                  style={{ clipPath: 'polygon(15px 0, 100% 0, 100% calc(100% - 15px), calc(100% - 15px) 100%, 0 100%, 0 15px)' }}
                >
                  <div className="absolute inset-0 bg-gradient-to-br from-orange-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
                  <div className="relative z-10">
                    <div className="w-12 h-12 mx-auto mb-3 border border-orange-500/50 flex items-center justify-center bg-orange-950/30 group-hover:scale-110 transition-transform">
                      <svg className="w-6 h-6 text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <h3 className="text-lg font-mono font-bold text-orange-400 uppercase tracking-wider mb-2">
                      Normie
                    </h3>
                    <p className="text-xs text-gray-400 font-mono leading-relaxed">
                      Simple Language<br/>
                      Friendly Messages<br/>
                      Easy to Understand
                    </p>
                  </div>
                </button>
              </div>

              <p className="text-center text-xs text-gray-500 font-mono mt-6">
                You can change this later in settings
              </p>
            </div>
          </div>
        </div>
      )}
    </LanguageContext.Provider>
  );
}

export function useLanguage() {
  const context = useContext(LanguageContext);
  if (!context) throw new Error('useLanguage must be used within LanguageProvider');
  return context;
}
