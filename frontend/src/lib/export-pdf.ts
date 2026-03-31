import { marked } from 'marked'

/**
 * Opens a print-optimized window with beautifully rendered markdown content.
 * Uses the browser's native print dialog which supports "Save as PDF".
 */
export function exportToPdf(markdown: string, title: string) {
  const win = window.open('', '_blank')
  if (!win) return

  const html = marked.parse(markdown, { gfm: true, breaks: false }) as string
  const date = new Date().toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })

  win.document.write(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<title>${esc(title)}</title>
<style>
@page {
  margin: 0.9in 0.8in;
  size: A4;
}
@media print {
  .toolbar { display: none !important; }
  body { padding: 0; }
}
*, *::before, *::after { box-sizing: border-box; }
body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif;
  font-size: 10.5pt;
  line-height: 1.7;
  color: #1e293b;
  max-width: 720px;
  margin: 0 auto;
  padding: 32px 24px;
  -webkit-font-smoothing: antialiased;
}

/* ── Toolbar (screen only) ── */
.toolbar {
  position: sticky;
  top: 0;
  z-index: 10;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
  margin: -32px -24px 24px;
  padding: 12px 24px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.toolbar button {
  background: #111827;
  color: #fff;
  border: none;
  padding: 8px 20px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}
.toolbar button:hover { opacity: 0.9; }
.toolbar .hint { font-size: 12px; color: #9ca3af; }

/* ── Header ── */
.doc-header {
  margin-bottom: 28px;
  padding-bottom: 16px;
  border-bottom: 3px solid #111827;
}
.doc-header h1 {
  font-size: 20pt;
  font-weight: 800;
  color: #111827;
  letter-spacing: -0.02em;
  margin: 0 0 4px;
  line-height: 1.2;
}
.doc-header .meta {
  font-size: 9pt;
  color: #94a3b8;
  display: flex;
  align-items: center;
  gap: 8px;
}
.doc-header .meta .dot { width: 3px; height: 3px; border-radius: 50%; background: #cbd5e1; }
.doc-header .badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 8pt;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
  letter-spacing: 0.02em;
}

/* ── Content typography ── */
.content h1 {
  font-size: 17pt;
  font-weight: 800;
  color: #111827;
  margin: 32pt 0 10pt;
  padding-bottom: 8pt;
  border-bottom: 2px solid #e5e7eb;
  letter-spacing: -0.01em;
}
.content h2 {
  font-size: 14pt;
  font-weight: 700;
  color: #111827;
  margin: 26pt 0 8pt;
  padding-bottom: 6pt;
  border-bottom: 1px solid #f1f5f9;
}
.content h3 {
  font-size: 11.5pt;
  font-weight: 700;
  color: #1e293b;
  margin: 20pt 0 6pt;
}
.content h4 {
  font-size: 10.5pt;
  font-weight: 600;
  color: #334155;
  margin: 16pt 0 4pt;
}
.content p {
  margin: 0 0 10pt;
  color: #334155;
}
.content strong {
  font-weight: 700;
  color: #0f172a;
}
.content em {
  font-style: italic;
  color: #64748b;
}
.content a {
  color: #4f46e5;
  text-decoration: none;
  border-bottom: 1px solid #c7d2fe;
}

/* ── Lists ── */
.content ul, .content ol {
  margin: 0 0 10pt;
  padding-left: 20pt;
  color: #334155;
}
.content li {
  margin: 3pt 0;
  padding-left: 2pt;
}
.content li::marker {
  color: #94a3b8;
}

/* ── Code ── */
.content code {
  font-family: 'SF Mono', 'Fira Code', 'JetBrains Mono', 'Consolas', monospace;
  font-size: 9pt;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  padding: 1.5pt 5pt;
  color: #0f172a;
}
.content pre {
  margin: 12pt 0;
  background: #0f172a;
  border: 1px solid #1e293b;
  border-radius: 8px;
  padding: 14pt 18pt;
  overflow-x: auto;
}
.content pre code {
  background: none;
  border: none;
  padding: 0;
  color: #e2e8f0;
  font-size: 9pt;
  line-height: 1.6;
}

/* ── Blockquote ── */
.content blockquote {
  margin: 12pt 0;
  padding: 10pt 16pt;
  border-left: 4px solid #6366f1;
  background: #faf5ff;
  border-radius: 0 6px 6px 0;
  color: #475569;
}
.content blockquote p {
  margin: 0;
  color: #475569;
}

/* ── Table ── */
.content table {
  width: 100%;
  border-collapse: collapse;
  margin: 12pt 0;
  font-size: 9.5pt;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  overflow: hidden;
}
.content thead {
  background: #f8fafc;
}
.content th {
  text-align: left;
  padding: 8pt 10pt;
  border-bottom: 2px solid #e2e8f0;
  font-size: 8.5pt;
  font-weight: 700;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.content td {
  padding: 7pt 10pt;
  border-top: 1px solid #f1f5f9;
  color: #334155;
}
.content tr:nth-child(even) {
  background: #fafbfc;
}

/* ── HR ── */
.content hr {
  border: none;
  border-top: 1px solid #e5e7eb;
  margin: 20pt 0;
}

/* ── Footer ── */
.doc-footer {
  margin-top: 36pt;
  padding-top: 12pt;
  border-top: 1px solid #e5e7eb;
  font-size: 8pt;
  color: #94a3b8;
  text-align: center;
  display: flex;
  justify-content: space-between;
}
</style>
</head>
<body>
<div class="toolbar">
  <button onclick="window.print()">Save as PDF</button>
  <span class="hint">or press Ctrl/Cmd + P</span>
</div>
<div class="doc-header">
  <h1>${esc(title)}</h1>
  <div class="meta">
    <span class="badge">AI Generated</span>
    <span class="dot"></span>
    <span>UNM Platform</span>
    <span class="dot"></span>
    <span>${date}</span>
  </div>
</div>
<div class="content">
${html}
</div>
<div class="doc-footer">
  <span>UNM Platform — AI Advisor</span>
  <span>${date}</span>
</div>
</body>
</html>`)
  win.document.close()
}

function esc(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}
