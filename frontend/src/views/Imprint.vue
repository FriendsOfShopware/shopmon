<template>
    <div class="imprint-page">
        <header-container title="Legal notice" />

        <main-container class="panel">
            <div class="imprint-content">
                <section>
                    <h2>Contact</h2>
                    <div class="contact-info">
                        <p>As a Non-Profit Website, it's not required to provide a legal representative. If you have questions or concerns, please reach out to us.</p>
                        <p>E-Mail: shopmon at fos.gg</p>
                    </div>
                </section>
            </div>
        </main-container>
    </div>
</template>

<script setup lang="ts">
</script>
<style scoped>
.imprint-content {
    max-width: 800px;
    margin: 0 auto;
    padding: 2rem;
    line-height: 1.6;
}

section {
    margin-bottom: 2rem;
}

h2 {
    color: var(--primary-color);
    margin-bottom: 1rem;
    font-size: 1.5rem;
    font-weight: 600;
}

h3 {
    margin-bottom: 0.5rem;
    margin-top: 1rem;
    font-size: 1.2rem;
    font-weight: 500;
}

ul {
    margin-bottom: 1rem;
    padding-left: 1.5rem;
}

li {
    margin-bottom: 0.5rem;
}

p {
    margin-bottom: 1rem;
    color: var(--text-color);
}ß
ß
.contact-info p:last-child,
.tech-info p:last-child {
    margin-bottom: 0;
}

a {
    color: var(--primary-color);
    text-decoration: underline;
    transition: color 0.2s ease;
}

a:hover {
    color: var(--primary-color-dark, #2563eb);
}
</style>