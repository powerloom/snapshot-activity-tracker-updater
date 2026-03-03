import { useQuery } from '@tanstack/react-query';
import {
  getDashboardSummary,
  getEpochs,
  getEpochDetail,
  getValidators,
  getValidatorDetail,
  getSlots,
  getSlotDetail,
  getProjects,
  getTimeline,
} from '../api/client';

// Query keys
export const queryKeys = {
  health: ['health'] as const,
  summary: ['summary'] as const,
  epochs: ['epochs'] as const,
  epoch: (id: number) => ['epochs', id] as const,
  validators: ['validators'] as const,
  validator: (id: string) => ['validators', id] as const,
  slots: ['slots'] as const,
  slot: (id: string) => ['slots', id] as const,
  projects: ['projects'] as const,
  timeline: ['timeline'] as const,
};

// Dashboard summary
export const useDashboardSummary = () => {
  return useQuery({
    queryKey: queryKeys.summary,
    queryFn: getDashboardSummary,
    refetchInterval: 30000, // Refetch every 30 seconds
  });
};

// Epochs
export const useEpochs = () => {
  return useQuery({
    queryKey: queryKeys.epochs,
    queryFn: getEpochs,
    refetchInterval: 60000, // Refetch every minute
  });
};

export const useEpochDetail = (epochId: number) => {
  return useQuery({
    queryKey: queryKeys.epoch(epochId),
    queryFn: () => getEpochDetail(epochId),
    enabled: !!epochId,
  });
};

// Validators
export const useValidators = () => {
  return useQuery({
    queryKey: queryKeys.validators,
    queryFn: getValidators,
    refetchInterval: 60000,
  });
};

export const useValidatorDetail = (validatorId: string) => {
  return useQuery({
    queryKey: queryKeys.validator(validatorId),
    queryFn: () => getValidatorDetail(validatorId),
    enabled: !!validatorId,
  });
};

// Slots
export const useSlots = () => {
  return useQuery({
    queryKey: queryKeys.slots,
    queryFn: getSlots,
    refetchInterval: 60000,
  });
};

export const useSlotDetail = (slotId: string) => {
  return useQuery({
    queryKey: queryKeys.slot(slotId),
    queryFn: () => getSlotDetail(slotId),
    enabled: !!slotId,
  });
};

// Projects
export const useProjects = () => {
  return useQuery({
    queryKey: queryKeys.projects,
    queryFn: getProjects,
    refetchInterval: 60000,
  });
};

// Timeline
export const useTimeline = () => {
  return useQuery({
    queryKey: queryKeys.timeline,
    queryFn: getTimeline,
    refetchInterval: 60000,
  });
};
