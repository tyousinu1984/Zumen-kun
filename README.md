# 図面くん (Zumen-kun) 🏠

> 日本の不動産中介業向けAI自動化プラットフォーム  
> PDF物件図面の自動解析 + AI音声物確電話

## 📋 機能概要

| 機能 | 説明 |
|------|------|
| 📄 PDF自動解析 | 物件図面PDFをGemini 2.0 Flash Visionで解析し、物件名・号室・電話番号・賃料・ADを自動抽出 |
| 📞 AI物確電話 | Retell AIが管理会社に自動電話して空室状況・外国人可否を確認 |
| ⚡ リアルタイム更新 | Supabase Realtimeで処理状況をページ更新なしに表示 |
| 🌍 外国人可否管理 | AI電話の結果を自動判定してダッシュボードに表示 |
| ☑️ 一括削除 | チェックボックスで複数選択削除 / ワンクリック全削除 |

---

## 🏗️ アーキテクチャ

```
ブラウザ
  ↓ REST + Supabase Realtime
Next.js Frontend (localhost:3000)
  ↓ REST API
Go Backend - Gin (localhost:8082)
  ↓ Gemini 2.0 Flash Vision API (PDF直接送信)
  ↓ Supabase REST API (PostgREST via Kong)
Supabase Stack (localhost:8000)
  ├── PostgreSQL DB  (localhost:5433)
  ├── Kong API Gateway (localhost:8000)
  ├── PostgREST (REST API)
  ├── GoTrue (Auth)
  ├── Realtime (WebSocket)
  └── Studio ダッシュボード (localhost:3001)
```

> **注意**: Python Worker は現在未使用。PDF解析はGoバックエンドがGemini APIを直接呼び出す。

---

## 🚀 クイックスタート

### 1. 環境変数の設定

```bash
cd Zumen-kun
cp .env.example .env
```

`.env` を編集して以下を設定:

```env
# 必須: Google Gemini API (PDF解析)
GEMINI_API_KEY=your_gemini_api_key

# Retell AI (AI物確電話) — 未設定時は模擬通話
RETELL_API_KEY=your_retell_api_key
RETELL_AGENT_ID=your_agent_id

# ローカルSupabase (デフォルト値で動作可能)
# POSTGRES_PASSWORD=postgres
# JWT_SECRET=super-secret-jwt-token-with-at-least-32-characters-long
```

### 2. Docker で起動

```bash
# 全サービスをビルド＆起動（初回は5〜10分かかります）
docker-compose up -d --build

# ログを確認
docker-compose logs -f go-backend
```

### 3. 動作確認

```bash
# ヘルスチェック
curl http://localhost:8082/health

# PDFをアップロードして解析テスト
curl -X POST http://localhost:8082/api/v1/properties/parse \
  -F "file=@your_file.pdf"

# 物件一覧取得
curl http://localhost:8082/api/v1/properties | jq .
```

---

## 🌐 ローカル確認用URL一覧

| サービス | URL | 説明 |
|----------|-----|------|
| **フロントエンド** | http://localhost:3000 | 物件管理ダッシュボード |
| **Go Backend API** | http://localhost:8082 | REST API サーバー |
| **API ヘルスチェック** | http://localhost:8082/health | Go バックエンド死活確認 |
| **Supabase Kong** | http://localhost:8000 | Supabase APIゲートウェイ |
| **Supabase Studio** | http://localhost:3001 | DB管理ダッシュボード（テーブル閲覧・編集） |
| **PostgreSQL** | localhost:5433 | 直接DB接続 (user: postgres / pass: postgres) |

### Supabase Studio でDBを確認する方法
1. http://localhost:3001 を開く
2. Table Editor → `properties` テーブルを選択
3. 解析結果・通話記録などをGUIで確認可能

### PostgREST 直接クエリ例

```bash
# ANON_KEY
ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1sb2NhbCIsInJvbGUiOiJhbm9uIiwiZXhwIjoxOTgzODEyOTk2fQ.M5YxZEaHHJzS2YaxZ5KZokoZw7f4vGiOVu3_nsMln2c"

# 物件一覧 (新しい順)
curl "http://localhost:8000/rest/v1/properties?order=created_at.desc" \
  -H "apikey: $ANON_KEY" | jq .

# 件数確認
curl "http://localhost:8000/rest/v1/properties?select=count" \
  -H "apikey: $ANON_KEY" \
  -H "Prefer: count=exact" -I

# DBを直接確認 (psql)
docker exec zumen-supabase-db \
  psql -U postgres -d postgres \
  -c "SELECT name, room_number, price, status FROM public.properties ORDER BY created_at DESC LIMIT 10;"
```

---

## 📡 API エンドポイント

### Properties

| メソッド | パス | 説明 |
|---------|------|------|
| `POST` | `/api/v1/properties/parse` | PDF アップロード＆AI解析トリガー |
| `GET` | `/api/v1/properties` | 物件一覧取得 |
| `GET` | `/api/v1/properties/:id` | 物件詳細取得 |
| `PATCH` | `/api/v1/properties/:id` | 物件情報更新 |
| `DELETE` | `/api/v1/properties/:id` | 1件削除 |
| `DELETE` | `/api/v1/properties` | **全件削除** |
| `DELETE` | `/api/v1/properties/bulk` | **複数件削除** `{"ids":["uuid1","uuid2"]}` |
| `POST` | `/api/v1/properties/:id/parse` | AI解析を再実行 |
| `POST` | `/api/v1/properties/:id/call` | AI物確電話トリガー |
| `GET` | `/api/v1/properties/:id/calls` | 通話記録取得 |

