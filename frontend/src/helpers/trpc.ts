import { createTRPCClient, httpBatchLink } from "@trpc/client";
import type { inferRouterInputs, inferRouterOutputs } from "@trpc/server";
import type { AppRouter } from "../../../api/src/trpc/router";

export type RouterOutput = inferRouterOutputs<AppRouter>;
export type RouterInput = inferRouterInputs<AppRouter>;

export const trpcClient = createTRPCClient<AppRouter>({
  links: [
    httpBatchLink({
      url: "/trpc",
    }),
  ],
});
