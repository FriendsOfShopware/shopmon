import type { RouteLocationNormalized, Router } from "vue-router";

import { useReturnUrl } from "@/composables/useReturnUrl";
import { useSession, fetchSession } from "@/composables/useSession";
import { api, setToken } from "@/helpers/api";
import { i18n } from "@/i18n";
import { nextTick } from "vue";

export { routes } from "./routes";

let initialSessionLoaded = false;

export function setupRouterGuards(router: Router) {
  router.beforeEach(async (to: RouteLocationNormalized) => {
    // Handle OAuth/SSO callback code — exchange for token
    if (to.query.code && !to.query.state) {
      const { data } = await api.POST("/auth/exchange-code" as any, {
        body: { code: to.query.code as string },
      });
      if (data?.token) {
        setToken(data.token);
        await fetchSession();
      }
      const { code: _removed, ...remainingQuery } = to.query;
      return { path: to.path, query: remainingQuery };
    }

    const { session, loading } = useSession();

    // On first navigation, wait for the session to load
    if (!initialSessionLoaded) {
      if (loading.value) {
        await fetchSession();
      }
      initialSessionLoaded = true;
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
      "auth.callback",
    ];
    const authRequired = !publicPages.includes(to.name as string);
    const { setReturnUrl } = useReturnUrl();

    if (authRequired && !session.value) {
      setReturnUrl(to.fullPath);
      return { name: "account.login" };
    } else if (to.name === "account.login") {
      setReturnUrl("/app/dashboard");
    }

    // Redirect to onboarding if user has no active organization
    if (
      session.value &&
      !session.value.session.activeOrganizationId &&
      to.name !== "account.onboarding" &&
      to.name !== "account.organization.accept" &&
      to.name !== "account.organization.reject"
    ) {
      return { name: "account.onboarding" };
    }

    // Check admin routes
    if (to.path.startsWith("/admin") && session.value) {
      const userRole = session.value.user.role;
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
