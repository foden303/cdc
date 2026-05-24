import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api/client';
import { ENDPOINTS } from '@/lib/api/endpoints';
import { POLLING } from '@/config/constants';
import type {
  ListTopicsResponse,
  ListPartitionsResponse,
  ListMessagesResponse,
  ListDLQMessagesResponse,
  ListConsumersResponse,
  ReprocessDLQResponse,
} from '@/types/api';

/** Query key factory for explorer queries. */
export const explorerKeys = {
  topics: (page?: number) => ['topics', page] as const,
  consumers: (page?: number) => ['consumers', page] as const,
  partitions: (topic: string, page?: number) => ['partitions', topic, page] as const,
  messages: (params: Record<string, unknown>) => ['messages', params] as const,
  dlqMessages: (page?: number) => ['dlqMessages', page] as const,
};

/** Fetches topic list with 10s polling. */
export function useTopics(page = 1, limit = 25) {
  return useQuery({
    queryKey: explorerKeys.topics(page),
    queryFn: () =>
      api.get<ListTopicsResponse>(ENDPOINTS.topics, {
        'pagination.page': page,
        'pagination.limit': limit,
      }),
    refetchInterval: POLLING.TOPICS,
  });
}

/** Fetches partitions for a specific topic with 5s polling. */
export function usePartitions(topic: string, page = 1, limit = 25) {
  return useQuery({
    queryKey: explorerKeys.partitions(topic, page),
    queryFn: () =>
      api.get<ListPartitionsResponse>(ENDPOINTS.partitions, {
        topic,
        'pagination.page': page,
        'pagination.limit': limit,
      }),
    enabled: !!topic,
    refetchInterval: POLLING.PARTITIONS,
  });
}

/** Fetches flow consumers with lag/pending summary. */
export function useConsumers(page = 1, limit = 25) {
  return useQuery({
    queryKey: explorerKeys.consumers(page),
    queryFn: () =>
      api.get<ListConsumersResponse>(ENDPOINTS.consumers, {
        'pagination.page': page,
        'pagination.limit': limit,
      }),
    refetchInterval: POLLING.PARTITIONS,
  });
}

/** Fetches messages with filtering — manual refresh only (no auto-polling). */
export function useMessages(params: {
  status?: number;
  topic?: string;
  partition?: string;
  page?: number;
  limit?: number;
}) {
  const { status, topic, partition, page = 1, limit = 25 } = params;
  return useQuery({
    queryKey: explorerKeys.messages(params),
    queryFn: () =>
      api.get<ListMessagesResponse>(ENDPOINTS.messages, {
        status,
        topic,
        partition,
        'pagination.page': page,
        'pagination.limit': limit,
      }),
    refetchInterval: POLLING.MESSAGES, // 0 = manual only
  });
}

/** Fetches dead-letter queue messages. */
export function useDLQMessages(page = 1, limit = 25) {
  return useQuery({
    queryKey: explorerKeys.dlqMessages(page),
    queryFn: () =>
      api.get<ListDLQMessagesResponse>(ENDPOINTS.dlqMessages, {
        'pagination.page': page,
        'pagination.limit': limit,
      }),
    refetchInterval: POLLING.MESSAGES,
  });
}

/** Mutates to trigger DLQ reprocessing. */
export function useReprocessDLQ() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api.post<ReprocessDLQResponse>(ENDPOINTS.dlqReprocess, {}),
    onSuccess: () => {
      // Invalidate messages since they might get cleared or status updated
      qc.invalidateQueries({ queryKey: ['messages'] });
      qc.invalidateQueries({ queryKey: ['dlqMessages'] });
    },
  });
}
