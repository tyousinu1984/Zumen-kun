-- ============================================================
-- 図面くん (Zumen-kun) Database Schema
-- Supabase (PostgreSQL) で実行してください
-- ============================================================

-- 拡張機能の有効化
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- 物件情報テーブル (Go 主制御)
-- ============================================================
CREATE TABLE IF NOT EXISTS properties (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name           TEXT,                   -- 物件名（例: 鈴木ハイム）
  room_number    TEXT,                   -- 号室（例: 102）
  phone_number   TEXT,                   -- 管理会社の電話番号
  price          NUMERIC,                -- 賃料（万円）
  ad_rate        TEXT,                   -- AD（広告料）
  status         TEXT DEFAULT 'pending', -- pending/parsing/calling/completed/parse_failed/call_failed
  foreigner_ok   BOOLEAN,                -- 外国人可否（AI電話確認後）
  pdf_url        TEXT,                   -- Supabase Storage の PDF URL
  created_at     TIMESTAMPTZ DEFAULT NOW(),
  updated_at     TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- 通話記録テーブル (Python 書き込み, Go 読み取り)
-- ============================================================
CREATE TABLE IF NOT EXISTS call_records (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  property_id      UUID REFERENCES properties(id) ON DELETE CASCADE,
  transcript       TEXT,             -- 通話全文
  ai_summary       TEXT,             -- AI要約（空室状況・外国人可否など）
  recording_url    TEXT,             -- 録音ファイル URL (Retell AI)
  duration_seconds INTEGER,          -- 通話時間（秒）
  status           TEXT,             -- success/failed/timeout
  created_at       TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- インデックス（パフォーマンス最適化）
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_properties_status
  ON properties(status);

CREATE INDEX IF NOT EXISTS idx_properties_created_at
  ON properties(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_call_records_property_id
  ON call_records(property_id);

-- ============================================================
-- updated_at 自動更新トリガー
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_properties_updated_at
  BEFORE UPDATE ON properties
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- Row Level Security (RLS)
-- ============================================================
ALTER TABLE properties ENABLE ROW LEVEL SECURITY;
ALTER TABLE call_records ENABLE ROW LEVEL SECURITY;

-- Service role（Go Backend / Python Worker）は全操作可能
CREATE POLICY "service_role_all_properties"
  ON properties FOR ALL
  USING (true)
  WITH CHECK (true);

CREATE POLICY "service_role_all_call_records"
  ON call_records FOR ALL
  USING (true)
  WITH CHECK (true);

-- ============================================================
-- Realtime 有効化（フロントエンドのリアルタイム更新用）
-- ============================================================
ALTER PUBLICATION supabase_realtime ADD TABLE properties;
ALTER PUBLICATION supabase_realtime ADD TABLE call_records;

-- ============================================================
-- Supabase Storage バケット作成（SQLエディターでは実行不可）
-- Supabase Dashboard > Storage > New Bucket で作成:
--   Bucket name: property-pdfs
--   Public: false (プライベート)
-- ============================================================

-- ============================================================
-- サンプルデータ（開発・テスト用）
-- ============================================================
INSERT INTO properties (name, room_number, phone_number, price, ad_rate, status)
VALUES
  ('サンプルマンション', '101', '03-1234-5678', 8.5, '1ヶ月', 'completed'),
  ('テストアパート', '205', '06-9876-5432', 6.0, '2ヶ月', 'pending'),
  ('デモハイツ', '302', '03-5555-0000', 12.0, '0.5ヶ月', 'parsing')
ON CONFLICT DO NOTHING;
