<template>
  <Card class="mb-8 p-0">
    <Tabs :default-value="$props.labels[0]?.key">
      <TabsList class="grid w-full gap-1 overflow-x-auto rounded-none border-b bg-muted p-1 sm:grid-cols-2 lg:grid-cols-none lg:grid-flow-col lg:auto-cols-min">
        <TabsTrigger
          v-for="label in $props.labels"
          :key="label.key"
          :value="label.key"
          class="whitespace-nowrap"
        >
          <component :is="label.icon" v-if="label.icon" class="mr-1 size-4" />
          {{ label.title }}
          <Badge v-if="label.count !== undefined" variant="secondary" class="ml-1.5">
            {{ label.count }}
          </Badge>
        </TabsTrigger>
      </TabsList>

      <TabsContent
        v-for="label in $props.labels"
        :key="label.key"
        :value="label.key"
        class="mt-0"
      >
        <slot :name="`panel-${label.key}`" :label="label" />
      </TabsContent>
    </Tabs>
  </Card>
</template>

<script setup lang="ts" generic="T extends string">
import type { FunctionalComponent } from "vue";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

defineProps<{
  labels: Array<{
    key: string;
    title: string;
    count?: number;
    icon?: FunctionalComponent;
  }>;
}>();
</script>
