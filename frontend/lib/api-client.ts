/**
 * API client for Go Backend communication
 *
 * ブラウザからは Next.js のリライトプロキシ経由でアクセス:
 *   /api/backend/* → Go Backend /api/v1/*
 *
 * これによりビルド時に URL を焼き込む必要がなくなり、
 * ローカル/staging/本番すべて同一イメージで動作可能。
 */

// リライト: /api/backend/:path* → Go Backend /api/v1/:path*
// 例: /api/backend/properties → http://go-backend:8080/api/v1/properties
const API_BASE_URL = "/api/backend";

export interface Property {
  id: string;
  company_id: string;
  name: string;
  room_number: string;
  phone_number: string;
  price: number;
  ad_rate: string;
  status: "pending" | "parsing" | "calling" | "completed" | "parse_failed" | "call_failed";
  foreigner_ok: boolean | null;
  pdf_url: string;
  source_pdf_id: string; // 元PDFのpropertyID
  page_number: number;   // このレコードが何ページ目か
  created_at: string;
  updated_at: string;
}

export interface CallRecord {
  id: string;
  property_id: string;
  transcript: string;
  ai_summary: string;
  recording_url: string;
  duration_seconds: number;
  status: string;
  created_at: string;
}

export interface ParseResponse {
  property_id: string;
  pdf_url: string;
  status: string;
  message: string;
}

// ─────────────────────────────────────────────────────────
// Properties API
// ─────────────────────────────────────────────────────────

export async function listProperties(): Promise<Property[]> {
  const res = await fetch(`${API_BASE_URL}/properties`, {
    cache: "no-store",
  });

  if (!res.ok) {
    throw new Error(`Failed to fetch properties: ${res.statusText}`);
  }

  return res.json();
}

export async function getProperty(id: string): Promise<Property> {
  const res = await fetch(`${API_BASE_URL}/properties/${id}`, {
    cache: "no-store",
  });

  if (!res.ok) {
    throw new Error(`Failed to fetch property: ${res.statusText}`);
  }

  return res.json();
}

export async function uploadAndParsePDF(file: File): Promise<ParseResponse> {
  const formData = new FormData();
  formData.append("file", file);

  const res = await fetch(`${API_BASE_URL}/properties/parse`, {
    method: "POST",
    body: formData,
  });

  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.error || "Failed to upload PDF");
  }

  return res.json();
}

export async function triggerParse(propertyId: string): Promise<void> {
  const res = await fetch(
    `${API_BASE_URL}/properties/${propertyId}/parse`,
    {
      method: "POST",
    }
  );

  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.error || "Failed to trigger parse");
  }
}

export async function triggerCall(propertyId: string): Promise<void> {
  const res = await fetch(
    `${API_BASE_URL}/properties/${propertyId}/call`,
    { method: "POST" }
  );
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.error || "Failed to trigger call");
  }
}

export async function cancelCall(propertyId: string): Promise<void> {
  const res = await fetch(
    `${API_BASE_URL}/properties/${propertyId}/cancel-call`,
    { method: "POST" }
  );
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.error || "Failed to cancel call");
  }
}

export function getPdfUrl(propertyId: string): string {
  return `${API_BASE_URL}/properties/${propertyId}/pdf`;
}

export async function deleteProperty(propertyId: string): Promise<void> {
  const res = await fetch(
    `${API_BASE_URL}/properties/${propertyId}`,
    { method: "DELETE" }
  );
  if (!res.ok) throw new Error("Failed to delete property");
}

export async function deleteAllProperties(): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/properties`, {
    method: "DELETE",
  });
  if (!res.ok) throw new Error("Failed to delete all properties");
}

export async function deleteBulkProperties(ids: string[]): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/properties/bulk`, {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ids }),
  });
  if (!res.ok) throw new Error("Failed to bulk delete properties");
}

// ─────────────────────────────────────────────────────────
// Call Records API
// ─────────────────────────────────────────────────────────

export async function getCallRecords(propertyId: string): Promise<CallRecord[]> {
  const res = await fetch(
    `${API_BASE_URL}/properties/${propertyId}/calls`,
    {
      cache: "no-store",
    }
  );

  if (!res.ok) {
    throw new Error(`Failed to fetch call records: ${res.statusText}`);
  }

  return res.json();
}
