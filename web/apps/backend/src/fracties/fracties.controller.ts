import { Controller } from '@nestjs/common';
import {appContract, authContract} from '@fullstack-typescript-template/ts-rest-contracts';
import { tsRestHandler, TsRestHandler } from '@ts-rest/nest';
import { Public } from '../auth/auth.guard.js';
import {fractieContract} from "@fullstack-typescript-template/ts-rest-contracts/dist/contracts/fracties.contract.js";
import {FractiesService} from "./fracties.service.js";

@Public()
@Controller()
export class FractiesController {
  constructor(private readonly fractiesService: FractiesService ) {}

  @Public()
  @TsRestHandler(fractieContract.getAll)
  async handler() {
    const fracties = await this.fractiesService.getAll()

    return tsRestHandler(fractieContract, {
      getAll: async () => {
        return {
          status: 200,
          body: {
            fracties
          },
        };
      },
    });
  }
}
