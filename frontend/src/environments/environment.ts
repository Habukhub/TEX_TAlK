type EnvName = 'local' | 'prod';

interface EnvironmentConfig {
  production: boolean;
  env: EnvName;
  apiUrl: string;
}

const configs: Record<EnvName, EnvironmentConfig> = {
  local: {
    production: false,
    env: 'local',
    apiUrl: 'http://localhost:8080/api/v1',
  },
  prod: {
    production: true,
    env: 'prod',
    apiUrl: '__FRONTEND_API_URL__',
  },
};

function detectEnv(): EnvName {
  if (typeof window === 'undefined') {
    return 'prod';
  }

  const hostname = window.location.hostname;

  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return 'local';
  }

  return 'prod';
}

export const environment: EnvironmentConfig = configs[detectEnv()];
