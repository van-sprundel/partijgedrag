import { z } from 'zod';

export type RegisterBody = z.infer<typeof registerBodySchema>;
export type LoginBody = z.infer<typeof loginBodySchema>;
export type ForgotPasswordParam = z.infer<typeof forgotPasswordParamSchema>;
export type ResetPasswordBody = z.infer<typeof resetPasswordBodySchema>;
export type RefreshTokenBody = z.infer<typeof refreshTokenBodySchema>;

export const registerBodySchema = z.object({
  email: z.string().email(),
  password: z.string(),
});

export const loginBodySchema = z.object({
  email: z.string().email(),
  password: z.string(),
});

export const forgotPasswordParamSchema = z.object({
  email: z.string().email(),
});

export const forgotPasswordQueryParamSchema = z.object({
  redirectUrl: z.string(),
});

export const resetPasswordParamSchema = z.object({
  token: z.string(),
});

export const resetPasswordBodySchema = z.object({
  password: z.string(),
});

export const refreshTokenBodySchema = z.object({
  refreshToken: z.string(),
});
