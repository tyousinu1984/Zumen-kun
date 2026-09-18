-- ============================================================
-- Supabase 内部ロール作成スクリプト
-- supabase/postgres イメージが通常自動作成するロールを手動で定義
-- 注意: このファイルは 01_schema.sql より先に実行される (00_ prefix)
-- ============================================================

-- supabase_admin ロール
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'supabase_admin') THEN
    CREATE ROLE supabase_admin WITH LOGIN BYPASSRLS CREATEROLE CREATEDB REPLICATION PASSWORD 'postgres';
  END IF;
END $$;

-- supabase_auth_admin ロール
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'supabase_auth_admin') THEN
    CREATE ROLE supabase_auth_admin WITH LOGIN NOINHERIT CREATEROLE PASSWORD 'postgres';
  END IF;
END $$;

-- supabase_storage_admin ロール
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'supabase_storage_admin') THEN
    CREATE ROLE supabase_storage_admin WITH LOGIN NOINHERIT CREATEROLE PASSWORD 'postgres';
  END IF;
END $$;

-- authenticator ロール (PostgREST 用)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticator') THEN
    CREATE ROLE authenticator WITH LOGIN NOINHERIT NOCREATEROLE NOCREATEDB NOSUPERUSER PASSWORD 'postgres';
  END IF;
END $$;

-- anon ロール (PostgREST 匿名アクセス用)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
    CREATE ROLE anon WITH NOLOGIN NOINHERIT;
  END IF;
END $$;

-- authenticated ロール (PostgREST 認証済みユーザー用)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN
    CREATE ROLE authenticated WITH NOLOGIN NOINHERIT;
  END IF;
END $$;

-- service_role ロール (バイパス RLS 用)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'service_role') THEN
    CREATE ROLE service_role WITH NOLOGIN NOINHERIT BYPASSRLS;
  END IF;
END $$;

-- dashboard_user ロール (Studio 用)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'dashboard_user') THEN
    CREATE ROLE dashboard_user WITH NOLOGIN NOINHERIT CREATEROLE CREATEDB;
  END IF;
END $$;

-- supabase_realtime_admin ロール
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'supabase_realtime_admin') THEN
    CREATE ROLE supabase_realtime_admin WITH LOGIN BYPASSRLS CREATEROLE PASSWORD 'postgres';
  END IF;
END $$;

-- ロール権限付与
GRANT anon TO authenticator;
GRANT authenticated TO authenticator;
GRANT service_role TO authenticator;
GRANT supabase_admin TO postgres;

-- スキーマ作成
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS storage;
CREATE SCHEMA IF NOT EXISTS realtime;
CREATE SCHEMA IF NOT EXISTS _realtime;
CREATE SCHEMA IF NOT EXISTS extensions;

-- スキーマオーナー設定
ALTER SCHEMA auth OWNER TO supabase_auth_admin;
ALTER SCHEMA storage OWNER TO supabase_storage_admin;
ALTER SCHEMA realtime OWNER TO supabase_admin;
ALTER SCHEMA _realtime OWNER TO supabase_admin;

-- DB への接続・作成権限（全ロールに付与）
GRANT CONNECT ON DATABASE postgres TO supabase_admin, supabase_auth_admin, supabase_storage_admin, authenticator, supabase_realtime_admin, anon, authenticated, service_role;
GRANT CREATE ON DATABASE postgres TO supabase_admin, supabase_auth_admin, supabase_storage_admin;
GRANT ALL PRIVILEGES ON DATABASE postgres TO supabase_admin;

-- public スキーマオーナーを postgres に設定（superuser）
ALTER SCHEMA public OWNER TO postgres;

-- 全ロールに public スキーマ権限
GRANT ALL ON SCHEMA public TO supabase_admin, supabase_auth_admin, supabase_storage_admin, authenticator, anon, authenticated, service_role;

-- auth / storage スキーマ権限
GRANT ALL ON SCHEMA auth TO supabase_auth_admin, supabase_admin;
GRANT ALL ON SCHEMA storage TO supabase_storage_admin, supabase_admin;

-- public スキーマへのアクセス権
GRANT USAGE ON SCHEMA public TO anon, authenticated, service_role;
GRANT ALL ON ALL TABLES IN SCHEMA public TO anon, authenticated, service_role;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO anon, authenticated, service_role;
GRANT ALL ON ALL ROUTINES IN SCHEMA public TO anon, authenticated, service_role;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO anon, authenticated, service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO anon, authenticated, service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON ROUTINES TO anon, authenticated, service_role;
