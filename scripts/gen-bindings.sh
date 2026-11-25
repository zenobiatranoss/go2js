#!/bin/bash
# Writes bindings/dom, the binding of the browser that go2js ships, from the
# declarations of TypeScript (lib.dom.d.ts), cut to what a game or an app that
# draws needs.
#
#   scripts/gen-bindings.sh [path to lib.dom.d.ts]
#
# lib.dom.d.ts comes with the typescript package of version 5 (npm i typescript@5).
set -euo pipefail

cd "$(dirname "$0")/.."

DTS="${1:-node_modules/typescript/lib/lib.dom.d.ts}"
ROOTS=document,window,requestAnimationFrame,cancelAnimationFrame,HTMLCanvasElement,CanvasRenderingContext2D,KeyboardEvent,MouseEvent,TouchEvent,WheelEvent,PointerEvent,HTMLImageElement,HTMLAudioElement,ImageData,Path2D,WebGLRenderingContext,WebGL2RenderingContext,localStorage,performance,navigator,location,fetch,AudioContext

go run ./cmd/go2js bind -global -import dom -dts "$DTS" -name dom -only "$ROOTS" -depth 2 -o bindings/dom dom
