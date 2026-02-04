# UI Prototypes

This directory contains three different UI prototypes for the Consensus AI application.

## 1. Web Simple (Static HTML)

The simplest option - just open the HTML file directly in your browser:

```bash
open ui-prototypes/web-simple/index.html
```

No dependencies or build step required.

## 2. React App

A React single-page application with Tailwind CSS styling.

```bash
cd ui-prototypes/react-app
npm install
npm start
```

This will start a dev server at http://localhost:3000

## 3. Electron App (Desktop)

A native desktop application built with Electron.

```bash
cd ui-prototypes/electron-app
npm install
npm start
```

This will launch a native desktop window.

## Quick Reference

| Prototype | Type | Command |
|-----------|------|---------|
| web-simple | Static HTML | `open ui-prototypes/web-simple/index.html` |
| react-app | React SPA | `cd ui-prototypes/react-app && npm install && npm start` |
| electron-app | Desktop app | `cd ui-prototypes/electron-app && npm install && npm start` |
