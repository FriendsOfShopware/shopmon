<template>
    <footer class="app-footer">
        <div class="container">
            <div class="footer-links">
                <router-link :to="{ name: 'privacy.authenticated' }" class="footer-link">
                    Privacy
                </router-link>
                <span class="footer-separator">|</span>
                <router-link :to="{ name: 'imprint.authenticated' }" class="footer-link">
                    Legal Notice
                </router-link>
            </div>
        </div>
    </footer>
</template>

<style scoped>
.app-footer {
    margin-top: auto;
    padding: 2rem 0;
    border-top: 1px solid var(--panel-border-color);
    background-color: var(--panel-background);
}

.footer-links {
    text-align: center;
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 1rem;
}

.footer-link {
    color: var(--text-color-muted);
    text-decoration: none;
    font-size: 0.875rem;
    transition: color 0.2s ease;
}

.footer-link:hover {
    color: var(--primary-color);
}

.footer-separator {
    color: var(--text-color-muted);
    font-size: 0.875rem;
}
</style>