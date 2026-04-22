# 図面くん (Zumen-kun) v4.0 🏠

> 日本の不動産中介業向けAI自動化プラットフォーム  
> PDF物件図面の自動解析 + AI音声物確電話

## 📋 概要

| 機能 | 説明 |
|------|------|
| 📄 PDF自動解析 | 物件図面PDFをGPT-4o Visionで解析し、物件名・賃料・ADを自動抽出 |
| 📞 AI物確電話 | Retell AIが管理会社に自動電話して空室状況・外国人可否を確認 |
| ⚡ リアルタイム更新 | Supabase Realtimeで処理状況をページ更新なしに表示 |
| 🌍 外国人可否管理 | AI電話の結果を自動判定してダッシュボードに表示 |

## 🏗️ アーキテクチャ

```
Next.js Frontend (3000)
        ↓ REST API
Go Backend - Gin (8080)
   ↓ Redis Queue    ↓ Supabase SDK
Redis (6379)    Supabase DB
        ↓
Python Worker - FastAPI (8001)
   - GPT-4o Vision (PDF解析)
   - Retell AI (音声電話)
```

## 🚀 クイックスタート

### 1. 環境変数の設定

```bash
cd Zumen-kun
cp .env.example .env
```

`.env` を編集して以下を設定:
- `SUPABASE_URL` / `SUPABASE_KEY` — Supabase プロジェクト設定
- `OPENAI_API_KEY` — OpenAI APIキー
- `RETELL_API_KEY` / `RETELL_AGENT_ID` — Retell AI設定

### 2. Supabase データベース初期化

Supabase Dashboard の SQL Editor で `docs/schema.sql` を実行してください。

### 3. Docker で起動

```bash
# 全サービスをビルド＆起動
docker-compose up --build

# バックグラウンドで起動
docker-compose up -d
```

### 4. 開発環境で個別起動

#### Go Backend
```bash
cd go-backend
go mod download
go run ./cmd/server
```

#### Python Worker
```bash
cd python-worker
pip install -r requirements.txt
uvicorn src.main:app --reload --port 8001
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

## 🌐 アクセスURL

| サービス | URL |
|----------|-----|
| フロントエンド | http://localhost:3000 |
| Go Backend API | http://localhost:8080 |
| Python Worker | http://localhost:8001 |
| Redis | localhost:6379 |

## 📡 ヘルスチェック

```bash
# Go Backend
curl http://localhost:8080/health

# Python Worker
curl http://localhost:8001/health
```

## 📂 プロジェクト構造

```
Zumen-kun/
├── docker-compose.yml       # 全サービス統合起動
├── .env.example             # 環境変数テンプレート
│
├── go-backend/              # Go (Gin) メインバックエンド
│   ├── cmd/server/main.go   # エントリーポイント
│   ├── internal/
│   │   ├── api/handlers/    # REST APIハンドラー
│   │   ├── db/              # Supabase クライアント
│   │   └── redis/           # Redisタスク Producer
│   └── Dockerfile
│
├── python-worker/           # Python AI エグゼキューター
│   ├── src/
│   │   ├── main.py          # FastAPI アプリ + Webhookエンドポイント
│   │   ├── worker.py        # Redis コンシューマー
│   │   ├── parsers/         # PDF Vision 解析
│   │   └── voice_agent/     # Retell AI 統合
│   └── Dockerfile
│
├── frontend/                # Next.js フロントエンド
│   ├── app/                 # App Router
│   ├── components/          # UIコンポーネント
│   └── lib/                 # API クライアント + Supabase
│
└── docs/
    ├── schema.sql           # Supabase DBスキーマ
    └── API.md               # API仕様書
```

## 🔄 処理フロー

### PDF解析フロー
```
1. ユーザーが PDF ドロップ
2. Go Backend → Supabase Storage に保存
3. Go Backend → Redis `task:parse_pdf` に push
4. Python Worker → Redis から task を pop
5. Python Worker → pdf2image で300DPI変換
6. Python Worker → GPT-4o Vision API 呼び出し
7. Python Worker → Supabase properties テーブル更新
8. フロントエンド → Realtime で自動反映
```

### AI物確電話フロー
```
1. ユーザーが「AI物確」ボタンをクリック
2. Go Backend → Redis `task:voice_call` に push
3. Python Worker → Retell AI でエージェント通話開始
4. Retell AI → 管理会社に電話
5. Retell AI → Webhook で通話結果送信
6. Python Worker → call_records テーブルに保存
7. Python Worker → properties.foreigner_ok を更新
8. フロントエンド → Realtime で自動反映
```

## ⚠️ APIキー未設定時の動作

| サービス | 未設定時の動作 |
|----------|----------------|
| OpenAI | モックデータを返す（開発用サンプル物件情報） |
| Retell AI | 模擬通話を実行（事前定義のモック会話） |
| Supabase | ログのみ出力（DB更新をスキップ） |
| Redis | 起動時に警告ログを出力 |

## 🛠️ 技術スタック

| 層 | 技術 |
|----|------|
| Frontend | Next.js 15 + TypeScript + TailwindCSS |
| Backend | Go 1.22 + Gin Framework |
| AI Worker | Python 3.12 + FastAPI |
| DB | Supabase (PostgreSQL + Storage + Realtime) |
| Queue | Redis 7 |
| PDF処理 | pdf2image + Poppler |
| AI Vision | GPT-4o Vision API |
| 音声AI | Retell AI |
| Container | Docker + Docker Compose |

## 📖 ドキュメント

- [API仕様書](docs/API.md)
- [DBスキーマ](docs/schema.sql)
