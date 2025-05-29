"use client";

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { fetchProperties, deleteProperty } from '@/lib/api';
import { Property } from '@/types/property';
import PropertyList from '@/components/PropertyList';
import SimplyRETSProcessor from '@/components/SimplyRETSProcessor';
import AuthGuard from '@/components/AuthGuard';
import { auth } from '@/lib/auth';

function PropertiesPageContent() {
  const [properties, setProperties] = useState<Property[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  const loadProperties = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await fetchProperties();
      setProperties(Array.isArray(data) ? data : []);
    } catch (err: any) {
      console.error('Error loading properties:', err);
      setError(err.message || 'Failed to fetch properties');
      setProperties([]);
      if (err.message?.includes('401') || err.message?.includes('Unauthorized')) {
        auth.removeToken();
        router.push('/login');
      }
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    loadProperties();
  }, [loadProperties]);

  const handleEdit = (id: number) => {
    if (id) {
      router.push(`/properties/${id}/edit`);
    }
  };

  const handleDelete = async (id: number) => {
    if (!id) return;
    
    if (confirm('Are you sure you want to delete this property?')) {
      try {
        await deleteProperty(id.toString());
        setProperties(prevProperties => 
          prevProperties ? prevProperties.filter(p => p.id !== id) : []
        );
      } catch (err: any) {
        console.error('Error deleting property:', err);
        setError(err.message || 'Failed to delete property');
      }
    }
  };

  const handleLogout = () => {
    auth.removeToken();
    router.push('/login');
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Property Management</h1>
          <div className="flex space-x-4">
            <button
              onClick={() => router.push('/properties/create')}
              className="bg-blue-500 hover:bg-blue-600 text-white font-medium py-2 px-4 rounded-lg transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            >
              Add Property
            </button>
            <button
              onClick={handleLogout}
              className="bg-gray-500 hover:bg-gray-600 text-white font-medium py-2 px-4 rounded-lg transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
            >
              Logout
            </button>
          </div>
        </div>

        <SimplyRETSProcessor onProcessingComplete={loadProperties} />

        <PropertyList
          properties={properties}
          onEdit={handleEdit}
          onDelete={handleDelete}
          loading={loading}
          error={error}
        />
      </div>
    </div>
  );
}

export default function PropertiesPage() {
  return (
    <AuthGuard>
      <PropertiesPageContent />
    </AuthGuard>
  );
}