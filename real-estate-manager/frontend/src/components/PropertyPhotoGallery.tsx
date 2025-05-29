"use client";

import React, { useState } from 'react';
import { Photo } from '@/types/property';

interface PropertyPhotoGalleryProps {
  photos: Photo[];
  propertyName?: string;
}

const PropertyPhotoGallery: React.FC<PropertyPhotoGalleryProps> = ({ photos, propertyName }) => {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isFullscreen, setIsFullscreen] = useState(false);

  if (!photos || photos.length === 0) {
    return (
      <div className="h-64 bg-gray-200 flex items-center justify-center rounded-lg">
        <div className="text-center">
          <svg className="w-16 h-16 text-gray-400 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          <p className="text-gray-500">No photos available</p>
        </div>
      </div>
    );
  }

  const nextPhoto = () => {
    setCurrentIndex((prev) => (prev + 1) % photos.length);
  };

  const prevPhoto = () => {
    setCurrentIndex((prev) => (prev - 1 + photos.length) % photos.length);
  };

  const openFullscreen = () => {
    setIsFullscreen(true);
  };

  const closeFullscreen = () => {
    setIsFullscreen(false);
  };

  const currentPhoto = photos[currentIndex];

  return (
    <>
      {/* Main Gallery */}
      <div className="relative">
        {/* Main Image */}
        <div className="relative h-64 md:h-96 overflow-hidden rounded-lg bg-gray-200">
          <img
            src={currentPhoto.url}
            alt={currentPhoto.caption || `${propertyName} - Photo ${currentIndex + 1}`}
            className="w-full h-full object-cover cursor-pointer hover:scale-105 transition-transform duration-200"
            onClick={openFullscreen}
          />
          
          {/* Navigation Arrows */}
          {photos.length > 1 && (
            <>
              <button
                onClick={prevPhoto}
                className="absolute left-2 top-1/2 transform -translate-y-1/2 bg-black bg-opacity-50 hover:bg-opacity-75 text-white p-2 rounded-full transition-all duration-200"
                aria-label="Previous photo"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
              </button>
              
              <button
                onClick={nextPhoto}
                className="absolute right-2 top-1/2 transform -translate-y-1/2 bg-black bg-opacity-50 hover:bg-opacity-75 text-white p-2 rounded-full transition-all duration-200"
                aria-label="Next photo"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </button>
            </>
          )}

          {/* Photo Counter */}
          <div className="absolute top-2 right-2 bg-black bg-opacity-70 text-white text-sm px-3 py-1 rounded-full">
            {currentIndex + 1} / {photos.length}
          </div>

          {/* Fullscreen Button */}
          <button
            onClick={openFullscreen}
            className="absolute bottom-2 right-2 bg-black bg-opacity-50 hover:bg-opacity-75 text-white p-2 rounded-full transition-all duration-200"
            aria-label="View fullscreen"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
            </svg>
          </button>
        </div>

        {/* Thumbnail Strip */}
        {photos.length > 1 && (
          <div className="flex space-x-2 mt-4 overflow-x-auto pb-2">
            {photos.map((photo, index) => (
              <button
                key={index}
                onClick={() => setCurrentIndex(index)}
                className={`flex-shrink-0 relative ${
                  index === currentIndex 
                    ? 'ring-2 ring-blue-500' 
                    : 'ring-1 ring-gray-300 hover:ring-gray-400'
                } rounded-lg overflow-hidden transition-all duration-200`}
              >
                <img
                  src={photo.url}
                  alt={photo.caption || `Thumbnail ${index + 1}`}
                  className="w-16 h-16 object-cover"
                />
                {index === currentIndex && (
                  <div className="absolute inset-0 bg-blue-500 bg-opacity-20"></div>
                )}
              </button>
            ))}
          </div>
        )}

        {/* Caption */}
        {currentPhoto.caption && (
          <p className="mt-2 text-sm text-gray-600 italic">
            {currentPhoto.caption}
          </p>
        )}
      </div>

      {/* Fullscreen Modal */}
      {isFullscreen && (
        <div className="fixed inset-0 z-50 bg-black bg-opacity-90 flex items-center justify-center p-4">
          <div className="relative max-w-7xl max-h-full">
            <img
              src={currentPhoto.url}
              alt={currentPhoto.caption || `${propertyName} - Photo ${currentIndex + 1}`}
              className="max-w-full max-h-full object-contain"
            />
            
            {/* Close Button */}
            <button
              onClick={closeFullscreen}
              className="absolute top-4 right-4 bg-black bg-opacity-50 hover:bg-opacity-75 text-white p-3 rounded-full transition-all duration-200"
              aria-label="Close fullscreen"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>

            {/* Navigation in Fullscreen */}
            {photos.length > 1 && (
              <>
                <button
                  onClick={prevPhoto}
                  className="absolute left-4 top-1/2 transform -translate-y-1/2 bg-black bg-opacity-50 hover:bg-opacity-75 text-white p-3 rounded-full transition-all duration-200"
                  aria-label="Previous photo"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                  </svg>
                </button>
                
                <button
                  onClick={nextPhoto}
                  className="absolute right-4 top-1/2 transform -translate-y-1/2 bg-black bg-opacity-50 hover:bg-opacity-75 text-white p-3 rounded-full transition-all duration-200"
                  aria-label="Next photo"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                </button>
              </>
            )}

            {/* Photo Info in Fullscreen */}
            <div className="absolute bottom-4 left-4 right-4 text-center">
              <div className="bg-black bg-opacity-50 text-white px-4 py-2 rounded-lg inline-block">
                <p className="text-sm">
                  {currentIndex + 1} of {photos.length}
                </p>
                {currentPhoto.caption && (
                  <p className="text-sm mt-1 italic">
                    {currentPhoto.caption}
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default PropertyPhotoGallery;
