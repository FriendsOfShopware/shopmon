import { type Ref, ref, watch, computed } from 'vue';

const DARK_MODE_STORAGE_KEY = 'shopmon-dark-mode';

// Shared state across all components
const isDarkMode = ref(false);
let initialized = false;

export function useDarkMode(): {
    darkMode: Ref<boolean>;
    toggleDarkMode: () => void;
    getThemeImage: (imagePath: string) => Ref<string>;
} {
    // Initialize only once
    if (!initialized) {
        initialized = true;
        initializeDarkMode();
    }

    function toggleDarkMode() {
        isDarkMode.value = !isDarkMode.value;
    }

    function getThemeImage(imagePath: string) {
        const computedImg = computed(() => {
            if (!isDarkMode.value) {
                return imagePath;
            }

            const lastDotIndex = imagePath.lastIndexOf('.');
            const pathWithoutExtension = imagePath.substring(0, lastDotIndex);
            const extension = imagePath.substring(lastDotIndex);
            return `${pathWithoutExtension}-dark${extension}`;
        });
        
        return computedImg.value;
    }

    return {
        darkMode: isDarkMode,
        toggleDarkMode,
        getThemeImage,
    };
}

function initializeDarkMode() {
    // Get initial preference
    const stored = localStorage.getItem(DARK_MODE_STORAGE_KEY);
    if (stored !== null) {
        isDarkMode.value = stored === 'true';
    } else {
        isDarkMode.value = window.matchMedia(
            '(prefers-color-scheme: dark)',
        ).matches;
    }

    // Apply initial state
    updateDarkModeClass(isDarkMode.value);

    // Watch for changes
    watch(isDarkMode, (newValue) => {
        localStorage.setItem(DARK_MODE_STORAGE_KEY, String(newValue));
        updateDarkModeClass(newValue);
    });

    // Listen for system preference changes
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', (e) => {
        // Only update if user hasn't set a preference
        if (localStorage.getItem(DARK_MODE_STORAGE_KEY) === null) {
            isDarkMode.value = e.matches;
        }
    });

    // Listen for changes in other tabs
    window.addEventListener('storage', (e) => {
        if (e.key === DARK_MODE_STORAGE_KEY && e.newValue !== null) {
            isDarkMode.value = e.newValue === 'true';
        }
    });
}

function updateDarkModeClass(isDark: boolean): void {
    if (isDark) {
        document.documentElement.classList.add('dark');
    } else {
        document.documentElement.classList.remove('dark');
    }
}
