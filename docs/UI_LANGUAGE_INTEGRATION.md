# UI Language & Integration Guide

## Overview

This document describes the new language mode system and integrated UI panels for Notes and Tasks.

## Features Implemented

### 1. Language Mode System

**Two Modes:**
- **Techy Mode**: Technical jargon, terminal-style language (e.g., "INITIALIZE_CONNECTION", "DATA_LOGS", "TASK_QUEUE")
- **Normie Mode**: Simple, friendly language (e.g., "Start a chat", "Notes", "Tasks")

**Implementation:**
- `LanguageContext.tsx` - Global context provider
- Modal on first launch to select preference
- Stored in localStorage
- Can be changed anytime via settings menu

**Usage in Components:**
```tsx
import { useLanguage } from '@/context/LanguageContext';

function MyComponent() {
  const { t } = useLanguage();
  
  return (
    <div>
      <h1>{t('SYSTEM_ONLINE', 'System Ready')}</h1>
    </div>
  );
}
```

### 2. Integrated Panels

**Notes Panel** (`NotesPanel.tsx`):
- Slides in from right side
- Full CRUD operations (Create, Read, Update, Delete)
- Pin/unpin notes
- Matches cyberpunk design system
- Language mode support

**Tasks Panel** (`TasksPanel.tsx`):
- Slides in from right side
- Add, complete, delete tasks
- Filter by status (all, active, completed)
- Task counter
- Matches cyberpunk design system
- Language mode support

**Benefits:**
- No page navigation required
- Quick access from any view
- Consistent with main app design
- Better mobile experience

### 3. Updated Sidebar

**Changes:**
- Navigation tabs now open panels instead of navigating
- Added settings dropdown with language mode toggle
- Integrated panel management
- Maintains conversation list

**Navigation Flow:**
```
Sidebar Tabs:
├── Chats → Main chat view (default)
├── Notes → Opens NotesPanel (right side)
└── Tasks → Opens TasksPanel (right side)

Settings Menu:
├── Language Mode Toggle
│   ├── Techy
│   └── Normie
└── Logout
```

## Design System

