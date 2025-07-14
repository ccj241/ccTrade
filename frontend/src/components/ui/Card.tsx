import React from 'react';
import { motion } from 'framer-motion';
import { cn } from '@/utils/cn';

interface CardProps {
  children: React.ReactNode;
  className?: string;
  hover?: boolean;
  glass?: boolean;
}

export const Card: React.FC<CardProps> = ({ 
  children, 
  className,
  hover = false,
  glass = false,
}) => {
  const baseClasses = glass
    ? 'backdrop-blur-md bg-white/10 dark:bg-black/10 border border-white/20 dark:border-white/10'
    : 'bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700';
    
  const Component = hover ? motion.div : 'div';
  const hoverProps = hover ? {
    whileHover: { y: -4, transition: { duration: 0.2 } },
    className: cn(
      baseClasses,
      'rounded-xl shadow-lg transition-shadow duration-300 hover:shadow-2xl',
      className
    )
  } : {
    className: cn(baseClasses, 'rounded-xl shadow-lg', className)
  };

  return (
    <Component {...hoverProps}>
      {children}
    </Component>
  );
};

export const CardHeader: React.FC<{ children: React.ReactNode; className?: string }> = ({ 
  children, 
  className 
}) => (
  <div className={cn('px-6 py-4 border-b border-gray-200 dark:border-gray-700', className)}>
    {children}
  </div>
);

export const CardContent: React.FC<{ children: React.ReactNode; className?: string }> = ({ 
  children, 
  className 
}) => (
  <div className={cn('px-6 py-4', className)}>
    {children}
  </div>
);

export const CardFooter: React.FC<{ children: React.ReactNode; className?: string }> = ({ 
  children, 
  className 
}) => (
  <div className={cn('px-6 py-4 border-t border-gray-200 dark:border-gray-700', className)}>
    {children}
  </div>
);