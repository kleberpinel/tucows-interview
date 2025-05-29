"use client";

import { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { Property } from '@/types/property';
import { getPropertyById } from '@/lib/api';
import PropertyPhotoGallery from '@/components/PropertyPhotoGallery';
import AuthGuard from '@/components/AuthGuard';

function PropertyDetailPageContent() {
  const [property, setProperty] = useState<Property | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  const params = useParams();

  const propertyId = params.id as string;

  useEffect(() => {
    const loadProperty = async () => {
      try {
        setLoading(true);
        setError(null);
        const data = await getPropertyById(propertyId);
        setProperty(data);
      } catch (err: any) {
        console.error('Error loading property:', err);
        setError(err.message || 'Failed to fetch property');
        if (err.message?.includes('401') || err.message?.includes('Unauthorized')) {
          router.push('/login');
        }
      } finally {
        setLoading(false);
      }
    };

    if (propertyId) {
      loadProperty();
    }
  }, [propertyId, router]);

  const formatPrice = (price: number): string => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(price);
  };

  const formatSquareFeet = (sqft: number): string => {
    return new Intl.NumberFormat('en-US').format(sqft);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
        <span className="ml-2 text-gray-600">Loading property...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4 max-w-md mx-auto">
          <strong className="font-bold">Error: </strong>
          <span className="block sm:inline">{error}</span>
        </div>
        <button
          onClick={() => router.back()}
          className="bg-gray-500 hover:bg-gray-600 text-white font-medium py-2 px-4 rounded-lg"
        >
          Go Back
        </button>
      </div>
    );
  }

  if (!property) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">Property not found</p>
        <button
          onClick={() => router.back()}
          className="mt-4 bg-gray-500 hover:bg-gray-600 text-white font-medium py-2 px-4 rounded-lg"
        >
          Go Back
        </button>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <button
            onClick={() => router.back()}
            className="flex items-center text-gray-600 hover:text-gray-900 transition-colors duration-200"
          >
            <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back to Properties
          </button>
          
          <div className="flex space-x-4">
            <button
              onClick={() => router.push(`/properties/${property.id}/edit`)}
              className="bg-yellow-500 hover:bg-yellow-600 text-white font-medium py-2 px-4 rounded-lg transition-colors duration-200"
            >
              Edit Property
            </button>
          </div>
        </div>

        <div className="bg-white shadow-lg rounded-lg overflow-hidden">
          {/* Photo Gallery */}
          <div className="p-6">
            <PropertyPhotoGallery photos={property.photos || []} propertyName={property.name} />
          </div>

          {/* Property Information */}
          <div className="p-6 border-t">
            <div className="grid md:grid-cols-2 gap-8">
              {/* Left Column */}
              <div>
                <h1 className="text-3xl font-bold text-gray-900 mb-4">
                  {property.name}
                </h1>

                <div className="flex items-center text-gray-600 mb-6">
                  <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" clipRule="evenodd" />
                  </svg>
                  <span className="text-lg">{property.location}</span>
                </div>

                <div className="mb-6">
                  <span className="text-4xl font-bold text-green-600">
                    {formatPrice(property.price)}
                  </span>
                </div>

                {property.description && (
                  <div className="mb-6">
                    <h2 className="text-xl font-semibold text-gray-900 mb-3">Description</h2>
                    <p className="text-gray-700 leading-relaxed">
                      {property.description}
                    </p>
                  </div>
                )}
              </div>

              {/* Right Column */}
              <div>
                <h2 className="text-xl font-semibold text-gray-900 mb-4">Property Details</h2>
                
                <div className="space-y-4">
                  {/* Property Type */}
                  {property.property_type && (
                    <div className="flex justify-between py-2 border-b border-gray-200">
                      <span className="font-medium text-gray-700">Property Type</span>
                      <span className="text-gray-900 capitalize">{property.property_type}</span>
                    </div>
                  )}

                  {/* Bedrooms */}
                  {property.bedrooms && (
                    <div className="flex justify-between py-2 border-b border-gray-200">
                      <span className="font-medium text-gray-700">Bedrooms</span>
                      <span className="text-gray-900">{property.bedrooms}</span>
                    </div>
                  )}

                  {/* Bathrooms */}
                  {property.bathrooms && (
                    <div className="flex justify-between py-2 border-b border-gray-200">
                      <span className="font-medium text-gray-700">Bathrooms</span>
                      <span className="text-gray-900">{property.bathrooms}</span>
                    </div>
                  )}

                  {/* Square Feet */}
                  {property.square_feet && (
                    <div className="flex justify-between py-2 border-b border-gray-200">
                      <span className="font-medium text-gray-700">Square Feet</span>
                      <span className="text-gray-900">{formatSquareFeet(property.square_feet)} sq ft</span>
                    </div>
                  )}

                  {/* Lot Size */}
                  {property.lot_size && (
                    <div className="flex justify-between py-2 border-b border-gray-200">
                      <span className="font-medium text-gray-700">Lot Size</span>
                      <span className="text-gray-900">{property.lot_size}</span>
                    </div>
                  )}

                  {/* Year Built */}
                  {property.year_built && (
                    <div className="flex justify-between py-2 border-b border-gray-200">
                      <span className="font-medium text-gray-700">Year Built</span>
                      <span className="text-gray-900">{property.year_built}</span>
                    </div>
                  )}

                  {/* MLS Number */}
                  {property.mls_number && (
                    <div className="flex justify-between py-2 border-b border-gray-200">
                      <span className="font-medium text-gray-700">MLS Number</span>
                      <span className="text-gray-900">{property.mls_number}</span>
                    </div>
                  )}

                  {/* External ID */}
                  {property.external_id && (
                    <div className="flex justify-between py-2 border-b border-gray-200">
                      <span className="font-medium text-gray-700">External ID</span>
                      <span className="text-gray-900">{property.external_id}</span>
                    </div>
                  )}
                </div>

                {/* Timestamps */}
                <div className="mt-8 pt-6 border-t border-gray-200">
                  <h3 className="text-lg font-semibold text-gray-900 mb-3">Record Information</h3>
                  <div className="space-y-2 text-sm text-gray-600">
                    {property.created_at && (
                      <div>
                        <span className="font-medium">Created:</span>
                        <span className="ml-2">{new Date(property.created_at).toLocaleString()}</span>
                      </div>
                    )}
                    {property.updated_at && (
                      <div>
                        <span className="font-medium">Updated:</span>
                        <span className="ml-2">{new Date(property.updated_at).toLocaleString()}</span>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function PropertyDetailPage() {
  return (
    <AuthGuard>
      <PropertyDetailPageContent />
    </AuthGuard>
  );
}
