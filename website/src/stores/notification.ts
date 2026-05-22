import { create } from 'zustand';

export type NotificationType = 'success' | 'error' | 'info' | 'warning';

export interface Notification {
  id: string;
  type: NotificationType;
  title: string;
  message?: string;
  duration?: number;
}

interface NotificationState {
  notifications: Notification[];
  addNotification: (n: Omit<Notification, 'id'>) => void;
  removeNotification: (id: string) => void;
  clearAll: () => void;
}

/** Notification queue for toast messages. Used by API error interceptor. */
export const useNotificationStore = create<NotificationState>((set) => ({
  notifications: [],
  addNotification: (n) => {
    const id = `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;
    set((s) => ({ notifications: [...s.notifications, { ...n, id }] }));

    // Auto-remove after duration (default 5s)
    const duration = n.duration ?? 5000;
    if (duration > 0) {
      setTimeout(() => {
        set((s) => ({
          notifications: s.notifications.filter((item) => item.id !== id),
        }));
      }, duration);
    }
  },
  removeNotification: (id) =>
    set((s) => ({
      notifications: s.notifications.filter((n) => n.id !== id),
    })),
  clearAll: () => set({ notifications: [] }),
}));
