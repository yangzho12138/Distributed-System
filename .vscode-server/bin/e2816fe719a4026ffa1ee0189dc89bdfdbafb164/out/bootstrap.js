"use strict";(function(u,l){typeof exports=="object"?module.exports=l():u.MonacoBootstrap=l()})(this,function(){const u=typeof require=="function"?require("module"):void 0,l=typeof require=="function"?require("path"):void 0,f=typeof require=="function"?require("fs"):void 0;if(Error.stackTraceLimit=100,typeof process<"u"&&!process.env.VSCODE_HANDLES_SIGPIPE){let e=!1;process.on("SIGPIPE",()=>{e||(e=!0,console.error(new Error("Unexpected SIGPIPE")))})}function S(e){if(!l||!u||typeof process>"u"){console.warn("enableASARSupport() is only available in node.js environments");return}const r=e?l.join(e,"node_modules"):l.join(__dirname,"../node_modules");let n;if(e&&process.platform==="win32"){const s=e.substr(0,1);let i;s.toLowerCase()!==s?i=s.toLowerCase():i=s.toUpperCase(),n=i+r.substr(1)}else n=void 0;const t=`${r}.asar`,L=n?`${n}.asar`:void 0,c=u._resolveLookupPaths;u._resolveLookupPaths=function(s,i){const o=c(s,i);if(Array.isArray(o)){let d=!1;for(let a=0,P=o.length;a<P;a++)if(o[a]===r){d=!0,o.splice(a,0,t);break}else if(o[a]===n){d=!0,o.splice(a,0,L);break}!d&&e&&o.push(t)}return o}}function E(e,r){let n=e.replace(/\\/g,"/");n.length>0&&n.charAt(0)!=="/"&&(n=`/${n}`);let t;return r.isWindows&&n.startsWith("//")?t=encodeURI(`${r.scheme||"file"}:${n}`):t=encodeURI(`${r.scheme||"file"}://${r.fallbackAuthority||""}${n}`),t.replace(/#/g,"%23")}function A(){const e=h();let r={availableLanguages:{}};if(e&&e.env.VSCODE_NLS_CONFIG)try{r=JSON.parse(e.env.VSCODE_NLS_CONFIG)}catch{}if(r._resolvedLanguagePackCoreLocation){const n=Object.create(null);r.loadBundle=function(t,L,c){const s=n[t];if(s){c(void 0,s);return}b(r._resolvedLanguagePackCoreLocation,`${t.replace(/\//g,"!")}.nls.json`).then(function(i){const o=JSON.parse(i);n[t]=o,c(void 0,o)}).catch(i=>{try{r._corruptedFile&&N(r._corruptedFile,"corrupted").catch(function(o){console.error(o)})}finally{c(i,void 0)}})}}return r}function p(){return(typeof self=="object"?self:typeof global=="object"?global:{}).vscode}function h(){const e=p();if(e)return e.process;if(typeof process<"u")return process}function _(){const e=p();if(e)return e.ipcRenderer}async function b(...e){const r=_();if(r)return r.invoke("vscode:readNlsFile",...e);if(f&&l)return(await f.promises.readFile(l.join(...e))).toString();throw new Error("Unsupported operation (read NLS files)")}function N(e,r){const n=_();if(n)return n.invoke("vscode:writeNlsFile",e,r);if(f)return f.promises.writeFile(e,r);throw new Error("Unsupported operation (write NLS files)")}return{enableASARSupport:S,setupNLS:A,fileUriFromPath:E}});

//# sourceMappingURL=https://ticino.blob.core.windows.net/sourcemaps/e2816fe719a4026ffa1ee0189dc89bdfdbafb164/core/bootstrap.js.map
