#!/usr/bin/env python3
# Exercise the union-aware editor build over the Language Server Protocol.
import os,subprocess,json,pathlib,threading,queue,tempfile
root=str(pathlib.Path(__file__).resolve().parent.parent)
workspace=tempfile.TemporaryDirectory(prefix="union-gopls-test-")
directory=pathlib.Path(workspace.name)
(directory/'go.mod').write_text('module unioneditor\n\ngo 1.26\n')
source="""package main
import "fmt"
type Foo union { Bar int; Baz string }
type Result[T any] union { Ok T; Err string }
func(r Result[T]) Get() T {
 switch data := r.(variant) { case Ok: return data; case Err: panic(data) }
}
func main() { fmt.Println(Result[int].Ok(42).Get()) }
"""
(directory/'main.go').write_text(source)
uri=(directory/'main.go').as_uri()
env=dict(os.environ,PATH=root+'/bin:'+os.environ['PATH'],GOROOT=root,GOCACHE=os.environ.get('GOCACHE',str(directory/'build-cache')),GOPLSCACHE=str(directory/'index-cache'),GOWORK='off',GOTOOLCHAIN='local',GOFLAGS='-buildvcs=false',GOTELEMETRY='off')
p=subprocess.Popen([root+'/bin/gopls'],stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.DEVNULL,env=env,cwd=str(directory))
responses=queue.Queue()
def reader():
 while True:
  header={}
  while True:
   line=p.stdout.readline()
   if not line:return
   if line==b'\r\n':break
   k,v=line.decode().split(':',1);header[k.lower()]=v.strip()
  responses.put(json.loads(p.stdout.read(int(header['content-length']))))
threading.Thread(target=reader,daemon=True).start()
def send(method,params,rid=None):
 obj={'jsonrpc':'2.0','method':method,'params':params}
 if rid is not None:obj['id']=rid
 payload=json.dumps(obj).encode();p.stdin.write(b'Content-Length: '+str(len(payload)).encode()+b'\r\n\r\n'+payload);p.stdin.flush()
 if rid is not None:
  while True:
   response=responses.get(timeout=60)
   if response.get('id')==rid:return response

def pos(offset):return {'line':source[:offset].count('\n'),'character':len(source[:offset].rsplit('\n',1)[-1])}
try:
 send('initialize',{'processId':None,'rootUri':directory.as_uri(),'capabilities':{}},1)
 send('initialized',{})
 send('textDocument/didOpen',{'textDocument':{'uri':uri,'languageId':'go','version':1,'text':source}})
 at=source.index('Result[int].Ok')+len('Result[int].')
 r=send('textDocument/hover',{'textDocument':{'uri':uri},'position':pos(at+1)},2)
 assert 'func Ok(value int) Result[int]' in r['result']['contents']['value'],r
 r=send('textDocument/completion',{'textDocument':{'uri':uri},'position':pos(at),'context':{'triggerKind':2,'triggerCharacter':'.'}},3)
 result=r.get('result') or {};items=result.get('items',[]) if isinstance(result,dict) else result
 labels=[i.get('label') for i in items]
 assert {'Ok','Err'}.issubset(labels),labels
 assert not any('union·' in label for label in labels),labels
 send('shutdown',{},4);send('exit',{})
finally:
 p.terminate()
 p.wait(timeout=10)

for text,want in [
 (source+'\nfunc missing(x Foo){switch x.(variant){case Bar:}}\n','non-exhaustive variant switch: missing Baz'),
 (source.replace('Result[T any]','Result[T]'),'missing constraint for type parameter T (use [T any])'),
]:
 (directory/'main.go').write_text(text)
 r=subprocess.run([root+'/bin/gopls','check','-severity=error','main.go'],cwd=str(directory),env=env,text=True,capture_output=True,timeout=60)
 assert want in r.stdout+r.stderr,r.stdout+r.stderr
workspace.cleanup()
print('gopls union smoke tests passed')
