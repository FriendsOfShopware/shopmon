<template>
    <div class="main-container">
        <slot />
    </div>

</template>

<script setup lang="ts">
</script>

<style scoped>
.main-container {
    grid-area: content;

    &:has(.data-table) {
        overflow: hidden;
    }
}
</style>