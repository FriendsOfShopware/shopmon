interface ShopwareExtensionCompatibility {
  name: string;
  label: string;
  iconPath: string;
  status: {
    name: string;
    label: string;
    type: string;
  };
}

interface ShopwareVersion {
  [version: string]: string[];
}

export const checkExtensionCompatibility = async (
  currentVersion: string,
  futureVersion: string,
  extensions: { name: string; version: string }[],
) => {
  const url = new URL("https://api.shopware.com/swplatform/autoupdate");
  url.searchParams.set("language", "en-GB");
  url.searchParams.set("shopwareVersion", currentVersion);

  const checkExtensionCompatibilityApiResp = await fetch(url.toString(), {
    method: "POST",
    body: JSON.stringify({
      futureShopwareVersion: futureVersion,
      plugins: extensions,
    }),
  });

  return (await checkExtensionCompatibilityApiResp.json()) as ShopwareExtensionCompatibility[];
};

export const getLatestShopwareVersion = async () => {
  const installApiResp = await fetch(
    "https://raw.githubusercontent.com/FriendsOfShopware/shopware-static-data/main/data/all-supported-php-versions-by-shopware-version.json",
    {
      method: "GET",
      headers: {
        Accept: "application/json",
        "User-Agent": "Shopmon",
      },
    },
  );

  return (await installApiResp.json()) as ShopwareVersion[];
};
