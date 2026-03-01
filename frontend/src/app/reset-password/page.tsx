'use client';

import { useState, useEffect, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { api } from '@/lib/api';
import Link from 'next/link';

function ResetPasswordForm() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const [token, setToken] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    const tokenParam = searchParams.get('token');
    if (tokenParam) {
      setToken(tokenParam);
    } else {
      setError('Invalid reset link');
    }
  }, [searchParams]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!password || !confirmPassword || !token) return;

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    setLoading(true);
    setError('');

    const res = await api.resetPassword(token, password);
    setLoading(false);

    if (res.success) {
      router.push('/login?success=Password reset successful! You can now log in.');
    } else {
      setError(res.error || 'An error occurred');
    }
  }

  if (!token && !error) {
    return (
      <div className="min-h-screen bg-[#050505] flex items-center justify-center">
        <div className="text-cyan-500">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#050505] flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="bg-[#0a0a0f] border border-cyan-900/30 rounded-lg p-8 shadow-2xl">
          <div className="text-center mb-8">
            <h1 className="text-2xl font-bold text-cyan-100 mb-2">Set New Password</h1>
            <p className="text-gray-400 text-sm">Enter your new password below</p>
          </div>

          {error && (
            <div className="mb-6 p-4 bg-red-950/30 border border-red-500/30 rounded text-red-400 text-sm">
              {error}
            </div>
          )}

          {token ? (
            <form onSubmit={handleSubmit} className="space-y-6">
              <div>
                <label className="block text-cyan-300 text-sm font-medium mb-2">
                  New Password
                </label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full px-4 py-3 bg-[#050508] border border-cyan-800/50 rounded text-cyan-100 placeholder-cyan-900 focus:outline-none focus:border-cyan-500 focus:shadow-[0_0_15px_rgba(6,182,212,0.1)]"
                  placeholder="Enter new password"
                  required
                  disabled={loading}
                  minLength={8}
                />
              </div>

              <div>
                <label className="block text-cyan-300 text-sm font-medium mb-2">
                  Confirm Password
                </label>
                <input
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="w-full px-4 py-3 bg-[#050508] border border-cyan-800/50 rounded text-cyan-100 placeholder-cyan-900 focus:outline-none focus:border-cyan-500 focus:shadow-[0_0_15px_rgba(6,182,212,0.1)]"
                  placeholder="Confirm new password"
                  required
                  disabled={loading}
                  minLength={8}
                />
              </div>

              <button
                type="submit"
                disabled={loading || !password || !confirmPassword || password !== confirmPassword}
                className="w-full py-3 bg-cyan-600 hover:bg-cyan-500 disabled:bg-cyan-800 disabled:cursor-not-allowed text-white font-medium rounded transition-colors"
              >
                {loading ? 'Resetting...' : 'Reset Password'}
              </button>
            </form>
          ) : (
            <div className="text-center">
              <p className="text-red-400 mb-4">Invalid or expired reset link</p>
              <Link href="/forgot-password" className="text-cyan-400 hover:text-cyan-300">
                Request a new reset link
              </Link>
            </div>
          )}

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

export default function ResetPasswordPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-[#050505] flex items-center justify-center">
        <div className="text-cyan-500">Loading...</div>
      </div>
    }>
      <ResetPasswordForm />
    </Suspense>
  );
}
