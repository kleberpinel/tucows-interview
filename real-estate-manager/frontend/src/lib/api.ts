import { config } from './config';
import { Property, CreatePropertyRequest, UpdatePropertyRequest, ProcessingJobResponse, ProcessingStatus } from '@/types/property';

const API_BASE_URL = config.apiUrl;

interface LoginResponse {
  token: string;
}

interface RegisterResponse {
  message: string;
}

class ApiClient {
  private baseURL: string;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const token = localStorage.getItem('token');

    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
        ...options.headers,
      },
      ...options,
    };

    const response = await fetch(url, config);

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`HTTP error! status: ${response.status} - ${errorText}`);
    }

    const responseText = await response.text();
    
    // Handle empty responses
    if (!responseText || responseText.trim() === '') {
      return (endpoint.includes('DELETE') ? undefined : []) as T;
    }

    try {
      const parsed = JSON.parse(responseText);
      return parsed;
    } catch (parseError) {
      console.error('Failed to parse response:', parseError);
      throw new Error('Invalid response format from server');
    }
  }

  // Auth methods
  async login(username: string, password: string): Promise<LoginResponse> {
    return this.request<LoginResponse>('/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
  }

  async register(username: string, password: string, email: string): Promise<RegisterResponse> {
    return this.request<RegisterResponse>('/register', {
      method: 'POST',
      body: JSON.stringify({ username, password, email }),
    });
  }

  // Property methods
  async getProperties(): Promise<Property[]> {
    const result = await this.request<Property[]>('/properties');
    return Array.isArray(result) ? result : [];
  }

  async getProperty(id: string): Promise<Property> {
    return this.request<Property>(`/properties/${id}`);
  }

  async createProperty(property: CreatePropertyRequest): Promise<Property> {
    return this.request<Property>('/properties', {
      method: 'POST',
      body: JSON.stringify(property),
    });
  }

  async updateProperty(id: string | string[], property: UpdatePropertyRequest): Promise<Property> {
    const propertyId = Array.isArray(id) ? id[0] : id;
    return this.request<Property>(`/properties/${propertyId}`, {
      method: 'PUT',
      body: JSON.stringify(property),
    });
  }

  async deleteProperty(id: string): Promise<void> {
    return this.request<void>(`/properties/${id}`, {
      method: 'DELETE',
    });
  }

  // SimplyRETS methods
  async startPropertyProcessing(limit: number = 50): Promise<ProcessingJobResponse> {
    return this.request<ProcessingJobResponse>('/simplyrets/process', {
      method: 'POST',
      body: JSON.stringify({ limit }),
    });
  }

  async getJobStatus(jobId: string): Promise<ProcessingStatus> {
    return this.request<ProcessingStatus>(`/simplyrets/jobs/${jobId}/status`);
  }

  async cancelJob(jobId: string): Promise<{ message: string; job_id: string }> {
    return this.request<{ message: string; job_id: string }>(`/simplyrets/jobs/${jobId}`, {
      method: 'DELETE',
    });
  }

  async getSimplyRETSHealth(): Promise<{ status: string; service: string; timestamp: string }> {
    return this.request<{ status: string; service: string; timestamp: string }>('/simplyrets/health');
  }
}

export const apiClient = new ApiClient(API_BASE_URL);

// Export individual functions for backward compatibility
export const login = (username: string, password: string) => 
  apiClient.login(username, password);

export const register = (username: string, password: string, email: string) => 
  apiClient.register(username, password, email);

export const getProperties = () => 
  apiClient.getProperties();

export const fetchProperties = () => 
  apiClient.getProperties();

export const getPropertyById = (id: string | string[]) => 
  apiClient.getProperty(Array.isArray(id) ? id[0] : id);

export const createProperty = (property: CreatePropertyRequest) => 
  apiClient.createProperty(property);

export const updateProperty = (id: string | string[], property: UpdatePropertyRequest) => 
  apiClient.updateProperty(id, property);

export const deleteProperty = (id: string) => 
  apiClient.deleteProperty(id);

// SimplyRETS API functions
export const startPropertyProcessing = (limit: number = 50) => 
  apiClient.request<ProcessingJobResponse>('/simplyrets/process', {
    method: 'POST',
    body: JSON.stringify({ limit }),
  });

export const getJobStatus = (jobId: string) => 
  apiClient.request<ProcessingStatus>(`/simplyrets/jobs/${jobId}/status`);

export const cancelJob = (jobId: string) => 
  apiClient.request<{ message: string; job_id: string }>(`/simplyrets/jobs/${jobId}`, {
    method: 'DELETE',
  });

export const getSimplyRETSHealth = () => 
  apiClient.request<{ status: string; service: string; timestamp: string }>('/simplyrets/health');