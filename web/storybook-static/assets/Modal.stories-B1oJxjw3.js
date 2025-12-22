import{j as e}from"./jsx-runtime-BYYWji4R.js";import{r as c,R as p,a as me}from"./index-ClcD9ViR.js";import{c as se,a as _t}from"./cn-JCLedEej.js";import{p as Xt,r as Jt}from"./bugs-Ot56VXf3.js";import{o as N,U as E,y as L,e as Nt,u as H,C as I,t as wt,a as Mt,l as ge,I as _,O as Ne,c as Kt}from"./keyboard-DNQdDreo.js";import{n as Qt,s as ae,c as Zt,E as Ot,e as es,N as ts,l as pe,t as fe}from"./use-root-containers-Dht_jFDM.js";import{O as re,M as T,y as S,a as Bt,N as ss}from"./use-is-mounted-62Q7yJH9.js";import{n as St}from"./use-owner-CcFJWZQm.js";import{u as we,s as Me}from"./hidden-BnOBa8gi.js";import{t as as,u as ns,d as ne,y as rs,q as ie}from"./transition-CF-Dw_VV.js";import{G as os,w as ls}from"./description-Bf17jbPp.js";import{B as u}from"./Button-DlbfybpJ.js";import"./_commonjsHelpers-Cpj98o6Y.js";import"./index-BUAr5TKG.js";function xe(t,s){let a=c.useRef([]),n=N(t);c.useEffect(()=>{let r=[...a.current];for(let[o,i]of s.entries())if(a.current[o]!==i){let l=n(s,r);return a.current=s,l}},[n,...s])}function is(t){function s(){document.readyState!=="loading"&&(t(),document.removeEventListener("DOMContentLoaded",s))}typeof window<"u"&&typeof document<"u"&&(document.addEventListener("DOMContentLoaded",s),s())}let M=[];is(()=>{function t(s){s.target instanceof HTMLElement&&s.target!==document.body&&M[0]!==s.target&&(M.unshift(s.target),M=M.filter(a=>a!=null&&a.isConnected),M.splice(10))}window.addEventListener("click",t,{capture:!0}),window.addEventListener("mousedown",t,{capture:!0}),window.addEventListener("focus",t,{capture:!0}),document.body.addEventListener("click",t,{capture:!0}),document.body.addEventListener("mousedown",t,{capture:!0}),document.body.addEventListener("focus",t,{capture:!0})});function Dt(t){if(!t)return new Set;if(typeof t=="function")return new Set(t());let s=new Set;for(let a of t.current)a.current instanceof HTMLElement&&s.add(a.current);return s}let ds="div";var Tt=(t=>(t[t.None=1]="None",t[t.InitialFocus=2]="InitialFocus",t[t.TabLock=4]="TabLock",t[t.FocusLock=8]="FocusLock",t[t.RestoreFocus=16]="RestoreFocus",t[t.All=30]="All",t))(Tt||{});function cs(t,s){let a=c.useRef(null),n=L(a,s),{initialFocus:r,containers:o,features:i=30,...l}=t;Nt()||(i=1);let d=St(a);ps({ownerDocument:d},!!(i&16));let g=fs({ownerDocument:d,container:a,initialFocus:r},!!(i&2));hs({ownerDocument:d,container:a,containers:o,previousActiveElement:g},!!(i&8));let f=Qt(),O=N(b=>{let x=a.current;x&&(w=>w())(()=>{H(f.current,{[ae.Forwards]:()=>{re(x,T.First,{skipElements:[b.relatedTarget]})},[ae.Backwards]:()=>{re(x,T.Last,{skipElements:[b.relatedTarget]})}})})}),J=Xt(),k=c.useRef(!1),C={ref:n,onKeyDown(b){b.key=="Tab"&&(k.current=!0,J.requestAnimationFrame(()=>{k.current=!1}))},onBlur(b){let x=Dt(o);a.current instanceof HTMLElement&&x.add(a.current);let w=b.relatedTarget;w instanceof HTMLElement&&w.dataset.headlessuiFocusGuard!=="true"&&(Et(x,w)||(k.current?re(a.current,H(f.current,{[ae.Forwards]:()=>T.Next,[ae.Backwards]:()=>T.Previous})|T.WrapAround,{relativeTo:b.target}):b.target instanceof HTMLElement&&S(b.target)))}};return p.createElement(p.Fragment,null,!!(i&4)&&p.createElement(we,{as:"button",type:"button","data-headlessui-focus-guard":!0,onFocus:O,features:Me.Focusable}),I({ourProps:C,theirProps:l,defaultTag:ds,name:"FocusTrap"}),!!(i&4)&&p.createElement(we,{as:"button",type:"button","data-headlessui-focus-guard":!0,onFocus:O,features:Me.Focusable}))}let us=E(cs),A=Object.assign(us,{features:Tt});function ms(t=!0){let s=c.useRef(M.slice());return xe(([a],[n])=>{n===!0&&a===!1&&wt(()=>{s.current.splice(0)}),n===!1&&a===!0&&(s.current=M.slice())},[t,M,s]),N(()=>{var a;return(a=s.current.find(n=>n!=null&&n.isConnected))!=null?a:null})}function ps({ownerDocument:t},s){let a=ms(s);xe(()=>{s||(t==null?void 0:t.activeElement)===(t==null?void 0:t.body)&&S(a())},[s]),Zt(()=>{s&&S(a())})}function fs({ownerDocument:t,container:s,initialFocus:a},n){let r=c.useRef(null),o=Bt();return xe(()=>{if(!n)return;let i=s.current;i&&wt(()=>{if(!o.current)return;let l=t==null?void 0:t.activeElement;if(a!=null&&a.current){if((a==null?void 0:a.current)===l){r.current=l;return}}else if(i.contains(l)){r.current=l;return}a!=null&&a.current?S(a.current):re(i,T.First)===ss.Error&&console.warn("There are no focusable elements inside the <FocusTrap />"),r.current=t==null?void 0:t.activeElement})},[n]),r}function hs({ownerDocument:t,container:s,containers:a,previousActiveElement:n},r){let o=Bt();Ot(t==null?void 0:t.defaultView,"focus",i=>{if(!r||!o.current)return;let l=Dt(a);s.current instanceof HTMLElement&&l.add(s.current);let d=n.current;if(!d)return;let g=i.target;g&&g instanceof HTMLElement?Et(l,g)?(n.current=g,S(g)):(i.preventDefault(),i.stopPropagation(),S(d)):S(n.current)},!0)}function Et(t,s){for(let a of t)if(a.contains(s))return!0;return!1}function gs(t,s){return t===s&&(t!==0||1/t===1/s)||t!==t&&s!==s}const xs=typeof Object.is=="function"?Object.is:gs,{useState:vs,useEffect:ys,useLayoutEffect:bs,useDebugValue:Cs}=me;function ks(t,s,a){const n=s(),[{inst:r},o]=vs({inst:{value:n,getSnapshot:s}});return bs(()=>{r.value=n,r.getSnapshot=s,de(r)&&o({inst:r})},[t,n,s]),ys(()=>(de(r)&&o({inst:r}),t(()=>{de(r)&&o({inst:r})})),[t]),Cs(n),n}function de(t){const s=t.getSnapshot,a=t.value;try{const n=s();return!xs(a,n)}catch{return!0}}function js(t,s,a){return s()}const Ns=typeof window<"u"&&typeof window.document<"u"&&typeof window.document.createElement<"u",ws=!Ns,Ms=ws?js:ks,Os="useSyncExternalStore"in me?(t=>t.useSyncExternalStore)(me):Ms;function Bs(t){return Os(t.subscribe,t.getSnapshot,t.getSnapshot)}function Ss(t,s){let a=t(),n=new Set;return{getSnapshot(){return a},subscribe(r){return n.add(r),()=>n.delete(r)},dispatch(r,...o){let i=s[r].call(a,...o);i&&(a=i,n.forEach(l=>l()))}}}function Ds(){let t;return{before({doc:s}){var a;let n=s.documentElement;t=((a=s.defaultView)!=null?a:window).innerWidth-n.clientWidth},after({doc:s,d:a}){let n=s.documentElement,r=n.clientWidth-n.offsetWidth,o=t-r;a.style(n,"paddingRight",`${o}px`)}}}function Ts(){return as()?{before({doc:t,d:s,meta:a}){function n(r){return a.containers.flatMap(o=>o()).some(o=>o.contains(r))}s.microTask(()=>{var r;if(window.getComputedStyle(t.documentElement).scrollBehavior!=="auto"){let l=Mt();l.style(t.documentElement,"scrollBehavior","auto"),s.add(()=>s.microTask(()=>l.dispose()))}let o=(r=window.scrollY)!=null?r:window.pageYOffset,i=null;s.addEventListener(t,"click",l=>{if(l.target instanceof HTMLElement)try{let d=l.target.closest("a");if(!d)return;let{hash:g}=new URL(d.href),f=t.querySelector(g);f&&!n(f)&&(i=f)}catch{}},!0),s.addEventListener(t,"touchstart",l=>{if(l.target instanceof HTMLElement)if(n(l.target)){let d=l.target;for(;d.parentElement&&n(d.parentElement);)d=d.parentElement;s.style(d,"overscrollBehavior","contain")}else s.style(l.target,"touchAction","none")}),s.addEventListener(t,"touchmove",l=>{if(l.target instanceof HTMLElement)if(n(l.target)){let d=l.target;for(;d.parentElement&&d.dataset.headlessuiPortal!==""&&!(d.scrollHeight>d.clientHeight||d.scrollWidth>d.clientWidth);)d=d.parentElement;d.dataset.headlessuiPortal===""&&l.preventDefault()}else l.preventDefault()},{passive:!1}),s.add(()=>{var l;let d=(l=window.scrollY)!=null?l:window.pageYOffset;o!==d&&window.scrollTo(0,o),i&&i.isConnected&&(i.scrollIntoView({block:"nearest"}),i=null)})})}}:{}}function Es(){return{before({doc:t,d:s}){s.style(t.documentElement,"overflow","hidden")}}}function Ls(t){let s={};for(let a of t)Object.assign(s,a(s));return s}let B=Ss(()=>new Map,{PUSH(t,s){var a;let n=(a=this.get(t))!=null?a:{doc:t,count:0,d:Mt(),meta:new Set};return n.count++,n.meta.add(s),this.set(t,n),this},POP(t,s){let a=this.get(t);return a&&(a.count--,a.meta.delete(s)),this},SCROLL_PREVENT({doc:t,d:s,meta:a}){let n={doc:t,d:s,meta:Ls(a)},r=[Ts(),Ds(),Es()];r.forEach(({before:o})=>o==null?void 0:o(n)),r.forEach(({after:o})=>o==null?void 0:o(n))},SCROLL_ALLOW({d:t}){t.dispose()},TEARDOWN({doc:t}){this.delete(t)}});B.subscribe(()=>{let t=B.getSnapshot(),s=new Map;for(let[a]of t)s.set(a,a.documentElement.style.overflow);for(let a of t.values()){let n=s.get(a.doc)==="hidden",r=a.count!==0;(r&&!n||!r&&n)&&B.dispatch(a.count>0?"SCROLL_PREVENT":"SCROLL_ALLOW",a),a.count===0&&B.dispatch("TEARDOWN",a)}});function Is(t,s,a){let n=Bs(B),r=t?n.get(t):void 0,o=r?r.count>0:!1;return ge(()=>{if(!(!t||!s))return B.dispatch("PUSH",t,a),()=>B.dispatch("POP",t,a)},[s,t]),o}let ce=new Map,R=new Map;function Oe(t,s=!0){ge(()=>{var a;if(!s)return;let n=typeof t=="function"?t():t.current;if(!n)return;function r(){var i;if(!n)return;let l=(i=R.get(n))!=null?i:1;if(l===1?R.delete(n):R.set(n,l-1),l!==1)return;let d=ce.get(n);d&&(d["aria-hidden"]===null?n.removeAttribute("aria-hidden"):n.setAttribute("aria-hidden",d["aria-hidden"]),n.inert=d.inert,ce.delete(n))}let o=(a=R.get(n))!=null?a:0;return R.set(n,o+1),o!==0||(ce.set(n,{"aria-hidden":n.getAttribute("aria-hidden"),inert:n.inert}),n.setAttribute("aria-hidden","true"),n.inert=!0),r},[t,s])}let ve=c.createContext(()=>{});ve.displayName="StackContext";var he=(t=>(t[t.Add=0]="Add",t[t.Remove=1]="Remove",t))(he||{});function As(){return c.useContext(ve)}function Rs({children:t,onUpdate:s,type:a,element:n,enabled:r}){let o=As(),i=N((...l)=>{s==null||s(...l),o(...l)});return ge(()=>{let l=r===void 0||r===!0;return l&&i(0,a,n),()=>{l&&i(1,a,n)}},[i,a,n,r]),p.createElement(ve.Provider,{value:i},t)}var zs=(t=>(t[t.Open=0]="Open",t[t.Closed=1]="Closed",t))(zs||{}),Fs=(t=>(t[t.SetTitleId=0]="SetTitleId",t))(Fs||{});let $s={0(t,s){return t.titleId===s.id?t:{...t,titleId:s.id}}},oe=c.createContext(null);oe.displayName="DialogContext";function X(t){let s=c.useContext(oe);if(s===null){let a=new Error(`<${t} /> is missing a parent <Dialog /> component.`);throw Error.captureStackTrace&&Error.captureStackTrace(a,X),a}return s}function Ps(t,s,a=()=>[document.body]){Is(t,s,n=>{var r;return{containers:[...(r=n.containers)!=null?r:[],a]}})}function Ws(t,s){return H(s.type,$s,t,s)}let qs="div",Us=Ne.RenderStrategy|Ne.Static;function Gs(t,s){let a=_(),{id:n=`headlessui-dialog-${a}`,open:r,onClose:o,initialFocus:i,role:l="dialog",__demoMode:d=!1,...g}=t,[f,O]=c.useState(0),J=c.useRef(!1);l=function(){return l==="dialog"||l==="alertdialog"?l:(J.current||(J.current=!0,console.warn(`Invalid role [${l}] passed to <Dialog />. Only \`dialog\` and and \`alertdialog\` are supported. Using \`dialog\` instead.`)),"dialog")}();let k=ns();r===void 0&&k!==null&&(r=(k&ne.Open)===ne.Open);let C=c.useRef(null),b=L(C,s),x=St(C),w=t.hasOwnProperty("open")||k!==null,ye=t.hasOwnProperty("onClose");if(!w&&!ye)throw new Error("You have to provide an `open` and an `onClose` prop to the `Dialog` component.");if(!w)throw new Error("You provided an `onClose` prop to the `Dialog`, but forgot an `open` prop.");if(!ye)throw new Error("You provided an `open` prop to the `Dialog`, but forgot an `onClose` prop.");if(typeof r!="boolean")throw new Error(`You provided an \`open\` prop to the \`Dialog\`, but the value is not a boolean. Received: ${r}`);if(typeof o!="function")throw new Error(`You provided an \`onClose\` prop to the \`Dialog\`, but the value is not a function. Received: ${o}`);let v=r?0:1,[K,Lt]=c.useReducer(Ws,{titleId:null,descriptionId:null,panelRef:c.createRef()}),D=N(()=>o(!1)),be=N(m=>Lt({type:0,id:m})),Q=Nt()?d?!1:v===0:!1,Z=f>1,Ce=c.useContext(oe)!==null,[It,At]=es(),Rt={get current(){var m;return(m=K.panelRef.current)!=null?m:C.current}},{resolveContainers:le,mainTreeNodeRef:ee,MainTreeNode:zt}=ts({portals:It,defaultContainers:[Rt]}),Ft=Z?"parent":"leaf",ke=k!==null?(k&ne.Closing)===ne.Closing:!1,$t=Ce||ke?!1:Q,Pt=c.useCallback(()=>{var m,j;return(j=Array.from((m=x==null?void 0:x.querySelectorAll("body > *"))!=null?m:[]).find(y=>y.id==="headlessui-portal-root"?!1:y.contains(ee.current)&&y instanceof HTMLElement))!=null?j:null},[ee]);Oe(Pt,$t);let Wt=Z?!0:Q,qt=c.useCallback(()=>{var m,j;return(j=Array.from((m=x==null?void 0:x.querySelectorAll("[data-headlessui-portal]"))!=null?m:[]).find(y=>y.contains(ee.current)&&y instanceof HTMLElement))!=null?j:null},[ee]);Oe(qt,Wt),rs(le,m=>{m.preventDefault(),D()},!(!Q||Z));let Ut=!(Z||v!==0);Ot(x==null?void 0:x.defaultView,"keydown",m=>{Ut&&(m.defaultPrevented||m.key===Kt.Escape&&(m.preventDefault(),m.stopPropagation(),D()))}),Ps(x,!(ke||v!==0||Ce),le),c.useEffect(()=>{if(v!==0||!C.current)return;let m=new ResizeObserver(j=>{for(let y of j){let te=y.target.getBoundingClientRect();te.x===0&&te.y===0&&te.width===0&&te.height===0&&D()}});return m.observe(C.current),()=>m.disconnect()},[v,C,D]);let[Gt,Vt]=ls(),Yt=c.useMemo(()=>[{dialogState:v,close:D,setTitleId:be},K],[v,K,D,be]),je=c.useMemo(()=>({open:v===0}),[v]),Ht={ref:b,id:n,role:l,"aria-modal":v===0?!0:void 0,"aria-labelledby":K.titleId,"aria-describedby":Gt};return p.createElement(Rs,{type:"Dialog",enabled:v===0,element:C,onUpdate:N((m,j)=>{j==="Dialog"&&H(m,{[he.Add]:()=>O(y=>y+1),[he.Remove]:()=>O(y=>y-1)})})},p.createElement(pe,{force:!0},p.createElement(fe,null,p.createElement(oe.Provider,{value:Yt},p.createElement(fe.Group,{target:C},p.createElement(pe,{force:!1},p.createElement(Vt,{slot:je,name:"Dialog.Description"},p.createElement(A,{initialFocus:i,containers:le,features:Q?H(Ft,{parent:A.features.RestoreFocus,leaf:A.features.All&~A.features.FocusLock}):A.features.None},p.createElement(At,null,I({ourProps:Ht,theirProps:g,slot:je,defaultTag:qs,features:Us,visible:v===0,name:"Dialog"}))))))))),p.createElement(zt,null))}let Vs="div";function Ys(t,s){let a=_(),{id:n=`headlessui-dialog-overlay-${a}`,...r}=t,[{dialogState:o,close:i}]=X("Dialog.Overlay"),l=L(s),d=N(f=>{if(f.target===f.currentTarget){if(Jt(f.currentTarget))return f.preventDefault();f.preventDefault(),f.stopPropagation(),i()}}),g=c.useMemo(()=>({open:o===0}),[o]);return I({ourProps:{ref:l,id:n,"aria-hidden":!0,onClick:d},theirProps:r,slot:g,defaultTag:Vs,name:"Dialog.Overlay"})}let Hs="div";function _s(t,s){let a=_(),{id:n=`headlessui-dialog-backdrop-${a}`,...r}=t,[{dialogState:o},i]=X("Dialog.Backdrop"),l=L(s);c.useEffect(()=>{if(i.panelRef.current===null)throw new Error("A <Dialog.Backdrop /> component is being used, but a <Dialog.Panel /> component is missing.")},[i.panelRef]);let d=c.useMemo(()=>({open:o===0}),[o]);return p.createElement(pe,{force:!0},p.createElement(fe,null,I({ourProps:{ref:l,id:n,"aria-hidden":!0},theirProps:r,slot:d,defaultTag:Hs,name:"Dialog.Backdrop"})))}let Xs="div";function Js(t,s){let a=_(),{id:n=`headlessui-dialog-panel-${a}`,...r}=t,[{dialogState:o},i]=X("Dialog.Panel"),l=L(s,i.panelRef),d=c.useMemo(()=>({open:o===0}),[o]),g=N(f=>{f.stopPropagation()});return I({ourProps:{ref:l,id:n,onClick:g},theirProps:r,slot:d,defaultTag:Xs,name:"Dialog.Panel"})}let Ks="h2";function Qs(t,s){let a=_(),{id:n=`headlessui-dialog-title-${a}`,...r}=t,[{dialogState:o,setTitleId:i}]=X("Dialog.Title"),l=L(s);c.useEffect(()=>(i(n),()=>i(null)),[n,i]);let d=c.useMemo(()=>({open:o===0}),[o]);return I({ourProps:{ref:l,id:n},theirProps:r,slot:d,defaultTag:Ks,name:"Dialog.Title"})}let Zs=E(Gs),ea=E(_s),ta=E(Js),sa=E(Ys),aa=E(Qs),ue=Object.assign(Zs,{Backdrop:ea,Panel:ta,Overlay:sa,Title:aa,Description:os});const na=_t("relative bg-white rounded-lg shadow-xl-premium transform transition-all",{variants:{size:{sm:"w-full max-w-md",md:"w-full max-w-2xl",lg:"w-full max-w-4xl"},scrollable:{true:"flex flex-col max-h-[85vh]",false:""}},defaultVariants:{size:"md",scrollable:!1}}),ra=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M6 18L18 6M6 6l12 12"})}),h=p.forwardRef(({isOpen:t,onClose:s,title:a,children:n,footer:r,size:o="md",scrollable:i=!1,closeOnBackdropClick:l=!0,icon:d,className:g},f)=>{const O=()=>{l&&s()};return e.jsx(ie,{show:t,as:p.Fragment,children:e.jsxs(ue,{as:"div",className:"relative z-50",onClose:O,initialFocus:void 0,children:[e.jsx(ie.Child,{as:p.Fragment,enter:"ease-out duration-300",enterFrom:"opacity-0",enterTo:"opacity-100",leave:"ease-in duration-200",leaveFrom:"opacity-100",leaveTo:"opacity-0",children:e.jsx("div",{className:"fixed inset-0 bg-black/50 backdrop-blur-sm"})}),e.jsx("div",{className:"fixed inset-0 overflow-y-auto",children:e.jsx("div",{className:"flex min-h-full items-center justify-center p-4",children:e.jsx(ie.Child,{as:p.Fragment,enter:"ease-out duration-300",enterFrom:"opacity-0 scale-95 translate-y-4",enterTo:"opacity-100 scale-100 translate-y-0",leave:"ease-in duration-200",leaveFrom:"opacity-100 scale-100 translate-y-0",leaveTo:"opacity-0 scale-95 translate-y-4",children:e.jsxs(ue.Panel,{ref:f,className:se(na({size:o,scrollable:i}),g),children:[(a||d)&&e.jsx("div",{className:se("px-6 py-5 border-b border-gray-200",i&&"flex-shrink-0"),children:e.jsxs("div",{className:"flex items-start justify-between gap-4",children:[e.jsxs("div",{className:"flex items-start gap-3 flex-1 min-w-0",children:[d&&e.jsx("div",{className:"flex-shrink-0 w-6 h-6 text-trust-deep mt-0.5",children:d}),a&&e.jsx(ue.Title,{as:"h3",className:"text-lg font-semibold text-gray-900 leading-6",children:a})]}),e.jsx("button",{type:"button",onClick:s,"aria-label":"Close modal",className:"flex-shrink-0 inline-flex items-center justify-center w-8 h-8 text-gray-400 hover:text-gray-500 hover:bg-gray-100 rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-trust-deep focus:ring-offset-2",children:e.jsx(ra,{})})]})}),e.jsx("div",{className:se("px-6 py-5",i?"overflow-y-auto flex-1":"overflow-visible",!a&&!d&&"pt-6"),children:n}),r&&e.jsx("div",{className:se("px-6 py-4 border-t border-gray-200 bg-gray-50 rounded-b-lg",i&&"flex-shrink-0"),children:r})]})})})})]})})});h.displayName="Modal";h.__docgenInfo={description:`Modal component for displaying overlay dialogs.
Uses Headless UI Dialog for full accessibility and focus management.

