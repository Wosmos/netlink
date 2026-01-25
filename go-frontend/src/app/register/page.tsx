'use client';

import { useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { useRouter } from 'next/navigation';
import Link from 'next/link';

export default function RegisterPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const { register } = useAuth();
  const router = useRouter();

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (password !== confirmPassword) {
      setError('SECURITY MISMATCH: PASSWORDS DO NOT ALIGN');
      return;
    }
    if (password.length < 8) {
      setError('SECURITY ALERT: PASSWORD LENGTH < 8 CHARS');
      return;
    }

    setLoading(true);

    // Artificial delay for effect
    const minTime = new Promise(resolve => setTimeout(resolve, 800));

    const [result] = await Promise.all([
      register(email, password),
      minTime
    ]);

    if (result.success) {
      setSuccess('REGISTRATION COMPLETE. VERIFICATION LINK DISPATCHED.');
      setTimeout(() => router.push('/login'), 2000);
    } else {
      setError(result.error || 'REGISTRATION SEQUENCE FAILED');
      setLoading(false); // Only stop loading on error, on success we wait for redirect
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#050505] px-4 relative overflow-hidden font-sans selection:bg-cyan-500/30">
      
      {/* --- CSS/ANIMATIONS --- */}
      <style jsx global>{`
        @keyframes gridMove {
          0% { transform: perspective(500px) rotateX(60deg) translateY(0); }
          100% { transform: perspective(500px) rotateX(60deg) translateY(50px); }
        }
        @keyframes scan {
          0% { background-position: 0% 0%; }
          100% { background-position: 0% 100%; }
        }
        @keyframes glitch-skew {
          0% { transform: skew(0deg); }
          20% { transform: skew(-2deg); }
          40% { transform: skew(2deg); }
          60% { transform: skew(-1deg); }
          80% { transform: skew(1deg); }
          100% { transform: skew(0deg); }
        }
        .scifi-grid {
          position: absolute;
          inset: -100% 0 0 0;
          background-image: 
            linear-gradient(to right, rgba(6, 182, 212, 0.1) 1px, transparent 1px),
            linear-gradient(to bottom, rgba(6, 182, 212, 0.1) 1px, transparent 1px);
          background-size: 60px 60px;
          animation: gridMove 20s linear infinite;
          opacity: 0.2;
          z-index: 0;
          pointer-events: none;
        }
        .clip-box {
          clip-path: polygon(
            0 0, 
            100% 0, 
            100% calc(100% - 20px), 
            calc(100% - 20px) 100%, 
            0 100%
          );
        }
        .clip-btn {
          clip-path: polygon(
            15px 0, 
            100% 0, 
            100% calc(100% - 15px), 
            calc(100% - 15px) 100%, 
            0 100%, 
            0 15px
          );
        }
      `}</style>

      {/* Background Layer */}
      <div className="scifi-grid"></div>
      <div className="absolute inset-0 z-10 pointer-events-none opacity-5 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] bg-repeat"></div>
      <div className="absolute inset-0 bg-radial-gradient from-cyan-900/10 via-transparent to-transparent pointer-events-none"></div>

      {/* --- REGISTER CARD --- */}
      <div className="max-w-md w-full relative z-20 group">
        
        {/* Border Glow */}
        <div className="absolute -inset-0.5 bg-gradient-to-r from-cyan-600 via-blue-600 to-cyan-600 opacity-20 group-hover:opacity-40 transition duration-1000 blur-sm"></div>

        <div className="bg-[#0c0c12] border border-cyan-900/50 p-8 clip-box shadow-[0_0_50px_rgba(6,182,212,0.1)] relative">
          
          {/* Header */}
          <div className="text-center mb-8 relative">
             <div className="inline-block relative">
                <h1 className="text-4xl font-black text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-blue-500 tracking-tighter uppercase drop-shadow-[0_0_10px_rgba(6,182,212,0.5)]">
                Netlink
                </h1>
                <div className="absolute -bottom-2 left-0 w-full h-[1px] bg-cyan-900/50">
                    <div className="absolute top-0 left-0 h-full w-1/3 bg-cyan-500/50 animate-[scan_2s_linear_infinite]"></div>
                </div>
            </div>
            <p className="mt-4 text-xs font-mono text-cyan-500/50 tracking-[0.2em] uppercase">
              New Unit Initialization
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            
            {/* Alerts */}
            {error && (
              <div className="bg-red-950/20 border border-red-500/30 text-red-400 p-3 text-xs font-mono flex items-center gap-2 animate-[glitch-skew_0.3s_ease-out]">
                <span className="text-red-500 text-lg">⚠</span>
                <span className="uppercase tracking-wide">{error}</span>
              </div>
            )}
            {success && (
              <div className="bg-emerald-950/20 border border-emerald-500/30 text-emerald-400 p-3 text-xs font-mono flex items-center gap-2">
                <span className="text-emerald-500 text-lg">✓</span>
                <span className="uppercase tracking-wide">{success}</span>
              </div>
            )}

            {/* Email Input */}
            <div className="group/input">
              <label htmlFor="email" className="block text-[10px] font-mono text-cyan-400/70 mb-1 uppercase tracking-wider group-focus-within/input:text-cyan-400 transition-colors">
                Primary ID (Email)
              </label>
              <div className="relative">
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  className="block w-full bg-[#050508] border border-cyan-900 text-cyan-50 px-4 py-2.5 focus:border-cyan-500/80 focus:ring-1 focus:ring-cyan-500/50 focus:outline-none transition-all font-mono text-sm placeholder-cyan-900/50"
                  placeholder="USER@NET.COM"
                />
                <div className="absolute top-0 right-0 w-2 h-2 border-t border-r border-cyan-500/30 group-focus-within/input:border-cyan-400 transition-colors"></div>
                <div className="absolute bottom-0 left-0 w-2 h-2 border-b border-l border-cyan-500/30 group-focus-within/input:border-cyan-400 transition-colors"></div>
              </div>
            </div>

            {/* Password Input */}
<div className="group/input">
  <label htmlFor="password" className="block text-[10px] font-mono text-cyan-400/70 mb-1 uppercase tracking-wider group-focus-within/input:text-cyan-400 transition-colors">
    Define Security Key
  </label>
  <div className="relative">
    {/* Fake visual layer (shows *) */}
    <div className="w-full bg-[#050508] border border-cyan-900 px-4 py-2.5 font-mono text-sm text-cyan-50">
      {password ? '*'.repeat(password.length) : (
        <span className="text-cyan-900/50">MIN. 8 CHARS</span>
      )}
    </div>

    {/* Real password input (invisible but functional) */}
    <input
      id="password"
      type="password"
      value={password}
      onChange={(e) => setPassword(e.target.value)}
      required
      className="absolute inset-0 w-full h-full opacity-0 cursor-text focus:outline-none"
      placeholder="MIN. 8 CHARS"
      autoComplete="new-password"
    />

    {/* Decorative corners */}
    <div className="absolute top-0 right-0 w-2 h-2 border-t border-r border-cyan-500/30 group-focus-within/input:border-cyan-400 transition-colors"></div>
    <div className="absolute bottom-0 left-0 w-2 h-2 border-b border-l border-cyan-500/30 group-focus-within/input:border-cyan-400 transition-colors"></div>
  </div>
</div>

{/* Confirm Password Input */}
<div className="group/input">
  <label htmlFor="confirmPassword" className="block text-[10px] font-mono text-cyan-400/70 mb-1 uppercase tracking-wider group-focus-within/input:text-cyan-400 transition-colors">
    Verify Security Key
  </label>
  <div className="relative">
    {/* Fake visual layer (shows *) */}
    <div className="w-full bg-[#050508] border border-cyan-900 px-4 py-2.5 font-mono text-sm text-cyan-50">
      {confirmPassword ? '*'.repeat(confirmPassword.length) : (
        <span className="text-cyan-900/50">CONFIRM KEY</span>
      )}
    </div>

    {/* Real password input (invisible but functional) */}
    <input
      id="confirmPassword"
      type="password"
      value={confirmPassword}
      onChange={(e) => setConfirmPassword(e.target.value)}
      required
      className="absolute inset-0 w-full h-full opacity-0 cursor-text focus:outline-none"
      placeholder="CONFIRM KEY"
      autoComplete="new-password"
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
              className="
                w-full relative overflow-hidden group/btn mt-2
                bg-cyan-950/50 border border-cyan-500/50 
                text-cyan-400 font-bold font-mono tracking-widest uppercase
                py-3.5 clip-btn
                hover:bg-cyan-500 hover:text-black hover:shadow-[0_0_20px_rgba(6,182,212,0.4)]
                focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-cyan-500 focus:ring-offset-black
                transition-all duration-300
                disabled:opacity-50 disabled:cursor-wait
              "
            >
              <div className="relative z-10 flex items-center justify-center gap-2">
                {loading ? (
                  <>
                    <span className="w-3 h-3 border-2 border-black/50 border-t-black rounded-full animate-spin"></span>
                    <span className="text-xs">ESTABLISHING ID...</span>
                  </>
                ) : (
                  <>
                    <span className="text-sm">CREATE ACCOUNT</span>
                    <svg className="w-3 h-3 group-hover/btn:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" /></svg>
                  </>
                )}
              </div>
              
              {/* Button Scan Effect */}
              <div className="absolute inset-0 bg-white/20 translate-y-full group-hover/btn:translate-y-[-200%] transition-transform duration-700 pointer-events-none"></div>
            </button>

            {/* Footer Link */}
            <div className="text-center mt-6">
              <p className="text-xs text-gray-500">
                IDENTITY EXISTS?{' '}
                <Link href="/login" className="text-cyan-600 hover:text-cyan-400 font-mono uppercase tracking-wide border-b border-transparent hover:border-cyan-400 transition-all ml-1">
                  ACCESS TERMINAL
                </Link>
              </p>
            </div>

            {/* Tech Decoration Bottom */}
            <div className="flex justify-between items-end opacity-30 mt-8 pt-4 border-t border-cyan-900/30">
                <span className="text-[10px] font-mono text-cyan-600">MOD_REG_V1</span>
                <div className="flex gap-0.5">
                    <div className="w-3 h-[2px] bg-cyan-500"></div>
                    <div className="w-1 h-[2px] bg-cyan-500"></div>
                </div>
            </div>

          </form>
        </div>
      </div>
    </div>
  );
}