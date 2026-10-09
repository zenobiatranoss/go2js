package javascript

var templateFuncs = map[string]string{
	"New":  "go2jsTemplateNew",
	"Must": "go2jsTemplateMust",
}

var templateTypes = map[string]string{
	"FuncMap":  "go2jsTemplateFuncMap",
	"Template": "go2jsTemplateNew",
	"HTML":     "go2jsTemplate",
	"JS":       "go2jsTemplate",
	"URL":      "go2jsTemplate",
	"CSS":      "go2jsTemplate",
	"HTMLAttr": "go2jsTemplate",
}

func init() {
	extendedStdlibFuncs()
	moreStdlibFuncs()
	functions := make(map[string]string, len(templateFuncs)+len(templateTypes))
	for name, value := range templateFuncs {
		functions[name] = value
	}
	for name := range templateTypes {
		functions[name] = ""
	}
	supported := funcSet(functions)
	stdlibFuncMaps["text/template"] = functions
	stdlibFuncMaps["html/template"] = functions
	supportedStdlibPackages["text/template"] = supported
	supportedStdlibPackages["html/template"] = supported
	for name, value := range templateTypes {
		packageTypes["template."+name] = value
	}
	for _, alias := range []string{"htmltemplate", "texttemplate"} {
		supportedStdlibPackages[alias] = supported
	}
}

