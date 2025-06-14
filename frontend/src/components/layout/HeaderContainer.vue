<template>
    <header>
        <h1 class="header-title">{{ title }}</h1>
        <div class="header-actions">
            <slot />
        </div>
    </header>
</template>

<script lang="ts" setup>
defineProps<{ title: string }>();
</script>

<style>
header {
    margin: 0 auto 1rem;
    display: flex;
    justify-content: space-between;
    width: 100%;
    grid-area: header;
}

.header-title {
    font-size: 1.875rem;
    line-height: 2.25rem;
    font-weight: 700;
    color: #fff;
}

.header-actions {
    display: flex;
    gap: .5rem;

    .btn {
        background: #0284c7;
        border-color: transparent;
        color: #fff;
        transition: background .4s;

        &:hover {
            background: #0284c777;
        }
    }
}
</style>
