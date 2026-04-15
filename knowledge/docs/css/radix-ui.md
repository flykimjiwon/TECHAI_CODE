# Radix UI Primitives Reference

## Installation & Import Pattern
```bash
npm install radix-ui
```
```tsx
// v1.x: individual packages
import * as Dialog from "@radix-ui/react-dialog"
// v2+ (unified): from "radix-ui"
import { Dialog, DropdownMenu, Tooltip } from "radix-ui"
```

## Dialog
```tsx
import * as Dialog from "@radix-ui/react-dialog"

<Dialog.Root open={open} onOpenChange={setOpen}>
  <Dialog.Trigger asChild>
    <button>Open Dialog</button>
  </Dialog.Trigger>

  <Dialog.Portal>
    <Dialog.Overlay className="fixed inset-0 bg-black/50 data-[state=open]:animate-in data-[state=closed]:animate-out" />
    <Dialog.Content className="fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-lg p-6 shadow-xl">
      <Dialog.Title className="text-lg font-semibold">Title</Dialog.Title>
      <Dialog.Description className="text-sm text-gray-600">Description</Dialog.Description>

      {/* Content */}
      <p>Dialog body content</p>

      <Dialog.Close asChild>
        <button className="absolute top-4 right-4">✕</button>
      </Dialog.Close>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
```

## DropdownMenu
```tsx
import * as DropdownMenu from "@radix-ui/react-dropdown-menu"

<DropdownMenu.Root>
  <DropdownMenu.Trigger asChild>
    <button>Options</button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Portal>
    <DropdownMenu.Content className="bg-white rounded-md shadow-lg p-1 min-w-48" sideOffset={5}>
      <DropdownMenu.Label className="px-2 py-1 text-xs text-gray-500">Account</DropdownMenu.Label>
      <DropdownMenu.Item className="px-2 py-1.5 rounded cursor-pointer hover:bg-gray-100" onSelect={() => {}}>
        Profile
      </DropdownMenu.Item>
      <DropdownMenu.Separator className="my-1 border-t border-gray-200" />

      {/* CheckboxItem */}
      <DropdownMenu.CheckboxItem checked={checked} onCheckedChange={setChecked}>
        <DropdownMenu.ItemIndicator>✓</DropdownMenu.ItemIndicator>
        Show notifications
      </DropdownMenu.CheckboxItem>

      {/* RadioGroup */}
      <DropdownMenu.RadioGroup value={color} onValueChange={setColor}>
        <DropdownMenu.RadioItem value="red">
          <DropdownMenu.ItemIndicator>●</DropdownMenu.ItemIndicator>
          Red
        </DropdownMenu.RadioItem>
        <DropdownMenu.RadioItem value="blue">Blue</DropdownMenu.RadioItem>
      </DropdownMenu.RadioGroup>

      {/* Sub menu */}
      <DropdownMenu.Sub>
        <DropdownMenu.SubTrigger>More options →</DropdownMenu.SubTrigger>
        <DropdownMenu.Portal>
          <DropdownMenu.SubContent>
            <DropdownMenu.Item>Sub item</DropdownMenu.Item>
          </DropdownMenu.SubContent>
        </DropdownMenu.Portal>
      </DropdownMenu.Sub>
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>
```

## Tooltip
```tsx
import * as Tooltip from "@radix-ui/react-tooltip"

{/* Provider wraps the app (once, near root) */}
<Tooltip.Provider delayDuration={400} skipDelayDuration={300}>
  <Tooltip.Root>
    <Tooltip.Trigger asChild>
      <button>Hover me</button>
    </Tooltip.Trigger>
    <Tooltip.Portal>
      <Tooltip.Content
        className="bg-gray-900 text-white text-xs px-2 py-1 rounded shadow-lg"
        sideOffset={4}
        side="top"            // "top" | "bottom" | "left" | "right"
        align="center"        // "start" | "center" | "end"
      >
        Tooltip text
        <Tooltip.Arrow className="fill-gray-900" />
      </Tooltip.Content>
    </Tooltip.Portal>
  </Tooltip.Root>
</Tooltip.Provider>
```

