# Interactive System Design Mind Map

This is an interactive version of the system design mind map with zoom, pan, and search capabilities.

## Files

- `interactive-mind-map.html` - The main HTML file with the interactive visualization and navigation
- `interactive-mind-map.md` - The markdown source data for the System Design mind map
- `java-mind-map.md` - The markdown source data for the Java mind map (Basics to Pro)
- `system-design-course-mind-map.md` - The markdown source data for the System Design Course mind map (Week by Week)
- `package.json` - npm scripts for easy server startup
- `preview_mindmap.js` - Node.js script to preview the mind map (optional wrapper)
- `preview_mindmap.py` - Python script to preview the mind map (alternative)

## How to View

### Option 1: Using npx http-server (Recommended - No Dependencies!)

Simply run:
```bash
# Navigate to the mind-map-interactive directory
cd mind-map-interactive

# Start the server (automatically opens browser)
npx http-server -p 8888 -o interactive-mind-map.html
```

Or use the npm script:
```bash
npm start
```

This requires **no installation** - `npx` downloads and runs `http-server` on the fly!

### Option 2: Using the Node.js Preview Script

If you want a wrapper with better output:
```bash
# Navigate to the mind-map-interactive directory
cd mind-map-interactive

# Run the Node.js script
node preview_mindmap.js
```

### Option 3: Direct File Opening
Simply open `interactive-mind-map.html` in your web browser. However, due to browser security restrictions, this might not work properly with loading the markdown file. Using an HTTP server is recommended.

### Option 4: Alternative Servers
You can use any HTTP server to serve the files:
```bash
# Using Python's built-in server
python -m http.server 8888

# Using the Python preview script
python preview_mindmap.py

# Using PHP
php -S localhost:8888
```

## Features

### Mind Maps Available
- **Home** - System Design Mind Map: Comprehensive guide to large-scale software systems
- **Java** - Java: From Basics to Pro: Complete Java journey from fundamentals to advanced topics including Spring, Spring Boot, Spring Security, JPA, and more
- **Course** - System Design Course: Week by Week: Organized course content covering Week 1-6 topics including caching, databases, scaling, distributed systems, CDN, storage systems, and more

### Navigation
- **Top Navigation Bar** - Switch between different mind maps using the Home, Java, and Course buttons
- **Click** nodes to expand/collapse branches
- **Drag** to pan around the mind map
- **Scroll** or use zoom buttons to zoom in/out
- **Search** for specific topics using the search box

### Keyboard Shortcuts
- `+` or `=` - Zoom in
- `-` or `_` - Zoom out
- `0` - Fit to screen
- `Ctrl/Cmd + F` - Focus search box

### Controls
- **Zoom In/Out** - Adjust the view scale
- **Fit Screen** - Reset view to fit entire mind map
- **Expand All** - Show all nodes
- **Collapse All** - Collapse all branches except root
- **Reset View** - Return to initial state

### Touch Support
On touch devices:
- Pinch to zoom
- Drag to pan

## Customization

### Changing Colors
Edit the color array in the JavaScript section of the HTML file:
```javascript
const colors = [
    '#FF6B6B', // Root
    '#4ECDC4', // Level 1
    // ... add more colors
];
```

### Adding Content
Edit the markdown files (`interactive-mind-map.md` or `java-mind-map.md`) using standard markdown heading hierarchy:
```markdown
# Root Topic
## Main Category
### Subcategory
- Item 1
- Item 2
```

### Adding New Mind Maps
To add a new mind map:
1. Create a new markdown file (e.g., `new-topic-mind-map.md`)
2. Add a new entry to the `mindMaps` object in `interactive-mind-map.html`:
```javascript
newTopic: {
    file: 'new-topic-mind-map.md',
    title: 'New Topic Mind Map',
    colors: [/* color array */]
}
```
3. Add a navigation button in the HTML:
```html
<button class="nav-button" id="nav-newtopic" data-route="newTopic">📚 New Topic</button>
```
4. Add a click handler in the JavaScript section

## Troubleshooting

1. **Mind map not loading**: Make sure both HTML and markdown files are in the same directory
2. **CORS errors**: Use `npx http-server` or any HTTP server (don't open HTML file directly)
3. **Search not working**: Ensure JavaScript is enabled in your browser
4. **npx not found**: Make sure Node.js (v14+) is installed. Download from [nodejs.org](https://nodejs.org/)

## Browser Compatibility
Works best in modern browsers:
- Chrome/Edge 80+
- Firefox 75+
- Safari 13+
