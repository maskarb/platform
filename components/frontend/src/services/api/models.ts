/**
 * Models API
 * Fetches available LLM models, filtered by feature flags
 */

import { apiClient } from './client';

export type Model = {
  value: string;
  label: string;
  description?: string;
  default?: boolean;
};

type ModelsResponse = {
  models: Model[];
};

/**
 * Get available models for a project
 * Models may be filtered by feature flags on the backend
 */
export async function listAvailableModels(projectName: string): Promise<Model[]> {
  const response = await apiClient.get<ModelsResponse>(
    `/projects/${projectName}/models`
  );
  return response.models || [];
}
