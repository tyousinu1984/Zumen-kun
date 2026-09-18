/**
 * Next.js Route Handler によるリバースプロキシ
 *
 * ブラウザ: /api/backend/* → このハンドラー → Go Backend /api/v1/*
 *
 * rewrites() と違い、Route Handler は真のランタイム実行なので
 * API_URL 環境変数がデプロイ時に正しく解決される。
 *
 * 環境別 API_URL:
 *   ローカル Docker: API_URL=http://go-backend:8080
 *   staging:        API_URL=https://api.staging.zumen-kun.com
 *   本番:           API_URL=https://api.zumen-kun.com
 */

import { NextRequest, NextResponse } from "next/server";

const API_URL = process.env.API_URL || "http://localhost:8082";

async function proxyRequest(request: NextRequest, params: { path: string[] }) {
  const path = params.path.join("/");
  const url = new URL(request.url);
  const targetUrl = `${API_URL}/api/v1/${path}${url.search}`;

  // リクエストヘッダーをコピー（Hostヘッダーは除外）
  const headers = new Headers();
  request.headers.forEach((value, key) => {
    if (key.toLowerCase() !== "host") {
      headers.set(key, value);
    }
  });

  try {
    const response = await fetch(targetUrl, {
      method: request.method,
      headers,
      body:
        request.method !== "GET" && request.method !== "HEAD"
          ? request.body
          : undefined,
      // Node.js fetch ではストリームをそのまま渡す
      // @ts-expect-error duplex is required for streaming body
      duplex: "half",
    });

    // レスポンスヘッダーをコピー
    const responseHeaders = new Headers();
    response.headers.forEach((value, key) => {
      // セキュリティ上の理由でいくつかのヘッダーは除外
      if (!["transfer-encoding", "connection"].includes(key.toLowerCase())) {
        responseHeaders.set(key, value);
      }
    });

    return new NextResponse(response.body, {
      status: response.status,
      headers: responseHeaders,
    });
  } catch (error) {
    console.error(`Proxy error to ${targetUrl}:`, error);
    return NextResponse.json(
      { error: "Go Backend への接続に失敗しました", target: targetUrl },
      { status: 502 }
    );
  }
}

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  return proxyRequest(request, await params);
}

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  return proxyRequest(request, await params);
}

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  return proxyRequest(request, await params);
}

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  return proxyRequest(request, await params);
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  return proxyRequest(request, await params);
}
