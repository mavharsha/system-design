# Custom Prompts

This folder contains custom prompts for various use cases.

## Using with Cursor

These prompts are set up for easy access in Cursor using slash commands:

| Command | Prompt File | Description |
|---------|-------------|-------------|
| `/review` | code-reviewer.md | Quick code review notes |
| `/explain` | code-explainer.md | Code explanation notes |
| `/debug` | debugging-assistant.md | Debugging analysis |
| `/perf` | performance-optimizer.md | Performance optimization |
| `/design` | system-design-reviewer.md | System design review |
| `/arch` | architecture-advisor.md | Architecture guidance |
| `/notes` | technical-writer.md | Technical notes |

**To use:** Select code/text → Type `/` in Cursor chat → Choose command

See [../.cursor/README.md](../.cursor/README.md) for setup details.

## Available Prompts

### Development & Code Quality
- **[code-reviewer.md](code-reviewer.md)** - Quick code review notes highlighting key issues and improvements
- **[code-explainer.md](code-explainer.md)** - Concise notes explaining code for learning and reference
- **[debugging-assistant.md](debugging-assistant.md)** - Structured debugging notes to identify and fix issues
- **[performance-optimizer.md](performance-optimizer.md)** - Performance analysis notes with actionable optimizations

### Architecture & Design
- **[system-design-reviewer.md](system-design-reviewer.md)** - Focused review notes on system designs
- **[architecture-advisor.md](architecture-advisor.md)** - Concise architectural guidance notes with practical recommendations

### Documentation
- **[technical-writer.md](technical-writer.md)** - Create clear, concise technical notes for quick reference

## Organization

You can organize your prompts in the following ways:

- **By Category**: Create subfolders like `ai-prompts/`, `templates/`, `code-generation/`
- **By Project**: Organize prompts specific to different projects or domains
- **By Tool**: Separate prompts for different AI tools or applications

## File Naming Convention

Consider using descriptive names for your prompt files:
- `system-design-reviewer.md` - Prompt for reviewing system design documents
- `code-explainer.md` - Prompt for explaining complex code
- `technical-writer.md` - Prompt for technical documentation writing
- `architecture-advisor.md` - Prompt for architecture recommendations

## Note-Taking Philosophy

These prompts are designed to create **concise, actionable notes** rather than comprehensive documentation:
- Focus on essentials and key insights
- Use bullet points and structured formats
- Prioritize by importance/impact
- Keep it scannable (5-10 minute read)
- Highlight the "why" and trade-offs
- Include practical examples

## Example Prompt Structure

```markdown
# [Prompt Name]

## Purpose
Brief description of what notes this prompt creates

## Prompt
[Structured instructions for creating notes]

**Notes Style:**
- Keep it concise
- Focus on actionable items
- Prioritize key information
```

## Tips

- Keep prompts versioned if they evolve over time
- Document the use case and expected outcomes
- Include examples of successful outputs when relevant
- Update prompts based on what works well

