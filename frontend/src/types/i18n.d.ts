import type { DefineLocaleMessage } from "vue-i18n";
import type en from "../locales/en.json";
import type { Component } from "vue";

type MessageSchema = typeof en;

declare module "vue-i18n" {
  interface DefineLocaleMessage extends MessageSchema {}
}

declare module "vue" {
  interface ComponentCustomProperties {
    $t(key: string): string;
    $t(key: string, named: Record<string, unknown>): string;
    $t(key: string, list: unknown[]): string;
    $t(key: string, ...args: unknown[]): string;
  }
}

declare module "vue-router" {
  interface RouteMeta {
    titleKey?: string;
    icon?: Component;
  }
}
