import { ref, type Ref } from "vue";
import type { AccountChangelog } from "@/api/generated";

export function useEnvironmentChangelogModal() {
  const viewEnvironmentChangelogDialog: Ref<boolean> = ref(false);
  const dialogEnvironmentChangelog: Ref<AccountChangelog | null> = ref(null);

  function openEnvironmentChangelog(environmentChangelog: AccountChangelog | null) {
    dialogEnvironmentChangelog.value = environmentChangelog;
    viewEnvironmentChangelogDialog.value = true;
  }

  function closeEnvironmentChangelog() {
    viewEnvironmentChangelogDialog.value = false;
    dialogEnvironmentChangelog.value = null;
  }

  return {
    viewEnvironmentChangelogDialog,
    dialogEnvironmentChangelog,
    openEnvironmentChangelog,
    closeEnvironmentChangelog,
  };
}
