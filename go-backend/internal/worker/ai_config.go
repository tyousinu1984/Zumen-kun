package worker

// ============================================================
// 🤖 AI設定ファイル - ここを編集するだけでAIの動作が変わります
// AI Configuration - Edit this file to change AI behavior
// ============================================================

// Gemini モデル設定
const (
	// GeminiModel - 使用するGeminiモデル名
	// 利用可能なモデル一覧: https://ai.google.dev/gemini-api/docs/models
	//
	// 推奨モデル（2026年5月時点 v1 API 利用可能モデル）:
	//   "gemini-2.5-flash"   - 最新・高精度 (PDF直接対応・推奨)
	//   "gemini-2.0-flash"   - 高速・軽量
	//   "gemini-2.0-flash-lite" - 超高速・低コスト
	GeminiModel = "gemini-2.5-flash"

	// GeminiAPIVersion - Gemini API バージョン
	// "v1"      - 安定版（推奨）
	// "v1beta"  - ベータ版（新機能あり・不安定の可能性）
	GeminiAPIVersion = "v1"

	// GeminiMaxOutputTokens - レスポンスの最大トークン数
	// 大きなPDF（多ページ）の場合は増やしてください
	GeminiMaxOutputTokens = 8192

	// GeminiTimeoutSeconds - Gemini API タイムアウト（秒）
	// 大きなPDFは時間がかかるため余裕を持たせています
	GeminiTimeoutSeconds = 300 // 5分
)

// PDF解析設定
const (
	// PDFMaxSizeBytes - アップロード可能なPDFの最大サイズ
	// デフォルト: 50MB
	PDFMaxSizeBytes = 50 * 1024 * 1024

	// PDFMimeType - PDFのMIMEタイプ（変更不要）
	PDFMimeType = "application/pdf"
)
