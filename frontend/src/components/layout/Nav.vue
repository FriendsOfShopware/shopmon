<script setup lang="ts">
import { useAuthStore } from '@/stores/auth.store';
import { useNotificationStore } from '@/stores/notification.store';

import {
    Disclosure,
    DisclosureButton,
    DisclosurePanel,
    Menu,
    MenuButton,
    MenuItem,
    MenuItems,
    Popover, 
    PopoverButton, 
    PopoverPanel,
} from '@headlessui/vue';

import FaGear from '~icons/fa6-solid/gear';
import FaPowerOff from '~icons/fa6-solid/power-off';

const authStore = useAuthStore();
const notificationStore = useNotificationStore();

const navigation = [
  { name: 'Dashboard', route: '/' },
  { name: 'My Shops', route: '/account/shops'},
];

const userNavigation = [
    { name: 'Settings', route: '/account/settings', icon: FaGear, },
    { name: 'Logout', route: '/logout', icon: FaPowerOff },
];
</script>

<template>
  <Disclosure v-if="authStore.user" v-slot="{ open }" as="nav" class="bg-sky-500">

    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex items-center justify-between h-16">
      <div class="flex items-center">
        <div class="flex-shrink-0">
          <img class="h-8 w-8" src="https://tailwindui.com/img/logos/workflow-mark.svg?color=blue&shade=300"
            alt="Workflow">
        </div>
        <div class="hidden md:block">
          <div class="ml-10 flex items-baseline space-x-4">
            <router-link v-for="item in navigation" :key="item.name" :to="item.route" :class="[
              item.route == $route.path
                ? 'bg-sky-600 text-white hover:text-white'
                : 'text-white hover:bg-sky-600 hover:bg-opacity-75 hover:text-white',
              'px-3 py-2 rounded-md text-sm font-medium',
            ]" :aria-current="item.route == $route.path ? 'page' : undefined">
              {{ item.name }}
            </router-link>
          </div>
        </div>
      </div>
      <div class="hidden md:block">
        <div class="ml-4 flex items-center md:ml-6">
          <Popover v-slot="{ open }" class="relative">
            <PopoverButton
              @click="notificationStore.markAllRead"
              class="
              h-8
              w-8
              bg-sky-400
              p-1
              rounded-full
              text-sky-200
              hover:text-white
              focus:outline-none
              flex
              justify-center
              items-center
            ">
              <span class="sr-only">View notifications</span>
              <icon-fa6-solid:bell class="h-5 w-5" aria-hidden="true" />
              <div v-if="notificationStore.unreadNotificationCount > 0" class="absolute -right-2 -top-1 ml-2 bg-red-500 rounded-full px-0.5 py-0.5 min-w-[20px] text-xs font-medium text-white">
                {{ notificationStore.unreadNotificationCount }}
              </div>
            </PopoverButton>

            <transition
              enter-active-class="transition ease-out duration-100"
              enter-from-class="transform opacity-0 scale-95" enter-to-class="transform opacity-100 scale-100"
              leave-active-class="transition ease-in duration-75" leave-from-class="transform opacity-100 scale-100"
              leave-to-class="transform opacity-0 scale-95"
            >
              <PopoverPanel class="absolute right-0 z-10 mt-2 w-screen max-w-[300px] px-4 sm:px-0">
                <div class="overflow-hidden rounded-md shadow-lg bg-white">
                  <div class="p-4 font-medium border-b border-grey-200 flex justify-between">
                    Notifications
                    <button class="flex items-center" @click="notificationStore.deleteAllNotifications">
                      <icon-fa6-solid:trash class="opacity-30 hover:opacity-100" />
                    </button>
                  </div>
                  <ul v-if="notificationStore.notifications.length > 0" class="divide-y divide-gray-200 overflow-y-auto max-h-80">
                    <li v-for="notification in notificationStore.notifications" class="px-4 py-2 flex gap-2 odd:bg-gray-50 hover:bg-sky-50">
                      <div class="shrink-0">
                        <icon-fa6-solid:circle-xmark v-if="notification.level === 'error'" class="text-red-600" />
                        <icon-fa6-solid:circle-info v-else class="text-yellow-400" />
                      </div>
                      <div>
                        <div class="font-medium">
                          {{ notification.title }}
                        </div>                          
                        <div class="text-gray-500">
                          {{ notification.message }}
                        </div>
                        <router-link :to="notification.link" type="a" v-if="notification.link">more ...</router-link>
                      </div>
                      <div>
                        <button @click="notificationStore.deleteNotification(notification.id)">
                          <icon-fa6-solid:xmark v-if="notification.level === 'error'" class="opacity-30 hover:opacity-100" />
                        </button>                        
                      </div>
                      
                    </li>
                  </ul>
                  <div v-else class="p-4">
                    Not much going on here
                  </div>                
                </div>
              </PopoverPanel>
            </transition>
          </Popover>
          

          <!-- Profile dropdown -->
          <Menu as="div" class="ml-3 relative">
            <div>
              <MenuButton class="
                  max-w-xs
                  bg-sky-400
                  rounded-full
                  flex
                  items-center
                  text-sm text-white
                  focus:outline-none
                ">
                <span class="sr-only">Open user menu</span>
                <img class="h-8 w-8 rounded-full" :src="authStore.user.avatar" alt="">
              </MenuButton>
            </div>
            <transition 
              enter-active-class="transition ease-out duration-100"
              enter-from-class="transform opacity-0 scale-95" enter-to-class="transform opacity-100 scale-100"
              leave-active-class="transition ease-in duration-75" leave-from-class="transform opacity-100 scale-100"
              leave-to-class="transform opacity-0 scale-95">
              <MenuItems class="
                  origin-top-right
                  absolute
                  right-0
                  mt-2
                  w-48
                  rounded-md
                  shadow-lg
                  py-1
                  bg-white
                  ring-1 ring-black ring-opacity-5
                  focus:outline-none
                  z-10
                ">
                <div class="px-4 py-2 border-b">
                  <div class="text-base font-medium">
                    {{ authStore.user.username }}
                  </div>
                  <div class="text-sm font-medium text-sky-700">
                    {{ authStore.user.email }}
                  </div>
                </div>
                <MenuItem v-for="item in userNavigation" :key="item.name" v-slot="{ active }">
                <button :class="[
                  active ? 'bg-gray-100' : '',
                  'w-full text-left px-4 py-2 text-sm text-gray-700',
                ]" @click="
                      item.route === '/logout'
                        ? authStore.logout()
                        : $router.push(item.route)
                    ">
                  <component :is="item.icon" class="w-4 h-4 inline-block text-sky-700 mr-1" />
                  {{ item.name }}
                </button>
                </MenuItem>
              </MenuItems>
            </transition>
          </Menu>
        </div>
      </div>

      <div class="-mr-2 flex md:hidden">
        <!-- Mobile menu button -->
        <DisclosureButton class="
            inline-flex
            items-center
            justify-center
            p-2
            text-white
            focus:outline-none
          ">
          <span class="sr-only">Open main menu</span>
          <icon-fa6-solid:bars-staggered v-if="!open" class="block h-6 w-6" aria-hidden="true" />
          <icon-fa6-solid:xmark v-else class="block h-6 w-6" aria-hidden="true" />
        </DisclosureButton>
      </div>
    </div>

    <!-- Mobile menu -->
    <DisclosurePanel class="md:hidden w-full absolute bg-sky-500 z-10 drop-shadow-md">
      <div class="px-2 pt-2 pb-3 space-y-1 sm:px-3">
        <DisclosureButton v-for="item in navigation" :key="item.name" as="a" :class="[
          item.route == $route.path
            ? 'bg-sky-700 text-white'
            : 'text-white hover:bg-sky-400 hover:bg-opacity-75',
          'block px-3 py-2 rounded-md text-base font-medium',
        ]" :aria-current="item.route == $route.path ? 'page' : undefined" @click="$router.push(item.route)">
          {{ item.name }}
        </DisclosureButton>
      </div>
      <div class="pt-4 pb-3 border-t border-sky-400">
        <div class="flex items-center px-5">
          <div class="flex-shrink-0">
            <img class="h-10 w-10 rounded-full" :src="authStore.user.avatar" alt="">
          </div>
          <div class="ml-3">
            <div class="text-base font-medium text-white">
              {{ authStore.user.username }}
            </div>
            <div class="text-sm font-medium text-sky-200">
              {{ authStore.user.email }}
            </div>
          </div>
          <button type="button" class="
              w-8
              h-8
              ml-auto
              bg-sky-400
              p-1
              border-2 border-transparent
              rounded-full
              text-sky-200
              flex
              justify-center
              items-center
              hover:text-white
              focus:outline-none
            ">
            <span class="sr-only">View notifications</span>
            <icon-fa6-solid:bell class="h-5 w-5" aria-hidden="true" />
          </button>
        </div>
        <div class="mt-3 px-2 space-y-1">
          <DisclosureButton v-for="item in userNavigation" :key="item.name" as="a" class="
              block
              px-3
              py-1
              rounded-md
              text-base
              font-medium
              text-white
              hover:bg-sky-400 hover:bg-opacity-75
            " @click="
              item.route === '/logout'
                ? authStore.logout()
                : $router.push(item.route)
            ">
            <component :is="item.icon" class="w-4 h-4 inline-block text-sky-200 mr-1" />
            {{ item.name }}
          </DisclosureButton>
        </div>
      </div>
    </DisclosurePanel>
  </Disclosure>
</template>