@example
\`\`\`tsx
<Modal
  isOpen={isOpen}
  onClose={() => setIsOpen(false)}
  title="Confirm Action"
>
  Are you sure you want to proceed?
</Modal>

<Modal
  isOpen={isOpen}
  onClose={() => setIsOpen(false)}
  title="User Profile"
  size="lg"
  footer={
    <div className="flex gap-3 justify-end">
      <Button variant="outline" onClick={() => setIsOpen(false)}>
        Cancel
      </Button>
      <Button variant="primary" onClick={handleSave}>
        Save Changes
      </Button>
    </div>
  }
>
  <form>...</form>
</Modal>
\`\`\``,methods:[],displayName:"Modal",props:{isOpen:{required:!0,tsType:{name:"boolean"},description:"Whether the modal is open"},onClose:{required:!0,tsType:{name:"signature",type:"function",raw:"() => void",signature:{arguments:[],return:{name:"void"}}},description:"Callback when modal should close"},title:{required:!1,tsType:{name:"string"},description:"Optional modal title"},children:{required:!0,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Modal content"},footer:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Optional footer content (typically buttons)"},size:{required:!1,tsType:{name:"union",raw:"'sm' | 'md' | 'lg'",elements:[{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"},{name:"literal",value:"'lg'"}]},description:"Modal size",defaultValue:{value:"'md'",computed:!1}},scrollable:{required:!1,tsType:{name:"boolean"},description:"Enable scrollable content area",defaultValue:{value:"false",computed:!1}},closeOnBackdropClick:{required:!1,tsType:{name:"boolean"},description:"Whether clicking backdrop closes the modal",defaultValue:{value:"true",computed:!1}},icon:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Optional icon to display in header"},className:{required:!1,tsType:{name:"string"},description:"Additional class names for the modal panel"}}};const ka={title:"Design System/Overlays/Modal",component:h,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{isOpen:{control:"boolean",description:"Whether the modal is open"},title:{control:"text",description:"Optional modal title"},size:{control:"select",options:["sm","md","lg"],description:"Modal size variant"},scrollable:{control:"boolean",description:"Enable scrollable content area"},closeOnBackdropClick:{control:"boolean",description:"Whether clicking backdrop closes the modal"},children:{control:"text",description:"Modal content"}}},z={render:t=>{const[s,a]=c.useState(!1);return e.jsxs(e.Fragment,{children:[e.jsx(u,{onClick:()=>a(!0),children:"Open Modal"}),e.jsx(h,{...t,isOpen:s,onClose:()=>a(!1),children:e.jsx("p",{className:"text-gray-700",children:"This is a basic modal with a title and content. Click the X button or press ESC to close."})})]})},args:{title:"Basic Modal",size:"md",closeOnBackdropClick:!0,scrollable:!1}},F={render:()=>{const[t,s]=c.useState(!1);return e.jsxs(e.Fragment,{children:[e.jsx(u,{onClick:()=>s(!0),children:"Open Modal with Actions"}),e.jsxs(h,{isOpen:t,onClose:()=>s(!1),title:"Confirm Your Action",footer:e.jsxs("div",{className:"flex gap-3 justify-end",children:[e.jsx(u,{variant:"outline",onClick:()=>s(!1),children:"Cancel"}),e.jsx(u,{variant:"primary",onClick:()=>{alert("Action confirmed!"),s(!1)},children:"Confirm"})]}),children:[e.jsx("p",{className:"text-gray-700",children:"Modals with footer actions are the most common use case. The footer is styled with a gray background to separate it from the main content."}),e.jsx("p",{className:"text-gray-600 mt-3 text-sm",children:"Primary actions should be on the right, with cancel/secondary actions on the left."})]})]})},args:{title:"Modal with Actions",size:"md"}},$={render:()=>{const[t,s]=c.useState(null);return e.jsxs("div",{className:"flex gap-3",children:[e.jsx(u,{onClick:()=>s("sm"),children:"Small Modal"}),e.jsx(u,{onClick:()=>s("md"),children:"Medium Modal"}),e.jsx(u,{onClick:()=>s("lg"),children:"Large Modal"}),e.jsx(h,{isOpen:t==="sm",onClose:()=>s(null),title:"Small Modal",size:"sm",footer:e.jsx("div",{className:"flex justify-end",children:e.jsx(u,{onClick:()=>s(null),children:"Close"})}),children:e.jsx("p",{className:"text-gray-700",children:"Small modals (400px max-width) are perfect for simple confirmations and short messages."})}),e.jsx(h,{isOpen:t==="md",onClose:()=>s(null),title:"Medium Modal",size:"md",footer:e.jsx("div",{className:"flex justify-end",children:e.jsx(u,{onClick:()=>s(null),children:"Close"})}),children:e.jsx("p",{className:"text-gray-700",children:"Medium modals (600px max-width) are the default size, suitable for most forms and content."})}),e.jsxs(h,{isOpen:t==="lg",onClose:()=>s(null),title:"Large Modal",size:"lg",footer:e.jsx("div",{className:"flex justify-end",children:e.jsx(u,{onClick:()=>s(null),children:"Close"})}),children:[e.jsx("p",{className:"text-gray-700",children:"Large modals (800px max-width) are ideal for complex forms, detailed content, or data tables."}),e.jsxs("div",{className:"mt-4 p-4 bg-gray-50 rounded-lg border border-gray-200",children:[e.jsx("h4",{className:"font-semibold text-gray-900 mb-2",children:"Additional Content"}),e.jsx("p",{className:"text-gray-600 text-sm",children:"Large modals provide more space for complex layouts and multiple sections."})]})]})]})},args:{title:"Size Variants",size:"md"}},P={render:()=>{const[t,s]=c.useState(!1),a=Array.from({length:20},(n,r)=>e.jsxs("p",{className:"text-gray-700 mb-3",children:["This is paragraph ",r+1," of long content. When content exceeds the viewport height, the modal body becomes scrollable while the header and footer remain fixed. This ensures important actions are always visible."]},r));return e.jsxs(e.Fragment,{children:[e.jsx(u,{onClick:()=>s(!0),children:"Open Scrollable Modal"}),e.jsx(h,{isOpen:t,onClose:()=>s(!1),title:"Terms and Conditions",size:"md",scrollable:!0,footer:e.jsxs("div",{className:"flex gap-3 justify-end",children:[e.jsx(u,{variant:"outline",onClick:()=>s(!1),children:"Decline"}),e.jsx(u,{variant:"primary",onClick:()=>{alert("Terms accepted!"),s(!1)},children:"Accept Terms"})]}),children:e.jsxs("div",{className:"prose prose-sm max-w-none",children:[e.jsx("p",{className:"text-gray-700 font-medium mb-4",children:"Please read and accept the following terms and conditions to continue."}),a]})})]})},args:{title:"Scrollable Modal",scrollable:!0}},W={render:()=>{const[t,s]=c.useState(!1);return e.jsxs(e.Fragment,{children:[e.jsx(u,{onClick:()=>s(!0),children:"Open Centered Modal"}),e.jsx(h,{isOpen:t,onClose:()=>s(!1),title:"Centered Modal",size:"md",children:e.jsxs("div",{className:"text-center py-6",children:[e.jsx("div",{className:"w-16 h-16 bg-trust-light rounded-full flex items-center justify-center mx-auto mb-4",children:e.jsx("svg",{className:"w-8 h-8 text-trust-deep",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"})})}),e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-2",children:"Action Successful"}),e.jsx("p",{className:"text-gray-600",children:"Your changes have been saved successfully. All updates are now live."})]})})]})},args:{title:"Centered Content",size:"sm"}},q={render:()=>{const[t,s]=c.useState(!1),[a,n]=c.useState(!1),r=()=>{n(!0),setTimeout(()=>{n(!1),s(!1),alert("Item deleted successfully")},1500)};return e.jsxs(e.Fragment,{children:[e.jsx(u,{variant:"danger",onClick:()=>s(!0),children:"Delete Item"}),e.jsxs(h,{isOpen:t,onClose:()=>s(!1),title:"Confirm Deletion",size:"sm",closeOnBackdropClick:!1,icon:e.jsx("svg",{className:"w-6 h-6 text-error-primary",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"})}),footer:e.jsxs("div",{className:"flex gap-3 justify-end",children:[e.jsx(u,{variant:"outline",onClick:()=>s(!1),disabled:a,children:"Cancel"}),e.jsx(u,{variant:"danger",onClick:r,isLoading:a,children:"Delete"})]}),children:[e.jsx("p",{className:"text-gray-700",children:"Are you sure you want to delete this item? This action cannot be undone."}),e.jsx("div",{className:"mt-4 p-3 bg-error-light border border-error-primary/20 rounded-md",children:e.jsx("p",{className:"text-sm text-error-dark font-medium",children:"Warning: This is a destructive action"})})]})]})},args:{title:"Confirmation",size:"sm"}},U={render:()=>{const[t,s]=c.useState(!1),[a,n]=c.useState(!1),r=o=>{o.preventDefault(),n(!0),setTimeout(()=>{n(!1),s(!1),alert("Form submitted successfully!")},1500)};return e.jsxs(e.Fragment,{children:[e.jsx(u,{onClick:()=>s(!0),children:"Create New User"}),e.jsx(h,{isOpen:t,onClose:()=>s(!1),title:"Create New User",size:"md",footer:e.jsxs("div",{className:"flex gap-3 justify-end",children:[e.jsx(u,{variant:"outline",onClick:()=>s(!1),disabled:a,children:"Cancel"}),e.jsx(u,{variant:"primary",type:"submit",form:"user-form",isLoading:a,children:"Create User"})]}),children:e.jsxs("form",{id:"user-form",onSubmit:r,className:"space-y-4",children:[e.jsxs("div",{children:[e.jsx("label",{htmlFor:"name",className:"block text-sm font-medium text-gray-700 mb-1",children:"Full Name"}),e.jsx("input",{type:"text",id:"name",name:"name",required:!0,className:"w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-trust-deep focus:border-transparent",placeholder:"John Doe"})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"email",className:"block text-sm font-medium text-gray-700 mb-1",children:"Email Address"}),e.jsx("input",{type:"email",id:"email",name:"email",required:!0,className:"w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-trust-deep focus:border-transparent",placeholder:"john.doe@example.com"})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"role",className:"block text-sm font-medium text-gray-700 mb-1",children:"Role"}),e.jsxs("select",{id:"role",name:"role",required:!0,className:"w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-trust-deep focus:border-transparent",children:[e.jsx("option",{value:"",children:"Select a role..."}),e.jsx("option",{value:"admin",children:"Administrator"}),e.jsx("option",{value:"user",children:"Standard User"}),e.jsx("option",{value:"viewer",children:"Viewer"})]})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"bio",className:"block text-sm font-medium text-gray-700 mb-1",children:"Bio (Optional)"}),e.jsx("textarea",{id:"bio",name:"bio",rows:3,className:"w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-trust-deep focus:border-transparent resize-none",placeholder:"Tell us about yourself..."})]})]})})]})},args:{title:"Form Modal",size:"md"}},G={render:()=>{const[t,s]=c.useState(!1),[a,n]=c.useState(!1);return e.jsxs("div",{className:"flex gap-3",children:[e.jsx(u,{onClick:()=>s(!0),children:"Dismissible on Backdrop Click"}),e.jsx(u,{onClick:()=>n(!0),children:"Non-Dismissible on Backdrop Click"}),e.jsxs(h,{isOpen:t,onClose:()=>s(!1),title:"Dismissible Modal",size:"sm",closeOnBackdropClick:!0,children:[e.jsx("p",{className:"text-gray-700 mb-3",children:"This modal can be dismissed by clicking the backdrop (dark area outside the modal)."}),e.jsx("p",{className:"text-sm text-gray-600",children:"Try clicking outside this box or pressing ESC."})]}),e.jsxs(h,{isOpen:a,onClose:()=>n(!1),title:"Non-Dismissible Modal",size:"sm",closeOnBackdropClick:!1,footer:e.jsx("div",{className:"flex justify-end",children:e.jsx(u,{onClick:()=>n(!1),children:"Close"})}),children:[e.jsx("p",{className:"text-gray-700 mb-3",children:"This modal cannot be dismissed by clicking the backdrop. You must use the close button or press ESC."}),e.jsx("div",{className:"p-3 bg-warning-light border border-warning-primary/20 rounded-md",children:e.jsx("p",{className:"text-sm text-warning-dark",children:"Use this for critical actions that require explicit user acknowledgment."})})]})]})},args:{title:"Backdrop Interaction",closeOnBackdropClick:!0}},V={render:t=>{const[s,a]=c.useState(!1);return e.jsxs(e.Fragment,{children:[e.jsx(u,{onClick:()=>a(!0),children:"Open Playground Modal"}),e.jsx(h,{...t,isOpen:s,onClose:()=>a(!1),footer:t.footer||e.jsxs("div",{className:"flex gap-3 justify-end",children:[e.jsx(u,{variant:"outline",onClick:()=>a(!1),children:"Cancel"}),e.jsx(u,{variant:"primary",onClick:()=>a(!1),children:"Confirm"})]}),children:t.children||e.jsxs("div",{children:[e.jsx("p",{className:"text-gray-700 mb-3",children:"Use the controls below to customize this modal's appearance and behavior."}),e.jsxs("ul",{className:"list-disc list-inside text-sm text-gray-600 space-y-1",children:[e.jsx("li",{children:"Change the size (sm, md, lg)"}),e.jsx("li",{children:"Toggle scrollable content"}),e.jsx("li",{children:"Control backdrop click behavior"}),e.jsx("li",{children:"Add or remove the title"})]})]})})]})},args:{title:"Playground Modal",size:"md",scrollable:!1,closeOnBackdropClick:!0}},Y={render:()=>{const[t,s]=c.useState(!1),[a,n]=c.useState(!1),[r,o]=c.useState(!1);return e.jsxs("div",{className:"flex flex-col gap-3",children:[e.jsx(u,{onClick:()=>s(!0),children:"Grant Consent Request"}),e.jsx(u,{variant:"danger",onClick:()=>n(!0),children:"Revoke Consent"}),e.jsx(u,{variant:"outline",onClick:()=>o(!0),children:"View Consent Details"}),e.jsx(h,{isOpen:t,onClose:()=>s(!1),title:"Grant Consent",size:"md",icon:e.jsx("svg",{className:"w-6 h-6 text-trust-deep",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"})}),footer:e.jsxs("div",{className:"flex gap-3 justify-end",children:[e.jsx(u,{variant:"outline",onClick:()=>s(!1),children:"Cancel"}),e.jsx(u,{variant:"primary",onClick:()=>{alert("Consent granted!"),s(!1)},children:"Grant Access"})]}),children:e.jsxs("div",{className:"space-y-4",children:[e.jsxs("p",{className:"text-gray-700",children:[e.jsx("strong",{children:"Analytics Dashboard"})," is requesting access to the following data:"]}),e.jsxs("ul",{className:"list-disc list-inside text-gray-700 space-y-2 pl-2",children:[e.jsx("li",{children:"Basic profile information (name, email)"}),e.jsx("li",{children:"Usage statistics and activity logs"}),e.jsx("li",{children:"Preference settings"})]}),e.jsx("div",{className:"p-4 bg-info-light border border-info-primary/20 rounded-md",children:e.jsx("p",{className:"text-sm text-info-dark",children:"This permission will be valid for 30 days and can be revoked at any time."})})]})}),e.jsxs(h,{isOpen:a,onClose:()=>n(!1),title:"Revoke Consent",size:"sm",closeOnBackdropClick:!1,icon:e.jsx("svg",{className:"w-6 h-6 text-error-primary",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"})}),footer:e.jsxs("div",{className:"flex gap-3 justify-end",children:[e.jsx(u,{variant:"outline",onClick:()=>n(!1),children:"Keep Consent"}),e.jsx(u,{variant:"danger",onClick:()=>{alert("Consent revoked"),n(!1)},children:"Revoke Access"})]}),children:[e.jsxs("p",{className:"text-gray-700 mb-3",children:["Are you sure you want to revoke consent for ",e.jsx("strong",{children:"Marketing Platform"}),"?"]}),e.jsx("p",{className:"text-sm text-gray-600",children:"This application will immediately lose access to your data and may stop functioning properly."})]}),e.jsx(h,{isOpen:r,onClose:()=>o(!1),title:"Consent Details",size:"lg",scrollable:!0,footer:e.jsx("div",{className:"flex justify-end",children:e.jsx(u,{onClick:()=>o(!1),children:"Close"})}),children:e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Application"}),e.jsx("p",{className:"text-gray-700",children:"Analytics Dashboard v2.1"})]}),e.jsxs("div",{children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Granted Permissions"}),e.jsxs("ul",{className:"list-disc list-inside text-gray-700 space-y-1",children:[e.jsx("li",{children:"Read basic profile information"}),e.jsx("li",{children:"Access usage statistics"}),e.jsx("li",{children:"View activity logs (last 90 days)"})]})]}),e.jsxs("div",{children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Consent Timeline"}),e.jsxs("div",{className:"space-y-3",children:[e.jsxs("div",{className:"flex items-start gap-3 p-3 bg-gray-50 rounded-md",children:[e.jsx("div",{className:"text-xs text-gray-500 mt-0.5",children:"Dec 1, 2025"}),e.jsxs("div",{className:"flex-1",children:[e.jsx("p",{className:"text-sm text-gray-900 font-medium",children:"Consent Granted"}),e.jsx("p",{className:"text-xs text-gray-600",children:"Initial access granted for 30 days"})]})]}),e.jsxs("div",{className:"flex items-start gap-3 p-3 bg-gray-50 rounded-md",children:[e.jsx("div",{className:"text-xs text-gray-500 mt-0.5",children:"Dec 15, 2025"}),e.jsxs("div",{className:"flex-1",children:[e.jsx("p",{className:"text-sm text-gray-900 font-medium",children:"Permissions Updated"}),e.jsx("p",{className:"text-xs text-gray-600",children:"Added access to activity logs"})]})]})]})]}),e.jsxs("div",{className:"p-4 bg-warning-light border border-warning-primary/20 rounded-md",children:[e.jsx("p",{className:"text-sm text-warning-dark font-medium mb-1",children:"Expires in 15 days"}),e.jsx("p",{className:"text-xs text-gray-700",children:"This consent will expire on December 31, 2025. You will need to renew access after this date."})]})]})})]})},args:{title:"Consent Management"}};var Be,Se,De,Te,Ee;z.parameters={...z.parameters,docs:{...(Be=z.parameters)==null?void 0:Be.docs,source:{originalSource:`{
  render: args => {
    const [isOpen, setIsOpen] = useState(false);
    return <>
        <Button onClick={() => setIsOpen(true)}>Open Modal</Button>
        <Modal {...args} isOpen={isOpen} onClose={() => setIsOpen(false)}>
          <p className="text-gray-700">
            This is a basic modal with a title and content. Click the X button or press ESC to close.
          </p>
        </Modal>
      </>;
  },
  args: {
    title: 'Basic Modal',
    size: 'md',
    closeOnBackdropClick: true,
    scrollable: false
  }
}`,...(De=(Se=z.parameters)==null?void 0:Se.docs)==null?void 0:De.source},description:{story:"Default modal with title and basic content",...(Ee=(Te=z.parameters)==null?void 0:Te.docs)==null?void 0:Ee.description}}};var Le,Ie,Ae,Re,ze;F.parameters={...F.parameters,docs:{...(Le=F.parameters)==null?void 0:Le.docs,source:{originalSource:`{
  render: () => {
    const [isOpen, setIsOpen] = useState(false);
    return <>
        <Button onClick={() => setIsOpen(true)}>Open Modal with Actions</Button>
        <Modal isOpen={isOpen} onClose={() => setIsOpen(false)} title="Confirm Your Action" footer={<div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setIsOpen(false)}>
                Cancel
              </Button>
              <Button variant="primary" onClick={() => {
          alert('Action confirmed!');
          setIsOpen(false);
        }}>
                Confirm
              </Button>
            </div>}>
          <p className="text-gray-700">
            Modals with footer actions are the most common use case. The footer is
            styled with a gray background to separate it from the main content.
          </p>
          <p className="text-gray-600 mt-3 text-sm">
            Primary actions should be on the right, with cancel/secondary actions on the left.
          </p>
        </Modal>
      </>;
  },
  args: {
    title: 'Modal with Actions',
    size: 'md'
  }
}`,...(Ae=(Ie=F.parameters)==null?void 0:Ie.docs)==null?void 0:Ae.source},description:{story:"Modal with footer actions (primary use case)",...(ze=(Re=F.parameters)==null?void 0:Re.docs)==null?void 0:ze.description}}};var Fe,$e,Pe,We,qe;$.parameters={...$.parameters,docs:{...(Fe=$.parameters)==null?void 0:Fe.docs,source:{originalSource:`{
  render: () => {
    const [openModal, setOpenModal] = useState<'sm' | 'md' | 'lg' | null>(null);
    return <div className="flex gap-3">
        <Button onClick={() => setOpenModal('sm')}>Small Modal</Button>
        <Button onClick={() => setOpenModal('md')}>Medium Modal</Button>
        <Button onClick={() => setOpenModal('lg')}>Large Modal</Button>

        <Modal isOpen={openModal === 'sm'} onClose={() => setOpenModal(null)} title="Small Modal" size="sm" footer={<div className="flex justify-end">
              <Button onClick={() => setOpenModal(null)}>Close</Button>
            </div>}>
          <p className="text-gray-700">
            Small modals (400px max-width) are perfect for simple confirmations and short messages.
          </p>
        </Modal>

        <Modal isOpen={openModal === 'md'} onClose={() => setOpenModal(null)} title="Medium Modal" size="md" footer={<div className="flex justify-end">
              <Button onClick={() => setOpenModal(null)}>Close</Button>
            </div>}>
          <p className="text-gray-700">
            Medium modals (600px max-width) are the default size, suitable for most forms and content.
          </p>
        </Modal>

        <Modal isOpen={openModal === 'lg'} onClose={() => setOpenModal(null)} title="Large Modal" size="lg" footer={<div className="flex justify-end">
              <Button onClick={() => setOpenModal(null)}>Close</Button>
            </div>}>
          <p className="text-gray-700">
            Large modals (800px max-width) are ideal for complex forms, detailed content, or data tables.
          </p>
          <div className="mt-4 p-4 bg-gray-50 rounded-lg border border-gray-200">
            <h4 className="font-semibold text-gray-900 mb-2">Additional Content</h4>
            <p className="text-gray-600 text-sm">
              Large modals provide more space for complex layouts and multiple sections.
            </p>
          </div>
        </Modal>
      </div>;
  },
  args: {
    title: 'Size Variants',
    size: 'md'
  }
}`,...(Pe=($e=$.parameters)==null?void 0:$e.docs)==null?void 0:Pe.source},description:{story:"Different modal sizes: small, medium, large",...(qe=(We=$.parameters)==null?void 0:We.docs)==null?void 0:qe.description}}};var Ue,Ge,Ve,Ye,He;P.parameters={...P.parameters,docs:{...(Ue=P.parameters)==null?void 0:Ue.docs,source:{originalSource:`{
  render: () => {
    const [isOpen, setIsOpen] = useState(false);
    const longContent = Array.from({
      length: 20
    }, (_, i) => <p key={i} className="text-gray-700 mb-3">
        This is paragraph {i + 1} of long content. When content exceeds the viewport height,
        the modal body becomes scrollable while the header and footer remain fixed. This
        ensures important actions are always visible.
      </p>);
    return <>
        <Button onClick={() => setIsOpen(true)}>Open Scrollable Modal</Button>
        <Modal isOpen={isOpen} onClose={() => setIsOpen(false)} title="Terms and Conditions" size="md" scrollable footer={<div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setIsOpen(false)}>
                Decline
              </Button>
              <Button variant="primary" onClick={() => {
          alert('Terms accepted!');
          setIsOpen(false);
        }}>
                Accept Terms
              </Button>
            </div>}>
          <div className="prose prose-sm max-w-none">
            <p className="text-gray-700 font-medium mb-4">
              Please read and accept the following terms and conditions to continue.
            </p>
            {longContent}
          </div>
        </Modal>
      </>;
  },
  args: {
    title: 'Scrollable Modal',
    scrollable: true
  }
}`,...(Ve=(Ge=P.parameters)==null?void 0:Ge.docs)==null?void 0:Ve.source},description:{story:"Scrollable content for long modal bodies",...(He=(Ye=P.parameters)==null?void 0:Ye.docs)==null?void 0:He.description}}};var _e,Xe,Je,Ke,Qe;W.parameters={...W.parameters,docs:{...(_e=W.parameters)==null?void 0:_e.docs,source:{originalSource:`{
  render: () => {
    const [isOpen, setIsOpen] = useState(false);
    return <>
        <Button onClick={() => setIsOpen(true)}>Open Centered Modal</Button>
        <Modal isOpen={isOpen} onClose={() => setIsOpen(false)} title="Centered Modal" size="md">
          <div className="text-center py-6">
            <div className="w-16 h-16 bg-trust-light rounded-full flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-trust-deep" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              Action Successful
            </h3>
            <p className="text-gray-600">
              Your changes have been saved successfully. All updates are now live.
            </p>
          </div>
        </Modal>
      </>;
  },
  args: {
    title: 'Centered Content',
    size: 'sm'
  }
}`,...(Je=(Xe=W.parameters)==null?void 0:Xe.docs)==null?void 0:Je.source},description:{story:"Centered modal layout (default behavior)",...(Qe=(Ke=W.parameters)==null?void 0:Ke.docs)==null?void 0:Qe.description}}};var Ze,et,tt,st,at;q.parameters={...q.parameters,docs:{...(Ze=q.parameters)==null?void 0:Ze.docs,source:{originalSource:`{
  render: () => {
    const [isOpen, setIsOpen] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const handleDelete = () => {
      setIsDeleting(true);
      // Simulate async action
      setTimeout(() => {
        setIsDeleting(false);
        setIsOpen(false);
        alert('Item deleted successfully');
      }, 1500);
    };
    return <>
        <Button variant="danger" onClick={() => setIsOpen(true)}>
          Delete Item
        </Button>
        <Modal isOpen={isOpen} onClose={() => setIsOpen(false)} title="Confirm Deletion" size="sm" closeOnBackdropClick={false} icon={<svg className="w-6 h-6 text-error-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>} footer={<div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setIsOpen(false)} disabled={isDeleting}>
                Cancel
              </Button>
              <Button variant="danger" onClick={handleDelete} isLoading={isDeleting}>
                Delete
              </Button>
            </div>}>
          <p className="text-gray-700">
            Are you sure you want to delete this item? This action cannot be undone.
          </p>
          <div className="mt-4 p-3 bg-error-light border border-error-primary/20 rounded-md">
            <p className="text-sm text-error-dark font-medium">
              Warning: This is a destructive action
            </p>
          </div>
        </Modal>
      </>;
  },
  args: {
    title: 'Confirmation',
    size: 'sm'
  }
}`,...(tt=(et=q.parameters)==null?void 0:et.docs)==null?void 0:tt.source},description:{story:"Confirmation modal pattern (common use case)",...(at=(st=q.parameters)==null?void 0:st.docs)==null?void 0:at.description}}};var nt,rt,ot,lt,it;U.parameters={...U.parameters,docs:{...(nt=U.parameters)==null?void 0:nt.docs,source:{originalSource:`{
  render: () => {
    const [isOpen, setIsOpen] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const handleSubmit = (e: React.FormEvent) => {
      e.preventDefault();
      setIsSubmitting(true);
      // Simulate async submission
      setTimeout(() => {
        setIsSubmitting(false);
        setIsOpen(false);
        alert('Form submitted successfully!');
      }, 1500);
    };
    return <>
        <Button onClick={() => setIsOpen(true)}>Create New User</Button>
        <Modal isOpen={isOpen} onClose={() => setIsOpen(false)} title="Create New User" size="md" footer={<div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setIsOpen(false)} disabled={isSubmitting}>
                Cancel
              </Button>
              <Button variant="primary" type="submit" form="user-form" isLoading={isSubmitting}>
                Create User
              </Button>
            </div>}>
          <form id="user-form" onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                Full Name
              </label>
              <input type="text" id="name" name="name" required className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-trust-deep focus:border-transparent" placeholder="John Doe" />
            </div>

            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                Email Address
              </label>
              <input type="email" id="email" name="email" required className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-trust-deep focus:border-transparent" placeholder="john.doe@example.com" />
            </div>

            <div>
              <label htmlFor="role" className="block text-sm font-medium text-gray-700 mb-1">
                Role
              </label>
              <select id="role" name="role" required className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-trust-deep focus:border-transparent">
                <option value="">Select a role...</option>
                <option value="admin">Administrator</option>
                <option value="user">Standard User</option>
                <option value="viewer">Viewer</option>
              </select>
            </div>

            <div>
              <label htmlFor="bio" className="block text-sm font-medium text-gray-700 mb-1">
                Bio (Optional)
              </label>
              <textarea id="bio" name="bio" rows={3} className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-trust-deep focus:border-transparent resize-none" placeholder="Tell us about yourself..." />
            </div>
          </form>
        </Modal>
      </>;
  },
  args: {
    title: 'Form Modal',
    size: 'md'
  }
}`,...(ot=(rt=U.parameters)==null?void 0:rt.docs)==null?void 0:ot.source},description:{story:"Form modal with input fields",...(it=(lt=U.parameters)==null?void 0:lt.docs)==null?void 0:it.description}}};var dt,ct,ut,mt,pt;G.parameters={...G.parameters,docs:{...(dt=G.parameters)==null?void 0:dt.docs,source:{originalSource:`{
  render: () => {
    const [dismissible, setDismissible] = useState(false);
    const [nonDismissible, setNonDismissible] = useState(false);
    return <div className="flex gap-3">
        <Button onClick={() => setDismissible(true)}>
          Dismissible on Backdrop Click
        </Button>
        <Button onClick={() => setNonDismissible(true)}>
          Non-Dismissible on Backdrop Click
        </Button>

        <Modal isOpen={dismissible} onClose={() => setDismissible(false)} title="Dismissible Modal" size="sm" closeOnBackdropClick={true}>
          <p className="text-gray-700 mb-3">
            This modal can be dismissed by clicking the backdrop (dark area outside the modal).
          </p>
          <p className="text-sm text-gray-600">
            Try clicking outside this box or pressing ESC.
          </p>
        </Modal>

        <Modal isOpen={nonDismissible} onClose={() => setNonDismissible(false)} title="Non-Dismissible Modal" size="sm" closeOnBackdropClick={false} footer={<div className="flex justify-end">
              <Button onClick={() => setNonDismissible(false)}>
                Close
              </Button>
            </div>}>
          <p className="text-gray-700 mb-3">
            This modal cannot be dismissed by clicking the backdrop. You must use the close button or press ESC.
          </p>
          <div className="p-3 bg-warning-light border border-warning-primary/20 rounded-md">
            <p className="text-sm text-warning-dark">
              Use this for critical actions that require explicit user acknowledgment.
            </p>
          </div>
        </Modal>
      </div>;
  },
  args: {
    title: 'Backdrop Interaction',
    closeOnBackdropClick: true
  }
}`,...(ut=(ct=G.parameters)==null?void 0:ct.docs)==null?void 0:ut.source},description:{story:"Backdrop interaction behavior",...(pt=(mt=G.parameters)==null?void 0:mt.docs)==null?void 0:pt.description}}};var ft,ht,gt,xt,vt;V.parameters={...V.parameters,docs:{...(ft=V.parameters)==null?void 0:ft.docs,source:{originalSource:`{
  render: args => {
    const [isOpen, setIsOpen] = useState(false);
    return <>
        <Button onClick={() => setIsOpen(true)}>Open Playground Modal</Button>
        <Modal {...args} isOpen={isOpen} onClose={() => setIsOpen(false)} footer={args.footer || <div className="flex gap-3 justify-end">
                <Button variant="outline" onClick={() => setIsOpen(false)}>
                  Cancel
                </Button>
                <Button variant="primary" onClick={() => setIsOpen(false)}>
                  Confirm
                </Button>
              </div>}>
          {args.children || <div>
              <p className="text-gray-700 mb-3">
                Use the controls below to customize this modal's appearance and behavior.
              </p>
              <ul className="list-disc list-inside text-sm text-gray-600 space-y-1">
                <li>Change the size (sm, md, lg)</li>
                <li>Toggle scrollable content</li>
                <li>Control backdrop click behavior</li>
                <li>Add or remove the title</li>
              </ul>
            </div>}
        </Modal>
      </>;
  },
  args: {
    title: 'Playground Modal',
    size: 'md',
    scrollable: false,
    closeOnBackdropClick: true
  }
}`,...(gt=(ht=V.parameters)==null?void 0:ht.docs)==null?void 0:gt.source},description:{story:"Interactive playground with all controls",...(vt=(xt=V.parameters)==null?void 0:xt.docs)==null?void 0:vt.description}}};var yt,bt,Ct,kt,jt;Y.parameters={...Y.parameters,docs:{...(yt=Y.parameters)==null?void 0:yt.docs,source:{originalSource:`{
  render: () => {
    const [grantModal, setGrantModal] = useState(false);
    const [revokeModal, setRevokeModal] = useState(false);
    const [detailsModal, setDetailsModal] = useState(false);
    return <div className="flex flex-col gap-3">
        <Button onClick={() => setGrantModal(true)}>
          Grant Consent Request
        </Button>
        <Button variant="danger" onClick={() => setRevokeModal(true)}>
          Revoke Consent
        </Button>
        <Button variant="outline" onClick={() => setDetailsModal(true)}>
          View Consent Details
        </Button>

        {/* Grant Consent Modal */}
        <Modal isOpen={grantModal} onClose={() => setGrantModal(false)} title="Grant Consent" size="md" icon={<svg className="w-6 h-6 text-trust-deep" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>} footer={<div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setGrantModal(false)}>
                Cancel
              </Button>
              <Button variant="primary" onClick={() => {
          alert('Consent granted!');
          setGrantModal(false);
        }}>
                Grant Access
              </Button>
            </div>}>
          <div className="space-y-4">
            <p className="text-gray-700">
              <strong>Analytics Dashboard</strong> is requesting access to the following data:
            </p>
            <ul className="list-disc list-inside text-gray-700 space-y-2 pl-2">
              <li>Basic profile information (name, email)</li>
              <li>Usage statistics and activity logs</li>
              <li>Preference settings</li>
            </ul>
            <div className="p-4 bg-info-light border border-info-primary/20 rounded-md">
              <p className="text-sm text-info-dark">
                This permission will be valid for 30 days and can be revoked at any time.
              </p>
            </div>
          </div>
        </Modal>

        {/* Revoke Consent Modal */}
        <Modal isOpen={revokeModal} onClose={() => setRevokeModal(false)} title="Revoke Consent" size="sm" closeOnBackdropClick={false} icon={<svg className="w-6 h-6 text-error-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>} footer={<div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setRevokeModal(false)}>
                Keep Consent
              </Button>
              <Button variant="danger" onClick={() => {
          alert('Consent revoked');
          setRevokeModal(false);
        }}>
                Revoke Access
              </Button>
            </div>}>
          <p className="text-gray-700 mb-3">
            Are you sure you want to revoke consent for <strong>Marketing Platform</strong>?
          </p>
          <p className="text-sm text-gray-600">
            This application will immediately lose access to your data and may stop functioning properly.
          </p>
        </Modal>

        {/* Consent Details Modal */}
        <Modal isOpen={detailsModal} onClose={() => setDetailsModal(false)} title="Consent Details" size="lg" scrollable footer={<div className="flex justify-end">
              <Button onClick={() => setDetailsModal(false)}>Close</Button>
            </div>}>
          <div className="space-y-6">
            <div>
              <h4 className="text-sm font-semibold text-gray-900 mb-2">Application</h4>
              <p className="text-gray-700">Analytics Dashboard v2.1</p>
            </div>

            <div>
              <h4 className="text-sm font-semibold text-gray-900 mb-2">Granted Permissions</h4>
              <ul className="list-disc list-inside text-gray-700 space-y-1">
                <li>Read basic profile information</li>
                <li>Access usage statistics</li>
                <li>View activity logs (last 90 days)</li>
              </ul>
            </div>

            <div>
              <h4 className="text-sm font-semibold text-gray-900 mb-2">Consent Timeline</h4>
              <div className="space-y-3">
                <div className="flex items-start gap-3 p-3 bg-gray-50 rounded-md">
                  <div className="text-xs text-gray-500 mt-0.5">Dec 1, 2025</div>
                  <div className="flex-1">
                    <p className="text-sm text-gray-900 font-medium">Consent Granted</p>
                    <p className="text-xs text-gray-600">Initial access granted for 30 days</p>
                  </div>
                </div>
                <div className="flex items-start gap-3 p-3 bg-gray-50 rounded-md">
                  <div className="text-xs text-gray-500 mt-0.5">Dec 15, 2025</div>
                  <div className="flex-1">
                    <p className="text-sm text-gray-900 font-medium">Permissions Updated</p>
                    <p className="text-xs text-gray-600">Added access to activity logs</p>
                  </div>
                </div>
              </div>
            </div>

            <div className="p-4 bg-warning-light border border-warning-primary/20 rounded-md">
              <p className="text-sm text-warning-dark font-medium mb-1">Expires in 15 days</p>
              <p className="text-xs text-gray-700">
                This consent will expire on December 31, 2025. You will need to renew access after this date.
              </p>
            </div>
          </div>
        </Modal>
      </div>;
  },
  args: {
    title: 'Consent Management'
  }
}`,...(Ct=(bt=Y.parameters)==null?void 0:bt.docs)==null?void 0:Ct.source},description:{story:"Real-world consent management examples",...(jt=(kt=Y.parameters)==null?void 0:kt.docs)==null?void 0:jt.description}}};const ja=["Default","WithFooterActions","Sizes","ScrollableContent","CenteredLayout","Confirmation","FormModal","BackdropInteraction","Playground","ConsentManagementExamples"];export{G as BackdropInteraction,W as CenteredLayout,q as Confirmation,Y as ConsentManagementExamples,z as Default,U as FormModal,V as Playground,P as ScrollableContent,$ as Sizes,F as WithFooterActions,ja as __namedExportsOrder,ka as default};
