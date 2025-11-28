import type {
	Motion as MotionContract,
	Party as PartyContract,
	Politician as PoliticianContract,
	Vote as VoteContract,
	VoteType,
} from "../contracts/index.js";

export function mapCaseToMotion(zaak: {
	id: string;
	subject?: string | null;
	title?: string | null;
	citationTitle?: string | null;
	number?: string | null;
	status?: string | null;
	type?: string | null;
	bulletPoints?: unknown;
	startedAt?: Date | null;
	updatedAt?: Date | null;
}): MotionContract {
	const bulletPoints = (zaak.bulletPoints as string[]) || [];

	return {
		id: zaak.id,
		title: zaak.subject || zaak.title || "Untitled Motion",
		description: zaak.title,
		shortTitle: zaak.citationTitle,
		motionNumber: zaak.number,
		status: zaak.status ?? "unknown",
		category: zaak.type,
		bulletPoints:
			bulletPoints && Array.isArray(bulletPoints)
				? bulletPoints.filter((bp): bp is string => typeof bp === "string")
				: [],
		categories: undefined, // will be set by the caller if needed
		originalId: zaak.id,
		createdAt: zaak.startedAt ?? new Date(),
		updatedAt: zaak.updatedAt ?? new Date(),
	};
}

export function mapPartyToContract(party: {
	id: string;
	nameNl?: string | null;
	shortName?: string | null;
	seats?: string | number | null;
	contentType?: string | null;
	activeFrom?: Date | null;
	activeTo?: Date | null;
	logoData?: Buffer | string | null;
	updatedAt?: Date | null;
	apiUpdatedAt?: Date | null;
}): PartyContract {
	return {
		id: party.id,
		name: party.nameNl ?? party.shortName ?? "",
		shortName: party.shortName ?? "",
		seats: Number(party.seats ?? 0),
		contentType: party.contentType ?? "image/png",
		activeFrom: party.activeFrom ?? null,
		activeTo: party.activeTo ?? null,
		logoData: party.logoData
			? Buffer.isBuffer(party.logoData)
				? Buffer.from(party.logoData).toString("base64")
				: party.logoData
			: null,
		createdAt: party.updatedAt ?? new Date(),
		updatedAt: party.apiUpdatedAt ?? new Date(),
	};
}

export function mapPoliticianToContract(politician: {
	id: string;
	firstNames?: string | null;
	lastName?: string | null;
	prefix?: string | null;
	updatedAt?: Date | null;
}): PoliticianContract {
	return {
		id: politician.id,
		firstName: politician.firstNames || "",
		lastName: politician.lastName || "",
		fullName:
			`${politician.firstNames || ""} ${politician.prefix || ""} ${politician.lastName || ""}`.trim(),
		partyId: undefined, // this needs to be handled separately
		createdAt: politician.updatedAt || new Date(),
		updatedAt: politician.updatedAt || new Date(),
	};
}

function mapVoteTypeToContract(voteType: string | null): VoteType {
	switch (voteType) {
		case "Voor":
			return "FOR";
		case "Tegen":
			return "AGAINST";
		case "Niet deelgenomen":
			return "NEUTRAL";
		case "FOR":
		case "AGAINST":
		case "NEUTRAL":
			return voteType;
		default:
			return "NEUTRAL";
	}
}

export function mapVoteToContract(vote: {
	id: string;
	type?: string | null;
	partyId?: string | null;
	politicianId?: string | null;
	updatedAt?: Date | null;
	apiUpdatedAt?: Date | null;
	party?: {
		id: string;
		nameNl?: string | null;
		shortName?: string | null;
		seats?: string | number | null;
		contentType?: string | null;
		activeFrom?: Date | null;
		activeTo?: Date | null;
		logoData?: Buffer | string | null;
		updatedAt?: Date | null;
		apiUpdatedAt?: Date | null;
	} | null;
	politician?: {
		id: string;
		firstNames?: string | null;
		lastName?: string | null;
		prefix?: string | null;
		updatedAt?: Date | null;
	} | null;
	decision?: {
		case?: {
			id: string;
			subject?: string | null;
			title?: string | null;
			citationTitle?: string | null;
			number?: string | null;
			status?: string | null;
			type?: string | null;
			bulletPoints?: unknown;
			startedAt?: Date | null;
			updatedAt?: Date | null;
		} | null;
	} | null;
}): VoteContract {
	const motion = vote.decision?.case
		? mapCaseToMotion(vote.decision.case)
		: undefined;
	return {
		id: vote.id,
		motionId: motion?.id || "",
		partyId: vote.partyId || "",
		politicianId: vote.politicianId || "",
		voteType: mapVoteTypeToContract(vote.type ?? null),
		reasoning: null, // Not available in your schema
		createdAt: vote.updatedAt || new Date(),
		updatedAt: vote.apiUpdatedAt || new Date(),
		motion: motion,
		party: vote.party ? mapPartyToContract(vote.party) : undefined,
		politician: vote.politician
			? mapPoliticianToContract(vote.politician)
			: undefined,
	};
}
