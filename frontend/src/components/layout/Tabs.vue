<template>
    <div class="panel tab-container">
        <tab-group>
            <tab-list class="tabs-list">
                <tab v-for="label in props.labels" :key="label.key" v-slot="{ selected }" as="template">
                    <button
                        class="tab"
                        :class="{ 'tab-active': selected }"
                        type="button"
                    >
                        <component :is="label.icon" v-if="label.icon" class="icon"/>
                        {{ label.title }}
                        <span v-if="label.count !== undefined" class="pill">
                            {{ label.count }}
                        </span>
                    </button>
                </tab>
            </tab-list>

            <tab-panels class="tab-panels">
                <tab-panel v-for="label in props.labels" :key="label.key" class="tab-panel">
                    <slot :name="`panel-${label.key}`" :label="label">
                        <div class="tab-panel-default">
                            <strong>{{ label.title }} Tab Panel</strong>.
                            Use
                            <pre class="tab-panel-code">&lt;template #panel-{{ label.key }}="{ label }"&gt;...&lt;/template&gt;</pre>
                            to fill with content
                        </div>
                    </slot>
                </tab-panel>
            </tab-panels>
        </tab-group>
    </div>
</template>

<script setup lang="ts" generic="T extends string">
import {TabGroup, TabList, Tab, TabPanels, TabPanel} from '@headlessui/vue';
import type {FunctionalComponent} from 'vue';

const props = defineProps<{
    labels: Array<{
        key: string;
        title: string;
        count?: number;
        icon?: FunctionalComponent;
    }>;
}>();
</script>

<style scoped>
.tab-container {
    width: 100%;
    margin-bottom: 4rem;
    padding: 0;
    overflow: hidden;
}

.tabs-list {
    border-bottom: 1px solid var(--panel-border-color);
    padding: 0 1.5rem;
    gap: 1rem;
    margin-bottom: -1px;
    display: grid;
    overflow-x: auto;
    overflow-y: hidden;

    @media (min-width: 640px) {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    @media (min-width: 1024px) {
        grid-auto-columns: min-content;
        grid-template-columns: none;
        grid-auto-flow: column;
    }
}

.tab {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem 0.75rem;
    font-size: 1rem;
    font-weight: 500;
    position: relative;
    transition: all 0.2s ease-in-out;
    white-space: nowrap;

    &::before {
        content: '';
        position: absolute;
        bottom: -1px;
        left: 50%;
        transform: translateX(-50%);
        width: 0;
        height: 3px;
        background-color: var(--primary-color);
        transition: all 0.2s ease-in-out;
    }

    &:after {
        content: '';
        position: absolute;
        bottom: -1px;
        left: -1.5rem;
        right: -1.5rem;
        height: 1px;
        background-color: var(--panel-border-color);
    }

    &:hover,
    &-active {
        color: #0ea5e9;

        &::before {
            width: 100%;
        }
    }

    .icon {
        margin-right: 0.5rem;
    }

    .pill {
        margin-left: 0.5rem;
    }
}

.tab-panels {
    margin-top: 1px;
}

.tab-panel {
    overflow-y: auto;

    &-default {
        padding: 1.5rem;
    }

    &-code {
        background-color: var(--panel-border-color);
        display: inline-block;
        padding: 0.125rem 0.25rem;
        border-radius: 0.25rem;
        font-size: 0.75rem;
    }
}
</style>
