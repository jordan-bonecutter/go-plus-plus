#!/usr/bin/env python3
"""Compare generated ! programs with explicit error checks in several build modes."""
import os
import pathlib
import random
import subprocess
import tempfile

ROOT = pathlib.Path(__file__).resolve().parent.parent
RANDOM = random.Random(81723)


def fixture():
    functions = []
    names = []
    for case in range(96):
        next_call = 0
        named = case % 2 == 0
        boolean = case % 3 == 0

        def integer(depth):
            nonlocal next_call
            if depth == 0 or RANDOM.randrange(3) == 0:
                call = ("call", next_call)
                next_call += 1
                return call
            return (RANDOM.choice(["+", "-", "*"]), integer(depth - 1), integer(depth - 1))

        def condition(depth):
            if depth == 0:
                return (">", integer(0), ("constant", RANDOM.randrange(2, 6)))
            return (RANDOM.choice(["&&", "||"]), condition(depth - 1), condition(depth - 1))

        tree = condition(2) if boolean else integer(2)

        def bang(node):
            op, *args = node
            if op == "call":
                return f"stage(mask, {args[0]}, &n)!"
            if op == "constant":
                return str(args[0])
            return f"({bang(args[0])} {op} {bang(args[1])})"

        statements = []
        next_var = 0

        def reference(node, indent=""):
            nonlocal next_var
            op, *args = node
            if op == "constant":
                return str(args[0])
            variable = f"v{next_var}"
            next_var += 1
            if op == "call":
                statements.append(f"{indent}{variable}, e{variable} := stage(mask, {args[0]}, &n)")
                result = "n" if named else "0"
                statements.append(f"{indent}if e{variable} != nil {{ return {result}, e{variable} }}")
            else:
                lhs = reference(args[0], indent)
                if op in ("&&", "||"):
                    statements.append(f"{indent}{variable} := {lhs}")
                    test = variable if op == "&&" else "!" + variable
                    statements.append(f"{indent}if {test} {{")
                    rhs = reference(args[1], indent + " ")
                    statements.append(f"{indent} {variable} = {rhs}")
                    statements.append(f"{indent}}}")
                else:
                    rhs = reference(args[1], indent)
                    statements.append(f"{indent}{variable} := {lhs} {op} {rhs}")
            return variable

        result = reference(tree)
        signature = "(n int, err error)" if named else "(int, error)"
        init = "n = 13; err = fault(-1)" if named else "n := 13"
        end = "return" if named else "return n, nil"
        for suffix, expression, prelude in [("bang", bang(tree), ""), ("reference", result, "\n".join(statements))]:
            action = f"if {expression} {{ n=71 }} else {{ n=83 }}" if boolean else f"n = {expression}"
            # Half the programs return through an iterator callback.
            if case % 4 in (2, 3):
                body = "for range seq {\n" + prelude + "\n" + action + "; break\n}"
            else:
                body = prelude + "\n" + action
            functions.append(f"func case{case}_{suffix}(mask int) {signature} {{\n{init}\ndefer func(){{ seen=append(seen,1000+n); n+=3 }}()\n{body}\n{end}\n}}")
        names.append(f'{{"case{case}",case{case}_bang,case{case}_reference,{1 << next_call}}}')
    return '''package main
import("fmt";"reflect")
type fault int
func(e fault)Error()string{return fmt.Sprint(int(e))}
var seen []int
func stage(mask,id int,n *int)(int,error){seen=append(seen,id);*n+=id+1;if mask&(1<<id)!=0{return 99,fault(id)};return id+2,nil}
func seq(y func(int)bool){for i:=0;i<2;i++{if !y(i){return}}}
type result struct{N int; E error; Trace []int; Panic any}
func evaluate(fn func(int)(int,error),mask int)(r result){seen=nil;defer func(){r.Panic=recover();r.Trace=append([]int(nil),seen...)}();r.N,r.E=fn(mask);return}
''' + "\n".join(functions) + '''
func main(){cases:=[]struct{name string;bang,reference func(int)(int,error);masks int}{
''' + ",\n".join(names) + ''',
};count:=0;for _,c:=range cases{for mask:=0;mask<c.masks;mask++{got,want:=evaluate(c.bang,mask),evaluate(c.reference,mask);if !reflect.DeepEqual(got,want){panic(fmt.Sprintf("%s mask %d: got %#v, want %#v",c.name,mask,got,want))};count++}};fmt.Printf("%d differential cases passed\\n",count)}
'''


def main():
    with tempfile.TemporaryDirectory(prefix="bang-differential-") as temporary:
        source = pathlib.Path(temporary) / "main.go"
        source.write_text(fixture())
        env = dict(os.environ, GOROOT=str(ROOT), GOTOOLCHAIN="local", GOWORK="off")
        for flags in ([], ["-gcflags=-l"], ["-gcflags=-N -l"], ["-race"]):
            command = [str(ROOT / "bin/go"), "run", *flags, str(source)]
            result = subprocess.run(command, env=env, text=True, capture_output=True, timeout=180)
            if result.returncode:
                # Keep the exact reproducer if compilation or behavior differs.
                retained = ROOT / "test" / "bang_differential_failure.go"
                retained.write_text(source.read_text())
                raise AssertionError(f"{command}:\n{result.stdout}{result.stderr}\nReproducer: {retained}")
            print(" ".join(flags) or "optimized", result.stdout.strip(), flush=True)


if __name__ == "__main__":
    main()
