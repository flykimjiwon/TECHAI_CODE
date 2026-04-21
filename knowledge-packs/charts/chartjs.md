# Chart.js + react-chartjs-2 Reference

## Installation
```bash
npm install chart.js react-chartjs-2
```

## Setup (Register Components)
```tsx
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  LineElement,
  PointElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  LineElement,
  PointElement,
  ArcElement,
  Title,
  Tooltip,
  Legend
)
```

## Bar Chart
```tsx
import { Bar } from 'react-chartjs-2'

const data = {
  labels: ['Jan', 'Feb', 'Mar', 'Apr'],
  datasets: [
    {
      label: 'Sales',
      data: [12, 19, 3, 5],
      backgroundColor: 'rgba(75, 192, 192, 0.2)',
      borderColor: 'rgba(75, 192, 192, 1)',
      borderWidth: 1,
    },
  ],
}

const options = {
  responsive: true,
  plugins: {
    legend: {
      position: 'top' as const,
    },
    title: {
      display: true,
      text: 'Monthly Sales',
    },
  },
}

<Bar data={data} options={options} />
```

## Line Chart
```tsx
import { Line } from 'react-chartjs-2'

const data = {
  labels: ['Jan', 'Feb', 'Mar'],
  datasets: [
    {
      label: 'Revenue',
      data: [65, 59, 80],
      borderColor: 'rgb(75, 192, 192)',
      backgroundColor: 'rgba(75, 192, 192, 0.2)',
      tension: 0.1, // 곡선
    },
  ],
}

<Line data={data} options={options} />
```

## Pie Chart
```tsx
import { Pie } from 'react-chartjs-2'

const data = {
  labels: ['Red', 'Blue', 'Yellow'],
  datasets: [
    {
      data: [300, 50, 100],
      backgroundColor: [
        'rgba(255, 99, 132, 0.2)',
        'rgba(54, 162, 235, 0.2)',
        'rgba(255, 206, 86, 0.2)',
      ],
      borderColor: [
        'rgba(255, 99, 132, 1)',
        'rgba(54, 162, 235, 1)',
        'rgba(255, 206, 86, 1)',
      ],
      borderWidth: 1,
    },
  ],
}

<Pie data={data} />
```

## Doughnut Chart
```tsx
import { Doughnut } from 'react-chartjs-2'

<Doughnut data={data} options={options} />
```

## Radar Chart
```tsx
import { Radar } from 'react-chartjs-2'
import { RadialLinearScale } from 'chart.js'

ChartJS.register(RadialLinearScale)

const data = {
  labels: ['Eating', 'Drinking', 'Sleeping', 'Designing', 'Coding'],
  datasets: [
    {
      label: 'My Skills',
      data: [65, 59, 90, 81, 56],
      backgroundColor: 'rgba(255, 99, 132, 0.2)',
      borderColor: 'rgba(255, 99, 132, 1)',
    },
  ],
}

<Radar data={data} />
```

## Options Structure
```tsx
const options = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: true,
      position: 'top' as const,
    },
    title: {
      display: true,
      text: 'Chart Title',
    },
  },
  scales: {
    y: {
      beginAtZero: true,
      ticks: {
        callback: (value: number) => `$${value}`,
      },
    },
  },
}
```

## 다중 데이터셋
```tsx
const data = {
  labels: ['Jan', 'Feb', 'Mar'],
  datasets: [
    {
      label: 'Dataset 1',
      data: [65, 59, 80],
      borderColor: 'rgb(255, 99, 132)',
      backgroundColor: 'rgba(255, 99, 132, 0.5)',
    },
    {
      label: 'Dataset 2',
      data: [28, 48, 40],
      borderColor: 'rgb(54, 162, 235)',
      backgroundColor: 'rgba(54, 162, 235, 0.5)',
    },
  ],
}
```
