import { implement } from "@orpc/server";
import { type Party as PartyModel, Prisma } from "@prisma/client";
import { apiContract, type VoteType } from "../contracts/index.js";
import { db } from "../lib/db.js";
import {
	mapCaseToMotion,
	mapPartyToContract,
	mapVoteToContract,
} from "../utils/mappers.js";

const os = implement(apiContract);

export const motionRouter = {
	getAll: os.motions.getAll.handler(async ({ input }) => {
		const { limit, offset, category, status } = input;

		const where: Prisma.CaseWhereInput = {
			type: "Motie",
		};
		if (category) {
			where.type = category;
		}
		if (status) {
			where.status = status;
		}

		const [cases, total] = await Promise.all([
			db.case.findMany({
				where,
				orderBy: { date: "desc" },
				take: limit,
				skip: offset,
			}),
			db.case.count({ where }),
		]);

		const motions = cases.map((c) => {
			return mapCaseToMotion(c);
		});

		return {
			motions,
			total,
			hasMore: offset + limit < total,
		};
	}),

	getById: os.motions.getById.handler(async ({ input }) => {
		const c = await db.case.findUnique({
			where: { id: input.id },
		});

		if (!c) {
			return null;
		}

		return mapCaseToMotion(c);
	}),

	getForCompass: os.motions.getForCompass.handler(async ({ input }) => {
		const { count, excludeIds = [], categoryIds, after } = input;

		const whereConditions = [
			Prisma.sql`"soort" = 'Motie'`,
			Prisma.sql`"bullet_points" IS NOT NULL AND jsonb_array_length("bullet_points") > 0`,
		];

		if (excludeIds.length > 0) {
			whereConditions.push(
				Prisma.sql`id NOT IN (${Prisma.join(
					excludeIds.map((id) => Prisma.sql`${id}`),
					",",
				)})`,
			);
		}

		if (after) {
			whereConditions.push(Prisma.sql`"gestart_op" >= ${after}`);
		}

		if (categoryIds && categoryIds.length > 0) {
			whereConditions.push(
				Prisma.sql`EXISTS (
                SELECT 1 FROM "zaak_categories"
                WHERE "zaak_id" = "zaken".id
                AND "category_id" IN (${Prisma.join(
									categoryIds.map((id) => Prisma.sql`${id}`),
									",",
								)})
            )`,
			);
		}

		const randomCaseIds = await db.$queryRaw<Array<{ id: string }>>`
        SELECT id FROM "zaken"
        WHERE ${Prisma.join(whereConditions, " AND ")}
        ORDER BY RANDOM()
        LIMIT ${count}
    `;

		const cases = await db.case.findMany({
			where: {
				id: {
					in: randomCaseIds.map((c) => c.id),
				},
			},
			include: {
				caseCategories: {
					include: {
						category: true,
					},
				},
			},
		});

		const motions = cases.map((c) => {
			const motion = mapCaseToMotion(c);

			if (c.caseCategories) {
				motion.categories = c.caseCategories
					.filter((cc) => cc.category)
					.map((cc) => ({
						id: cc.category.id,
						name: cc.category.name,
						type: cc.category.type as "generic" | "hot_topic",
						description: cc.category.description,
						keywords: cc.category.keywords,
						createdAt: cc.category.createdAt || new Date(),
						updatedAt: cc.category.updatedAt || new Date(),
					}));
			}

			return motion;
		});

		return motions;
	}),

	getCategories: os.motions.getCategories.handler(async ({ input }) => {
		const { type } = input;

		const where: Prisma.MotionCategoryWhereInput = {};
		if (type !== "all") {
			where.type = type;
		}

		const categories = await db.motionCategory.findMany({
			where,
			orderBy: [
				{ type: "asc" }, // hot_topic first, then generic
				{ name: "asc" },
			],
		});

		return categories.map((cat) => ({
			id: cat.id,
			name: cat.name,
			type: cat.type as "generic" | "hot_topic",
			description: cat.description,
			keywords: cat.keywords,
			createdAt: cat.createdAt || new Date(),
			updatedAt: cat.updatedAt || new Date(),
		}));
	}),

	getVotes: os.motions.getVotes.handler(async ({ input }) => {
		const votesWithRelations = await db.vote.findMany({
			where: {
				decisionId: input.motionId,
			},
			orderBy: [{ partyId: "asc" }, { politicianId: "asc" }],
		});

		const votes = votesWithRelations.map((v) => mapVoteToContract(v));

		const partyVoteMap = new Map<
			string,
			{ party: PartyModel; votes: string[] }
		>();

		const parties = await db.party.findMany({
			where: {
				id: {
					in: votesWithRelations
						.map((v) => v.partyId)
						.filter((p) => p !== null) as string[],
				},
			},
		});

		votesWithRelations.forEach((vote) => {
			if (vote.partyId) {
				const party = parties.find((p) => p.id === vote.partyId);
				if (party) {
					if (!partyVoteMap.has(vote.partyId)) {
						partyVoteMap.set(vote.partyId, {
							party: party,
							votes: [],
						});
					}
					if (vote.type) {
						partyVoteMap.get(vote.partyId)?.votes.push(vote.type);
					}
				}
			}
		});

		const partyPositions = Array.from(partyVoteMap.values()).map(
			({ party, votes: partyVotes }) => {
				if (partyVotes.length === 0) {
					return {
						party: mapPartyToContract(party),
						position: "NEUTRAL" as VoteType,
						count: 0,
					};
				}
				const voteCounts = partyVotes.reduce(
					(acc, vote) => {
						acc[vote] = (acc[vote] || 0) + 1;
						return acc;
					},
					{} as Record<string, number>,
				);

				const majorityVoteEntry = Object.entries(voteCounts).reduce((a, b) =>
					a[1] > b[1] ? a : b,
				);

				const majorityVote = majorityVoteEntry[0];
				const count = majorityVoteEntry[1];

				return {
					party: mapPartyToContract(party),
					position: majorityVote as VoteType,
					count,
				};
			},
		);

		return { votes, partyPositions };
	}),

	getStatistics: os.motions.getStatistics.handler(async () => {
		const statistics = await db.case.aggregate({
			where: {
				type: "Motie",
			},
			_count: {
				id: true,
			},
			_min: {
				startedAt: true,
			},
			_max: {
				startedAt: true,
			},
		});

		return {
			count: statistics._count.id,
			lastMotionDate: statistics._max.startedAt,
			firstMotionDate: statistics._min.startedAt,
		};
	}),
};