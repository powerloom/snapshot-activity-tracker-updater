import axios from 'axios';
import type {
  HealthResponse,
  DashboardSummary,
  EpochsList,
  EpochDetail,
  ValidatorsList,
  ValidatorDetail,
  SlotsList,
  SlotDetail,
  ProjectsList,
  Timeline,
} from './types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Health check
export const getHealth = async (): Promise<HealthResponse> => {
  const response = await apiClient.get<HealthResponse>('/health');
  return response.data;
};

// Dashboard summary
export const getDashboardSummary = async (): Promise<DashboardSummary> => {
  const response = await apiClient.get<DashboardSummary>('/dashboard/summary');
  return response.data;
};

// Epochs (?offset=&limit=; limit 0 = all retained epochs up to server max page size)
export const getEpochs = async (params?: {
  offset?: number;
  limit?: number;
}): Promise<EpochsList> => {
  const search = new URLSearchParams();
  if (params?.offset != null) search.set('offset', String(params.offset));
  if (params?.limit != null) search.set('limit', String(params.limit));
  const q = search.toString();
  const response = await apiClient.get<EpochsList>(`/epochs${q ? `?${q}` : ''}`);
  return response.data;
};

export const getEpochDetail = async (epochId: number): Promise<EpochDetail> => {
  const response = await apiClient.get<EpochDetail>(`/epochs/${epochId}`);
  return response.data;
};

// Validators
export const getValidators = async (): Promise<ValidatorsList> => {
  const response = await apiClient.get<ValidatorsList>('/validators');
  return response.data;
};

export const getValidatorDetail = async (validatorId: string): Promise<ValidatorDetail> => {
  const response = await apiClient.get<ValidatorDetail>(`/validators/${validatorId}`);
  return response.data;
};

// Slots
export const getSlots = async (): Promise<SlotsList> => {
  const response = await apiClient.get<SlotsList>('/slots');
  return response.data;
};

export const getSlotDetail = async (slotId: string): Promise<SlotDetail> => {
  const response = await apiClient.get<SlotDetail>(`/slots/${slotId}`);
  return response.data;
};

// Projects
export const getProjects = async (): Promise<ProjectsList> => {
  const response = await apiClient.get<ProjectsList>('/projects');
  return response.data;
};

// Timeline (?offset=&limit=)
export const getTimeline = async (params?: {
  offset?: number;
  limit?: number;
}): Promise<Timeline> => {
  const search = new URLSearchParams();
  if (params?.offset != null) search.set('offset', String(params.offset));
  if (params?.limit != null) search.set('limit', String(params.limit));
  const q = search.toString();
  const response = await apiClient.get<Timeline>(`/timeline${q ? `?${q}` : ''}`);
  return response.data;
};

export default apiClient;
