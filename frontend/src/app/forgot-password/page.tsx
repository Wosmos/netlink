'use client';

import { useState } from 'react';
import { api } from '@/lib/api';
import Link from 'next/link';

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!email.trim()) return;

    setLoading(true);
    setError('');
    setMessage('');

    const res = await api.forgotPassword(email.trim().toLowerCase());
    setLoading(false);

    if (res.success) {
      setMessage(res.message || 'If that email is registered, you will receive a reset link');
      setEmail('');
    } else {
      setError(res.error || 'An error occurred');
    }
  }

  return (
    <div className="min-h-screen bg-[#050505] flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="bg-[#0a0a0f] border border-cyan-900/30 rounded-lg p-8 shadow-2xl">
          <div className="text-center mb-8">
            <h1 className="text-2xl font-bold text-cyan-100 mb-2">Reset Password</h1>
            <p className="text-gray-400 text-sm">Enter your email to receive a reset link</p>
          </div>

          {message && (
            <div className="mb-6 p-4 bg-emerald-950/30 border border-emerald-500/30 rounded text-emerald-400 text-sm">
              {message}
            </div>
          )}

          {error && (
            <div className="mb-6 p-4 bg-red-950/30 border border-red-500/30 rounded text-red-400 text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label className="block text-cyan-300 text-sm font-medium mb-2">
                Email Address
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-3 bg-[#050508] border border-cyan-800/50 rounded text-cyan-100 placeholder-cyan-900 focus:outline-none focus:border-cyan-500 focus:shadow-[0_0_15px_rgba(6,182,212,0.1)]"
                placeholder="Enter your email"
                required
                disabled={loading}
              />
            </div>

            <button
              type="submit"
              disabled={loading || !email.trim()}
              className="w-full py-3 bg-cyan-600 hover:bg-cyan-500 disabled:bg-cyan-800 disabled:cursor-not-allowed text-white font-medium rounded transition-colors"
            >
              {loading ? 'Sending...' : 'Send Reset Link'}
            </button>
          </form>

          <div className="mt-6 text-center">
            <Link href="/login" className="text-cyan-400 hover:text-cyan-300 text-sm">
              Back to Login
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}