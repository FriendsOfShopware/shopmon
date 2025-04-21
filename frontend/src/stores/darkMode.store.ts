import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useDarkModeStore = defineStore('darkMode', () => {
    const darkMode = ref(window.matchMedia('(prefers-color-scheme: dark)').matches);

    function setDarkMode(isDarkMode: boolean) {
        darkMode.value = isDarkMode;
        updateDarkModeClass();
    }

    function toggleDarkMode(): void {
        setDarkMode(!darkMode.value);
    }

    function updateDarkModeClass(): void {
        if (darkMode.value) {
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }
    }

    return {
        darkMode,
        setDarkMode,
        toggleDarkMode,
        updateDarkModeClass
    };
});
