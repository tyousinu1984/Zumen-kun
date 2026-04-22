"use client";

import { useEffect, useState } from "react";
import {
  listProperties,
  triggerCall,
  triggerParse,
  deleteProperty,
  type Property,
} from "@/lib/api-client";
import { subscribeToProperties } from "@/lib/supabase";
import {
  cn,
  formatDate,
  formatPrice,
  getForeignerColor,
  getForeignerLabel,
  getStatusColor,
  getStatusLabel,
} from "@/lib/utils";

interface PropertyTableProps {
  refreshTrigger?: number;
}

export function PropertyTable({ refreshTrigger }: PropertyTableProps) {
  const [properties, setProperties] = useState<Property[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const fetchProperties = async () => {
    try {
      const data = await listProperties();
      setProperties(data);
      setError(null);
    } catch (err) {
      setError(
        "物件一覧の取得に失敗しました。Go Backendが起動しているか確認してください。"
      );
    } finally {
      setLoading(false);
    }
  };

  // Initial load
  useEffect(() => {
    fetchProperties();
  }, [refreshTrigger]);

  // Supabase Realtime subscription
  useEffect(() => {
    const unsubscribe = subscribeToProperties((payload) => {
      if (payload.eventType === "INSERT") {
        setProperties((prev) => [payload.new as unknown as Property, ...prev]);
      } else if (payload.eventType === "UPDATE") {
        setProperties((prev) =>
          prev.map((p) =>
            p.id === (payload.new as { id: string }).id
              ? { ...p, ...(payload.new as Partial<Property>) }
              : p
          )
        );
      } else if (payload.eventType === "DELETE") {
        setProperties((prev) =>
          prev.filter((p) => p.id !== (payload.old as { id: string }).id)
        );
      }
    });

    return unsubscribe;
  }, []);

  const handleParse = async (propertyId: string) => {
    setActionLoading(propertyId + "-parse");
    try {
      await triggerParse(propertyId);
    } catch (err) {
      alert(err instanceof Error ? err.message : "解析トリガーに失敗しました");
    } finally {
      setActionLoading(null);
      fetchProperties();
    }
  };

  const handleCall = async (propertyId: string) => {
    setActionLoading(propertyId + "-call");
    try {
      await triggerCall(propertyId);
    } catch (err) {
      alert(err instanceof Error ? err.message : "物確トリガーに失敗しました");
    } finally {
      setActionLoading(null);
      fetchProperties();
    }
  };

  const handleDelete = async (propertyId: string, propertyName: string) => {
    if (!confirm(`「${propertyName || "この物件"}」を削除しますか？`)) return;
    try {
      await deleteProperty(propertyId);
      setProperties((prev) => prev.filter((p) => p.id !== propertyId));
    } catch (err) {
      alert("削除に失敗しました");
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="flex flex-col items-center gap-3 text-gray-400">
          <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin" />
          <p>物件一覧を読み込み中...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-amber-50 border border-amber-200 rounded-xl p-6 text-center">
        <p className="text-amber-800 font-medium">⚠️ {error}</p>
        <p className="text-amber-600 text-sm mt-1">
          開発環境: Go Backendを起動してください
        </p>
        <button
          onClick={fetchProperties}
          className="mt-3 px-4 py-2 text-sm bg-amber-100 hover:bg-amber-200 text-amber-800 rounded-lg transition-colors"
        >
          再試行
        </button>
      </div>
    );
  }

  if (properties.length === 0) {
    return (
      <div className="bg-white border border-gray-200 rounded-xl p-12 text-center">
        <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <svg className="w-8 h-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
        </div>
        <p className="text-gray-500 font-medium">物件がまだ登録されていません</p>
        <p className="text-gray-400 text-sm mt-1">
          上のアップロードエリアからPDFをアップロードしてください
        </p>
      </div>
    );
  }

  return (
    <div className="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="text-left px-4 py-3 font-semibold text-gray-600">物件名</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-600">号室</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-600">賃料</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-600">AD</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-600">ステータス</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-600">外国人可否</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-600">登録日時</th>
              <th className="text-left px-4 py-3 font-semibold text-gray-600">アクション</th>
            </tr>
          </thead>
          <tbody>
            {properties.map((property) => (
              <tr
                key={property.id}
                className="border-b border-gray-100 hover:bg-gray-50 transition-colors"
              >
                <td className="px-4 py-3">
                  <div className="font-medium text-gray-900">
                    {property.name || (
                      <span className="text-gray-400 italic text-xs">解析待ち</span>
                    )}
                  </div>
                  {property.phone_number && (
                    <div className="text-xs text-gray-400 mt-0.5">
                      📞 {property.phone_number}
                    </div>
                  )}
                </td>
                <td className="px-4 py-3 text-gray-700">
                  {property.room_number || "---"}
                </td>
                <td className="px-4 py-3 font-medium text-gray-900">
                  {formatPrice(property.price)}
                </td>
                <td className="px-4 py-3">
                  {property.ad_rate ? (
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-700">
                      {property.ad_rate}
                    </span>
                  ) : (
                    <span className="text-gray-300">---</span>
                  )}
                </td>
                <td className="px-4 py-3">
                  <span
                    className={cn(
                      "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium",
                      getStatusColor(property.status),
                      (property.status === "parsing" || property.status === "calling") &&
                        "status-active"
                    )}
                  >
                    {property.status === "parsing" && (
                      <span className="w-1.5 h-1.5 bg-blue-500 rounded-full mr-1.5 animate-pulse" />
                    )}
                    {property.status === "calling" && (
                      <span className="w-1.5 h-1.5 bg-yellow-500 rounded-full mr-1.5 animate-pulse" />
                    )}
                    {getStatusLabel(property.status)}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span
                    className={cn(
                      "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium",
                      getForeignerColor(property.foreigner_ok)
                    )}
                  >
                    {getForeignerLabel(property.foreigner_ok)}
                  </span>
                </td>
                <td className="px-4 py-3 text-gray-400 text-xs">
                  {formatDate(property.created_at)}
                </td>
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    {/* Re-parse button */}
                    {property.pdf_url && property.status !== "parsing" && (
                      <button
                        onClick={() => handleParse(property.id)}
                        disabled={actionLoading === property.id + "-parse"}
                        className="px-2.5 py-1.5 text-xs font-medium bg-blue-50 hover:bg-blue-100 text-blue-700 rounded-lg transition-colors disabled:opacity-50"
                        title="AI解析を再実行"
                      >
                        {actionLoading === property.id + "-parse" ? "..." : "📄 解析"}
                      </button>
                    )}

                    {/* AI物確 button */}
                    {property.phone_number && property.status !== "calling" && (
                      <button
                        onClick={() => handleCall(property.id)}
                        disabled={actionLoading === property.id + "-call"}
                        className="px-2.5 py-1.5 text-xs font-medium bg-green-50 hover:bg-green-100 text-green-700 rounded-lg transition-colors disabled:opacity-50"
                        title="AIが自動で物確電話をかけます"
                      >
                        {actionLoading === property.id + "-call" ? "..." : "📞 AI物確"}
                      </button>
                    )}

                    {/* Delete button */}
                    <button
                      onClick={() => handleDelete(property.id, property.name)}
                      className="px-2.5 py-1.5 text-xs font-medium bg-red-50 hover:bg-red-100 text-red-700 rounded-lg transition-colors"
                      title="削除"
                    >
                      🗑
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
