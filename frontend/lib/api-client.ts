/**
 * API client for Go Backend communication
 */

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface Property {
  id: string;
  name: string;
  room_number: string;
  phone_number: string;
  price: number;
  ad_rate: string;
  status: "pending" | "parsing" | "calling" | "completed" | "parse_failed" | "call_failed";
  foreigner_ok: boolean | null;
  pdf_url: string;
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
  const res = await fetch(`${API_BASE_URL}/api/v1/properties`, {
    cache: "no-store",
  });

  if (!res.ok) {
    throw new Error(`Failed to fetch properties: ${res.statusText}`);
  }

  return res.json();
}

export async function getProperty(id: string): Promise<Property> {
  const res = await fetch(`${API_BASE_URL}/api/v1/properties/${id}`, {
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

  const res = await fetch(`${API_BASE_URL}/api/v1/properties/parse`, {
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
    `${API_BASE_URL}/api/v1/properties/${propertyId}/parse`,
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
    `${API_BASE_URL}/api/v1/properties/${propertyId}/call`,
    {
      method: "POST",
    }
  );

  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.error || "Failed to trigger call");
  }
}

export async function deleteProperty(propertyId: string): Promise<void> {
  const res = await fetch(
    `${API_BASE_URL}/api/v1/properties/${propertyId}`,
    {
      method: "DELETE",
    }
  );

  if (!res.ok) {
    throw new Error("Failed to delete property");
  }
}

// ─────────────────────────────────────────────────────────
// Call Records API
// ─────────────────────────────────────────────────────────

export async function getCallRecords(propertyId: string): Promise<CallRecord[]> {
  const res = await fetch(
    `${API_BASE_URL}/api/v1/properties/${propertyId}/calls`,
    {
      cache: "no-store",
    }
  );

  if (!res.ok) {
    throw new Error(`Failed to fetch call records: ${res.statusText}`);
  }

  return res.json();
}
