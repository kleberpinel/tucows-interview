export interface Photo {
  url: string;
  local_url?: string;
  caption?: string;
}

export interface Property {
  id: number;
  name: string;
  location: string;
  price: number;
  description: string;
  photos?: Photo[];
  external_id?: string;
  mls_number?: string;
  property_type?: string;
  bedrooms?: number;
  bathrooms?: number;
  square_feet?: number;
  lot_size?: string;
  year_built?: number;
  created_at?: string;
  updated_at?: string;
}

export interface CreatePropertyRequest {
  name: string;
  location: string;
  price: number;
  description: string;
  photos?: Photo[];
  property_type?: string;
  bedrooms?: number;
  bathrooms?: number;
  square_feet?: number;
  lot_size?: string;
  year_built?: number;
}

export interface UpdatePropertyRequest {
  name?: string;
  location?: string;
  price?: number;
  description?: string;
  photos?: Photo[];
  property_type?: string;
  bedrooms?: number;
  bathrooms?: number;
  square_feet?: number;
  lot_size?: string;
  year_built?: number;
}

// SimplyRETS API related types
export interface ProcessingJobResponse {
  job_id: string;
  message: string;
  limit: number;
  started_at: string;
}

export interface ProcessingStatus {
  id?: number;
  status: 'running' | 'completed' | 'failed' | 'cancelled';
  total_properties: number;
  processed_count: number;
  failed_count: number;
  started_at: string;
  completed_at?: string;
  error_message?: string;
}