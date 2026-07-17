#!/bin/sh
# Claude Code PostToolUse hook: Edit/Write されたファイルを自動フォーマットする
# stdin に hook 入力(JSON)が渡される。tool_input.file_path を対象にする。
# フォーマッタが見つからない場合は何もせず正常終了する（hook で作業を妨げない）。

file=$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("tool_input",{}).get("file_path",""))' 2>/dev/null)
[ -z "$file" ] || [ ! -f "$file" ] && exit 0

repo_root=$(cd "$(dirname "$0")/.." && pwd)

case "$file" in
  "$repo_root"/backend/*.go)
    if command -v gofmt >/dev/null 2>&1; then
      gofmt -w "$file"
    elif [ -x "$HOME/go/bin/gofmt" ]; then
      "$HOME/go/bin/gofmt" -w "$file"
    fi
    ;;
  "$repo_root"/frontend/src/*.ts|"$repo_root"/frontend/src/*.tsx|"$repo_root"/frontend/e2e/*.ts)
    prettier_js="$repo_root/frontend/node_modules/prettier/bin/prettier.cjs"
    [ -f "$prettier_js" ] || exit 0
    # node が PATH に無い環境（WSL 等）では vscode-server 同梱の node にフォールバックする
    node_bin=$(command -v node 2>/dev/null)
    [ -z "$node_bin" ] && node_bin=$(ls -t "$HOME"/.vscode-server/bin/*/node 2>/dev/null | head -n 1)
    [ -n "$node_bin" ] && "$node_bin" "$prettier_js" --write "$file" >/dev/null 2>&1
    ;;
esac
exit 0
