<template>
  <div class="container">
    <header-container title="Legal notice" />

    <Panel>
      <section>
        <h2>Contact</h2>
        <div class="contact-info">
          <p>
            As a Non-Profit Website, it's not required to provide a legal representative. If you
            have questions or concerns, please reach out to us.
          </p>
          <p>E-Mail: shopmon at fos.gg</p>
        </div>
      </section>
    </Panel>
  </div>
</template>

<script setup lang="ts"></script>
<style scoped>
section {
  margin-bottom: 2rem;
}

h2 {
  color: var(--primary-color);
  margin-bottom: 1rem;
  font-size: 1.5rem;
  font-weight: 600;
}

p {
  margin-bottom: 1rem;
  color: var(--text-color);
}

.contact-info p:last-child {
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
