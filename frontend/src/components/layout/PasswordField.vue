<template>
    <div>
        <div class="relative">
            <label for="password" class="sr-only">Password</label>
        <Field id="password" :name="name" :type="passwordType" autocomplete="current-password" required=""
            class="field field-password" placeholder="Password" :class="{ 'has-error': error }" />
        <div class="absolute right-0 inset-y-0 flex items-center pr-3 cursor-pointer z-10">
            <icon-fa6-solid:eye class="w-[18px]" v-if="passwordType == 'password'" @click="passwordType = 'text'" />
            <icon-fa6-solid:eye-slash class="w-[18px]" v-else @click="passwordType = 'password'" />
        </div>
        </div>

        <div class="text-red-700">
            {{ error }}
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { Form, Field, configure } from 'vee-validate';

defineProps<{ name: string, error: string | undefined }>();

const passwordType = ref('password');
</script>
