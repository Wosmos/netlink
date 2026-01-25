'use client';
import { useAuth } from '@/context/AuthContext';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import Sidebar from '@/components/Sidebar';

export default function ChatLayout({ children }: { children: React.ReactNode }) {
  const { user, loading, logout } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !user) {
      router.push('/login');
    }
  }, [user, loading, router]);

  if (loading || !user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#050505]">
        <div className="flex flex-col items-center gap-4">
          <div className="w-16 h-16 border-4 border-cyan-500/30 border-t-cyan-500 rounded-full animate-spin"></div>
          <span className="text-cyan-500 font-mono animate-pulse tracking-[0.2em]">AUTHENTICATING...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen flex bg-[#050505] overflow-hidden">
      <Sidebar user={user} onLogout={logout} />
      <main className="flex-1 flex min-w-0">{children}</main>
    </div>
  );
}
