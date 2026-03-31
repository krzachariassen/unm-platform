// Gauge component extracted from CognitiveLoadView

const LEVEL_STYLE = {
  low:    { label: 'Low',    color: '#22c55e', bg: '#dcfce7' },
  medium: { label: 'Medium', color: '#f59e0b', bg: '#fef3c7' },
  high:   { label: 'High',   color: '#ef4444', bg: '#fee2e2' },
} as const

type LevelKey = keyof typeof LEVEL_STYLE

export interface SpeedometerDimension {
  abbr: string
  level: string
  value: number
}

export interface SpeedometerProps {
  level: string
  dimensions?: SpeedometerDimension[]
}

function arcPath(cx: number, cy: number, r: number, startDeg: number, endDeg: number): string {
  const s = (startDeg * Math.PI) / 180
  const e = (endDeg * Math.PI) / 180
  const x1 = cx + r * Math.cos(s)
  const y1 = cy + r * Math.sin(s)
  const x2 = cx + r * Math.cos(e)
  const y2 = cy + r * Math.sin(e)
  const large = endDeg - startDeg > 180 ? 1 : 0
  return `M ${x1} ${y1} A ${r} ${r} 0 ${large} 1 ${x2} ${y2}`
}

export function Speedometer({ level, dimensions }: SpeedometerProps) {
  const s = LEVEL_STYLE[level as LevelKey] ?? LEVEL_STYLE.low
  const needle = level === 'high' ? 0.85 : level === 'medium' ? 0.5 : 0.2

  const r = 44
  const cx = 52
  const cy = 52
  const startAngle = -210
  const sweep = 240
  const endAngle = startAngle + sweep
  const greenEnd = startAngle + sweep * 0.33
  const yellowEnd = startAngle + sweep * 0.66

  const needleAngle = startAngle + sweep * needle
  const needleRad = (needleAngle * Math.PI) / 180
  const nx = cx + (r - 14) * Math.cos(needleRad)
  const ny = cy + (r - 14) * Math.sin(needleRad)

  return (
    <div className="flex flex-col items-center">
      <svg width={104} height={72} viewBox="0 0 104 72">
        <path d={arcPath(cx, cy, r, startAngle, greenEnd)}  fill="none" stroke="#dcfce7" strokeWidth={7} strokeLinecap="round" />
        <path d={arcPath(cx, cy, r, greenEnd, yellowEnd)}   fill="none" stroke="#fef3c7" strokeWidth={7} strokeLinecap="round" />
        <path d={arcPath(cx, cy, r, yellowEnd, endAngle)}   fill="none" stroke="#fee2e2" strokeWidth={7} strokeLinecap="round" />
        <line
          x1={cx} y1={cy} x2={nx} y2={ny}
          stroke={s.color} strokeWidth={2.5} strokeLinecap="round"
        />
        <circle cx={cx} cy={cy} r={4} fill={s.color} />
      </svg>
      <span className="text-xs font-semibold mt-1" style={{ color: s.color }}>{s.label}</span>
      {dimensions && dimensions.length > 0 && (
        <div className="flex gap-1 mt-2 flex-wrap justify-center">
          {dimensions.map((d) => {
            const ds = LEVEL_STYLE[d.level as LevelKey] ?? LEVEL_STYLE.low
            return (
              <span
                key={d.abbr}
                className="text-[10px] font-medium px-1.5 py-0.5 rounded"
                style={{ background: ds.bg, color: ds.color }}
              >
                {d.abbr}
              </span>
            )
          })}
        </div>
      )}
    </div>
  )
}
