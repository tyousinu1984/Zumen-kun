"use client";

import { useEffect } from "react";

interface PDFModalProps {
  propertyId: string;
  propertyName?: string;
  pdfUrl: string;
  pageNumber?: number; // 表示するページ番号
  onClose: () => void;
}

export function PDFModal({ propertyId, propertyName, pdfUrl, pageNumber, onClose }: PDFModalProps) {
  // #page=N フラグメント付きURL（pageNumber が指定されている場合）
  const iframeSrc = pageNumber && pageNumber > 0
    ? `${pdfUrl}#page=${pageNumber}&toolbar=1&navpanes=0&scrollbar=1`
    : `${pdfUrl}#toolbar=1&navpanes=0&scrollbar=1`;
  const linkHref = pageNumber && pageNumber > 0
    ? `${pdfUrl}#page=${pageNumber}`
    : pdfUrl;
  // ESCキーで閉じる
  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [onClose]);

  // スクロール無効化
  useEffect(() => {
    document.body.style.overflow = "hidden";
    return () => { document.body.style.overflow = ""; };
  }, []);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div className="relative flex flex-col bg-white rounded-2xl shadow-2xl overflow-hidden"
        style={{ width: "min(90vw, 960px)", height: "min(90vh, 860px)" }}
      >
        {/* ヘッダー */}
        <div className="flex items-center justify-between px-5 py-3 bg-gray-50 border-b border-gray-200 flex-shrink-0">
          <div className="flex items-center gap-2">
            <span className="text-lg">📄</span>
            <div>
              <p className="font-semibold text-gray-800 text-sm">
                {propertyName || "物件図面"}
              </p>
              <p className="text-xs text-gray-400">ID: {propertyId.slice(0, 8)}...</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {/* 別タブで開く */}
            <a
              href={pdfUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-blue-600 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 border border-blue-200 rounded-lg transition-colors"
            >
              <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                  d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
              別タブで開く
            </a>
            {/* 閉じるボタン */}
            <button
              onClick={onClose}
              className="flex items-center justify-center w-8 h-8 text-gray-500 hover:text-gray-700 hover:bg-gray-200 rounded-lg transition-colors"
              title="閉じる (ESC)"
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        {/* PDF 表示エリア */}
        <div className="flex-1 bg-gray-100 relative">
          <iframe
            src={iframeSrc}
            className="w-full h-full border-0"
            title={`物件図面 - ${propertyName || propertyId}${pageNumber ? ` (p.${pageNumber})` : ""}`}
          />
        </div>

        {/* フッター */}
        <div className="flex items-center justify-between px-5 py-2 bg-gray-50 border-t border-gray-200 flex-shrink-0">
          <p className="text-xs text-gray-400">
            ESC キーまたは外側クリックで閉じます
          </p>
          <button
            onClick={onClose}
            className="px-4 py-1.5 text-xs text-gray-600 hover:text-gray-800 bg-white hover:bg-gray-100 border border-gray-200 rounded-lg transition-colors"
          >
            閉じる
          </button>
        </div>
      </div>
    </div>
  );
}
