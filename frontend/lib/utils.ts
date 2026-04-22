import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatPrice(price: number | null | undefined): string {
  if (!price) return "---";
  return `${price}万円`;
}

export function formatDate(dateStr: string | null | undefined): string {
  if (!dateStr) return "---";
  return new Date(dateStr).toLocaleDateString("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function getStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    pending: "待機中",
    parsing: "解析中...",
    calling: "物確中...",
    completed: "完了",
    parse_failed: "解析失敗",
    call_failed: "物確失敗",
  };
  return labels[status] || status;
}

export function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    pending: "bg-gray-100 text-gray-700",
    parsing: "bg-blue-100 text-blue-700",
    calling: "bg-yellow-100 text-yellow-700",
    completed: "bg-green-100 text-green-700",
    parse_failed: "bg-red-100 text-red-700",
    call_failed: "bg-red-100 text-red-700",
  };
  return colors[status] || "bg-gray-100 text-gray-700";
}

export function getForeignerLabel(foreignerOk: boolean | null): string {
  if (foreignerOk === null || foreignerOk === undefined) return "未確認";
  return foreignerOk ? "✅ 外国人可" : "❌ 外国人不可";
}

export function getForeignerColor(foreignerOk: boolean | null): string {
  if (foreignerOk === null || foreignerOk === undefined)
    return "bg-gray-100 text-gray-500";
  return foreignerOk
    ? "bg-green-100 text-green-700"
    : "bg-red-100 text-red-700";
}
