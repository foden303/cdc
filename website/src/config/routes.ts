/**
 * Route paths — single source of truth for navigation.
 * Used by React Router, Sidebar, and Breadcrumb components.
 */
export const ROUTES = {
  DASHBOARD: '/',
  EXPLORER: '/explorer',
  EXPLORER_TOPICS: '/explorer/topics',
  EXPLORER_TOPIC_DETAIL: '/explorer/topics/:topic',
  EXPLORER_TOPIC_PARTITION: '/explorer/topics/:topic/partitions/:partition',
  EXPLORER_CONSUMERS: '/explorer/consumers',
  EXPLORER_CONSUMER_DETAIL: '/explorer/consumers/:consumer',
  EXPLORER_DLQ: '/explorer/dlq',
  MANAGER: '/manager',
  MANAGER_SOURCES: '/manager/sources',
  MANAGER_SINKS: '/manager/sinks',
  MANAGER_FLOWS: '/manager/flows',
  MANAGER_FLOW_DETAIL: '/manager/flows/:id',
} as const;

/** Navigation items for the sidebar. */
export const NAV_ITEMS = [
  {
    key: 'dashboard',
    labelKey: 'nav.dashboard',
    path: ROUTES.DASHBOARD,
    icon: 'LayoutDashboard' as const,
  },
  {
    key: 'explorer',
    labelKey: 'nav.explorer',
    path: ROUTES.EXPLORER_TOPICS,
    icon: 'Search' as const,
    children: [
      { key: 'topics', labelKey: 'nav.topics', path: ROUTES.EXPLORER_TOPICS },
      { key: 'consumers', labelKey: 'nav.consumers', path: ROUTES.EXPLORER_CONSUMERS },
      { key: 'dlq', labelKey: 'nav.dlq', path: ROUTES.EXPLORER_DLQ },
    ],
  },
  {
    key: 'manager',
    labelKey: 'nav.manager',
    path: ROUTES.MANAGER_SOURCES,
    icon: 'Settings' as const,
    children: [
      { key: 'sources', labelKey: 'nav.sources', path: ROUTES.MANAGER_SOURCES },
      { key: 'sinks', labelKey: 'nav.sinks', path: ROUTES.MANAGER_SINKS },
      { key: 'flows', labelKey: 'nav.flows', path: ROUTES.MANAGER_FLOWS },
    ],
  },
] as const;
