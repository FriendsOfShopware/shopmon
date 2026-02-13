<template>
    <div class="header">
        <h1>Shopmon - The Open-Source Dashboard for Shopware Developers</h1>
    </div>

    <div class="panel panel-intro">
        <h2>Everything at a glance</h2>
        Monitor all your Shopware environments from one central dashboard. Shop status, performance metrics, extension updates, and security alerts - all in real-time without switching between tools.
    </div>

    <div class="row">
        <div class="col-highlights panel">
            <img src="/home/dashboard.svg" alt="Shopmon Dashboard" />
            <h3>Dashboard overview</h3>
            <p>Maintain a complete overview of all your shops with the most important information using the Shopmon Dashboard</p>
        </div>

        <div class="col-highlights panel">
            <img src="/home/automatic-performancechecks.svg" alt="Automatic Performancechecks"/>

            <h3>Automatic performance checks</h3>
            <p>Get notified when your shop status changes based on Frosh-Tools performance checks</p>
        </div>

        <div class="col-highlights panel">
            <img src="/home/speed.png" alt="Sitespeed.io logo">
            <h3>Performance monitoring</h3>
            <p>Daily automated page speed checks with an overview of the most important metrics over time</p>
        </div>
    </div>

    <div class="row row-center">
        <div class="col-text">
            <h4>Shopware versions & environment state dashboard</h4>
            <p>Stay in control with the Shopmon Dashboard. View all your environments with information like Shopware version, environment status based on Frosh-Tools tests, and the latest changelogs.</p>
        </div>

        <div class="col-img">
            <img :src="getThemeImage('/home/shopmon-dashboard.png')" alt="Shopmon dashboard">
        </div>
    </div>

    <div class="row row-center">
        <div class="col-img">
            <img :src="getThemeImage('/home/shopmon-performance-checks.png')" alt="Shopmon performance checks">
        </div>

        <div class="col-text">
            <h4>Frosh Tools performance checks</h4>
            <p>Get valuable insights and warnings about your environment. When Frosh Tools is installed, performance checks are automatically displayed and integrated into your environment status. If a task gets stuck or a new security update is available, you'll be immediately notified through your environment status. Additionally, you'll receive an email whenever your environment status changes.</p>
        </div>
    </div>

    <div class="row row-center">
        <div class="col-text">
            <h4>Extensions and updates</h4>
            <p>Keep track of all your installed extensions either in the project view or in the comprehensive extension overview. Get all the important information such as available updates, installed versions, and Shopware Store ratings. Connect with other Shopmon users to share solutions and discuss known issues directly and without detours.</p>
        </div>

        <div class="col-img">
            <img :src="getThemeImage('/home/shopmon-extensions.png')" alt="Shopmon extensions">
        </div>
    </div>

    <div class="row row-center">
        <div class="col-img">
            <img :src="getThemeImage('/home/shopmon-sitespeed.png')" alt="Shopmon sitespeed">
        </div>

        <div class="col-text">
            <h4>Automatic sitespeed.io checks</h4>
            <p>If you want, we perform daily sitespeed checks on up to 5 URLs of your environment and show you the most important information directly in an overview.</p>
        </div>
    </div>

    <div class="row">
        <div>
            <h4>And much more ...</h4>
            <p>Passkey Login | GitHub Login | Remote Cache Clear | Scheduled Task Overview | Remote Task Rescheduling | Queue Overview | Environment Changelogs | More Features in Development</p>
        </div>
    </div>

    <div class="section-primary">
        <div class="row row-center row-primary">
            <div class="col" style="text-align:right;">
                <h4>Ready to monitor your Shopware?</h4>
                <p>Join Shopmon to monitor your environments. Start monitoring your first shop in under 2 minutes - no credit card required, no commitments, just pure monitoring.</p>
            </div>

            <router-link :to="{ name: 'account.register' }" class="btn btn-primary-outline">
                Start monitoring for free
            </router-link>
        </div>
    </div>

    <div class="row">
        <div class="col-highlights panel">
            <icon-fluent:money-off-24-regular
                class="icon"
                aria-hidden="true"
            />
            <h3>Forever Free</h3>
            <p>No hidden costs, subscription fees, or limitations. Shopmon is completely free to use for everyone, forever.</p>
        </div>

        <div class="col-highlights panel">
            <icon-fluent:people-community-12-filled
                class="icon"
                aria-hidden="true"
            />
            <h3>Community Driven</h3>
            <p>Built by the Friends of Shopware community for the community. <a href="https://github.com/FriendsOfShopware" target="_blank">Join</a> our growing network of developers contributing daily.</p>
        </div>

        <div class="col-highlights panel">
            <icon-mdi:github
                class="icon"
                aria-hidden="true"
            />
            <h3>Open Source</h3>
            <p>Full transparency into how your data is handled. Inspect, modify, and contribute to the codebase on <a href="https://github.com/FriendsOfShopware/shopmon" target="_blank">GitHub</a>.</p>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useDarkMode } from '@/composables/useDarkMode';
const { getThemeImage } = useDarkMode();
</script>

<style scoped>
h1 {
    color: #FFF;
    font-size: clamp(20px, 46px, 4vw);
}

h2 {
    color: var(--primary-color);
    font-size: 38px;
    margin-bottom: 1.5rem;
}

h3 {
    color: var(--primary-color);
    font-size: 26px;
    margin-bottom: 1.5rem;
}

h4 {
    color: var(--home-heading-color);
    font-size: 32px;
    margin-bottom: 1.5rem;
}

.header {
    width: 80%;
    max-width: 1200px;
    text-align: center;
    margin: 5rem auto;
}

.panel-intro {
    width: 80%;
    max-width: 1200px;
    margin-left: auto;
    margin-right: auto;
    text-align: center;
}

.section-primary {
    padding-top: 5rem;
    margin-bottom: 5rem;
    color: #FFF;
    background: var(--primary-color);

    h4 {
        color: #FFF;
    }
}

.row {
    width: 80%;
    max-width: 1200px;
    margin: 0 auto;
    display: flex;
    gap: 4%;
    padding-bottom: 5rem;

    + .row {
        padding-top: 5rem;
    }

    @media all and (max-width: 800px) {
        flex-direction: column;
    }

    &-center {
        align-items: center;
    }

    .col-highlights {
        flex: 0 0 auto;
        width: 30.666666%;
        text-align: center;
        padding: 3rem 2rem;

        svg, img {
            height: 100px;
            width: 100%;
            object-fit: contain;
            margin-bottom: 2rem;
        }
    }

    .col {
        width: 50%;
    }

    .col-img {
        flex: 0 0 auto;
        width: 58%;

        img {
            border-radius: 1rem;
            border: 5px solid #FFF;
            box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);

            .dark & {
                box-shadow: unset;
            }

            @media all and (max-width: 800px) {
                border-radius: 0.5rem;
            }
        }

        @media all and (max-width: 800px) {
            order: 2;
        }
    }

    .col-text {
        flex: 0 0 auto;
        width: 38%;

        @media all and (max-width: 800px) {
            order: 1;
        }
    }

    @media all and (max-width: 800px) {
        margin-bottom: 0;

        [class^="col-"] {
            flex: 1;
            width: 100%;
            margin-bottom: 2rem;
        }
    }
}
</style>