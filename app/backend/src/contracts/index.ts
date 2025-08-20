import { oc } from "@orpc/contract";
import { z } from "zod";

// TODO narrow down vote types
const VoteTypeSchema = z.string();

const PartySchema = z.object({
	id: z.string(),
	name: z.string(),
	shortName: z.string(),
	color: z.string().nullable(),
	seats: z.number(),
	activeFrom: z.date().nullable(),
	activeTo: z.date().nullable(),
	createdAt: z.date(),
	updatedAt: z.date(),
});

const PoliticianSchema = z.object({
	id: z.string(),
	firstName: z.string(),
	lastName: z.string(),
	fullName: z.string(),
	partyId: z.string().optional(),
	createdAt: z.date().optional(),
	updatedAt: z.date().optional(),
	party: PartySchema.optional(),
});

const MotionSchema = z.object({
	id: z.string(),
	title: z.string(),
	description: z.string().nullable(),
	shortTitle: z.string().nullable(),
	motionNumber: z.string().nullable(),
	date: z.date().nullable(),
	status: z.string(),
	category: z.string().nullable(),
	bulletPoints: z.array(z.string()),
	originalId: z.string().nullable(),
	createdAt: z.date(),
	updatedAt: z.date(),
});

const VoteSchema = z.object({
	id: z.string(),
	motionId: z.string(),
	partyId: z.string(),
	politicianId: z.string(),
	voteType: VoteTypeSchema,
	reasoning: z.string().nullable(),
	createdAt: z.date(),
	updatedAt: z.date(),
	motion: MotionSchema.optional(),
	party: PartySchema.optional(),
	politician: PoliticianSchema.optional(),
});

const UserAnswerSchema = z.object({
	motionId: z.string(),
	answer: z.enum(["agree", "disagree", "neutral"]),
});

const PartyResultSchema = z.object({
	party: PartySchema,
	score: z.number(),
	agreement: z.number(),
	totalVotes: z.number(),
	matchingVotes: z.number(),
});

const CompassResultSchema = z.object({
	id: z.string(),
	totalAnswers: z.number(),
	partyResults: z.array(PartyResultSchema),
	createdAt: z.date(),
});

// Motion contracts
const motionGetAllContract = oc
	.input(
		z.object({
			limit: z.number().min(1).max(100).default(20),
			offset: z.number().min(0).default(0),
			category: z.string().optional(),
			status: z.string().optional(),
		}),
	)
	.output(
		z.object({
			motions: z.array(MotionSchema),
			total: z.number(),
			hasMore: z.boolean(),
		}),
	);

const motionGetByIdContract = oc
	.input(z.object({ id: z.string() }))
	.output(MotionSchema.nullable());

const motionGetForCompassContract = oc
	.input(
		z.object({
			count: z.number().min(1).max(50).default(20),
			excludeIds: z.array(z.string()).optional(),
		}),
	)
	.output(z.array(MotionSchema));

const motionGetVotesContract = oc
	.input(z.object({ motionId: z.string() }))
	.output(
		z.object({
			votes: z.array(VoteSchema),
			partyPositions: z.array(
				z.object({
					party: PartySchema,
					position: VoteTypeSchema,
					count: z.number(),
				}),
			),
		}),
	);

// Party contracts
const partyGetAllContract = oc
	.input(z.object({ activeOnly: z.boolean().default(true) }))
	.output(z.array(PartySchema));

const partyGetByIdContract = oc
	.input(z.object({ id: z.string() }))
	.output(PartySchema.nullable());

const partyGetWithVotesContract = oc
	.input(
		z.object({
			partyId: z.string(),
			motionIds: z.array(z.string()).optional(),
		}),
	)
	.output(
		z.object({
			party: PartySchema,
			votes: z.array(VoteSchema),
		}),
	);

// Compass contracts
const compassSubmitAnswersContract = oc
	.input(z.object({ answers: z.array(UserAnswerSchema) }))
	.output(CompassResultSchema);

const compassGetResultsContract = oc
	.input(z.object({ sessionId: z.string() }))
	.output(CompassResultSchema.nullable());

const compassGetMotionDetailsContract = oc
	.input(
		z.object({
			motionId: z.string(),
			includeVotes: z.boolean().default(true),
		}),
	)
	.output(
		z.object({
			motion: MotionSchema,
			votes: z.array(VoteSchema).optional(),
			partyPositions: z
				.array(
					z.object({
						party: PartySchema,
						position: VoteTypeSchema,
						count: z.number(),
					}),
				)
				.optional(),
		}),
	);

// Contract router
export const apiContract = {
	motions: {
		getAll: motionGetAllContract,
		getById: motionGetByIdContract,
		getForCompass: motionGetForCompassContract,
		getVotes: motionGetVotesContract,
	},
	parties: {
		getAll: partyGetAllContract,
		getById: partyGetByIdContract,
		getWithVotes: partyGetWithVotesContract,
	},
	compass: {
		submitAnswers: compassSubmitAnswersContract,
		getResults: compassGetResultsContract,
		getMotionDetails: compassGetMotionDetailsContract,
	},
};

// Type exports
export type ApiContract = typeof apiContract;
export type VoteType = z.infer<typeof VoteTypeSchema>;
export type Party = z.infer<typeof PartySchema>;
export type Politician = z.infer<typeof PoliticianSchema>;
export type Motion = z.infer<typeof MotionSchema>;
export type Vote = z.infer<typeof VoteSchema>;
export type UserAnswer = z.infer<typeof UserAnswerSchema>;
export type PartyResult = z.infer<typeof PartyResultSchema>;
export type CompassResult = z.infer<typeof CompassResultSchema>;
