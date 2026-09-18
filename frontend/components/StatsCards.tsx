"use client";

import type { Property } from "@/lib/api-client";

interface StatsCardsProps {
  properties: Property[];
}

export function StatsCards({ properties }: StatsCardsProps) {
  const total = properties.length;
  const completed = properties.filter((p) => p.status === "completed").length;
  const parsing = properties.filter((p) => p.status === "parsing").length;
  const calling = properties.filter((p) => p.status === "calling").length;
  const foreignerOk = properties.filter((p) => p.foreigner_ok === true).length;
  const foreignerNg = properties.filter((p) => p.foreigner_ok === false).length;

  const stats = [
    {
      label: "総物件数",
      value: total,
      icon: "🏢",
      color: "bg-blue-50 text-blue-700",
      border: "border-blue-100",
    },
    {
      label: "解析完了",
      value: completed,
      icon: "✅",
      color: "bg-green-50 text-green-700",
      border: "border-green-100",
    },
    {
      label: "解析中",
      value: parsing,
      icon: "⚙️",
      color: "bg-blue-50 text-blue-600",
      border: "border-blue-100",
      pulse: parsing > 0,
    },
    {
      label: "物確中",
      value: calling,
      icon: "📞",
      color: "bg-yellow-50 text-yellow-700",
      border: "border-yellow-100",
      pulse: calling > 0,
    },
    {
      label: "外国人可",
      value: foreignerOk,
      icon: "🌍",
      color: "bg-emerald-50 text-emerald-700",
      border: "border-emerald-100",
    },
    {
      label: "外国人不可",
      value: foreignerNg,
      icon: "🚫",
      color: "bg-red-50 text-red-700",
      border: "border-red-100",
    },
  ];

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
      {stats.map((stat) => (
        <div
          key={stat.label}
          className={`bg-white rounded-xl border ${stat.border} p-4 text-center shadow-sm`}
        >
          <div className={`inline-flex items-center justify-center w-10 h-10 rounded-full ${stat.color} mb-2`}>
            <span className="text-lg">{stat.icon}</span>
          </div>
          <div className="flex items-center justify-center gap-1">
            <span className="text-2xl font-bold text-gray-900">{stat.value}</span>
            {stat.pulse && (
              <span className="w-2 h-2 bg-blue-500 rounded-full animate-pulse" />
            )}
          </div>
          <p className="text-xs text-gray-500 mt-0.5">{stat.label}</p>
        </div>
      ))}
    </div>
  );
}
