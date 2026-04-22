"use client";

import { useCallback, useEffect, useState } from "react";
import { listProperties, type Property } from "@/lib/api-client";
import { PDFUploader } from "@/components/PDFUploader";
import { PropertyTable } from "@/components/PropertyTable";
import { StatsCards } from "@/components/StatsCards";
import { subscribeToProperties } from "@/lib/supabase";

export default function DashboardPage() {
  const [properties, setProperties] = useState<Property[]>([]);
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [realtimeStatus, setRealtimeStatus] = useState<"connecting" | "connected" | "disconnected">("connecting");

  // Fetch properties for stats
  const fetchProperties = useCallback(async () => {
    try {
      const data = await listProperties();
      setProperties(data);
    } catch {
      // Ignore errors here - PropertyTable shows its own error state
    }
  }, []);

  useEffect(() => {
    fetchProperties();
  }, [fetchProperties, refreshTrigger]);

  // Realtime subscription
  useEffect(() => {
    const unsubscribe = subscribeToProperties((payload) => {
      setRealtimeStatus("connected");
      fetchProperties();
    });

    // If supabase is not configured, still show UI
    const timer = setTimeout(() => {
      setRealtimeStatus("disconnected");
    }, 3000);

    return () => {
      unsubscribe();
      clearTimeout(timer);
    };
  }, [fetchProperties]);

  const handleUploadSuccess = (propertyId: string) => {
    // Trigger refresh after upload
    setTimeout(() => {
      setRefreshTrigger((t) => t + 1);
    }, 1000);
  };

  return (
    <div className="space-y-8">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">物件管理ダッシュボード</h2>
          <p className="text-gray-500 mt-1">
            PDFをアップロードしてAI自動解析・物確電話を実行します
          </p>
        </div>

        {/* Realtime indicator */}
        <div className="flex items-center gap-2 px-3 py-1.5 bg-white rounded-full border border-gray-200 text-xs">
          <span
            className={`w-2 h-2 rounded-full ${
              realtimeStatus === "connected"
                ? "bg-green-500 animate-pulse"
                : realtimeStatus === "connecting"
                ? "bg-yellow-500 animate-pulse"
                : "bg-gray-400"
            }`}
          />
          <span className="text-gray-600">
            {realtimeStatus === "connected"
              ? "リアルタイム更新中"
              : realtimeStatus === "connecting"
              ? "接続中..."
              : "Supabase未設定（ポーリング無効）"}
          </span>
        </div>
      </div>

      {/* Stats Cards */}
      <StatsCards properties={properties} />

      {/* PDF Upload Section */}
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
        <div className="flex items-center gap-2 mb-4">
          <div className="w-6 h-6 bg-blue-100 rounded flex items-center justify-center">
            <svg className="w-3.5 h-3.5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
            </svg>
          </div>
          <h3 className="font-semibold text-gray-900">物件図面PDF アップロード</h3>
        </div>

        <PDFUploader onSuccess={handleUploadSuccess} />

        {/* Process flow guide */}
        <div className="mt-4 flex items-center justify-center gap-2 text-xs text-gray-400">
          <span className="flex items-center gap-1">
            <span className="w-5 h-5 bg-gray-100 rounded-full flex items-center justify-center text-gray-600 font-bold">1</span>
            PDFアップロード
          </span>
          <span className="text-gray-300">→</span>
          <span className="flex items-center gap-1">
            <span className="w-5 h-5 bg-blue-100 rounded-full flex items-center justify-center text-blue-600 font-bold">2</span>
            AI図面解析
          </span>
          <span className="text-gray-300">→</span>
          <span className="flex items-center gap-1">
            <span className="w-5 h-5 bg-green-100 rounded-full flex items-center justify-center text-green-600 font-bold">3</span>
            AI物確電話
          </span>
          <span className="text-gray-300">→</span>
          <span className="flex items-center gap-1">
            <span className="w-5 h-5 bg-purple-100 rounded-full flex items-center justify-center text-purple-600 font-bold">4</span>
            外国人可否確定
          </span>
        </div>
      </div>

      {/* Properties Table */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-6 h-6 bg-gray-100 rounded flex items-center justify-center">
              <svg className="w-3.5 h-3.5 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                  d="M3 10h18M3 14h18m-9-4v8m-7 0h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
            </div>
            <h3 className="font-semibold text-gray-900">
              物件一覧{" "}
              <span className="text-gray-400 font-normal text-sm">
                ({properties.length}件)
              </span>
            </h3>
          </div>
          <button
            onClick={() => setRefreshTrigger((t) => t + 1)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-gray-500 hover:text-gray-700 bg-white hover:bg-gray-50 border border-gray-200 rounded-lg transition-colors"
          >
            <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            更新
          </button>
        </div>

        <PropertyTable refreshTrigger={refreshTrigger} />
      </div>

      {/* Development Notice */}
      <div className="bg-amber-50 border border-amber-200 rounded-xl p-4">
        <div className="flex items-start gap-3">
          <span className="text-amber-500 text-lg flex-shrink-0">💡</span>
          <div className="text-sm">
            <p className="font-semibold text-amber-800">開発環境について</p>
            <ul className="mt-1 space-y-1 text-amber-700">
              <li>• Go Backend が未起動の場合、API呼び出しはエラーになります</li>
              <li>• <code className="bg-amber-100 px-1 rounded">docker-compose up</code> で全サービスを起動できます</li>
              <li>• OpenAI APIキー未設定時はモックデータが返されます</li>
              <li>• Retell AI APIキー未設定時は模擬通話が実行されます</li>
              <li>• Supabase未設定時はリアルタイム更新が無効になります</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}
