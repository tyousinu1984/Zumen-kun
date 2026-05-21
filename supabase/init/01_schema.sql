-- ============================================================
-- 図面くん (Zumen-kun) - ローカルSupabase DB 初期化
-- supabase/postgres イメージ用
-- 注意: このスクリプトは supabase_admin/service_role/anon ロール作成前に
--       実行されるため、ロール依存の処理は DO $$ で保護する
-- ============================================================

-- ============================================================
-- 会社テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS public.companies (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- ユーザーテーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS public.users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id    UUID REFERENCES public.companies(id) ON DELETE CASCADE,
  email         TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  role          TEXT NOT NULL DEFAULT 'user',   -- 'admin' or 'user'
  display_name  TEXT,
  is_active     BOOLEAN NOT NULL DEFAULT true,
  created_at    TIMESTAMPTZ DEFAULT NOW(),
  updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email      ON public.users(email);
CREATE INDEX IF NOT EXISTS idx_users_company_id ON public.users(company_id);

DROP TRIGGER IF EXISTS trigger_users_updated_at ON public.users;
CREATE TRIGGER trigger_users_updated_at
  BEFORE UPDATE ON public.users
  FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();

-- ============================================================
-- 物件情報テーブル
-- ============================================================

CREATE TABLE IF NOT EXISTS public.properties (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id     UUID REFERENCES public.companies(id) ON DELETE CASCADE,
  name           TEXT,
  room_number    TEXT,
  phone_number   TEXT,
  price          NUMERIC,
  ad_rate        TEXT,
  status         TEXT DEFAULT 'pending',
  foreigner_ok   BOOLEAN,
  pdf_url        TEXT,
  source_pdf_id  UUID,            -- 元PDFの最初のproperty ID（ページジャンプ用）
  page_number    INTEGER,          -- このレコードが何ページ目か
  created_at     TIMESTAMPTZ DEFAULT NOW(),
  updated_at     TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- 通話記録テーブル
-- ============================================================
CREATE TABLE IF NOT EXISTS public.call_records (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  property_id      UUID REFERENCES public.properties(id) ON DELETE CASCADE,
  transcript       TEXT,
  ai_summary       TEXT,
  recording_url    TEXT,
  duration_seconds INTEGER,
  status           TEXT,
  created_at       TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- インデックス
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_properties_status     ON public.properties(status);
CREATE INDEX IF NOT EXISTS idx_properties_created_at ON public.properties(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_call_records_prop_id  ON public.call_records(property_id);

-- ============================================================
-- updated_at トリガー
-- ============================================================
CREATE OR REPLACE FUNCTION public.update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_properties_updated_at ON public.properties;
CREATE TRIGGER trigger_properties_updated_at
  BEFORE UPDATE ON public.properties
  FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();

-- ============================================================
-- RLS (Row Level Security) + ポリシー
-- service_role/anon ロールが存在する場合のみ設定
-- ============================================================
DO $$
BEGIN
  -- RLS有効化
  ALTER TABLE public.properties  ENABLE ROW LEVEL SECURITY;
  ALTER TABLE public.call_records ENABLE ROW LEVEL SECURITY;

  -- service_role ポリシー
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'service_role') THEN
    DROP POLICY IF EXISTS "service_role_all_properties"   ON public.properties;
    DROP POLICY IF EXISTS "service_role_all_call_records" ON public.call_records;

    CREATE POLICY "service_role_all_properties"
      ON public.properties FOR ALL TO service_role USING (true) WITH CHECK (true);

    CREATE POLICY "service_role_all_call_records"
      ON public.call_records FOR ALL TO service_role USING (true) WITH CHECK (true);
  END IF;

  -- anon ポリシー（読み取り専用）
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
    DROP POLICY IF EXISTS "anon_read_properties" ON public.properties;

    CREATE POLICY "anon_read_properties"
      ON public.properties FOR SELECT TO anon USING (true);
  END IF;

EXCEPTION WHEN OTHERS THEN
  RAISE NOTICE 'RLS setup skipped: %', SQLERRM;
END $$;

-- ============================================================
-- Supabase Storage バケット（storage スキーマが存在する場合のみ）
-- ============================================================
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'storage' AND table_name = 'buckets'
  ) THEN
    INSERT INTO storage.buckets (id, name, public)
    VALUES ('property-pdfs', 'property-pdfs', false)
    ON CONFLICT (id) DO NOTHING;

    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'service_role') THEN
      IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'storage' AND tablename = 'objects'
        AND policyname = 'service_role_storage'
      ) THEN
        CREATE POLICY "service_role_storage" ON storage.objects
          FOR ALL TO service_role USING (true) WITH CHECK (true);
      END IF;
    END IF;
  END IF;
EXCEPTION WHEN OTHERS THEN
  RAISE NOTICE 'Storage setup skipped: %', SQLERRM;
END $$;

-- ============================================================
-- Realtime 有効化（publication が存在する場合のみ）
-- ============================================================
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_publication WHERE pubname = 'supabase_realtime') THEN
    IF NOT EXISTS (
      SELECT 1 FROM pg_publication_tables
      WHERE pubname = 'supabase_realtime' AND tablename = 'properties'
    ) THEN
      ALTER PUBLICATION supabase_realtime ADD TABLE public.properties;
    END IF;
    IF NOT EXISTS (
      SELECT 1 FROM pg_publication_tables
      WHERE pubname = 'supabase_realtime' AND tablename = 'call_records'
    ) THEN
      ALTER PUBLICATION supabase_realtime ADD TABLE public.call_records;
    END IF;
  END IF;
EXCEPTION WHEN OTHERS THEN
  RAISE NOTICE 'Realtime setup skipped: %', SQLERRM;
END $$;

