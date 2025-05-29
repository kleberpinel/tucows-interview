export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

export const config = {
  apiUrl: API_BASE_URL,
} as const;