import { Controller, Req, Res, HttpStatus } from '@nestjs/common';
import { AuthService } from './auth.service.js';
import { LoggedInRequest, Public } from './auth.guard.js';
import { SendgridService } from '../sendgrid/sendgrid.service.js';
import { ApiTags } from '@nestjs/swagger';
import { authContract } from '@fullstack-typescript-template/ts-rest-contracts';
import { tsRestHandler, TsRestHandler } from '@ts-rest/nest';
import { Response } from 'express';
import { httpException, notFoundException } from '../utils.js';

@ApiTags()
@Controller()
export class AuthController {
  constructor(
    private readonly authService: AuthService,
    private readonly sendgridService: SendgridService,
  ) {}

  @Public()
  @TsRestHandler(authContract.register)
  register(@Res({ passthrough: true }) res: Response) {
    return tsRestHandler(authContract.register, async ({ body }) => {
      const result = await this.authService.register(body);
      const loginResult = await this.authService.login(body);

      res.cookie('auth', loginResult, {
        httpOnly: true,
        maxAge: 1000 * 60 * 60 * 24 * 30,
      });

      return { status: 200, body: result };
    });
  }

  @Public()
  @TsRestHandler(authContract.login)
  login(@Res({ passthrough: true }) res: Response) {
    return tsRestHandler(authContract.login, async ({ body }) => {
      const result = await this.authService.login(body);

      res.cookie('auth', result, {
        httpOnly: true,
        // secure: process.env.NODE_ENV === 'production ',
        // sameSite: 'strict',
        maxAge: 1000 * 60 * 60 * 24 * 30,
      });
      return {
        body: {
          message: 'Successfully logged in',
        },
        status: 200,
      };
    });
  }

  @Public()
  @TsRestHandler(authContract.forgotPassword)
  forgotPassword() {
    return tsRestHandler(authContract.forgotPassword, async ({ params, query }) => {
      const redirectUrl = query.redirectUrl;
      if (!redirectUrl) {
        httpException('No redirectUrl provided', HttpStatus.BAD_REQUEST);
      }

      await this.authService.sendPasswordResetEmail(params.email, redirectUrl);

      return {
        status: 200,
        body: { message: 'Successfully sent password reset email' },
      };
    });
  }

  @TsRestHandler(authContract.me)
  me(@Req() req: LoggedInRequest) {
    return tsRestHandler(authContract.me, async () => {
      const { email } = req.user;
      const [user] = await this.authService.getMe(email);

      if (!user) {
        notFoundException();
      }

      return { status: 200, body: user };
    });
  }

  @Public()
  @TsRestHandler(authContract.passwordReset)
  passwordReset() {
    return tsRestHandler(authContract.passwordReset, async ({ params, body }) => {
      await this.authService.resetPassword(body.password, params.token);

      return {
        status: 200,
        body: {
          message: 'Successfully updated password',
        },
      };
    });
  }

  @Public()
  @TsRestHandler(authContract.refreshToken)
  refreshToken() {
    return tsRestHandler(authContract.refreshToken, async ({ body }) => {
      const result = await this.authService.refreshSession(body.refreshToken);

      return { status: 200, body: result };
    });
  }

  @TsRestHandler(authContract.logout)
  logout(@Req() req: LoggedInRequest) {
    return tsRestHandler(authContract.logout, async () => {
      await this.authService.logout(req.user.sub);

      return {
        status: 200,
        body: {
          message: 'Successfully logged out',
        },
      };
    });
  }
}
