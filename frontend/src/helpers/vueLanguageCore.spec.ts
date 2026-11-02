import {
  createVueLanguagePlugin,
  forEachEmbeddedCode,
  getDefaultCompilerOptions,
} from "@vue/language-core";
import ts from "typescript";
import { describe, expect, it } from "vitest";

describe("Vue language-core patch", () => {
  it.each([
    '<Input placeholder="https://example.com" />',
    '<Input :placeholder="`https://example.com\nsecond line`" />',
    '<Input :model-value="value // inline comment\n" />',
  ])("generates valid TypeScript for %s", (template) => {
    const plugin = createVueLanguagePlugin(
      ts,
      { strict: true },
      getDefaultCompilerOptions(),
      (fileName: string) => fileName,
    );
    const source = `<script setup lang="ts">const value = 'test';\n</script><template>${template}</template>`;
    const root = plugin.createVirtualCode!(
      `/compiler-regression-${template.length}.vue`,
      "vue",
      ts.ScriptSnapshot.fromString(source),
      { getAssociatedScript: () => undefined },
    );
    expect(root).toBeDefined();
    const script = [...forEachEmbeddedCode(root!)].find((code) => code.id === "script_ts");
    expect(script).toBeDefined();
    const generated = script!.snapshot.getText(0, script!.snapshot.getLength());
    const parsed = ts.createSourceFile("component.ts", generated, ts.ScriptTarget.Latest, true);
    const program = ts.createProgram(
      ["component.ts"],
      { noLib: true },
      {
        ...ts.createCompilerHost({}),
        getSourceFile: (fileName) => (fileName === "component.ts" ? parsed : undefined),
      },
    );
    expect(program.getSyntacticDiagnostics().map((diagnostic) => diagnostic.messageText)).toEqual(
      [],
    );
    const constructorProps = generated.slice(
      generated.indexOf("new __VLS_"),
      generated.indexOf("}));"),
    );
    expect(constructorProps).toContain(
      template.includes("second line")
        ? "https://example.com\nsecond line"
        : template.includes("inline comment")
          ? "// inline comment\n"
          : "https://example.com",
    );
  });
});
