import {
  ArgumentsHost,
  Catch,
  ExceptionFilter,
  HttpException,
  HttpStatus,
  UnauthorizedException,
} from '@nestjs/common';
import { Request, Response } from 'express';
import { globalResponseError } from './global.response.error.js';

@Catch()
export class GlobalExceptionFilter implements ExceptionFilter {
  catch(exception: any, host: ArgumentsHost) {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse<Response>();
    const request = ctx.getRequest<Request>();
    let message = exception.message;
    const code = exception.code;
    let status = HttpStatus.INTERNAL_SERVER_ERROR;

    if (exception.constructor) {
      switch (exception.constructor) {
        case HttpException:
          status = (exception as HttpException).getStatus();
          message = (exception as HttpException).getResponse();
          break;
        case UnauthorizedException:
          status = HttpStatus.UNAUTHORIZED;
          break;
        default:
          break;
      }
    }

    response
      .status(status)
      .json(globalResponseError(status, message, code, request));
  }
}
