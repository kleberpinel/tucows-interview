"use client";

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { getPropertyById, updateProperty } from '@/lib/api';
import { Property, UpdatePropertyRequest } from '@/types/property';
import PropertyForm from '@/components/PropertyForm';
import AuthGuard from '@/components/AuthGuard';

function EditPropertyPageContent() {
  const router = useRouter();
  const params = useParams();
  const id = params.id as string;
  const [property, setProperty] = useState<Property | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (id) {
      const fetchProperty = async () => {
        try {
          const data = await getPropertyById(id);
          setProperty(data);
        } catch (err: any) {
          setError('Failed to fetch property');
        } finally {
          setLoading(false);
        }
      };
      fetchProperty();
    }
  }, [id]);

  const handleUpdate = async (updatedProperty: UpdatePropertyRequest) => {
    await updateProperty(id, updatedProperty);
    router.push('/properties');
  };

  if (loading) return <div className="text-center py-8">Loading...</div>;
  if (error) return <div className="text-center py-8 text-red-500">{error}</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center mb-8">
          <button
            onClick={() => router.back()}
            className="mr-4 text-gray-600 hover:text-gray-800"
          >
            ← Back
          </button>
          <h1 className="text-3xl font-bold text-gray-900">Edit Property</h1>
        </div>
        
        {property && (
          <PropertyForm
            initialData={property}
            onSubmit={handleUpdate}
            isEditing={true}
          />
        )}
      </div>
    </div>
  );
}

export default function EditPropertyPage() {
  return (
    <AuthGuard>
      <EditPropertyPageContent />
    </AuthGuard>
  );
}