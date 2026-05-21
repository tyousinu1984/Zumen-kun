"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function LoginPage() {
  const router = useRouter();
  const [mode, setMode] = useState<"login" | "register">("login");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Login form
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  // Register form (追加フィールド)
  const [companyName, setCompanyName] = useState("");
  const [displayName, setDisplayName] = useState("");

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.error || "ログインに失敗しました");
        return;
      }
      localStorage.setItem("zumen_token", data.token);
      localStorage.setItem("zumen_user", JSON.stringify({
        user_id: data.user_id,
        email: data.email,
        role: data.role,
        company_id: data.company_id,
        display_name: data.display_name,
      }));
      router.push("/");
    } catch {
      setError("ネットワークエラーが発生しました");
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email,
          password,
          company_name: companyName,
          display_name: displayName || undefined,
          role: "admin",
        }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.error || "登録に失敗しました");
        return;
      }
      localStorage.setItem("zumen_token", data.token);
      localStorage.setItem("zumen_user", JSON.stringify({
        user_id: data.user_id,
        email: data.email,
        role: data.role,
        company_id: data.company_id,
        display_name: data.display_name,
      }));
      router.push("/");
    } catch {
      setError("ネットワークエラーが発生しました");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-blue-950 to-slate-900 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {/* ロゴ */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-600 rounded-2xl mb-4 shadow-lg">
            <span className="text-3xl">📐</span>
          </div>
          <h1 className="text-3xl font-bold text-white">図面くん</h1>
          <p className="text-slate-400 mt-1 text-sm">不動産AI物確プラットフォーム</p>
        </div>

        {/* カード */}
        <div className="bg-white rounded-2xl shadow-2xl p-8">
          {/* タブ */}
          <div className="flex rounded-xl bg-slate-100 p-1 mb-6">
            <button
              onClick={() => { setMode("login"); setError(null); }}
              className={`flex-1 py-2 text-sm font-medium rounded-lg transition-all ${
                mode === "login"
                  ? "bg-white text-slate-900 shadow"
                  : "text-slate-500 hover:text-slate-700"
              }`}
            >
              ログイン
            </button>
            <button
              onClick={() => { setMode("register"); setError(null); }}
              className={`flex-1 py-2 text-sm font-medium rounded-lg transition-all ${
                mode === "register"
                  ? "bg-white text-slate-900 shadow"
                  : "text-slate-500 hover:text-slate-700"
              }`}
            >
              新規登録
            </button>
          </div>

          {/* エラー */}
          {error && (
            <div className="mb-4 px-4 py-3 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm">
              ⚠️ {error}
            </div>
          )}

          {/* フォーム */}
          <form onSubmit={mode === "login" ? handleLogin : handleRegister} className="space-y-4">
            {mode === "register" && (
              <>
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-1">
                    会社名 <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="text"
                    value={companyName}
                    onChange={(e) => setCompanyName(e.target.value)}
                    placeholder="例：株式会社〇〇不動産"
                    required
                    className="w-full px-4 py-3 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-1">
                    表示名
                  </label>
                  <input
                    type="text"
                    value={displayName}
                    onChange={(e) => setDisplayName(e.target.value)}
                    placeholder="例：田中 太郎"
                    className="w-full px-4 py-3 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
              </>
            )}

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">
                メールアドレス <span className="text-red-500">*</span>
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="example@company.com"
                required
                autoComplete="email"
                className="w-full px-4 py-3 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">
                パスワード <span className="text-red-500">*</span>
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder={mode === "register" ? "6文字以上" : "パスワード"}
                required
                minLength={6}
                autoComplete={mode === "login" ? "current-password" : "new-password"}
                className="w-full px-4 py-3 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-xl transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {loading ? (
                <>
                  <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  処理中...
                </>
              ) : mode === "login" ? (
                "ログイン"
              ) : (
                "アカウント作成"
              )}
            </button>
          </form>
        </div>

        <p className="text-center text-slate-500 text-xs mt-6">
          © 2025 図面くん — AI不動産物確自動化
        </p>
      </div>
    </div>
  );
}
