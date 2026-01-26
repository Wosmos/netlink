# Quick Reference: UI Components & Language System

## Language Mode Usage

### Basic Usage
```tsx
import { useLanguage } from '@/context/LanguageContext';

function MyComponent() {
  const { t, mode, setMode } = useLanguage();
  
  return (
    <div>
      {/* Simple translation */}
      <h1>{t('SYSTEM_ACTIVE', 'System Active')}</h1>
      
      {/* Check current mode */}
      {mode === 'techy' && <TechyFeature />}
      
      {/* Change mode */}
      <button onClick={() => setMode('normie')}>
        Switch to Normie
      </button>
    </div>
  );
}
```

## Common Translations

```tsx
// Headers
t('DATA_LOGS', 'Notes')
t('TASK_QUEUE', 'Tasks')
t('SYSTEM_DIAGNOSTICS', 'Settings')

// Actions
t('INITIALIZE', 'Start')
t('TERMINATE', 'Delete')
t('EXECUTE', 'Submit')
t('ABORT', 'Cancel')

// Status
t('ONLINE', 'Online')
t('OFFLINE', 'Offline')
t('PROCESSING...', 'Loading...')
t('COMPLETE', 'Done')

// Messages
t('NO_DATA_FOUND', 'Nothing here yet')
t('CONNECTION_LOST', 'Disconnected')
t('OPERATION_SUCCESSFUL', 'Success!')
```

## Design System Classes

### Buttons
```tsx
// Primary
className="bg-cyan-900/30 border border-cyan-500/50 text-cyan-400 
           hover:bg-cyan-500 hover:text-black transition-all"

// Secondary  
className="border border-gray-700 text-gray-500 
           hover:border-cyan-500/50 hover:text-cyan-400"

// Danger
className="text-red-500 hover:text-red-400"

// With clipped corners
style={{ clipPath: 'polygon(8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%, 0 8px)' }}
```

### Inputs
```tsx
className="bg-[#050508] border border-cyan-800/50 text-cyan-100 
           font-mono text-sm focus:outline-none focus:border-cyan-500 
           placeholder-cyan-900"
```

### Panels/Modals
```tsx
// Panel container
className="fixed inset-y-0 right-0 w-full sm:w-96 
           bg-[#0a0a0f] border-l border-cyan-900/30 z-50"

// Modal overlay
className="fixed inset-0 bg-black/70 backdrop-blur-sm z-50"

// Modal content
className="bg-[#0a0a0f] border-2 border-cyan-500/50 p-6"
```

### Icons/Avatars
```tsx
// User avatar
className="w-10 h-10 bg-cyan-950/50 border border-cyan-500/50 
           text-cyan-400 flex items-center justify-center"
style={{ clipPath: 'polygon(20% 0, 100% 0, 100% 80%, 80% 100%, 0 100%, 0 20%)' }}

// Status indicator
className="w-3 h-3 rounded-full border-2 border-[#0a0a0f] 
           bg-emerald-500" // online
className="w-3 h-3 rounded-full border-2 border-[#0a0a0f] 
           bg-gray-600" // offline
```

### Text Styles
```tsx
// Header
className="text-xl font-mono text-cyan-400 uppercase tracking-widest"

// Subheader
className="text-sm font-mono text-cyan-100 uppercase tracking-wider"

// Body
className="text-sm text-cyan-100 font-mono"

// Muted
className="text-xs text-gray-500 font-mono"

// Label
className="text-[10px] text-cyan-500/50 uppercase font-mono"
```

## Component Patterns

### Loading State
```tsx
<div className="flex flex-col items-center gap-4">
  <div className="w-12 h-12 border-4 border-cyan-500/30 
                  border-t-cyan-500 rounded-full animate-spin"></div>
  <span className="text-cyan-500 font-mono text-xs animate-pulse">
    {t('LOADING_DATA...', 'Loading...')}
  </span>
</div>
```

### Empty State
```tsx
<div className="text-center p-8">
  <div className="w-16 h-16 mx-auto mb-4 border border-cyan-500/30 
                  flex items-center justify-center bg-cyan-950/20" 
       style={{ clipPath: 'polygon(20% 0, 100% 0, 100% 80%, 80% 100%, 0 100%, 0 20%)' }}>
    <svg className="w-8 h-8 text-cyan-500/50">...</svg>
  </div>
  <p className="text-cyan-500/50 font-mono text-xs uppercase">
    {t('NO_DATA_FOUND', 'Nothing here yet')}
  </p>
</div>
```

