import { AxiosError } from "axios";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import api from "@/lib/api";
import { PassportData } from "@/types";

interface PassportResponse {
  passport: PassportData;
}

interface ParsePassportInput {
  file: File;
  signal?: AbortSignal;
  showSuccessToast?: boolean;
}

export function usePassportData(candidateId?: string, enabled: boolean = true) {
  return useQuery({
    queryKey: ["passport-data", candidateId],
    queryFn: async () => {
      const response = await api.get<PassportResponse>(`/candidates/${candidateId}/passport`);
      return response.data.passport;
    },
    enabled: enabled && Boolean(candidateId),
    retry: (failureCount, error) => {
      const status = (error as AxiosError)?.response?.status;
      if (status === 404 || status === 403) {
        return false;
      }
      return failureCount < 1;
    },
  });
}

export function useParsePassport(candidateId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      file,
      signal,
      showSuccessToast = true,
    }: ParsePassportInput) => {
      const formData = new FormData();
      formData.append("file", file);

      const endpoint = candidateId
        ? `/candidates/${candidateId}/passport/parse`
        : "/candidates/passport/parse-preview";

      const response = await api.post<PassportResponse>(endpoint, formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
        signal,
      });

      return {
        passport: response.data.passport,
        showSuccessToast,
      };
    },
    onSuccess: ({ passport, showSuccessToast }) => {
      if (candidateId) {
        queryClient.setQueryData(["passport-data", candidateId], passport);
      }
      if (showSuccessToast) {
        toast.success("Passport data extracted successfully.");
      }
    },
    onError: (error) => {
      if ((error as AxiosError)?.code === "ERR_CANCELED") {
        return;
      }
      const message =
        (error as AxiosError<{ error?: string }>)?.response?.data?.error ||
        (error as Error)?.message ||
        "Failed to extract passport data";
      toast.error(message);
    },
  });
}
