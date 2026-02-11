---
name: design-system-documentation
description: Official design system documentation for Reefline dashboard, covering layout patterns, visual styles, and core components.
---

# Reefline Design System

This document outlines the UI/UX patterns and design principles used in the Reefline dashboard. Use these patterns when creating new pages or components to ensure consistency.

## 1. Core Principles

*   **Minimalist & Border-Focused:** We rely heavily on borders (`border-border`) to define structure rather than heavy shadows or distinct background colors for cards.
*   **Subtle Depth:** We use very subtle gradients (`from-primary/5`) and opacities to indicate active states or hover effects, avoiding flat solid colors for interactive elements.
*   **Data Density:** The interface is designed for power users (DevOps/Engineering), favoring dense information display (tables, lists, grids) with `text-sm` or `text-xs` sizing.
*   **Visual Polish:** Use of `backdrop-blur`, custom scrollbars, and `DotPattern` backgrounds adds a premium feel.

## 2. Layout Patterns

### A. Sidebar-Detail Layout
Used in **Settings** and **Overview** pages.

*   **Structure:**
    *   **Container:** `flex flex-col h-screen` (or `h-[calc(100vh-offset)]`).
    *   **Sidebar:** Fixed width (`w-64` or `w-56`), `border-r`, `bg-card/30` or transparent.
    *   **Main Content:** `flex-1`, `overflow-y-auto`.
    *   **Scrollbars:** Always use the custom transparent scrollbar utility.

```tsx
<div className="flex h-full">
  <aside className="w-64 border-r border-border shrink-0 overflow-y-auto [&::-webkit-scrollbar]:w-1.5 ...">
    {/* Navigation or Filters */}
  </aside>
  <main className="flex-1 overflow-y-auto [&::-webkit-scrollbar]:w-1.5 ...">
    {/* Page Content */}
  </main>
</div>
```

### B. "Grid System" Layout
Used in **Integrations** page.

*   **Concept:** A grid of cards that visually merge into a table-like structure.
*   **Implementation:** Cards use `rounded-none`, `border-0`, `border-r`, and `border-b`. The container handles the outer boundaries (e.g., `border-l`, `border-t`).

```tsx
// Container
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 border-l border-t border-border">
  {items.map(item => <Card key={item.id} />)}
</div>

// Integration Card Component
<Card className="rounded-none border-0 border-r border-b border-border shadow-none ...">
  {/* Content */}
</Card>
```

## 3. Component Patterns

### A. Page Header
Standardized header for all pages.

*   **Title:** `text-3xl` (or `2xl`), `font-medium`, `tracking-tight`.
*   **Gradient Effect:** `bg-gradient-to-br from-foreground to-foreground/70 bg-clip-text text-transparent`.
*   **Subtitle:** `text-muted-foreground`.

```tsx
<div className="p-4 md:p-6 border-b border-border">
  <h1 className="text-3xl font-medium tracking-tight bg-gradient-to-br from-foreground to-foreground/70 bg-clip-text text-transparent">
    Page Title
  </h1>
  <p className="text-muted-foreground mt-1">
    Page description goes here.
  </p>
</div>
```

### B. Section Header with Dotted Background
Used to separate sections in Settings or Sidebars.

*   **Component:** Uses `DotPattern` (from `@/components/ui/dot-pattern`).
*   **Style:** `border-b border-border`, light background (`bg-muted/5`).

```tsx
function SectionHeader({ title, icon: Icon }) {
  return (
    <div className="relative border-b border-border overflow-hidden bg-muted/5">
       <DotPattern className="absolute inset-0 fill-neutral-400/20" />
       <div className="relative z-10 flex items-center gap-2 px-4 py-2">
         <Icon className="h-4 w-4 text-muted-foreground" />
         <h3 className="text-sm font-medium">{title}</h3>
       </div>
    </div>
  )
}
```

### C. Sidebar Navigation Item
Used for switching tabs or sections.

*   **States:**
    *   **Idle:** `text-muted-foreground`.
    *   **Hover:** `text-foreground`, subtle border color change.
    *   **Active:** `bg-primary/[0.03]`, `text-foreground`, `border-primary` highlight (often a left border or icon box highlight).

### D. Custom Scrollbar
Apply these utility classes to any scrollable element (`overflow-y-auto`):

```css
[&::-webkit-scrollbar]:w-1.5
[&::-webkit-scrollbar-track]:bg-transparent
[&::-webkit-scrollbar-thumb]:bg-accent/30
[&::-webkit-scrollbar-thumb]:rounded-full
[&::-webkit-scrollbar-thumb:hover]:bg-muted-background/50
```

## 4. Visual Assets & Icons

*   **Icons:** Lucide React (`lucide-react`). Size `h-4 w-4` is standard for UI elements, `h-5 w-5` for primary actions.
*   **Colors:**
    *   `primary`: Brand color (often used with low opacity for backgrounds).
    *   `muted-foreground`: Secondary text.
    *   `border`: Standard borders.
    *   `destructive`: Error states or dangerous actions.

## 5. Examples

### Settings Page Pattern
Refer to: `frontend/dashboard/src/pages/settings/settings.page.tsx`
*   Features "Danger Zone" pattern (red background/border).
*   Form layouts with `max-w` constraints or `grid-cols-2`.

### Overview List Pattern
Refer to: `frontend/dashboard/src/pages/overview/overview.page.tsx`
*   Features "List Row" design for dense data items.
*   Horizontal flex layout with specific widths for columns (`min-w-[300px]`, `w-24`, etc.).
