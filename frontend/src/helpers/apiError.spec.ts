import { expect, it } from "vitest";
import { apiErrorMessage, requireMutation } from "./apiError";
it.each([{}, null, false, "", { message: "" }])(
  "provides fallback for unhelpful error %j",
  (error) => {
    expect(apiErrorMessage(error, "Try again")).toBe("Try again");
    expect(() => requireMutation({ error }, "Try again")).toThrow("Try again");
  },
);
it("accepts success and preserves useful error messages", () => {
  expect(() => requireMutation({}, "Try again")).not.toThrow();
  expect(() => requireMutation({ error: "Denied" }, "Try again")).toThrow("Denied");
  expect(() => requireMutation({ error: { message: "Denied" } }, "Try again")).toThrow("Denied");
});
