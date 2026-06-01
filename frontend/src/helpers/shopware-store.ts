export function shopwareStoreSearchUrl(extensionName: string): string {
  return `https://store.shopware.com/en/search?search=${encodeURIComponent(extensionName)}`;
}
