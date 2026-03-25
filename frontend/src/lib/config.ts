export const config = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? '/api',
  appTitle: import.meta.env.VITE_APP_TITLE ?? 'UNM Platform',
  environment: import.meta.env.VITE_ENVIRONMENT ?? 'local',
  isProduction: import.meta.env.VITE_ENVIRONMENT === 'production',
} as const;
