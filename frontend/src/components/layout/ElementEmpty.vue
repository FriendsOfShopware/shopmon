<template>
    <div class="element-empty">
        <svg class="element-empty-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
        </svg>

        <h3 v-if="title" class="element-empty-title">{{ title }}</h3>

        <p class="element-empty-description">
            <slot>
                Get started by adding your first element.
            </slot>
        </p>

        <div class="element-empty-cta">
            <router-link :to="route" class="btn btn-primary">
                <icon-fa6-solid:plus class="icon" aria-hidden="true" />
                {{ button }}
            </router-link>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router';

withDefaults(
    defineProps<{
        title?: string;
        button?: string;
        route: RouteLocationRaw;
    }>(),
    {
        title: 'No elements',
        button: 'Add element',
    },
);
</script>

<style scoped>
.element-empty {
    text-align: center;

    &-icon {
        width: 3rem;
        height: 3rem;
        margin: 0 auto 0.5rem;
    }

    &-title {
        font-size: 1.25rem;
        line-height: 1.75rem;
        font-weight: 500;
    }

    &-description {
        margin-top: 0.25rem;
    }

    &-cta {
        margin-top: 1.5rem;
    }
}
</style>
