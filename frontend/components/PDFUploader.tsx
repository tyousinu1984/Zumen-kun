"use client";

import { useCallback, useState } from "react";
import { useDropzone } from "react-dropzone";
import { uploadAndParsePDF } from "@/lib/api-client";

interface PDFUploaderProps {
  onSuccess?: (propertyId: string) => void;
}

export function PDFUploader({ onSuccess }: PDFUploaderProps) {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const onDrop = useCallback(
    async (acceptedFiles: File[]) => {
      const file = acceptedFiles[0];
      if (!file) return;

      setUploading(true);
      setError(null);
      setSuccess(null);

      try {
        const result = await uploadAndParsePDF(file);
        setSuccess(
          `✅ アップロード完了！物件ID: ${result.property_id.slice(0, 8)}... AI解析を開始しました。`
        );
        onSuccess?.(result.property_id);
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : "アップロードに失敗しました。Go Backendが起動しているか確認してください。"
        );
      } finally {
        setUploading(false);
      }
    },
    [onSuccess]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: { "application/pdf": [".pdf"] },
    maxFiles: 1,
    disabled: uploading,
  });

  return (
    <div className="space-y-3">
      <div
        {...getRootProps()}
        className={`
          border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-all
          ${isDragActive
            ? "border-blue-500 bg-blue-50 scale-[1.01]"
            : "border-gray-300 bg-white hover:border-blue-400 hover:bg-blue-50"
          }
          ${uploading ? "opacity-60 cursor-not-allowed" : ""}
        `}
      >
        <input {...getInputProps()} />

        <div className="flex flex-col items-center gap-3">
          {uploading ? (
            <>
              <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin" />
              <p className="text-blue-600 font-medium">アップロード中...</p>
              <p className="text-sm text-gray-400">AI解析のキューに追加しています</p>
            </>
          ) : (
            <>
              <div className="w-14 h-14 bg-blue-50 rounded-full flex items-center justify-center">
                <svg
                  className="w-7 h-7 text-blue-500"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={1.5}
                    d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                  />
                </svg>
              </div>
              {isDragActive ? (
                <p className="text-blue-600 font-semibold">ここにドロップ！</p>
              ) : (
                <>
                  <p className="text-gray-700 font-medium">
                    物件図面PDFをドラッグ＆ドロップ
                  </p>
                  <p className="text-sm text-gray-400">
                    またはクリックしてファイルを選択
                  </p>
                  <p className="text-xs text-gray-300">PDFファイルのみ対応</p>
                </>
              )}
            </>
          )}
        </div>
      </div>

      {success && (
        <div className="flex items-start gap-2 p-3 bg-green-50 border border-green-200 rounded-lg text-sm text-green-700">
          <span className="flex-shrink-0">✅</span>
          <span>{success}</span>
        </div>
      )}

      {error && (
        <div className="flex items-start gap-2 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
          <span className="flex-shrink-0">❌</span>
          <span>{error}</span>
        </div>
      )}
    </div>
  );
}
