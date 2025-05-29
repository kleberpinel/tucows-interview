"use client";

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Property } from '@/types/property';
import { fetchProperties } from '@/lib/api';

interface PropertyListProps {
  properties: Property[] | null;
  onEdit: (id: number) => void;
  onDelete: (id: number) => void;
  loading?: boolean;
  error?: string | null;
}

const formatPrice = (price: number): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(price);
};

const PropertyList: React.FC<PropertyListProps> = ({ 
  properties, 
  onEdit, 
  onDelete, 
  loading = false,
  error = null
}) => {
  const router = useRouter();

  const handleViewProperty = (id: number) => {
    if (id) {
      router.push(`/properties/${id}`);
    }
  };
  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
        <span className="ml-2 text-gray-600">Loading properties...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
        <strong className="font-bold">Error: </strong>
        <span className="block sm:inline">{error}</span>
      </div>
    );
  }

  // Handle null or undefined properties
  if (!properties || !Array.isArray(properties)) {
    return (
      <div className="text-center py-12 text-gray-500">
        <p className="text-lg">No properties data available.</p>
      </div>
    );
  }

  if (properties.length === 0) {
    return (
      <div className="text-center py-12 text-gray-500">
        <svg className="mx-auto h-24 w-24 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
        </svg>
        <h3 className="mt-4 text-lg font-medium text-gray-900">No properties found</h3>
        <p className="mt-2 text-sm text-gray-500">
          Get started by creating your first property listing.
        </p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {properties.map((property) => {
        // Safety check for each property
        if (!property || typeof property !== 'object') {
          return null;
        }

        return (
          <div 
            key={property.id || Math.random()} 
            className="bg-white shadow-md rounded-lg overflow-hidden hover:shadow-lg transition-shadow duration-200 cursor-pointer"
            onClick={() => property.id && handleViewProperty(property.id)}
          >
            {/* Property Image */}
            {property.photos && property.photos.length > 0 ? (
              <div className="h-48 overflow-hidden relative">
                <img
                  src={property.photos[0].url}
                  alt={property.photos[0].caption || property.name || 'Property image'}
                  className="w-full h-full object-cover hover:scale-105 transition-transform duration-200"
                />
                {property.photos.length > 1 && (
                  <div className="absolute top-2 right-2 bg-black bg-opacity-70 text-white text-xs px-2 py-1 rounded">
                    +{property.photos.length - 1} more
                  </div>
                )}
              </div>
            ) : (
              <div className="h-48 bg-gray-200 flex items-center justify-center">
                <svg className="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>
            )}

            <div className="p-6">
              <h3 className="text-xl font-semibold text-gray-900 mb-2 truncate">
                {property.name || 'Unnamed Property'}
              </h3>
              
              <div className="flex items-center text-gray-600 mb-3">
                <svg className="w-4 h-4 mr-2 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" clipRule="evenodd" />
                </svg>
                <span className="text-sm truncate">{property.location || 'Unknown Location'}</span>
              </div>
              
              <div className="mb-4">
                <span className="text-2xl font-bold text-green-600">
                  {property.price ? formatPrice(Number(property.price)) : '$0'}
                </span>
              </div>

              {/* Property Details */}
              <div className="flex flex-wrap gap-2 mb-3">
                {property.bedrooms && (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                    {property.bedrooms} bed{property.bedrooms > 1 ? 's' : ''}
                  </span>
                )}
                {property.bathrooms && (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                    {property.bathrooms} bath{property.bathrooms > 1 ? 's' : ''}
                  </span>
                )}
                {property.square_feet && (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                    {property.square_feet.toLocaleString()} sq ft
                  </span>
                )}
                {property.property_type && (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                    {property.property_type}
                  </span>
                )}
              </div>
              
              {property.description && (
                <p className="text-gray-700 text-sm mb-4 line-clamp-3">
                  {property.description}
                </p>
              )}
              
              <div className="flex space-x-2 mt-4">
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    property.id && onEdit(property.id);
                  }}
                  className="flex-1 bg-yellow-500 hover:bg-yellow-600 text-white font-medium py-2 px-4 rounded text-sm transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-yellow-500 focus:ring-offset-2"
                  aria-label={`Edit ${property.name || 'property'}`}
                  disabled={!property.id}
                >
                  Edit
                </button>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    property.id && onDelete(property.id);
                  }}
                  className="flex-1 bg-red-500 hover:bg-red-600 text-white font-medium py-2 px-4 rounded text-sm transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
                  aria-label={`Delete ${property.name || 'property'}`}
                  disabled={!property.id}
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};

export default PropertyList;