### Color Palette
- **Primary**: Cyan (#06B6D4) - Main accent color
- **Secondary**: Orange (#F97316) - Group chats, normie mode
- **Background**: Dark (#0a0a0f, #050508) - Main backgrounds
- **Success**: Emerald (#10B981) - Online status, completed tasks
- **Danger**: Red (#EF4444) - Delete actions, offline status

### Typography
- **Font**: Monospace (system font)
- **Headers**: Uppercase with wide letter spacing
- **Body**: Normal case, readable size
- **Emphasis**: Tracking (letter-spacing) for important text

### Visual Elements
- **Clipped Corners**: `clip-path: polygon(...)` for futuristic look
- **Borders**: Thin, colored borders with transparency
- **Backgrounds**: Dark with subtle gradients
- **Grid Pattern**: Animated grid background for depth
- **Glow Effects**: Box shadows with color for emphasis

### Component Patterns

**Button Styles:**
```tsx
// Primary Action
className="bg-cyan-900/30 border border-cyan-500/50 text-cyan-400 hover:bg-cyan-500 hover:text-black"

// Secondary Action
className="border border-gray-700 text-gray-500 hover:border-cyan-500/50"

// Danger Action
className="text-red-500 hover:text-red-400"
```

**Input Styles:**
```tsx
className="bg-[#050508] border border-cyan-800/50 text-cyan-100 focus:border-cyan-500"
```

**Panel Container:**
```tsx
className="fixed inset-y-0 right-0 w-full sm:w-96 bg-[#0a0a0f] border-l border-cyan-900/30 z-50"
```

## Language Translation Examples

### Common Translations

| Techy | Normie |
|-------|--------|
| SELECT_TARGET | Select a Chat |
| INITIALIZE_CONNECTION | Start a Chat |
| DATA_LOGS | Notes |
| TASK_QUEUE | Tasks |
| NO_CONNECTIONS | No chats yet |
| TERMINATE DATA LOG? | Delete this note? |
| NEW_LOG | New Note |
| ENTER_DATA... | Start writing... |
| LOADING_DATA... | Loading... |
| SYSTEM_ONLINE | System Ready |
| CONNECTION_ACTIVE | Online |
| CONNECTION_LOST | Offline |

### Adding New Translations

When adding new text to the UI:

1. Use the `t()` function from `useLanguage()`
2. First parameter: Techy version (uppercase, technical)
3. Second parameter: Normie version (normal case, friendly)

```tsx
const { t } = useLanguage();

// Example
<button>{t('EXECUTE_COMMAND', 'Submit')}</button>
<p>{t('PROCESSING_REQUEST...', 'Please wait...')}</p>
<h1>{t('SYSTEM_DIAGNOSTICS', 'Settings')}</h1>
```

## Mobile Responsiveness

### Breakpoints
- **Mobile**: < 640px (sm)
- **Tablet**: 640px - 1024px
- **Desktop**: > 1024px (lg)

### Panel Behavior
- **Mobile**: Full-screen overlay
- **Desktop**: Side panel (384px width)

### Sidebar Behavior
- **Mobile**: Collapsible with hamburger menu
- **Desktop**: Always visible

## File Structure

```
src/
├── context/
│   ├── AuthContext.tsx
│   └── LanguageContext.tsx ✨ NEW
├── components/
│   ├── Sidebar.tsx (updated)
│   ├── NotesPanel.tsx ✨ NEW
│   └── TasksPanel.tsx ✨ NEW
├── app/
│   ├── layout.tsx (updated with LanguageProvider)
│   └── chat/
│       ├── page.tsx (updated with language support)
│       ├── notes/ (deprecated - now uses panel)
│       └── tasks/ (deprecated - now uses panel)
```

## Testing Checklist

- [ ] Language mode modal appears on first launch
- [ ] Can select Techy or Normie mode
- [ ] Preference persists after refresh
- [ ] Can change mode in settings
- [ ] Notes panel opens/closes correctly
- [ ] Tasks panel opens/closes correctly
- [ ] Panels work on mobile (full screen)
- [ ] Panels work on desktop (side-by-side)
- [ ] All CRUD operations work in panels
- [ ] Design matches main app theme
- [ ] Translations display correctly in both modes

## Future Enhancements

### Potential Additions
1. **More Language Modes**: Add "Professional", "Casual", "Gamer" modes
2. **Custom Translations**: Allow users to customize specific terms
3. **Localization**: Add support for multiple languages (i18n)
4. **Theme Variants**: Different color schemes per mode
5. **Accessibility**: Screen reader support for mode differences

### Panel Improvements
1. **Drag & Drop**: Reorder notes/tasks
2. **Categories**: Organize notes into folders
3. **Tags**: Add tags to notes and tasks
4. **Search**: Search within notes/tasks
5. **Export**: Export notes as markdown/PDF

## Troubleshooting

### Language mode not persisting
- Check localStorage: `localStorage.getItem('netlink_language_mode')`
- Clear and try again: `localStorage.removeItem('netlink_language_mode')`

### Panels not opening
- Check z-index conflicts
- Verify state management in Sidebar
- Check console for errors

### Design inconsistencies
- Verify Tailwind classes are correct
- Check for conflicting CSS
- Ensure all components use design system colors

## API Integration

The panels use existing API endpoints:

**Notes:**
- `GET /api/notes` - Get all notes
- `POST /api/notes` - Create note
- `PUT /api/notes/:id` - Update note
- `DELETE /api/notes/:id` - Delete note
- `POST /api/notes/pin?id=:id` - Toggle pin

**Tasks:**
- `GET /api/tasks` - Get all tasks
- `POST /api/tasks` - Create task
- `POST /api/tasks/toggle?id=:id` - Toggle completion
- `POST /api/tasks/delete?id=:id` - Delete task

No backend changes required for this feature.

## Performance Considerations

- Language preference loaded once on mount
- Panels lazy-loaded (only rendered when open)
- LocalStorage used for persistence (no API calls)
- Smooth animations with CSS transitions
- Optimized re-renders with proper state management

## Accessibility

- Keyboard navigation supported
- Focus management in modals
- ARIA labels on interactive elements
- Color contrast meets WCAG AA standards
- Screen reader friendly text alternatives

---

**Last Updated**: January 2026
**Version**: 1.0.0
**Status**: ✅ Implemented and Ready
