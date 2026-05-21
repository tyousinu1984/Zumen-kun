import type { NextConfig } from "next";

// ============================================================
// API URL 設計方針
// ============================================================
// ブラウザからは常に /api/backend/* を叩く（Next.js プロキシ経由）
// Next.js サーバーサイドが API_URL（ランタイム環境変数）に転送する
//
// 環境別の設定:
//   ローカル: API_URL=http://go-backend:8080  (Dockerネットワーク内)
//   staging:  API_URL=https://api.staging.zumen-kun.com
//   本番:     API_URL=https://api.zumen-kun.com
//
// NEXT_PUBLIC_API_URL はもう不要（ビルド時に焼き込み不要）
// ============================================================

const nextConfig: NextConfig = {
  output: "standalone",
  experimental: {
    serverActions: {
      allowedOrigins: ["localhost:3000"],
    },
  },
  // rewrites は削除 → app/api/backend/[...path]/route.ts で処理
  // Route Handler はランタイムで API_URL を解決するため環境分離が正しく動く
};

export default nextConfig;
