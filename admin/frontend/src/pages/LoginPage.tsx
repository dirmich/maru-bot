import { useAuth } from '../context/AuthContext';
import { ShieldCheck } from 'lucide-react';

export default function LoginPage() {
  const { login } = useAuth();

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-3xl shadow-xl p-10 border">
        <div className="flex flex-col items-center text-center mb-10">
          <div className="w-16 h-16 bg-blue-600 rounded-2xl flex items-center justify-center text-white mb-6 shadow-lg shadow-blue-200">
            <ShieldCheck className="w-10 h-10" />
          </div>
          <h1 className="text-3xl font-extrabold tracking-tight">Maru Admin</h1>
          <p className="text-gray-500 mt-2">Centralized management for Marubot instances</p>
        </div>

        <button
          onClick={login}
          className="w-full flex items-center justify-center gap-3 bg-white border-2 border-gray-100 py-4 px-6 rounded-2xl font-bold text-gray-700 hover:bg-gray-50 hover:border-blue-100 transition-all active:scale-95 shadow-sm"
        >
          <img src="https://www.gstatic.com/firebasejs/ui/2.0.0/images/auth/google.svg" className="w-5 h-5" alt="" />
          Continue with Google
        </button>

        <p className="mt-8 text-center text-xs text-gray-400 leading-relaxed px-4">
          By continuing, you agree to access the MaruBot administrative interface.
        </p>
      </div>
      
      <p className="mt-8 text-sm text-gray-500 font-medium">
        &copy; 2026 MaruBot Open Source Project
      </p>
    </div>
  );
}
