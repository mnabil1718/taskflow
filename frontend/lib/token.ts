const ACCESS_KEY = "tf_access";
const REFRESH_KEY = "tf_refresh";

function setCookie(name: string, value: string, maxAge: number) {
  document.cookie = `${name}=${value}; path=/; max-age=${maxAge}; SameSite=Lax`;
}

function deleteCookie(name: string) {
  document.cookie = `${name}=; path=/; max-age=0`;
}

export const tokenStorage = {
  getAccess: (): string | null =>
    typeof window !== "undefined" ? localStorage.getItem(ACCESS_KEY) : null,

  getRefresh: (): string | null =>
    typeof window !== "undefined" ? localStorage.getItem(REFRESH_KEY) : null,

  set: (access: string, refresh: string): void => {
    localStorage.setItem(ACCESS_KEY, access);
    localStorage.setItem(REFRESH_KEY, refresh);
    // Mirror access token to a cookie so Next.js middleware can read it
    setCookie(ACCESS_KEY, access, 15 * 60); // 15 min, matches JWT_ACCESS_EXPIRY
  },

  clear: (): void => {
    localStorage.removeItem(ACCESS_KEY);
    localStorage.removeItem(REFRESH_KEY);
    deleteCookie(ACCESS_KEY);
  },
};