func templateRuntimeSource() string {
	return `
function go2jsTemplateFuncMap() { return {}; }
function go2jsTemplate(name) {
  var t = { __go2js_template:true, name:name==null?"":String(name), tree:null, funcs:{}, defs:{} };
  t.Parse = function(text){ return [t, go2jsTemplateParse(t,text)]; };
  t.Execute = function(w,d){ return go2jsTemplateExecute(t,w,d); };
  t.ExecuteTemplate = function(n,w,d){ return go2jsTemplateExecuteTemplate(t,n,w,d); };
  t.Funcs = function(f){ if (f!==null&&f!==undefined){ for (var k of Object.keys(f)) t.funcs[k]=f[k]; } return t; };
  return t;
}
function go2jsTemplateNew(name){ return go2jsTemplate(name); }
function go2jsTemplateMust(r){ var v=Array.isArray(r)?r:[r,null]; if (v[1]!==null&&v[1]!==undefined) throw v[1]; return v[0]; }
function go2jsTemplateLineOf(src,off){ if (off<0) off=0; if (off>src.length) off=src.length; var n=1; for (var i=0;i<off&&i<src.length;i++) if (src[i]==='\n') n++; return n; }
function go2jsTemplateLineCol(src,off){ if (off<0) off=0; if (off>src.length) off=src.length; var s=0; for (var i=off;i>=0;i--){ if (src[i]==='\n'){ s=i+1; break; } if (i===0){ s=0; break; } } return off-s; }
function go2jsTemplateParseFail(msg,off){ throw {go2jsTemplateParseMsg:msg, go2jsTemplateParseOffset:typeof off==='number'?off:0}; }
function go2jsTemplateExecFail(msg,kind,off,token){ return {msg:msg,kind:kind||'',offset:off==null?null:off,token:token||''}; }
function go2jsTemplateParse(template,text){ var src=String(text); try{ var p=go2jsTemplateParseNodes(src,null,0); template.tree=p.nodes; template.defs={}; template.src=src; go2jsTemplateRegisterDefs(template.tree,template); return null; }catch(e){ if (e&&e.go2jsTemplateParseMsg){ var line=go2jsTemplateLineOf(src,e.go2jsTemplateParseOffset||0); return new Error('template: '+(template.name||'')+':'+line+': '+e.go2jsTemplateParseMsg); } return e; } }
function go2jsTemplateRegisterDefs(nodes,t){ if (!nodes) return; for (var n of nodes){ if (n.t==='define'){ var nm=go2jsTemplateGetDefineName(n); if (nm){ t.defs=t.defs||{}; t.defs[nm]={name:nm,tree:n.body,funcs:t.funcs,defs:t.defs}; } } } }
function go2jsTemplateWrite(w,t){ if (!w) return; var ws=go2jsWriterMethod(w,'WriteString'); if (ws!==null){ ws(t); return; } var wr=go2jsWriterMethod(w,'Write'); if (wr===null){ throw new TypeError('template: writer does not implement io.Writer'); } wr(go2jsStringToBytes(t)); }

function go2jsTemplateCheckKw(body, ctxWord) {
  var stages = go2jsTemplateSplitStages(body);
  for (var i=0;i<stages.length;i++){
    var toks = go2jsTemplateTokenize(stages[i].text);
    for (var j=0;j<toks.length;j++){
      if (toks[j].k==="word" && (toks[j].v==="end"||toks[j].v==="else")){
        if (ctxWord) return { kw: toks[j].v, loc: "in " + ctxWord };
        if (j===0) return { kw: toks[j].v, loc: "in command" };
        return { kw: toks[j].v, loc: "in operand" };
      }
    }
  }
  return null;
}
function go2jsTemplateAfterName(action) {
  var a = action.trim();
  var m = a.match(/^"([^"]*)"|'([^']*)'/);
  if (m) return a.slice(m[0].length).trim();
  var sp = a.search(/\s/);
  if (sp === -1) return "";
  return a.slice(sp).trim();
}
function go2jsTemplateParseNodes(text, stops, offset) {
  var nodes = []; var index = offset; var stopSet = stops || [];
  for (;;) {
    var open = text.indexOf("{{", index);
    if (open === -1) { if (index < text.length) nodes.push({t:"text",v:text.slice(index)}); return {nodes, stop:"", action:"", index:text.length}; }
    var literal = text.slice(index,open); var cursor = open+2; var trimLeft=false;
    if (text[cursor]==='-' && cursor+1<text.length && (text[cursor+1]===' '||text[cursor+1]==='\t'||text[cursor+1]==='\n'||text[cursor+1]==='\r'||text[cursor+1]==='}')) { trimLeft=true; cursor++; }
    var close = text.indexOf("}}", cursor); if (close === -1) go2jsTemplateParseFail("unexpected EOF", text.length);
    var bodyRaw = text.slice(cursor,close); var trimRight=false;
    if (bodyRaw.length>0 && bodyRaw[bodyRaw.length-1]==='-') { trimRight=true; bodyRaw = bodyRaw.slice(0,-1); }
    var bodyLead = bodyRaw.length - bodyRaw.replace(/^\s+/, "").length;
    var bodyOff = cursor + bodyLead;
    var body = bodyRaw.trim();
    if (trimLeft) literal = literal.replace(/[\s\r\n]+$/,"");
    if (literal!=="") nodes.push({t:"text",v:literal});
    index = close+2;
    if (trimRight) { while (index<text.length && /\s/.test(text[index])) index++; }
    if (body==="") continue;
    var mkw = body.match(/^(end|else|if|with|range|define|block|template)(?=[\s(]|$)/);
    if (mkw) {
      var kw = mkw[1];
      if (stopSet.indexOf(kw)!==-1) { return {nodes, stop:kw, action:body.slice(kw.length).trim(), index}; }
      if (kw==="end") go2jsTemplateParseFail("unexpected {{end}}", open);
      if (kw==="else") { if (stopSet.indexOf("else")!==-1||stopSet.length>0) return {nodes,stop:"else",action:body.slice(4).trim(),index}; go2jsTemplateParseFail("unexpected else", open); }
      var restRaw = body.slice(kw.length);
      var lead2 = restRaw.length - restRaw.replace(/^\s+/, "").length;
      var kwActOff = bodyOff + kw.length + lead2;
      if (kw==="define"||kw==="block") {
        var rest = restRaw.trim(); if (rest==="") go2jsTemplateParseFail("missing value for "+kw+" clause", open);
        if (kw==="block" && go2jsTemplateAfterName(rest)==="") go2jsTemplateParseFail("missing value for block clause", open);
        var blk = go2jsTemplateParseBlock(text,index,kw,rest,open,kwActOff); nodes.push(blk.node); index=blk.index; continue;
      }
      if (kw==="template") {
        var rest = restRaw.trim(); if (rest==="") go2jsTemplateParseFail("missing value for template clause", open);
        nodes.push({t:"template", action:rest, pos:kwActOff, actOff:kwActOff}); continue;
      }
      if (kw==="if"||kw==="with"||kw==="range") {
        var rest = restRaw.trim(); if (rest==="") go2jsTemplateParseFail("missing value for "+kw, open);
        var bad = go2jsTemplateCheckKw(rest, kw); if (bad) go2jsTemplateParseFail("unexpected <"+bad.kw+"> "+bad.loc, open);
        var blk = go2jsTemplateParseBlock(text,index,kw,rest,open,kwActOff); nodes.push(blk.node); index=blk.index; continue;
      }
    }
    var kbad = go2jsTemplateCheckKw(body, null); if (kbad) go2jsTemplateParseFail("unexpected <"+kbad.kw+"> "+kbad.loc, open);
    nodes.push({t:"action", action:body, pos:open, actOff:bodyOff});
  }
}
function go2jsTemplateParseBlock(text, index, keyword, action, actionOff, actOff) {
  var head = go2jsTemplateParseNodes(text, ["else","end"], index);
  var node = {t:keyword, action:action, body:head.nodes, alt:[], stop:"", pos:actionOff!==undefined?actionOff:index, actOff:actOff!==undefined?actOff:(actionOff!==undefined?actionOff:index), namePos:actionOff!==undefined?actionOff:index};
  if (head.stop === "end") return {node, index:head.index};
  if (head.stop === "else") {
    var rest = head.action.trim();
    if (rest === "") {
      var branch = go2jsTemplateParseNodes(text, ["else","end"], head.index);
      node.alt = branch.nodes;
      if (branch.stop === "else") go2jsTemplateParseFail("expected end; found {{else}}", head.index-2);
      if (branch.stop === "end") { node.stop="end"; return {node,index:branch.index}; }
      return {node,index:branch.index};
    }
    var first = rest.split(/[\s(]/)[0];
    if (first === "if") {
      var nested = go2jsTemplateParseBlock(text, head.index, "if", rest.slice(2).trim(), head.index-2);
      node.alt = [nested.node];
      return {node, index:nested.index};
    }
    go2jsTemplateParseFail('unexpected <'+first+'> in else', head.index-2);
  }
  var branch = go2jsTemplateParseNodes(text, ["else","end"], head.index);
  node.alt = branch.nodes;
  if (branch.stop === "end") { node.stop="end"; return {node,index:branch.index}; }
  if (branch.stop === "else") go2jsTemplateParseFail("expected end; found {{else}}", branch.index-2);
  return {node,index:branch.index};
}
function go2jsTemplateGetDefineName(node){
  if (!node || node.t !== 'define') return '';
  var a = node.action || '';
  var m = a.match(/^"([^"]+)"|'([^']+)'|([^\s()]+)$/);
  if (!m) return '';
  return m[1]||m[2]||m[3]||'';
}
function go2jsTemplateFindTemplate(template, name) {
  if (template === null || template === undefined) return null;
  if (template.defs && template.defs[name]) return template.defs[name];
  if (template.tree && name === (template.name || '')) return template;
  return null;
}
function go2jsTemplateExecute(template, writer, data) {
  var ctx = { name: template && template.name ? template.name : "", rootName: template && template.name ? template.name : "", source: (template && template.src) ? template.src : "", defs: {}, node: null, stage: null, out: "" };
  try {
    var root = data; var vars = { $: data };
    var tree = template.tree || [];
    template.defs = template.defs || {};
    try { if (typeof template._root === "undefined") template._root = template; } catch(e) {}
    go2jsTemplateRegisterDefs(tree, template);
    ctx.defs = template.defs; ctx.root = template;
    var res = go2jsTemplateRenderNodes(tree, data, root, template.funcs, vars, ctx);
    go2jsTemplateWrite(writer, ctx.out + (res||""));
    return null;
  } catch (err) {
    go2jsTemplateWrite(writer, ctx.out);
    return go2jsTemplateExecError(err, ctx);
  }
}
function go2jsTemplateExecError(err, ctx) {
  if (err === null || err === undefined) return null;
  if (typeof err === 'string') return new Error(err);
  if (err instanceof Error) return err;
  if (typeof err !== 'object') return new Error(String(err));
  var kind = err.kind || '';
  var ctxName = (ctx && ctx.name) ? ctx.name : '';
  var rootName = (ctx && ctx.rootName !== undefined) ? ctx.rootName : ctxName;
  var src = (ctx && ctx.source !== undefined) ? String(ctx.source) : '';
  var msg = err.msg || 'error';
  if (kind === 'field') {
    var off = typeof err.offset === 'number' ? err.offset : (typeof err.tokenAbs === 'number' ? err.tokenAbs : 0);
    var line = go2jsTemplateLineOf(src, off); var col = go2jsTemplateLineCol(src, off);
    var display = err.token || (ctx && ctx.stage && ctx.stage.head) || '';
    return new Error('template: ' + rootName + ':' + line + ':' + col + ': executing "' + ctxName + '" at <' + display + '>: ' + msg);
  }
  if (kind === 'token' || kind === 'call') {
    var off = typeof err.offset === 'number' ? err.offset : ((ctx && ctx.stage && ctx.stage.headAbs !== undefined) ? ctx.stage.headAbs : 0);
    var line = go2jsTemplateLineOf(src, off); var col = go2jsTemplateLineCol(src, off);
    var display = err.display !== undefined ? err.display : ((ctx && ctx.stage && ctx.stage.text) ? ctx.stage.text : ((ctx && ctx.stage && ctx.stage.head) || ''));
    return new Error('template: ' + rootName + ':' + line + ':' + col + ': executing "' + ctxName + '" at <' + display + '>: ' + msg);
  }
  if (kind === 'node' || kind === 'action') {
    var node = ctx && ctx.node ? ctx.node : null;
    var off = node && typeof node.pos === 'number' ? node.pos : 0;
    var line = go2jsTemplateLineOf(src, off); var col = go2jsTemplateLineCol(src, off);
    var disp = (node && node.actionDisp) ? node.actionDisp : ((node && node.action) ? ('{{' + node.action + '}}') : '');
    return new Error('template: ' + rootName + ':' + line + ':' + col + ': executing "' + ctxName + '" at <' + disp + '>: ' + msg);
  }
  var line = go2jsTemplateLineOf(src, 0);
  return new Error('template: ' + ctxName + ':' + line + ': ' + msg);
}
function go2jsTemplateParseRangeFull(action) {
  var a = action.trim();
  var colon = a.indexOf(":=");
  if (colon === -1) return { pipeline: a };
  var left = a.slice(0, colon).trim();
  var right = a.slice(colon+2).trim();
  var parts = left.split(/\s*,\s*/);
  if (parts.length === 2) return { key: parts[0].trim(), val: parts[1].trim(), pipeline: right };
  if (parts.length === 1) return { key: parts[0].trim(), pipeline: right };
  return { pipeline: right };
}
function go2jsTemplateParseRange(action) {
  var a = action.trim();
  var colon = a.indexOf(":=");
  if (colon === -1) return {};
  var left = a.slice(0, colon).trim();
  var right = a.slice(colon+2).trim();
  var parts = left.split(/\s*,\s*/);
  if (parts.length === 2) return { key: parts[0].trim(), val: parts[1].trim() };
  if (parts.length === 1) return { key: parts[0].trim() };
  return {};
}
function go2jsTemplateVars(parent, extra) {
  var n = Object.create(parent || {});
  if (extra) { for (var k in extra) if (Object.prototype.hasOwnProperty.call(extra, k)) n[k] = extra[k]; }
  return n;
}
function go2jsTemplateRangeItems(r) {
  var out = [];
  if (typeof go2jsInterfaceValue === "function") r = go2jsInterfaceValue(r);
  if (r === null || r === undefined) return out;
  if (typeof r === "number") {
    for (var i = 0; i < r; i++) { out.push({ key: "i", keyval: i, value: i }); }
    return out;
  }
  if (Array.isArray(r)) {
    for (var i = 0; i < r.length; i++) { out.push({ key: "i", keyval: i, value: r[i] }); }
    return out;
  }
  if (r instanceof go2jsNativeMap) {
    var keys = Array.from(r.keys());
    try { keys.sort(function(a,b){ if (typeof a==="number"&&typeof b==="number") return a-b; var sa=String(a), sb=String(b); if (sa<sb)return -1; if (sa>sb)return 1; return 0; }); } catch(e) {}
    for (var j=0; j<keys.length; j++) { var k=keys[j]; out.push({ key: "k", keyval: k, value: r.get(k) }); }
    return out;
  }
  if (typeof r === "string") { var s=r; for (var i=0;i<s.length;i++){ out.push({ key: "i", keyval: i, value: s[i] }); } return out; }
  if (typeof r === "object") { var keys=Object.keys(r); try{keys.sort();}catch(e){} for (var j=0;j<keys.length;j++){ out.push({ key:"k", keyval:keys[j], value:r[keys[j]]}); } return out; }
  return out;
}
function go2jsTemplateGetTemplateName(action) {
  var a = action.trim();
  var m = a.match(/^"([^"]*)"|'([^']*)'|([^\s]+)(?:[\s(]|$)/);
  if (m) return m[1]||m[2]||m[3]||'';
  return a;
}
function go2jsTemplateEvalTemplateData(action, dot, root, funcs, vars, ctx) {
  var a = action.trim();
  var rest = go2jsTemplateTemplatePipeline(action);
  if (rest === null) return null;
  return go2jsTemplateEval(rest, dot, root, funcs, vars, ctx);
}
function go2jsTemplateTemplatePipeline(action) {
  var a = action.trim();
  var m = a.match(/^"([^"]*)"|'([^']*)'/);
  if (m) { var rest = a.slice(m[0].length).trim(); return rest === "" ? null : rest; }
  var sp = a.indexOf(" ");
  if (sp === -1) return null;
  return a.slice(sp+1).trim();
}
function go2jsTemplateEmit(ctx, s) { if (s !== undefined && s !== null && s !== "") ctx.out += s; }
function go2jsTemplateRenderNodes(nodes, dot, root, funcs, vars, ctx) {
  if (!nodes || nodes.length === 0) return;
  for (var i = 0; i < nodes.length; i++) {
    var node = nodes[i];
    ctx.node = node;
    if (node.t === "text") { go2jsTemplateEmit(ctx, node.v); continue; }
    if (node.t === "action") {
      ctx.actionBase = node.actOff !== undefined ? node.actOff : 0;
      var v = go2jsTemplateEvalAction(node.action, dot, root, funcs, vars, ctx);
      go2jsTemplateEmit(ctx, go2jsTemplateActionString(v));
      continue;
    }
    if (node.t === "if") {
      ctx.actionBase = node.actOff !== undefined ? node.actOff : 0;
      var cond = go2jsTemplateEvalCond(node.action, dot, root, funcs, vars, ctx);
      if (go2jsTemplateTruth(cond)) {
        go2jsTemplateRenderNodes(node.body || [], dot, root, funcs, vars, ctx);
      } else if (node.alt && node.alt.length > 0) {
        go2jsTemplateRenderNodes(node.alt, dot, root, funcs, vars, ctx);
      }
      continue;
    }
    if (node.t === "with") {
      ctx.actionBase = node.actOff !== undefined ? node.actOff : 0;
      var val = go2jsTemplateEvalCond(node.action, dot, root, funcs, vars, ctx);
      if (go2jsTemplateTruth(val)) {
        var vvars = go2jsTemplateVars(vars, { $: val });
        go2jsTemplateRenderNodes(node.body || [], val, root, funcs, vvars, ctx);
      } else if (node.alt && node.alt.length > 0) {
        go2jsTemplateRenderNodes(node.alt, dot, root, funcs, vars, ctx);
      }
      continue;
    }
    if (node.t === "range") {
      ctx.actionBase = node.actOff !== undefined ? node.actOff : 0;
      var rinfo = go2jsTemplateParseRangeFull(node.action);
      var r = go2jsTemplateEval(rinfo.pipeline || node.action, dot, root, funcs, vars, ctx);
      var rk = go2jsInterfaceValue ? go2jsInterfaceValue(r) : r;
      if (!(rk === null || rk === undefined || (typeof rk === "number" && Math.floor(rk) === rk) || Array.isArray(rk) || rk instanceof go2jsNativeMap)) {
        var re = go2jsTemplateExecFail("range can't iterate over " + go2jsTemplatePrintValue(rk), "action", 0);
        ctx.node = { pos: node.actOff, actionDisp: node.action };
        throw re;
      }
      var items = go2jsTemplateRangeItems(r);
      if (items.length === 0) {
        if (node.alt && node.alt.length > 0) {
          go2jsTemplateRenderNodes(node.alt, dot, root, funcs, vars, ctx);
        }
        continue;
      }
      for (var j = 0; j < items.length; j++) {
        var it = items[j];
        var ndot = it.value !== undefined ? it.value : dot;
        var nv = go2jsTemplateVars(vars, { $: root });
        if (rinfo.key) nv[rinfo.key] = it.keyval;
        if (rinfo.val) nv[rinfo.val] = it.value;
        go2jsTemplateRenderNodes(node.body || [], ndot, root, funcs, nv, ctx);
      }
      continue;
    }
    if (node.t === "define") {
      continue;
    }
    if (node.t === "block") {
      var bpipe = go2jsTemplateTemplatePipeline(node.action);
      var bval = bpipe === null ? dot : go2jsTemplateEval(bpipe, dot, root, funcs, vars, ctx);
      var bdot = bval !== null && bval !== undefined ? bval : dot;
      var name = go2jsTemplateGetTemplateName(node.action) || node.name || "";
      var oldName = ctx.name;
      if (ctx && ctx.root && ctx.root.defs && ctx.root.defs[name]) {
        var tdef = ctx.root.defs[name];
        ctx.name = name;
        go2jsTemplateRenderNodes(tdef.tree || tdef.body || [], bdot, root, funcs, { $: bdot }, ctx);
      } else {
        ctx.name = name || ctx.name;
        go2jsTemplateRenderNodes(node.body || [], bdot, root, funcs, { $: bdot }, ctx);
      }
      ctx.name = oldName;
      continue;
    }
    if (node.t === "template") {
      var tn = go2jsTemplateGetTemplateName(node.action);
      var tdef = null;
      if (ctx && ctx.defs) tdef = ctx.defs[tn];
      if (!tdef && ctx && ctx.root && ctx.root.defs) tdef = ctx.root.defs[tn];
      if (!tdef) {
        ctx.node = { pos: node.pos, actionDisp: "{{template " + node.action + "}}" };
        throw go2jsTemplateExecFail('template "' + tn + '" not defined', "action", node.pos);
      }
      var tval = go2jsTemplateEvalTemplateData(node.action, dot, root, funcs, vars, ctx);
      var tdot = tval;
      var oldName3 = ctx.name;
      ctx.name = tn;
      go2jsTemplateRenderNodes(tdef.tree || tdef.body || [], tdot, root, funcs, { $: tdot }, ctx);
      ctx.name = oldName3;
      continue;
    }
  }
}
function go2jsTemplateEvalCond(action, dot, root, funcs, vars, ctx) {
  var a = action.trim();
  var idx = a.indexOf(":=");
  if (idx !== -1) {
    var nm = a.slice(0, idx).trim();
    var pipe = a.slice(idx + 2).trim();
    var v = go2jsTemplateEval(pipe, dot, root, funcs, vars, ctx);
    if (nm.startsWith("$")) vars[nm] = v;
    return v;
  }
  return go2jsTemplateEval(a, dot, root, funcs, vars, ctx);
}
function go2jsTemplateEvalAction(action, dot, root, funcs, vars, ctx) {
  var a = action;
  var assign = a.indexOf(":=");
  var equals = a.indexOf(" = ");
  if (assign !== -1) {
    var name = a.slice(0, assign).trim();
    if (name.startsWith("$")) {
      vars[name] = go2jsTemplateEval(a.slice(assign+2).trim(), dot, root, funcs, vars, ctx);
      return "";
    }
  }
  if (equals !== -1 && a.indexOf("==") === -1 && a.indexOf("!=") === -1) {
    var name = a.slice(0, equals).trim();
    if (name.startsWith("$")) {
      vars[name] = go2jsTemplateEval(a.slice(equals+3).trim(), dot, root, funcs, vars, ctx);
      return "";
    }
  }
  return go2jsTemplateEval(a, dot, root, funcs, vars, ctx);
}
function go2jsTemplateSplit(pipeline, sep) {
  var parts=[]; var d=0, q="", cur="";
  for (var ch of pipeline) {
    if (q!==""){ cur+=ch; if (ch===q) q=""; continue; }
    if (ch==='"'||ch==="'"){ q=ch; cur+=ch; continue; }
    if (ch==="(") d++; else if (ch===")") d--;
    if (ch===sep && d===0){ parts.push(cur); cur=""; continue; }
    cur+=ch;
  }
  parts.push(cur);
  return parts;
}
function go2jsTemplateEval(pipeline, dot, root, funcs, vars, ctx) {
  var stages = go2jsTemplateSplitStages(pipeline);
  if (stages.length===0) return "";
  var value = go2jsTemplateEvalCommand(stages[0].text, dot, root, funcs, vars, ctx, false, null, stages[0].off);
  for (var i=1;i<stages.length;i++) {
    var fn = go2jsTemplateEvalCommand(stages[i].text, dot, root, funcs, vars, ctx, true, value, stages[i].off);
    if (typeof fn !== "function") throw go2jsTemplateExecFail("non executable command in pipeline stage "+(i+1), "token", (ctx.stage&&ctx.stage.headAbs)||0);
    var savedStage = ctx.stage;
    if (fn.__go2jsStage) ctx.stage = fn.__go2jsStage;
    try { value = go2jsCallNow(fn,null,[value]); } finally { ctx.stage = savedStage; }
  }
  return value;
}
function go2jsTemplateSplitStages(pipeline) {
  var parts=[]; var d=0, q="", cur="", curOff=0, started=false;
  for (var i=0;i<pipeline.length;i++) {
    var ch=pipeline[i];
    if (q!==""){ cur+=ch; if (ch===q) q=""; continue; }
    if (ch==='"'||ch==="'"){ if(!started){started=true;curOff=i;} q=ch; cur+=ch; continue; }
    if (ch==="(") d++; else if (ch===")") d--;
    if (ch==="|" && d===0){ parts.push({text:cur.trim(), off:curOff}); cur=""; started=false; continue; }
    if (!started && !/\s/.test(ch)){ started=true; curOff=i; }
    cur+=ch;
  }
  parts.push({text:cur.trim(), off:curOff});
  return parts.filter(function(p){ return p.text !== ""; });
}
function go2jsTemplateTokenize(command) {
  var tokens=[]; var i=0;
  while (i<command.length) {
    var ch=command[i];
    if (/\s/.test(ch)){ i++; continue; }
    if (ch==="("){ var d=1,c=i+1,b=""; while(c<command.length&&d>0){ if (command[c]==="(")d++; else if (command[c]===")"){ d--; if (d===0) break; } b+=command[c]; c++; } tokens.push({k:"group",v:b,o:i,r:"("+b+")"}); i=c+1; continue; }
    if (ch==='"'||ch==="'"){ var c=i+1,b=""; while(c<command.length&&command[c]!==ch){ if (ch==='"'&&command[c]==='\\') c++; b+=command[c]; c++; } tokens.push({k:"string",v:b,o:i,r:ch+b+ch}); i=c+1; continue; }
    var w=""; while(i<command.length&&!/\s/.test(command[i])&&command[i]!=="("){ w+=command[i]; i++; }
    if (w!=="") tokens.push({k:"word",v:w,o:i-w.length,r:w});
  }
  return tokens;
}
function go2jsTemplateEvalToken(token, dot, root, funcs, vars, ctx) {
  if (token.k === "group") return go2jsTemplateEval(token.v, dot, root, funcs, vars, ctx);
  if (token.k === "string") return token.v;
  return go2jsTemplateResolve(token, [], dot, root, funcs, vars, ctx);
}
function go2jsTemplateEvalCommand(command, dot, root, funcs, vars, ctx, hasPipe, pipeVal, stageOff) {
  var tokens = go2jsTemplateTokenize(command.trim());
  if (tokens.length === 0) return "";
  var prev = ctx.stage;
  var base = (typeof stageOff === "number" ? stageOff : 0) + ((ctx && typeof ctx.actionBase === "number") ? ctx.actionBase : 0);
  ctx.stage = { text: command.trim(), head: tokens[0].r || tokens[0].v, headAbs: base + (tokens[0].o !== undefined ? tokens[0].o : 0), base: base };
  var args = [];
  for (var i=1; i<tokens.length; i++) {
    args.push(go2jsTemplateEvalToken(tokens[i], dot, root, funcs, vars, ctx));
  }
  var res;
  if (hasPipe) {
    res = go2jsTemplateResolveForPipe(tokens[0], args, dot, root, funcs, vars, ctx);
    if (typeof res === "function") res.__go2jsStage = { text: command.trim(), head: tokens[0].r || tokens[0].v, headAbs: base + (tokens[0].o !== undefined ? tokens[0].o : 0), base: base };
  } else {
    res = go2jsTemplateResolve(tokens[0], args, dot, root, funcs, vars, ctx);
  }
  ctx.stage = prev;
  return res;
}
function go2jsTemplateArity(word) {
  switch (word) {
    case "and": case "or": return { min: 1, max: Infinity };
    case "not": case "len": return { min: 1, max: 1 };
    case "index": return { min: 1, max: Infinity };
    case "print": case "println": return { min: 0, max: Infinity };
    case "printf": case "slice": case "call": return { min: 1, max: Infinity };
    case "html": case "js": case "urlquery": return { min: 1, max: 1 };
    case "eq": return { min: 1, max: Infinity };
    case "ne": case "lt": case "le": case "gt": case "ge": return { min: 2, max: 2 };
  }
  return null;
}
function go2jsTemplateWantString(ar) { return ar.min === ar.max ? String(ar.min) : ("at least " + ar.min); }
function go2jsTemplateArityErr(word, ar, got, ctx) {
  var e = go2jsTemplateExecFail("wrong number of args for " + word + ": want " + go2jsTemplateWantString(ar) + " got " + got, "token", 0);
  e.display = word;
  e.offset = ctx && ctx.stage ? ctx.stage.headAbs : 0;
  return e;
}
function go2jsTemplateTypeName(v) {
  if (v === null || v === undefined) return "<nil>";
  if (typeof v === "boolean") return "bool";
  if (typeof v === "string") return "string";
  if (typeof v === "number") return (Math.floor(v) === v && isFinite(v)) ? "int" : "float64";
  return typeof v;
}
function go2jsTemplateCallBuiltin(b, args, ctx) {
  try { return go2jsCallNow(b, null, args); }
  catch (e) {
    if (e && e.msg !== undefined && (e.kind === "call" || e.kind === "token")) {
      e.display = ctx && ctx.stage ? ctx.stage.text : e.display;
      e.offset = ctx && ctx.stage && typeof ctx.stage.headAbs === "number" ? ctx.stage.headAbs : e.offset;
    }
    throw e;
  }
}
function go2jsTemplateResolve(token, args, dot, root, funcs, vars, ctx) {
  if (token.k === "group") return go2jsTemplateEval(token.v, dot, root, funcs, vars, ctx);
  if (token.k === "string") return token.v;
  var word = token.v;
  if (word === ".") return go2jsTemplateCall(dot, args);
  if (word.startsWith("$")) {
    var d2 = word.indexOf(".");
    var name = d2 === -1 ? word : word.slice(0,d2);
    var base = (vars[name] !== undefined) ? vars[name] : (name === "$" ? root : null);
    return go2jsTemplateField(word.slice(name.length), base, args, dot, ctx, token);
  }
  if (/^-?\.[0-9]+$/.test(word)) return parseFloat(word);
  if (word.startsWith(".")) return go2jsTemplateField(word, dot, args, dot, ctx, token);
  if (word === "true") return true;
  if (word === "false") return false;
  if (word === "nil") { var ne = go2jsTemplateExecFail("nil is not a command", "call", 0); ne.display = "nil"; ne.offset = ctx && ctx.stage ? ctx.stage.headAbs : 0; throw ne; }
  if (/^-?[0-9]+$/.test(word)) return parseInt(word,10);
  if (/^-?\.[0-9]+$/.test(word) || /^-?[0-9]+\.[0-9]+$/.test(word)) return parseFloat(word);
  var ar = go2jsTemplateArity(word);
  if (ar && (args.length < ar.min || args.length > ar.max)) throw go2jsTemplateArityErr(word, ar, args.length, ctx);
  var b = go2jsTemplateBuiltin(word, funcs, vars);
  if (b !== undefined) return go2jsTemplateCallBuiltin(b, args, ctx);
  if (funcs && funcs[word] !== undefined) {
    var fn = funcs[word];
    return typeof fn === "function" ? go2jsTemplateCall(fn,args) : fn;
  }
  if (args.length === 0) return "";
  throw go2jsTemplateExecFail('function "' + word + '" not defined', 'token', token.o);
}
function go2jsTemplateCall(fn, args){ if (typeof fn !== "function") return fn; return go2jsCallNow(fn,null,args); }

function go2jsMapIndex(m,k){ return go2jsMapGet(m,k,null); }
function go2jsTemplatePrintValue(v) {
  if (typeof go2jsInterfaceValue === "function") v = go2jsInterfaceValue(v);
  if (v === null || v === undefined) return "<nil>";
  if (typeof v === "string") return v;
  if (typeof v === "boolean") return v ? "true" : "false";
  if (typeof v === "number") return go2jsFormatFloatDefault ? go2jsFormatFloatDefault(v) : String(v);
  if (typeof v === "function") return "<nil>";
  return go2jsFormat(v);
}
function go2jsTemplateHTMLEscape(s) {
  return String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&#34;").replace(/'/g, "&#39;");
}
function go2jsTemplateJSEscape(s) {
  var str = String(s);
  var out = "";
  for (var i = 0; i < str.length; i++) {
    var c = str[i];
    if (c === "\\") out += "\\\\";
    else if (c === "'") out += "\\'";
    else if (c === '"') out += '\\"';
    else if (c === "<") out += "\\u003C";
    else if (c === ">") out += "\\u003E";
    else if (c === "&") out += "\\u0026";
    else if (c === "=") out += "\\u003D";
    else out += c;
  }
  return out;
}
function go2jsTemplateURLQueryEscape(s) {
  var str = String(s);
  var out = "";
  var hex = "0123456789ABCDEF";
  for (var i = 0; i < str.length; i++) {
    var c = str[i];
    if ((c >= "A" && c <= "Z") || (c >= "a" && c <= "z") || (c >= "0" && c <= "9") || c === "-" || c === "_" || c === "." || c === "~") { out += c; continue; }
    if (c === " ") { out += "+"; continue; }
    var code = str.charCodeAt(i);
    if (code < 0x80) { out += "%" + hex[(code >> 4) & 15] + hex[code & 15]; }
    else { var bytes = go2jsStringToBytes(c); for (var b = 0; b < bytes.length; b++) { var by = bytes[b]; out += "%" + hex[(by >> 4) & 15] + hex[by & 15]; } }
  }
  return out;
}
function go2jsTemplateActionString(v){
  if (v===null||v===undefined) return "<no value>";
  return go2jsTemplateString(v);
}
function go2jsTemplateString(v){ if (v===null||v===undefined) return ""; if (typeof v==="string") return v; if (typeof v==="function") return ""; if (typeof v==="boolean") return v?"true":"false"; if (typeof v==="number") return go2jsFormatFloatDefault?go2jsFormatFloatDefault(v):String(v); if (v.__go2js_reflectValue===true) return go2jsTemplateString(v.v); return go2jsFormat(v); }
function go2jsTemplateProperty(v, name){
  if (v===null||v===undefined) return null;
  if (typeof go2jsInterfaceValue === "function") v = go2jsInterfaceValue(v);
  if (v===null||v===undefined) return null;
  if (name.startsWith("$")) return null;
  if (v instanceof go2jsNativeMap) return go2jsMapGet(v, name, null);
  if (v.hasOwnProperty && v.hasOwnProperty(name)) {
    var d = v[name];
    if (typeof d === "function") return d.bind(v);
    return d;
  }
  var d2 = v[name];
  if (typeof d2 === "function") return d2.bind(v);
  return d2;
}
function go2jsTemplateField(word, base, args, dot, ctx, token){
  if (word === ".") return go2jsTemplateCall(typeof go2jsInterfaceValue === "function" ? go2jsInterfaceValue(base) : base, args);
  var parts = word.split(".").slice(1);
  var v = typeof go2jsInterfaceValue === "function" ? go2jsInterfaceValue(base) : base;
  var fieldBase = (ctx && typeof ctx.actionBase === "number") ? ctx.actionBase : 0;
  for (var i=0;i<parts.length;i++){
    var p = parts[i];
    if (v===null||v===undefined) throw go2jsTemplateExecFail("can't evaluate field "+p+" in type "+go2jsGoTypeName(v), "field", fieldBase+((token&&token.o!==undefined)?token.o:0), word);
    var isMap = (v instanceof go2jsNativeMap);
    if (isMap) {
      v = go2jsMapGet(v, p, null);
      continue;
    }
    if (v.hasOwnProperty && v.hasOwnProperty(p)) {
      var d = v[p];
      if (typeof d === "function") v = d.bind(v);
      else v = d;
      continue;
    }
    throw go2jsTemplateExecFail("can't evaluate field "+p+" in type "+go2jsGoTypeName(v), "field", fieldBase+((token&&token.o!==undefined)?token.o:0), word);
  }
  if (v===null||v===undefined) return null;
  if (typeof v === "function") {
    if (args.length>0) return go2jsTemplateCall(v,args);
  }
  return v;
}
function go2jsTemplateTruth(v){
  if (typeof go2jsInterfaceValue === "function") v = go2jsInterfaceValue(v);
  if (v===null||v===undefined||v===false) return false;
  if (v===true) return true;
  if (typeof v==="number") return v!==0;
  if (typeof v==="string") return v.length>0;
  if (Array.isArray(v)) return v.length>0;
  if (v instanceof go2jsNativeMap) return v.size>0;
  if (typeof v==="object") return Object.keys(v).length>0;
  return true;
}
function go2jsTemplateBuiltin(word, funcs, vars){
  switch(word){
    case "and":
      return function(){ var args=Array.prototype.slice.call(arguments); if (args.length===0) return false; for (var a of args){ if (!go2jsTemplateTruth(a)) return a; } return args[args.length-1]; };
    case "or":
      return function(){ var args=Array.prototype.slice.call(arguments); if (args.length===0) return false; for (var a of args){ if (go2jsTemplateTruth(a)) return a; } return args[args.length-1]; };
    case "not":
      return function(v){ return !go2jsTemplateTruth(v); };
    case "len":
      return function(v){ if (v===null||v===undefined) return 0; if (typeof v==="string"||Array.isArray(v)) return v.length; if (v instanceof go2jsNativeMap) return v.size; if (typeof v==="object") return Object.keys(v).length; return 0; };
    case "index":
      return function(v){ var ks=Array.prototype.slice.call(arguments,1); var cur=v; if (typeof go2jsInterfaceValue === "function") cur = go2jsInterfaceValue(cur); for (var i=0;i<ks.length;i++){ var k=ks[i]; if (cur===null||cur===undefined) return null; if (Array.isArray(cur)){ cur=cur[Number(k)]; continue; } if (cur instanceof go2jsNativeMap){ cur = go2jsMapGet(cur, k, null); continue; } cur = cur[k]; } return cur===undefined?null:cur; };
    case "print":
      return function(){ var args=Array.prototype.slice.call(arguments); return args.map(x=>go2jsTemplatePrintValue(x)).join(""); };
    case "println":
      return function(){ var args=Array.prototype.slice.call(arguments); return args.map(x=>go2jsTemplatePrintValue(x)).join(" ")+"\n"; };
    case "html":
      return function(v){ return go2jsTemplateHTMLEscape(go2jsTemplatePrintValue(v)); };
    case "js":
      return function(v){ return go2jsTemplateJSEscape(go2jsTemplatePrintValue(v)); };
    case "urlquery":
      return function(v){ return go2jsTemplateURLQueryEscape(go2jsTemplatePrintValue(v)); };
    case "printf":
      return function(f){ if (typeof f !== "string") throw go2jsTemplateExecFail("wrong type for value; expected string; got " + go2jsTemplateTypeName(f), "call", 0); var args=Array.prototype.slice.call(arguments,1); return go2jsSprintf.apply(null,[f].concat(args)); };
    case "call":
      return function(fn){ var a=Array.prototype.slice.call(arguments,1); return go2jsCallNow(fn,null,a); };
    case "slice":
      return function(v){ var b=Array.prototype.slice.call(arguments,1); var cur=v; if (typeof cur==="string"){ var s=String(cur); var l=typeof b[0] === "undefined"?0:Number(b[0]), r=typeof b[1] === "undefined"?s.length:Number(b[1]); return s.slice(l,r); } if (Array.isArray(cur)){ var a=cur; var l=typeof b[0] === "undefined"?0:Number(b[0]); var r=typeof b[1] === "undefined"?a.length:Number(b[1]); return a.slice(l,r); } return cur; };
    case "eq":
    case "ne":
    case "lt":
    case "le":
    case "gt":
    case "ge":
      // basic
      return function(){ var args=Array.prototype.slice.call(arguments); return go2jsTemplateBuiltinCmp(word, args); };
  }
  return undefined;
}
function go2jsTemplateEqual(left,right){
  if (left===right) return true;
  if (left===null||left===undefined||right===null||right===undefined) return false;
  if (typeof left==="number" && typeof right==="number") return left===right;
  if (typeof left==="string" && typeof right==="string") return left===right;
  return go2jsEqual(left,right)===true;
}
function go2jsTemplateCompare(left,right){
  if (typeof left==="string" && typeof right==="string") return left<right?-1:left>right?1:0;
  var a=Number(left), b=Number(right);
  return a<b?-1:a>b?1:0;
}
function go2jsTemplateBuiltinCmp(op,args){
  if (op==="eq"){ if (args.length===1) throw go2jsTemplateExecFail("error calling eq: missing argument for comparison","call",0); var x=args[0]; for (var i=1;i<args.length;i++) if (go2jsTemplateEqual(x,args[i])) return true; return false; }
  if (op==="ne"){ if (args.length!==2) throw go2jsTemplateExecFail("wrong number of args for ne: want 2 got "+args.length,"token",0); return !go2jsTemplateEqual(args[0],args[1]); }
  if (op==="lt"){ if (args.length!==2) throw go2jsTemplateExecFail("wrong number of args for lt: want 2 got "+args.length,"token",0); return go2jsTemplateCompare(args[0],args[1])<0; }
  if (op==="le"){ if (args.length!==2) throw go2jsTemplateExecFail("wrong number of args for le: want 2 got "+args.length,"token",0); return go2jsTemplateCompare(args[0],args[1])<=0; }
  if (op==="gt"){ if (args.length!==2) throw go2jsTemplateExecFail("wrong number of args for gt: want 2 got "+args.length,"token",0); return go2jsTemplateCompare(args[0],args[1])>0; }
  if (op==="ge"){ if (args.length!==2) throw go2jsTemplateExecFail("wrong number of args for ge: want 2 got "+args.length,"token",0); return go2jsTemplateCompare(args[0],args[1])>=0; }
  return false;
}
function go2jsTemplateResolveForPipe(token, args, dot, root, funcs, vars, ctx) {
  if (token.k === "group") return go2jsTemplateEval(token.v, dot, root, funcs, vars, ctx);
  if (token.k === "string") return token.v;
  var word = token.v;
  if (word.startsWith("$")) {
    var d2 = word.indexOf(".");
    var name = d2 === -1 ? word : word.slice(0,d2);
    var base = (vars[name] !== undefined) ? vars[name] : (name === "$" ? root : null);
    var f = go2jsTemplateField(word.slice(name.length), base, [], dot, ctx, token);
    if (typeof f !== "function") throw go2jsTemplateExecFail("non executable command in pipeline stage", "token", token.o);
    return function(v){ var all = args.slice(); all.push(v); return go2jsCallNow(f,null,all); };
  }
  if (word.startsWith(".")) {
    var f = go2jsTemplateField(word, dot, [], dot, ctx, token);
    if (typeof f !== "function") throw go2jsTemplateExecFail("non executable command in pipeline stage", "token", token.o);
    return function(v){ var all = args.slice(); all.push(v); return go2jsCallNow(f,null,all); };
  }
  if (/^-?\d/.test(word) || word==="true"||word==="false"||word==="nil"||word===".") {
    throw go2jsTemplateExecFail("non executable command in pipeline stage", "token", token.o);
  }
  var ar2 = go2jsTemplateArity(word);
  if (ar2 && (args.length + 1 < ar2.min || args.length + 1 > ar2.max)) throw go2jsTemplateArityErr(word, ar2, args.length + 1, ctx);
  var b = go2jsTemplateBuiltin(word, funcs, vars);
  if (b !== undefined) {
    return function(v){ var all = args.slice(); all.push(v); return go2jsTemplateCallBuiltin(b, all, ctx); };
  }
  if (funcs && funcs[word] !== undefined) {
    var fn = funcs[word];
    if (typeof fn !== "function") return fn;
    return function(v){ var all = args.slice(); all.push(v); return go2jsCallNow(fn,null,all); };
  }
  throw go2jsTemplateExecFail('function "' + word + '" not defined', "token", token.o);
}
`
}
