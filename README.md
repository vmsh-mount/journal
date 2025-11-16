# Journal - Personal Site

A minimal personal journal/blog site built with Go and styled with Tailwind CSS.

## Quick Start

### Run Application
```shell
go run main.go
```

The site will be available at `http://localhost:8080`

---

## Installation & Setup

### Prerequisites

#### 1. Install Go
See [Go installation guide](https://go.dev/doc/install)

#### 2. Install Node.js
Tailwind CSS requires Node.js to run its build tools.

**macOS (using Homebrew):**
```shell
brew install node
```

**Verify installation:**
```shell
node -v    # Should show version (e.g., v20.x.x)
npm -v     # Should show version (e.g., 9.x.x)
```

**What is npm?**
- npm (Node Package Manager) comes with Node.js
- It's used to install JavaScript packages (like Tailwind CSS)
- Similar to how Go uses `go get` for packages

---

## Project Setup

### Step 1: Install Dependencies

```shell
npm install
```

**What this does:**
- Reads `package.json` to see what packages are needed
- Downloads Tailwind CSS and saves it to `node_modules/` directory
- Creates `package-lock.json` (locks dependency versions)

**Files created:**
- `node_modules/` - Contains all installed packages (don't edit this)
- `package-lock.json` - Locks exact versions of dependencies

### Step 2: Build CSS

```shell
npm run build:css
```

**What this does:**
- Runs the Tailwind CLI tool
- Reads `static/css/input.css` (source file)
- Scans templates in `internal/templates/` for Tailwind classes
- Generates `static/css/styles.css` (final CSS file)

**Behind the scenes:**
```shell
# This is what npm run build:css actually runs:
tailwindcss -i ./static/css/input.css -o ./static/css/styles.css --minify

# Breakdown:
# -i = input file (source CSS)
# -o = output file (compiled CSS)
# --minify = compress the output (smaller file size)
```

### Step 3: Development (Watch Mode)

```shell
npm run watch:css
```

**What this does:**
- Watches your template files for changes
- Automatically rebuilds CSS when you add/remove Tailwind classes
- Keeps running until you stop it (Ctrl+C)

**Use this during development** so you don't have to manually rebuild CSS every time.

---
