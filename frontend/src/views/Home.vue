<template>
  <div class="header">
    <h1>{{ $t('home.title') }}</h1>
  </div>

  <Panel class="panel-intro">
    <h2>{{ $t('home.subtitle') }}</h2>
    {{ $t('home.description') }}
  </Panel>

  <div class="row">
    <Panel class="col-highlights">
      <img src="/home/dashboard.svg" :alt="$t('home.altDashboard')" />
      <h3>{{ $t('home.dashboardOverview') }}</h3>
      <p>
        {{ $t('home.dashboardOverviewDesc') }}
      </p>
    </Panel>

    <Panel class="col-highlights">
      <img src="/home/automatic-performancechecks.svg" :alt="$t('home.altAutoChecks')" />

      <h3>{{ $t('home.autoPerformanceChecks') }}</h3>
      <p>{{ $t('home.autoPerformanceChecksDesc') }}</p>
    </Panel>

    <Panel class="col-highlights">
      <img src="/home/speed.png" :alt="$t('home.altSitespeedLogo')" />
      <h3>{{ $t('home.performanceMonitoring') }}</h3>
      <p>
        {{ $t('home.performanceMonitoringDesc') }}
      </p>
    </Panel>
  </div>

  <div class="row row-center">
    <div class="col-text">
      <h4>{{ $t('home.environmentState') }}</h4>
      <p>
        {{ $t('home.environmentStateDesc') }}
      </p>
    </div>

    <div class="col-img">
      <img :src="getThemeImage('/home/shopmon-dashboard.png')" :alt="$t('home.altDashboard')" />
    </div>
  </div>

  <div class="row row-center">
    <div class="col-img">
      <img
        :src="getThemeImage('/home/shopmon-performance-checks.png')"
        :alt="$t('home.altPerformanceChecks')"
      />
    </div>

    <div class="col-text">
      <h4>{{ $t('home.froshToolsTitle') }}</h4>
      <p>
        {{ $t('home.froshToolsDesc') }}
      </p>
    </div>
  </div>

  <div class="row row-center">
    <div class="col-text">
      <h4>{{ $t('home.extensionsTitle') }}</h4>
      <p>
        {{ $t('home.extensionsDesc') }}
      </p>
    </div>

    <div class="col-img">
      <img :src="getThemeImage('/home/shopmon-extensions.png')" :alt="$t('home.altExtensions')" />
    </div>
  </div>

  <div class="row row-center">
    <div class="col-img">
      <img :src="getThemeImage('/home/shopmon-sitespeed.png')" :alt="$t('home.altSitespeed')" />
    </div>

    <div class="col-text">
      <h4>{{ $t('home.sitespeedTitle') }}</h4>
      <p>
        {{ $t('home.sitespeedDesc') }}
      </p>
    </div>
  </div>

  <div class="row">
    <div>
      <h4>{{ $t('home.andMore') }}</h4>
      <p>
        {{ $t('home.featureList') }}
      </p>
    </div>
  </div>

  <div v-if="sponsors.length" class="row">
    <sponsor-showcase
      :sponsors="sponsors"
      title="Sponsors"
      description="Shopmon is supported by companies that help keep the project free and moving forward for the Shopware community."
      title-tag="h4"
    />
  </div>

  <div class="section-primary">
    <div class="row row-center row-primary">
      <div class="col" style="text-align: right">
        <h4>{{ $t('home.readyTitle') }}</h4>
        <p>
          {{ $t('home.readyDesc') }}
        </p>
      </div>

      <router-link :to="{ name: 'account.register' }" class="btn btn-primary-outline">
        {{ $t('home.startFree') }}
      </router-link>
    </div>
  </div>

  <div class="row">
    <Panel class="col-highlights">
      <icon-fluent:money-off-24-regular class="icon" aria-hidden="true" />
      <h3>{{ $t('home.foreverFree') }}</h3>
      <p>
        {{ $t('home.foreverFreeDesc') }}
      </p>
    </Panel>

    <Panel class="col-highlights">
      <icon-fluent:people-community-12-filled class="icon" aria-hidden="true" />
      <h3>{{ $t('home.communityDriven') }}</h3>
      <p>
        {{ $t('home.communityDrivenDesc') }}
      </p>
    </Panel>

    <Panel class="col-highlights">
      <icon-mdi:github class="icon" aria-hidden="true" />
      <h3>{{ $t('home.openSource') }}</h3>
      <p>
        {{ $t('home.openSourceDesc') }}
      </p>
    </Panel>
  </div>
</template>

<script setup lang="ts">
import SponsorShowcase from "@/components/SponsorShowcase.vue";
import { useDarkMode } from "@/composables/useDarkMode";
import { sponsors } from "@/data/sponsors";

const { getThemeImage } = useDarkMode();
</script>

<style scoped>
h1 {
  color: #fff;
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
  color: #fff;
  background: var(--primary-color);

  h4 {
    color: #fff;
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

  .col-highlights {
    flex: 0 0 auto;
    width: 30.666666%;
    text-align: center;
    padding: 3rem 2rem;

    svg,
    img {
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
      border: 5px solid #fff;
      box-shadow:
        0 1px 3px 0 rgba(0, 0, 0, 0.1),
        0 1px 2px 0 rgba(0, 0, 0, 0.06);

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

.row-center {
  align-items: center;
}

.dark .col-img img {
  box-shadow: unset;
}
</style>
