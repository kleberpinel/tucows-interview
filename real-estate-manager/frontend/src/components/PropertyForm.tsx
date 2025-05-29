"use client";

import { useState } from 'react';
import { Property, CreatePropertyRequest, UpdatePropertyRequest } from '@/types/property';

interface PropertyFormProps {
  initialData?: Partial<Property>;
  onSubmit: (property: CreatePropertyRequest | UpdatePropertyRequest) => Promise<void>;
  isEditing?: boolean;
}

const PropertyForm: React.FC<PropertyFormProps> = ({ 
  initialData, 
  onSubmit, 
  isEditing = false 
}) => {
  const [formData, setFormData] = useState({
    name: initialData?.name || '',
    location: initialData?.location || '',
    price: initialData?.price?.toString() || '',
    description: initialData?.description || '',
    property_type: initialData?.property_type || '',
    bedrooms: initialData?.bedrooms?.toString() || '',
    bathrooms: initialData?.bathrooms?.toString() || '',
    square_feet: initialData?.square_feet?.toString() || '',
    lot_size: initialData?.lot_size || '',
    year_built: initialData?.year_built?.toString() || ''
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const priceValue = parseFloat(formData.price);
      if (isNaN(priceValue) || priceValue <= 0) {
        throw new Error('Price must be a valid positive number');
      }

      await onSubmit({
        name: formData.name.trim(),
        location: formData.location.trim(),
        price: priceValue,
        description: formData.description.trim(),
        property_type: formData.property_type.trim() || undefined,
        bedrooms: formData.bedrooms ? parseInt(formData.bedrooms) : undefined,
        bathrooms: formData.bathrooms ? parseFloat(formData.bathrooms) : undefined,
        square_feet: formData.square_feet ? parseInt(formData.square_feet) : undefined,
        lot_size: formData.lot_size.trim() || undefined,
        year_built: formData.year_built ? parseInt(formData.year_built) : undefined
      });
    } catch (err: any) {
      setError(err.message || 'Failed to save property');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic Information */}
        <div className="bg-white p-6 rounded-lg border border-gray-200">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Basic Information</h3>
          
          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                Property Name *
              </label>
              <input
                id="name"
                type="text"
                name="name"
                required
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                value={formData.name}
                onChange={handleChange}
                placeholder="Enter property name"
              />
            </div>

            <div>
              <label htmlFor="property_type" className="block text-sm font-medium text-gray-700">
                Property Type
              </label>
              <select
                id="property_type"
                name="property_type"
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                value={formData.property_type}
                onChange={handleChange}
              >
                <option value="">Select type</option>
                <option value="residential">Residential</option>
                <option value="commercial">Commercial</option>
                <option value="condo">Condo</option>
                <option value="townhouse">Townhouse</option>
                <option value="land">Land</option>
                <option value="other">Other</option>
              </select>
            </div>
          </div>

          <div className="mt-4">
            <label htmlFor="location" className="block text-sm font-medium text-gray-700">
              Location *
            </label>
            <input
              id="location"
              type="text"
              name="location"
              required
              className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              value={formData.location}
              onChange={handleChange}
              placeholder="Enter property location"
            />
          </div>

          <div className="mt-4">
            <label htmlFor="price" className="block text-sm font-medium text-gray-700">
              Price ($) *
            </label>
            <input
              id="price"
              type="number"
              name="price"
              required
              step="0.01"
              min="0"
              className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              value={formData.price}
              onChange={handleChange}
              placeholder="Enter price"
            />
          </div>

          <div className="mt-4">
            <label htmlFor="description" className="block text-sm font-medium text-gray-700">
              Description
            </label>
            <textarea
              id="description"
              name="description"
              rows={4}
              className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              value={formData.description}
              onChange={handleChange}
              placeholder="Enter property description (optional)"
            />
          </div>
        </div>

        {/* Property Details */}
        <div className="bg-white p-6 rounded-lg border border-gray-200">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Property Details</h3>
          
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div>
              <label htmlFor="bedrooms" className="block text-sm font-medium text-gray-700">
                Bedrooms
              </label>
              <input
                id="bedrooms"
                type="number"
                name="bedrooms"
                min="0"
                step="1"
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                value={formData.bedrooms}
                onChange={handleChange}
                placeholder="Number of bedrooms"
              />
            </div>

            <div>
              <label htmlFor="bathrooms" className="block text-sm font-medium text-gray-700">
                Bathrooms
              </label>
              <input
                id="bathrooms"
                type="number"
                name="bathrooms"
                min="0"
                step="0.5"
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                value={formData.bathrooms}
                onChange={handleChange}
                placeholder="Number of bathrooms"
              />
            </div>

            <div>
              <label htmlFor="square_feet" className="block text-sm font-medium text-gray-700">
                Square Feet
              </label>
              <input
                id="square_feet"
                type="number"
                name="square_feet"
                min="0"
                step="1"
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                value={formData.square_feet}
                onChange={handleChange}
                placeholder="Square footage"
              />
            </div>

            <div>
              <label htmlFor="lot_size" className="block text-sm font-medium text-gray-700">
                Lot Size
              </label>
              <input
                id="lot_size"
                type="text"
                name="lot_size"
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                value={formData.lot_size}
                onChange={handleChange}
                placeholder="e.g., 0.25 acres, 10,000 sq ft"
              />
            </div>

            <div>
              <label htmlFor="year_built" className="block text-sm font-medium text-gray-700">
                Year Built
              </label>
              <input
                id="year_built"
                type="number"
                name="year_built"
                min="1800"
                max={new Date().getFullYear()}
                step="1"
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                value={formData.year_built}
                onChange={handleChange}
                placeholder="Year built"
              />
            </div>
          </div>
        </div>

        {/* Submit Button */}
        <div className="flex justify-end space-x-4">
          <button
            type="button"
            onClick={() => window.history.back()}
            className="bg-gray-500 hover:bg-gray-600 text-white font-medium py-2 px-6 rounded-lg transition-colors duration-200"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading}
            className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-6 rounded-lg disabled:opacity-50 transition-colors duration-200"
          >
            {loading ? 'Saving...' : (isEditing ? 'Update Property' : 'Create Property')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default PropertyForm;