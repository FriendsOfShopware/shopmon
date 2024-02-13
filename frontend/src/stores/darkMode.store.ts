import { defineStore } from 'pinia';

interface DarkModeState {
  darkMode: boolean;
}

export const useDarkModeStore = defineStore('darkMode', {
  state: (): DarkModeState => ({
    darkMode: window.matchMedia('(prefers-color-scheme: dark)').matches,
  }),
  
  actions: {
    setDarkMode(darkMode: boolean) {
        this.darkMode = darkMode;
        this.updateDarkModeClass();
    },

    toggleDarkMode(): void {
      this.setDarkMode(!this.darkMode);
    },

    updateDarkModeClass(): void {
        if (this.darkMode) {
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }
    },
  }
});
