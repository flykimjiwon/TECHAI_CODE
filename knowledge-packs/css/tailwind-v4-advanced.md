# Tailwind CSS v4 Advanced Features

## @theme Directive (Design Tokens in CSS)
```css
/* app/globals.css */
@import "tailwindcss";

@theme {
  --color-brand: oklch(60% 0.2 250);
  --color-brand-dark: oklch(45% 0.2 250);
  --font-sans: "Geist", sans-serif;
  --radius-card: 1rem;
  --spacing-section: 5rem;
  --shadow-card: 0 4px 24px oklch(0% 0 0 / 10%);
}
```
Custom tokens are auto-available as utilities: `bg-brand`, `text-brand-dark`, `rounded-card`, `mt-section`.

## @source Directive (Explicit Content Sources)
```css
@import "tailwindcss";
@source "../node_modules/@my-ui/components/src";
@source "../../packages/shared/src/**/*.tsx";
```
Tells Tailwind to scan additional paths for class names.

## Dynamic Spacing Scale
v4 uses a continuous scale — any numeric value works:
```tsx
<div className="p-3.5 mt-13 gap-7 w-92">
<div className="top-[17px] left-[calc(50%-8px)]">    {/* arbitrary */}
```
No more "only multiples of 4" restriction. Full range: 0–96+ with 0.5 steps.

## Container Queries (Built-in)
No plugin needed in v4:
```tsx
{/* Define container */}
<div className="@container">
  {/* Respond to container width */}
  <div className="grid grid-cols-1 @sm:grid-cols-2 @lg:grid-cols-3">
  <p className="text-sm @md:text-base @xl:text-lg">
</div>

{/* Named containers */}
<div className="@container/sidebar">
  <div className="@lg/sidebar:hidden">Hidden when sidebar >= lg</div>
</div>
```
Breakpoints: `@xs` (320px), `@sm` (384px), `@md` (448px), `@lg` (512px), `@xl` (576px), `@2xl` (672px).

## 3D Transforms
```tsx
{/* Perspective */}
<div className="perspective-500">
  <div className="rotate-x-12 rotate-y-6 rotate-z-3">3D card</div>
</div>

{/* Transform style */}
<div className="transform-style-3d">
  <div className="translate-z-4 backface-hidden">Front</div>
  <div className="rotate-y-180 backface-hidden">Back</div>
</div>

{/* Scale 3D */}
<div className="scale-z-75 hover:scale-z-100 transition-transform">
```

## OKLCH Color System
```css
@theme {
  /* OKLCH: Lightness Chroma Hue */
  --color-primary: oklch(65% 0.22 250);     /* blue */
  --color-success: oklch(70% 0.18 145);     /* green */
  --color-danger:  oklch(60% 0.25 25);      /* red */
  --color-muted:   oklch(60% 0.05 250);     /* desaturated */
}
```
```tsx
{/* Arbitrary OKLCH in utilities */}
<div className="bg-[oklch(70%_0.15_200)] text-[oklch(30%_0.05_200)]">
```
OKLCH is perceptually uniform — equal chroma changes look equal across hues.

## Gradient Enhancements
```tsx
{/* Linear with angle */}
<div className="bg-linear-45 from-blue-500 to-purple-500">
<div className="bg-linear-[135deg] from-pink-400 via-red-400 to-orange-400">

{/* Conic gradient */}
<div className="bg-conic from-blue-500 to-purple-500">
<div className="bg-conic-[at_50%_40%] from-red-500 via-yellow-500 to-red-500">

{/* Radial gradient */}
<div className="bg-radial from-white to-blue-500">
<div className="bg-radial-[at_25%_25%] from-yellow-400 to-transparent">

{/* Interpolation mode */}
<div className="bg-linear-45 from-blue-500 to-red-500 bg-[in_oklch]">
```

## Performance: v3 vs v4

| Feature | v3 | v4 |
|---|---|---|
| Build tool | PostCSS (Node) | Lightning CSS (Rust) |
| Config file | `tailwind.config.js` | CSS `@theme {}` |
| Scan engine | Regex heuristics | Token-aware parser |
| Full build (large project) | ~800ms | ~100ms |
| Incremental rebuild | ~200ms | ~5ms |
| Arbitrary variants | Limited | Any CSS selector |
| Container queries | Plugin (`@tailwindcss/container-queries`) | Built-in |
| 3D transforms | Plugin or arbitrary | Built-in utilities |
| Color system | HSL/RGB | OKLCH-first |
| Dynamic values | Only whitelisted scales | Full continuous scale |

## Arbitrary Variants (v4 Enhancement)
```tsx
{/* Any CSS selector as variant */}
<div className="[&>li]:mb-2 [&:nth-child(3)]:bg-blue-100">
<div className="[@media(hover:hover)]:hover:underline">
<div className="[.dark_&]:text-white">     {/* ancestor has .dark */}
<div className="group-[.is-published]:block">
```

## 자주 쓰는 v4 패턴

### CSS Variables as Utilities
```tsx
{/* Use CSS vars directly */}
<div className="bg-(--brand-color) text-(--text-primary)">

{/* In @theme, then use */}
// @theme { --color-brand: oklch(60% 0.2 250); }
<div className="bg-brand hover:bg-brand/80">
```

### Starting Style (Entry Animation)
```tsx
<div className="@starting-style:opacity-0 transition-opacity duration-300">
  Fades in on first render
</div>
```
