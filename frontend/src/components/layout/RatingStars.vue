<script lang="ts" setup>
    defineProps<{rating: number}>()
</script>

<template>
    <span :data-tooltip="`${rating / 2} from 5`" class="tooltip-position-center">
        <template v-if="rating !== null" v-for="n in 5">
        <icon-fa6-regular:star v-if="(rating / 2) - n < -.5" />
        <icon-fa6-solid:star-half-stroke v-else-if="(rating / 2) - n === -.5" />
        <icon-fa6-solid:star v-else />
    </template>
    </span>    
</template>