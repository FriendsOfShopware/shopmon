import type { Component } from "vue";
import type { RouteLocationRaw } from "vue-router";

export interface BreadcrumbItem {
  label: string;
  to?: RouteLocationRaw;
  href?: string;
  icon?: Component;
}
