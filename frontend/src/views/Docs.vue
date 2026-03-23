<template>
  <header-container title="Documentation" />
  <main-container>
    <div class="docs">
      <!-- Table of Contents -->
      <nav class="docs-nav">
        <h3 class="docs-title">
          <icon-fa6-solid:list class="icon" />
          Contents
        </h3>
        <ul class="docs-toc">
          <li><a href="#getting-started">Getting Started</a></li>
          <li><a href="#connecting-shop">Connecting a Shop</a></li>
          <li><a href="#organizations">Organizations</a></li>
          <li><a href="#projects">Projects</a></li>
          <li><a href="#dashboard-overview">Dashboard Overview</a></li>
          <li><a href="#health-checks">Health Checks</a></li>
          <li><a href="#extensions">Extensions</a></li>
          <li><a href="#scheduled-tasks">Scheduled Tasks</a></li>
          <li><a href="#queue">Queue Monitoring</a></li>
          <li><a href="#sitespeed">Performance Monitoring</a></li>
          <li><a href="#deployments">Deployments</a></li>
          <li><a href="#changelog">Changelog</a></li>
          <li><a href="#notifications">Notifications</a></li>
          <li><a href="#shop-token">Bypass Authentication</a></li>
          <li><a href="#packages-mirror">Packages Mirror</a></li>
          <li><a href="#sso">Single Sign-On (SSO)</a></li>
        </ul>
      </nav>

      <!-- Getting Started -->
      <section id="getting-started" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:rocket class="icon" />
          Getting Started
        </h2>
        <div class="docs-content">
          <p>
            Shopmon is an open-source monitoring dashboard for Shopware 6 shops. It lets you monitor
            all your Shopware environments from one central place — shop status, performance
            metrics, extension updates, and security alerts in real-time.
          </p>
          <p>To start monitoring, you need to:</p>
          <ol>
            <li>
              Create an <a href="#organizations">Organization</a> to group your team and shops
            </li>
            <li>Create a <a href="#projects">Project</a> within that organization</li>
            <li><a href="#connecting-shop">Connect a Shop</a> by providing API credentials</li>
          </ol>
        </div>
      </section>

      <!-- Connecting a Shop -->
      <section id="connecting-shop" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:plug class="icon" />
          Connecting a Shop
        </h2>
        <div class="docs-content">
          <p>There are two ways to connect your Shopware shop to Shopmon:</p>

          <h4>Option 1: Using the Shopmon Plugin (Recommended)</h4>
          <p>
            Install the
            <a href="https://github.com/FriendsOfShopware/FroshShopmon" target="_blank"
              >FroshShopmon Plugin</a
            >
            in your Shopware shop. The plugin automatically creates the required integration and
            provides a base64-encoded connection string that you can paste into Shopmon during shop
            setup.
          </p>
          <ol>
            <li>Install and activate the FroshShopmon plugin in your Shopware Administration</li>
            <li>Copy the base64 connection string from the plugin configuration</li>
            <li>
              In Shopmon, go to
              <router-link :to="{ name: 'account.shops.new' }">Add Shop</router-link> and click
              "Connect using Shopmon Plugin"
            </li>
            <li>Paste the connection string and the credentials will be filled automatically</li>
          </ol>

          <h4>Option 2: Manual Integration</h4>
          <p>
            Create an integration in your Shopware Administration under
            <strong>Settings &rarr; System &rarr; Integrations</strong>.
          </p>
          <p>The integration needs the following permissions:</p>
          <ul>
            <li>
              <strong>Read access</strong> to: system info, plugins, scheduled tasks, message queue
              stats
            </li>
            <li>
              <strong>Write access</strong> to: cache operations (for remote cache clear), scheduled
              task rescheduling
            </li>
          </ul>
          <p>
            For the exact list of required permissions, see the
            <a
              href="https://github.com/FriendsOfShopware/FroshShopmon?tab=readme-ov-file#permissions"
              target="_blank"
              >plugin documentation</a
            >.
          </p>
          <p>
            After creating the integration, copy the <strong>Client-ID</strong> and
            <strong>Client-Secret</strong> and enter them when
            <router-link :to="{ name: 'account.shops.new' }">adding a new shop</router-link>.
          </p>

          <Alert type="info">
            Your Client-Secret is encrypted before being stored. Shopmon uses it exclusively to
            authenticate with your Shopware API.
          </Alert>
        </div>
      </section>

      <!-- Organizations -->
      <section id="organizations" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:building class="icon" />
          Organizations
        </h2>
        <div class="docs-content">
          <p>
            Organizations are the top-level grouping entity. Every shop belongs to an organization,
            and organizations can have multiple members.
          </p>

          <h4>Managing Members</h4>
          <p>
            Organization owners can invite new members by email. Invitees receive an email with
            accept/reject links. Expired invitations are automatically cleaned up.
          </p>

          <h4>Creating an Organization</h4>
          <p>
            Go to
            <router-link :to="{ name: 'account.organizations.list' }">My Organisations</router-link>
            and click "New Organization". Give it a name to get started.
          </p>
        </div>
      </section>

      <!-- Projects -->
      <section id="projects" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:folder class="icon" />
          Projects
        </h2>
        <div class="docs-content">
          <p>
            Projects sit within an organization and group related shops together — for example, a
            client project with staging and production environments.
          </p>
          <ul>
            <li>Each project has a name, description, and optional Git repository URL</li>
            <li>
              The Git URL is used to link commits in the <a href="#deployments">Deployments</a> view
            </li>
            <li>
              Projects can have <strong>API Keys</strong> with specific scopes (e.g.
              <code>deployments</code>) for CI/CD integration
            </li>
            <li>
              Projects can manage <a href="#packages-mirror">Packages Mirror</a> tokens to serve
              Shopware store packages through a fast CDN
            </li>
          </ul>
        </div>
      </section>

      <!-- Dashboard -->
      <section id="dashboard-overview" class="docs-section">
        <h2 class="docs-title">
          <icon-ri:dashboard-fill class="icon" />
          Dashboard Overview
        </h2>
        <div class="docs-content">
          <p>
            The <router-link :to="{ name: 'account.dashboard' }">Dashboard</router-link> gives you a
            quick overview of all your shops and organizations. Each shop card shows:
          </p>
          <ul>
            <li><strong>Shop name</strong> and favicon</li>
            <li><strong>Project</strong> it belongs to</li>
            <li><strong>Shopware version</strong></li>
            <li>
              <strong>Status indicator</strong> (green, yellow, or red) based on health checks
            </li>
          </ul>
          <p>
            The dashboard also shows recent changes across all your shops, so you can see what has
            been updated at a glance.
          </p>
        </div>
      </section>

      <!-- Health Checks -->
      <section id="health-checks" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:circle-check class="icon" />
          Health Checks
        </h2>
        <div class="docs-content">
          <p>
            After each scrape, Shopmon runs automated health checks and assigns a status (green,
            yellow, or red) to your shop. The overall shop status reflects the worst check result.
          </p>

          <h4>Built-in Checks</h4>
          <dl class="docs-check-list">
            <div class="docs-check-item">
              <dt>Security</dt>
              <dd>
                Cross-references your Shopware version against known CVE advisories. Warns if the
                SwagPlatformSecurity plugin is missing or outdated.
              </dd>
            </div>

            <div class="docs-check-item">
              <dt>Environment</dt>
              <dd>
                Verifies that your Shopware environment is set to <code>production</code> or
                <code>staging</code> (not <code>dev</code>).
              </dd>
            </div>

            <div class="docs-check-item">
              <dt>Scheduled Tasks</dt>
              <dd>
                Flags any scheduled task that is overdue by more than its configured interval.
              </dd>
            </div>

            <div class="docs-check-item">
              <dt>Admin Worker</dt>
              <dd>
                Warns if the Admin Worker is still enabled — it should be disabled in favour of a
                proper message queue worker in production.
              </dd>
            </div>

            <div class="docs-check-item">
              <dt>Frosh Tools Integration</dt>
              <dd>
                If
                <a href="https://github.com/FriendsOfShopware/FroshTools" target="_blank"
                  >FroshTools</a
                >
                is installed, additional checks are displayed including PHP version, MySQL version,
                opcache status, memory limits, queue storage type, and mail configuration.
              </dd>
            </div>
          </dl>

          <Alert type="info">
            You can suppress individual checks per shop using the ignore feature if they are known
            false positives for your environment.
          </Alert>
        </div>
      </section>

      <!-- Extensions -->
      <section id="extensions" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:plug class="icon" />
          Extensions
        </h2>
        <div class="docs-content">
          <p>Shopmon tracks all installed plugins and apps in your Shopware shop:</p>
          <ul>
            <li>View installed versions and available updates</li>
            <li>See Shopware Store ratings and links</li>
            <li>Get notified when updates are available</li>
            <li>View extension changelogs for available updates</li>
          </ul>

          <h4>Cross-Shop Extension View</h4>
          <p>
            The
            <router-link :to="{ name: 'account.extension.list' }">My Extensions</router-link> page
            aggregates all extensions across all your shops. This helps you quickly identify which
            shops have outdated extensions or inconsistent versions.
          </p>

          <h4>Compatibility Check</h4>
          <p>
            When a new Shopware version is available, use the built-in
            <strong>Compatibility Check</strong>
            to verify whether your installed extensions are compatible with the target version
            before upgrading.
          </p>
        </div>
      </section>

      <!-- Scheduled Tasks -->
      <section id="scheduled-tasks" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:list-check class="icon" />
          Scheduled Tasks
        </h2>
        <div class="docs-content">
          <p>
            Monitor all registered scheduled tasks in your Shopware shop. The overview shows each
            task's name, status, interval, and next/last execution time.
          </p>
          <ul>
            <li>Overdue tasks are highlighted with a warning</li>
            <li>
              You can <strong>reschedule stuck tasks</strong> directly from Shopmon — this resets
              the task status back to "scheduled" so it will be picked up again
            </li>
          </ul>
        </div>
      </section>

      <!-- Queue -->
      <section id="queue" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:bars-staggered class="icon" />
          Queue Monitoring
        </h2>
        <div class="docs-content">
          <p>
            View the message queue sizes of your Shopware shop. This helps you identify if messages
            are piling up and workers aren't keeping up.
          </p>
        </div>
      </section>

      <!-- Sitespeed -->
      <section id="sitespeed" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:rocket class="icon" />
          Performance Monitoring (Sitespeed)
        </h2>
        <div class="docs-content">
          <p>
            Shopmon can perform daily automated page speed checks using
            <a href="https://www.sitespeed.io" target="_blank">sitespeed.io</a> and display the
            results over time.
          </p>

          <h4>What's Measured</h4>
          <ul>
            <li><strong>TTFB</strong> — Time to First Byte</li>
            <li><strong>Fully Loaded</strong> — Total page load time</li>
            <li><strong>LCP</strong> — Largest Contentful Paint</li>
            <li><strong>FCP</strong> — First Contentful Paint</li>
            <li><strong>CLS</strong> — Cumulative Layout Shift</li>
            <li><strong>Transfer Size</strong> — Total bytes transferred</li>
          </ul>

          <h4>Setup</h4>
          <p>
            Configure up to 5 URLs per shop for monitoring. Measurements are taken daily and
            displayed in a time-series chart. When linked with
            <a href="#deployments">Deployments</a>, performance changes can be correlated with
            specific releases.
          </p>
        </div>
      </section>

      <!-- Deployments -->
      <section id="deployments" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:code-branch class="icon" />
          Deployments
        </h2>
        <div class="docs-content">
          <p>
            Track deployment history for your shops. Deployments can be created automatically from
            your CI/CD pipeline using a project API key.
          </p>

          <h4>CI/CD Integration</h4>
          <p>
            The easiest way to integrate deployments is using
            <a href="https://github.com/FriendsOfShopware/shopmon-cli" target="_blank"
              >shopmon-cli</a
            >, which handles authentication and reporting automatically. Add it to your deployment
            pipeline to track every release.
          </p>
          <ol>
            <li>
              Create an API key for your project with the <code>deployments</code> scope (Project
              Settings &rarr; API Keys)
            </li>
            <li>
              Install <code>shopmon-cli</code> in your CI/CD pipeline and configure it with your API
              key
            </li>
            <li>
              The CLI will report deployment details (command, return code, execution time, git
              reference, composer packages) and upload the deployment log output
            </li>
          </ol>

          <h4>What's Tracked</h4>
          <ul>
            <li>Deployment name (auto-generated if not provided)</li>
            <li>Command executed, return code, and execution time</li>
            <li>Git reference (linked to your project's Git URL)</li>
            <li>Composer package versions at time of deployment</li>
            <li>Full deployment log output</li>
          </ul>

          <Alert type="info">
            Sitespeed measurements are automatically tagged with the latest deployment, allowing you
            to correlate performance changes with specific releases.
          </Alert>
        </div>
      </section>

      <!-- Changelog -->
      <section id="changelog" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:file-waveform class="icon" />
          Changelog
        </h2>
        <div class="docs-content">
          <p>
            Shopmon automatically detects and records changes to your shop environment on each
            scrape:
          </p>
          <ul>
            <li>Shopware version upgrades or downgrades</li>
            <li>Extension installations, updates, removals</li>
            <li>Extension activations and deactivations</li>
          </ul>
          <p>
            The changelog provides a complete audit trail of what changed in your environment and
            when.
          </p>
        </div>
      </section>

      <!-- Notifications -->
      <section id="notifications" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:bell class="icon" />
          Notifications
        </h2>
        <div class="docs-content">
          <p>Shopmon notifies you when something changes in your shop environment:</p>

          <h4>In-App Notifications</h4>
          <p>
            Notifications appear in the notification bell in the top bar. They are triggered when a
            shop's status worsens or when authentication errors occur during scraping.
          </p>

          <h4>Email Alerts</h4>
          <p>
            You can subscribe to email alerts per shop. When subscribed, you'll receive an email
            whenever the shop's status changes. Duplicate alerts are prevented — you won't be
            spammed for the same issue.
          </p>
        </div>
      </section>

      <!-- Shop Token -->
      <section id="shop-token" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:key class="icon" />
          Bypass Authentication
        </h2>
        <div class="docs-content">
          <p>
            If your Shopware shop is protected by HTTP authentication (e.g. Basic Auth on a staging
            environment), Shopmon won't be able to reach it by default.
          </p>
          <p>
            Each shop in Shopmon has a unique <strong>Shop Token</strong>. Configure your web server
            or reverse proxy to allow requests that include the header
            <code>shopmon-shop-token</code> with this token value to bypass authentication.
          </p>
          <p>
            The token is displayed on the shop information page and during shop creation. You can
            copy it to your clipboard with the copy button.
          </p>
        </div>
      </section>

      <!-- Packages Mirror -->
      <section id="packages-mirror" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:cube class="icon" />
          Packages Mirror
        </h2>
        <div class="docs-content">
          <p>
            Shopmon can act as a proxy for Shopware store packages through a Global CDN, delivering
            packages with ~80ms response times compared to ~6s from
            <code>packages.shopware.com</code> — up to 75x faster.
          </p>

          <h4>How It Works</h4>
          <p>
            Add your Shopware store token to a project in the project settings. Shopmon registers it
            with the packages mirror, which syncs available packages every hour. Your Composer
            installs then pull packages from the mirror instead of the Shopware store directly.
          </p>

          <h4>Setup</h4>
          <ol>
            <li>
              Go to your project settings and add your Shopware store token in the
              <strong>Packages Tokens</strong> section
            </li>
            <li>
              Remove <code>packages.shopware.com</code> from the <code>repositories</code> in your
              <code>composer.json</code>
            </li>
            <li>
              Add the mirror URL as a Composer repository (the exact snippet is shown in the project
              settings after adding a token)
            </li>
            <li>
              Authenticate with your Shopware store token using
              <code>composer config --auth</code>
            </li>
          </ol>

          <Alert type="info">
            Tokens are validated against <code>packages.shopware.com</code> before being saved.
            Syncing happens automatically every hour, but you can also trigger a manual sync from
            the project settings.
          </Alert>
        </div>
      </section>

      <!-- SSO -->
      <section id="sso" class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:right-to-bracket class="icon" />
          Single Sign-On (SSO)
        </h2>
        <div class="docs-content">
          <p>
            Organizations can configure OpenID Connect (OIDC) based Single Sign-On for their
            members.
          </p>
          <ul>
            <li>
              Configure your OIDC provider (issuer, client ID/secret, endpoints) in the organization
              settings
            </li>
            <li>
              Auto-discovery via <code>/.well-known/openid-configuration</code> is supported — just
              provide the issuer URL
            </li>
            <li>
              Users with matching email domains will be automatically redirected to the configured
              SSO provider on login
            </li>
          </ul>
          <p>
            SSO configuration requires the <strong>update</strong> permission on the organization.
          </p>
        </div>
      </section>

      <!-- Scraping Info -->
      <section class="docs-section">
        <h2 class="docs-title">
          <icon-fa6-solid:rotate class="icon" />
          How Scraping Works
        </h2>
        <div class="docs-content">
          <p>
            Shopmon automatically scrapes your shop once per hour. During each scrape, it collects:
          </p>
          <ul>
            <li>Shopware version</li>
            <li>All installed plugins and apps with their versions and status</li>
            <li>Available updates from the Shopware Store</li>
            <li>Scheduled task statuses</li>
            <li>Message queue sizes</li>
            <li>Cache configuration (environment, HTTP cache, cache adapter)</li>
            <li>Shop favicon</li>
          </ul>
          <p>
            After collecting data, all <a href="#health-checks">health checks</a> are executed and
            the shop status is updated. If the status changes,
            <a href="#notifications">notifications</a>
            are triggered.
          </p>
          <p>
            You can also trigger a manual refresh from any shop's detail page using the refresh
            button in the header.
          </p>

          <h4>Remote Actions</h4>
          <p>From the shop detail page, you can also:</p>
          <ul>
            <li>
              <strong>Clear Cache</strong> — Remotely clears your Shopware HTTP and application
              cache
            </li>
            <li>
              <strong>Reschedule Tasks</strong> — Reset stuck scheduled tasks back to "scheduled"
              status
            </li>
          </ul>
        </div>
      </section>
    </div>
  </main-container>
</template>

<script setup lang="ts"></script>

<style scoped>
.docs {
  max-width: 800px;
}

.docs-nav,
.docs-section {
  background-color: var(--panel-background);
  padding: 1.25rem;
  border-radius: 0.375rem;
  box-shadow:
    0 1px 3px 0 rgba(0, 0, 0, 0.1),
    0 1px 2px 0 rgba(0, 0, 0, 0.06);
  margin-bottom: 2rem;
}

.dark .docs-nav,
.dark .docs-section {
  box-shadow: none;
}

.docs-title {
  font-size: 1.125rem;
  line-height: 1.75rem;
  font-weight: 500;
  padding-bottom: 0.25rem;
  margin-bottom: 1rem;
  border-bottom: 1px solid var(--panel-border-color);

  .icon {
    margin-right: 0.25rem;
  }
}

.docs-toc {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;

  li a {
    display: block;
    padding: 0.375rem 0.75rem;
    border-radius: 0.25rem;
    font-size: 0.9375rem;
    transition: background-color 0.15s;

    &:hover {
      background-color: var(--item-hover-background);
    }
  }
}

.docs-section {
  scroll-margin-top: 1rem;
}

.docs-content {
  line-height: 1.75;

  p {
    margin-bottom: 1rem;
  }

  p:last-child {
    margin-bottom: 0;
  }

  h4 {
    font-size: 1.1rem;
    font-weight: 600;
    margin-top: 1.5rem;
    margin-bottom: 0.75rem;

    &:first-child {
      margin-top: 0;
    }
  }

  ol,
  ul {
    margin-bottom: 1rem;
    padding-left: 1.5rem;
  }

  ol {
    list-style: decimal;
  }

  ul {
    list-style: disc;
  }

  li {
    margin-bottom: 0.375rem;
  }

  code {
    background-color: var(--item-background);
    padding: 0.125rem 0.375rem;
    border-radius: 0.25rem;
    font-size: 0.875rem;
  }

  a {
    text-decoration: underline;
    text-decoration-color: var(--link-color);
    text-underline-offset: 2px;
  }
}

.docs-check-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 1rem;
}

.docs-check-item {
  padding: 0.75rem;
  background-color: var(--item-background);
  border-radius: 0.375rem;

  dt {
    font-weight: 600;
    margin-bottom: 0.25rem;
  }

  dd {
    color: var(--text-color-muted);
  }
}
</style>
