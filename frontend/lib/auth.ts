// 認証ユーティリティ（localStorage ベース）

export interface AuthUser {
  user_id: string;
  email: string;
  role: string;
  company_id: string;
  display_name: string;
}

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("zumen_token");
}

export function getUser(): AuthUser | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem("zumen_user");
  if (!raw) return null;
  try {
    return JSON.parse(raw) as AuthUser;
  } catch {
    return null;
  }
}

export function logout() {
  localStorage.removeItem("zumen_token");
  localStorage.removeItem("zumen_user");
  window.location.href = "/login";
}

export function isLoggedIn(): boolean {
  return !!getToken();
}

/** API リクエスト用ヘッダー（JWT 付き） */
export function authHeaders(): Record<string, string> {
  const token = getToken();
  return token
    ? { Authorization: `Bearer ${token}`, "Content-Type": "application/json" }
    : { "Content-Type": "application/json" };
}
