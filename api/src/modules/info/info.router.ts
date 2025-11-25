import { z } from 'zod';
import { publicProcedure, router } from '#src/trpc/index.ts';
import * as InfoService from './info.service.ts';

export const infoRouter = router({
    checkExtensionCompatibility: publicProcedure
        .input(
            z.object({
                currentVersion: z.string(),
                futureVersion: z.string(),
                extensions: z.array(
                    z.object({
                        name: z.string(),
                        version: z.string(),
                    }),
                ),
            }),
        )
        .query(async ({ input }) => {
            return await InfoService.checkExtensionCompatibility(
                input.currentVersion,
                input.futureVersion,
                input.extensions,
            );
        }),
    getLatestShopwareVersion: publicProcedure.query(async () => {
        return await InfoService.getLatestShopwareVersion();
    }),
});
