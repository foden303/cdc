import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { TopBar } from './TopBar';
import { useSidebarStore } from '@/stores/sidebar';
import { cn } from '@/lib/utils';

/** Main app shell — sidebar + topbar + routed content area. */
export function Shell() {
  const collapsed = useSidebarStore((s) => s.collapsed);

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      <Sidebar />
      <div
        className={cn(
          'flex flex-1 flex-col transition-all duration-300',
          collapsed ? 'ml-16' : 'ml-64',
        )}
      >
        <TopBar />
        <main className="flex-1 overflow-y-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
