import { HttpException, HttpStatus, Injectable, UnauthorizedException } from '@nestjs/common';
import bcrypt from 'bcryptjs';
import { UserService } from '../user/user.service.js';
import { JwtService } from '@nestjs/jwt';
import { SendgridService } from '../sendgrid/sendgrid.service.js';
import { LoginBody, RegisterBody } from '@fullstack-typescript-template/ts-rest-contracts';
import date from 'date-and-time';
import { ACCESS_TOKEN_EXPIRES_IN, REFRESH_TOKEN_EXPIRES_IN, REFRESH_TOKEN_EXPIRES_IN_MINUTES } from './contants.js';
import { Cron, CronExpression } from '@nestjs/schedule';
import { randomUUID } from 'node:crypto';
import { DbService } from '../db/db.service.js';
import { httpException } from '../utils.js';

@Injectable()
export class AuthService {
  constructor(
    private readonly userService: UserService,
    private jwtService: JwtService,
    private sendGridService: SendgridService,
    private db: DbService,
  ) {}

  @Cron(CronExpression.EVERY_30_MINUTES)
  async deleteExpiredSessions() {
    await this.db.userSession.deleteExpiredSessions();
    await this.db.passwordResetRequest.deleteExpiredPasswordResetRequests();
  }

  async register(body: RegisterBody) {
    body.password = hashPassword(body.password);

    const result = await this.userService.create(body);
    const [user] = await this.userService.findOne(body.email);
    if (!user) throw new HttpException('Could not create user.', HttpStatus.INTERNAL_SERVER_ERROR);

    return result;
  }
  async login(body: LoginBody) {
    const [user] = await this.userService.findOneUnsafe(body.email);

    if (!user || !validatePassword(body.password, user.password)) {
      throw new HttpException('Wrong combination of email and password.', HttpStatus.UNAUTHORIZED);
    }

    const payload = { sub: user.id, email: user.email, date: new Date() };

    const [session] = await this.createSession(user.id, payload);
    if (!session) throw new HttpException('Could not create session.', HttpStatus.INTERNAL_SERVER_ERROR);

    return {
      accessToken: session.accessToken,
      refreshToken: session.refreshToken,
    };
  }
  async sendPasswordResetEmail(email: string, redirectUrl: string) {
    const [user] = await this.userService.findOne(email);
    if (!user) {
      return { message: 'Successfully Sent Password Reset Email' };
    }

    const [existingPasswordResetRequest] = await this.db.passwordResetRequest.getPasswordResetRequestByUserId({
      userId: user.id,
    });
    if (existingPasswordResetRequest) {
      await this.db.passwordResetRequest.deletePasswordResetRequestByUserId({
        userId: user.id,
      });
    }

    const [passwordResetRequest] = await this.db.passwordResetRequest.createPasswordResetRequest({
      userId: user.id,
      token: randomUUID(),
      expiresAt: date.addMinutes(new Date(), 15),
    });
    if (!passwordResetRequest) {
      throw new HttpException('Could not create password reset request', HttpStatus.INTERNAL_SERVER_ERROR);
    }

    try {
      await this.sendGridService.sendPasswordResetEmail(email, passwordResetRequest.token, redirectUrl);
    } catch (e) {
      throw new HttpException('Could not send password reset email', HttpStatus.INTERNAL_SERVER_ERROR, {
        cause: e,
      });
    }

    return { message: 'Successfully Sent Password Reset Email' };
  }

  async resetPassword(password: string, token: string) {
    try {
      const [passwordResetRequest] = await this.db.passwordResetRequest.getPasswordResetRequestByToken({
        token,
      });
      if (!passwordResetRequest) {
        httpException('No password reset request found', HttpStatus.INTERNAL_SERVER_ERROR);
      }
      await this.db.passwordResetRequest.deletePasswordResetRequestByToken({
        token,
      });
      await this.userService.updatePassword(passwordResetRequest.userId, password);
      return {
        status: HttpStatus.OK,
        body: { message: 'Successfully Reset Password' },
      };
    } catch {
      throw new UnauthorizedException();
    }
  }

  async getMe(email: string) {
    return await this.userService.findOne(email);
  }

  async createSession(userId: number, payload: Buffer | object) {
    const expiresAt = date.addMinutes(new Date(), REFRESH_TOKEN_EXPIRES_IN_MINUTES);
    const accessToken = this.jwtService.sign(payload, {
      expiresIn: ACCESS_TOKEN_EXPIRES_IN,
    });
    const refreshToken = this.jwtService.sign(payload, {
      expiresIn: REFRESH_TOKEN_EXPIRES_IN,
    });

    return await this.db.userSession.createSession({
      userId,
      expiresAt,
      accessToken,
      refreshToken,
    });
  }

  async deleteSession(accessToken: string) {
    try {
      await this.db.userSession.deleteSessionByAccessToken({ accessToken });
    } catch {
      throw new HttpException('Could not delete session', HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }

  async logout(userId: number) {
    try {
      await this.db.userSession.deleteAllSessionsOfUser({ userId });
    } catch {
      throw new HttpException('Could not delete session', HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }

  async refreshSession(refreshToken: string) {
    const { sub, email } = await this.jwtService.verifyAsync<{
      sub: number;
      email: string;
    }>(refreshToken);

    const [session] = await this.db.userSession.getSessionByRefreshToken({
      refreshToken,
    });
    if (!session) {
      throw new HttpException('Could not find session', HttpStatus.INTERNAL_SERVER_ERROR);
    }

    await this.deleteSession(session.accessToken);

    const [newSession] = await this.createSession(session.userId, {
      sub,
      email,
    });
    if (!newSession) throw new HttpException('Could not create session', HttpStatus.INTERNAL_SERVER_ERROR);

    return {
      accessToken: newSession.accessToken,
      refreshToken: newSession.refreshToken,
    };
  }
}

export function hashPassword(password: string) {
  return bcrypt.hashSync(password, 10);
}

export function validatePassword(password: string, hash: string) {
  return bcrypt.compareSync(password, hash);
}
