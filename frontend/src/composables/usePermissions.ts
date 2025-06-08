import { authClient } from '@/helpers/auth-client';
import { type Ref, ref, watch } from 'vue';

// Cache for permission results
const permissionCache = new Map<
    string,
    { allowed: boolean; timestamp: number }
>();

// Cache TTL in milliseconds (5 minutes)
const CACHE_TTL = 5 * 60 * 1000;

interface PermissionCheck {
    organizationId?: string | null;
    permissions: Parameters<
        typeof authClient.organization.hasPermission
    >[0]['permissions'];
}

export function usePermissions(
    permissionCheck: Ref<PermissionCheck> | PermissionCheck,
) {
    const isLoading = ref(false);
    const allowed = ref(false);

    // Create a cache key from the permission check
    const createCacheKey = (check: PermissionCheck): string => {
        return JSON.stringify({
            orgId: check.organizationId,
            permissions: check.permissions,
        });
    };

    // Check if cached result is still valid
    const getCachedResult = (key: string) => {
        const cached = permissionCache.get(key);
        if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
            return cached.allowed;
        }
        return null;
    };

    // Store result in cache
    const setCachedResult = (key: string, result: boolean) => {
        permissionCache.set(key, {
            allowed: result,
            timestamp: Date.now(),
        });
    };

    // Get the current permission check value (reactive or static)
    const getCurrentCheck = (): PermissionCheck => {
        return 'value' in permissionCheck
            ? permissionCheck.value
            : permissionCheck;
    };

    // Perform the permission check
    const checkPermissions = async () => {
        const currentCheck = getCurrentCheck();

        // If organizationId is null or undefined, set allowed to false and don't make API call
        if (!currentCheck.organizationId) {
            allowed.value = false;
            isLoading.value = false;
            return;
        }

        const cacheKey = createCacheKey(currentCheck);

        // Check cache first
        const cachedResult = getCachedResult(cacheKey);
        if (cachedResult !== null) {
            allowed.value = cachedResult;
            return;
        }

        isLoading.value = true;

        try {
            const response = await authClient.organization.hasPermission({
                organizationId: currentCheck.organizationId,
                permissions: currentCheck.permissions,
            });

            const result = response.data?.success ?? false;
            allowed.value = result;

            // Cache the result
            setCachedResult(cacheKey, result);
        } finally {
            isLoading.value = false;
        }
    };

    // Watch for changes in permission check parameters, especially organizationId
    watch(
        () => {
            const currentCheck = getCurrentCheck();
            return [
                currentCheck.organizationId,
                JSON.stringify(currentCheck.permissions),
            ];
        },
        (newValues, oldValues) => {
            // Always check permissions when organizationId changes from null to a value
            // or when permissions change
            if (
                !oldValues ||
                newValues[0] !== oldValues[0] ||
                newValues[1] !== oldValues[1]
            ) {
                checkPermissions();
            }
        },
        { immediate: true },
    );

    return allowed;
}
