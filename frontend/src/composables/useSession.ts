import { ref, computed } from "vue";
import { getToken } from "../helpers/api";
import { getSession, setActiveOrganization as apiSetActiveOrganization } from "@/api/generated";

interface SessionUser {
  id: string;
  name: string;
  email: string;
  emailVerified: boolean;
  image: string | null;
  role: string;
  notifications?: string[];
}

interface SessionInfo {
  id: string;
  userId: string;
  expiresAt: string;
  activeOrganizationId: string | null;
  impersonatedBy?: string;
  token?: string;
}

export interface SessionData {
  user: SessionUser;
  session: SessionInfo;
}

const session = ref<SessionData | null>(null);
const loading = ref(true);
const _fetched = ref(false);
let _pending: Promise<SessionData | null> | null = null;

export async function fetchSession(): Promise<SessionData | null> {
  if (_pending) {
    return _pending;
  }

  const token = getToken();
  if (!token) {
    session.value = null;
    loading.value = false;
    _fetched.value = true;
    return null;
  }

  loading.value = true;
  _pending = getSession().then(({ data, error }) => {
    if (error || !data?.user) {
      session.value = null;
    } else {
      session.value = data as SessionData;
    }
    loading.value = false;
    _fetched.value = true;
    _pending = null;
    return session.value;
  });
  return _pending;
}

export async function setActiveOrganization(organizationId: string): Promise<void> {
  await apiSetActiveOrganization({
    body: { organizationId },
  });
  _fetched.value = false;
  _pending = null;
  await fetchSession();
}

export function useSession() {
  if (!_fetched.value) {
    void fetchSession();
  }

  const activeOrganizationId = computed(() => session.value?.session.activeOrganizationId ?? null);

  return { session, loading, fetchSession, activeOrganizationId, setActiveOrganization };
}

export function clearSession() {
  session.value = null;
}
