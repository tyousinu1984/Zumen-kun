# 図面くん API 仕様書

## Go Backend API (Port: 8080)

### ベースURL
```
http://localhost:8080/api/v1
```

### ヘルスチェック
```
GET /health
```
レスポンス例:
```json
{
  "status": "ok",
  "service": "zumen-kun-go-backend",
  "timestamp": "2024-01-01T00:00:00Z",
  "redis": "connected"
}
```

---

### 物件管理 API

#### PDFアップロード & 解析開始
```
POST /api/v1/properties/parse
Content-Type: multipart/form-data

FormData:
  file: (PDF ファイル)
```
レスポンス:
```json
{
  "property_id": "uuid",
  "pdf_url": "https://...",
  "status": "parsing",
  "message": "PDF uploaded and parse task queued successfully"
}
```

#### 物件一覧取得
```
GET /api/v1/properties
```

#### 物件詳細取得
```
GET /api/v1/properties/:id
```

#### 物件作成
```
POST /api/v1/properties
Content-Type: application/json

{
  "name": "物件名",
  "room_number": "101",
  "phone_number": "03-1234-5678",
  "price": 8.5,
  "ad_rate": "1ヶ月",
  "pdf_url": "https://..."
}
```

#### 物件更新
```
PATCH /api/v1/properties/:id
Content-Type: application/json

{
  "name": "新しい物件名",
  "status": "completed",
  "foreigner_ok": true
}
```

#### 物件削除
```
DELETE /api/v1/properties/:id
```

#### AI解析 トリガー（再実行）
```
POST /api/v1/properties/:id/parse
```

#### AI物確電話 トリガー
```
POST /api/v1/properties/:id/call
```

#### 通話記録一覧
```
GET /api/v1/properties/:id/calls
```

---

### ステータス一覧

| Status | 説明 |
|--------|------|
| `pending` | 登録済み・未処理 |
| `parsing` | PDF解析中（Python Worker 処理中） |
| `calling` | AI物確電話中（Retell AI 通話中） |
| `completed` | 処理完了 |
| `parse_failed` | PDF解析失敗 |
| `call_failed` | 物確電話失敗 |

---

## Python Worker API (Port: 8001)

### ヘルスチェック
```
GET /health
```
レスポンス例:
```json
{
  "status": "ok",
  "service": "zumen-kun-python-worker",
  "worker_running": true,
  "redis": "connected",
  "supabase": "configured",
  "openai": "configured",
  "retell": "not configured"
}
```

### Retell AI Webhook
```
POST /webhook/retell
Content-Type: application/json

{
  "call_id": "call_xxx",
  "call_status": "ended",
  "transcript": "通話記録...",
  "recording_url": "https://...",
  "duration_ms": 45000,
  "metadata": {
    "property_id": "uuid"
  }
}
```

---

## Redis キュー仕様 (Go ↔ Python DTO)

### `task:parse_pdf` キュー
```json
{
  "property_id": "uuid",
  "file_url": "https://supabase.../property.pdf",
  "timestamp": 1704067200
}
```

### `task:voice_call` キュー
```json
{
  "property_id": "uuid",
  "phone_number": "03-1234-5678",
  "property_name": "鈴木ハイム",
  "room_number": "102",
  "timestamp": 1704067200
}
```
