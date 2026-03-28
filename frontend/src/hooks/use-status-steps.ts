import { AxiosError } from "axios";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import api from "@/lib/api";
import { useCurrentUser } from "@/hooks/use-auth";
import { usePairingStore } from "@/stores/pairing-store";
import { CandidateProgress, CandidateStatus, UserRole } from "@/types";

interface StatusStepApiError {
  error?: string;
  message?: string;
}

export function useCandidateProgress(
  candidateId?: string,
  enabled: boolean = true,
) {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const requiresWorkspace = user?.role === UserRole.FOREIGN_AGENT;
  const canQuery =
    enabled &&
    !!candidateId &&
    !!user &&
    (!requiresWorkspace || (isPairingReady && !!activePairingId));

  return useQuery({
    queryKey: ["candidate-progress", candidateId, activePairingId],
    queryFn: async () => {
      const response = await api.get<CandidateProgress>(
        `/candidates/${candidateId}/status-steps`,
      );
      return response.data;
    },
    enabled: canQuery,
    staleTime: 60_000,
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => {
      const status = (error as AxiosError)?.response?.status;
      if (status === 403 || status === 404) {
        return false;
      }

      return failureCount < 2;
    },
  });
}

export function useUpdateStatusStep(candidateId: string) {
  const queryClient = useQueryClient();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const progressQueryKey = [
    "candidate-progress",
    candidateId,
    activePairingId,
  ] as const;
  const candidateQueryKey = [
    "candidate",
    candidateId,
    activePairingId,
  ] as const;

  return useMutation({
    mutationFn: async ({
      step_name,
      status,
      notes,
    }: {
      step_name: string;
      status: string;
      notes?: string;
    }) => {
      const encodedStepName = encodeURIComponent(step_name);
      const response = await api.patch<CandidateProgress>(
        `/candidates/${candidateId}/status-steps/${encodedStepName}`,
        { status, notes },
      );
      return response.data;
    },
    onMutate: async ({ step_name, status, notes }) => {
      await queryClient.cancelQueries({ queryKey: progressQueryKey });

      const previousProgress =
        queryClient.getQueryData<CandidateProgress>(progressQueryKey);

      if (previousProgress) {
        const nextSteps = previousProgress.steps.map((step) => {
          if (step.step_name !== step_name) {
            return step;
          }

          return {
            ...step,
            step_status:
              status as CandidateProgress["steps"][number]["step_status"],
            notes: typeof notes === "string" ? notes : step.notes,
            completed_at:
              status === "completed" ? new Date().toISOString() : undefined,
            updated_at: new Date().toISOString(),
          };
        });

        const completedCount = nextSteps.filter(
          (step) => step.step_status === "completed",
        ).length;
        const hasCompletedAllSteps =
          nextSteps.length > 0 && completedCount === nextSteps.length;

        queryClient.setQueryData<CandidateProgress>(progressQueryKey, {
          ...previousProgress,
          steps: nextSteps,
          progress_percentage: nextSteps.length
            ? (completedCount / nextSteps.length) * 100
            : 0,
          overall_status: hasCompletedAllSteps
            ? CandidateStatus.COMPLETED
            : CandidateStatus.IN_PROGRESS,
          last_updated_at: new Date().toISOString(),
        });
      }

      return { previousProgress };
    },
    onSuccess: (progress) => {
      queryClient.setQueryData(progressQueryKey, progress);
      queryClient.setQueryData(candidateQueryKey, (previousCandidate: any) =>
        previousCandidate
          ? {
              ...previousCandidate,
              status: progress.overall_status,
            }
          : previousCandidate,
      );
      queryClient.setQueriesData(
        { queryKey: ["candidates"] },
        (previous: any) => {
          if (!previous?.data || !Array.isArray(previous.data)) {
            return previous;
          }

          return {
            ...previous,
            data: previous.data.map((candidate: any) =>
              candidate?.id === candidateId
                ? {
                    ...candidate,
                    status: progress.overall_status,
                  }
                : candidate,
            ),
          };
        },
      );
      toast.success("Status step updated successfully");
    },
    onError: (error, _variables, context) => {
      if (context?.previousProgress) {
        queryClient.setQueryData(progressQueryKey, context.previousProgress);
      }

      const response = (error as AxiosError<StatusStepApiError>).response;

      if (response?.status === 403) {
        toast.error(
          "Only the Ethiopian agency that created this candidate can update process steps.",
        );
        return;
      }

      if (response?.status === 400 && response.data?.error) {
        toast.error(response.data.error);
        return;
      }

      toast.error(
        response?.data?.error ||
          response?.data?.message ||
          "Failed to update status step",
      );
    },
  });
}
