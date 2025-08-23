import type { router as BackendRouter } from "@backend/src/index.js";
import { createORPCClient } from "@orpc/client";
import { RPCLink } from "@orpc/client/fetch";
import type { RouterClient } from "@orpc/server";
import { createTanstackQueryUtils } from "@orpc/tanstack-query";

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:3001";

export const link = new RPCLink({
	url: `${BASE_URL}/api`,
});

export const client: RouterClient<typeof BackendRouter> =
	createORPCClient(link);
export const orpc = createTanstackQueryUtils(client);

export type {
	CompassResult,
	Motion,
	MotionDetail,
	Party,
	PartyResult,
	UserAnswer,
	Vote,
	VoteType,
} from "../../../backend/src/contracts/index.js";
