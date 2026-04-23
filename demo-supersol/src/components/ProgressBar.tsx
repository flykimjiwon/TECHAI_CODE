"use client";

interface ProgressBarProps {
  current: number;
  target: number;
  label?: string;
  color?: string;
}

export default function ProgressBar({
  current,
  target,
  label,
  color = "bg-blue-600",
}: ProgressBarProps) {
  const pct = Math.min(Math.round((current / target) * 100), 100);

  return (
    <div className="space-y-1.5">
      {label && (
        <div className="flex justify-between text-xs">
          <span className="text-gray-500">{label}</span>
          <span className="font-bold text-blue-600">{pct}%</span>
        </div>
      )}
      <div className="w-full h-2 bg-gray-100 rounded-full overflow-hidden">
        <div
          className={`h-full rounded-full transition-all duration-300 ${color}`}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}
