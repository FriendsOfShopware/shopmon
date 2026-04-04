<template>
  <header-container title="Documentation" />
  <main-container>
    <div class="max-w-[800px]">
      <!-- Table of Contents -->
      <Card class="mb-8">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:list class="mr-1 inline" />
            Contents
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div class="flex flex-col gap-1">
            <a v-for="item in tocItems" :key="item.href" :href="item.href" class="block px-3 py-1.5 rounded text-[15px] hover:bg-accent transition-colors">{{ item.label }}</a>
          </div>
        </CardContent>
      </Card>

      <!-- Getting Started -->
      <Card id="getting-started" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:rocket class="mr-1 inline" />
            Getting Started
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0 [&_ol]:mb-4 [&_ol]:pl-6 [&_ol]:list-decimal [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
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
        </CardContent>
      </Card>

      <!-- Connecting a Shop -->
      <Card id="connecting-shop" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:plug class="mr-1 inline" />
            Connecting a Shop
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0 [&_ol]:mb-4 [&_ol]:pl-6 [&_ol]:list-decimal [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
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
              <router-link :to="{ name: 'account.environments.new' }">Add Shop</router-link> and
              click "Connect using Shopmon Plugin"
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
            <router-link :to="{ name: 'account.environments.new' }">adding a new shop</router-link>.
          </p>

          <Alert>
            <AlertDescription>
              Your Client-Secret is encrypted before being stored. Shopmon uses it exclusively to
              authenticate with your Shopware API.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>

      <!-- Organizations -->
      <Card id="organizations" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:building class="mr-1 inline" />
            Organizations
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0 [&_ol]:mb-4 [&_ol]:pl-6 [&_ol]:list-decimal [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
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
        </CardContent>
      </Card>

      <!-- Projects -->
      <Card id="projects" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:folder class="mr-1 inline" />
            Projects
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0 [&_ol]:mb-4 [&_ol]:pl-6 [&_ol]:list-decimal [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
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
        </CardContent>
      </Card>

      <!-- Dashboard -->
      <Card id="dashboard-overview" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-ri:dashboard-fill class="mr-1 inline" />
            Dashboard Overview
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0 [&_ol]:mb-4 [&_ol]:pl-6 [&_ol]:list-decimal [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
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
        </CardContent>
      </Card>

      <!-- Health Checks -->
      <Card id="health-checks" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:circle-check class="mr-1 inline" />
            Health Checks
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0 [&_ol]:mb-4 [&_ol]:pl-6 [&_ol]:list-decimal [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
          <p>
            After each scrape, Shopmon runs automated health checks and assigns a status (green,
            yellow, or red) to your shop. The overall shop status reflects the worst check result.
          </p>

          <h4>Built-in Checks</h4>
          <div class="flex flex-col gap-4 mb-4">
            <div class="p-3 bg-muted rounded-md">
              <dt class="font-semibold mb-1">Security</dt>
              <dd class="text-muted-foreground">
                Cross-references your Shopware version against known CVE advisories. Warns if the
                SwagPlatformSecurity plugin is missing or outdated.
              </dd>
            </div>

            <div class="p-3 bg-muted rounded-md">
              <dt class="font-semibold mb-1">Environment</dt>
              <dd class="text-muted-foreground">
                Verifies that your Shopware environment is set to <code>production</code> or
                <code>staging</code> (not <code>dev</code>).
              </dd>
            </div>

            <div class="p-3 bg-muted rounded-md">
              <dt class="font-semibold mb-1">Scheduled Tasks</dt>
              <dd class="text-muted-foreground">
                Flags any scheduled task that is overdue by more than its configured interval.
              </dd>
            </div>

            <div class="p-3 bg-muted rounded-md">
              <dt class="font-semibold mb-1">Admin Worker</dt>
              <dd class="text-muted-foreground">
                Warns if the Admin Worker is still enabled — it should be disabled in favour of a
                proper message queue worker in production.
              </dd>
            </div>

            <div class="p-3 bg-muted rounded-md">
              <dt class="font-semibold mb-1">Frosh Tools Integration</dt>
              <dd class="text-muted-foreground">
                If
                <a href="https://github.com/FriendsOfShopware/FroshTools" target="_blank" class="underline underline-offset-2 decoration-primary"
                  >FroshTools</a
                >
                is installed, additional checks are displayed including PHP version, MySQL version,
                opcache status, memory limits, queue storage type, and mail configuration.
              </dd>
            </div>
          </div>

          <Alert>
            <AlertDescription>
              You can suppress individual checks per shop using the ignore feature if they are known
              false positives for your environment.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>

      <!-- Extensions -->
      <Card id="extensions" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:plug class="mr-1 inline" />
            Extensions
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0 [&_ol]:mb-4 [&_ol]:pl-6 [&_ol]:list-decimal [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
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
        </CardContent>
      </Card>

      <!-- Scheduled Tasks -->
      <Card id="scheduled-tasks" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:list-check class="mr-1 inline" />
            Scheduled Tasks
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0 [&_ol]:mb-4 [&_ol]:pl-6 [&_ol]:list-decimal [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
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
        </CardContent>
      </Card>

      <!-- Queue -->
      <Card id="queue" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:bars-staggered class="mr-1 inline" />
            Queue Monitoring
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0">
          <p>
            View the message queue sizes of your Shopware shop. This helps you identify if messages
            are piling up and workers aren't keeping up.
          </p>
        </CardContent>
      </Card>

      <!-- Sitespeed -->
      <Card id="sitespeed" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:rocket class="mr-1 inline" />
            Performance Monitoring (Sitespeed)
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0 [&_ol]:mb-4 [&_ol]:pl-6 [&_ol]:list-decimal [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
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
        </CardContent>
      </Card>

      <!-- Deployments -->
      <Card id="deployments" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:code-branch class="mr-1 inline" />
            Deployments
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0 [&_ol]:mb-4 [&_ol]:pl-6 [&_ol]:list-decimal [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
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

          <Alert>
            <AlertDescription>
              Sitespeed measurements are automatically tagged with the latest deployment, allowing you
              to correlate performance changes with specific releases.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>

      <!-- Changelog -->
      <Card id="changelog" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:file-waveform class="mr-1 inline" />
            Changelog
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5">
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
        </CardContent>
      </Card>

      <!-- Notifications -->
      <Card id="notifications" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:bell class="mr-1 inline" />
            Notifications
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0">
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
        </CardContent>
      </Card>

      <!-- Shop Token -->
      <Card id="shop-token" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:key class="mr-1 inline" />
            Bypass Authentication
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm">
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
        </CardContent>
      </Card>

      <!-- Packages Mirror -->
      <Card id="packages-mirror" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:cube class="mr-1 inline" />
            Packages Mirror
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0 [&_ol]:mb-4 [&_ol]:pl-6 [&_ol]:list-decimal [&_li]:mb-1.5 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
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

          <Alert>
            <AlertDescription>
              Tokens are validated against <code class="bg-muted px-1.5 py-0.5 rounded text-sm">packages.shopware.com</code> before being saved.
              Syncing happens automatically every hour, but you can also trigger a manual sync from
              the project settings.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>

      <!-- SSO -->
      <Card id="sso" class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:right-to-bracket class="mr-1 inline" />
            Single Sign-On (SSO)
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5 [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
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
        </CardContent>
      </Card>

      <!-- Scraping Info -->
      <Card class="mb-8 scroll-mt-4">
        <CardHeader>
          <CardTitle>
            <icon-fa6-solid:rotate class="mr-1 inline" />
            How Scraping Works
          </CardTitle>
        </CardHeader>
        <CardContent class="leading-7 [&_p]:mb-4 [&_p:last-child]:mb-0 [&_h4]:text-lg [&_h4]:font-semibold [&_h4]:mt-6 [&_h4]:mb-3 [&_h4:first-child]:mt-0 [&_ul]:mb-4 [&_ul]:pl-6 [&_ul]:list-disc [&_li]:mb-1.5 [&_a]:underline [&_a]:underline-offset-2 [&_a]:decoration-primary">
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
        </CardContent>
      </Card>
    </div>
  </main-container>
</template>

<script setup lang="ts">
const tocItems = [
  { href: "#getting-started", label: "Getting Started" },
  { href: "#connecting-shop", label: "Connecting a Shop" },
  { href: "#organizations", label: "Organizations" },
  { href: "#projects", label: "Projects" },
  { href: "#dashboard-overview", label: "Dashboard Overview" },
  { href: "#health-checks", label: "Health Checks" },
  { href: "#extensions", label: "Extensions" },
  { href: "#scheduled-tasks", label: "Scheduled Tasks" },
  { href: "#queue", label: "Queue Monitoring" },
  { href: "#sitespeed", label: "Performance Monitoring" },
  { href: "#deployments", label: "Deployments" },
  { href: "#changelog", label: "Changelog" },
  { href: "#notifications", label: "Notifications" },
  { href: "#shop-token", label: "Bypass Authentication" },
  { href: "#packages-mirror", label: "Packages Mirror" },
  { href: "#sso", label: "Single Sign-On (SSO)" },
];
</script>
