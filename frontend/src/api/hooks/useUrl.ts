import { useMutation, useQuery } from "@tanstack/react-query";
import { apiClient } from "@/api/client";
import { UrlSchema, CreateShortenRequestSchema } from "@/api/schemas/url";

export const useShortenUrl = () => {
  return useMutation({
    mutationFn: async (originalUrl: string) => {
      const request = CreateShortenRequestSchema.parse({
        original_url: originalUrl,
      });

      const response = await apiClient.post("/shorten", request);

      return UrlSchema.parse(response.data);
    },
  });
};

export const useGetUrlByShortCode = (shortCode: string | null) => {
  return useQuery({
    queryKey: ["url", shortCode],
    queryFn: async () => {
      if (!shortCode) return null;
      const response = await apiClient.get(`/urls/short/${shortCode}`);
      return UrlSchema.parse(response.data);
    },
    enabled: !!shortCode,
  });
};
