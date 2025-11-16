#!/usr/bin/env node
/**
 * Simple Node.js server wrapper for previewing the interactive mind map.
 * Uses http-server via npx (no dependencies needed).
 */

const { spawn } = require('child_process');
const { exec } = require('child_process');
const path = require('path');
const fs = require('fs');

const PORT = 8888;
const HOST = 'localhost';
const URL = `http://${HOST}:${PORT}/interactive-mind-map.html`;

// Check if files exist
const scriptDir = __dirname;
const htmlFile = path.join(scriptDir, 'interactive-mind-map.html');
const mdFile = path.join(scriptDir, 'interactive-mind-map.md');

if (!fs.existsSync(htmlFile)) {
    console.error('❌ Error: interactive-mind-map.html not found!');
    console.error('Make sure to run this script from the same directory as the HTML file.');
    process.exit(1);
}

if (!fs.existsSync(mdFile)) {
    console.warn('⚠️  Warning: interactive-mind-map.md not found!');
    console.warn('The mind map will not load without this file.');
}

console.log('🚀 Starting Mind Map Preview Server');
console.log('='.repeat(50));
console.log(`📍 Server will run at: ${URL}`);
console.log(`📁 Serving directory: ${scriptDir}`);
console.log('='.repeat(50));
console.log('\n✨ Features available:');
console.log('  • Click nodes to expand/collapse');
console.log('  • Drag to pan around');
console.log('  • Scroll to zoom in/out');
console.log('  • Use search box to find topics');
console.log('  • Keyboard shortcuts: +/- for zoom, 0 to fit, Ctrl+F to search');
console.log('\n🛑 Press Ctrl+C to stop the server\n');

// Start http-server using npx
const server = spawn('npx', ['http-server', '-p', PORT.toString(), '-o', 'interactive-mind-map.html'], {
    cwd: scriptDir,
    stdio: 'inherit',
    shell: true
});

// Handle process termination
process.on('SIGINT', () => {
    console.log('\n\n👋 Server stopped. Goodbye!');
    server.kill();
    process.exit(0);
});

server.on('error', (error) => {
    console.error('❌ Error starting server:', error.message);
    console.error('\n💡 Make sure Node.js is installed and npx is available.');
    process.exit(1);
});

