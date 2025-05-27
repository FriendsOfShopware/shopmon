import { initTRPC } from '@trpc/server';
import type { context } from './context.ts';

export const t = initTRPC.context<context>().create();

export const publicProcedure = t.procedure;
export const router = t.router;
