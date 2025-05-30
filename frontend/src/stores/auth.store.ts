import { authClient } from '@/helpers/auth-client';
import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
const session = authClient.useSession();

export const useAuthStore = defineStore('auth', () => {
    const userAvatar = ref(
        'https://api.dicebear.com/7.x/personas/svg/?seed=default?d=identicon',
    );

    setAvatar();

    watch(session, async () => {
        await Promise.all([setAvatar()]);
    });

    const returnUrl = ref<string | null>(null);

    async function setAvatar() {
        if (session.value.data?.user) {
            try {
                const userEmailSha256 = Array.from(
                    new Uint8Array(
                        await crypto.subtle.digest(
                            'SHA-256',
                            new TextEncoder().encode(
                                session.value.data.user.email,
                            ),
                        ),
                    ),
                )
                    .map((b) => b.toString(16).padStart(2, '0'))
                    .join('');

                userAvatar.value = `https://api.dicebear.com/7.x/personas/svg/?seed=${userEmailSha256}?d=identicon`;
            } catch (error) {
                console.error('Error generating user avatar:', error);
            }
        }
    }

    return {
        userAvatar,
        returnUrl,
    };
});
