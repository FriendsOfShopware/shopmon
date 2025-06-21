<template>
    <header>
        <div :class="[
            'header-title',
            {'mobile-hide': titleMobileHide}
            ]"
        >
            <div v-if="breadcrumb" class="header-breadcrumb">{{ breadcrumb }}</div>
            <h1>{{ title }}</h1>
        </div>

        <div class="header-actions">
            <slot />
        </div>
    </header>
</template>

<script lang="ts" setup>
defineProps<{ title: string, breadcrumb?: string, titleMobileHide?: boolean }>();
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
    h1 {
        font-size: 1.875rem;
        line-height: 2.25rem;
        font-weight: 700;
        color: #fff;
    }

    &.mobile-hide {
        display: none;

        + .header-actions {
            margin-left: auto;
        }

        @media all and (min-width: 1024px) {
            display: block;
        }
    }
}

.header-breadcrumb {
    font-size: 1rem;
    color: #fff;
    margin-bottom: 0.25rem;
}

.header-actions {
    display: flex;
    gap: .5rem;
    align-items: flex-start;

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
