import { oc } from "@orpc/contract";
import { z } from "zod";

const VoteTypeSchema = z.enum(["FOR", "AGAINST", "NEUTRAL"]);

const PartySchema = z.object({
  id: z.string(),
  name: z.string(),
  shortName: z.string(),
  seats: z.number(),
  activeFrom: z.coerce.date().nullable(),
  activeTo: z.coerce.date().nullable(),
  contentType: z.string().nullable(),
  logoData: z.string().nullable(), //base64
  createdAt: z.coerce.date(),
  updatedAt: z.coerce.date(),
});

const PoliticianSchema = z.object({
  id: z.string(),
  firstName: z.string(),
  lastName: z.string(),
  fullName: z.string(),
  partyId: z.string().optional(),
  createdAt: z.coerce.date().optional(),
  updatedAt: z.coerce.date().optional(),
  party: PartySchema.optional(),
});

const MotionCategorySchema = z.object({
  id: z.string(),
  name: z.string(),
  type: z.enum(["generic", "hot_topic"]),
  description: z.string().nullable(),
  keywords: z.array(z.string()),
  createdAt: z.coerce.date(),
  updatedAt: z.coerce.date(),
});

const MotionSchema = z.object({
  id: z.string(),
  title: z.string(),
  description: z.string().nullish(),
  shortTitle: z.string().nullish(),
  motionNumber: z.string().nullish(),
  status: z.string(),
  category: z.string().nullish(),
  bulletPoints: z
    .array(z.string())
    .nullish()
    .transform((value) => value ?? []),
  categories: z.array(MotionCategorySchema).optional(),
  originalId: z.string().nullish(),
  documentURL: z.string().nullish(),
  did: z.string().nullish(),
  createdAt: z.coerce.date(),
  updatedAt: z.coerce.date(),
});

const VoteSchema = z.object({
  id: z.string(),
  motionId: z.string(),
  partyId: z.string(),
  politicianId: z.string(),
  voteType: VoteTypeSchema,
  reasoning: z.string().nullable(),
  createdAt: z.coerce.date(),
  updatedAt: z.coerce.date(),
  motion: MotionSchema.optional(),
  party: PartySchema.optional(),
  politician: PoliticianSchema.optional(),
});

