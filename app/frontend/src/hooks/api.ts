import { useMutation, useQueries, useQuery } from "@tanstack/react-query";
import { orpc } from "../lib/api.js";
import { getStoredSessions } from "../lib/sessionStorage.js";

export const useMotions = (params?: {
	limit?: number;
	offset?: number;
	category?: string;
	status?: string;
	withVotes?: boolean;
	search?: string;
}) => {
	return useQuery(
		orpc.motions.getAll.queryOptions({
			input: {
				limit: 20,
				offset: 0,
				...params,
			},
		}),
	);
};

export const useMotion = (id: string) => {
	return useQuery(
		orpc.motions.getById.queryOptions({
			input: { id },
			enabled: !!id,
		}),
	);
};

export const useCompassMotions = (
	count: number = 20,
	excludeIds?: string[],
	categoryIds?: string[],
	after?: Date,
	search?: string,
	partyIds?: string[],
) => {
	return useQuery(
		orpc.motions.getForCompass.queryOptions({
			input: {
				count,
				excludeIds,
				categoryIds,
				after,
				search,
				partyIds,
			},
		}),
	);
};

export const useMotionCategories = (
	type: "generic" | "hot_topic" | "all" = "all",
) => {
	return useQuery(
		orpc.motions.getCategories.queryOptions({
			input: { type },
		}),
	);
};

export const useMotionVotes = (motionId: string) => {
	return useQuery(
		orpc.motions.getVotes.queryOptions({
			input: { motionId },
			enabled: !!motionId,
		}),
	);
};

export const useParties = (params?: {
	activeOnly?: boolean;
	dateFrom?: Date;
	dateTo?: Date;
}) => {
	const { activeOnly = true, dateFrom, dateTo } = params ?? {};
	const useRange = dateFrom !== undefined && dateTo !== undefined;

	return useQuery(
		useRange
			? orpc.parties.getInRange.queryOptions({
					input: { dateFrom, dateTo },
			  })
			: orpc.parties.getAll.queryOptions({
					input: { activeOnly },
			  }),
	);
};

export const useParty = (id: string) => {
	return useQuery(
		orpc.parties.getById.queryOptions({
			input: { id },
			enabled: !!id,
		}),
	);
};

export const usePartyVotes = (partyId: string, motionIds?: string[]) => {
	return useQuery(
		orpc.parties.getWithVotes.queryOptions({
			input: {
				partyId,
				motionIds,
			},
			enabled: !!partyId,
		}),
	);
};

export const useSubmitAnswers = () => {
	return useMutation(orpc.compass.submitAnswers.mutationOptions());
};

export const useCompassResults = (sessionId: string) => {
	return useQuery(
		orpc.compass.getResults.queryOptions({
			input: { sessionId },
			enabled: !!sessionId,
		}),
	);
};

export const useMotionDetails = (
	motionId: string,
	includeVotes: boolean = true,
) => {
	return useQuery(
		orpc.compass.getMotionDetails.queryOptions({
			input: {
				motionId,
				includeVotes,
			},
			enabled: !!motionId,
		}),
	);
};

export const usePartyLikenessMatrix = (filters?: {
	dateFrom?: Date;
	dateTo?: Date;
}) => {
	return useQuery(
		orpc.statistics.getPartyLikenessMatrix.queryOptions({ input: filters }),
	);
};

export const usePartyCategoryLikeness = (
	partyId: string,
	filters?: {
		dateFrom?: Date;
		dateTo?: Date;
	},
) => {
	return useQuery(
		orpc.statistics.getPartyCategoryLikeness.queryOptions({
			input: { partyId, ...filters },
			enabled: !!partyId,
		}),
	);
};

export const usePartyFocus = (
	partyId: string,
	filters?: {
		dateFrom?: Date;
		dateTo?: Date;
	},
) => {
	return useQuery(
		orpc.statistics.getPartyFocus.queryOptions({
			input: { partyId, ...filters },
			enabled: !!partyId,
		}),
	);
};

export const useCompassMotionsCount = (filters: {
	categoryIds?: string[];
	after?: Date;
	search?: string;
	partyIds?: string[];
}) => {
	return useQuery(
		orpc.motions.getForCompassCount.queryOptions({
			input: {
				categoryIds: filters.categoryIds,
				after: filters.after,
				search: filters.search,
				partyIds: filters.partyIds,
			},
		}),
	);
};

export const useRecentSessions = () => {
	const storedSessions = getStoredSessions();

	const queries = useQueries({
		queries: storedSessions.map((session) =>
			orpc.compass.getResults.queryOptions({
				input: { sessionId: session.id },
			}),
		),
	});

	return {
		sessions: queries
			.map((query, index) => ({
				id: storedSessions[index].id,
				createdAt: storedSessions[index].createdAt,
				data: query.data,
				isLoading: query.isLoading,
				error: query.error,
			}))
			.filter((session) => session.data !== null), // Only show sessions that exist
		isLoading: queries.some((q) => q.isLoading),
	};
};
