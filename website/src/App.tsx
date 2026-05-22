import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { TooltipProvider } from '@/components/ui/tooltip';
import { Toaster } from '@/components/ui/sonner';
import { Shell } from '@/components/layout/Shell';
import { Skeleton } from '@/components/ui/skeleton';
import { ROUTES } from '@/config/routes';

// Lazy-loaded pages for code splitting
const DashboardPage = lazy(() => import('@/features/dashboard/page'));
const ExplorerPage = lazy(() => import('@/features/explorer/page'));
const SourcesPage = lazy(() => import('@/features/manager/sources/page'));
const SinksPage = lazy(() => import('@/features/manager/sinks/page'));
const FlowsPage = lazy(() => import('@/features/manager/flows/page'));
const FlowDetailPage = lazy(() => import('@/features/manager/flows/detail'));

/** TanStack Query client with sensible defaults. */
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 3,
      staleTime: 2_000,
      refetchOnWindowFocus: true,
    },
  },
});

/** Page loading fallback. */
function PageLoader() {
  return (
    <div className="space-y-4 p-6">
      <Skeleton className="h-8 w-48" />
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-[120px]" />
        ))}
      </div>
      <Skeleton className="h-[200px]" />
    </div>
  );
}

/** Root App component — providers + routing. */
export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <BrowserRouter>
          <Routes>
            <Route element={<Shell />}>
              <Route
                index
                element={
                  <Suspense fallback={<PageLoader />}>
                    <DashboardPage />
                  </Suspense>
                }
              />
              <Route
                path={ROUTES.EXPLORER}
                element={
                  <Suspense fallback={<PageLoader />}>
                    <ExplorerPage />
                  </Suspense>
                }
              />
              <Route path="/manager" element={<Navigate to={ROUTES.MANAGER_SOURCES} replace />} />
              <Route
                path={ROUTES.MANAGER_SOURCES}
                element={
                  <Suspense fallback={<PageLoader />}>
                    <SourcesPage />
                  </Suspense>
                }
              />
              <Route
                path={ROUTES.MANAGER_SINKS}
                element={
                  <Suspense fallback={<PageLoader />}>
                    <SinksPage />
                  </Suspense>
                }
              />
              <Route
                path={ROUTES.MANAGER_FLOWS}
                element={
                  <Suspense fallback={<PageLoader />}>
                    <FlowsPage />
                  </Suspense>
                }
              />
              <Route
                path={ROUTES.MANAGER_FLOW_DETAIL}
                element={
                  <Suspense fallback={<PageLoader />}>
                    <FlowDetailPage />
                  </Suspense>
                }
              />
            </Route>
          </Routes>
        </BrowserRouter>
        <Toaster />
      </TooltipProvider>
    </QueryClientProvider>
  );
}
