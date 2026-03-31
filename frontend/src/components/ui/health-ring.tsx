// Circular health indicator extracted from DashboardPage

export interface HealthRingProps {
  /** Score 0-100 */
  value: number
  color: string
  size?: number
  strokeWidth?: number
}

export function HealthRing({ value, color, size = 80, strokeWidth = 6 }: HealthRingProps) {
  const radius = (size - strokeWidth) / 2
  const circumference = 2 * Math.PI * radius
  const offset = circumference - (value / 100) * circumference
  return (
    <svg width={size} height={size} style={{ transform: 'rotate(-90deg)' }}>
      <circle
        cx={size / 2} cy={size / 2} r={radius}
        fill="none" stroke="#f1f5f9" strokeWidth={strokeWidth}
      />
      <circle
        cx={size / 2} cy={size / 2} r={radius}
        fill="none" stroke={color} strokeWidth={strokeWidth}
        strokeDasharray={circumference} strokeDashoffset={offset}
        strokeLinecap="round"
        style={{ transition: 'stroke-dashoffset 1s ease-out' }}
      />
    </svg>
  )
}
