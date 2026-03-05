import { beforeEach, describe, expect, it, vi } from "vitest";

describe("useWhatsNew", () => {
  beforeEach(() => {
    vi.resetModules();

    const store = new Map<string, string>();
    vi.stubGlobal("localStorage", {
      getItem: (key: string) => store.get(key) ?? null,
      setItem: (key: string, value: string) => store.set(key, String(value)),
      removeItem: (key: string) => store.delete(key),
      clear: () => store.clear(),
    });
  });

  it("shows the current release when it has not been seen yet", async () => {
    const { useWhatsNew } = await import("./useWhatsNew");

    const { isOpen } = useWhatsNew();

    expect(isOpen.value).toBe(true);
  });

  it("keeps the modal closed when the current release was already seen", async () => {
    const { WHATS_NEW_VERSION, useWhatsNew } = await import("./useWhatsNew");

    localStorage.setItem("shopmon-whats-new", WHATS_NEW_VERSION);

    vi.resetModules();

    const reloadedComposable = await import("./useWhatsNew");
    const { isOpen } = reloadedComposable.useWhatsNew();

    expect(isOpen.value).toBe(false);
  });

  it("stores the current release when dismissed", async () => {
    const { WHATS_NEW_VERSION, useWhatsNew } = await import("./useWhatsNew");

    const { isOpen, dismiss } = useWhatsNew();
    dismiss();

    expect(isOpen.value).toBe(false);
    expect(localStorage.getItem("shopmon-whats-new")).toBe(WHATS_NEW_VERSION);
  });

  it("can be reopened manually after dismissal", async () => {
    const { useWhatsNew } = await import("./useWhatsNew");

    const { isOpen, dismiss, open } = useWhatsNew();
    dismiss();
    open();

    expect(isOpen.value).toBe(true);
  });
});
