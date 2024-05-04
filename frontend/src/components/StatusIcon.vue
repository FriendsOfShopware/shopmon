<template>
    <span v-if="tooltip" :data-tooltip="status" class="has-tooltip">
        <component
            class="icon icon-status"
            :is="getIconComponent(status)"
            :class="getIconClasses(status)"
        />
    </span>

    <template v-else>
        <component
            class="icon icon-status"
            :is="getIconComponent(status)"
            :class="getIconClasses(status)"
        />
    </template>
</template>

<script setup>
import FaCircleXmark from '~icons/fa6-solid/circle-xmark';
import FaCircleInfo from '~icons/fa6-solid/circle-info';
import FaCircleCheck from '~icons/fa6-solid/circle-check';
import FaCircle from '~icons/fa6-regular/circle';

const props = defineProps({
    status: {
        type: String,
        required: true,
    },
    tooltip: {
        type: Boolean,
        required: false,
        default: false
    }
});

function getIconComponent(status) {
    switch (status) {
        case 'red':
        case 'inactive':
            return FaCircleXmark;
        case 'yellow':
            return FaCircleInfo;
        case 'not installed':
            return FaCircle;
        default:
            return FaCircleCheck;
    }
}

function getIconClasses(status) {
    switch (status) {
        case 'red':
            return 'icon-error';
        case 'yellow':
            return 'icon-warning';
        case 'inactive':
        case 'not installed':
            return 'icon-muted';
        default:
            return 'icon-success';
    }
}
</script>

<style scoped>
.has-tooltip {
    &:before {
        bottom: calc(100% + 6px);
        left: -.25rem;
        text-transform: capitalize;
    }

    &:after {
        margin-bottom: -4px;
        left: .5rem;
    }
}
</style>
