<template>
    <div class="privacy-page">
        <header-container title="Privacy Policy" />
        <main-container class="panel">
            <div class="privacy-content">

            <section>
                <h2>Data We Collect</h2>
                <h3>Account Information</h3>
                <ul>
                    <li>Email address and name for account creation</li>
                    <li>GitHub User ID if GitHub has been used as Login Provider</li>
                    <li>Password (hashed using bcrypt)</li>
                    <li>User preferences and settings</li>
                </ul>

                <h3>Shop Monitoring Data</h3>
                <ul>
                    <li>Shopware instance URLs and API Credentials (encrypted outside the Database using AES-GCM)</li>
                    <li>Extension information and version details</li>
                    <li>Scheduled task information</li>
                </ul>

                <h3>Technical Data</h3>
                <ul>
                    <li>Log files for debugging and security purposes</li>
                    <li>Session data for authentication</li>
                    <li>Error tracking data (via Sentry referecing user id)</li>
                </ul>
            </section>

            <section>
                <h2>How We Use Your Data</h2>
                <ul>
                    <li>To provide shop monitoring services</li>
                    <li>To send notifications about your shops' status</li>
                    <li>To improve our service and troubleshoot issues</li>
                    <li>To ensure security and prevent abuse</li>
                </ul>
            </section>

            <section>
                <h2>Data Storage and Security</h2>
                <ul>
                    <li>The Server is hosted at Oracle Cloud at Frankfurt in Germany</li>
                    <li>Passwords are hashed, Shopware API Credentials are encrypted using AES-GCM</li>
                    <li>Backups are stored encrypted</li>
                    <li>Staging and Production are seperated and fixtures are used for development/testing</li>
                    <li>When user deletes his shop/account, the data will be directly deleted</li>
                </ul>
            </section>

            <section>
                <h2>Cookies and Tracking</h2>
                <p>
                    We do not include any external JavaScript library to the website.
                </p>
            </section>

            <section>
                <h2>Contact</h2>
                <p>
                    For any privacy-related questions or concerns, please reach out to us 
                    through our official GitHub repository or per Mail to shopmon at fos.gg.
                </p>
            </section>
            </div>
        </main-container>
    </div>
</template>

<script setup lang="ts">
</script>
<style scoped>
.privacy-content {
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
}
</style>