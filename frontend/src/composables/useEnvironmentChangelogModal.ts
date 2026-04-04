import { ref, type Ref } from "vue";
import type { components } from "@/types/api";

type AccountChangelog = components["schemas"]["AccountChangelog"];

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
