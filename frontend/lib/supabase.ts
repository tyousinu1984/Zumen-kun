/**
 * Supabase client for Realtime subscriptions
 */

import { createClient } from "@supabase/supabase-js";

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL || "";
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY || "";

// Only create client if credentials are provided
export const supabase =
  supabaseUrl && supabaseAnonKey
    ? createClient(supabaseUrl, supabaseAnonKey)
    : null;

export type RealtimeCallback<T> = (payload: {
  new: T;
  old: T;
  eventType: "INSERT" | "UPDATE" | "DELETE";
}) => void;

/**
 * Subscribe to property table changes
 * Enables real-time updates when Python Worker updates property status
 */
export function subscribeToProperties(
  callback: RealtimeCallback<Record<string, unknown>>
) {
  if (!supabase) {
    console.warn(
      "Supabase not configured - realtime updates unavailable"
    );
    return () => {};
  }

  const channel = supabase
    .channel("properties-changes")
    .on(
      "postgres_changes",
      {
        event: "*",
        schema: "public",
        table: "properties",
      },
      (payload) => {
        callback({
          new: payload.new as Record<string, unknown>,
          old: payload.old as Record<string, unknown>,
          eventType: payload.eventType as "INSERT" | "UPDATE" | "DELETE",
        });
      }
    )
    .subscribe();

  // Return unsubscribe function
  return () => {
    supabase?.removeChannel(channel);
  };
}

/**
 * Subscribe to call_records table changes
 */
export function subscribeToCallRecords(
  propertyId: string,
  callback: RealtimeCallback<Record<string, unknown>>
) {
  if (!supabase) {
    return () => {};
  }

  const channel = supabase
    .channel(`call-records-${propertyId}`)
    .on(
      "postgres_changes",
      {
        event: "*",
        schema: "public",
        table: "call_records",
        filter: `property_id=eq.${propertyId}`,
      },
      (payload) => {
        callback({
          new: payload.new as Record<string, unknown>,
          old: payload.old as Record<string, unknown>,
          eventType: payload.eventType as "INSERT" | "UPDATE" | "DELETE",
        });
      }
    )
    .subscribe();

  return () => {
    supabase?.removeChannel(channel);
  };
}
