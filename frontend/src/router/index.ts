import type { RouteLocationNormalized, Router } from "vue-router";

import { useReturnUrl } from "@/composables/useReturnUrl";
import { authClient } from "@/helpers/auth-client";
import { i18n } from "@/i18n";
import { nextTick } from "vue";

export { routes } from "./routes";

export function setupRouterGuards(router: Router) {
  const session = authClient.useSession();

  router.beforeEach(async (to: RouteLocationNormalized) => {
    if (session.value.isPending) {
      await new Promise((resolve) => {
        const a = setInterval(() => {
          if (!session.value.isPending) {
            clearInterval(a);
            resolve(true);
          }
        }, 50);
      });
    }

    // redirect to the login page if not logged in and trying to access a restricted page
    const publicPages = [
      "home",
      "privacy",
      "imprint",
      "account.login",
      "account.register",
      "account.confirm",
      "account.forgot.password",
      "account.forgot.password.confirm",
    ];
    const authRequired = !publicPages.includes(to.name as string);
    const { setReturnUrl } = useReturnUrl();

    if (authRequired && !session.value.data) {
      setReturnUrl(to.fullPath);
      return { name: "account.login" };
    } else if (to.name === "account.login") {
      setReturnUrl("/app/dashboard");
    }

    // Check admin routes
    if (to.path.startsWith("/admin") && session.value.data) {
      const userRole = session.value.data.user.role;
      if (userRole !== "admin") {
        return { name: "home" };
      }
    }
  });

  const DEFAULT_TITLE = "Shopmon";
  router.afterEach(async (to) => {
    if (typeof document === "undefined") return;

    await nextTick();

    const titleKey = to.meta.titleKey;
    if (typeof titleKey === "string") {
      document.title = i18n.global.t(titleKey) + `| ${DEFAULT_TITLE}`;
      return;
    }

    document.title = DEFAULT_TITLE;
  });
}
