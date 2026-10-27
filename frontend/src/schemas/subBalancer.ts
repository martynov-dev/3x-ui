import { z } from 'zod';

export const SubBalancerStrategySchema = z.enum([
  'leastLoad',
  'leastPing',
  'random',
  'roundRobin',
  'failover',
]);
export type SubBalancerStrategy = z.infer<typeof SubBalancerStrategySchema>;

export const SubBalancerSchema = z.object({
  id: z.number(),
  remark: z.string(),
  strategy: SubBalancerStrategySchema,
  inboundIds: z.array(z.number()),
  audience: z.enum(['', 'all', 'selected']).optional(),
  clientEmails: z.array(z.string()).nullish(),
  primaryInboundId: z.number().optional(),
  fallbackInboundId: z.number().optional(),
  memberWeights: z.record(z.string(), z.number()).nullish(),
  sortOrder: z.number(),
  enabled: z.boolean(),
  createdAt: z.number().optional(),
  updatedAt: z.number().optional(),
});
export type SubBalancer = z.infer<typeof SubBalancerSchema>;

export const SubBalancerListSchema = z.array(SubBalancerSchema);

export const SubBalancerFormSchema = z
  .object({
    remark: z
      .string()
      .trim()
      .min(1, 'pages.settings.subBalancers.errRemarkRequired')
      .max(256, 'pages.settings.subBalancers.errRemarkRequired'),
    strategy: SubBalancerStrategySchema,
    inboundIds: z
      .array(z.number().int().positive())
      .min(1, 'pages.settings.subBalancers.errInboundsRequired'),
    // inboundId (stringified) -> leastLoad weight; absent members weigh 1.0.
    memberWeights: z
      .record(
        z.string(),
        z
          .number({ error: 'pages.settings.subBalancers.errWeightPositive' })
          .positive('pages.settings.subBalancers.errWeightPositive'),
      )
      .optional(),
    sortOrder: z
      .number({ error: 'pages.settings.subBalancers.errSortOrder' })
      .int('pages.settings.subBalancers.errSortOrder')
      .min(1, 'pages.settings.subBalancers.errSortOrder'),
    enabled: z.boolean(),
    audience: z.enum(['all', 'selected']).optional(),
    clientEmails: z.array(z.string().trim().min(1)).optional(),
    primaryInboundId: z.number().int().positive().optional(),
    fallbackInboundId: z.number().int().positive().optional(),
  })
  .superRefine((value, ctx) => {
    if (value.audience === 'selected' && !value.clientEmails?.length) {
      ctx.addIssue({
        code: 'custom',
        path: ['clientEmails'],
        message: 'pages.settings.subBalancers.errClientsRequired',
      });
    }
    if (
      value.strategy === 'failover' &&
      (value.inboundIds.length !== 2 ||
        !value.primaryInboundId ||
        !value.fallbackInboundId ||
        value.primaryInboundId === value.fallbackInboundId ||
        !value.inboundIds.includes(value.primaryInboundId) ||
        !value.inboundIds.includes(value.fallbackInboundId))
    ) {
      ctx.addIssue({
        code: 'custom',
        path: ['primaryInboundId'],
        message: 'pages.settings.subBalancers.errFailoverMembers',
      });
    }
  });
export type SubBalancerFormValues = z.infer<typeof SubBalancerFormSchema>;
