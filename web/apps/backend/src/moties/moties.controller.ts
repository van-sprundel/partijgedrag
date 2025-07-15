import { Controller } from '@nestjs/common';
import { appContract, authContract } from '@fullstack-typescript-template/ts-rest-contracts';
import { tsRestHandler, TsRestHandler } from '@ts-rest/nest';
import { Public } from '../auth/auth.guard.js';
import { fractieContract } from '@fullstack-typescript-template/ts-rest-contracts/dist/contracts/fracties.contract.js';
import { MotiesService } from './moties.service.js';
import { motieContract } from '@fullstack-typescript-template/ts-rest-contracts/dist/contracts/moties.contract.js';

@Public()
@Controller()
export class MotiesController {
  constructor(private readonly motiesService: MotiesService) {}

  // changes

  @Public()
  @TsRestHandler(motieContract.getAll)
  async handler() {
    return await tsRestHandler(motieContract.getAll, async () => {
      const moties = await this.motiesService.getAll();
      return {
        body: {
          moties,
        },
        status: 200,
      };
    });
  }
}
