#!/usr/bin/env python3
"""Exercise postfix ! over gopls stdio, including overlay edits and analyses."""
import json
import os
import pathlib
import queue
import subprocess
import tempfile
import threading
import time

ROOT = pathlib.Path(__file__).resolve().parent.parent
SOURCE = '''package main
import ("fmt"; "bangeditor/model")
type Item struct { Field int }
func value() (Item, error) { return Item{7}, nil }
func pair() (int, string, error) { return 7, "ok", nil }
func only() error { return nil }
func prefixNegation() bool { return !
 false }
func run() (out int, err error) {
 item := value()!
 out = item.Field
 n, s := pair()!
 fmt.Println(n, s)
 only()!
 return
}
type Choice union { First int; Second int; Text string }
func choose() (Choice, error) { return Choice.First(7), nil }
func readChoice() (int, error) {
 switch p := choose()!.(variant) {
 case First: return p, nil
 case Second: return p + 1, nil
 case Text: return len(p), nil
 }
}
func modelInt() model.Result[int] { return model.Result[int].Value(7) }
func modelString() model.Result[string] { return model.Result[string].Value("s") }
type ModelAlias = model.Result[int]
func modelAlias() ModelAlias { return ModelAlias.Value(8) }
func First(n int) Choice { return Choice.First(n) }
func localOne() int { type Local union { Value int }; x := Local.Value(1); switch v := x.(variant) { case Value: return v } }
func localTwo() int { type Local union { Value int }; x := Local.Value(2); switch v := x.(variant) { case Value: return v } }
func main() { fmt.Println(run()); fmt.Println(readChoice()) }
'''


MODEL = """package model
type Result[T any] union { Value T; Text string }
func Make[T any](v T) Result[T] { return Result[T].Value(v) }
func (r Result[T]) HasValue() bool {
 switch r.(variant) { case Value: return true; case Text: return false }
}
"""


def apply_edits(text, edits):
    lines = text.splitlines(keepends=True)
    def offset(pos):
        return sum(map(len, lines[:pos["line"]])) + pos["character"]
    for edit in sorted(edits, key=lambda e: offset(e["range"]["start"]), reverse=True):
        lo, hi = (offset(edit["range"][end]) for end in ("start", "end"))
        text = text[:lo] + edit["newText"] + text[hi:]
    return text


