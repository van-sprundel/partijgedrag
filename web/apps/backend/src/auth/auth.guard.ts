import { CanActivate, ExecutionContext, Injectable, SetMetadata, UnauthorizedException } from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { Reflector } from '@nestjs/core';
import { AuthService } from './auth.service.js';
import { unauthorizedException } from '../utils.js';
import { Response, Request } from 'express';
import { DbService } from '../db/db.service.js';
export const IS_PUBLIC_KEY = 'isPublic';
export const Public = () => SetMetadata(IS_PUBLIC_KEY, true);

export type AuthenticatedRequest = Request & {
  user: {
    sub: number;
    email: string;
  };
  cookies?: Cookies;
};

export type Cookies = {
  auth: AuthCookie;
};

export type AuthCookie = {
  accessToken: string;
  refreshToken: string;
};

export type LoggedInRequest = Request & {
  user: {
    sub: number;
    email: string;
  };
};

@Injectable()
export class AuthGuard implements CanActivate {
  constructor(
    private jwtService: JwtService,
    private authService: AuthService,
    private reflector: Reflector,
    private db: DbService,
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const isPublic = this.reflector.getAllAndOverride<boolean>(IS_PUBLIC_KEY, [
      context.getHandler(),
      context.getClass(),
    ]);
    if (isPublic) {
      // 💡 See this condition
      return true;
    }

    const request = context.switchToHttp().getRequest<AuthenticatedRequest>();

    const response = context.switchToHttp().getResponse<Response>();

    const authCookie = this.extractAuthCookie(request);
    if (!authCookie) {
      unauthorizedException('No auth cookie found');
    }

    try {
      const [session] = await this.db.userSession.getSessionByAccessToken({
        accessToken: authCookie.accessToken,
      });
      if (!session) unauthorizedException(`No session found for access token ${JSON.stringify(authCookie)}`);

      const now = new Date();
      const expiresAt = new Date(session.expiresAt);
      if (now > expiresAt) {
        await this.db.userSession.deleteSessionByAccessToken({
          accessToken: authCookie.accessToken,
        });
        unauthorizedException('Session expired');
      }

      try {
        request['user'] = await this.jwtService.verifyAsync(session.accessToken);
      } catch {
        request['user'] = await this.jwtService.verifyAsync(authCookie.refreshToken);
        const newSession = await this.authService.refreshSession(authCookie.refreshToken);
        response.cookie('auth', newSession, {
          httpOnly: true,
          maxAge: 1000 * 60 * 60 * 24 * 30,
        });
      }
    } catch (e) {
      throw new UnauthorizedException('Unauthorized', {
        cause: e,
      });
    }
    return true;
  }

  private extractAuthCookie(
    request: AuthenticatedRequest & {
      cookies: Cookies;
    },
  ): AuthCookie | undefined {
    return request.cookies.auth;
  }
}
