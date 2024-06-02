<template>
    <header>
        <div class="container header">
            <h1 class="header-title">{{ title }}</h1>
            <div class="header-actions">
                <slot />
            </div>

        </div>
    </header>
</template>

<script lang="ts" setup>
defineProps<{ title: string }>();
</script>

<style scoped>
header {
    background-color: var(--panel-background);
    box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
}

.dark header {
    background-color: #26262680;
    border-bottom: 1px solid #4a4a4a;
}

.header {
    margin: 0 auto;
    padding-top: 1.5rem;
    padding-bottom: 1.5rem;
    display: flex;
    justify-content: space-between;
}

.header-title {
    font-size: 1.875rem;
    line-height: 2.25rem;
    font-weight: 700;
}

.header-actions {
    display: flex;
    gap: .5rem;
}
</style>
