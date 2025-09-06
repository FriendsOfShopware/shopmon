import { ref, type Ref } from 'vue';

interface ExtensionChangelog {
    version: string;
    text: string;
    creationDate: string;
    isCompatible: boolean;
}

interface Extension {
    name: string;
    label: string;
    changelog: ExtensionChangelog[] | null;
}

export function useExtensionChangelogModal() {
    const viewExtensionChangelogDialog: Ref<boolean> = ref(false);
    const dialogExtension: Ref<Extension | null> = ref(null);

    function openExtensionChangelog(extension: Extension | null) {
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