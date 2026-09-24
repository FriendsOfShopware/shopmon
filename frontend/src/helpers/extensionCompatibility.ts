export interface ExtensionCompatibilityStatus {
  label: string;
  type: string;
}

export interface CompatibilityExtension {
  active: boolean;
  compatibility?: ExtensionCompatibilityStatus | null;
}

export type CompatibilityState =
  | "incompatible"
  | "unknown"
  | "updateAvailable"
  | "compatible"
  | "inactive";

export function getCompatibilityState(extension: CompatibilityExtension): CompatibilityState {
  if (!extension.active) return "inactive";
  if (!extension.compatibility) return "unknown";
  if (extension.compatibility.type === "red") return "incompatible";
  if (extension.compatibility.label === "Available now") return "updateAvailable";
  return "compatible";
}

const severityOrder: Record<CompatibilityState, number> = {
  incompatible: 0,
  unknown: 1,
  updateAvailable: 2,
  compatible: 3,
  inactive: 4,
};

export function compareByCompatibilitySeverity(
  a: CompatibilityExtension,
  b: CompatibilityExtension,
): number {
  return severityOrder[getCompatibilityState(a)] - severityOrder[getCompatibilityState(b)];
}
