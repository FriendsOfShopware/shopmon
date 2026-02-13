declare module "*.mjml" {
  const content: (vars: Record<string, unknown>) => string;
  export default content;
}

declare global {
  const SENTRY_RELEASE: string;
}

export {};
