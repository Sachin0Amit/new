# Sovereign Design System

A unified, accessible, and responsive design system for the Sovereign Intelligence Core.

## Design Tokens

Tokens are defined in `web/src/styles/tokens.css` as CSS Custom Properties.

### Colors
| Token | Description |
|---|---|
| `--color-bg-primary` | Main background color |
| `--color-bg-secondary` | Secondary background (sidebar, cards) |
| `--color-accent` | Brand primary color (Deep Indigo) |
| `--color-text-primary` | Main text color |
| `--color-success` | Success/Healthy status |
| `--color-danger` | Error/Critical status |

### Layout & Spacing
| Token | Description |
|---|---|
| `--radius-md` | Standard border radius (8px) |
| `--radius-lg` | Large border radius (12px) |
| `--font-sans` | Primary sans-serif typeface |
| `--font-mono` | Monospaced typeface for data/logs |

## Components

Components are defined in `web/src/styles/components.css`.

### Buttons
- `.btn`: Base button styles.
- `.btn-primary`: High-emphasis brand action.
- `.btn-ghost`: Low-emphasis navigation/utility action.
- `.btn-danger`: Destructive actions.

### Data Display
- `.card`: Standard content container with border and radius.
- `.badge`: Tiny status labels or counters.
- `.status-dot`: 8px indicators for real-time node health.
- `.skeleton`: Animated loading state for async data.

## Accessibility

The system adheres to **WCAG 2.1 AA** standards:
- **Interactive States**: All buttons and links use `:focus-visible` with high-contrast outlines.
- **Color Contrast**: All text pairings meet the 4.5:1 minimum ratio.
- **Semantics**: Uses HTML5 landmarks (`role="main"`, `role="navigation"`) and ARIA labels.
- **Responsiveness**: Fluid layouts using CSS Grid (auto-fit) ensure usability from 320px to 4K.
