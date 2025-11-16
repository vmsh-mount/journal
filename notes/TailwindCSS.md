## Understanding Tailwind CSS

### What is Tailwind CSS?

Tailwind CSS is a **utility-first CSS framework**. Instead of writing custom CSS, you apply pre-built utility classes directly in your HTML templates. For example:

```html
<!-- Instead of writing CSS like: -->
<!-- .button { padding: 1rem; background: blue; } -->

<!-- You write: -->
<button class="px-4 py-4 bg-blue-500">Click me</button>
```

**Key Benefits:**
- **Faster development**: No context switching between HTML and CSS files
- **Smaller CSS files**: Only the classes you use are included in the final CSS
- **Consistent design**: Pre-defined spacing, colors, and sizes
- **Highly customizable**: Easy to extend with your own design tokens

### How Tailwind Works Under the Hood

1. **Source File (`input.css`)**: Contains Tailwind directives (`@tailwind base`, `@tailwind components`, `@tailwind utilities`)
2. **Content Scanning**: Tailwind scans your HTML templates (defined in `tailwind.config.js`) for class names
3. **CSS Generation**: Tailwind generates only the CSS for classes it finds
4. **Output File (`styles.css`)**: The final, optimized CSS file that your HTML references

**Example Flow:**
```
Your HTML: <div class="bg-blue-500 px-4">
           ↓
Tailwind scans and finds: bg-blue-500, px-4
           ↓
Generates CSS: .bg-blue-500 { background-color: #3b82f6; }
               .px-4 { padding-left: 1rem; padding-right: 1rem; }
           ↓
Output: styles.css (only contains CSS for classes you actually used)
```
---
## Configuration Deep Dive

### `package.json`

```json
{
  "scripts": {
    "build:css": "tailwindcss -i ./static/css/input.css -o ./static/css/styles.css --minify",
    "watch:css": "tailwindcss -i ./static/css/input.css -o ./static/css/styles.css --watch"
  },
  "devDependencies": {
    "tailwindcss": "^3.4.1"
  }
}
```

**Explanation:**
- **`scripts`**: Shortcuts for common commands (run with `npm run <script-name>`)
- **`devDependencies`**: Packages only needed during development (not in production)
- **`^3.4.1`**: Version constraint (^ means "compatible with 3.4.1")

### `tailwind.config.js`

```javascript
module.exports = {
  // Where to look for Tailwind classes
  content: [
    "./internal/templates/**/*.html",  // All HTML files in templates/
    "./internal/handlers/**/*.go",      // Go files (in case classes are in strings)
  ],
  
  theme: {
    extend: {
      // Custom font family
      fontFamily: {
        sans: ['Inter', 'system-ui', '...'],  // Default sans-serif font
      },
      
      // Custom max-width
      maxWidth: {
        'content': '680px',  // Use with: max-w-content
      },
      
      // Custom colors
      colors: {
        'text': '#111',           // Use with: text-text
        'text-muted': '#555',     // Use with: text-text-muted
      },
    },
  },
  
  plugins: [],  // Can add Tailwind plugins here
}
```

**Key Concepts:**

1. **`content`**: Tells Tailwind which files to scan for class names
    - `**/*.html` means "all HTML files in any subdirectory"
    - Only classes found in these files will be included in final CSS

2. **`theme.extend`**: Adds custom values without replacing defaults
    - `fontFamily.sans` extends the default sans-serif stack
    - `maxWidth.content` creates a new utility: `max-w-content`
    - `colors` adds custom color utilities

3. **Why `extend`?**
    - Keeps all default Tailwind utilities
    - Adds your custom values on top
    - If you used `theme` (without extend), you'd lose all defaults

### `static/css/input.css`

```css
@tailwind base;        /* Tailwind's base styles (normalize, etc.) */
@tailwind components;  /* Component classes (like .container) */
@tailwind utilities;   /* Utility classes (like .bg-blue-500, .px-4) */

@layer base {
  body {
    @apply bg-[#f8f7f2] text-text font-sans leading-relaxed text-base;
  }
}

@layer components {
  .container {
    @apply w-full max-w-content mx-auto px-4;
  }
  /* ... more component classes ... */
}
```

**Explanation:**

1. **`@tailwind` directives**: These are replaced with actual CSS during build
    - `base`: Resets, typography defaults
    - `components`: Reusable component classes
    - `utilities`: All utility classes (bg-*, px-*, etc.)

2. **`@layer`**: Organizes your custom CSS
    - `base`: Base element styles (body, headings, etc.)
    - `components`: Reusable component classes
    - `utilities`: Custom utility classes

