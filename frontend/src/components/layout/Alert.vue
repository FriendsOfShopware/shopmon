<template>
    <div :class="['alert', `alert-${type}`]">
        <div class="alert-icon">
            <icon-fa6-solid:circle-info aria-hidden="true" class="icon-info" v-if="type === 'info'" />
            <icon-fa6-solid:circle-check aria-hidden="true" class="icon-success" v-if="type === 'success'" />
            <icon-fa6-solid:circle-info aria-hidden="true" class="icon-warning" v-if="type === 'warning'" />
            <icon-fa6-solid:circle-xmark aria-hidden="true" class="icon-error" v-if="type === 'error'" />
        </div>
        
        <!-- eslint-disable vue/no-v-html -->
        <div class="alert-content" v-html="content" />
        <!-- eslint-enable vue/no-v-html -->
    </div>
</template>

<script lang="ts" setup>
defineProps<{ type: string; content: string }>();

</script>

<style scoped>
.alert {
    display: flex;
    align-items: flex-start;
    padding: 1rem;
    border-radius: 0.375rem;
    border: 1px solid transparent;
    margin-top: 0.75rem;
    
    &-icon {
        flex-shrink: 0;
        
        svg {
            width: 1.25rem;
            height: 1.25rem;
        }
    }
    
    &-content {
        flex: 1;
        margin-left: 0.75rem;
        
        @media (min-width: 768px) {
            display: flex;
            justify-content: space-between;
        }
    }

    &-info {
        border-color: var(--info-color);
        background-color: var(--info-background);
    }

    &-success {
        border-color: var(--success-color);
        background-color: var(--success-background);
    }

    &-warning {
        border-color: var(--warning-color);
        background-color: var(--warning-background);
    }

    &-error {
        border-color: var(--error-color);
        background-color: var(--error-background);
    }
}
</style>
