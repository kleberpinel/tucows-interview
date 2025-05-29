"use client";

import { useRouter } from 'next/navigation';
import { createProperty } from '@/lib/api';
import { CreatePropertyRequest } from '@/types/property';
import PropertyForm from '@/components/PropertyForm';
import AuthGuard from '@/components/AuthGuard';

function CreatePropertyPageContent() {
  const router = useRouter();

  const handleCreate = async (property: CreatePropertyRequest) => {
    await createProperty(property);
    router.push('/properties');
  };

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
          <h1 className="text-3xl font-bold text-gray-900">Create New Property</h1>
        </div>
        
        <PropertyForm
          onSubmit={handleCreate}
          isEditing={false}
        />
      </div>
    </div>
  );
}

export default function CreatePropertyPage() {
  return (
    <AuthGuard>
      <CreatePropertyPageContent />
    </AuthGuard>
  );
}