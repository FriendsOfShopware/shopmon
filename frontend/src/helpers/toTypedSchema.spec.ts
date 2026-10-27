import { z } from "zod";
import { toTypedSchema } from "./toTypedSchema";

const schema = toTypedSchema(
  z.object({
    name: z.string().min(1, "name required"),
    gitUrl: z.string().url("bad url").optional().or(z.literal("")),
    taskGraceMinutes: z.coerce.number({ error: "bad number" }).int("bad int").min(0, "bad min"),
  }),
);

it("returns parsed output for a valid form", async () => {
  const result = await schema.parse({
    name: "Shop",
    gitUrl: "",
    taskGraceMinutes: "4",
  });

  expect(result.errors).toEqual([]);
  expect(result.value).toEqual({ name: "Shop", gitUrl: "", taskGraceMinutes: 4 });
});

it("maps field errors and keeps union messages on the field", async () => {
  const result = await schema.parse({
    name: "",
    gitUrl: "not a url",
    taskGraceMinutes: "abc",
  });

  expect(result.value).toBeUndefined();
  expect(result.errors).toEqual([
    { path: "name", errors: ["name required"] },
    { path: "gitUrl", errors: ["bad url"] },
    { path: "taskGraceMinutes", errors: ["bad number"] },
  ]);
});

it("reports which fields are required", () => {
  expect(schema.describe?.("name")).toEqual({ required: true, exists: true });
  expect(schema.describe?.("gitUrl")).toEqual({ required: false, exists: true });
});
