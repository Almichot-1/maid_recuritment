'use client';

import { useCurrentUser } from '@/hooks/use-auth';
import { useSelectionProgress, useUpdateProgress, useUploadProgressDocument } from '@/hooks/use-selection-progress';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import { FileText, Download, Loader2 } from 'lucide-react';
import { UserRole, PROGRESS_STATUS, COC_TYPE, VISA_STATUS, TICKET_STATUS, ARRIVAL_STATUS } from '@/types';

interface ProgressTrackingProps {
  selectionId: string;
}

export function ProgressTracking({ selectionId }: ProgressTrackingProps) {
  const { user } = useCurrentUser();
  const { data: progress, isLoading } = useSelectionProgress(selectionId);
  const updateProgress = useUpdateProgress(selectionId);
  const uploadDocument = useUploadProgressDocument(selectionId);
  
  const isEthiopian = user?.role === UserRole.ETHIOPIAN_AGENT;
  const canEdit = isEthiopian;

  const handleUpdate = async (field: string, value: string) => {
    if (!canEdit) return;
    
    try {
      await updateProgress.mutateAsync({ [field]: value });
      toast.success('Progress updated successfully');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to update progress');
    }
  };

  const handleFileUpload = async (documentType: string, file: File) => {
    if (!canEdit) return;
    
    try {
      await uploadDocument.mutateAsync({ documentType, file });
      toast.success('Document uploaded successfully');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to upload document');
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  if (!progress) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-muted-foreground">
          Progress tracking will be available once the selection is approved.
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* COC Section */}
      <Card>
        <CardHeader>
          <CardTitle>COC (Certificate of Competency)</CardTitle>
          <CardDescription>Track COC processing status and type</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="coc_status">Status</Label>
              <Select
                value={progress.coc_status}
                onValueChange={(value) => handleUpdate('coc_status', value)}
                disabled={!canEdit || updateProgress.isPending}
              >
                <SelectTrigger id="coc_status">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={PROGRESS_STATUS.PENDING}>Pending</SelectItem>
                  <SelectItem value={PROGRESS_STATUS.IN_PROGRESS}>In Progress</SelectItem>
                  <SelectItem value={PROGRESS_STATUS.DONE}>Done</SelectItem>
                  <SelectItem value={PROGRESS_STATUS.FAILED}>Failed</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="coc_type">Type</Label>
              <Select
                value={progress.coc_type || ''}
                onValueChange={(value) => handleUpdate('coc_type', value)}
                disabled={!canEdit || updateProgress.isPending}
              >
                <SelectTrigger id="coc_type">
                  <SelectValue placeholder="Select type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">Not set</SelectItem>
                  <SelectItem value={COC_TYPE.ONLINE}>Online</SelectItem>
                  <SelectItem value={COC_TYPE.OFFLINE}>Offline</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label>Document</Label>
            {progress.coc_document ? (
              <div className="flex items-center gap-2">
                <FileText className="h-4 w-4 text-muted-foreground" />
                <a
                  href={progress.coc_document.file_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-blue-600 hover:underline"
                >
                  {progress.coc_document.file_name}
                </a>
                <Download className="h-4 w-4" />
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No document uploaded</p>
            )}
            {canEdit && (
              <Input
                type="file"
                accept=".pdf,.jpg,.jpeg,.png"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) handleFileUpload('coc', file);
                }}
                disabled={uploadDocument.isPending}
                className="mt-2"
              />
            )}
          </div>
        </CardContent>
      </Card>

      {/* Medical Section */}
      <Card>
        <CardHeader>
          <CardTitle>Medical</CardTitle>
          <CardDescription>Track medical examination status</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="medical_status">Status</Label>
            <Select
              value={progress.medical_status}
              onValueChange={(value) => handleUpdate('medical_status', value)}
              disabled={!canEdit || updateProgress.isPending}
            >
              <SelectTrigger id="medical_status">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={PROGRESS_STATUS.PENDING}>Pending</SelectItem>
                <SelectItem value={PROGRESS_STATUS.IN_PROGRESS}>In Progress</SelectItem>
                <SelectItem value={PROGRESS_STATUS.DONE}>Done</SelectItem>
                <SelectItem value={PROGRESS_STATUS.FAILED}>Failed</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>Document</Label>
            {progress.medical_document ? (
              <div className="flex items-center gap-2">
                <FileText className="h-4 w-4 text-muted-foreground" />
                <a
                  href={progress.medical_document.file_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-blue-600 hover:underline"
                >
                  {progress.medical_document.file_name}
                </a>
                <Download className="h-4 w-4" />
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No document uploaded</p>
            )}
            {canEdit && (
              <Input
                type="file"
                accept=".pdf,.jpg,.jpeg,.png"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) handleFileUpload('medical', file);
                }}
                disabled={uploadDocument.isPending}
                className="mt-2"
              />
            )}
          </div>
        </CardContent>
      </Card>

      {/* Visa Section */}
      <Card>
        <CardHeader>
          <CardTitle>Visa</CardTitle>
          <CardDescription>Track visa application status</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="visa_status">Status</Label>
            <Select
              value={progress.visa_status}
              onValueChange={(value) => handleUpdate('visa_status', value)}
              disabled={!canEdit || updateProgress.isPending}
            >
              <SelectTrigger id="visa_status">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={VISA_STATUS.PENDING}>Pending</SelectItem>
                <SelectItem value={VISA_STATUS.IN_PROGRESS}>In Progress</SelectItem>
                <SelectItem value={VISA_STATUS.APPROVED}>Approved</SelectItem>
                <SelectItem value={VISA_STATUS.REJECTED}>Rejected</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>Document</Label>
            {progress.visa_document ? (
              <div className="flex items-center gap-2">
                <FileText className="h-4 w-4 text-muted-foreground" />
                <a
                  href={progress.visa_document.file_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-blue-600 hover:underline"
                >
                  {progress.visa_document.file_name}
                </a>
                <Download className="h-4 w-4" />
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No document uploaded</p>
            )}
            {canEdit && (
              <Input
                type="file"
                accept=".pdf,.jpg,.jpeg,.png"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) handleFileUpload('visa', file);
                }}
                disabled={uploadDocument.isPending}
                className="mt-2"
              />
            )}
          </div>
        </CardContent>
      </Card>

      {/* Ticket Section */}
      <Card>
        <CardHeader>
          <CardTitle>Ticket</CardTitle>
          <CardDescription>Track ticket booking status</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="ticket_status">Status</Label>
            <Select
              value={progress.ticket_status}
              onValueChange={(value) => handleUpdate('ticket_status', value)}
              disabled={!canEdit || updateProgress.isPending}
            >
              <SelectTrigger id="ticket_status">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={TICKET_STATUS.PENDING}>Pending</SelectItem>
                <SelectItem value={TICKET_STATUS.BOOKED}>Booked</SelectItem>
                <SelectItem value={TICKET_STATUS.CONFIRMED}>Confirmed</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>Document</Label>
            {progress.ticket_document ? (
              <div className="flex items-center gap-2">
                <FileText className="h-4 w-4 text-muted-foreground" />
                <a
                  href={progress.ticket_document.file_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-blue-600 hover:underline"
                >
                  {progress.ticket_document.file_name}
                </a>
                <Download className="h-4 w-4" />
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No document uploaded</p>
            )}
            {canEdit && (
              <Input
                type="file"
                accept=".pdf,.jpg,.jpeg,.png"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) handleFileUpload('ticket', file);
                }}
                disabled={uploadDocument.isPending}
                className="mt-2"
              />
            )}
          </div>
        </CardContent>
      </Card>

      {/* Arrival Section */}
      <Card>
        <CardHeader>
          <CardTitle>Arrival</CardTitle>
          <CardDescription>Track arrival status and details</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="arrival_status">Status</Label>
            <Select
              value={progress.arrival_status}
              onValueChange={(value) => handleUpdate('arrival_status', value)}
              disabled={!canEdit || updateProgress.isPending}
            >
              <SelectTrigger id="arrival_status">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ARRIVAL_STATUS.NOT_ARRIVED}>Not Arrived</SelectItem>
                <SelectItem value={ARRIVAL_STATUS.IN_TRANSIT}>In Transit</SelectItem>
                <SelectItem value={ARRIVAL_STATUS.ARRIVED}>Arrived</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="destination_country">Destination Country</Label>
              <Input
                id="destination_country"
                type="text"
                value={progress.destination_country || ''}
                onBlur={(e) => handleUpdate('destination_country', e.target.value)}
                disabled={!canEdit || updateProgress.isPending}
                placeholder="e.g. UAE"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="arrival_city">Arrival City</Label>
              <Input
                id="arrival_city"
                type="text"
                value={progress.arrival_city || ''}
                onBlur={(e) => handleUpdate('arrival_city', e.target.value)}
                disabled={!canEdit || updateProgress.isPending}
                placeholder="Enter arrival city"
              />
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="departure_date">Departure Date</Label>
              <Input
                id="departure_date"
                type="date"
                value={progress.departure_date ? new Date(progress.departure_date).toISOString().split('T')[0] : ''}
                onChange={(e) => handleUpdate('departure_date', new Date(e.target.value).toISOString())}
                disabled={!canEdit || updateProgress.isPending}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="arrival_date">Arrival Date</Label>
              <Input
                id="arrival_date"
                type="date"
                value={progress.arrival_date ? new Date(progress.arrival_date).toISOString().split('T')[0] : ''}
                onChange={(e) => handleUpdate('arrival_date', new Date(e.target.value).toISOString())}
                disabled={!canEdit || updateProgress.isPending}
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label>Arrival Document</Label>
            {progress.arrival_document ? (
              <div className="flex items-center gap-2">
                <FileText className="h-4 w-4 text-muted-foreground" />
                <a
                  href={progress.arrival_document.file_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-blue-600 hover:underline"
                >
                  {progress.arrival_document.file_name}
                </a>
                <Download className="h-4 w-4" />
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No document uploaded</p>
            )}
            {canEdit && (
              <Input
                type="file"
                accept=".pdf,.jpg,.jpeg,.png"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) handleFileUpload('arrival', file);
                }}
                disabled={uploadDocument.isPending}
                className="mt-2"
              />
            )}
          </div>
        </CardContent>
      </Card>

      {!canEdit && (
        <Card className="border-yellow-200 bg-yellow-50 dark:border-yellow-800 dark:bg-yellow-950">
          <CardContent className="py-4 text-sm text-yellow-800 dark:text-yellow-200">
            You have read-only access to progress tracking. Only Ethiopian agents can update progress.
          </CardContent>
        </Card>
      )}
    </div>
  );
}
