import { ref, type Ref } from 'vue';
import type { ShopChangelog } from '@/types/shop';

export function useShopChangelogModal() {
    const viewShopChangelogDialog: Ref<boolean> = ref(false);
    const dialogShopChangelog: Ref<ShopChangelog | null> = ref(null);

    function openShopChangelog(shopChangelog: ShopChangelog | null) {
        dialogShopChangelog.value = shopChangelog;
        viewShopChangelogDialog.value = true;
    }

    function closeShopChangelog() {
        viewShopChangelogDialog.value = false;
        dialogShopChangelog.value = null;
    }

    return {
        viewShopChangelogDialog,
        dialogShopChangelog,
        openShopChangelog,
        closeShopChangelog,
    };
}