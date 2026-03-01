'use client';

import { useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { useRouter } from 'next/navigation';
import Link from 'next/link';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const router = useRouter();

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);

    // Simulate a slight delay for the "processing" animation effect
    const minTime = new Promise(resolve => setTimeout(resolve, 800));
    
    const [result] = await Promise.all([
      login(email, password),
      minTime
    ]);

    if (result.success) {
      router.push('/chat');
    } else {
      setError(result.error || 'AUTHENTICATION FAILED');
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#050505] px-4 relative overflow-hidden font-sans selection:bg-cyan-500/30">
      
      {/* Background Layer */}
      <div className="scifi-grid"></div>
      <div className="absolute inset-0 z-10 pointer-events-none opacity-5 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] bg-repeat"></div>
      
      {/* Radial Gradient for depth */}
      <div className="absolute inset-0 bg-radial-gradient from-cyan-900/10 via-transparent to-transparent pointer-events-none"></div>

      {/* --- LOGIN CARD --- */}
      <div className="max-w-md w-full relative z-20 group">
        
        {/* Animated decorative border glow */}
        <div className="absolute -inset-0.5 bg-gradient-to-r from-cyan-500 via-blue-500 to-cyan-500 opacity-20 group-hover:opacity-40 transition duration-1000 blur-sm"></div>

        <div className="bg-[#0c0c12] border border-cyan-900/50 p-8 clip-box shadow-[0_0_50px_rgba(6,182,212,0.1)] relative">
          
          {/* Header */}
          <div className="text-center mb-8 relative">
            <div className="inline-block relative">
                <h1 className="text-4xl font-black text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-blue-500 tracking-tighter uppercase drop-shadow-[0_0_10px_rgba(6,182,212,0.5)]">
                NetLink
                </h1>
                <div className="absolute -bottom-2 left-0 w-full h-[1px] bg-cyan-900/50">
                    <div className="absolute top-0 left-0 h-full w-1/3 bg-cyan-500/50 animate-[scan_2s_linear_infinite]"></div>
                </div>
            </div>
            <p className="mt-4 text-xs font-mono text-cyan-500/50 tracking-[0.2em] uppercase">
              Secure Terminal Access
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            
            {/* Error Message */}
            {error && (
              <div className="bg-red-950/20 border border-red-500/30 text-red-400 p-3 text-xs font-mono flex items-center gap-2 animate-[glitch-skew_0.3s_ease-out]">
                <span className="text-red-500 text-lg">⚠</span>
                <span className="uppercase tracking-wide">{error}</span>
              </div>
            )}

            {/* Email Input */}
            <div className="group/input">
              <label htmlFor="email" className="block text-xs font-mono text-cyan-400/70 mb-2 uppercase tracking-wider group-focus-within/input:text-cyan-400 transition-colors">
                User Identity (Email)
              </label>
              <div className="relative">
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  placeholder="ENTER_ID"
                  className="block w-full bg-[#050508] border border-cyan-900 text-cyan-50 px-4 py-3 focus:border-cyan-500/80 focus:ring-1 focus:ring-cyan-500/50 focus:outline-none focus:bg-[#0a0a0f] transition-all font-mono text-sm placeholder-cyan-900/50"
                />
                {/* Decorative corner */}
                <div className="absolute top-0 right-0 w-2 h-2 border-t border-r border-cyan-500/30 group-focus-within/input:border-cyan-400 transition-colors"></div>
                <div className="absolute bottom-0 left-0 w-2 h-2 border-b border-l border-cyan-500/30 group-focus-within/input:border-cyan-400 transition-colors"></div>
              </div>
            </div>

          
{/* Password Input */}
<div className="group/input">
  <label htmlFor="password" className="block text-xs font-mono text-cyan-400/70 mb-2 uppercase tracking-wider group-focus-within/input:text-cyan-400 transition-colors">
    Security Key (Password)
  </label>
  <div className="relative">
    {/* Fake visual layer (shows *) */}
    <div className="w-full bg-[#050508] border border-cyan-900 rounded-sm px-4 py-3 font-mono text-sm text-cyan-50">
      {password ? '*'.repeat(password.length) : (
        <span className="text-cyan-900/50">******</span>
      )}
    </div>

    {/* Real password input (invisible but functional) */}
    <input
      id="password"
      type="password"
      value={password}
      onChange={(e) => setPassword(e.target.value)}
      required
      autoComplete="current-password"
      className="absolute inset-0 w-full h-full opacity-0 cursor-text focus:outline-none"
    />

    {/* Decorative corners */}
    <div className="absolute top-0 right-0 w-2 h-2 border-t border-r border-cyan-500/30 group-focus-within/input:border-cyan-400 transition-colors"></div>
    <div className="absolute bottom-0 left-0 w-2 h-2 border-b border-l border-cyan-500/30 group-focus-within/input:border-cyan-400 transition-colors"></div>
  </div>
</div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={loading}
              className="w-full relative overflow-hidden group/btn bg-cyan-950/50 border border-cyan-500/50 text-cyan-400 font-bold font-mono tracking-widest uppercase py-4 clip-btn hover:bg-cyan-500 hover:text-black hover:shadow-[0_0_20px_rgba(6,182,212,0.4)] focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-cyan-500 focus:ring-offset-black transition-all duration-300 disabled:opacity-50 disabled:cursor-wait"
            >
              <div className="relative z-10 flex items-center justify-center gap-2">
                {loading ? (
                  <>
                    <span className="w-3 h-3 border-2 border-black/50 border-t-black rounded-full animate-spin"></span>
                    <span>AUTHENTICATING...</span>
                  </>
                ) : (
                  <>
                    <span>INITIATE UPLINK</span>
                    <svg className="w-4 h-4 group-hover/btn:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" /></svg>
                  </>
                )}
              </div>
              
              {/* Button Scan Effect */}
              <div className="absolute inset-0 bg-white/20 translate-y-full group-hover/btn:translate-y-[-200%] transition-transform duration-700 pointer-events-none"></div>
            </button>

            {/* Footer Links */}
            <div className="text-center mt-6 space-y-3">
              <p className="text-xs text-gray-500">
                NO ACCESS ID?{' '}
                <Link href="/register" className="text-cyan-600 hover:text-cyan-400 font-mono uppercase tracking-wide border-b border-transparent hover:border-cyan-400 transition-all ml-1">
                  REQUEST_CLEARANCE
                </Link>
              </p>
              <p className="text-xs text-gray-500">
                LOST SECURITY KEY?{' '}
                <Link href="/forgot-password" className="text-orange-500 hover:text-orange-400 font-mono uppercase tracking-wide border-b border-transparent hover:border-orange-400 transition-all ml-1">
                  RESET_ACCESS
                </Link>
              </p>
            </div>

            {/* Tech Decoration Bottom */}
            <div className="flex justify-between items-end opacity-30 mt-8 pt-4 border-t border-cyan-900/30">
                <span className="text-[10px] font-mono text-cyan-600">SYS_V.2.0.4</span>
                <div className="flex gap-1">
                    <div className="w-1 h-1 bg-cyan-500 rounded-full animate-pulse"></div>
                    <div className="w-1 h-1 bg-cyan-500 rounded-full animate-pulse delay-75"></div>
                    <div className="w-1 h-1 bg-cyan-500 rounded-full animate-pulse delay-150"></div>
                </div>
            </div>

          </form>
        </div>
      </div>
    </div>
  );
}