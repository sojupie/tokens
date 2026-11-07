// Build a delivery-only CSS pack. Exported tokens and component downloads retain
// their original bytes. Names referenced outside CSS stay public: controllers,
// inline styles and static consumers need no translation table or runtime shim.
import {readFile,writeFile,mkdir,readdir} from 'node:fs/promises';
import {createHash} from 'node:crypto';
import path from 'node:path';
import {rewriteProperties,shortName} from './short-css-names.ts';

const digest=(text:string)=>createHash('sha256').update(text).digest('hex');
export function deliveryStyles(styles:Record<string,string>,runtimeSources:string[]) {
 const prefixes=new Set(runtimeSources.flatMap(source=>source.match(/--[\w-]+/g)??[]));
 const counts=new Map<string,number>();
 for(const css of Object.values(styles)) rewriteProperties(css,name=>{counts.set(name,(counts.get(name)??0)+1);return name;});
 const names=[...counts].sort((a,b)=>b[1]-a[1]||(a[0]<b[0]?-1:a[0]>b[0]?1:0));
 const mapping:Record<string,string>={};let index=0;
 for(const [name] of names) {
  if([...prefixes].some(prefix=>name.startsWith(prefix)))continue;
  let shortened:string;do{shortened=shortName(index++);}while(counts.has(shortened));
  mapping[name]=shortened;
 }
 const output=Object.fromEntries(Object.entries(styles).map(([url,css])=>[url,rewriteProperties(css,name=>mapping[name]??name)]));
 return {id:digest(JSON.stringify(mapping)).slice(0,16),styles:output,shortened:Object.keys(mapping).length,properties:counts.size};
}

async function files(directory:string):Promise<string[]> {
 const entries=await readdir(directory,{withFileTypes:true});
 const result:string[]=[];
 for(const entry of entries){if(['generated','vendor','node_modules'].includes(entry.name))continue;const file=path.join(directory,entry.name);if(entry.isDirectory())result.push(...await files(file));else result.push(file);}
 return result;
}

export async function buildDeliveryStyles(assets:string) {
 const styles:Record<string,string>={};
 for(const directory of ['catalog/web/assets/styles','forge/components/generated','forge/components/generated/bundles']) {
  for(const name of (await readdir(directory)).sort())if(name.endsWith('.css') && name !== 'components-base.css'){
   const relative=directory.endsWith('/bundles')?'bundles/'+name:name;
   styles['/static/'+relative]=await readFile(path.join(directory,name),'utf8');
  }
 }
 styles['/static/site.css']=await readFile('forge/site/site.css','utf8');
 styles['/theme.css']=await readFile(path.join(assets,'theme.css'),'utf8');
 for(const name of (await readdir(path.join(assets,'themes'))).sort())if(name.endsWith('.css'))styles['/themes/'+name]=await readFile(path.join(assets,'themes',name),'utf8');
 const sources=(await Promise.all(['catalog/web','forge/components','forge/site'].map(files))).flat().filter(file=>/\.(?:go|templ|js|ts)$/.test(file)&&!file.includes('/internal/render/'));
 const result=deliveryStyles(styles,await Promise.all(sources.map(file=>readFile(file,'utf8'))));
 const directory=path.join(assets,'delivery-css');await mkdir(directory,{recursive:true});
 const manifest:{id:string;files:Record<string,{file:string;sourceSha256:string;sha256:string}>}={id:result.id,files:{}};
 for(const [url,css] of Object.entries(result.styles)){
  const hash=digest(css),file=hash+'.css';await writeFile(path.join(directory,file),css);
  manifest.files[url]={file,sourceSha256:digest(styles[url]),sha256:hash};
 }
 await writeFile(path.join(directory,'manifest.json'),JSON.stringify(manifest)+'\n');
 console.log(`Delivery CSS: shortened ${result.shortened}/${result.properties} properties; ${Object.values(styles).reduce((n,s)=>n+Buffer.byteLength(s),0)} → ${Object.values(result.styles).reduce((n,s)=>n+Buffer.byteLength(s),0)} bytes. Downloads retain original names.`);
 return manifest;
}