3. **`@apply`**: Uses Tailwind utilities inside CSS
    - `@apply bg-[#f8f7f2]` is equivalent to `background-color: #f8f7f2;`
    - Allows you to create component classes using Tailwind utilities

4. **Arbitrary values**: `bg-[#f8f7f2]` uses square brackets for custom values
    - Useful for one-off values not in your theme
    - `[#f8f7f2]` is a custom color
    - `[680px]` would be a custom size

---

## File Structure

```
journal/
├── package.json              # Node.js dependencies and scripts
├── package-lock.json         # Locked dependency versions
├── tailwind.config.js        # Tailwind configuration
├── node_modules/             # Installed packages (gitignored)
│
├── static/
│   └── css/
│       ├── input.css         # Source CSS (edit this)
│       └── styles.css        # Compiled CSS (generated, don't edit)
│
└── internal/
    └── templates/            # HTML templates (Tailwind scans these)
        ├── base.html
        ├── articles.html
        └── ...
```

**Important:**
- ✅ **Edit**: `input.css`, templates, `tailwind.config.js`
- ❌ **Don't edit**: `styles.css` (it's auto-generated)
- ❌ **Don't commit**: `node_modules/` (in `.gitignore`)

---

## Development Workflow

### Two-Terminal Setup

**Terminal 1 - CSS Watcher:**
```shell
npm run watch:css
```
Keeps running, watches for changes, auto-rebuilds CSS.

**Terminal 2 - Go Server:**
```shell
go run main.go
```
Runs your web server.

### Workflow Steps

1. **Make changes** to HTML templates (add/remove Tailwind classes)
2. **CSS auto-rebuilds** (if watch mode is running)
3. **Refresh browser** to see changes

**Example:**
```html
<!-- Change this: -->
<h1 class="text-2xl">Title</h1>

<!-- To this: -->
<h1 class="text-4xl font-bold text-blue-500">Title</h1>
```

The watcher detects the change, rebuilds CSS, and you refresh to see the new styling.

---

## Common Tailwind Concepts

### Spacing Scale

Tailwind uses a consistent spacing scale (based on 0.25rem = 4px):

- `p-1` = padding: 0.25rem (4px)
- `p-4` = padding: 1rem (16px)
- `p-8` = padding: 2rem (32px)
- `m-2` = margin: 0.5rem (8px)
- `mx-auto` = margin-left: auto; margin-right: auto (centers element)

### Responsive Design

```html
<div class="text-sm md:text-base lg:text-lg">
  Responsive text
</div>
```

- `text-sm`: Base size (mobile)
- `md:text-base`: Medium screens and up
- `lg:text-lg`: Large screens and up

### Color System

```html
<div class="bg-blue-500 text-white">
  Blue background, white text
</div>
```

- `bg-{color}-{shade}`: Background color
- `text-{color}-{shade}`: Text color
- Shades: 50 (lightest) to 900 (darkest), 500 is middle

### Custom Colors (Our Setup)

```html
<div class="text-text">Main text color</div>
<div class="text-text-muted">Muted text color</div>
```

These are defined in `tailwind.config.js` under `colors`.

### Flexbox & Grid

```html
<div class="flex items-center justify-between">
  <span>Left</span>
  <span>Right</span>
</div>
```

- `flex`: display: flex
- `items-center`: align-items: center
- `justify-between`: justify-content: space-between

---

## Troubleshooting

### CSS not updating?

1. **Check if watcher is running**: `npm run watch:css`
2. **Manually rebuild**: `npm run build:css`
3. **Hard refresh browser**: Cmd+Shift+R (Mac) or Ctrl+Shift+R (Windows)

### Classes not working?

1. **Check if class exists**: [Tailwind Docs](https://tailwindcss.com/docs)
2. **Verify file is scanned**: Check `content` in `tailwind.config.js`
3. **Rebuild CSS**: `npm run build:css`

### npm install fails?

1. **Check Node.js version**: `node -v` (should be 16+)
2. **Clear cache**: `npm cache clean --force`
3. **Delete node_modules**: `rm -rf node_modules && npm install`

---

## Production Build

Before deploying:

```shell
npm run build:css
```

This creates a minified `styles.css` file optimized for production.

**Note**: The `--minify` flag removes whitespace and comments, making the file smaller for faster loading.

---

## Resources

- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [Tailwind CSS Cheat Sheet](https://nerdcave.com/tailwind-cheat-sheet)
- [Inter Font](https://rsms.me/inter/)

---

## Project Structure

- `static/css/input.css` - Tailwind source file with directives
- `static/css/styles.css` - Compiled CSS (generated, don't edit directly)
- `tailwind.config.js` - Tailwind configuration
- `internal/templates/` - HTML templates (Tailwind scans these for classes)
- `package.json` - Node.js dependencies and build scripts

---