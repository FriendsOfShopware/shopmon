import { router } from '../index.ts';
import { shopRouter } from './shop.ts';

export const organizationRouter = router({
    shop: shopRouter,
});