### List Item
```tsx
<div className="p-4 hover:bg-cyan-950/20 border-b border-cyan-900/20 
                transition-colors group">
  <div className="flex items-center gap-3">
    {/* Avatar */}
    <div className="w-10 h-10 bg-cyan-950/50 border border-cyan-500/50 
                    text-cyan-400 flex items-center justify-center">
      A
    </div>
    
    {/* Content */}
    <div className="flex-1 min-w-0">
      <p className="font-mono text-sm text-cyan-100 truncate">Title</p>
      <p className="text-xs text-gray-500 truncate font-mono">Subtitle</p>
    </div>
    
    {/* Action (hidden until hover) */}
    <button className="opacity-0 group-hover:opacity-100 
                       text-gray-500 hover:text-cyan-400">
      <svg>...</svg>
    </button>
  </div>
</div>
```

### Form
```tsx
<form onSubmit={handleSubmit} className="space-y-4">
  <div>
    <label className="text-xs text-cyan-500/70 font-mono uppercase mb-1 block">
      {t('INPUT_LABEL', 'Label')}
    </label>
    <input
      type="text"
      className="w-full px-3 py-2 bg-[#050508] border border-cyan-800/50 
                 text-cyan-100 font-mono text-sm focus:outline-none 
                 focus:border-cyan-500"
    />
  </div>
  
  <button
    type="submit"
    className="w-full py-2 bg-cyan-900/30 border border-cyan-500/50 
               text-cyan-400 font-mono text-xs uppercase 
               hover:bg-cyan-500 hover:text-black"
  >
    {t('SUBMIT', 'Submit')}
  </button>
</form>
```

## Color Reference

```tsx
// Backgrounds
bg-[#0a0a0f]  // Main dark
bg-[#050508]  // Darker
bg-[#0c0c14]  // Modals

// Borders
border-cyan-900/30   // Subtle
border-cyan-500/50   // Medium
border-cyan-500      // Strong

// Text
text-cyan-400        // Primary
text-cyan-100        // Body
text-gray-500        // Muted
text-cyan-500/50     // Very muted

// Accents
text-orange-400      // Groups
text-emerald-500     // Success/Online
text-red-500         // Danger/Delete
text-yellow-500      // Warning/Pinned
```

## Responsive Utilities

```tsx
// Mobile first
className="w-full sm:w-96"           // Full width on mobile, 384px on desktop
className="p-3 sm:p-4"               // Smaller padding on mobile
className="text-xs sm:text-sm"       // Smaller text on mobile
className="hidden sm:block"          // Hide on mobile
className="block sm:hidden"          // Show only on mobile

// Sidebar collapse
className="fixed lg:relative"        // Fixed on mobile, relative on desktop
className="-translate-x-full lg:translate-x-0"  // Hidden on mobile by default
```

## Animation Classes

```tsx
// Transitions
className="transition-colors duration-200"
className="transition-all duration-300"
className="transition-transform"

// Hover effects
className="hover:scale-110"
className="hover:bg-cyan-950/20"
className="hover:text-cyan-400"

// Animations
className="animate-spin"             // Loading spinner
className="animate-pulse"            // Pulsing text
className="animate-bounce"           // Scroll down button

// Custom animations (add to globals.css)
@keyframes fadePop {
  0% { opacity: 0; transform: scale(0.98); }
  100% { opacity: 1; transform: scale(1); }
}
.msg-enter { animation: fadePop 0.2s ease-out forwards; }
```

## Z-Index Layers

```tsx
z-0   // Background elements
z-10  // Content
z-20  // Headers/Footers
z-30  // Dropdowns
z-40  // Modals
z-50  // Panels
z-[9999]  // Language selection modal (first launch)
```

## Common SVG Icons

```tsx
// Chat
<path d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8..." />

// Settings
<path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724..." />

// Close
<path d="M6 18L18 6M6 6l12 12" />

// Check
<path d="M5 13l4 4L19 7" />

// Plus
<path d="M12 4v16m8-8H4" />

// Delete
<path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862..." />
```

---

**Pro Tips:**
- Always use `t()` for user-facing text
- Keep techy mode uppercase and technical
- Keep normie mode friendly and clear
- Use monospace font for consistency
- Add hover states for interactivity
- Test on mobile and desktop
- Maintain color consistency
