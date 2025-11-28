import { randomBytes, randomUUID } from "node:crypto";
import { implement, ORPCError } from "@orpc/server";
import type { Party, UserAnswer, Vote, VoteType } from "../contracts/index.js";
import { apiContract, CompassResultSchema } from "../contracts/index.js";
import { sql, sqlOneOrNull } from "../services/db/sql-tag.js";
import {
	mapCaseToMotion,
	mapPartyToContract,
	mapVoteToContract,
} from "../utils/mappers.js";

const os = implement(apiContract);

function mapVoteType(dutchVoteType: string | null): VoteType {
	switch (dutchVoteType) {
		case "Voor":
			return "FOR";
		case "Tegen":
			return "AGAINST";
		case "Niet deelgenomen":
			return "NEUTRAL";
		default:
			return "NEUTRAL";
	}
}

export const compassRouter = {
	submitAnswers: os.compass.submitAnswers.handler(async ({ input }) => {
		const { answers } = input;

		const partyResults = await calculatePartyAlignment(answers);
		const id = randomUUID();
		const now = new Date();

		await sql`
			INSERT INTO user_sessions (id, answers, results, "createdAt", "updatedAt")
			VALUES (
				${id},
				${JSON.stringify(answers)},
				${JSON.stringify({
					totalAnswers: answers.length,
					partyResults,
					createdAt: now,
				})},
				${now},
				${now}
			)
		`;

		return {
			id,
			totalAnswers: answers.length,
			partyResults,
			createdAt: now,
		};
	}),

	getResults: os.compass.getResults.handler(async ({ input }) => {
		const session = await sqlOneOrNull<{
			id: string;
			answers: unknown;
			results: unknown | null;
			createdAt: Date;
			updatedAt: Date;
		}>`
			SELECT * FROM user_sessions WHERE id = ${input.sessionId}
		`;

		if (!session || !session.results) {
			return null;
		}

		const results = CompassResultSchema.parse({
			...(session.results as object),
			id: session.id,
		});

		const answers = session.answers as UserAnswer[];
		const motionDetails = await getMotionVoteDetails(answers);

		return {
			id: session.id,
			totalAnswers: results.totalAnswers,
			partyResults: results.partyResults,
			motionDetails,
			createdAt: session.createdAt,
		};
	}),

	getMotionDetails: os.compass.getMotionDetails.handler(async ({ input }) => {
		const { motionId, includeVotes } = input;

		const zaak = await sqlOneOrNull<{
			id: string;
			number: string | null;
			subject: string | null;
			type: string | null;
			title: string | null;
			citationTitle: string | null;
			alias: string | null;
			status: string | null;
			date: Date | null;
			startedAt: Date | null;
			organization: string | null;
			basis: string | null;
			term: string | null;
			meetingYear: string | null;
			sequenceNumber: string | null;
			currentTreatmentStatus: string | null;
			completed: boolean | null;
			largeProject: boolean | null;
			updatedAt: Date | null;
			apiUpdatedAt: Date | null;
			removed: boolean | null;
			cabinetAppreciation: string | null;
			completedAt: Date | null;
			chamber: string | null;
			bulletPoints: unknown | null;
			documentURL: string | null;
			did: string | null;
		}>`
			SELECT
				id,
				nummer as number,
				onderwerp as subject,
				soort as type,
				titel as title,
				citeertitel as "citationTitle",
				alias,
				status,
				datum as date,
				gestart_op as "startedAt",
				organisatie as organization,
				grondslagvoorhang as basis,
				termijn as term,
				vergaderjaar as "meetingYear",
				volgnummer as "sequenceNumber",
				huidige_behandelstatus as "currentTreatmentStatus",
				afgedaan as completed,
				groot_project as "largeProject",
				gewijzigd_op as "updatedAt",
				api_gewijzigd_op as "apiUpdatedAt",
				verwijderd as removed,
				kabinetsappreciatie as "cabinetAppreciation",
				datum_afgedaan as "completedAt",
				kamer as chamber,
				bullet_points as "bulletPoints",
				document_url as "documentURL",
				did
			FROM zaken
			WHERE id = ${motionId}
		`;

		if (!zaak) {
			throw new ORPCError("NOT_FOUND", { message: "Motion not found" });
		}

		const motion = mapCaseToMotion(zaak);

		let votes: Vote[] = [];
		let partyPositions: {
			party: Party;
			position: VoteType;
			count: number;
		}[] = [];

		if (includeVotes) {
			const decisions = await sql<{ id: string; caseId: string | null }>`
				SELECT id, zaak_id as "caseId"
				FROM besluiten
				WHERE zaak_id = ${motionId}
			`;
			const decisionIds = decisions.map((d) => d.id);

			if (decisionIds.length > 0) {
				const votesWithRelations = await sql<{
					id: string;
					decisionIdRaw: string | null;
					decisionId: string | null;
					type: string | null;
					partySize: string | null;
					actorName: string | null;
					actorParty: string | null;
					mistake: boolean | null;
					sidActorMember: string | null;
					sidActorParty: string | null;
					politicianIdRaw: string | null;
					politicianId: string | null;
					partyIdRaw: string | null;
					partyId: string | null;
					updatedAt: Date | null;
					apiUpdatedAt: Date | null;
				}>`
					SELECT
						id,
						besluit_id_raw as "decisionIdRaw",
						besluit_id as "decisionId",
						soort as type,
						fractie_grootte as "partySize",
						actor_naam as "actorName",
						actor_fractie as "actorParty",
						vergissing as mistake,
						sid_actor_lid as "sidActorMember",
						sid_actor_fractie as "sidActorParty",
						persoon_id_raw as "politicianIdRaw",
						persoon_id as "politicianId",
						fractie_id_raw as "partyIdRaw",
						fractie_id as "partyId",
						gewijzigd_op as "updatedAt",
						api_gewijzigd_op as "apiUpdatedAt"
					FROM stemmingen
					WHERE besluit_id = ANY(${decisionIds})
				`;

				votes = votesWithRelations.map((v) => mapVoteToContract(v));

				const partyVoteMap = new Map<
					string,
					{ party: PartyMapped; votes: VoteType[]; partySize: number }
				>();

				const partyIds = votesWithRelations
					.map((v) => v.partyId)
					.filter((p) => p !== null) as string[];

				if (partyIds.length > 0) {
					const parties = await sql<{
						id: string;
						number: string | null;
						shortName: string | null;
						nameNl: string | null;
						nameEn: string | null;
						seats: string | null;
						votesCount: string | null;
						activeFrom: Date | null;
						activeTo: Date | null;
						contentType: string | null;
						contentLength: string | null;
						updatedAt: Date | null;
						apiUpdatedAt: Date | null;
						logoData: any | null;
						removed: boolean | null;
					}>`
						SELECT
							id,
							nummer as number,
							afkorting as "shortName",
							naam_nl as "nameNl",
							naam_en as "nameEn",
							aantal_zetels as seats,
							aantal_stemmen as "votesCount",
							datum_actief as "activeFrom",
							datum_inactief as "activeTo",
							content_type as "contentType",
							content_length as "contentLength",
							gewijzigd_op as "updatedAt",
							api_gewijzigd_op as "apiUpdatedAt",
							logo_data as "logoData",
							verwijderd as removed
						FROM fracties
						WHERE id = ANY(${partyIds})
					`;

					votesWithRelations.forEach((vote) => {
						if (vote.partyId) {
							const party = parties.find((p) => p.id === vote.partyId);
							if (party) {
								if (!partyVoteMap.has(vote.partyId)) {
									partyVoteMap.set(vote.partyId, {
										party: party,
										votes: [],
										partySize: Number(vote.partySize || 0),
									});
								}
								if (vote.type) {
									partyVoteMap
										.get(vote.partyId)
										?.votes.push(mapVoteType(vote.type));
								}
							}
						}
					});
				}

				partyPositions = Array.from(partyVoteMap.values()).map(
					({ party, votes: partyVotes, partySize }) => {
						const voteCounts = partyVotes.reduce(
							(acc, vote) => {
								acc[vote] = (acc[vote] || 0) + 1;
								return acc;
							},
							{} as Record<string, number>,
						);

						const majorityVoteEntry =
							Object.keys(voteCounts).length > 0
								? Object.entries(voteCounts).reduce((a, b) =>
										a[1] > b[1] ? a : b,
									)
								: [];

						const majorityVote =
							(majorityVoteEntry[0] as VoteType) ?? ("NEUTRAL" as const);

						return {
							party: mapPartyToContract(party),
							position: majorityVote,
							count: partySize,
						};
					},
				);
			}
		}

		return {
			motion,
			votes: includeVotes ? votes : undefined,
			partyPositions: includeVotes ? partyPositions : undefined,
		};
	}),
};

