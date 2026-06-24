import AdminLayout from "@/layouts/AdminLayout.vue";
import type { RouteRecordRaw } from "vue-router";

const adminRoutes: RouteRecordRaw[] = [
  {
    path: "/admin",
    component: AdminLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: "",
        redirect: "/admin/dashboard",
      },
      {
        path: "dashboard",
        name: "admin.dashboard",
        component: () => import("@/views/admin/Dashboard.vue"),
      },
      {
        path: "users",
        name: "admin.users",
        component: () => import("@/views/admin/ListUsers.vue"),
      },
      {
        path: "users/:id",
        name: "admin.users.detail",
        component: () => import("@/views/admin/DetailUser.vue"),
      },
      {
        path: "organizations",
        name: "admin.organizations",
        component: () => import("@/views/admin/ListOrganizations.vue"),
      },
      {
        path: "organizations/:id",
        name: "admin.organizations.detail",
        component: () => import("@/views/admin/DetailOrganization.vue"),
      },
      {
        path: "environments",
        name: "admin.environments",
        component: () => import("@/views/admin/ListEnvironments.vue"),
      },
      {
        path: "environments/:id",
        name: "admin.environments.detail",
        component: () => import("@/views/admin/DetailEnvironment.vue"),
      },
      {
        path: "audit-log",
        name: "admin.auditLog",
        component: () => import("@/views/admin/ListAuditLog.vue"),
      },
    ],
  },
];

export default adminRoutes;
