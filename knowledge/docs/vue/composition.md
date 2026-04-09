# Vue 3 Composition API

## Core APIs

### setup()
```js
export default {
  setup(props, context) {
    // composition logic
    return { /* template bindings */ }
  }
}
```

### Reactivity

**ref()**: primitive 반응형 (`.value`로 접근)
```js
import { ref } from 'vue'
const count = ref(0)
count.value++ // script에서
// {{ count }} template에서 자동 unwrap
```

**reactive()**: object 반응형
```js
const state = reactive({ user: 'Alice', age: 30 })
state.user = 'Bob' // direct mutation
```

**computed()**: derived state
```js
const doubled = computed(() => count.value * 2)
```

**watch()**: explicit dependency
```js
watch(count, (newVal, oldVal) => {
  console.log(`${oldVal} → ${newVal}`)
})
```

**watchEffect()**: auto-track
```js
watchEffect(() => {
  console.log(count.value) // auto re-run on count change
})
```

### Lifecycle

```js
import { onMounted, onUnmounted } from 'vue'
onMounted(() => { /* DOM ready */ })
onUnmounted(() => { /* cleanup */ })
```

### Component Props/Emits

```js
const props = defineProps<{ msg: string }>()
const emit = defineEmits<{ (e: 'update', value: string): void }>()
emit('update', 'new value')
```

### Dependency Injection

```js
// Parent
provide('key', value)
// Child
const injected = inject('key')
```

## Template Syntax

```vue
<template>
  <div v-if="show">Conditional</div>
  <ul>
    <li v-for="item in items" :key="item.id">{{ item.name }}</li>
  </ul>
  <input v-model="text" @input="handleInput" :disabled="loading" />
</template>
```

**Directives**: `v-if`, `v-for`, `v-model` (two-way), `v-bind` (`:attr`), `v-on` (`@event`).
