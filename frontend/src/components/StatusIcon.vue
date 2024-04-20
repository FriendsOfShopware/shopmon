<template>
    <component
        :is="getIconComponent(status)"
        :class="getIconClasses(status)"
    />
</template>

<script setup>
import FaCircleXmark from '~icons/fa6-solid/circle-xmark';
import FaCircleInfo from '~icons/fa6-solid/circle-info';
import FaCircleCheck from '~icons/fa6-solid/circle-check';

const props = defineProps({
    status: {
        type: String,
        required: true,
    },
});

function getIconComponent(status) {
    switch (status) {
        case 'red':
            return FaCircleXmark;
        case 'yellow':
            return FaCircleInfo;
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
        default:
            return 'icon-success';
    }
}
</script>
