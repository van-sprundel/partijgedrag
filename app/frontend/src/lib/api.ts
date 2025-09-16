import type { router as BackendRouter } from "@backend/src/index.js";
import { createORPCClient } from "@orpc/client";
import { RPCLink } from "@orpc/client/fetch";
import type { RouterClient } from "@orpc/server";
import { createTanstackQueryUtils } from "@orpc/tanstack-query";

const getApiUrl = () => {
	if (import.meta.env.PROD) {
		// needs absolute path
		return `${window.location.origin}/api`;
	}
	return `${import.meta.env.VITE_API_URL}/api`;
};

export const link = new RPCLink({
	url: getApiUrl(),
});

export const client: RouterClient<typeof BackendRouter> =
	createORPCClient(link);
export const orpc = createTanstackQueryUtils(client);

export type {
	CompassResult,
	Motion,
	MotionCategory,
	MotionDetail,
	Party,
	PartyResult,
	UserAnswer,
	Vote,
	VoteType,
} from "../../../backend/src/contracts/index.js";
