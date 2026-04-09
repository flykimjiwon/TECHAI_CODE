# Apache ECharts

## 기본 구조

```js
import * as echarts from 'echarts'

const chart = echarts.init(document.getElementById('main'))

const option = {
  title: { text: 'Sales 2024' },
  tooltip: { trigger: 'axis' },
  legend: { data: ['Revenue', 'Profit'] },
  xAxis: { type: 'category', data: ['Jan', 'Feb', 'Mar'] },
  yAxis: { type: 'value' },
  series: [
    { name: 'Revenue', type: 'line', data: [120, 200, 150] },
    { name: 'Profit', type: 'bar', data: [50, 80, 60] }
  ]
}

chart.setOption(option)
```

## Series Types

```js
// Line chart
{ type: 'line', smooth: true, data: [...] }

// Bar chart
{ type: 'bar', stack: 'total', data: [...] }

// Pie chart
{ type: 'pie', radius: '50%', data: [
  { value: 335, name: 'A' },
  { value: 234, name: 'B' }
]}

// Scatter
{ type: 'scatter', data: [[10, 20], [30, 40]] }

// Radar
{ type: 'radar', data: [{ value: [4, 3, 5, 2] }] }

// Heatmap
{ type: 'heatmap', data: [[0, 0, 5], [0, 1, 3]] }
```

## Responsive

```js
window.addEventListener('resize', () => chart.resize())
```

## Large Dataset

```js
xAxis: {
  type: 'category',
  data: largeArray
}
dataZoom: [
  { type: 'slider', start: 0, end: 10 },
  { type: 'inside' }
]
series: [
  { type: 'line', large: true, largeThreshold: 2000, data: [...] }
]
```

## Theme

```js
echarts.registerTheme('dark', { ... })
const chart = echarts.init(dom, 'dark')
```

## React Integration

```jsx
import ReactECharts from 'echarts-for-react'

<ReactECharts option={option} style={{ height: 400 }} />
```

**패턴**: option 객체로 모든 설정. series.type으로 차트 종류 결정. dataZoom으로 대용량 데이터 핸들링. resize() 필수.
