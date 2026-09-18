"use client";

import { useEffect, useState } from "react";
import { getUser, logout, type AuthUser } from "@/lib/auth";

export function Header() {
  const [user, setUser] = useState<AuthUser | null>(null);

  useEffect(() => {
    setUser(getUser());
  }, []);

  return (
    <header className="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between">
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold text-sm">
          📐
        </div>
        <div>
          <h1 className="text-lg font-bold text-gray-900">図面くん</h1>
          {user?.company_id && (
            <p className="text-xs text-gray-400">AI不動産物確プラットフォーム</p>
          )}
        </div>
      </div>

      {user ? (
        <div className="flex items-center gap-4">
          <div className="text-right">
            <p className="text-sm font-medium text-gray-900">
              {user.display_name || user.email}
            </p>
            <p className="text-xs text-gray-400 capitalize">
              {user.role === "admin" ? "🔑 管理者" : "👤 ユーザー"}
            </p>
          </div>
          <button
            onClick={logout}
            className="px-3 py-1.5 text-xs font-medium text-gray-600 hover:text-red-600 bg-gray-100 hover:bg-red-50 rounded-lg transition-colors border border-gray-200"
          >
            ログアウト
          </button>
        </div>
      ) : (
        <a
          href="/login"
          className="px-4 py-2 text-sm font-medium bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
        >
          ログイン
        </a>
      )}
    </header>
  );
}
