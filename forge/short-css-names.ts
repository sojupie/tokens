// Experimental closed-corpus transform. Public token names remain unchanged.
import postcss from 'postcss';
import {readFile,writeFile} from 'node:fs/promises';
import {pathToFileURL} from 'node:url';

type Sheet = { text: string; url?: string };
type Page = { styles: Sheet[]; inline: string[] };
type Visitor = (name: string) => string;

const alphabet='abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-';
export function shortName(index: number){
  let length=1, capacity=alphabet.length;
  while(index>=capacity){index-=capacity;length++;capacity*=alphabet.length;}
  let result='';
  for(let i=0;i<length;i++){result=alphabet[index%alphabet.length]+result;index=Math.floor(index/alphabet.length);}
  return '--'+result;
}

// Visit identifiers outside strings and comments. Escaped/non-ASCII property
// names fail closed rather than silently producing an incomplete mapping.
function identifiers(text: string, visit: (name: string, index: number) => string){
  let result='',i=0;
  while(i<text.length){
    const start=i;
    if(text.startsWith('/*',i)){
      const end=text.indexOf('*/',i+2);i=end<0?text.length:end+2;
    }else if(text[i]==='"'||text[i]==="'"){
      const quote=text[i++];
      while(i<text.length){if(text[i++]==='\\'){i++;continue;}if(text[i-1]===quote)break;}
    }else if(text.startsWith('--',i)&&!/[\w-]/.test(text[i-1]||'')){
      i+=2;
      while(i<text.length&&/[\w-]/.test(text[i]))i++;
      if(text[i]==='\\'||text.charCodeAt(i)>127)throw Error('Escaped/non-ASCII custom property requires a CSS tokenizer');
      const name=text.slice(start,i);
      result+=name.length>2?visit(name,start):name;continue;
    }else{i++;}
    result+=text.slice(start,i);
  }
  return result;
}

function walk(css: string, visit: Visitor){
  const variable=(text: string,index: number)=>/\b(?:var|style)\(\s*$/i.test(text.slice(0,index).replace(/\/\*[\s\S]*?\*\//g,''));
  const root=postcss.parse(css);
  root.walkDecls(decl=>{
    if(decl.prop.startsWith('--')){
      if(!/^--[\w-]+$/.test(decl.prop))throw Error('Unsupported custom property '+decl.prop);
      decl.prop=visit(decl.prop);
    }
    const value=decl.raws.value?.raw ?? decl.value;
    decl.value=identifiers(value,(name,index)=>variable(value,index)?visit(name):name);
  });
  root.walkAtRules(rule=>{
    const params=rule.raws.params?.raw ?? rule.params;
    rule.params=identifiers(params,(name,index)=>
      rule.name==='property'||/^\s*:/.test(params.slice(index+name.length))||variable(params,index)?visit(name):name);
  });
  return root.toString();
}

export function shorten<P extends Page>(corpus: { pages: P[] }){
  const counts=new Map<string, number>();
  const collect: Visitor=name=>{counts.set(name,(counts.get(name)||0)+1);return name;};
  for(const page of corpus.pages){
    for(const sheet of page.styles){
      if(walk(sheet.text,collect)!==sheet.text)throw Error('Identity transform changed formatting: '+sheet.url);
    }
    for(const inline of page.inline)walk('x{'+inline+'}',collect);
  }
  const names=[...counts].sort((a,b)=>b[1]-a[1]||(a[0]<b[0]?-1:a[0]>b[0]?1:0));
  const mapping: Record<string,string> = {};
  let index=0;
  for(const [name] of names){
    mapping[name]=shortName(index++);
  }
  const rename: Visitor=name=>{if(!counts.has(name))throw Error('Unmapped '+name);return mapping[name];};
  const pages=corpus.pages.map(page=>({...page,
    styles:page.styles.map(sheet=>({...sheet,text:walk(sheet.text,rename)})),
    inline:page.inline.map(text=>walk('x{'+text+'}',rename).slice(2,-1)),
  }));
  return {mapping,occurrences:Object.fromEntries(names),pages};
}

if(process.argv[1]&&import.meta.url===pathToFileURL(process.argv[1]).href){
  const [input,output]=process.argv.slice(2);
  if(!input||!output)throw Error('Usage: node forge/short-css-names.ts input.json output.json');
  const corpus=JSON.parse(await readFile(input,'utf8'));
  const result=shorten(corpus);
  await writeFile(output,JSON.stringify(result));
}

// Release builds use the same CSS tokenizer with a conservative runtime boundary.
export { walk as rewriteProperties };