---

## 📂 プロジェクト構造

```
Zumen-kun/
├── docker-compose.yml           # 全サービス統合起動
├── .env.example                 # 環境変数テンプレート
│
├── go-backend/                  # Go (Gin) メインバックエンド ★主役
│   ├── cmd/server/main.go       # エントリーポイント・ルーティング
│   ├── internal/
│   │   ├── api/handlers/        # REST APIハンドラー
│   │   │   ├── pdf.go           # PDF受信 → Gemini解析
│   │   │   ├── property.go      # CRUD + 全削除/複数削除
│   │   │   └── call.go          # 通話記録
│   │   ├── db/
│   │   │   ├── supabase.go      # PostgREST 直接HTTP呼び出し
│   │   │   └── models/          # Property / CallRecord 構造体
│   │   ├── worker/
│   │   │   ├── gemini_parser.go # Gemini 2.0 Flash PDF解析
│   │   │   ├── ai_config.go     # AI設定 (Gemini/OpenAI切替)
│   │   │   └── prompts.go       # 解析プロンプト
│   │   └── config/config.go     # 環境変数設定
│   └── Dockerfile
│
├── python-worker/               # Python AI Worker (将来の拡張用・現在未使用)
│   └── src/
│       ├── voice_agent/         # Retell AI 統合 (将来実装)
│       └── parsers/             # OpenAI Vision代替 (将来実装)
│
├── frontend/                    # Next.js フロントエンド
│   ├── app/
│   │   ├── page.tsx             # ダッシュボード (全削除ボタン)
│   │   └── layout.tsx
│   ├── components/
│   │   ├── PropertyTable.tsx    # テーブル (チェックボックス多選削除)
│   │   ├── PDFUploader.tsx      # ドラッグ&ドロップアップロード
│   │   └── StatsCards.tsx       # 統計カード
│   └── lib/
│       ├── api-client.ts        # Go Backend APIクライアント
│       └── supabase.ts          # Supabase Realtime購読
│
├── supabase/
│   ├── init/                    # DB初期化SQL (コンテナ起動時に自動実行)
│   │   ├── 00_roles.sql         # ロール設定
│   │   └── 01_schema.sql        # テーブル定義
│   ├── kong.yml                 # Supabase Kongルーティング設定
│   └── pg_hba.conf              # PostgreSQL認証設定
│
└── docs/
    ├── schema.sql               # DBスキーマ (参照用)
    └── API.md                   # API仕様書
```

---

## 🔄 処理フロー

### PDF解析フロー（現在の実装）

```
1. ユーザーが PDF をドラッグ＆ドロップ
2. Go Backend → PDF を /uploads/ に一時保存
3. Go Backend → Supabase (PostgREST) に status=parsing でレコード作成
4. Go Backend → Gemini 2.0 Flash Vision API に PDF を直接送信
5. Gemini → 物件名・号室・電話番号・賃料・AD を JSON で返す
6. Go Backend → Supabase (PostgREST) に解析結果を PATCH
7. フロントエンド → Supabase Realtime で自動反映（5秒ポーリング含む）
```

### AI物確電話フロー（Retell AI統合後）

```
1. ユーザーが「AI物確」ボタンをクリック
2. Go Backend → status=calling に更新
3. Retell AI → 管理会社に電話
4. Retell AI → Webhook で通話結果送信
5. Go Backend → call_records テーブルに保存
6. Go Backend → properties.foreigner_ok を更新
```

---

## ⚠️ APIキー未設定時の動作

| サービス | 未設定時の動作 |
|----------|----------------|
| Gemini API | モックデータを返す（開発用サンプル物件情報） |
| Retell AI | 模擬通話を実行（事前定義のモック会話） |
| SUPABASE_URL/KEY | 起動時に警告ログを出力（DB操作をスキップ） |

---

## 🛠️ 技術スタック

| 層 | 技術 |
|----|------|
| Frontend | Next.js 15 + TypeScript + TailwindCSS + shadcn/ui |
| Backend | Go 1.22 + Gin Framework |
| DB | Supabase ローカル (PostgreSQL 15 + Realtime + Kong) |
| AI Vision | **Gemini 2.0 Flash** (PDF直接解析) |
| 音声AI | Retell AI (将来統合) |
| Container | Docker + Docker Compose |

---

## 🐛 トラブルシューティング

### DB に保存されない場合
```bash
# Go バックエンドのログを確認
docker logs zumen-go-backend 2>&1 | grep -E "(POST|📥|❌)" | tail -20

# Supabase Kong の疎通確認
docker exec zumen-go-backend wget -qO- http://supabase-kong:8000/rest/v1/properties \
  -H "apikey: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1sb2NhbCIsInJvbGUiOiJhbm9uIiwiZXhwIjoxOTgzODEyOTk2fQ.M5YxZEaHHJzS2YaxZ5KZokoZw7f4vGiOVu3_nsMln2c"
```

### Supabase Studio が開かない場合
```bash
# Studio コンテナのログ確認
docker logs zumen-supabase-studio 2>&1 | tail -20

# 全コンテナの状態確認
docker-compose ps
```

### DB を完全にリセットする場合
```bash
docker-compose down -v   # ボリュームごと削除
docker-compose up -d     # 再起動（スキーマ自動再作成）
```

---

## 📖 ドキュメント

- [API仕様書](docs/API.md)
- [DBスキーマ](docs/schema.sql)
