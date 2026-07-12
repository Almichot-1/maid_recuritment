'use client';

import { Badge } from '@/components/ui/badge';
import type { SelectionProgressSummary } from '@/types';
import { CheckCircle2, Clock, XCircle, AlertCircle } from 'lucide-react';

interface ProgressBadgesProps {
  progress: SelectionProgressSummary | undefined;
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'done':
    case 'approved':
    case 'confirmed':
    case 'arrived':
      return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
    case 'in_progress':
    case 'booked':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200';
    case 'failed':
    case 'rejected':
      return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
    case 'pending':
    case 'not_arrived':
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200';
  }
}

function getStatusIcon(status: string) {
  switch (status) {
    case 'done':
    case 'approved':
    case 'confirmed':
    case 'arrived':
      return <CheckCircle2 className="h-3 w-3" />;
    case 'in_progress':
    case 'booked':
      return <Clock className="h-3 w-3" />;
    case 'failed':
    case 'rejected':
      return <XCircle className="h-3 w-3" />;
    case 'pending':
    case 'not_arrived':
    default:
      return <AlertCircle className="h-3 w-3" />;
  }
}

function formatStatus(status: string): string {
  return status
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

export function ProgressBadges({ progress }: ProgressBadgesProps) {
  if (!progress) {
    return (
      <div className="text-sm text-muted-foreground">
        No progress tracking yet
      </div>
    );
  }

  const badges = [
    { label: 'COC', status: progress.coc_status },
    { label: 'Medical', status: progress.medical_status },
    { label: 'Visa', status: progress.visa_status },
    { label: 'Ticket', status: progress.ticket_status },
    { label: 'Arrival', status: progress.arrival_status },
  ];

  return (
    <div className="flex flex-wrap gap-1.5">
      {badges.map((badge) => (
        <Badge
          key={badge.label}
          variant="outline"
          className={`text-xs ${getStatusColor(badge.status)}`}
        >
          <span className="flex items-center gap-1">
            {getStatusIcon(badge.status)}
            <span className="font-medium">{badge.label}:</span>
            <span>{formatStatus(badge.status)}</span>
          </span>
        </Badge>
      ))}
    </div>
  );
}
