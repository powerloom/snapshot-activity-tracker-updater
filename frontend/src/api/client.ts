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

// Epochs
export const getEpochs = async (): Promise<EpochsList> => {
  const response = await apiClient.get<EpochsList>('/epochs');
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

// Timeline
export const getTimeline = async (): Promise<Timeline> => {
  const response = await apiClient.get<Timeline>('/timeline');
  return response.data;
};

export default apiClient;
