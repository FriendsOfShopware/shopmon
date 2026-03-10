export interface Sponsor {
  name: string;
  url: string;
  description?: string;
}

// Add sponsor logo files under /frontend/public/sponsors and list them here.
// The same data powers the public home page and the in-app What's New modal.
export const sponsors: Sponsor[] = [
  {
    name: "8mylez",
    url: "https://www.8mylez.com/",
    description: "8mylez is a Shopware Platinum & Technology Partner based in Paderborn, Germany.",
  },
  {
    name: "JCI - Just Code It",
    url: "mailto:mail@just-code-it.dev",
    description:
      "JCI builds and maintains modern Shopware platforms with a focus on custom development, performance, and long-term scalability.",
  },
];
