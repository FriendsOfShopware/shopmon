<template>
    <span
        v-if="rating !== null"
        :data-tooltip="`${rating / 2} from 5`"
        class="tooltip-position-center rating-stars"
    >
        <template
            v-for="n in 5"
            :key="n"
        >
            <icon-fa6-regular:star v-if="(rating / 2) - n < -.5" />
            <icon-fa6-solid:star-half-stroke v-else-if="(rating / 2) - n === -.5" />
            <icon-fa6-solid:star v-else />
        </template>
    </span>
</template>

<script lang="ts" setup>
defineProps<{ rating: number | null }>();
</script>

<style scoped>
.rating-stars {
    white-space: nowrap;
}
</style>