async function calculatePartyAlignment(answers: UserAnswer[]) {
	const motionIds = answers.map((a) => a.motionId);

	const decisions = await sql<{ id: string; caseId: string | null }>`
		SELECT id, zaak_id as "caseId"
		FROM besluiten
		WHERE zaak_id = ANY(${motionIds})
	`;
	const decisionIds = decisions.map((d) => d.id);

	if (decisionIds.length === 0) {
		return [];
	}

	const votes = await sql<{
		id: string;
		decisionId: string | null;
		type: string | null;
		partyId: string | null;
		actorParty: string | null;
	}>`
		SELECT
			id,
			besluit_id as "decisionId",
			soort as type,
			fractie_id as "partyId",
			actor_fractie as "actorParty"
		FROM stemmingen
		WHERE besluit_id = ANY(${decisionIds})
	`;

	const parties = await sql<{
		id: string;
		number: string | null;
		shortName: string | null;
		nameNl: string | null;
		nameEn: string | null;
		seats: string | null;
		votesCount: string | null;
		activeFrom: Date | null;
		activeTo: Date | null;
		contentType: string | null;
		contentLength: string | null;
		updatedAt: Date | null;
		apiUpdatedAt: Date | null;
		logoData: any | null;
		removed: boolean | null;
	}>`
		SELECT
			id,
			nummer as number,
			afkorting as "shortName",
			naam_nl as "nameNl",
			naam_en as "nameEn",
			aantal_zetels as seats,
			aantal_stemmen as "votesCount",
			datum_actief as "activeFrom",
			datum_inactief as "activeTo",
			content_type as "contentType",
			content_length as "contentLength",
			gewijzigd_op as "updatedAt",
			api_gewijzigd_op as "apiUpdatedAt",
			logo_data as "logoData",
			verwijderd as removed
		FROM fracties
		WHERE datum_inactief IS NULL OR datum_inactief >= NOW()
	`;

	const partyNameMap = new Map<string, PartyMapped>();
	parties.forEach((p) => {
		if (p.nameNl) partyNameMap.set(p.nameNl, p);
	});

	const partyScores = new Map<
		string,
		{
			party: Party;
			totalVotes: number;
			matchingVotes: number;
			score: number;
			agreement: number;
		}
	>();

	parties.forEach((party) => {
		partyScores.set(party.id, {
			party: mapPartyToContract(party),
			totalVotes: 0,
			matchingVotes: 0,
			score: 0,
			agreement: 0,
		});
	});

	answers.forEach((answer) => {
		const decision = decisions.find((d) => d.caseId === answer.motionId);
		if (!decision) return;

		const motionVotes = votes.filter((v) => v.decisionId === decision.id);

		const partyPositions = new Map<string, VoteType[]>();

		motionVotes.forEach((vote) => {
			const party = vote.partyId
				? parties.find((p) => p.id === vote.partyId)
				: vote.actorParty
					? partyNameMap.get(vote.actorParty)
					: undefined;

			if (party?.id && vote.type) {
				const partyId = party.id;
				if (!partyPositions.has(partyId)) {
					partyPositions.set(partyId, []);
				}
				partyPositions.get(partyId)?.push(mapVoteType(vote.type));
			}
		});

		partyPositions.forEach((votes, partyId) => {
			const partyScore = partyScores.get(partyId);
			if (!partyScore) return;

			partyScore.totalVotes++;

			const voteCounts = votes.reduce(
				(acc, vote) => {
					acc[vote] = (acc[vote] || 0) + 1;
					return acc;
				},
				{} as Record<VoteType, number>,
			);

			const majorityVote =
				Object.keys(voteCounts).length > 0
					? (Object.entries(voteCounts).reduce((a, b) =>
							a[1] > b[1] ? a : b,
						)[0] as VoteType)
					: null;

			// Skip neutral answers for agreement calculation
			if (answer.answer !== "neutral") {
				const userSupports = answer.answer === "agree";
				const partySupports = majorityVote === "FOR";

				if (
					(userSupports && partySupports) ||
					(!userSupports && !partySupports)
				) {
					partyScore.matchingVotes++;
					partyScore.score += 1;
				}
			} else {
				// Neutral answers contribute to score but not to matching votes
				partyScore.score += 0.5;
			}

			partyScore.agreement =
				partyScore.totalVotes > 0
					? (partyScore.matchingVotes / partyScore.totalVotes) * 100
					: 0;
		});
	});

	const minVotesThreshold = Math.max(1, Math.floor(answers.length * 0.25));

	const results = Array.from(partyScores.values())
		.filter((result) => result.totalVotes >= minVotesThreshold)
		.sort((a, b) => {
			if (Math.abs(b.agreement - a.agreement) > 0.1) {
				return b.agreement - a.agreement;
			}
			if (Math.abs(b.totalVotes - a.totalVotes) > 0) {
				return b.totalVotes - a.totalVotes;
			}
			return b.score - a.score;
		});

	return results.map((result) => ({
		party: result.party,
		score: Math.round(result.score * 100) / 100,
		agreement: Math.round(result.agreement * 100) / 100,
		totalVotes: result.totalVotes,
		matchingVotes: result.matchingVotes,
	}));
}

