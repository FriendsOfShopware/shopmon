export type DestinationProvider = "slack" | "telegram" | "discord" | "gotify" | "microsoftteams";
export type DestinationConnection = {
  provider: DestinationProvider;
  slackMode: "webhook" | "bot";
  webhookUrl: string;
  token: string;
  channel: string;
  serverUrl: string;
};

export function emptyConnection(provider: DestinationProvider = "slack"): DestinationConnection {
  return { provider, slackMode: "webhook", webhookUrl: "", token: "", channel: "", serverUrl: "" };
}

// Encode individual values, never the whole DSN. Users paste original credentials
// and URLs; go-notifier decodes each value exactly once.
export function buildDestinationDSN(connection: DestinationConnection): string {
  const token = connection.token.trim();
  const channel = connection.channel.trim();
  const query = (values: Record<string, string>) => new URLSearchParams(values).toString();
  const url = (raw: string, error: string) => {
    try {
      const value = new URL(raw.trim());
      if (value.protocol !== "https:" || value.username || value.password || value.hash)
        throw new Error();
      return value;
    } catch {
      throw new Error(error);
    }
  };
  switch (connection.provider) {
    case "slack": {
      if (connection.slackMode === "bot") {
        if (!/^xox(b-|p-|a-2)/.test(token)) throw new Error("invalidSlackToken");
        if (!channel) throw new Error("requiredChannel");
        return `slack://${encodeURIComponent(token)}@default?${query({ channel })}`;
      }
      const webhook = url(connection.webhookUrl, "invalidSlackWebhook");
      if (
        webhook.host !== "hooks.slack.com" ||
        !/^\/services\/[^/]+\/[^/]+\/[^/]+$/.test(webhook.pathname) ||
        webhook.search
      )
        throw new Error("invalidSlackWebhook");
      return `slack://default?${query({ webhook_url: webhook.href })}`;
    }
    case "telegram":
      if (!token) throw new Error("requiredToken");
      if (!channel) throw new Error("requiredChannel");
      return `telegram://${encodeURIComponent(token)}@default?${query({ channel })}`;
    case "discord": {
      const webhook = url(connection.webhookUrl, "invalidDiscordWebhook");
      const match = webhook.pathname.match(/^\/api\/(?:v\d+\/)?webhooks\/(\d+)\/([^/]+)\/?$/);
      if (
        !["discord.com", "discordapp.com", "canary.discord.com", "ptb.discord.com"].includes(
          webhook.host,
        ) ||
        !match
      )
        throw new Error("invalidDiscordWebhook");
      return `discord://${encodeURIComponent(decodeURIComponent(match[2]!))}@default?${query({ webhook_id: match[1]! })}`;
    }
    case "gotify": {
      const server = url(connection.serverUrl, "invalidGotifyServer");
      if (server.pathname !== "/" || server.search) throw new Error("invalidGotifyServer");
      if (!token) throw new Error("requiredToken");
      return `gotify://${encodeURIComponent(token)}@${server.host}`;
    }
    case "microsoftteams": {
      const webhook = url(connection.webhookUrl, "invalidTeamsWebhook");
      const match = webhook.pathname.match(/^\/webhook\/([^/]+)\/IncomingWebhook\/(.+)$/);
      // The pinned go-notifier Teams factory supports this connector format.
      // Reject other URL formats rather than silently sending to another host.
      if (webhook.host !== "outlook.office.com" || !match || webhook.search)
        throw new Error("invalidTeamsWebhook");
      return `microsoftteams://${encodeURIComponent(decodeURIComponent(match[1]!))}@default?${query({ token: decodeURIComponent(match[2]!) })}`;
    }
  }
}
