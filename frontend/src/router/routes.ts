import type { RouteRecordRaw } from "vue-router";

import AppLayout from "@/layouts/AppLayout.vue";
import LoginLayout from "@/layouts/LoginLayout.vue";
import DefaultLayout from "@/layouts/DefaultLayout.vue";
import EnvironmentDetailLayout from "@/layouts/EnvironmentDetailLayout.vue";
import adminRoutes from "@/router/admin";

import FaShop from "~icons/fa6-solid/shop";
import FaCircleCheck from "~icons/fa6-solid/circle-check";
import FaFileWaveform from "~icons/fa6-solid/file-waveform";
import FaListCheck from "~icons/fa6-solid/list-check";
import FaPlug from "~icons/fa6-solid/plug";
import FaRocket from "~icons/fa6-solid/rocket";
import FaCodeBranch from "~icons/fa6-solid/code-branch";
import Dashboard from "~icons/ri/dashboard-fill";
import FaFolder from "~icons/fa6-solid/folder";
import FaBuilding from "~icons/fa6-solid/building";
import FaBook from "~icons/fa6-solid/book";

export const routes: RouteRecordRaw[] = [
  {
    path: "/",
    component: DefaultLayout,
    children: [
      {
        path: "",
        name: "home",
        component: () => import("@/views/Home.vue"),
      },
      {
        path: "privacy",
        name: "privacy",
        component: () => import("@/views/Privacy.vue"),
      },
      {
        path: "imprint",
        name: "imprint",
        component: () => import("@/views/Imprint.vue"),
      },
    ],
  },
  {
    path: "/account",
    component: LoginLayout,
    children: [
      {
        name: "account.login",
        path: "login",
        component: () => import("@/views/auth/Login.vue"),
      },
      {
        name: "account.register",
        path: "register",
        component: () => import("@/views/auth/Register.vue"),
      },
      {
        name: "account.confirm",
        path: "confirm/:token",
        component: () => import("@/views/auth/AccountConfirm.vue"),
      },
      {
        name: "account.forgot.password",
        path: "forgot-password",
        component: () => import("@/views/auth/ForgotPassword.vue"),
      },
      {
        name: "account.forgot.password.confirm",
        path: "forgot-password/:token",
        component: () => import("@/views/auth/ForgotPasswordConfirm.vue"),
      },
    ],
  },
  {
    path: "/app",
    component: AppLayout,
    children: [
      {
        path: "dashboard",
        name: "account.dashboard",
        component: () => import("@/views/Dashboard.vue"),
        meta: {
          titleKey: "nav.dashboard",
          icon: Dashboard,
        },
      },
      {
        name: "account.settings",
        path: "settings",
        component: () => import("@/views/account/Settings.vue"),
        meta: {
          titleKey: "nav.settings",
          icon: FaShop,
        },
      },
      {
        name: "account.shop.list",
        path: "shops",
        component: () => import("@/views/shop/ListShops.vue"),
        meta: {
          titleKey: "nav.myShops",
          icon: FaFolder,
        },
      },
      {
        name: "account.environments.new",
        path: "environments/new",
        component: () => import("@/views/environment/AddEnvironment.vue"),
      },
      {
        name: "account.shops.new",
        path: "shops/new",
        component: () => import("@/views/shop/AddShop.vue"),
      },
      {
        name: "account.shops.edit",
        path: "shops/:shopId(\\d+)/edit",
        component: () => import("@/views/shop/EditShop.vue"),
        meta: {
          titleKey: "nav.editShop",
        },
      },
      {
        name: "account.environments.edit",
        path: "environments/:environmentId(\\d+)/edit",
        component: () => import("@/views/environment/EditEnvironment.vue"),
      },
      {
        path: "environments/:environmentId(\\d+)",
        component: EnvironmentDetailLayout,
        children: [
          {
            name: "account.environments.detail",
            path: "",
            component: () => import("@/views/environment/detail/DetailEnvironment.vue"),
            meta: {
              titleKey: "nav.environmentInformation",
              icon: FaShop,
            },
          },
          {
            name: "account.environments.detail.checks",
            path: "checks",
            component: () => import("@/views/environment/detail/DetailChecks.vue"),
            meta: {
              titleKey: "nav.checks",
              icon: FaCircleCheck,
            },
          },
          {
            name: "account.environments.detail.extensions",
            path: "extensions",
            component: () => import("@/views/environment/detail/DetailExtensions.vue"),
            meta: {
              titleKey: "nav.extensions",
              icon: FaPlug,
            },
          },
          {
            name: "account.environments.detail.tasks",
            path: "tasks",
            component: () => import("@/views/environment/detail/DetailScheduledTasks.vue"),
            meta: {
              titleKey: "nav.scheduledTasks",
              icon: FaListCheck,
            },
          },
          {
            name: "account.environments.detail.queue",
            path: "queue",
            component: () => import("@/views/environment/detail/DetailQueue.vue"),
            meta: {
              titleKey: "nav.queue",
              icon: FaCircleCheck,
            },
          },
          {
            name: "account.environments.detail.sitespeed",
            path: "sitespeed",
            component: () => import("@/views/environment/detail/DetailSitespeed.vue"),
            meta: {
              titleKey: "nav.sitespeed",
              icon: FaRocket,
            },
          },
          {
            name: "account.environments.detail.changelog",
            path: "changelog",
            component: () => import("@/views/environment/detail/DetailChangelog.vue"),
            meta: {
              titleKey: "nav.changelog",
              icon: FaFileWaveform,
            },
          },
          {
            name: "account.environments.detail.deployments",
            path: "deployments",
            component: () => import("@/views/environment/detail/DetailDeployments.vue"),
            meta: {
              titleKey: "nav.deployments",
              icon: FaCodeBranch,
            },
          },
          {
            name: "account.environments.detail.deployment",
            path: "deployments/:deploymentId(\\d+)",
            component: () => import("@/views/environment/detail/DetailDeployment.vue"),
            meta: {
              titleKey: "nav.deploymentDetails",
            },
          },
        ],
      },
      {
        name: "account.organizations.list",
        path: "organizations",
        component: () => import("@/views/organization/ListOrganizations.vue"),
        meta: {
          titleKey: "nav.myOrganization",
          icon: FaBuilding,
        },
      },
      {
        name: "account.organizations.new",
        path: "organizations/new",
        component: () => import("@/views/organization/AddOrganization.vue"),
      },
      {
        name: "account.onboarding",
        path: "onboarding",
        component: () => import("@/views/organization/OnboardingOrganization.vue"),
      },
      {
        name: "account.organizations.detail",
        path: "organization",
        component: () => import("@/views/organization/DetailOrganization.vue"),
        meta: {
          titleKey: "nav.myOrganization",
          icon: FaBuilding,
        },
      },
      {
        name: "account.organizations.edit",
        path: "organization/edit",
        component: () => import("@/views/organization/EditOrganization.vue"),
      },
      {
        name: "account.organizations.sso",
        path: "organization/sso",
        component: () => import("@/views/organization/ManageSSO.vue"),
      },
      {
        name: "account.extension.list",
        path: "extensions",
        component: () => import("@/views/account/ListExtensions.vue"),
        meta: {
          titleKey: "nav.myExtensions",
          icon: FaPlug,
        },
      },
      {
        name: "account.docs",
        path: "docs",
        component: () => import("@/views/Docs.vue"),
        meta: {
          titleKey: "nav.documentation",
          icon: FaBook,
        },
      },
      {
        name: "account.organization.accept",
        path: "organizations/accept/:token",
        component: () => import("@/views/organization/AcceptRejectInvitation.vue"),
        props: {
          action: "accept",
        },
      },
      {
        name: "account.organization.reject",
        path: "organizations/reject/:token",
        component: () => import("@/views/organization/AcceptRejectInvitation.vue"),
        props: {
          action: "reject",
        },
      },
    ],
  },
  ...adminRoutes,
  // catch all redirect to home page
  {
    path: "/:pathMatch(.*)*",
    name: "not-found",
    redirect: () => ({ name: "home" }),
  },
];
