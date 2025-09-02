import { useMutation, useQuery } from "@tanstack/react-query";
import { orpc } from "../lib/api.js";

export const useMotions = (params?: {
	limit?: number;
	offset?: number;
	category?: string;
	status?: string;
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
) => {
	return useQuery(
		orpc.motions.getForCompass.queryOptions({
			input: {
				count,
				excludeIds,
				categoryIds,
				after,
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

export const useParties = (activeOnly: boolean = true) => {
	return useQuery(
		orpc.parties.getAll.queryOptions({
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
