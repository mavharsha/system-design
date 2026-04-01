# Claude Code Custom Commands

This folder contains custom slash commands for Claude Code.

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
2. **Type `/` in Claude Code chat** to see available commands
3. **Choose a command** (e.g., `/review`)
4. The AI will apply the corresponding prompt from `prompts/` folder

## Manual Setup (If Needed)

If the commands don't appear automatically, you can:

### Method 1: Copy-Paste Prompts
1. Open the prompt file you want (e.g., `prompts/code-reviewer.md`)
2. Copy its content
3. Paste into Claude Code chat
4. Add your code/question below

### Method 2: Reference File
1. In Claude Code, reference the prompt file directly
2. This includes the prompt in your conversation

## Prompt Files Location

All prompt files are in: `../prompts/`

- `code-reviewer.md`
- `code-explainer.md`
- `debugging-assistant.md`
- `performance-optimizer.md`
- `system-design-reviewer.md`
- `architecture-advisor.md`
- `technical-writer.md`
