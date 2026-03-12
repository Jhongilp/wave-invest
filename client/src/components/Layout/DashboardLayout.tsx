import type { ReactNode } from 'react';

interface DashboardLayoutProps {
  sidebar: ReactNode;
  children: ReactNode;
}

export function DashboardLayout({ sidebar, children }: DashboardLayoutProps) {
  return (
    <div className="min-h-screen bg-gray-950 text-white flex">
      {/* Sidebar */}
      <aside className="w-80 bg-gray-900 border-r border-gray-800 flex flex-col">
        {/* Logo */}
        <div className="p-4 border-b border-gray-800">
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <span className="text-2xl">🌊</span>
            Wave Invest
          </h1>
        </div>

        {/* Sidebar Content */}
        <div className="flex-1 overflow-auto">
          {sidebar}
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-gray-800 text-xs text-gray-500">
          <p>Swing Trading Assistant</p>
          <p>Powered by Gemini AI</p>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col overflow-hidden">
        {children}
      </main>
    </div>
  );
}
