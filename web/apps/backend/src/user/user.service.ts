import { HttpException, HttpStatus, Injectable } from '@nestjs/common';
import { RegisterBody } from '@fullstack-typescript-template/ts-rest-contracts';
import { DbService } from '../db/db.service.js';

@Injectable()
export class UserService {
  constructor(private db: DbService) {}

  async findOne(email: string) {
    return await this.db.user.getUserByEmail({ email });
  }

  async findOneUnsafe(email: string) {
    return await this.db.user.getUnsafeUserByEmail({ email });
  }

  async create(newUserParams: RegisterBody) {
    if (await this.findOne(newUserParams.email).then((res) => res.length !== 0)) {
      throw new HttpException('User with this email already exists!', HttpStatus.BAD_REQUEST);
    }

    await this.db.user.createUser(newUserParams);

    return { message: 'Successfully created user' };
  }

  async updatePassword(id: number, password: string) {
    await this.db.user.updatePassword({ id, password });

    return { status: HttpStatus.OK, message: 'Successfully updated password' };
  }
}
