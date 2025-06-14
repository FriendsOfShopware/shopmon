import { router } from '../index.ts';
import { projectRouter } from './project.ts';
import { shopRouter } from './shop.ts';
import { ssoRouter } from './sso.ts';

export const organizationRouter = router({
    shop: shopRouter,
    sso: ssoRouter,
    project: projectRouter,
});
