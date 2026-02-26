// TypeScript types matching Go API responses

export interface HealthResponse {
  status: string;
  timestamp: string;
}

export interface DashboardSummary {
  total_epochs: number;
  total_validators: number;
  total_slots: number;
  active_validators: number;
  current_day: string;
}

export interface NetworkTopology {
  nodes: TopologyNode[];
  links: TopologyLink[];
}

export interface TopologyNode {
  id: string;
  type: 'validator' | 'slot' | 'project';
  label: string;
  properties?: Record<string, unknown>;
  x?: number;
  y?: number;
  fx?: number | null;
  fy?: number | null;
}

export interface TopologyLink {
  source: string;
  target: string;
  type: 'votes_for' | 'submits_to' | 'validates';
  properties?: Record<string, unknown>;
}

export interface EpochsList {
  epochs: EpochSummary[];
}

export interface EpochSummary {
  epoch_id: number;
  timestamp: number;
  total_validators: number;
  eligible_nodes_count: number;
  slot_count: number;
  aggregated_projects: number;
}

export interface EpochDetail {
  epoch_id: number;
  data_market: string;
  timestamp: number;
  total_validators: number;
  eligible_nodes_count: number;
  submission_counts: Record<string, number>;
  aggregated_projects: Record<string, string>;
  validator_batches?: ValidatorBatch[];
}

export interface ValidatorBatch {
  validator_id: string;
  batch_cid: string;
}

export interface ValidatorsList {
  validators: ValidatorSummary[];
}

export interface ValidatorSummary {
  validator_id: string;
  total_epochs: number;
  total_batches: number;
  last_active: number;
  recent_epochs?: number[];
}

export interface ValidatorDetail {
  validator_id: string;
  total_epochs: number;
  batches_by_epoch: Record<number, string>;
  projects: string[];
  last_active: number;
}

export interface SlotsList {
  slots: SlotSummary[];
}

export interface SlotSummary {
  slot_id: string;
  total_days: number;
  total_submits: number;
  eligible_count: number;
  last_active: number;
}

export interface SlotDetail {
  slot_id: string;
  submissions_by_day: Record<string, number>;
  eligible_by_day: Record<string, number>;
  total_submits: number;
  total_eligible: number;
}

export interface ProjectsList {
  projects: ProjectSummary[];
}

export interface ProjectSummary {
  project_id: string;
  vote_count: number;
  epochs: number;
  last_cid?: string;
}

export interface Timeline {
  events: TimelineEvent[];
}

export interface TimelineEvent {
  type: string;
  timestamp: number;
  data?: Record<string, unknown>;
}
