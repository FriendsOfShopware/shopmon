export interface ExtensionDiff {
    name: string,
    label: string,
    state: string,
    oldVersion: string | null,
    newVersion: string | null,
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
    shopId: number;
    shopOrganizationId: number;
    extensions: ExtensionDiff[];
    oldShopwareVersion: string | null;
    newShopwareVersion: string | null;
    date: string;
}
