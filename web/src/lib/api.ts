import axios from 'axios';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface Setting {
  id: number;
  connector_mode: string;
  mcp_server_cmd?: string;
  mcp_server_url?: string;
  mcp_auth?: string;
  note_target: string;
  polling_interval_sec: number;
  llm_base_url?: string;
  llm_api_key?: string;
  llm_model?: string;
  llm_timeout_sec: number;
  intent_threshold: number;
  smtp_host?: string;
  smtp_port?: number;
  smtp_user?: string;
  smtp_pass?: string;
  smtp_from?: string;
  provider_json: string;
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: number;
  comment_id: number;
  status: string;
  request_type: string;
  email?: string;
  prompt?: string;
  confidence?: number;
  provider_name?: string;
  provider_job_id?: string;
  result_object_key?: string;
  result_url?: string;
  error?: string;
  retry_count: number;
  created_at: string;
  updated_at: string;
  comment?: {
    id: number;
    note_target: string;
    comment_uid: string;
    user_name?: string;
    content: string;
    comment_created_at?: string;
    ingested_at: string;
  };
  deliveries?: Array<{
    id: number;
    task_id: number;
    email_to: string;
    status: string;
    sent_at?: string;
    error?: string;
  }>;
}

export interface TasksResponse {
  tasks: Task[];
  limit: number;
  offset: number;
}

const api = axios.create({
  baseURL: `${API_BASE}/api`,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const apiClient = {
  getSettings: async (): Promise<Setting> => {
    const response = await api.get<Setting>('/settings');
    return response.data;
  },

  updateSettings: async (settings: Partial<Setting>): Promise<Setting> => {
    const response = await api.put<Setting>('/settings', settings);
    return response.data;
  },

  runPoll: async (): Promise<{ message: string }> => {
    const response = await api.post<{ message: string }>('/poll/run');
    return response.data;
  },

  listTasks: async (limit: number = 100, offset: number = 0): Promise<TasksResponse> => {
    const response = await api.get<TasksResponse>(`/tasks?limit=${limit}&offset=${offset}`);
    return response.data;
  },

  getTask: async (id: number): Promise<Task> => {
    const response = await api.get<Task>(`/tasks/${id}`);
    return response.data;
  },

  healthCheck: async (): Promise<{ status: string }> => {
    const response = await api.get<{ status: string }>('/healthz');
    return response.data;
  },
};