## Popover
```tsx
import * as Popover from "@radix-ui/react-popover"

<Popover.Root>
  <Popover.Trigger asChild>
    <button>Open popover</button>
  </Popover.Trigger>
  <Popover.Portal>
    <Popover.Content
      className="bg-white rounded-lg shadow-xl p-4 w-72"
      sideOffset={8}
      align="start"
    >
      <p>Popover content</p>
      <Popover.Arrow className="fill-white" />
      <Popover.Close asChild>
        <button className="absolute top-2 right-2 text-gray-400">✕</button>
      </Popover.Close>
    </Popover.Content>
  </Popover.Portal>
</Popover.Root>
```

## Accordion
```tsx
import * as Accordion from "@radix-ui/react-accordion"

{/* Single open at a time */}
<Accordion.Root type="single" defaultValue="item-1" collapsible>
  <Accordion.Item value="item-1" className="border-b">
    <Accordion.Trigger className="flex w-full justify-between py-4 font-medium">
      Question 1
      <ChevronDownIcon className="transition-transform data-[state=open]:rotate-180" />
    </Accordion.Trigger>
    <Accordion.Content className="overflow-hidden data-[state=open]:animate-accordion-down data-[state=closed]:animate-accordion-up">
      <div className="pb-4">Answer 1</div>
    </Accordion.Content>
  </Accordion.Item>
</Accordion.Root>

{/* Multiple open simultaneously */}
<Accordion.Root type="multiple" defaultValue={["item-1"]}>
  <Accordion.Item value="item-1">...</Accordion.Item>
  <Accordion.Item value="item-2">...</Accordion.Item>
</Accordion.Root>
```

## Tabs
```tsx
import * as Tabs from "@radix-ui/react-tabs"

<Tabs.Root defaultValue="tab1" orientation="horizontal">
  <Tabs.List className="flex border-b" aria-label="Navigation">
    <Tabs.Trigger
      value="tab1"
      className="px-4 py-2 data-[state=active]:border-b-2 data-[state=active]:border-blue-500"
    >
      Tab 1
    </Tabs.Trigger>
    <Tabs.Trigger value="tab2">Tab 2</Tabs.Trigger>
  </Tabs.List>
  <Tabs.Content value="tab1" className="p-4">Content 1</Tabs.Content>
  <Tabs.Content value="tab2" className="p-4">Content 2</Tabs.Content>
</Tabs.Root>
```

## Common Patterns

### asChild (Polymorphic)
```tsx
{/* Merges props onto child element instead of rendering a wrapper */}
<Dialog.Trigger asChild>
  <Button variant="outline">Open</Button>   {/* Button renders, not a div+button */}
</Dialog.Trigger>
```

### forceMount (Keep in DOM)
```tsx
{/* Keeps content mounted even when closed (useful for animations) */}
<Dialog.Content forceMount>
  <motion.div animate={open ? "open" : "closed"}>...</motion.div>
</Dialog.Content>
```

### data-state Styling
```tsx
{/* Radix exposes data-[state=open|closed|active|inactive|checked|unchecked] */}
<Accordion.Trigger className="
  data-[state=open]:text-blue-600
  data-[state=open]:bg-blue-50
">
<Tooltip.Content className="
  data-[state=instant-open]:animate-in
  data-[state=delayed-open]:animate-in
  data-[state=closed]:animate-out
">
```

### Portal (Escape z-index)
```tsx
{/* Renders outside the current DOM hierarchy — avoids overflow:hidden clipping */}
<Dialog.Portal container={document.getElementById("modal-root")}>
  <Dialog.Content>...</Dialog.Content>
</Dialog.Portal>
```
