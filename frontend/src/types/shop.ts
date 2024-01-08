export interface ExtensionDiff {
    name: string,
    label: string,
    state: string,
    old_version: string | null,
    new_version: string | null,
    changelog: ExtensionChangelog[] | null,
    active: boolean,
}

export interface ExtensionChangelog {
    version: string
    text: string
    creationDate: string
    isCompatible: boolean;
}

export interface ShopChangelog {
    id: number;
    shop_id: number;
    extensions: ExtensionDiff[];
    old_shopware_version: string | null;
    new_shopware_version: string | null;
    date: string;
}
