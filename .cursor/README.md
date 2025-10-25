# Cursor Custom Commands

This folder contains custom slash commands for Cursor AI.

## Available Commands

After setup, you can use these commands by typing `/` in the chat:

- **`/review`** - Quick code review notes
- **`/explain`** - Code explanation notes
- **`/debug`** - Debugging analysis
- **`/perf`** - Performance optimization notes
- **`/design`** - System design review
- **`/arch`** - Architecture guidance
- **`/notes`** - Technical notes creation

## How to Use

1. **Select code or text** in your editor
2. **Type `/` in Cursor Chat** to see available commands
3. **Choose a command** (e.g., `/review`)
4. The AI will apply the corresponding prompt from `prompts/` folder

## Manual Setup (If Needed)

If the commands don't appear automatically, you can:

### Method 1: Copy-Paste Prompts
1. Open the prompt file you want (e.g., `prompts/code-reviewer.md`)
2. Copy its content
3. Paste into Cursor chat
4. Add your code/question below

### Method 2: Use @ Symbol
1. In Cursor chat, type `@` 
2. Select "File" and choose a prompt file
3. This includes the prompt in your conversation

### Method 3: Add to Cursor Settings
1. Open Cursor Settings (`Ctrl+,` or `Cmd+,`)
2. Go to "Cursor" → "Rules for AI"
3. Add custom rules or commands manually

## Prompt Files Location

All prompt files are in: `../prompts/`

- `code-reviewer.md`
- `code-explainer.md`
- `debugging-assistant.md`
- `performance-optimizer.md`
- `system-design-reviewer.md`
- `architecture-advisor.md`
- `technical-writer.md`

