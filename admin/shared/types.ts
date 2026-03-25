export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url?: string;
  is_superuser: boolean;
  created_at: Date;
}

export interface Instance {
  id: string;
  user_id: string;
  device_name: string;
  os: string;
  memory: number;
  storage: number;
  language: string;
  version: string;
  last_check_in: Date;
  created_at: Date;
}

export interface AdminStats {
  total_users: number;
  total_instances: number;
  platform_stats: {
    windows: number;
    macos: number;
    linux: number;
    other: number;
  };
}
