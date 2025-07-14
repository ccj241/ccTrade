import React from 'react';
import { cn } from '@/utils/cn';

interface SkeletonProps {
  className?: string;
  variant?: 'text' | 'circular' | 'rectangular';
  width?: string | number;
  height?: string | number;
}

export const Skeleton: React.FC<SkeletonProps> = ({
  className,
  variant = 'text',
  width,
  height,
}) => {
  const variants = {
    text: 'rounded',
    circular: 'rounded-full',
    rectangular: 'rounded-lg',
  };

  const style: React.CSSProperties = {
    width: width || '100%',
    height: height || (variant === 'text' ? '1em' : '100%'),
  };

  return (
    <div
      className={cn(
        'animate-pulse bg-gray-200 dark:bg-gray-700',
        variants[variant],
        className
      )}
      style={style}
    />
  );
};

export const SkeletonCard: React.FC = () => (
  <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6 space-y-4">
    <div className="flex items-center space-x-4">
      <Skeleton variant="circular" width={48} height={48} />
      <div className="flex-1 space-y-2">
        <Skeleton width="60%" height={20} />
        <Skeleton width="40%" height={16} />
      </div>
    </div>
    <div className="space-y-2">
      <Skeleton height={16} />
      <Skeleton height={16} />
      <Skeleton width="80%" height={16} />
    </div>
  </div>
);

export const SkeletonTable: React.FC<{ rows?: number }> = ({ rows = 5 }) => (
  <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg overflow-hidden">
    <div className="bg-gray-50 dark:bg-gray-700 px-6 py-3">
      <Skeleton height={20} width="30%" />
    </div>
    <div className="divide-y divide-gray-200 dark:divide-gray-700">
      {Array.from({ length: rows }).map((_, index) => (
        <div key={index} className="px-6 py-4 flex items-center space-x-4">
          <Skeleton variant="circular" width={40} height={40} />
          <div className="flex-1 space-y-2">
            <Skeleton width="60%" height={16} />
            <Skeleton width="40%" height={14} />
          </div>
          <Skeleton width={80} height={32} />
        </div>
      ))}
    </div>
  </div>
);