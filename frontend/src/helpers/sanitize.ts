import DOMPurify from "dompurify";

// Sanitizes untrusted HTML before it is rendered via v-html. All extension
// metadata (descriptions, installation manuals, changelogs) originates from the
// Shopware store / third-party producers and must be treated as attacker-
// controllable. Strips scripts, event handlers, and other XSS vectors while
// keeping the formatting markup the store provides.
export function sanitizeHtml(html?: string | null): string {
  if (!html) return "";
  return DOMPurify.sanitize(html, { USE_PROFILES: { html: true } });
}
