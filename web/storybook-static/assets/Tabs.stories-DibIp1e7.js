import{j as e}from"./jsx-runtime-BYYWji4R.js";import{r as d,R as A}from"./index-ClcD9ViR.js";import{c as te,a as ne}from"./cn-JCLedEej.js";import{U as $,I as Oe,y as U,l as ee,o as C,C as G,d as F,u as O,c as w,t as Be,O as ce}from"./keyboard-DNQdDreo.js";import{T as Je}from"./use-resolve-button-type-Bw633SU1.js";import{a as Ke,I as V,N as H,o as Ye,O as E,M as k}from"./use-is-mounted-62Q7yJH9.js";import{u as $e,s as Xe}from"./hidden-BnOBa8gi.js";import"./_commonjsHelpers-Cpj98o6Y.js";function Qe({onFocus:a}){let[s,t]=d.useState(!0),r=Ke();return s?A.createElement($e,{as:"button",type:"button",features:Xe.Focusable,onFocus:n=>{n.preventDefault();let l,o=50;function x(){if(o--<=0){l&&cancelAnimationFrame(l);return}if(a()){if(cancelAnimationFrame(l),!r.current)return;t(!1);return}l=requestAnimationFrame(x)}l=requestAnimationFrame(x)}}):null}const Ue=d.createContext(null);function Ze(){return{groups:new Map,get(a,s){var t;let r=this.groups.get(a);r||(r=new Map,this.groups.set(a,r));let n=(t=r.get(s))!=null?t:0;r.set(s,n+1);let l=Array.from(r.keys()).indexOf(s);function o(){let x=r.get(s);x>1?r.set(s,x-1):r.delete(s)}return[l,o]}}}function ea({children:a}){let s=d.useRef(Ze());return d.createElement(Ue.Provider,{value:s},a)}function Ge(a){let s=d.useContext(Ue);if(!s)throw new Error("You must wrap your component in a <StableCollection>");let t=aa(),[r,n]=s.current.get(a,t);return d.useEffect(()=>n,[]),r}function aa(){var a,s,t;let r=(t=(s=(a=d.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED)==null?void 0:a.ReactCurrentOwner)==null?void 0:s.current)!=null?t:null;if(!r)return Symbol();let n=[],l=r;for(;l;)n.push(l.index),l=l.return;return"$."+n.join(".")}var sa=(a=>(a[a.Forwards=0]="Forwards",a[a.Backwards=1]="Backwards",a))(sa||{}),ta=(a=>(a[a.Less=-1]="Less",a[a.Equal=0]="Equal",a[a.Greater=1]="Greater",a))(ta||{}),ra=(a=>(a[a.SetSelectedIndex=0]="SetSelectedIndex",a[a.RegisterTab=1]="RegisterTab",a[a.UnregisterTab=2]="UnregisterTab",a[a.RegisterPanel=3]="RegisterPanel",a[a.UnregisterPanel=4]="UnregisterPanel",a))(ra||{});let na={0(a,s){var t;let r=V(a.tabs,m=>m.current),n=V(a.panels,m=>m.current),l=r.filter(m=>{var y;return!((y=m.current)!=null&&y.hasAttribute("disabled"))}),o={...a,tabs:r,panels:n};if(s.index<0||s.index>r.length-1){let m=O(Math.sign(s.index-a.selectedIndex),{[-1]:()=>1,0:()=>O(Math.sign(s.index),{[-1]:()=>0,0:()=>0,1:()=>1}),1:()=>0});if(l.length===0)return o;let y=O(m,{0:()=>r.indexOf(l[0]),1:()=>r.indexOf(l[l.length-1])});return{...o,selectedIndex:y===-1?a.selectedIndex:y}}let x=r.slice(0,s.index),j=[...r.slice(s.index),...x].find(m=>l.includes(m));if(!j)return o;let f=(t=r.indexOf(j))!=null?t:a.selectedIndex;return f===-1&&(f=a.selectedIndex),{...o,selectedIndex:f}},1(a,s){if(a.tabs.includes(s.tab))return a;let t=a.tabs[a.selectedIndex],r=V([...a.tabs,s.tab],l=>l.current),n=a.selectedIndex;return a.info.current.isControlled||(n=r.indexOf(t),n===-1&&(n=a.selectedIndex)),{...a,tabs:r,selectedIndex:n}},2(a,s){return{...a,tabs:a.tabs.filter(t=>t!==s.tab)}},3(a,s){return a.panels.includes(s.panel)?a:{...a,panels:V([...a.panels,s.panel],t=>t.current)}},4(a,s){return{...a,panels:a.panels.filter(t=>t!==s.panel)}}},ie=d.createContext(null);ie.displayName="TabsDataContext";function z(a){let s=d.useContext(ie);if(s===null){let t=new Error(`<${a} /> is missing a parent <Tab.Group /> component.`);throw Error.captureStackTrace&&Error.captureStackTrace(t,z),t}return s}let le=d.createContext(null);le.displayName="TabsActionsContext";function de(a){let s=d.useContext(le);if(s===null){let t=new Error(`<${a} /> is missing a parent <Tab.Group /> component.`);throw Error.captureStackTrace&&Error.captureStackTrace(t,de),t}return s}function ia(a,s){return O(s.type,na,a,s)}let la=d.Fragment;function da(a,s){let{defaultIndex:t=0,vertical:r=!1,manual:n=!1,onChange:l,selectedIndex:o=null,...x}=a;const j=r?"vertical":"horizontal",f=n?"manual":"auto";let m=o!==null,y=F({isControlled:m}),T=U(s),[g,c]=d.useReducer(ia,{info:y,selectedIndex:o??t,tabs:[],panels:[]}),P=d.useMemo(()=>({selectedIndex:g.selectedIndex}),[g.selectedIndex]),b=F(l||(()=>{})),u=F(g.tabs),h=d.useMemo(()=>({orientation:j,activation:f,...g}),[j,f,g]),I=C(v=>(c({type:1,tab:v}),()=>c({type:2,tab:v}))),D=C(v=>(c({type:3,panel:v}),()=>c({type:4,panel:v}))),R=C(v=>{L.current!==v&&b.current(v),m||c({type:0,index:v})}),L=F(m?a.selectedIndex:g.selectedIndex),ae=d.useMemo(()=>({registerTab:I,registerPanel:D,change:R}),[]);ee(()=>{c({type:0,index:o??t})},[o]),ee(()=>{if(L.current===void 0||g.tabs.length<=0)return;let v=V(g.tabs,i=>i.current);v.some((i,N)=>g.tabs[N]!==i)&&R(v.indexOf(g.tabs[L.current]))});let se={ref:T};return A.createElement(ea,null,A.createElement(le.Provider,{value:ae},A.createElement(ie.Provider,{value:h},h.tabs.length<=0&&A.createElement(Qe,{onFocus:()=>{var v,i;for(let N of u.current)if(((v=N.current)==null?void 0:v.tabIndex)===0)return(i=N.current)==null||i.focus(),!0;return!1}}),G({ourProps:se,theirProps:x,slot:P,defaultTag:la,name:"Tabs"}))))}let oa="div";function ca(a,s){let{orientation:t,selectedIndex:r}=z("Tab.List"),n=U(s);return G({ourProps:{ref:n,role:"tablist","aria-orientation":t},theirProps:a,slot:{selectedIndex:r},defaultTag:oa,name:"Tabs.List"})}let ma="button";function ba(a,s){var t,r;let n=Oe(),{id:l=`headlessui-tabs-tab-${n}`,...o}=a,{orientation:x,activation:j,selectedIndex:f,tabs:m,panels:y}=z("Tab"),T=de("Tab"),g=z("Tab"),c=d.useRef(null),P=U(c,s);ee(()=>T.registerTab(c),[T,c]);let b=Ge("tabs"),u=m.indexOf(c);u===-1&&(u=b);let h=u===f,I=C(i=>{var N;let W=i();if(W===H.Success&&j==="auto"){let _e=(N=Ye(c))==null?void 0:N.activeElement,oe=g.tabs.findIndex(qe=>qe.current===_e);oe!==-1&&T.change(oe)}return W}),D=C(i=>{let N=m.map(W=>W.current).filter(Boolean);if(i.key===w.Space||i.key===w.Enter){i.preventDefault(),i.stopPropagation(),T.change(u);return}switch(i.key){case w.Home:case w.PageUp:return i.preventDefault(),i.stopPropagation(),I(()=>E(N,k.First));case w.End:case w.PageDown:return i.preventDefault(),i.stopPropagation(),I(()=>E(N,k.Last))}if(I(()=>O(x,{vertical(){return i.key===w.ArrowUp?E(N,k.Previous|k.WrapAround):i.key===w.ArrowDown?E(N,k.Next|k.WrapAround):H.Error},horizontal(){return i.key===w.ArrowLeft?E(N,k.Previous|k.WrapAround):i.key===w.ArrowRight?E(N,k.Next|k.WrapAround):H.Error}}))===H.Success)return i.preventDefault()}),R=d.useRef(!1),L=C(()=>{var i;R.current||(R.current=!0,(i=c.current)==null||i.focus({preventScroll:!0}),T.change(u),Be(()=>{R.current=!1}))}),ae=C(i=>{i.preventDefault()}),se=d.useMemo(()=>{var i;return{selected:h,disabled:(i=a.disabled)!=null?i:!1}},[h,a.disabled]),v={ref:P,onKeyDown:D,onMouseDown:ae,onClick:L,id:l,role:"tab",type:Je(a,c),"aria-controls":(r=(t=y[u])==null?void 0:t.current)==null?void 0:r.id,"aria-selected":h,tabIndex:h?0:-1};return G({ourProps:v,theirProps:o,slot:se,defaultTag:ma,name:"Tabs.Tab"})}let ua="div";function pa(a,s){let{selectedIndex:t}=z("Tab.Panels"),r=U(s),n=d.useMemo(()=>({selectedIndex:t}),[t]);return G({ourProps:{ref:r},theirProps:a,slot:n,defaultTag:ua,name:"Tabs.Panels"})}let xa="div",ga=ce.RenderStrategy|ce.Static;function va(a,s){var t,r,n,l;let o=Oe(),{id:x=`headlessui-tabs-panel-${o}`,tabIndex:j=0,...f}=a,{selectedIndex:m,tabs:y,panels:T}=z("Tab.Panel"),g=de("Tab.Panel"),c=d.useRef(null),P=U(c,s);ee(()=>g.registerPanel(c),[g,c,x]);let b=Ge("panels"),u=T.indexOf(c);u===-1&&(u=b);let h=u===m,I=d.useMemo(()=>({selected:h}),[h]),D={ref:P,id:x,role:"tabpanel","aria-labelledby":(r=(t=y[u])==null?void 0:t.current)==null?void 0:r.id,tabIndex:h?j:-1};return!h&&((n=f.unmount)==null||n)&&!((l=f.static)!=null&&l)?A.createElement($e,{as:"span","aria-hidden":"true",...D}):G({ourProps:D,theirProps:f,slot:I,defaultTag:xa,features:ga,visible:h,name:"Tabs.Panel"})}let ha=$(ba),ya=$(da),fa=$(ca),Na=$(pa),ja=$(va),M=Object.assign(ha,{Group:ya,List:fa,Panels:Na,Panel:ja});const Ta=ne("flex gap-1",{variants:{variant:{underline:"border-b border-gray-200",pill:"bg-gray-100 rounded-lg p-1",button:"gap-2"},orientation:{horizontal:"flex-row",vertical:"flex-col"}},compoundVariants:[{variant:"underline",orientation:"vertical",className:"border-b-0 border-r border-gray-200"}],defaultVariants:{variant:"underline",orientation:"horizontal"}}),wa=ne("relative inline-flex items-center justify-center gap-2 font-medium transition-all duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-navy-600 disabled:opacity-50 disabled:cursor-not-allowed",{variants:{variant:{underline:"border-b-2 border-transparent hover:text-navy-700 hover:border-gray-300",pill:"rounded-md hover:bg-white/60",button:"border border-gray-300 rounded-md hover:border-gray-400 hover:bg-gray-50"},size:{sm:"px-3 py-1.5 text-sm",md:"px-4 py-2 text-base",lg:"px-5 py-3 text-lg"},orientation:{horizontal:"",vertical:"w-full"},selected:{true:"",false:""}},compoundVariants:[{variant:"underline",selected:!0,className:"text-navy-800 border-navy-600 font-semibold"},{variant:"underline",selected:!1,className:"text-gray-600"},{variant:"pill",selected:!0,className:"bg-white text-navy-800 shadow-sm font-semibold"},{variant:"pill",selected:!1,className:"text-gray-700"},{variant:"button",selected:!0,className:"bg-navy-600 text-white border-navy-600 shadow-sm font-semibold hover:bg-navy-700"},{variant:"button",selected:!1,className:"bg-white text-gray-700"},{variant:"underline",orientation:"vertical",className:"border-b-0 border-r-2"}],defaultVariants:{variant:"underline",size:"md",orientation:"horizontal",selected:!1}}),ka=ne("focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-navy-600 rounded-md",{variants:{size:{sm:"mt-3",md:"mt-4",lg:"mt-6"},orientation:{horizontal:"",vertical:"ml-6"}},compoundVariants:[{orientation:"vertical",size:"sm",className:"ml-4 mt-0"},{orientation:"vertical",size:"md",className:"ml-6 mt-0"},{orientation:"vertical",size:"lg",className:"ml-8 mt-0"}],defaultVariants:{size:"md",orientation:"horizontal"}}),p=A.forwardRef(({tabs:a,children:s,variant:t="underline",size:r="md",orientation:n="horizontal",defaultTab:l,selectedTab:o,onTabChange:x,className:j,...f},m)=>{const y=A.Children.toArray(s);y.length!==a.length&&console.warn(`Tabs: Number of children (${y.length}) does not match number of tabs (${a.length})`);const T=l?a.findIndex(b=>b.id===l):0,g=o?a.findIndex(b=>b.id===o):void 0,c=b=>{x&&a[b]&&x(a[b].id)},P=o!==void 0;return e.jsx("div",{ref:m,className:te("w-full",n==="vertical"&&"flex",j),...f,children:e.jsxs(M.Group,{selectedIndex:P?g:void 0,defaultIndex:P?void 0:T,onChange:c,vertical:n==="vertical",children:[e.jsx(M.List,{className:Ta({variant:t,orientation:n}),children:a.map(b=>e.jsx(M,{disabled:b.disabled,className:({selected:u})=>wa({variant:t,size:r,orientation:n,selected:u}),children:({selected:u})=>e.jsxs(e.Fragment,{children:[b.icon&&e.jsx("span",{className:te("flex-shrink-0",r==="sm"&&"w-4 h-4",r==="md"&&"w-5 h-5",r==="lg"&&"w-6 h-6"),"aria-hidden":"true",children:b.icon}),e.jsx("span",{children:b.label})]})},b.id))}),e.jsx(M.Panels,{className:te(n==="vertical"&&"flex-1"),children:y.map((b,u)=>{var h;return e.jsx(M.Panel,{className:ka({size:r,orientation:n}),children:b},((h=a[u])==null?void 0:h.id)||u)})})]})})});p.displayName="Tabs";p.__docgenInfo={description:`Tabs component for organizing content into tabbed panels.
Supports both controlled and uncontrolled modes.

@example
\`\`\`tsx
// Uncontrolled
<Tabs
  tabs={[
    { id: 'tab1', label: 'Tab 1' },
    { id: 'tab2', label: 'Tab 2' },
  ]}
  defaultTab="tab1"
>
  <div>Panel 1</div>
  <div>Panel 2</div>
</Tabs>

// Controlled
<Tabs
  tabs={tabs}
  selectedTab={activeTab}
  onTabChange={setActiveTab}
>
  {panels}
</Tabs>
\`\`\``,methods:[],displayName:"Tabs",props:{tabs:{required:!0,tsType:{name:"Array",elements:[{name:"TabItem"}],raw:"TabItem[]"},description:"Array of tab items"},children:{required:!0,tsType:{name:"union",raw:"React.ReactNode[] | React.ReactNode",elements:[{name:"Array",elements:[{name:"ReactReactNode",raw:"React.ReactNode"}],raw:"React.ReactNode[]"},{name:"ReactReactNode",raw:"React.ReactNode"}]},description:"Tab panel content - must match tabs array length"},variant:{required:!1,tsType:{name:"union",raw:"'underline' | 'pill' | 'button'",elements:[{name:"literal",value:"'underline'"},{name:"literal",value:"'pill'"},{name:"literal",value:"'button'"}]},description:"Visual style variant",defaultValue:{value:"'underline'",computed:!1}},size:{required:!1,tsType:{name:"union",raw:"'sm' | 'md' | 'lg'",elements:[{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"},{name:"literal",value:"'lg'"}]},description:"Tab size",defaultValue:{value:"'md'",computed:!1}},orientation:{required:!1,tsType:{name:"union",raw:"'horizontal' | 'vertical'",elements:[{name:"literal",value:"'horizontal'"},{name:"literal",value:"'vertical'"}]},description:"Tab orientation",defaultValue:{value:"'horizontal'",computed:!1}},defaultTab:{required:!1,tsType:{name:"string"},description:"Default selected tab ID (uncontrolled mode)"},selectedTab:{required:!1,tsType:{name:"string"},description:"Selected tab ID (controlled mode)"},onTabChange:{required:!1,tsType:{name:"signature",type:"function",raw:"(tabId: string) => void",signature:{arguments:[{type:{name:"string"},name:"tabId"}],return:{name:"void"}}},description:"Callback when tab changes"}},composes:["Omit"]};const Ma={title:"Design System/Navigation/Tabs",component:p,parameters:{layout:"padded"},tags:["autodocs"]},We=()=>e.jsx("svg",{fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"})}),Sa=()=>e.jsx("svg",{fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"})}),Fe=()=>e.jsxs("svg",{fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:[e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"}),e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M15 12a3 3 0 11-6 0 3 3 0 016 0z"})]}),He=()=>e.jsx("svg",{fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"})}),S=[{id:"overview",label:"Overview"},{id:"details",label:"Details"},{id:"settings",label:"Settings"}],re=[{id:"home",label:"Home",icon:e.jsx(We,{})},{id:"profile",label:"Profile",icon:e.jsx(Sa,{})},{id:"analytics",label:"Analytics",icon:e.jsx(He,{})},{id:"settings",label:"Settings",icon:e.jsx(Fe,{})}],Pa=[{id:"granted",label:"Granted"},{id:"pending",label:"Pending"},{id:"revoked",label:"Revoked"}],_={render:()=>e.jsxs(p,{tabs:S,defaultTab:"overview",className:"w-full max-w-2xl",children:[e.jsxs("div",{className:"p-4 bg-gray-50 rounded border border-gray-200",children:[e.jsx("h3",{className:"font-semibold text-gray-900 mb-2",children:"Overview"}),e.jsx("p",{className:"text-gray-700",children:"This is the overview panel with general information and summary data."})]}),e.jsxs("div",{className:"p-4 bg-gray-50 rounded border border-gray-200",children:[e.jsx("h3",{className:"font-semibold text-gray-900 mb-2",children:"Details"}),e.jsx("p",{className:"text-gray-700",children:"Detailed information and specific data points are displayed here."})]}),e.jsxs("div",{className:"p-4 bg-gray-50 rounded border border-gray-200",children:[e.jsx("h3",{className:"font-semibold text-gray-900 mb-2",children:"Settings"}),e.jsx("p",{className:"text-gray-700",children:"Configuration options and preferences can be adjusted here."})]})]})},q={render:()=>e.jsxs("div",{className:"space-y-12 w-full max-w-2xl",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4",children:"Underline (Default)"}),e.jsxs(p,{tabs:S,defaultTab:"overview",variant:"underline",children:[e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Overview content"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Details content"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Settings content"})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4",children:"Pill"}),e.jsxs(p,{tabs:S,defaultTab:"overview",variant:"pill",children:[e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Overview content"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Details content"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Settings content"})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4",children:"Button"}),e.jsxs(p,{tabs:S,defaultTab:"overview",variant:"button",children:[e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Overview content"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Details content"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Settings content"})]})]})]})},B={render:()=>{const a=[{id:"active",label:"Active"},{id:"disabled",label:"Disabled",disabled:!0},{id:"another",label:"Another Active"},{id:"locked",label:"Locked",disabled:!0}];return e.jsxs(p,{tabs:a,defaultTab:"active",className:"w-full max-w-2xl",children:[e.jsx("div",{className:"p-4 bg-emerald-50 border border-emerald-200 rounded",children:e.jsx("p",{className:"text-emerald-800",children:"Active tab content is accessible"})}),e.jsx("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded",children:e.jsx("p",{className:"text-gray-500",children:"This tab is disabled"})}),e.jsx("div",{className:"p-4 bg-emerald-50 border border-emerald-200 rounded",children:e.jsx("p",{className:"text-emerald-800",children:"Another active tab"})}),e.jsx("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded",children:e.jsx("p",{className:"text-gray-500",children:"This tab is locked"})})]})}},J={render:()=>e.jsxs("div",{className:"space-y-12 w-full max-w-3xl",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4",children:"Vertical Underline"}),e.jsxs(p,{tabs:S,defaultTab:"overview",variant:"underline",orientation:"vertical",children:[e.jsxs("div",{className:"p-4 bg-gray-50 rounded border border-gray-200",children:[e.jsx("h4",{className:"font-semibold text-gray-900 mb-2",children:"Overview"}),e.jsx("p",{className:"text-gray-700",children:"Vertical layout with underline variant."})]}),e.jsxs("div",{className:"p-4 bg-gray-50 rounded border border-gray-200",children:[e.jsx("h4",{className:"font-semibold text-gray-900 mb-2",children:"Details"}),e.jsx("p",{className:"text-gray-700",children:"Detailed view in vertical orientation."})]}),e.jsxs("div",{className:"p-4 bg-gray-50 rounded border border-gray-200",children:[e.jsx("h4",{className:"font-semibold text-gray-900 mb-2",children:"Settings"}),e.jsx("p",{className:"text-gray-700",children:"Settings panel in vertical layout."})]})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4",children:"Vertical Button"}),e.jsxs(p,{tabs:re,defaultTab:"home",variant:"button",orientation:"vertical",children:[e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Home dashboard content"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Profile information"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Analytics and reports"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Settings and preferences"})]})]})]})},K={render:()=>e.jsxs("div",{className:"space-y-12 w-full max-w-2xl",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4",children:"Small"}),e.jsxs(p,{tabs:S,defaultTab:"overview",size:"sm",children:[e.jsx("div",{className:"p-3 bg-gray-50 rounded text-sm",children:"Small tab content"}),e.jsx("div",{className:"p-3 bg-gray-50 rounded text-sm",children:"Small tab content"}),e.jsx("div",{className:"p-3 bg-gray-50 rounded text-sm",children:"Small tab content"})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4",children:"Medium (Default)"}),e.jsxs(p,{tabs:S,defaultTab:"overview",size:"md",children:[e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Medium tab content"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Medium tab content"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Medium tab content"})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4",children:"Large"}),e.jsxs(p,{tabs:S,defaultTab:"overview",size:"lg",children:[e.jsx("div",{className:"p-6 bg-gray-50 rounded text-lg",children:"Large tab content"}),e.jsx("div",{className:"p-6 bg-gray-50 rounded text-lg",children:"Large tab content"}),e.jsx("div",{className:"p-6 bg-gray-50 rounded text-lg",children:"Large tab content"})]})]})]})},Y={render:()=>e.jsxs("div",{className:"space-y-12 w-full max-w-2xl",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4",children:"Underline with Icons"}),e.jsxs(p,{tabs:re,defaultTab:"home",variant:"underline",children:[e.jsxs("div",{className:"p-4 bg-gray-50 rounded",children:[e.jsx("h4",{className:"font-semibold text-gray-900 mb-2",children:"Home Dashboard"}),e.jsx("p",{className:"text-gray-700",children:"Welcome to your dashboard."})]}),e.jsxs("div",{className:"p-4 bg-gray-50 rounded",children:[e.jsx("h4",{className:"font-semibold text-gray-900 mb-2",children:"User Profile"}),e.jsx("p",{className:"text-gray-700",children:"Manage your profile information."})]}),e.jsxs("div",{className:"p-4 bg-gray-50 rounded",children:[e.jsx("h4",{className:"font-semibold text-gray-900 mb-2",children:"Analytics"}),e.jsx("p",{className:"text-gray-700",children:"View your analytics and insights."})]}),e.jsxs("div",{className:"p-4 bg-gray-50 rounded",children:[e.jsx("h4",{className:"font-semibold text-gray-900 mb-2",children:"Settings"}),e.jsx("p",{className:"text-gray-700",children:"Configure your preferences."})]})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4",children:"Pill with Icons"}),e.jsxs(p,{tabs:re,defaultTab:"home",variant:"pill",children:[e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Home content with icons"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Profile content with icons"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Analytics content with icons"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded",children:"Settings content with icons"})]})]})]})},X={render:()=>{const[a,s]=d.useState("granted");return e.jsxs("div",{className:"w-full max-w-2xl space-y-4",children:[e.jsxs("div",{className:"p-4 bg-navy-50 border border-navy-200 rounded-lg",children:[e.jsxs("p",{className:"text-sm text-navy-800 mb-2",children:[e.jsx("strong",{children:"Controlled mode:"})," The parent component manages the active tab state."]}),e.jsxs("p",{className:"text-sm text-navy-700",children:["Current active tab: ",e.jsx("code",{className:"px-2 py-0.5 bg-navy-100 rounded",children:a})]})]}),e.jsxs(p,{tabs:Pa,selectedTab:a,onTabChange:s,variant:"pill",children:[e.jsxs("div",{className:"p-6 bg-emerald-50 border border-emerald-200 rounded",children:[e.jsx("h4",{className:"font-semibold text-emerald-900 mb-3",children:"Granted Permissions"}),e.jsxs("ul",{className:"space-y-2 text-emerald-800",children:[e.jsx("li",{children:"• Read access to documents"}),e.jsx("li",{children:"• Edit personal profile"}),e.jsx("li",{children:"• View analytics dashboard"})]})]}),e.jsxs("div",{className:"p-6 bg-amber-50 border border-amber-200 rounded",children:[e.jsx("h4",{className:"font-semibold text-amber-900 mb-3",children:"Pending Approvals"}),e.jsxs("ul",{className:"space-y-2 text-amber-800",children:[e.jsx("li",{children:"• Admin access - awaiting approval"}),e.jsx("li",{children:"• Delete permissions - under review"})]})]}),e.jsxs("div",{className:"p-6 bg-red-50 border border-red-200 rounded",children:[e.jsx("h4",{className:"font-semibold text-red-900 mb-3",children:"Revoked Access"}),e.jsxs("ul",{className:"space-y-2 text-red-800",children:[e.jsx("li",{children:"• Export data - revoked on 2024-01-15"}),e.jsx("li",{children:"• Share externally - revoked on 2024-01-10"})]})]})]}),e.jsxs("div",{className:"flex gap-2 pt-4",children:[e.jsx("button",{onClick:()=>s("granted"),className:"px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors",children:"Go to Granted"}),e.jsx("button",{onClick:()=>s("pending"),className:"px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors",children:"Go to Pending"}),e.jsx("button",{onClick:()=>s("revoked"),className:"px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors",children:"Go to Revoked"})]})]})}},Q={render:()=>{const[a,s]=d.useState("current"),t=[{id:"current",label:"Current Consents",icon:e.jsx(We,{})},{id:"history",label:"History",icon:e.jsx(He,{})},{id:"preferences",label:"Preferences",icon:e.jsx(Fe,{})}];return e.jsxs("div",{className:"w-full max-w-3xl p-6 bg-white border border-gray-200 rounded-lg shadow-md-premium",children:[e.jsxs("div",{className:"mb-6",children:[e.jsx("h2",{className:"text-2xl font-display font-semibold text-navy-900 mb-2",children:"Consent Management"}),e.jsx("p",{className:"text-gray-600",children:"Manage your data sharing permissions and consent preferences"})]}),e.jsxs(p,{tabs:t,selectedTab:a,onTabChange:s,variant:"underline",size:"md",children:[e.jsxs("div",{className:"space-y-4",children:[e.jsxs("div",{className:"p-4 border border-gray-200 rounded-lg",children:[e.jsxs("div",{className:"flex items-start justify-between mb-2",children:[e.jsx("h4",{className:"font-semibold text-gray-900",children:"Healthcare Provider Access"}),e.jsx("span",{className:"px-2 py-1 text-xs font-medium bg-emerald-100 text-emerald-800 rounded",children:"Active"})]}),e.jsx("p",{className:"text-sm text-gray-600 mb-3",children:"Allows Dr. Smith to access your medical records"}),e.jsxs("div",{className:"flex items-center justify-between text-xs text-gray-500",children:[e.jsx("span",{children:"Granted: Jan 15, 2024"}),e.jsx("span",{children:"Expires: Jan 15, 2025"})]})]}),e.jsxs("div",{className:"p-4 border border-gray-200 rounded-lg",children:[e.jsxs("div",{className:"flex items-start justify-between mb-2",children:[e.jsx("h4",{className:"font-semibold text-gray-900",children:"Research Study Participation"}),e.jsx("span",{className:"px-2 py-1 text-xs font-medium bg-emerald-100 text-emerald-800 rounded",children:"Active"})]}),e.jsx("p",{className:"text-sm text-gray-600 mb-3",children:"Anonymous data sharing for clinical research"}),e.jsxs("div",{className:"flex items-center justify-between text-xs text-gray-500",children:[e.jsx("span",{children:"Granted: Dec 1, 2023"}),e.jsx("span",{children:"Expires: Dec 1, 2024"})]})]})]}),e.jsxs("div",{className:"space-y-3",children:[e.jsxs("div",{className:"p-3 bg-gray-50 border border-gray-200 rounded",children:[e.jsxs("div",{className:"flex items-center justify-between mb-1",children:[e.jsx("span",{className:"text-sm font-medium text-gray-900",children:"Lab Results Access"}),e.jsx("span",{className:"text-xs text-gray-500",children:"Revoked"})]}),e.jsx("p",{className:"text-xs text-gray-600",children:"Revoked on Jan 10, 2024"})]}),e.jsxs("div",{className:"p-3 bg-gray-50 border border-gray-200 rounded",children:[e.jsxs("div",{className:"flex items-center justify-between mb-1",children:[e.jsx("span",{className:"text-sm font-medium text-gray-900",children:"Pharmacy Access"}),e.jsx("span",{className:"text-xs text-gray-500",children:"Expired"})]}),e.jsx("p",{className:"text-xs text-gray-600",children:"Expired on Dec 31, 2023"})]}),e.jsxs("div",{className:"p-3 bg-gray-50 border border-gray-200 rounded",children:[e.jsxs("div",{className:"flex items-center justify-between mb-1",children:[e.jsx("span",{className:"text-sm font-medium text-gray-900",children:"Insurance Verification"}),e.jsx("span",{className:"text-xs text-gray-500",children:"Completed"})]}),e.jsx("p",{className:"text-xs text-gray-600",children:"Completed on Nov 15, 2023"})]})]}),e.jsxs("div",{className:"space-y-4",children:[e.jsxs("div",{className:"p-4 border border-gray-200 rounded-lg",children:[e.jsx("h4",{className:"font-semibold text-gray-900 mb-3",children:"Default Consent Duration"}),e.jsxs("select",{className:"w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-navy-600",children:[e.jsx("option",{children:"30 days"}),e.jsx("option",{children:"90 days"}),e.jsx("option",{selected:!0,children:"1 year"}),e.jsx("option",{children:"Until revoked"})]})]}),e.jsxs("div",{className:"p-4 border border-gray-200 rounded-lg",children:[e.jsx("h4",{className:"font-semibold text-gray-900 mb-3",children:"Notification Preferences"}),e.jsxs("div",{className:"space-y-2",children:[e.jsxs("label",{className:"flex items-center gap-2",children:[e.jsx("input",{type:"checkbox",checked:!0,className:"rounded"}),e.jsx("span",{className:"text-sm text-gray-700",children:"Email me when consent is requested"})]}),e.jsxs("label",{className:"flex items-center gap-2",children:[e.jsx("input",{type:"checkbox",checked:!0,className:"rounded"}),e.jsx("span",{className:"text-sm text-gray-700",children:"Notify before consent expires"})]}),e.jsxs("label",{className:"flex items-center gap-2",children:[e.jsx("input",{type:"checkbox",className:"rounded"}),e.jsx("span",{className:"text-sm text-gray-700",children:"Weekly consent summary"})]})]})]})]})]})]})}},Z={render:a=>{const[s,t]=d.useState(a.defaultTab||"overview");return e.jsxs(p,{...a,selectedTab:s,onTabChange:t,className:"w-full max-w-2xl",children:[e.jsxs("div",{className:"p-4 bg-gray-50 rounded border border-gray-200",children:[e.jsx("h3",{className:"font-semibold text-gray-900 mb-2",children:"Overview Panel"}),e.jsx("p",{className:"text-gray-700",children:"This is the overview content panel."})]}),e.jsxs("div",{className:"p-4 bg-gray-50 rounded border border-gray-200",children:[e.jsx("h3",{className:"font-semibold text-gray-900 mb-2",children:"Details Panel"}),e.jsx("p",{className:"text-gray-700",children:"This is the details content panel."})]}),e.jsxs("div",{className:"p-4 bg-gray-50 rounded border border-gray-200",children:[e.jsx("h3",{className:"font-semibold text-gray-900 mb-2",children:"Settings Panel"}),e.jsx("p",{className:"text-gray-700",children:"This is the settings content panel."})]})]})},args:{tabs:S,variant:"underline",size:"md",orientation:"horizontal",defaultTab:"overview"},argTypes:{variant:{control:"select",options:["underline","pill","button"],description:"Visual style of the tabs"},size:{control:"select",options:["sm","md","lg"],description:"Size of the tabs"},orientation:{control:"select",options:["horizontal","vertical"],description:"Tab orientation"},defaultTab:{control:"text",description:"ID of the initially selected tab"}}};var me,be,ue;_.parameters={..._.parameters,docs:{...(me=_.parameters)==null?void 0:me.docs,source:{originalSource:`{
  render: () => <Tabs tabs={basicTabs} defaultTab="overview" className="w-full max-w-2xl">
      <div className="p-4 bg-gray-50 rounded border border-gray-200">
        <h3 className="font-semibold text-gray-900 mb-2">Overview</h3>
        <p className="text-gray-700">
          This is the overview panel with general information and summary data.
        </p>
      </div>
      <div className="p-4 bg-gray-50 rounded border border-gray-200">
        <h3 className="font-semibold text-gray-900 mb-2">Details</h3>
        <p className="text-gray-700">
          Detailed information and specific data points are displayed here.
        </p>
      </div>
      <div className="p-4 bg-gray-50 rounded border border-gray-200">
        <h3 className="font-semibold text-gray-900 mb-2">Settings</h3>
        <p className="text-gray-700">
          Configuration options and preferences can be adjusted here.
        </p>
      </div>
    </Tabs>
}`,...(ue=(be=_.parameters)==null?void 0:be.docs)==null?void 0:ue.source}}};var pe,xe,ge;q.parameters={...q.parameters,docs:{...(pe=q.parameters)==null?void 0:pe.docs,source:{originalSource:`{
  render: () => <div className="space-y-12 w-full max-w-2xl">
      {/* Underline variant */}
      <div>
        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4">
          Underline (Default)
        </h3>
        <Tabs tabs={basicTabs} defaultTab="overview" variant="underline">
          <div className="p-4 bg-gray-50 rounded">Overview content</div>
          <div className="p-4 bg-gray-50 rounded">Details content</div>
          <div className="p-4 bg-gray-50 rounded">Settings content</div>
        </Tabs>
      </div>

      {/* Pill variant */}
      <div>
        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4">
          Pill
        </h3>
        <Tabs tabs={basicTabs} defaultTab="overview" variant="pill">
          <div className="p-4 bg-gray-50 rounded">Overview content</div>
          <div className="p-4 bg-gray-50 rounded">Details content</div>
          <div className="p-4 bg-gray-50 rounded">Settings content</div>
        </Tabs>
      </div>

      {/* Button variant */}
      <div>
        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4">
          Button
        </h3>
        <Tabs tabs={basicTabs} defaultTab="overview" variant="button">
          <div className="p-4 bg-gray-50 rounded">Overview content</div>
          <div className="p-4 bg-gray-50 rounded">Details content</div>
          <div className="p-4 bg-gray-50 rounded">Settings content</div>
        </Tabs>
      </div>
    </div>
}`,...(ge=(xe=q.parameters)==null?void 0:xe.docs)==null?void 0:ge.source}}};var ve,he,ye;B.parameters={...B.parameters,docs:{...(ve=B.parameters)==null?void 0:ve.docs,source:{originalSource:`{
  render: () => {
    const tabsWithDisabled: TabItem[] = [{
      id: 'active',
      label: 'Active'
    }, {
      id: 'disabled',
      label: 'Disabled',
      disabled: true
    }, {
      id: 'another',
      label: 'Another Active'
    }, {
      id: 'locked',
      label: 'Locked',
      disabled: true
    }];
    return <Tabs tabs={tabsWithDisabled} defaultTab="active" className="w-full max-w-2xl">
        <div className="p-4 bg-emerald-50 border border-emerald-200 rounded">
          <p className="text-emerald-800">Active tab content is accessible</p>
        </div>
        <div className="p-4 bg-gray-50 border border-gray-200 rounded">
          <p className="text-gray-500">This tab is disabled</p>
        </div>
        <div className="p-4 bg-emerald-50 border border-emerald-200 rounded">
          <p className="text-emerald-800">Another active tab</p>
        </div>
        <div className="p-4 bg-gray-50 border border-gray-200 rounded">
          <p className="text-gray-500">This tab is locked</p>
        </div>
      </Tabs>;
  }
}`,...(ye=(he=B.parameters)==null?void 0:he.docs)==null?void 0:ye.source}}};var fe,Ne,je;J.parameters={...J.parameters,docs:{...(fe=J.parameters)==null?void 0:fe.docs,source:{originalSource:`{
  render: () => <div className="space-y-12 w-full max-w-3xl">
      {/* Vertical underline */}
      <div>
        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4">
          Vertical Underline
        </h3>
        <Tabs tabs={basicTabs} defaultTab="overview" variant="underline" orientation="vertical">
          <div className="p-4 bg-gray-50 rounded border border-gray-200">
            <h4 className="font-semibold text-gray-900 mb-2">Overview</h4>
            <p className="text-gray-700">Vertical layout with underline variant.</p>
          </div>
          <div className="p-4 bg-gray-50 rounded border border-gray-200">
            <h4 className="font-semibold text-gray-900 mb-2">Details</h4>
            <p className="text-gray-700">Detailed view in vertical orientation.</p>
          </div>
          <div className="p-4 bg-gray-50 rounded border border-gray-200">
            <h4 className="font-semibold text-gray-900 mb-2">Settings</h4>
            <p className="text-gray-700">Settings panel in vertical layout.</p>
          </div>
        </Tabs>
      </div>

      {/* Vertical button */}
      <div>
        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4">
          Vertical Button
        </h3>
        <Tabs tabs={iconTabs} defaultTab="home" variant="button" orientation="vertical">
          <div className="p-4 bg-gray-50 rounded">Home dashboard content</div>
          <div className="p-4 bg-gray-50 rounded">Profile information</div>
          <div className="p-4 bg-gray-50 rounded">Analytics and reports</div>
          <div className="p-4 bg-gray-50 rounded">Settings and preferences</div>
        </Tabs>
      </div>
    </div>
}`,...(je=(Ne=J.parameters)==null?void 0:Ne.docs)==null?void 0:je.source}}};var Te,we,ke;K.parameters={...K.parameters,docs:{...(Te=K.parameters)==null?void 0:Te.docs,source:{originalSource:`{
  render: () => <div className="space-y-12 w-full max-w-2xl">
      {/* Small */}
      <div>
        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4">
          Small
        </h3>
        <Tabs tabs={basicTabs} defaultTab="overview" size="sm">
          <div className="p-3 bg-gray-50 rounded text-sm">Small tab content</div>
          <div className="p-3 bg-gray-50 rounded text-sm">Small tab content</div>
          <div className="p-3 bg-gray-50 rounded text-sm">Small tab content</div>
        </Tabs>
      </div>

      {/* Medium (default) */}
      <div>
        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4">
          Medium (Default)
        </h3>
        <Tabs tabs={basicTabs} defaultTab="overview" size="md">
          <div className="p-4 bg-gray-50 rounded">Medium tab content</div>
          <div className="p-4 bg-gray-50 rounded">Medium tab content</div>
          <div className="p-4 bg-gray-50 rounded">Medium tab content</div>
        </Tabs>
      </div>

      {/* Large */}
      <div>
        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4">
          Large
        </h3>
        <Tabs tabs={basicTabs} defaultTab="overview" size="lg">
          <div className="p-6 bg-gray-50 rounded text-lg">Large tab content</div>
          <div className="p-6 bg-gray-50 rounded text-lg">Large tab content</div>
          <div className="p-6 bg-gray-50 rounded text-lg">Large tab content</div>
        </Tabs>
      </div>
    </div>
}`,...(ke=(we=K.parameters)==null?void 0:we.docs)==null?void 0:ke.source}}};var Se,Pe,Ae;Y.parameters={...Y.parameters,docs:{...(Se=Y.parameters)==null?void 0:Se.docs,source:{originalSource:`{
  render: () => <div className="space-y-12 w-full max-w-2xl">
      {/* Underline with icons */}
      <div>
        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4">
          Underline with Icons
        </h3>
        <Tabs tabs={iconTabs} defaultTab="home" variant="underline">
          <div className="p-4 bg-gray-50 rounded">
            <h4 className="font-semibold text-gray-900 mb-2">Home Dashboard</h4>
            <p className="text-gray-700">Welcome to your dashboard.</p>
          </div>
          <div className="p-4 bg-gray-50 rounded">
            <h4 className="font-semibold text-gray-900 mb-2">User Profile</h4>
            <p className="text-gray-700">Manage your profile information.</p>
          </div>
          <div className="p-4 bg-gray-50 rounded">
            <h4 className="font-semibold text-gray-900 mb-2">Analytics</h4>
            <p className="text-gray-700">View your analytics and insights.</p>
          </div>
          <div className="p-4 bg-gray-50 rounded">
            <h4 className="font-semibold text-gray-900 mb-2">Settings</h4>
            <p className="text-gray-700">Configure your preferences.</p>
          </div>
        </Tabs>
      </div>

      {/* Pill with icons */}
      <div>
        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4">
          Pill with Icons
        </h3>
        <Tabs tabs={iconTabs} defaultTab="home" variant="pill">
          <div className="p-4 bg-gray-50 rounded">Home content with icons</div>
          <div className="p-4 bg-gray-50 rounded">Profile content with icons</div>
          <div className="p-4 bg-gray-50 rounded">Analytics content with icons</div>
          <div className="p-4 bg-gray-50 rounded">Settings content with icons</div>
        </Tabs>
      </div>
    </div>
}`,...(Ae=(Pe=Y.parameters)==null?void 0:Pe.docs)==null?void 0:Ae.source}}};var Ie,Ce,De;X.parameters={...X.parameters,docs:{...(Ie=X.parameters)==null?void 0:Ie.docs,source:{originalSource:`{
  render: () => {
    const [activeTab, setActiveTab] = useState('granted');
    return <div className="w-full max-w-2xl space-y-4">
        <div className="p-4 bg-navy-50 border border-navy-200 rounded-lg">
          <p className="text-sm text-navy-800 mb-2">
            <strong>Controlled mode:</strong> The parent component manages the active tab state.
          </p>
          <p className="text-sm text-navy-700">
            Current active tab: <code className="px-2 py-0.5 bg-navy-100 rounded">{activeTab}</code>
          </p>
        </div>

        <Tabs tabs={permissionTabs} selectedTab={activeTab} onTabChange={setActiveTab} variant="pill">
          <div className="p-6 bg-emerald-50 border border-emerald-200 rounded">
            <h4 className="font-semibold text-emerald-900 mb-3">Granted Permissions</h4>
            <ul className="space-y-2 text-emerald-800">
              <li>• Read access to documents</li>
              <li>• Edit personal profile</li>
              <li>• View analytics dashboard</li>
            </ul>
          </div>
          <div className="p-6 bg-amber-50 border border-amber-200 rounded">
            <h4 className="font-semibold text-amber-900 mb-3">Pending Approvals</h4>
            <ul className="space-y-2 text-amber-800">
              <li>• Admin access - awaiting approval</li>
              <li>• Delete permissions - under review</li>
            </ul>
          </div>
          <div className="p-6 bg-red-50 border border-red-200 rounded">
            <h4 className="font-semibold text-red-900 mb-3">Revoked Access</h4>
            <ul className="space-y-2 text-red-800">
              <li>• Export data - revoked on 2024-01-15</li>
              <li>• Share externally - revoked on 2024-01-10</li>
            </ul>
          </div>
        </Tabs>

        <div className="flex gap-2 pt-4">
          <button onClick={() => setActiveTab('granted')} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors">
            Go to Granted
          </button>
          <button onClick={() => setActiveTab('pending')} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors">
            Go to Pending
          </button>
          <button onClick={() => setActiveTab('revoked')} className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors">
            Go to Revoked
          </button>
        </div>
      </div>;
  }
}`,...(De=(Ce=X.parameters)==null?void 0:Ce.docs)==null?void 0:De.source}}};var Re,Ee,ze;Q.parameters={...Q.parameters,docs:{...(Re=Q.parameters)==null?void 0:Re.docs,source:{originalSource:`{
  render: () => {
    const [activeTab, setActiveTab] = useState('current');
    const consentTabs: TabItem[] = [{
      id: 'current',
      label: 'Current Consents',
      icon: <HomeIcon />
    }, {
      id: 'history',
      label: 'History',
      icon: <ChartIcon />
    }, {
      id: 'preferences',
      label: 'Preferences',
      icon: <SettingsIcon />
    }];
    return <div className="w-full max-w-3xl p-6 bg-white border border-gray-200 rounded-lg shadow-md-premium">
        <div className="mb-6">
          <h2 className="text-2xl font-display font-semibold text-navy-900 mb-2">
            Consent Management
          </h2>
          <p className="text-gray-600">
            Manage your data sharing permissions and consent preferences
          </p>
        </div>

        <Tabs tabs={consentTabs} selectedTab={activeTab} onTabChange={setActiveTab} variant="underline" size="md">
          {/* Current Consents */}
          <div className="space-y-4">
            <div className="p-4 border border-gray-200 rounded-lg">
              <div className="flex items-start justify-between mb-2">
                <h4 className="font-semibold text-gray-900">Healthcare Provider Access</h4>
                <span className="px-2 py-1 text-xs font-medium bg-emerald-100 text-emerald-800 rounded">
                  Active
                </span>
              </div>
              <p className="text-sm text-gray-600 mb-3">
                Allows Dr. Smith to access your medical records
              </p>
              <div className="flex items-center justify-between text-xs text-gray-500">
                <span>Granted: Jan 15, 2024</span>
                <span>Expires: Jan 15, 2025</span>
              </div>
            </div>
            <div className="p-4 border border-gray-200 rounded-lg">
              <div className="flex items-start justify-between mb-2">
                <h4 className="font-semibold text-gray-900">Research Study Participation</h4>
                <span className="px-2 py-1 text-xs font-medium bg-emerald-100 text-emerald-800 rounded">
                  Active
                </span>
              </div>
              <p className="text-sm text-gray-600 mb-3">
                Anonymous data sharing for clinical research
              </p>
              <div className="flex items-center justify-between text-xs text-gray-500">
                <span>Granted: Dec 1, 2023</span>
                <span>Expires: Dec 1, 2024</span>
              </div>
            </div>
          </div>

          {/* History */}
          <div className="space-y-3">
            <div className="p-3 bg-gray-50 border border-gray-200 rounded">
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm font-medium text-gray-900">Lab Results Access</span>
                <span className="text-xs text-gray-500">Revoked</span>
              </div>
              <p className="text-xs text-gray-600">Revoked on Jan 10, 2024</p>
            </div>
            <div className="p-3 bg-gray-50 border border-gray-200 rounded">
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm font-medium text-gray-900">Pharmacy Access</span>
                <span className="text-xs text-gray-500">Expired</span>
              </div>
              <p className="text-xs text-gray-600">Expired on Dec 31, 2023</p>
            </div>
            <div className="p-3 bg-gray-50 border border-gray-200 rounded">
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm font-medium text-gray-900">Insurance Verification</span>
                <span className="text-xs text-gray-500">Completed</span>
              </div>
              <p className="text-xs text-gray-600">Completed on Nov 15, 2023</p>
            </div>
          </div>

          {/* Preferences */}
          <div className="space-y-4">
            <div className="p-4 border border-gray-200 rounded-lg">
              <h4 className="font-semibold text-gray-900 mb-3">Default Consent Duration</h4>
              <select className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-navy-600">
                <option>30 days</option>
                <option>90 days</option>
                <option selected>1 year</option>
                <option>Until revoked</option>
              </select>
            </div>
            <div className="p-4 border border-gray-200 rounded-lg">
              <h4 className="font-semibold text-gray-900 mb-3">Notification Preferences</h4>
              <div className="space-y-2">
                <label className="flex items-center gap-2">
                  <input type="checkbox" checked className="rounded" />
                  <span className="text-sm text-gray-700">Email me when consent is requested</span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="checkbox" checked className="rounded" />
                  <span className="text-sm text-gray-700">Notify before consent expires</span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="checkbox" className="rounded" />
                  <span className="text-sm text-gray-700">Weekly consent summary</span>
                </label>
              </div>
            </div>
          </div>
        </Tabs>
      </div>;
  }
}`,...(ze=(Ee=Q.parameters)==null?void 0:Ee.docs)==null?void 0:ze.source}}};var Le,Me,Ve;Z.parameters={...Z.parameters,docs:{...(Le=Z.parameters)==null?void 0:Le.docs,source:{originalSource:`{
  render: args => {
    const [activeTab, setActiveTab] = useState(args.defaultTab || 'overview');
    return <Tabs {...args} selectedTab={activeTab} onTabChange={setActiveTab} className="w-full max-w-2xl">
        <div className="p-4 bg-gray-50 rounded border border-gray-200">
          <h3 className="font-semibold text-gray-900 mb-2">Overview Panel</h3>
          <p className="text-gray-700">This is the overview content panel.</p>
        </div>
        <div className="p-4 bg-gray-50 rounded border border-gray-200">
          <h3 className="font-semibold text-gray-900 mb-2">Details Panel</h3>
          <p className="text-gray-700">This is the details content panel.</p>
        </div>
        <div className="p-4 bg-gray-50 rounded border border-gray-200">
          <h3 className="font-semibold text-gray-900 mb-2">Settings Panel</h3>
          <p className="text-gray-700">This is the settings content panel.</p>
        </div>
      </Tabs>;
  },
  args: {
    tabs: basicTabs,
    variant: 'underline',
    size: 'md',
    orientation: 'horizontal',
    defaultTab: 'overview'
  },
  argTypes: {
    variant: {
      control: 'select',
      options: ['underline', 'pill', 'button'],
      description: 'Visual style of the tabs'
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
      description: 'Size of the tabs'
    },
    orientation: {
      control: 'select',
      options: ['horizontal', 'vertical'],
      description: 'Tab orientation'
    },
    defaultTab: {
      control: 'text',
      description: 'ID of the initially selected tab'
    }
  }
}`,...(Ve=(Me=Z.parameters)==null?void 0:Me.docs)==null?void 0:Ve.source}}};const Va=["Default","Variants","DisabledTabs","Vertical","Sizes","WithIcons","Controlled","ConsentManagementExample","Playground"];export{Q as ConsentManagementExample,X as Controlled,_ as Default,B as DisabledTabs,Z as Playground,K as Sizes,q as Variants,J as Vertical,Y as WithIcons,Va as __namedExportsOrder,Ma as default};
