import { describe, it, expect, beforeEach, vi } from "vitest";

// Use the real locale catalogs so the test exercises actual keys.
vi.mock("@/i18n", async () => {
  const { createI18n } = await import("vue-i18n");
  const en = (await import("@/locales/en.json")).default;
  const de = (await import("@/locales/de.json")).default;
  return {
    i18n: createI18n({
      legacy: false,
      locale: "en",
      fallbackLocale: "en",
      messages: { en, de } as any,
    }),
  };
});

import { i18n } from "@/i18n";
import {
  notificationReasons,
  translateCheckMessage,
  translateNotificationMessage,
  translateNotificationTitle,
  translateReason,
} from "@/helpers/i18n";
import type { components } from "@/types/api";

type Check = components["schemas"]["EnvironmentCheck"];
type Notification = components["schemas"]["Notification"];

function check(partial: Partial<Check>): Check {
  return { id: "x", level: "green", message: "", ...partial } as Check;
}

describe("translateCheckMessage", () => {
  beforeEach(() => {
    (i18n.global.locale as unknown as { value: string }).value = "en";
  });

  it("renders a known key", () => {
    expect(translateCheckMessage(check({ messageKey: "check.frosh.phpOutdated" }))).toBe(
      "PHP version is outdated",
    );
  });

  it("interpolates params", () => {
    expect(
      translateCheckMessage(check({ messageKey: "check.task.overdue", params: { name: "foo" } })),
    ).toBe("Scheduled task 'foo' is overdue");
  });

  it("appends the recommendation suffix when current/recommended present", () => {
    expect(
      translateCheckMessage(
        check({
          messageKey: "check.frosh.phpOutdated",
          params: { current: "7.4", recommended: "8.2" },
        }),
      ),
    ).toBe("PHP version is outdated (current: 7.4, recommended: 8.2)");
  });

  it("falls back to the snippet param for an unknown key", () => {
    expect(
      translateCheckMessage(
        check({
          messageKey: "check.frosh.somethingCustom",
          params: { snippet: "somethingCustom" },
        }),
      ),
    ).toBe("somethingCustom");
  });

  it("falls back to the stored message when no key is set", () => {
    expect(translateCheckMessage(check({ message: "legacy English text" }))).toBe(
      "legacy English text",
    );
  });

  it("renders in the active locale", () => {
    (i18n.global.locale as unknown as { value: string }).value = "de";
    expect(translateCheckMessage(check({ messageKey: "check.frosh.phpOutdated" }))).toBe(
      "PHP-Version ist veraltet",
    );
  });
});

describe("translateNotification", () => {
  beforeEach(() => {
    (i18n.global.locale as unknown as { value: string }).value = "en";
  });

  function notification(partial: Partial<Notification>): Notification {
    return {
      id: 1,
      userId: "u",
      key: "k",
      level: "warning",
      title: "",
      message: "",
      link: null,
      read: false,
      createdAt: "2026-01-01T00:00:00Z",
      ...partial,
    } as Notification;
  }

  it("renders title/message from keys with params", () => {
    const n = notification({
      titleKey: "notification.statusDegraded.title",
      messageKey: "notification.statusDegraded.message",
      params: { name: "Shop", from: "green", to: "red" },
    });
    expect(translateNotificationTitle(n)).toBe("Environment: Shop status changed");
    expect(translateNotificationMessage(n)).toBe("Status changed from green to red");
  });

  it("falls back to stored title/message for legacy rows", () => {
    const n = notification({ title: "Legacy title", message: "Legacy message" });
    expect(translateNotificationTitle(n)).toBe("Legacy title");
    expect(translateNotificationMessage(n)).toBe("Legacy message");
  });

  it("extracts reasons from params", () => {
    const reasons = [{ level: "red", messageKey: "check.worker.enabled" }];
    const n = notification({ params: { reasons } });
    expect(notificationReasons(n)).toHaveLength(1);
    expect(translateReason(notificationReasons(n)[0])).toBe(
      "The admin worker is enabled. This can cause performance issues. Consider using the CLI worker instead.",
    );
  });

  it("returns no reasons when absent", () => {
    expect(notificationReasons(notification({}))).toEqual([]);
  });
});
