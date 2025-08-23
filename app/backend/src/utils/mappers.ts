import { Case, ParliamentaryDocument, Party, Politician, Vote } from "@prisma/client";
import { Motion as MotionContract, Party as PartyContract, Politician as PoliticianContract, Vote as VoteContract } from "../contracts/index.js";

export function mapCaseToMotion(zaak: Case, dossier?: ParliamentaryDocument | null): MotionContract {
    return {
        id: zaak.id,
        title: zaak.title || zaak.subject || "Untitled Motion",
        description: zaak.subject,
        shortTitle: zaak.citationTitle,
        motionNumber: zaak.number,
        date: zaak.date,
        status: zaak.status || "unknown",
        category: zaak.type,
        bulletPoints:
            dossier &&
            dossier.bulletPoints != null &&
            Array.isArray(dossier.bulletPoints)
                ? dossier.bulletPoints.filter(
                    (bp): bp is string => typeof bp === "string",
                )
                : [],
        originalId: zaak.id,
        createdAt: zaak.startedAt || new Date(),
        updatedAt: zaak.updatedAt || new Date(),
    };
}

export function mapPartyToContract(party: Party): PartyContract {
    return {
        id: party.id,
        name: party.nameNl || party.shortName || "",
        shortName: party.shortName || "",
        color: null, // Not available in your schema
        seats: Number(party.seats) || 0,
        activeFrom: party.activeFrom,
        activeTo: party.activeTo,
        createdAt: party.updatedAt || new Date(),
        updatedAt: party.apiUpdatedAt || new Date(),
    };
}

export function mapPoliticianToContract(politician: Politician): PoliticianContract {
    return {
        id: politician.id,
        firstName: politician.firstNames || "",
        lastName: politician.lastName || "",
        fullName:
            `${politician.firstNames || ""} ${politician.prefix || ""} ${politician.lastName || ""}`.trim(),
        partyId: undefined, // This needs to be handled separately
        createdAt: politician.updatedAt || new Date(),
        updatedAt: politician.updatedAt || new Date(),
    };
}

export function mapVoteToContract(vote: Vote & { party?: Party | null, politician?: Politician | null, decision?: { case?: Case | null } | null }): VoteContract {
    const motion = vote.decision?.case ? mapCaseToMotion(vote.decision.case) : undefined;
    return {
        id: vote.id,
        motionId: motion?.id || "",
        partyId: vote.partyId || "",
        politicianId: vote.politicianId || "",
        voteType: vote.type,
        reasoning: null, // Not available in your schema
        createdAt: vote.updatedAt || new Date(),
        updatedAt: vote.apiUpdatedAt || new Date(),
        motion: motion,
        party: vote.party ? mapPartyToContract(vote.party) : undefined,
        politician: vote.politician ? mapPoliticianToContract(vote.politician) : undefined,
    };
}