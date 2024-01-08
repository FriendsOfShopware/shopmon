import { initTRPC } from '@trpc/server';
import { context } from './context';

export const t = initTRPC.context<context>().create();

export const publicProcedure = t.procedure;
export const router = t.router;
