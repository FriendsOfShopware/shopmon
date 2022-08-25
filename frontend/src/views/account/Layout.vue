<template>
  <div
    class="
      min-h-full
      flex
      items-center
      justify-center
      py-12
      px-4
      sm:px-6
      lg:px-8
    "
  >
    <div class="max-w-md w-full space-y-8">
      <img
        class="mx-auto h-12 w-auto mb-6"
        src="https://tailwindui.com/img/logos/workflow-mark.svg?color=blue&shade=500"
        alt="Workflow"
      />
      <router-view></router-view>
    </div>
  </div>
</template>
