import { HttpException, NotFoundException, UnauthorizedException } from '@nestjs/common';
import { HttpExceptionOptions } from '@nestjs/common/exceptions/http.exception.js';

export function httpException(
  response: string | Record<string, any>,
  status: number,
  options?: HttpExceptionOptions,
): never {
  console.error('HTTP EXCEPTION', {
    response,
    status,
    options,
  });
  throw new HttpException(response, status, options);
}

export function notFoundException(objectOrError?: any, options?: HttpExceptionOptions): never {
  console.error('NOT FOUND EXCEPTION', {
    objectOrError,
    options,
  });
  throw new NotFoundException(objectOrError, options);
}

export function unauthorizedException(objectOrError?: any, options?: HttpExceptionOptions): never {
  console.error('UNAUTHORIZED EXCEPTION', {
    objectOrError,
    options,
  });
  throw new UnauthorizedException(objectOrError, options);
}