const DecisionSchema = z.object({
  id: z.string(),
  agendaPointId: z.string().nullable(),
  caseId: z.string().nullable(),
  votingType: z.string().nullable(),
  decisionType: z.string().nullable(),
  decisionText: z.string().nullable(),
  comment: z.string().nullable(),
  status: z.string().nullable(),
  agendaPointCaseDecisionOrder: z.bigint().nullable(),
  updatedAt: z.coerce.date().nullable(),
  apiUpdatedAt: z.coerce.date().nullable(),
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

const MotionDetailSchema = z.object({
  motionId: z.string(),
  userAnswer: z.enum(["agree", "disagree", "neutral"]),
  motion: MotionSchema.nullable(),
  partyPositions: z.array(
    z.object({
      party: PartySchema,
      position: VoteTypeSchema,
      voteCount: z.number(),
      agreesWithUser: z.boolean(),
    }),
  ),
});

export const CompassResultSchema = z.object({
  id: z.string(),
  totalAnswers: z.number(),
  partyResults: z.array(PartyResultSchema),
  motionDetails: z.array(MotionDetailSchema).optional(),
  createdAt: z.coerce.date(),
});

export const UserSessionSchema = z.object({
  id: z.string(),
  answers: UserAnswerSchema,
  results: CompassResultSchema,
  createdAt: z.coerce.date(),
  updatedAt: z.coerce.date(),
});

// Motion contracts
const motionGetAllContract = oc
  .input(
    z.object({
      limit: z.number().min(1).max(100).default(20),
      offset: z.number().min(0).default(0),
      category: z.string().optional(),
      status: z.string().optional(),
      withVotes: z.boolean().optional(),
      search: z.string().optional(),
      partyIds: z.array(z.string()).optional(),
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
      categoryIds: z.array(z.string()).optional(),
      after: z.coerce.date().optional(),
      search: z.string().optional(),
      partyIds: z.array(z.string()).optional(),
    }),
  )
  .output(z.array(MotionSchema));

const motionGetForCompassCountContract = oc
  .input(
    z.object({
      categoryIds: z.array(z.string()).optional(),
      after: z.coerce.date().optional(),
      search: z.string().optional(),
      partyIds: z.array(z.string()).optional(),
    }),
  )
  .output(z.object({ count: z.number() }));

const motionGetCategoriesContract = oc
  .input(
    z.object({ type: z.enum(["generic", "hot_topic", "all"]).default("all") }),
  )
  .output(z.array(MotionCategorySchema));

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

const motionGetRecentContract = oc
  .input(z.object({ limit: z.number().min(1).max(50).default(10) }))
  .output(z.array(MotionSchema));

// Party contracts
const partyGetAllContract = oc
  .input(z.object({ activeOnly: z.boolean().default(true) }))
  .output(z.array(PartySchema));

const partyGetInRangeContract = oc
  .input(
    z.object({
      dateFrom: z.coerce.date(),
      dateTo: z.coerce.date(),
    }),
  )
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

const motionGetStatisticsContract = oc.output(
  z.object({
    count: z.number(),
    firstMotionDate: z.coerce.date().nullable(),
    lastMotionDate: z.coerce.date().nullable(),
  }),
);

const dateFilterSchema = z.object({
  dateFrom: z.coerce.date().optional(),
  dateTo: z.coerce.date().optional(),
});

// For Party Likeness Matrix
export const PartyLikenessSchema = z.object({
  party1Id: z.string(),
  party1Name: z.string(),
  party2Id: z.string(),
  party2Name: z.string(),
  commonMotions: z.number(),
  sameVotes: z.number(),
  likenessPercentage: z.number(),
});

// For Party Focus
export const PartyFocusCategorySchema = z.object({
  categoryId: z.string(),
  categoryName: z.string(),
  categoryType: z.string().nullable(),
  motionCount: z.coerce.number(),
});

export const PartyFocusSchema = z.object({
  party: PartySchema,
  categories: z.array(PartyFocusCategorySchema),
});

// For Party-Category Likeness
export const PartyCategoryLikenessSchema = z.object({
  categoryId: z.string(),
  categoryName: z.string(),
  party2Id: z.string(),
  party2Name: z.string(),
  likenessPercentage: z.number(),
});

// New statistics contracts
const statisticsGetPartyLikenessMatrixContract = oc
  .input(dateFilterSchema.optional())
  .output(z.array(PartyLikenessSchema));

const statisticsGetPartyFocusContract = oc
  .input(z.object({ partyId: z.string() }).merge(dateFilterSchema))
  .output(PartyFocusSchema.nullable());

const statisticsGetPartyCategoryLikenessContract = oc
  .input(z.object({ partyId: z.string() }).merge(dateFilterSchema))
  .output(z.array(PartyCategoryLikenessSchema));

export const apiContract = {
  motions: {
    getAll: motionGetAllContract,
    getById: motionGetByIdContract,
    getForCompass: motionGetForCompassContract,
    getForCompassCount: motionGetForCompassCountContract,
    getCategories: motionGetCategoriesContract,
    getStatistics: motionGetStatisticsContract,
    getVotes: motionGetVotesContract,
    getRecent: motionGetRecentContract,
  },
  parties: {
    getAll: partyGetAllContract,
    getInRange: partyGetInRangeContract,
    getById: partyGetByIdContract,
    getWithVotes: partyGetWithVotesContract,
  },
  compass: {
    submitAnswers: compassSubmitAnswersContract,
    getResults: compassGetResultsContract,
    getMotionDetails: compassGetMotionDetailsContract,
  },
  statistics: {
    getPartyLikenessMatrix: statisticsGetPartyLikenessMatrixContract,
    getPartyFocus: statisticsGetPartyFocusContract,
    getPartyCategoryLikeness: statisticsGetPartyCategoryLikenessContract,
  },
};

// Type exports
export type ApiContract = typeof apiContract;
export type VoteType = z.infer<typeof VoteTypeSchema>;
export type Party = z.infer<typeof PartySchema>;
export type Politician = z.infer<typeof PoliticianSchema>;
export type MotionCategory = z.infer<typeof MotionCategorySchema>;
export type Motion = z.infer<typeof MotionSchema>;
export type Decision = z.infer<typeof DecisionSchema>;
export type Vote = z.infer<typeof VoteSchema>;
export type UserAnswer = z.infer<typeof UserAnswerSchema>;
export type PartyResult = z.infer<typeof PartyResultSchema>;
export type MotionDetail = z.infer<typeof MotionDetailSchema>;
export type CompassResult = z.infer<typeof CompassResultSchema>;
export type PartyLikeness = z.infer<typeof PartyLikenessSchema>;
export type PartyFocus = z.infer<typeof PartyFocusSchema>;
export type PartyFocusCategory = z.infer<typeof PartyFocusCategorySchema>;
export type PartyCategoryLikeness = z.infer<typeof PartyCategoryLikenessSchema>;
export type UserSession = z.infer<typeof UserSessionSchema>;
