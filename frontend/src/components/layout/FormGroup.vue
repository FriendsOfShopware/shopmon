<template>
    <div class="form-group">
        <div class="sidebar">
            <h3 class="sidebar-title">{{ title }}</h3>

            <p v-if="subTitle" class="sidebar-subtitle">{{ subTitle }}</p>

            <Alert type="info" v-if="$slots.info">
                <slot name="info"></slot>
            </Alert>
        </div>
        
        <div class="content">
            <slot />
        </div>
    </div>
</template>

<script lang="ts" setup>
defineProps<{ title: string; subTitle?: string }>();
</script>

<style>
.form-group {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1rem;
    padding-bottom: 3rem;
    
    @media (min-width: 768px) {
        grid-template-columns: 1fr 2fr;
    }
}

.sidebar {
    order: 2;
    
    @media (min-width: 768px) {
        order: 1;
    }

    &-title {
        font-size: 1.125rem;
        font-weight: 500;
    }
    
    &-subtitle {
        color: #6b7280;
    }
}

.content {
    order: 1;
    background-color: var(--panel-background);
    padding: 1.25rem;
    border-radius: 0.375rem;
    box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
    overflow: hidden;

    .form-group-table & {
        padding: 0;
    }
    
    @media (min-width: 768px) {
        order: 2;
    }
    
    @media (min-width: 640px) {
        padding: 1.5rem;
    }

    > :not(:first-child) {
        margin-top: 1.5rem;;
    }

    .dark & {
        box-shadow: none;
    }
    
    label {
        display: block;
        margin-bottom: .25rem;
        font-weight: 500;
        cursor: pointer;
    }
}
</style>