def run(binary):
    with tempfile.TemporaryDirectory(prefix="bang-lsp-") as temporary:
        directory = pathlib.Path(temporary)
        path = directory / "main.go"
        path.write_text(SOURCE)
        model = directory / "model" / "model.go"
        model.parent.mkdir()
        model.write_text(MODEL)
        (directory / "go.mod").write_text("module bangeditor\n\ngo 1.26\n")
        uri = path.as_uri()
        env = dict(os.environ, PATH=str(ROOT / "bin") + ":" + os.environ["PATH"],
                   GOROOT=str(ROOT), GOCACHE=os.environ.get("GOCACHE", str(directory / "cache")),
                   GOPLSCACHE=str(directory / "index"), GOWORK="off", GOTOOLCHAIN="local",
                   GOTELEMETRY="off", GOFLAGS="-buildvcs=false")
        with (directory / "server.log").open("w+") as log:
            server = subprocess.Popen([str(binary)], cwd=directory, env=env,
                                      stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=log)
            messages = queue.Queue()
            notifications = []
            next_id = 0

            def reader():
                try:
                    while True:
                        headers = {}
                        while True:
                            line = server.stdout.readline()
                            if not line:
                                return
                            if line == b"\r\n":
                                break
                            key, value = line.decode().split(":", 1)
                            headers[key.lower()] = value.strip()
                        messages.put(json.loads(server.stdout.read(int(headers["content-length"]))))
                except Exception as error:
                    messages.put({"readerError": str(error)})

            threading.Thread(target=reader, daemon=True).start()

            def send(method, params, request_id=None):
                message = {"jsonrpc": "2.0", "method": method, "params": params}
                if request_id is not None:
                    message["id"] = request_id
                write(message)

            def write(message):
                payload = json.dumps(message).encode()
                server.stdin.write(b"Content-Length: " + str(len(payload)).encode() + b"\r\n\r\n" + payload)
                server.stdin.flush()

            def receive(deadline):
                message = messages.get(timeout=max(0.01, deadline - time.monotonic()))
                assert "readerError" not in message, message
                if "method" in message and "id" in message:
                    if message["method"] == "workspace/configuration":
                        result = [{} for _ in message["params"]["items"]]
                    else:
                        result = None
                    write({"jsonrpc": "2.0", "id": message["id"], "result": result})
                else:
                    notifications.append(message)
                return message

            def request(method, params, expected_error=None):
                nonlocal next_id
                next_id += 1
                send(method, params, next_id)
                deadline = time.monotonic() + 60
                while True:
                    response = receive(deadline)
                    if response.get("id") == next_id and "method" not in response:
                        if expected_error is not None:
                            assert expected_error in response.get("error", {}).get("message", ""), response
                            return None
                        assert "error" not in response, response
                        return response.get("result")

            def position(source, needle, offset=0):
                before = source[:source.index(needle) + offset]
                return {"line": before.count("\n"), "character": len(before.rsplit("\n", 1)[-1])}

            def at(needle, offset=0):
                return {"textDocument": {"uri": uri}, "position": position(SOURCE, needle, offset)}

            def diagnostics(version, expected=None):
                deadline = time.monotonic() + 60
                cursor = 0
                while True:
                    for message in notifications[cursor:]:
                        if message.get("method") == "textDocument/publishDiagnostics":
                            params = message["params"]
                            if params.get("uri") != uri or params.get("version") != version:
                                continue
                            errors = [d["message"] for d in params["diagnostics"] if d.get("severity") == 1]
                            if expected is None:
                                assert not errors, errors
                                return
                            if any(expected in error for error in errors):
                                return
                    cursor = len(notifications)
                    receive(deadline)

            try:
                request("initialize", {"processId": None, "rootUri": directory.as_uri(),
                    "capabilities": {}, "initializationOptions": {
                        "staticcheck": True,
                        "analyses": {"nilness": True, "unusedwrite": True},
                        "diagnosticsDelay": "10ms"}})
                send("initialized", {})
                send("textDocument/didOpen", {"textDocument": {
                    "uri": uri, "languageId": "go", "version": 1, "text": SOURCE}})
                diagnostics(1)
                def actions(start, end, kind):
                    return request("textDocument/codeAction", {
                        "textDocument": {"uri": uri}, "range": {"start": start, "end": end},
                        "context": {"diagnostics": [], "only": [kind], "triggerKind": 2}}) or []
                extract_start = position(SOURCE, "item :=")
                extract_end = position(SOURCE, "out = item.Field", len("out = item.Field"))
                assert not actions(extract_start, extract_end, "refactor.extract"), "unsafe extraction offered"
                safe = actions(position(SOURCE, "out = item.Field"), extract_end, "refactor.extract")
                assert any("function" in a["title"].lower() for a in safe), safe
                bang_start = position(SOURCE, "value()!")
                bang_end = position(SOURCE, "value()!", len("value()!"))
                assert not actions(bang_start, bang_end, "refactor.inline"), "unsafe inline offered"
                def reject_fix(fix, start, end):
                    request("workspace/executeCommand", {"command": "gopls.apply_fix", "arguments": [{
                        "Fix": fix, "Location": {"uri": uri, "range": {"start": start, "end": end}},
                        "ResolveEdits": True}]}, "postfix !")
                reject_fix("extract_function", extract_start, extract_end)
                callee_start = position(SOURCE, "fmt.Println(readChoice())", len("fmt.Println("))
                callee_end = position(SOURCE, "fmt.Println(readChoice())", len("fmt.Println(readChoice()"))
                reject_fix("inline_call", callee_start, callee_end)
                hover = request("textDocument/hover", at("item :=", 1))
                assert "Item" in hover["contents"]["value"], hover
                definition = request("textDocument/definition", at("value()!", 1))
                assert definition, definition
                completion = request("textDocument/completion", at("item.Field", len("item.")))
                items = completion.get("items", []) if isinstance(completion, dict) else completion
                assert any(item["label"] == "Field" for item in items), items
                signature = request("textDocument/signatureHelp", at("value()!", len("value(")))
                assert "error" in signature["signatures"][0]["label"], signature
                rename = request("textDocument/rename", dict(at("item :=", 1), newName="renamed"))
                assert "renamed" in json.dumps(rename), rename
                variant_hover = request("textDocument/hover", at("Choice.First", len("Choice.")))
                assert "func First" in variant_hover["contents"]["value"], variant_hover
                variant_rename = request("textDocument/rename", dict(at("Choice.First", len("Choice.")), newName="Primary"))
                assert json.dumps(variant_rename).count('"newText": "Primary"') >= 3, variant_rename
                local_at = at("Local.Value", len("Local."))
                local_rename = request("textDocument/rename", dict(local_at, newName="Number"))
                assert json.dumps(local_rename).count('"newText": "Number"') == 3, local_rename
                local_refs = request("textDocument/references", dict(local_at, context={"includeDeclaration": True}))
                assert len(local_refs) == 3, local_refs
                generic_at = at("model.Result[int].Value", len("model.Result[int]."))
                generic_refs = request("textDocument/references", dict(generic_at, context={"includeDeclaration": True}))
                assert len(generic_refs) >= 6, generic_refs
                generic_rename = request("textDocument/rename", dict(generic_at, newName="Payload"))
                assert json.dumps(generic_rename).count('"newText": "Payload"') >= 6, generic_rename
                request("textDocument/rename", dict(generic_at, newName="Text"), "already has variant")
                request("textDocument/rename", dict(generic_at, newName="value"), "cannot become unexported")
                # Compile all edits, including the generic declaration and its reverse dependencies.
                changes = generic_rename.get("changes", {})
                for change in generic_rename.get("documentChanges", []):
                    changes[change["textDocument"]["uri"]] = change["edits"]
                assert uri in changes and model.as_uri() in changes, changes
                path.write_text(apply_edits(SOURCE, changes[uri]))
                model.write_text(apply_edits(MODEL, changes[model.as_uri()]))
                try:
                    checked = subprocess.run([str(ROOT / "bin/go"), "test", "./..."], cwd=directory,
                        env=env, text=True, capture_output=True, timeout=60)
                    assert checked.returncode == 0, checked.stdout + checked.stderr
                finally:
                    path.write_text(SOURCE)
                    model.write_text(MODEL)
                edits = request("textDocument/formatting", {"textDocument": {"uri": uri},
                    "options": {"tabSize": 8, "insertSpaces": False}})
                formatted = apply_edits(SOURCE, edits)
                assert formatted.count("!") == SOURCE.count("!"), formatted
                for version, text, expected in [
                    (2, SOURCE + "\nfunc invalid() { only()! }\n", "enclosing function"),
                    (3, SOURCE + "\nfunc invalid() error { valuePlain()!; return nil }; func valuePlain() int {return 1}\n", "call to return a final error"),
                    (4, formatted, None),
                ]:
                    send("textDocument/didChange", {"textDocument": {"uri": uri, "version": version},
                        "contentChanges": [{"text": text}]})
                    diagnostics(version, expected)
                request("shutdown", {})
                send("exit", {})
                server.wait(timeout=10)
            finally:
                if server.poll() is None:
                    server.terminate()
                    server.wait(timeout=10)
            log.seek(0)
            output = log.read()
            for forbidden in ("panic:", "internal error", "failed to analyze", "analysis failed"):
                assert forbidden not in output.lower(), output
    print("gopls bang diagnostics, completion, hover, definition, rename, signatures, refactoring guards, formatting, and overlay tests passed")


if __name__ == "__main__":
    import sys
    run(pathlib.Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / "bin/gopls")
