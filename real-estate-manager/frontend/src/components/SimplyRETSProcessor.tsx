"use client";

import React, { useState, useEffect } from 'react';
import { ProcessingJobResponse, ProcessingStatus } from '@/types/property';
import { startPropertyProcessing, getJobStatus, cancelJob } from '@/lib/api';

interface SimplyRETSProcessorProps {
  onProcessingComplete?: () => void;
}

const SimplyRETSProcessor: React.FC<SimplyRETSProcessorProps> = ({ onProcessingComplete }) => {
  const [isProcessing, setIsProcessing] = useState(false);
  const [currentJob, setCurrentJob] = useState<ProcessingJobResponse | null>(null);
  const [jobStatus, setJobStatus] = useState<ProcessingStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [limit, setLimit] = useState(50);

  useEffect(() => {
    let interval: NodeJS.Timeout;

    if (currentJob && isProcessing) {
      interval = setInterval(async () => {
        try {
          const status = await getJobStatus(currentJob.job_id);
          setJobStatus(status);
          
          if (status.status === 'completed' || status.status === 'failed' || status.status === 'cancelled') {
            setIsProcessing(false);
            if (status.status === 'completed' && onProcessingComplete) {
              onProcessingComplete();
            }
          }
        } catch (err: any) {
          console.error('Error fetching job status:', err);
          setError(err.message);
          setIsProcessing(false);
        }
      }, 2000); // Poll every 2 seconds
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [currentJob, isProcessing, onProcessingComplete]);

  const handleStartProcessing = async () => {
    if (limit <= 0 || limit > 500) {
      setError('Limit must be between 1 and 500');
      return;
    }

    try {
      setError(null);
      setIsProcessing(true);
      
      const response = await startPropertyProcessing(limit);
      setCurrentJob(response);
      setJobStatus({
        status: 'running',
        total_properties: 0,
        processed_count: 0,
        failed_count: 0,
        started_at: new Date().toISOString()
      });
    } catch (err: any) {
      console.error('Error starting processing:', err);
      setError(err.message || 'Failed to start processing');
      setIsProcessing(false);
    }
  };

  const handleCancelProcessing = async () => {
    if (!currentJob) return;

    try {
      await cancelJob(currentJob.job_id);
      setIsProcessing(false);
      setJobStatus(prev => prev ? { ...prev, status: 'cancelled' } : null);
    } catch (err: any) {
      console.error('Error cancelling job:', err);
      setError(err.message || 'Failed to cancel job');
    }
  };

  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'completed':
        return 'text-green-600 bg-green-100';
      case 'failed':
        return 'text-red-600 bg-red-100';
      case 'cancelled':
        return 'text-gray-600 bg-gray-100';
      case 'processing':
        return 'text-blue-600 bg-blue-100';
      default:
        return 'text-yellow-600 bg-yellow-100';
    }
  };

  const getProgressPercentage = () => {
    if (!jobStatus || !jobStatus.total_properties) return 0;
    return Math.round((jobStatus.processed_count / jobStatus.total_properties) * 100);
  };

  return (
    <div className="bg-white shadow-md rounded-lg p-6 mb-6">
      <h2 className="text-xl font-semibold text-gray-900 mb-4">
        SimplyRETS Property Import
      </h2>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      {!isProcessing ? (
        <div className="space-y-4">
          <div>
            <label htmlFor="limit" className="block text-sm font-medium text-gray-700 mb-2">
              Number of properties to import (1-500):
            </label>
            <input
              id="limit"
              type="number"
              min="1"
              max="500"
              value={limit}
              onChange={(e) => setLimit(parseInt(e.target.value) || 0)}
              className="block w-32 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>

          <button
            onClick={handleStartProcessing}
            disabled={isProcessing}
            className="bg-blue-500 hover:bg-blue-600 text-white font-medium py-2 px-4 rounded-lg transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
          >
            Start Import
          </button>

          <p className="text-sm text-gray-600">
            This will fetch property listings from SimplyRETS, download images, and save them to your database.
            Properties are processed in batches of 10 with parallel image downloading.
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium text-gray-900">
              Processing Properties...
            </h3>
            <button
              onClick={handleCancelProcessing}
              className="bg-red-500 hover:bg-red-600 text-white font-medium py-1 px-3 rounded text-sm transition-colors duration-200"
            >
              Cancel
            </button>
          </div>

          {jobStatus && (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(jobStatus.status)}`}>
                  {jobStatus.status.toUpperCase()}
                </span>
                <span className="text-sm text-gray-600">
                  Job ID: {currentJob?.job_id}
                </span>
              </div>

              {jobStatus.total_properties > 0 && (
                <div>
                  <div className="flex justify-between text-sm text-gray-600 mb-1">
                    <span>Progress</span>
                    <span>{jobStatus.processed_count} / {jobStatus.total_properties} properties</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-blue-600 h-2 rounded-full transition-all duration-500"
                      style={{ width: `${getProgressPercentage()}%` }}
                    ></div>
                  </div>
                  <div className="text-center text-sm text-gray-600 mt-1">
                    {getProgressPercentage()}%
                  </div>
                </div>
              )}

              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="font-medium text-gray-700">Processed:</span>
                  <span className="ml-2 text-green-600">{jobStatus.processed_count}</span>
                </div>
                <div>
                  <span className="font-medium text-gray-700">Failed:</span>
                  <span className="ml-2 text-red-600">{jobStatus.failed_count}</span>
                </div>
              </div>

              {jobStatus.error_message && (
                <p className="text-sm text-red-600 italic">
                  {jobStatus.error_message}
                </p>
              )}

              {jobStatus.started_at && (
                <p className="text-xs text-gray-500">
                  Started: {new Date(jobStatus.started_at).toLocaleString()}
                </p>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default SimplyRETSProcessor;
