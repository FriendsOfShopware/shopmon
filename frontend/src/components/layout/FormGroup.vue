<template>
    <div class="form-group">
        <h3 class="panel-title">{{ title }}</h3>
        <p v-if="subTitle" class="form-group-subtitle">{{ subTitle }}</p>

        <Alert v-if="$slots.info" type="info">
            <slot name="info" />
        </Alert>
        
        <div class="content">
            <slot />
        </div>
    </div>
</template>

<script lang="ts" setup>
defineProps<{ title: string; subTitle?: string }>();
</script>

<style scoped>
.form-group-title {
    margin-bottom: 0.75rem;
}

.form-group + .form-group {
    margin-top: 2rem;
}

.content {
    margin: 1.5rem 0 0;
    background-color: var(--panel-background);

    .form-group-table & {
        padding: 0;
    }

    > :not(:first-child) {
        margin-top: 1rem;;
    }

    .dark & {
        box-shadow: none;
    }
    
    label {
        display: block;
        margin-bottom: .25rem;
        font-weight: 500;
        cursor: pointer;
    }
}
</style>
