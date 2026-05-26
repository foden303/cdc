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
const ExplorerTopicsPage = lazy(() => import('@/features/explorer/topics/page'));
const ExplorerTopicDetailPage = lazy(() => import('@/features/explorer/topics/detail'));
const ExplorerConsumersPage = lazy(() => import('@/features/explorer/consumers/page'));
const ExplorerMessagesPage = lazy(() => import('@/features/explorer/messages/page'));
const ExplorerDLQPage = lazy(() => import('@/features/explorer/dlq/page'));
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
                element={<Navigate to={ROUTES.EXPLORER_TOPICS} replace />}
              />
              <Route
                path={ROUTES.EXPLORER_TOPICS}
                element={
                  <Suspense fallback={<PageLoader />}>
                    <ExplorerTopicsPage />
                  </Suspense>
                }
              />
              <Route
                path={ROUTES.EXPLORER_TOPIC_DETAIL}
                element={
                  <Suspense fallback={<PageLoader />}>
                    <ExplorerTopicDetailPage />
                  </Suspense>
                }
              />
              <Route
                path={ROUTES.EXPLORER_TOPIC_PARTITION}
                element={
                  <Suspense fallback={<PageLoader />}>
                    <ExplorerMessagesPage />
                  </Suspense>
                }
              />
              <Route
                path={ROUTES.EXPLORER_CONSUMERS}
                element={
                  <Suspense fallback={<PageLoader />}>
                    <ExplorerConsumersPage />
                  </Suspense>
                }
              />
              <Route
                path={ROUTES.EXPLORER_DLQ}
                element={
                  <Suspense fallback={<PageLoader />}>
                    <ExplorerDLQPage />
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
