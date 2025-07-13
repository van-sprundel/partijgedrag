import { Controller } from '@nestjs/common';
import { AppService } from './app.service.js';
import { appContract } from '@fullstack-typescript-template/ts-rest-contracts';
import { tsRestHandler, TsRestHandler } from '@ts-rest/nest';
import { Public } from './auth/auth.guard.js';

@Public()
@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}

  @Public()
  @TsRestHandler(appContract)
  handler() {
    return tsRestHandler(appContract, {
      getAppInfo: async () => {
        return {
          status: 200,
          body: {
            name: 'Full Stack Typescript Template',
            version: '0.0.0',
          },
        };
      },
    });
  }
}
