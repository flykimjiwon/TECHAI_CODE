# Tailwind CSS v4 Reference

## Setup (CSS-first)
```css
/* app/globals.css */
@import "tailwindcss";
```

## Layout
```tsx
<div className="flex items-center justify-between">
<div className="grid grid-cols-3 gap-4">
<div className="container mx-auto px-4">
```

## Spacing
- Padding: `p-4` `px-6` `py-2` `pt-8` (scale: 0, 1, 2, 3, 4, 6, 8, 12, 16, 24, 32)
- Margin: `m-4` `mx-auto` `my-8` `-mt-4` (negative spacing)
- Gap: `gap-4` `gap-x-2` `gap-y-6`

## Typography
```tsx
<h1 className="text-4xl font-bold text-gray-900">
<p className="text-sm text-gray-600 leading-relaxed">
<span className="font-medium uppercase tracking-wide">
```

## Colors (50-950 shades)
```tsx
<div className="bg-blue-500 text-white border-blue-700">
<div className="bg-gray-100 text-gray-900 hover:bg-gray-200">
```
Common: gray, red, blue, green, yellow, purple, pink

## Borders & Shadows
```tsx
<div className="border border-gray-200 rounded-lg shadow-sm">
<div className="border-2 border-dashed rounded-xl shadow-lg">
```

## Responsive Breakpoints
```tsx
<div className="w-full md:w-1/2 lg:w-1/3">
  {/* sm: 640px, md: 768px, lg: 1024px, xl: 1280px, 2xl: 1536px */}
</div>
```

## Dark Mode
```tsx
<div className="bg-white dark:bg-gray-800 text-gray-900 dark:text-white">
```

## Transitions & Hover
```tsx
<button className="transition-colors hover:bg-blue-600 active:scale-95">
<div className="transition-all duration-300 hover:shadow-xl">
```

## 자주 쓰는 패턴

### Card
```tsx
<div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
  <h3 className="text-xl font-semibold mb-2">Title</h3>
  <p className="text-gray-600">Content</p>
</div>
```

### Button
```tsx
<button className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors">
  Click me
</button>
```

### Responsive Grid
```tsx
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
  {items.map(item => <Card key={item.id} {...item} />)}
</div>
```
