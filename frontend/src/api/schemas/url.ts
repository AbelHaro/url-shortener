import { z } from "zod";

export const ApiUrlSchema = z.object({
  id: z.string(),
  original_url: z.url(),
  short_code: z.string(),
});

export const UrlSchema = ApiUrlSchema.transform((data) => ({
  id: data.id,
  originalUrl: data.original_url,
  shortCode: data.short_code,
}));

export type Url = z.infer<typeof UrlSchema>;

export const CreateShortenRequestSchema = z.object({
  original_url: z.url(),
});

export type CreateShortenRequest = z.infer<typeof CreateShortenRequestSchema>;

export const GetUrlByShortCodeRequestSchema = z.object({
  short_code: z.string(),
});

export type GetUrlByShortCodeRequest = z.infer<typeof GetUrlByShortCodeRequestSchema>;
