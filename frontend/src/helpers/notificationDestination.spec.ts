import { expect, it } from "vitest";
import { buildDestinationDSN, emptyConnection } from "./notificationDestination";

it("encodes original Slack webhook URLs exactly once", () => {
  const dsn = buildDestinationDSN({
    ...emptyConnection(),
    webhookUrl: " https://hooks.slack.com/services/T/B/a%2Fb ",
  });
  expect(new URL(dsn).searchParams.get("webhook_url")).toBe(
    "https://hooks.slack.com/services/T/B/a%2Fb",
  );
});
it("encodes token and channel delimiters without changing credentials", () => {
  const dsn = buildDestinationDSN({
    ...emptyConnection("telegram"),
    token: " 123:a@b/?#% ",
    channel: " @alerts ",
  });
  const url = new URL(dsn);
  expect(decodeURIComponent(url.username)).toBe("123:a@b/?#%");
  expect(url.searchParams.get("channel")).toBe("@alerts");
});
it("supports Slack bot tokens", () => {
  expect(
    buildDestinationDSN({
      ...emptyConnection(),
      slackMode: "bot",
      token: "xoxb-secret",
      channel: "C123",
    }),
  ).toBe("slack://xoxb-secret@default?channel=C123");
});
it("extracts Discord credentials from a copied webhook", () => {
  expect(
    buildDestinationDSN({
      ...emptyConnection("discord"),
      webhookUrl: "https://discord.com/api/webhooks/123/token",
    }),
  ).toBe("discord://token@default?webhook_id=123");
});
it("normalizes a Gotify URL while preserving a custom port", () => {
  expect(
    buildDestinationDSN({
      ...emptyConnection("gotify"),
      serverUrl: " https://NOTIFY.example.com:8443/ ",
      token: "a@b",
    }),
  ).toBe("gotify://a%40b@notify.example.com:8443");
});
it("extracts the Teams connector ID and full token path", () => {
  const dsn = buildDestinationDSN({
    ...emptyConnection("microsoftteams"),
    webhookUrl: "https://outlook.office.com/webhook/id/IncomingWebhook/token/part",
  });
  expect(new URL(dsn).username).toBe("id");
  expect(new URL(dsn).searchParams.get("token")).toBe("token/part");
});
it.each([
  { ...emptyConnection(), webhookUrl: "https://hooks.slack.com.evil.test/services/T/B/token" },
  { ...emptyConnection(), webhookUrl: "http://hooks.slack.com/services/T/B/token" },
  { ...emptyConnection(), webhookUrl: "https://user:pass@hooks.slack.com/services/T/B/token" },
  { ...emptyConnection("discord"), webhookUrl: "https://example.com/api/webhooks/123/token" },
  { ...emptyConnection("gotify"), serverUrl: "https://example.com/gotify", token: "secret" },
  {
    ...emptyConnection("microsoftteams"),
    webhookUrl: "https://example.logic.azure.com/workflows/secret",
  },
])("rejects unsupported URLs without exposing credentials", (connection) => {
  expect(() => buildDestinationDSN(connection)).toThrow();
});
