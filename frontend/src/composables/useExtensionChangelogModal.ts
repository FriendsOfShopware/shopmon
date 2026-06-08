import { ref, type Ref } from "vue";

interface ExtensionChangelog {
  version: string;
  text: string;
  creationDate: string;
  isCompatible: boolean;
}

export interface ExtensionWithChangelog {
  name: string;
  label: string;
  changelog?: ExtensionChangelog[] | string | null;
}

export function useExtensionChangelogModal() {
  const viewExtensionChangelogDialog: Ref<boolean> = ref(false);
  const dialogExtension: Ref<ExtensionWithChangelog | null> = ref(null);

  function openExtensionChangelog(extension: ExtensionWithChangelog | null) {
    dialogExtension.value = extension;
    viewExtensionChangelogDialog.value = true;
  }

  function closeExtensionChangelog() {
    viewExtensionChangelogDialog.value = false;
  }

  return {
    viewExtensionChangelogDialog,
    dialogExtension,
    openExtensionChangelog,
    closeExtensionChangelog,
  };
}
