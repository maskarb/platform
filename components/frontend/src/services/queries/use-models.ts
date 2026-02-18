/**
 * React Query hooks for Models
 * Fetches available LLM models with feature flag filtering
 */

import { useQuery } from '@tanstack/react-query';
import * as modelsApi from '../api/models';

export const modelKeys = {
  all: ['models'] as const,
  list: (projectName: string) => [...modelKeys.all, 'list', projectName] as const,
};

/**
 * Hook to fetch available models for a project
 * Models are filtered by feature flags on the backend
 */
export function useAvailableModels(projectName: string) {
  return useQuery({
    queryKey: modelKeys.list(projectName),
    queryFn: () => modelsApi.listAvailableModels(projectName),
    enabled: !!projectName,
    staleTime: 60 * 60 * 1000, // 1 hour - models rarely change
    gcTime: 24 * 60 * 60 * 1000, // 24 hours cache
  });
}

// Re-export Model type for convenience
export type { Model } from '../api/models';
