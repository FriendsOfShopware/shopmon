<template>
    <svg class="shops-empty-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
    </svg>

    <h3 class="shops-empty-title">No shops</h3>

    <p class="shops-empty-description">Get started by adding your first Shop.</p>

    <div class="shops-empty-cta">
        <router-link to="/account/shops/new" class="btn btn-primary">
            <icon-fa6-solid:plus class="icon" aria-hidden="true" />
            Add Shop
        </router-link>
    </div>
</template>

<style scoped>
.shops-empty {
    text-align: center;

    &-icon {
        width: 3rem;
        height: 3rem;
        margin: 0 auto 0.5rem;
    }

    &-title {
        font-size: 1.25rem;
        line-height: 1.75rem;
        font-weight: 500;
    }

    &-description {
        margin-top: 0.25rem;
    }

    &-cta {
        margin-top: 1.5rem;
    }
}
</style>
