import { Request } from 'express';
export function globalResponseError(
  statusCode: number,
  message: string,
  code: string,
  request: Request,
): GlobalResponseError {
  return {
    statusCode: statusCode,
    message,
    code,
    timestamp: new Date().toISOString(),
    path: request.url,
    method: request.method,
  };
}

export interface GlobalResponseError {
  statusCode: number;
  message: string;
  code: string;
  timestamp: string;
  path: string;
  method: string;
}
