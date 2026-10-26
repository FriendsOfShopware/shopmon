import type { ZodType } from "zod";
import type { TypedSchema } from "vee-validate";

// @vee-validate/zod@4 only accepts Zod 3. Zod 4 validates through the public
// safeParse API, which is enough for these flat object forms.
type PartialDeep<T> = T extends (...args: never[]) => unknown
  ? T
  : T extends readonly unknown[]
    ? T
    : T extends object
      ? { [K in keyof T]?: PartialDeep<T[K]> }
      : T;

interface FormIssue {
  message: string;
  path: PropertyKey[];
  errors?: FormIssue[][];
  unionErrors?: { issues: FormIssue[] }[];
}

function normalizeFormPath(path: PropertyKey[]): string {
  const segments = path.map(String);
  if (segments.length === 0) return "";

  let fullPath = segments[0] ?? "";
  for (let i = 1; i < segments.length; i++) {
    const segment = segments[i] ?? "";
    if (Number(segment) >= 0) {
      fullPath += `[${segment}]`;
      continue;
    }
    fullPath += `.${segment}`;
  }
  return fullPath;
}

function collectIssues(issues: readonly FormIssue[], errors: Map<string, string[]>) {
  for (const issue of issues) {
    const nested = issue.errors?.flat() ?? issue.unionErrors?.flatMap((entry) => entry.issues);
    if (nested && nested.length > 0) {
      collectIssues(nested, errors);
      if (issue.path.length === 0) continue;
    }

    const path = normalizeFormPath(issue.path);
    const messages = errors.get(path) ?? [];
    messages.push(issue.message);
    errors.set(path, messages);
  }
}

function schemaAt(schema: ZodType, path?: string): ZodType | undefined {
  if (!path) return schema;
  const shape = (schema as { shape?: Record<string, ZodType> }).shape;
  return shape?.[path];
}

export function toTypedSchema<TOutput, TInput = TOutput>(
  schema: ZodType<TOutput, TInput>,
): TypedSchema<PartialDeep<TInput>, TOutput> {
  return {
    __type: "VVTypedSchema",
    async parse(values) {
      const result = await schema.safeParseAsync(values);
      if (result.success) {
        return { value: result.data, errors: [] };
      }

      const errors = new Map<string, string[]>();
      collectIssues(result.error.issues as FormIssue[], errors);
      return {
        errors: [...errors.entries()].map(([path, messages]) => ({
          path: path || undefined,
          errors: messages,
        })),
      };
    },
    cast(values) {
      const result = schema.safeParse(values);
      return (result.success ? result.data : values) as PartialDeep<TInput>;
    },
    describe(path) {
      const field = schemaAt(schema, path);
      if (!field) {
        return { required: false, exists: false };
      }
      return { required: !field.safeParse(undefined).success, exists: true };
    },
  };
}
