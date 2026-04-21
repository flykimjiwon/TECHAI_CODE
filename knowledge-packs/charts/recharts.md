# Recharts Reference

## Installation
```bash
npm install recharts
```

## ResponsiveContainer Pattern
```tsx
import { ResponsiveContainer } from 'recharts'

<ResponsiveContainer width="100%" height={400}>
  {/* Chart component */}
</ResponsiveContainer>
```

## BarChart
```tsx
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts'

const data = [
  { name: 'Jan', value: 400 },
  { name: 'Feb', value: 300 },
]

<BarChart width={600} height={300} data={data}>
  <CartesianGrid strokeDasharray="3 3" />
  <XAxis dataKey="name" />
  <YAxis />
  <Tooltip />
  <Legend />
  <Bar dataKey="value" fill="#8884d8" />
</BarChart>
```

## LineChart
```tsx
import { LineChart, Line, XAxis, YAxis, Tooltip } from 'recharts'

<LineChart width={600} height={300} data={data}>
  <XAxis dataKey="name" />
  <YAxis />
  <Tooltip />
  <Line type="monotone" dataKey="value" stroke="#8884d8" strokeWidth={2} />
</LineChart>
```

## PieChart
```tsx
import { PieChart, Pie, Cell, Tooltip, Legend } from 'recharts'

const data = [
  { name: 'Group A', value: 400 },
  { name: 'Group B', value: 300 },
]

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042']

<PieChart width={400} height={400}>
  <Pie data={data} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={80} label>
    {data.map((entry, index) => (
      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
    ))}
  </Pie>
  <Tooltip />
  <Legend />
</PieChart>
```

## AreaChart
```tsx
import { AreaChart, Area, XAxis, YAxis, Tooltip } from 'recharts'

<AreaChart width={600} height={300} data={data}>
  <XAxis dataKey="name" />
  <YAxis />
  <Tooltip />
  <Area type="monotone" dataKey="value" stroke="#8884d8" fill="#8884d8" />
</AreaChart>
```

## ComposedChart
```tsx
import { ComposedChart, Line, Bar, Area, XAxis, YAxis, Tooltip } from 'recharts'

<ComposedChart width={600} height={300} data={data}>
  <XAxis dataKey="name" />
  <YAxis />
  <Tooltip />
  <Bar dataKey="value1" fill="#8884d8" />
  <Line type="monotone" dataKey="value2" stroke="#ff7300" />
</ComposedChart>
```

## Custom Tooltip
```tsx
const CustomTooltip = ({ active, payload, label }: any) => {
  if (!active || !payload) return null

  return (
    <div className="bg-white p-4 border rounded shadow">
      <p className="font-bold">{label}</p>
      {payload.map((entry: any, index: number) => (
        <p key={index} style={{ color: entry.color }}>
          {entry.name}: {entry.value}
        </p>
      ))}
    </div>
  )
}

<BarChart data={data}>
  <Tooltip content={<CustomTooltip />} />
</BarChart>
```

## 자주 쓰는 데이터 형식
```tsx
// 단일 시리즈
const data = [
  { name: 'Page A', value: 400 },
  { name: 'Page B', value: 300 },
]

// 다중 시리즈
const data = [
  { name: 'Jan', sales: 4000, revenue: 2400 },
  { name: 'Feb', sales: 3000, revenue: 1398 },
]
```
