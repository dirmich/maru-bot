import React, { useState } from 'react';
import { 
  Users, 
  Monitor, 
  BarChart3, 
  Settings, 
  LogOut, 
  LayoutDashboard, 
  Cpu, 
  HardDrive, 
  Globe 
} from 'lucide-react';
import { 
  BarChart, 
  Bar, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  PieChart, 
  Pie, 
  Cell 
} from 'recharts';

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444'];

export default function App() {
  const [activeTab, setActiveTab] = useState('dashboard');

  const stats = [
    { title: 'Total Users', value: '124', icon: <Users className="w-6 h-6" />, color: 'bg-blue-500' },
    { title: 'Active Installs', value: '86', icon: <Monitor className="w-6 h-6" />, color: 'bg-green-500' },
    { title: 'Windows', value: '52', icon: <Cpu className="w-6 h-6" />, color: 'bg-purple-500' },
    { title: 'macOS/Linux', value: '34', icon: <Globe className="w-6 h-6" />, color: 'bg-orange-500' },
  ];

  const platData = [
    { name: 'Windows', value: 52 },
    { name: 'macOS', value: 20 },
    { name: 'Linux', value: 14 },
  ];

  return (
    <div className="flex h-screen bg-gray-50 text-gray-900 font-sans">
      {/* Sidebar */}
      <aside className="w-64 bg-white border-r flex flex-col">
        <div className="p-6 flex items-center gap-3">
          <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold">M</div>
          <span className="text-xl font-bold tracking-tight">Maru Admin</span>
        </div>

        <nav className="flex-1 px-4 space-y-2 mt-4">
          <button 
            onClick={() => setActiveTab('dashboard')}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${activeTab === 'dashboard' ? 'bg-blue-50 text-blue-600 font-semibold' : 'text-gray-500 hover:bg-gray-100'}`}
          >
            <LayoutDashboard className="w-5 h-5" /> Dashboard
          </button>
          <button 
            onClick={() => setActiveTab('users')}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${activeTab === 'users' ? 'bg-blue-50 text-blue-600 font-semibold' : 'text-gray-500 hover:bg-gray-100'}`}
          >
            <Users className="w-5 h-5" /> Users
          </button>
          <button 
            onClick={() => setActiveTab('installs')}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${activeTab === 'installs' ? 'bg-blue-50 text-blue-600 font-semibold' : 'text-gray-500 hover:bg-gray-100'}`}
          >
            <BarChart3 className="w-5 h-5" /> Statistics
          </button>
        </nav>

        <div className="p-4 border-t">
          <button className="w-full flex items-center gap-3 px-4 py-3 text-red-500 hover:bg-red-50 rounded-xl transition-all">
            <LogOut className="w-5 h-5" /> Logout
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-y-auto p-8">
        <header className="flex justify-between items-center mb-10">
          <div>
            <h1 className="text-3xl font-extrabold tracking-tight">System Overview</h1>
            <p className="text-gray-500 mt-1">Real-time status of Marubot instances worldwide.</p>
          </div>
          <div className="flex items-center gap-4">
            <div className="text-right">
              <p className="text-sm font-semibold">Shin David (Super)</p>
              <p className="text-xs text-gray-400">oldtv.cf@gmail.com</p>
            </div>
            <div className="w-10 h-10 bg-gray-200 rounded-full border-2 border-blue-500 overflow-hidden">
               <img src="https://api.dicebear.com/7.x/avataaars/svg?seed=Shin" alt="avatar" />
            </div>
          </div>
        </header>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-10">
          {stats.map((s, i) => (
            <div key={i} className="bg-white p-6 rounded-3xl border shadow-sm hover:shadow-md transition-all">
              <div className="flex items-center justify-between mb-4">
                <div className={`${s.color} p-2 rounded-xl text-white shadow-lg`}>
                  {s.icon}
                </div>
                <span className="text-xs font-bold text-green-500 bg-green-50 px-2 py-1 rounded-full">+12%</span>
              </div>
              <p className="text-gray-500 text-sm font-medium">{s.title}</p>
              <p className="text-3xl font-bold mt-1 tracking-tight">{s.value}</p>
            </div>
          ))}
        </div>

        {/* Charts Row */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-10">
          <div className="bg-white p-8 rounded-3xl border shadow-sm">
            <h3 className="text-lg font-bold mb-6">Installation Trends</h3>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={platData}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} />
                  <XAxis dataKey="name" axisLine={false} tickLine={false} />
                  <YAxis axisLine={false} tickLine={false} />
                  <Tooltip cursor={{ fill: '#f3f4f6' }} contentStyle={{ borderRadius: '12px', border: 'none', boxShadow: '0 10px 15px -3px rgb(0 0 0 / 0.1)' }} />
                  <Bar dataKey="value" fill="#3b82f6" radius={[6, 6, 0, 0]} barSize={40} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="bg-white p-8 rounded-3xl border shadow-sm">
            <h3 className="text-lg font-bold mb-6">Platform Distribution</h3>
            <div className="h-64 flex items-center justify-center">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={platData}
                    innerRadius={60}
                    outerRadius={80}
                    paddingAngle={5}
                    dataKey="value"
                  >
                    {platData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip contentStyle={{ borderRadius: '12px', border: 'none' }} />
                </PieChart>
              </ResponsiveContainer>
              <div className="absolute flex flex-col items-center">
                <span className="text-2xl font-bold">86</span>
                <span className="text-xs text-gray-400">Total</span>
              </div>
            </div>
          </div>
        </div>

        {/* User Table (Placeholder) */}
        <div className="bg-white rounded-3xl border shadow-sm overflow-hidden">
          <div className="p-6 border-b flex justify-between items-center bg-gray-50/50">
            <h3 className="text-lg font-bold">Recent Users</h3>
            <button className="text-blue-600 text-sm font-semibold hover:underline">View all</button>
          </div>
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="text-gray-400 font-medium">
                <th className="px-6 py-4">User</th>
                <th className="px-6 py-4">Status</th>
                <th className="px-6 py-4">System</th>
                <th className="px-6 py-4">Last Active</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {[1, 2, 3].map((_, i) => (
                <tr key={i} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4 flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-blue-100"></div>
                    <div>
                      <p className="font-semibold text-gray-900">User {i}</p>
                      <p className="text-xs text-gray-400">user{i}@example.com</p>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="px-2 py-1 bg-green-100 text-green-700 rounded-full text-[10px] font-bold uppercase">Online</span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                       <Cpu className="w-4 h-4 text-gray-400" /> Windows 11
                    </div>
                  </td>
                  <td className="px-6 py-4 text-gray-500">2 mins ago</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </main>
    </div>
  );
}
