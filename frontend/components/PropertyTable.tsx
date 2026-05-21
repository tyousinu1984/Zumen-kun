"use client";

import { useEffect, useState } from "react";
import {
  listProperties,
  triggerCall,
  cancelCall,
  triggerParse,
  deleteProperty,
  deleteBulkProperties,
  getPdfUrl,
  type Property,
} from "@/lib/api-client";
import { subscribeToProperties } from "@/lib/supabase";
import { PDFModal } from "@/components/PDFModal";
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
  onSelectionChange?: (ids: string[]) => void;
  onDataChange?: (properties: Property[]) => void;
}

export function PropertyTable({ refreshTrigger, onSelectionChange, onDataChange }: PropertyTableProps) {
  const [properties, setProperties] = useState<Property[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [bulkDeleting, setBulkDeleting] = useState(false);
  const [pdfModal, setPdfModal] = useState<{ propertyId: string; propertyName: string; pageNumber?: number } | null>(null);

  const fetchProperties = async () => {
    try {
      const data = await listProperties();
      setProperties(data);
      onDataChange?.(data);
      setError(null);
    } catch (err) {
      setError(
        "物件一覧の取得に失敗しました。Go Backendが起動しているか確認してください。"
      );
    } finally {
      setLoading(false);
    }
  };

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

  // Polling fallback
  useEffect(() => {
    const interval = setInterval(() => { fetchProperties(); }, 5000);
    return () => clearInterval(interval);
  }, []);

  // 選択状態が変わったら親に通知
  useEffect(() => {
    onSelectionChange?.(Array.from(selectedIds));
  }, [selectedIds, onSelectionChange]);

  // ─── チェックボックス操作 ───
  const isAllSelected = properties.length > 0 && selectedIds.size === properties.length;
  const isPartialSelected = selectedIds.size > 0 && selectedIds.size < properties.length;

  const toggleAll = () => {
    if (isAllSelected) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(properties.map((p) => p.id)));
    }
  };

  const toggleOne = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  // ─── 個別削除 ───
  const handleDelete = async (propertyId: string, propertyName: string) => {
    if (!confirm(`「${propertyName || "この物件"}」を削除しますか？`)) return;
    try {
      await deleteProperty(propertyId);
      setProperties((prev) => prev.filter((p) => p.id !== propertyId));
      setSelectedIds((prev) => { const n = new Set(prev); n.delete(propertyId); return n; });
    } catch (err) {
      alert("削除に失敗しました");
    }
  };

  // ─── 選択削除 ───
  const handleBulkDelete = async () => {
    const ids = Array.from(selectedIds);
    if (ids.length === 0) return;
    if (!confirm(`選択した ${ids.length} 件を削除しますか？`)) return;
    setBulkDeleting(true);
    try {
      await deleteBulkProperties(ids);
      setProperties((prev) => prev.filter((p) => !selectedIds.has(p.id)));
      setSelectedIds(new Set());
    } catch (err) {
      alert("一括削除に失敗しました");
    } finally {
      setBulkDeleting(false);
    }
  };

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
        <p className="text-amber-600 text-sm mt-1">開発環境: Go Backendを起動してください</p>
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
        <p className="text-gray-400 text-sm mt-1">上のアップロードエリアからPDFをアップロードしてください</p>
      </div>
    );
  }

  return (
    <>
    {/* PDF プレビューモーダル */}
    {pdfModal && (
      <PDFModal
        propertyId={pdfModal.propertyId}
        propertyName={pdfModal.propertyName}
        pdfUrl={getPdfUrl(pdfModal.propertyId)}
        pageNumber={pdfModal.pageNumber}
        onClose={() => setPdfModal(null)}
      />
    )}
    <div className="space-y-2">
      {/* 一括操作バー */}
      {selectedIds.size > 0 && (
        <div className="flex items-center gap-3 px-4 py-2.5 bg-blue-50 border border-blue-200 rounded-xl">
          <span className="text-sm font-medium text-blue-800">
            {selectedIds.size} 件を選択中
          </span>
          <button
            onClick={handleBulkDelete}
            disabled={bulkDeleting}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-red-500 hover:bg-red-600 text-white rounded-lg transition-colors disabled:opacity-50"
          >
            {bulkDeleting ? (
              <span className="w-3 h-3 border-2 border-white border-t-transparent rounded-full animate-spin" />
            ) : (
              "🗑"
            )}
            選択削除 ({selectedIds.size}件)
          </button>
          <button
            onClick={() => setSelectedIds(new Set())}
            className="text-xs text-blue-600 hover:text-blue-800 transition-colors"
          >
            選択解除
          </button>
        </div>
      )}

      {/* テーブル */}
      <div className="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-200">
                {/* 全選択チェックボックス */}
                <th className="px-4 py-3 w-10">
                    <input
                      type="checkbox"
                      aria-label="全件選択"
                      checked={isAllSelected}
                      ref={(el) => { if (el) el.indeterminate = isPartialSelected; }}
                      onChange={toggleAll}
                      className="w-4 h-4 rounded border-gray-300 text-blue-600 cursor-pointer"
                    />
                </th>
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
                  className={cn(
                    "border-b border-gray-100 hover:bg-gray-50 transition-colors",
                    selectedIds.has(property.id) && "bg-blue-50 hover:bg-blue-50"
                  )}
                >
                  {/* 行チェックボックス */}
                  <td className="px-4 py-3">
                    <input
                      type="checkbox"
                      aria-label={`${property.name || property.id} を選択`}
                      checked={selectedIds.has(property.id)}
                      onChange={() => toggleOne(property.id)}
                      className="w-4 h-4 rounded border-gray-300 text-blue-600 cursor-pointer"
                    />
                  </td>
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
                    <div className="flex items-center gap-1.5 flex-wrap">
                      {/* 図面プレビューボタン:
                          - pdf_url あり → そのまま表示（page_number があればジャンプ）
                          - source_pdf_id あり → 元PDFのIDでアクセス + ページジャンプ
                      */}
                      {(property.pdf_url || property.source_pdf_id) && (() => {
                        // 表示に使うpropertyId（元PDFのID or 自身のID）
                        const pdfPropertyId = property.source_pdf_id || property.id;
                        return (
                          <button
                            onClick={() => setPdfModal({
                              propertyId: pdfPropertyId,
                              propertyName: property.name,
                              pageNumber: property.page_number || undefined,
                            })}
                            className="px-2.5 py-1.5 text-xs font-medium bg-gray-50 hover:bg-gray-100 text-gray-600 rounded-lg transition-colors border border-gray-200"
                            title={property.page_number ? `図面をプレビュー（${property.page_number}ページ目）` : "図面をプレビュー"}
                          >
                            🔍 図面{property.page_number ? ` p.${property.page_number}` : ""}
                          </button>
                        );
                      })()}
                      {/* AI解析ボタン */}
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
                      {/* AI物確ボタン (calling中は非表示) */}
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
                      {/* キャンセルボタン (calling中のみ表示) */}
                      {property.status === "calling" && (
                        <button
                          onClick={async () => {
                            setActionLoading(property.id + "-cancel");
                            try {
                              await cancelCall(property.id);
                              fetchProperties();
                            } catch {
                              alert("キャンセルに失敗しました");
                            } finally {
                              setActionLoading(null);
                            }
                          }}
                          disabled={actionLoading === property.id + "-cancel"}
                          className="px-2.5 py-1.5 text-xs font-medium bg-orange-50 hover:bg-orange-100 text-orange-700 border border-orange-200 rounded-lg transition-colors disabled:opacity-50"
                          title="物確電話をキャンセル"
                        >
                          {actionLoading === property.id + "-cancel" ? "..." : "🛑 キャンセル"}
                        </button>
                      )}
                      {/* 削除ボタン */}
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
    </div>
    </>
  );
}
