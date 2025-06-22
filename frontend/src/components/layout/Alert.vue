<template>
    <div :class="['alert', `alert-${type}`]">
        <div class="alert-icon">
            <component
                :is="getIconComponent(type)"
                class="icon icon-status"
                :class="`icon-${type}`"
            />
        </div>

        <div class="alert-content">
            <slot />
        </div>
    </div>
</template>

<script lang="ts" setup>
import FaCircleCheck from '~icons/fa6-solid/circle-check';
import FaCircleInfo from '~icons/fa6-solid/circle-info';
import FaCircleXmark from '~icons/fa6-solid/circle-xmark';

defineProps<{ type: string }>();

function getIconComponent(type) {
    switch (type) {
        case 'error':
            return FaCircleXmark;
        case 'info':
        case 'warning':
            return FaCircleInfo;
        default:
            return FaCircleCheck;
    }
}
</script>

<style scoped>
.alert {
    display: flex;
    align-items: flex-start;
    padding: 1rem;
    border-radius: 0.375rem;
    border: 1px solid transparent;
    
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
        padding-top: .15rem;
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
