import { Controller } from '@nestjs/common';
import { appContract, authContract } from '@fullstack-typescript-template/ts-rest-contracts';
import { tsRestHandler, TsRestHandler } from '@ts-rest/nest';
import { Public } from '../auth/auth.guard.js';
import { fractieContract } from '@fullstack-typescript-template/ts-rest-contracts/dist/contracts/fracties.contract.js';
import { FractiesService } from './fracties.service.js';
import { motieContract } from '@fullstack-typescript-template/ts-rest-contracts/dist/contracts/moties.contract.js';

@Public()
@Controller()
export class FractiesController {
  constructor(private readonly fractiesService: FractiesService) {}

  // changes

  @Public()
  @TsRestHandler(fractieContract.getAll)
  async handler() {
    return await tsRestHandler(fractieContract.getAll, async ({ query }) => {
      const moties = await this.fractiesService.getAll(query.page, query.pageSize);
      return {
        body: {
          moties,
        },
        status: 200,
      };
    });
  }
}
