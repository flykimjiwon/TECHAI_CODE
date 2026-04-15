# Framer Motion Reference

## Installation
```bash
npm install framer-motion
```

## motion.* Components
```tsx
import { motion } from "framer-motion"

<motion.div animate={{ opacity: 1, x: 0 }} />
<motion.span animate={{ scale: 1.2 }} />
<motion.button whileTap={{ scale: 0.95 }} />
<motion.img animate={{ rotate: 360 }} />
```

## initial / animate / exit
```tsx
<motion.div
  initial={{ opacity: 0, y: -20 }}
  animate={{ opacity: 1, y: 0 }}
  exit={{ opacity: 0, y: 20 }}
  transition={{ duration: 0.3, ease: "easeOut" }}
/>

{/* Transition options */}
transition={{
  duration: 0.5,
  delay: 0.1,
  ease: "easeInOut",        // "linear" | "easeIn" | "easeOut" | "easeInOut"
  type: "spring",           // spring physics
  stiffness: 300,
  damping: 20,
  repeat: Infinity,
  repeatType: "reverse",    // "loop" | "reverse" | "mirror"
}}
```

## Variants (staggerChildren)
```tsx
const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,    // delay between each child
      delayChildren: 0.2,      // delay before first child
    },
  },
}

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
}

<motion.ul variants={containerVariants} initial="hidden" animate="visible">
  {items.map((item) => (
    <motion.li key={item.id} variants={itemVariants}>
      {item.label}
    </motion.li>
  ))}
</motion.ul>
```

## Gesture Props
```tsx
<motion.button
  whileHover={{ scale: 1.05, backgroundColor: "#3b82f6" }}
  whileTap={{ scale: 0.95 }}
  whileFocus={{ outline: "2px solid #3b82f6" }}
  whileDrag={{ scale: 1.1, rotate: 5 }}
  drag
  dragConstraints={{ left: -100, right: 100, top: -50, bottom: 50 }}
  onHoverStart={() => console.log("hover")}
  onTap={() => console.log("tap")}
/>
```

## AnimatePresence
```tsx
import { AnimatePresence, motion } from "framer-motion"

{/* Must have key prop on child */}
<AnimatePresence mode="wait">
  {isVisible && (
    <motion.div
      key="modal"
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      exit={{ opacity: 0, scale: 0.9 }}
    />
  )}
</AnimatePresence>

{/* mode options: "sync" (default) | "wait" | "popLayout" */}
<AnimatePresence mode="wait" initial={false}>
  <motion.div key={route} ...>  {/* Page transitions */}
</AnimatePresence>
```

## Layout Animations
```tsx
{/* Animate layout changes automatically */}
<motion.div layout>
  {isExpanded ? <FullContent /> : <Summary />}
</motion.div>

{/* Shared element transition */}
<motion.div layoutId="card-image">
  <img src={thumbnail} />
</motion.div>
{/* In another location: */}
<motion.div layoutId="card-image">
  <img src={fullImage} />
</motion.div>

{/* Layout types */}
<motion.div layout="position">   {/* Only position */}
<motion.div layout="size">       {/* Only size */}
<motion.div layout="preserve-aspect">
```

## useMotionValue & useSpring
```tsx
import { useMotionValue, useSpring, useTransform, motion } from "framer-motion"

function Component() {
  const x = useMotionValue(0)
  const springX = useSpring(x, { stiffness: 300, damping: 20 })
  const rotate = useTransform(x, [-200, 200], [-30, 30])
  const opacity = useTransform(x, [-100, 0, 100], [0, 1, 0])

  return (
    <motion.div
      style={{ x: springX, rotate, opacity }}
      drag="x"
      dragConstraints={{ left: -200, right: 200 }}
      onDrag={(_, info) => x.set(info.offset.x)}
    />
  )
}
```

## Scroll Animations
```tsx
import { useScroll, useTransform, useSpring, motion } from "framer-motion"

function ParallaxSection() {
  const ref = useRef(null)
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start end", "end start"],  // [enter, exit] viewport edges
  })

  const y = useTransform(scrollYProgress, [0, 1], ["-20%", "20%"])
  const opacity = useTransform(scrollYProgress, [0, 0.3, 0.7, 1], [0, 1, 1, 0])

  return (
    <section ref={ref}>
      <motion.div style={{ y, opacity }}>Parallax content</motion.div>
    </section>
  )
}

{/* whileInView — simpler scroll trigger */}
<motion.div
  initial={{ opacity: 0, y: 40 }}
  whileInView={{ opacity: 1, y: 0 }}
  viewport={{ once: true, margin: "-100px" }}
  transition={{ duration: 0.5 }}
/>
```

## MotionConfig
```tsx
import { MotionConfig } from "framer-motion"

{/* Apply transition defaults to all children */}
<MotionConfig transition={{ duration: 0.2, ease: "easeOut" }}>
  <motion.div animate={{ opacity: 1 }} />  {/* Uses config transition */}
  <motion.button whileTap={{ scale: 0.95 }} />
</MotionConfig>

{/* Reduce motion for accessibility */}
<MotionConfig reducedMotion="user">
  {/* Respects prefers-reduced-motion media query */}
</MotionConfig>
```

## 자주 쓰는 패턴

### Fade In List
```tsx
const list = { hidden: { opacity: 0 }, show: { opacity: 1, transition: { staggerChildren: 0.08 } } }
const item = { hidden: { opacity: 0, y: 16 }, show: { opacity: 1, y: 0 } }

<motion.div variants={list} initial="hidden" animate="show">
  {data.map((d) => <motion.div key={d.id} variants={item}>{d.name}</motion.div>)}
</motion.div>
```

### Modal with AnimatePresence
```tsx
<AnimatePresence>
  {open && (
    <motion.div
      key="overlay"
      className="fixed inset-0 bg-black/50"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      onClick={onClose}
    >
      <motion.div
        className="fixed inset-x-4 top-1/2 -translate-y-1/2 bg-white rounded-xl p-6"
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.9, opacity: 0 }}
        onClick={(e) => e.stopPropagation()}
      >
        {children}
      </motion.div>
    </motion.div>
  )}
</AnimatePresence>
```
