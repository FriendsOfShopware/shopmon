export interface PluginConnectionData {
  url: string;
  clientId: string;
  clientSecret: string;
}

export function parsePluginData(rawInput: string): PluginConnectionData {
  let input = rawInput.trim();
  if (!input) {
    throw new Error("EMPTY_INPUT");
  }

  // Remove surrounding single or double quotes if present
  if (
    (input.startsWith('"') && input.endsWith('"')) ||
    (input.startsWith("'") && input.endsWith("'"))
  ) {
    input = input.slice(1, -1).trim();
  }

  if (!input) {
    throw new Error("EMPTY_INPUT");
  }

  let jsonString = "";
  if (input.startsWith("{")) {
    jsonString = input;
  } else {
    // Remove all internal whitespace/newlines
    let cleaned = input.replace(/\s+/g, "");
    // Convert URL-safe base64 characters (- -> +, _ -> /)
    cleaned = cleaned.replace(/-/g, "+").replace(/_/g, "/");
    // Pad base64 string if necessary
    while (cleaned.length % 4 !== 0) {
      cleaned += "=";
    }
    try {
      jsonString = atob(cleaned);
    } catch {
      throw new Error("INVALID_FORMAT");
    }
  }

  let data: Record<string, unknown>;
  try {
    data = JSON.parse(jsonString);
  } catch {
    throw new Error("INVALID_FORMAT");
  }

  const url =
    data.url || data.shopUrl || data.shop_url || data.apiUrl || data.api_url || data.endpoint;
  const clientId = data.clientId || data.client_id || data.accessKey || data.access_key;
  const clientSecret =
    data.clientSecret || data.client_secret || data.secretAccessKey || data.secret_access_key;

  if (
    typeof url !== "string" ||
    !url ||
    typeof clientId !== "string" ||
    !clientId ||
    typeof clientSecret !== "string" ||
    !clientSecret
  ) {
    throw new Error("INVALID_DATA");
  }

  return {
    url,
    clientId,
    clientSecret,
  };
}