async function getMotionVoteDetails(answers: UserAnswer[]) {
	const motionIds = answers.map((a) => a.motionId);

	const motions = await sql<{
		id: string;
		number: string | null;
		subject: string | null;
		type: string | null;
		title: string | null;
		citationTitle: string | null;
		alias: string | null;
		status: string | null;
		date: Date | null;
		startedAt: Date | null;
		organization: string | null;
		basis: string | null;
		term: string | null;
		meetingYear: string | null;
		sequenceNumber: string | null;
		currentTreatmentStatus: string | null;
		completed: boolean | null;
		largeProject: boolean | null;
		updatedAt: Date | null;
		apiUpdatedAt: Date | null;
		removed: boolean | null;
		cabinetAppreciation: string | null;
		completedAt: Date | null;
		chamber: string | null;
		bulletPoints: unknown | null;
		documentURL: string | null;
		did: string | null;
	}>`
		SELECT
			id,
			nummer as number,
			onderwerp as subject,
			soort as type,
			titel as title,
			citeertitel as "citationTitle",
			alias,
			status,
			datum as date,
			gestart_op as "startedAt",
			organisatie as organization,
			grondslagvoorhang as basis,
			termijn as term,
			vergaderjaar as "meetingYear",
			volgnummer as "sequenceNumber",
			huidige_behandelstatus as "currentTreatmentStatus",
			afgedaan as completed,
			groot_project as "largeProject",
			gewijzigd_op as "updatedAt",
			api_gewijzigd_op as "apiUpdatedAt",
			verwijderd as removed,
			kabinetsappreciatie as "cabinetAppreciation",
			datum_afgedaan as "completedAt",
			kamer as chamber,
			bullet_points as "bulletPoints",
			document_url as "documentURL",
			did
		FROM zaken
		WHERE id = ANY(${motionIds})
	`;

	const decisions = await sql<{ id: string; caseId: string | null }>`
		SELECT id, zaak_id as "caseId"
		FROM besluiten
		WHERE zaak_id = ANY(${motionIds})
	`;
	const decisionIds = decisions.map((d) => d.id);

	if (decisionIds.length === 0) {
		return answers.map((answer) => ({
			motionId: answer.motionId,
			userAnswer: answer.answer,
			motion: null,
			partyPositions: [],
		}));
	}

	const votes = await sql<{
		id: string;
		decisionId: string | null;
		type: string | null;
		partyId: string | null;
		actorParty: string | null;
		partySize: string | null;
	}>`
		SELECT
			id,
			besluit_id as "decisionId",
			soort as type,
			fractie_id as "partyId",
			actor_fractie as "actorParty",
			fractie_grootte as "partySize"
		FROM stemmingen
		WHERE besluit_id = ANY(${decisionIds})
	`;

	const partyIdsFromVotes = votes
		.map((v) => v.partyId)
		.filter((p) => p !== null) as string[];
	const partyNamesFromVotes = votes
		.map((v) => v.actorParty)
		.filter((p) => p !== null) as string[];

	const parties =
		partyIdsFromVotes.length > 0 || partyNamesFromVotes.length > 0
			? await sql<{
					id: string;
					number: string | null;
					shortName: string | null;
					nameNl: string | null;
					nameEn: string | null;
					seats: string | null;
					votesCount: string | null;
					activeFrom: Date | null;
					activeTo: Date | null;
					contentType: string | null;
					contentLength: string | null;
					updatedAt: Date | null;
					apiUpdatedAt: Date | null;
					logoData: any | null;
					removed: boolean | null;
				}>`
			SELECT
				id,
				nummer as number,
				afkorting as "shortName",
				naam_nl as "nameNl",
				naam_en as "nameEn",
				aantal_zetels as seats,
				aantal_stemmen as "votesCount",
				datum_actief as "activeFrom",
				datum_inactief as "activeTo",
				content_type as "contentType",
				content_length as "contentLength",
				gewijzigd_op as "updatedAt",
				api_gewijzigd_op as "apiUpdatedAt",
				logo_data as "logoData",
				verwijderd as removed
			FROM fracties
			WHERE id = ANY(${partyIdsFromVotes})
			   OR naam_nl = ANY(${partyNamesFromVotes})
		`
			: [];

	const partyNameMap = new Map<string, PartyMapped>();
	parties.forEach((p) => {
		if (p.nameNl) partyNameMap.set(p.nameNl, p);
	});

	return answers.map((answer) => {
		const motion = motions.find((m) => m.id === answer.motionId);
		const decision = decisions.find((d) => d.caseId === answer.motionId);

		const motionVotes = decision
			? votes.filter((v) => v.decisionId === decision.id)
			: [];

		const partyVotes = new Map<
			string,
			{
				party: PartyMapped;
				votes: VoteType[];
				majorityVote: VoteType;
				partySize: number;
			}
		>();

		motionVotes.forEach((vote) => {
			const party = vote.partyId
				? parties.find((p) => p.id === vote.partyId)
				: vote.actorParty
					? partyNameMap.get(vote.actorParty)
					: undefined;

			if (party?.id && vote.type) {
				const partyId = party.id;
				if (!partyVotes.has(partyId)) {
					partyVotes.set(partyId, {
						party: party,
						votes: [],
						majorityVote: "NEUTRAL",
						partySize: Number(vote.partySize || 0),
					});
				}
				partyVotes.get(partyId)?.votes.push(mapVoteType(vote.type));
			}
		});

		partyVotes.forEach((partyData, _partyId) => {
			const voteCounts = partyData.votes.reduce(
				(acc, vote) => {
					acc[vote as VoteType] = (acc[vote as VoteType] || 0) + 1;
					return acc;
				},
				{} as Record<VoteType, number>,
			);

			const majorityVote =
				Object.keys(voteCounts).length > 0
					? Object.entries(voteCounts).reduce((a, b) =>
							a[1] > b[1] ? a : b,
						)[0]
					: null;

			partyData.majorityVote = (majorityVote ?? "NEUTRAL") as VoteType;
		});

		const partyPositions = Array.from(partyVotes.values()).map(
			({ party, majorityVote, partySize }) => ({
				party: mapPartyToContract(party),
				position: majorityVote,
				voteCount: partySize,
				agreesWithUser:
					majorityVote !== "NEUTRAL" &&
					((answer.answer === "agree" && majorityVote === "FOR") ||
						(answer.answer === "disagree" && majorityVote === "AGAINST") ||
						answer.answer === "neutral"),
			}),
		);

		return {
			motionId: answer.motionId,
			userAnswer: answer.answer,
			motion: motion ? mapCaseToMotion(motion) : null,
			partyPositions,
		};
	});
}

export interface PartyMapped {
	id: string;
	number: bigint | null;
	shortName: string | null;
	nameNl: string | null;
	nameEn: string | null;
	seats: bigint | null;
	votesCount: bigint | null;
	activeFrom: Date | null;
	activeTo: Date | null;
	contentType: string | null;
	contentLength: bigint | null;
	updatedAt: Date | null;
	apiUpdatedAt: Date | null;
	logoData: Buffer | null;
	removed: boolean | null;
}
