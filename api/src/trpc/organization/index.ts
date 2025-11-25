import { router } from '#src/trpc/index.ts';
import { projectRouter } from './project.ts';
import { ssoRouter } from './sso.ts';

export const organizationRouter = router({
    project: projectRouter,
    sso: ssoRouter,
});
