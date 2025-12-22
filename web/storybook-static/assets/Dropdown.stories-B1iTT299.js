import{j as e}from"./jsx-runtime-BYYWji4R.js";import{r as h,R as D}from"./index-ClcD9ViR.js";import{c as H,a as St}from"./cn-JCLedEej.js";import{p as kt,r as At}from"./bugs-Ot56VXf3.js";import{l as J,U as G,y as X,o as w,u as Dt,C as Y,I as oe,a as ne,c as b,O as de}from"./keyboard-DNQdDreo.js";import{y as Rt,s as Bt,u as Lt,d as K,q as Tt}from"./transition-CF-Dw_VV.js";import{n as Pt}from"./use-owner-CcFJWZQm.js";import{T as Et}from"./use-resolve-button-type-Bw633SU1.js";import{s as $t,u as Ft,f as Ot,c as I}from"./use-text-value-DhDQuHs4.js";import{o as Wt,I as Ut,h as Vt,T as _t,_ as qt,M as ue,D as Mt}from"./use-is-mounted-62Q7yJH9.js";import{B as y}from"./Button-DlbfybpJ.js";import"./_commonjsHelpers-Cpj98o6Y.js";function Ht({container:t,accept:a,walk:i,enabled:s=!0}){let n=h.useRef(a),c=h.useRef(i);h.useEffect(()=>{n.current=a,c.current=i},[a,i]),J(()=>{if(!t||!s)return;let l=Wt(t);if(!l)return;let r=n.current,d=c.current,x=Object.assign(p=>r(p),{acceptNode:r}),m=l.createTreeWalker(t,NodeFilter.SHOW_ELEMENT,x,!1);for(;m.nextNode();)d(m.currentNode)},[t,s,n,c])}var Qt=(t=>(t[t.Open=0]="Open",t[t.Closed=1]="Closed",t))(Qt||{}),Jt=(t=>(t[t.Pointer=0]="Pointer",t[t.Other=1]="Other",t))(Jt||{}),Kt=(t=>(t[t.OpenMenu=0]="OpenMenu",t[t.CloseMenu=1]="CloseMenu",t[t.GoToItem=2]="GoToItem",t[t.Search=3]="Search",t[t.ClearSearch=4]="ClearSearch",t[t.RegisterItem=5]="RegisterItem",t[t.UnregisterItem=6]="UnregisterItem",t))(Kt||{});function se(t,a=i=>i){let i=t.activeItemIndex!==null?t.items[t.activeItemIndex]:null,s=Ut(a(t.items.slice()),c=>c.dataRef.current.domRef.current),n=i?s.indexOf(i):null;return n===-1&&(n=null),{items:s,activeItemIndex:n}}let Gt={1(t){return t.menuState===1?t:{...t,activeItemIndex:null,menuState:1}},0(t){return t.menuState===0?t:{...t,__demoMode:!1,menuState:0}},2:(t,a)=>{var i;let s=se(t),n=Ot(a,{resolveItems:()=>s.items,resolveActiveIndex:()=>s.activeItemIndex,resolveId:c=>c.id,resolveDisabled:c=>c.dataRef.current.disabled});return{...t,...s,searchQuery:"",activeItemIndex:n,activationTrigger:(i=a.trigger)!=null?i:1}},3:(t,a)=>{let i=t.searchQuery!==""?0:1,s=t.searchQuery+a.value.toLowerCase(),n=(t.activeItemIndex!==null?t.items.slice(t.activeItemIndex+i).concat(t.items.slice(0,t.activeItemIndex+i)):t.items).find(l=>{var r;return((r=l.dataRef.current.textValue)==null?void 0:r.startsWith(s))&&!l.dataRef.current.disabled}),c=n?t.items.indexOf(n):-1;return c===-1||c===t.activeItemIndex?{...t,searchQuery:s}:{...t,searchQuery:s,activeItemIndex:c,activationTrigger:1}},4(t){return t.searchQuery===""?t:{...t,searchQuery:"",searchActiveItemIndex:null}},5:(t,a)=>{let i=se(t,s=>[...s,{id:a.id,dataRef:a.dataRef}]);return{...t,...i}},6:(t,a)=>{let i=se(t,s=>{let n=s.findIndex(c=>c.id===a.id);return n!==-1&&s.splice(n,1),s});return{...t,...i,activationTrigger:1}}},ae=h.createContext(null);ae.displayName="MenuContext";function Z(t){let a=h.useContext(ae);if(a===null){let i=new Error(`<${t} /> is missing a parent <Menu /> component.`);throw Error.captureStackTrace&&Error.captureStackTrace(i,Z),i}return a}function Xt(t,a){return Dt(a.type,Gt,t,a)}let Yt=h.Fragment;function Zt(t,a){let{__demoMode:i=!1,...s}=t,n=h.useReducer(Xt,{__demoMode:i,menuState:i?0:1,buttonRef:h.createRef(),itemsRef:h.createRef(),items:[],searchQuery:"",activeItemIndex:null,activationTrigger:1}),[{menuState:c,itemsRef:l,buttonRef:r},d]=n,x=X(a);Rt([r,l],(u,S)=>{var f;d({type:1}),Vt(S,_t.Loose)||(u.preventDefault(),(f=r.current)==null||f.focus())},c===0);let m=w(()=>{d({type:1})}),p=h.useMemo(()=>({open:c===0,close:m}),[c,m]),g={ref:x};return D.createElement(ae.Provider,{value:n},D.createElement(Bt,{value:Dt(c,{0:K.Open,1:K.Closed})},Y({ourProps:g,theirProps:s,slot:p,defaultTag:Yt,name:"Menu"})))}let er="button";function tr(t,a){var i;let s=oe(),{id:n=`headlessui-menu-button-${s}`,...c}=t,[l,r]=Z("Menu.Button"),d=X(l.buttonRef,a),x=kt(),m=w(f=>{switch(f.key){case b.Space:case b.Enter:case b.ArrowDown:f.preventDefault(),f.stopPropagation(),r({type:0}),x.nextFrame(()=>r({type:2,focus:I.First}));break;case b.ArrowUp:f.preventDefault(),f.stopPropagation(),r({type:0}),x.nextFrame(()=>r({type:2,focus:I.Last}));break}}),p=w(f=>{switch(f.key){case b.Space:f.preventDefault();break}}),g=w(f=>{if(At(f.currentTarget))return f.preventDefault();t.disabled||(l.menuState===0?(r({type:1}),x.nextFrame(()=>{var k;return(k=l.buttonRef.current)==null?void 0:k.focus({preventScroll:!0})})):(f.preventDefault(),r({type:0})))}),u=h.useMemo(()=>({open:l.menuState===0}),[l]),S={ref:d,id:n,type:Et(t,l.buttonRef),"aria-haspopup":"menu","aria-controls":(i=l.itemsRef.current)==null?void 0:i.id,"aria-expanded":l.menuState===0,onKeyDown:m,onKeyUp:p,onClick:g};return Y({ourProps:S,theirProps:c,slot:u,defaultTag:er,name:"Menu.Button"})}let rr="div",sr=de.RenderStrategy|de.Static;function ir(t,a){var i,s;let n=oe(),{id:c=`headlessui-menu-items-${n}`,...l}=t,[r,d]=Z("Menu.Items"),x=X(r.itemsRef,a),m=Pt(r.itemsRef),p=kt(),g=Lt(),u=g!==null?(g&K.Open)===K.Open:r.menuState===0;h.useEffect(()=>{let o=r.itemsRef.current;o&&r.menuState===0&&o!==(m==null?void 0:m.activeElement)&&o.focus({preventScroll:!0})},[r.menuState,r.itemsRef,m]),Ht({container:r.itemsRef.current,enabled:r.menuState===0,accept(o){return o.getAttribute("role")==="menuitem"?NodeFilter.FILTER_REJECT:o.hasAttribute("role")?NodeFilter.FILTER_SKIP:NodeFilter.FILTER_ACCEPT},walk(o){o.setAttribute("role","none")}});let S=w(o=>{var C,q;switch(p.dispose(),o.key){case b.Space:if(r.searchQuery!=="")return o.preventDefault(),o.stopPropagation(),d({type:3,value:o.key});case b.Enter:if(o.preventDefault(),o.stopPropagation(),d({type:1}),r.activeItemIndex!==null){let{dataRef:j}=r.items[r.activeItemIndex];(q=(C=j.current)==null?void 0:C.domRef.current)==null||q.click()}Mt(r.buttonRef.current);break;case b.ArrowDown:return o.preventDefault(),o.stopPropagation(),d({type:2,focus:I.Next});case b.ArrowUp:return o.preventDefault(),o.stopPropagation(),d({type:2,focus:I.Previous});case b.Home:case b.PageUp:return o.preventDefault(),o.stopPropagation(),d({type:2,focus:I.First});case b.End:case b.PageDown:return o.preventDefault(),o.stopPropagation(),d({type:2,focus:I.Last});case b.Escape:o.preventDefault(),o.stopPropagation(),d({type:1}),ne().nextFrame(()=>{var j;return(j=r.buttonRef.current)==null?void 0:j.focus({preventScroll:!0})});break;case b.Tab:o.preventDefault(),o.stopPropagation(),d({type:1}),ne().nextFrame(()=>{qt(r.buttonRef.current,o.shiftKey?ue.Previous:ue.Next)});break;default:o.key.length===1&&(d({type:3,value:o.key}),p.setTimeout(()=>d({type:4}),350));break}}),f=w(o=>{switch(o.key){case b.Space:o.preventDefault();break}}),k=h.useMemo(()=>({open:r.menuState===0}),[r]),_={"aria-activedescendant":r.activeItemIndex===null||(i=r.items[r.activeItemIndex])==null?void 0:i.id,"aria-labelledby":(s=r.buttonRef.current)==null?void 0:s.id,id:c,onKeyDown:S,onKeyUp:f,role:"menu",tabIndex:0,ref:x};return Y({ourProps:_,theirProps:l,slot:k,defaultTag:rr,features:sr,visible:u,name:"Menu.Items"})}let nr=h.Fragment;function or(t,a){let i=oe(),{id:s=`headlessui-menu-item-${i}`,disabled:n=!1,...c}=t,[l,r]=Z("Menu.Item"),d=l.activeItemIndex!==null?l.items[l.activeItemIndex].id===s:!1,x=h.useRef(null),m=X(a,x);J(()=>{if(l.__demoMode||l.menuState!==0||!d||l.activationTrigger===0)return;let j=ne();return j.requestAnimationFrame(()=>{var re,ce;(ce=(re=x.current)==null?void 0:re.scrollIntoView)==null||ce.call(re,{block:"nearest"})}),j.dispose},[l.__demoMode,x,d,l.menuState,l.activationTrigger,l.activeItemIndex]);let p=$t(x),g=h.useRef({disabled:n,domRef:x,get textValue(){return p()}});J(()=>{g.current.disabled=n},[g,n]),J(()=>(r({type:5,id:s,dataRef:g}),()=>r({type:6,id:s})),[g,s]);let u=w(()=>{r({type:1})}),S=w(j=>{if(n)return j.preventDefault();r({type:1}),Mt(l.buttonRef.current)}),f=w(()=>{if(n)return r({type:2,focus:I.Nothing});r({type:2,focus:I.Specific,id:s})}),k=Ft(),_=w(j=>k.update(j)),o=w(j=>{k.wasMoved(j)&&(n||d||r({type:2,focus:I.Specific,id:s,trigger:0}))}),C=w(j=>{k.wasMoved(j)&&(n||d&&r({type:2,focus:I.Nothing}))}),q=h.useMemo(()=>({active:d,disabled:n,close:u}),[d,n,u]);return Y({ourProps:{id:s,ref:m,role:"menuitem",tabIndex:n===!0?void 0:-1,"aria-disabled":n===!0?!0:void 0,disabled:void 0,onClick:S,onFocus:f,onPointerEnter:_,onMouseEnter:_,onPointerMove:o,onMouseMove:o,onPointerLeave:C,onMouseLeave:C},theirProps:c,slot:q,defaultTag:nr,name:"Menu.Item"})}let ar=G(Zt),lr=G(tr),cr=G(ir),dr=G(or),Q=Object.assign(ar,{Button:lr,Items:cr,Item:dr});const ur=St("absolute z-50 mt-2 rounded-md bg-white shadow-lg-premium ring-1 ring-black ring-opacity-5 focus:outline-none overflow-hidden",{variants:{size:{sm:"min-w-[160px]",md:"min-w-[200px]",lg:"min-w-[280px]"},align:{left:"left-0 origin-top-left",right:"right-0 origin-top-right"}},defaultVariants:{size:"md",align:"left"}}),mr=St("flex items-start gap-3 px-4 py-2.5 text-left transition-colors duration-150 cursor-pointer",{variants:{size:{sm:"text-sm",md:"text-sm",lg:"text-base"},disabled:{true:"cursor-not-allowed opacity-50",false:""},destructive:{true:"text-red-600 hover:bg-red-50 hover:text-red-700",false:"text-gray-900 hover:bg-navy-50 hover:text-navy-900"}},defaultVariants:{size:"md",disabled:!1,destructive:!1}}),pr=()=>e.jsx("svg",{className:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M5 13l4 4L19 7"})}),gr=()=>e.jsx("svg",{className:"w-4 h-4 ml-1",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M19 9l-7 7-7-7"})}),v=D.forwardRef(({items:t,trigger:a,onSelect:i,align:s="left",size:n="md",className:c,triggerClassName:l},r)=>{const d=D.useMemo(()=>{const m=[];let p={items:[]};return t.forEach((g,u)=>{g.section&&g.section!==p.section?(p.items.length>0&&m.push(p),p={section:g.section,items:[g]}):p.items.push(g),(g.divider||u===t.length-1)&&(m.push(p),p={items:[]})}),m},[t]),x=m=>{m.disabled||(m.onClick&&m.onClick(),i&&i(m))};return e.jsx(Q,{as:"div",className:"relative inline-block text-left",ref:r,children:({open:m})=>e.jsxs(e.Fragment,{children:[e.jsx(Q.Button,{className:H("inline-flex items-center",l),children:typeof a=="string"?e.jsxs("span",{className:"inline-flex items-center",children:[a,e.jsx(gr,{})]}):a}),e.jsx(Tt,{show:m,as:D.Fragment,enter:"transition ease-out duration-200",enterFrom:"opacity-0 scale-95",enterTo:"opacity-100 scale-100",leave:"transition ease-in duration-150",leaveFrom:"opacity-100 scale-100",leaveTo:"opacity-0 scale-95",children:e.jsx(Q.Items,{className:H(ur({size:n,align:s}),c),children:e.jsx("div",{className:"py-1",children:d.map((p,g)=>e.jsxs(D.Fragment,{children:[p.section&&e.jsx("div",{className:"px-4 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wider",children:p.section}),p.items.map(u=>e.jsx(Q.Item,{disabled:u.disabled,children:({active:S})=>e.jsxs("button",{type:"button",onClick:()=>x(u),className:H(mr({size:n,disabled:u.disabled,destructive:u.destructive}),S&&!u.disabled&&"bg-navy-50","w-full"),disabled:u.disabled,children:[u.selected?e.jsx("span",{className:"flex-shrink-0 w-4 h-4 text-navy-700",children:e.jsx(pr,{})}):u.icon?e.jsx("span",{className:"flex-shrink-0 w-4 h-4",children:u.icon}):e.jsx("span",{className:"flex-shrink-0 w-4 h-4"}),e.jsxs("div",{className:"flex-1 min-w-0",children:[e.jsx("div",{className:"font-medium truncate",children:u.label}),u.description&&e.jsx("div",{className:H("mt-0.5 text-xs text-gray-500 truncate",u.destructive&&"text-red-500"),children:u.description})]})]})},u.id)),g<d.length-1&&e.jsx("div",{className:"my-1 border-t border-gray-200"})]},g))})})})]})})});v.displayName="Dropdown";v.__docgenInfo={description:`Dropdown menu component for action lists and contextual menus.
Uses Headless UI Menu for full accessibility and keyboard navigation.

@example
\`\`\`tsx
<Dropdown
  trigger={<Button>Actions</Button>}
  items={[
    { id: '1', label: 'Edit', icon: <EditIcon /> },
    { id: '2', label: 'Delete', destructive: true, icon: <TrashIcon /> },
  ]}
  onSelect={(item) => console.log('Selected:', item)}
/>

<Dropdown
  trigger={<Button>User Menu</Button>}
  align="right"
  items={[
    { id: '1', label: 'Profile', section: 'Account' },
    { id: '2', label: 'Settings', divider: true },
    { id: '3', label: 'Sign Out', destructive: true },
  ]}
/>
\`\`\``,methods:[],displayName:"Dropdown",props:{items:{required:!0,tsType:{name:"Array",elements:[{name:"DropdownItem"}],raw:"DropdownItem[]"},description:"Array of dropdown items"},trigger:{required:!0,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Trigger element (button, link, etc.)"},onSelect:{required:!1,tsType:{name:"signature",type:"function",raw:"(item: DropdownItem) => void",signature:{arguments:[{type:{name:"DropdownItem"},name:"item"}],return:{name:"void"}}},description:"Callback when an item is selected"},align:{required:!1,tsType:{name:"union",raw:"'left' | 'right'",elements:[{name:"literal",value:"'left'"},{name:"literal",value:"'right'"}]},description:"Dropdown alignment relative to trigger",defaultValue:{value:"'left'",computed:!1}},size:{required:!1,tsType:{name:"union",raw:"'sm' | 'md' | 'lg'",elements:[{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"},{name:"literal",value:"'lg'"}]},description:"Dropdown size",defaultValue:{value:"'md'",computed:!1}},className:{required:!1,tsType:{name:"string"},description:"Additional class names for the menu panel"},triggerClassName:{required:!1,tsType:{name:"string"},description:"Additional class names for the trigger wrapper"}}};const $r={title:"Design System/Overlays/Dropdown",component:v,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{items:{control:"object",description:"Array of dropdown menu items"},trigger:{control:"text",description:"Trigger element (button, link, etc.)"},align:{control:"select",options:["left","right"],description:"Dropdown alignment relative to trigger"},size:{control:"select",options:["sm","md","lg"],description:"Dropdown size variant"},onSelect:{action:"selected",description:"Callback when an item is selected"}}},ee=()=>e.jsx("svg",{className:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"})}),z=()=>e.jsx("svg",{className:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"})}),te=()=>e.jsx("svg",{className:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"})}),le=()=>e.jsx("svg",{className:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"})}),M=()=>e.jsx("svg",{className:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"})}),Nt=()=>e.jsx("svg",{className:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"})}),N=()=>e.jsx("svg",{className:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"})}),zt=()=>e.jsxs("svg",{className:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:[e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"}),e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M15 12a3 3 0 11-6 0 3 3 0 016 0z"})]}),Ct=()=>e.jsx("svg",{className:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"})}),fr=[{id:"1",label:"Edit"},{id:"2",label:"Duplicate"},{id:"3",label:"Archive"},{id:"4",label:"Delete",destructive:!0}],vr=[{id:"1",label:"Profile",section:"Account"},{id:"2",label:"Settings",divider:!0},{id:"3",label:"Billing",section:"Organization"},{id:"4",label:"Team Members",divider:!0},{id:"5",label:"Documentation",section:"Support"},{id:"6",label:"Contact Support",divider:!0},{id:"7",label:"Sign Out",destructive:!0}],hr=[{id:"1",label:"Edit",icon:e.jsx(ee,{})},{id:"2",label:"Duplicate",icon:e.jsx(te,{})},{id:"3",label:"Archive",icon:e.jsx(le,{})},{id:"4",label:"Share",icon:e.jsx(M,{})},{id:"5",label:"Download",icon:e.jsx(Nt,{})},{id:"6",label:"Delete",icon:e.jsx(z,{}),destructive:!0}],xr=[{id:"1",label:"Private",description:"Only you can see this",icon:e.jsx(N,{})},{id:"2",label:"Team",description:"Visible to your team members",icon:e.jsx(M,{})},{id:"3",label:"Public",description:"Anyone with the link can view",icon:e.jsx(M,{})}],br=[{id:"1",label:"Edit",icon:e.jsx(ee,{})},{id:"2",label:"Duplicate",icon:e.jsx(te,{}),disabled:!0},{id:"3",label:"Archive",icon:e.jsx(le,{}),disabled:!0},{id:"4",label:"Share",icon:e.jsx(M,{})},{id:"5",label:"Download",icon:e.jsx(Nt,{})},{id:"6",label:"Delete",icon:e.jsx(z,{}),destructive:!0}],jr=[{id:"1",label:"English"},{id:"2",label:"Spanish"},{id:"3",label:"French"},{id:"4",label:"German"},{id:"5",label:"Italian"}],yr=[{id:"1",label:"View Details"},{id:"2",label:"Edit Properties"},{id:"3",label:"Duplicate",divider:!0},{id:"4",label:"Archive",description:"Hide from active list",destructive:!0},{id:"5",label:"Delete",description:"Permanently remove this item",icon:e.jsx(z,{}),destructive:!0}],wr=[{id:"1",label:"John Doe",description:"john.doe@example.com",icon:e.jsx(N,{})},{id:"2",label:"Jane Smith",description:"jane.smith@example.com",icon:e.jsx(N,{}),divider:!0},{id:"3",label:"Invite Team Member",description:"Send an email invitation",icon:e.jsx(M,{})}],ie=[{id:"1",label:"Edit",icon:e.jsx(ee,{})},{id:"2",label:"Duplicate",icon:e.jsx(te,{})},{id:"3",label:"Delete",icon:e.jsx(z,{}),destructive:!0}],me=[{id:"1",label:"Profile",icon:e.jsx(N,{})},{id:"2",label:"Settings",icon:e.jsx(zt,{})},{id:"3",label:"Sign Out",icon:e.jsx(Ct,{}),destructive:!0}],Ir=[{id:"1",label:"John Doe",description:"john.doe@example.com",icon:e.jsx(N,{}),divider:!0},{id:"2",label:"Profile",icon:e.jsx(N,{}),section:"Account"},{id:"3",label:"Settings",icon:e.jsx(zt,{}),divider:!0},{id:"4",label:"Documentation",section:"Help"},{id:"5",label:"Contact Support",divider:!0},{id:"6",label:"Sign Out",icon:e.jsx(Ct,{}),destructive:!0}],Sr=[{id:"1",label:"View Details",description:"See full consent information"},{id:"2",label:"Edit Permissions",description:"Modify granted access",divider:!0},{id:"3",label:"Extend Duration",description:"Extend expiration date",section:"Manage"},{id:"4",label:"Pause Access",description:"Temporarily suspend permissions",divider:!0},{id:"5",label:"Revoke Consent",description:"Permanently remove all access",icon:e.jsx(z,{}),destructive:!0}],kr=[{id:"1",label:"First Item",icon:e.jsx(ee,{})},{id:"2",label:"Second Item",icon:e.jsx(te,{}),description:"With description"},{id:"3",label:"Disabled Item",icon:e.jsx(le,{}),disabled:!0},{id:"4",label:"Selected Item",icon:e.jsx(M,{}),selected:!0,divider:!0},{id:"5",label:"Destructive Action",icon:e.jsx(z,{}),destructive:!0}],A={render:t=>e.jsx("div",{className:"h-64 flex items-start justify-center",children:e.jsx(v,{...t,items:t.items||fr,trigger:e.jsx(y,{children:"Actions"})})}),args:{align:"left",size:"md"}},R={render:()=>e.jsx("div",{className:"h-96 flex items-start justify-center",children:e.jsx(v,{items:vr,trigger:e.jsx(y,{children:"User Menu"}),onSelect:t=>console.log("Selected:",t.label)})}),args:{align:"left",size:"md"}},B={render:()=>e.jsx("div",{className:"h-80 flex items-start justify-center",children:e.jsx(v,{items:hr,trigger:e.jsx(y,{children:"Actions"}),onSelect:t=>console.log("Selected:",t.label)})}),args:{align:"left",size:"md"}},L={render:()=>e.jsx("div",{className:"h-80 flex items-start justify-center",children:e.jsx(v,{items:xr,trigger:e.jsx(y,{children:"Change Visibility"}),size:"lg",onSelect:t=>console.log("Selected:",t.label)})}),args:{align:"left",size:"lg"}},T={render:()=>e.jsx("div",{className:"h-80 flex items-start justify-center",children:e.jsx(v,{items:br,trigger:e.jsx(y,{children:"Actions"}),onSelect:t=>alert(`Selected: ${t.label}`)})}),args:{align:"left",size:"md"}},P={render:()=>{const[t,a]=h.useState("2"),i=jr.map(s=>({...s,selected:t===s.id}));return e.jsx("div",{className:"h-80 flex items-start justify-center",children:e.jsx(v,{items:i,trigger:e.jsx(y,{children:"Select Language"}),onSelect:s=>a(s.id)})})},args:{align:"left",size:"md"}},E={render:()=>e.jsx("div",{className:"h-80 flex items-start justify-center",children:e.jsx(v,{items:yr,trigger:e.jsx(y,{variant:"outline",children:"Manage Item"}),size:"lg",onSelect:t=>{t.destructive?confirm(`Are you sure you want to ${t.label.toLowerCase()}?`)&&alert(`${t.label} confirmed`):alert(`Selected: ${t.label}`)}})}),args:{align:"left",size:"lg"}},$={render:()=>e.jsx("div",{className:"h-80 flex items-start justify-center",children:e.jsx(v,{items:wr,trigger:e.jsx(y,{children:"Assign To"}),size:"lg",onSelect:t=>alert(`Assigned to: ${t.label}`)})}),args:{align:"left",size:"lg"}},F={render:()=>e.jsxs("div",{className:"h-80 flex items-start justify-center gap-4",children:[e.jsx(v,{items:ie,trigger:e.jsx(y,{size:"sm",children:"Small"}),size:"sm"}),e.jsx(v,{items:ie,trigger:e.jsx(y,{size:"md",children:"Medium"}),size:"md"}),e.jsx(v,{items:ie,trigger:e.jsx(y,{size:"lg",children:"Large"}),size:"lg"})]}),args:{align:"left"}},O={render:()=>e.jsxs("div",{className:"h-80 w-full flex items-start justify-between px-8",children:[e.jsx(v,{items:me,trigger:e.jsx(y,{children:"Align Left"}),align:"left"}),e.jsx(v,{items:me,trigger:e.jsx(y,{children:"Align Right"}),align:"right"})]}),args:{}},W={render:()=>e.jsx("div",{className:"h-96 flex items-start justify-end pr-8",children:e.jsx(v,{items:Ir,trigger:e.jsxs("button",{className:"flex items-center gap-2 px-3 py-2 rounded-md hover:bg-gray-100 transition-colors",children:[e.jsx("div",{className:"w-8 h-8 bg-navy-600 text-white rounded-full flex items-center justify-center text-sm font-semibold",children:"JD"}),e.jsx("svg",{className:"w-4 h-4 text-gray-600",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M19 9l-7 7-7-7"})})]}),align:"right",size:"md",onSelect:t=>{t.id==="6"?alert("Signing out..."):alert(`Selected: ${t.label}`)}})}),args:{}},U={render:()=>e.jsx("div",{className:"h-96 flex items-start justify-center",children:e.jsx(v,{items:Sr,trigger:e.jsx(y,{variant:"outline",children:"Manage Consent"}),size:"lg",onSelect:t=>{t.destructive?confirm("Are you sure you want to revoke this consent?")&&alert("Consent revoked"):alert(`Selected: ${t.label}`)}})}),args:{}},V={render:t=>e.jsx("div",{className:"h-96 flex items-start justify-center",children:e.jsx(v,{...t,items:t.items||kr,trigger:t.trigger||e.jsx(y,{children:"Open Dropdown"})})}),args:{align:"left",size:"md"}};var pe,ge,fe,ve,he;A.parameters={...A.parameters,docs:{...(pe=A.parameters)==null?void 0:pe.docs,source:{originalSource:`{
  render: args => {
    return <div className="h-64 flex items-start justify-center">
        <Dropdown {...args} items={args.items || defaultItems} trigger={<Button>Actions</Button>} />
      </div>;
  },
  args: {
    align: 'left',
    size: 'md'
  }
}`,...(fe=(ge=A.parameters)==null?void 0:ge.docs)==null?void 0:fe.source},description:{story:"Default dropdown with basic items",...(he=(ve=A.parameters)==null?void 0:ve.docs)==null?void 0:he.description}}};var xe,be,je,ye,we;R.parameters={...R.parameters,docs:{...(xe=R.parameters)==null?void 0:xe.docs,source:{originalSource:`{
  render: () => {
    return <div className="h-96 flex items-start justify-center">
        <Dropdown items={sectionsItems} trigger={<Button>User Menu</Button>} onSelect={item => console.log('Selected:', item.label)} />
      </div>;
  },
  args: {
    align: 'left',
    size: 'md'
  }
}`,...(je=(be=R.parameters)==null?void 0:be.docs)==null?void 0:je.source},description:{story:"Dropdown with section grouping",...(we=(ye=R.parameters)==null?void 0:ye.docs)==null?void 0:we.description}}};var Ie,Se,ke,De,Me;B.parameters={...B.parameters,docs:{...(Ie=B.parameters)==null?void 0:Ie.docs,source:{originalSource:`{
  render: () => {
    return <div className="h-80 flex items-start justify-center">
        <Dropdown items={withIconsItems} trigger={<Button>Actions</Button>} onSelect={item => console.log('Selected:', item.label)} />
      </div>;
  },
  args: {
    align: 'left',
    size: 'md'
  }
}`,...(ke=(Se=B.parameters)==null?void 0:Se.docs)==null?void 0:ke.source},description:{story:"Dropdown with icons",...(Me=(De=B.parameters)==null?void 0:De.docs)==null?void 0:Me.description}}};var Ne,ze,Ce,Ae,Re;L.parameters={...L.parameters,docs:{...(Ne=L.parameters)==null?void 0:Ne.docs,source:{originalSource:`{
  render: () => {
    return <div className="h-80 flex items-start justify-center">
        <Dropdown items={withDescriptionsItems} trigger={<Button>Change Visibility</Button>} size="lg" onSelect={item => console.log('Selected:', item.label)} />
      </div>;
  },
  args: {
    align: 'left',
    size: 'lg'
  }
}`,...(Ce=(ze=L.parameters)==null?void 0:ze.docs)==null?void 0:Ce.source},description:{story:"Dropdown items with descriptions",...(Re=(Ae=L.parameters)==null?void 0:Ae.docs)==null?void 0:Re.description}}};var Be,Le,Te,Pe,Ee;T.parameters={...T.parameters,docs:{...(Be=T.parameters)==null?void 0:Be.docs,source:{originalSource:`{
  render: () => {
    return <div className="h-80 flex items-start justify-center">
        <Dropdown items={disabledItems} trigger={<Button>Actions</Button>} onSelect={item => alert(\`Selected: \${item.label}\`)} />
      </div>;
  },
  args: {
    align: 'left',
    size: 'md'
  }
}`,...(Te=(Le=T.parameters)==null?void 0:Le.docs)==null?void 0:Te.source},description:{story:"Dropdown with disabled items",...(Ee=(Pe=T.parameters)==null?void 0:Pe.docs)==null?void 0:Ee.description}}};var $e,Fe,Oe,We,Ue;P.parameters={...P.parameters,docs:{...($e=P.parameters)==null?void 0:$e.docs,source:{originalSource:`{
  render: () => {
    const [selectedId, setSelectedId] = useState('2');
    const items = languageItems.map(item => ({
      ...item,
      selected: selectedId === item.id
    }));
    return <div className="h-80 flex items-start justify-center">
        <Dropdown items={items} trigger={<Button>Select Language</Button>} onSelect={item => setSelectedId(item.id)} />
      </div>;
  },
  args: {
    align: 'left',
    size: 'md'
  }
}`,...(Oe=(Fe=P.parameters)==null?void 0:Fe.docs)==null?void 0:Oe.source},description:{story:"Dropdown with checkmarks for selected items",...(Ue=(We=P.parameters)==null?void 0:We.docs)==null?void 0:Ue.description}}};var Ve,_e,qe,He,Qe;E.parameters={...E.parameters,docs:{...(Ve=E.parameters)==null?void 0:Ve.docs,source:{originalSource:`{
  render: () => {
    return <div className="h-80 flex items-start justify-center">
        <Dropdown items={destructiveItems} trigger={<Button variant="outline">Manage Item</Button>} size="lg" onSelect={item => {
        if (item.destructive) {
          if (confirm(\`Are you sure you want to \${item.label.toLowerCase()}?\`)) {
            alert(\`\${item.label} confirmed\`);
          }
        } else {
          alert(\`Selected: \${item.label}\`);
        }
      }} />
      </div>;
  },
  args: {
    align: 'left',
    size: 'lg'
  }
}`,...(qe=(_e=E.parameters)==null?void 0:_e.docs)==null?void 0:qe.source},description:{story:"Dropdown with destructive actions",...(Qe=(He=E.parameters)==null?void 0:He.docs)==null?void 0:Qe.description}}};var Je,Ke,Ge,Xe,Ye;$.parameters={...$.parameters,docs:{...(Je=$.parameters)==null?void 0:Je.docs,source:{originalSource:`{
  render: () => {
    return <div className="h-80 flex items-start justify-center">
        <Dropdown items={complexContentItems} trigger={<Button>Assign To</Button>} size="lg" onSelect={item => alert(\`Assigned to: \${item.label}\`)} />
      </div>;
  },
  args: {
    align: 'left',
    size: 'lg'
  }
}`,...(Ge=(Ke=$.parameters)==null?void 0:Ke.docs)==null?void 0:Ge.source},description:{story:"Dropdown with complex/rich content",...(Ye=(Xe=$.parameters)==null?void 0:Xe.docs)==null?void 0:Ye.description}}};var Ze,et,tt,rt,st;F.parameters={...F.parameters,docs:{...(Ze=F.parameters)==null?void 0:Ze.docs,source:{originalSource:`{
  render: () => {
    return <div className="h-80 flex items-start justify-center gap-4">
        <Dropdown items={sizesItems} trigger={<Button size="sm">Small</Button>} size="sm" />
        <Dropdown items={sizesItems} trigger={<Button size="md">Medium</Button>} size="md" />
        <Dropdown items={sizesItems} trigger={<Button size="lg">Large</Button>} size="lg" />
      </div>;
  },
  args: {
    align: 'left'
  }
}`,...(tt=(et=F.parameters)==null?void 0:et.docs)==null?void 0:tt.source},description:{story:"Dropdown sizes comparison",...(st=(rt=F.parameters)==null?void 0:rt.docs)==null?void 0:st.description}}};var it,nt,ot,at,lt;O.parameters={...O.parameters,docs:{...(it=O.parameters)==null?void 0:it.docs,source:{originalSource:`{
  render: () => {
    return <div className="h-80 w-full flex items-start justify-between px-8">
        <Dropdown items={alignmentItems} trigger={<Button>Align Left</Button>} align="left" />
        <Dropdown items={alignmentItems} trigger={<Button>Align Right</Button>} align="right" />
      </div>;
  },
  args: {}
}`,...(ot=(nt=O.parameters)==null?void 0:nt.docs)==null?void 0:ot.source},description:{story:"Dropdown alignment options",...(lt=(at=O.parameters)==null?void 0:at.docs)==null?void 0:lt.description}}};var ct,dt,ut,mt,pt;W.parameters={...W.parameters,docs:{...(ct=W.parameters)==null?void 0:ct.docs,source:{originalSource:`{
  render: () => {
    return <div className="h-96 flex items-start justify-end pr-8">
        <Dropdown items={userAccountMenuItems} trigger={<button className="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-gray-100 transition-colors">
              <div className="w-8 h-8 bg-navy-600 text-white rounded-full flex items-center justify-center text-sm font-semibold">
                JD
              </div>
              <svg className="w-4 h-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </button>} align="right" size="md" onSelect={item => {
        if (item.id === '6') {
          alert('Signing out...');
        } else {
          alert(\`Selected: \${item.label}\`);
        }
      }} />
      </div>;
  },
  args: {}
}`,...(ut=(dt=W.parameters)==null?void 0:dt.docs)==null?void 0:ut.source},description:{story:"User account menu (real-world example)",...(pt=(mt=W.parameters)==null?void 0:mt.docs)==null?void 0:pt.description}}};var gt,ft,vt,ht,xt;U.parameters={...U.parameters,docs:{...(gt=U.parameters)==null?void 0:gt.docs,source:{originalSource:`{
  render: () => {
    return <div className="h-96 flex items-start justify-center">
        <Dropdown items={consentActionsMenuItems} trigger={<Button variant="outline">Manage Consent</Button>} size="lg" onSelect={item => {
        if (item.destructive) {
          if (confirm('Are you sure you want to revoke this consent?')) {
            alert('Consent revoked');
          }
        } else {
          alert(\`Selected: \${item.label}\`);
        }
      }} />
      </div>;
  },
  args: {}
}`,...(vt=(ft=U.parameters)==null?void 0:ft.docs)==null?void 0:vt.source},description:{story:"Consent management actions dropdown",...(xt=(ht=U.parameters)==null?void 0:ht.docs)==null?void 0:xt.description}}};var bt,jt,yt,wt,It;V.parameters={...V.parameters,docs:{...(bt=V.parameters)==null?void 0:bt.docs,source:{originalSource:`{
  render: args => {
    return <div className="h-96 flex items-start justify-center">
        <Dropdown {...args} items={args.items || playgroundItems} trigger={args.trigger || <Button>Open Dropdown</Button>} />
      </div>;
  },
  args: {
    align: 'left',
    size: 'md'
  }
}`,...(yt=(jt=V.parameters)==null?void 0:jt.docs)==null?void 0:yt.source},description:{story:"Interactive playground with all controls",...(It=(wt=V.parameters)==null?void 0:wt.docs)==null?void 0:It.description}}};const Fr=["Default","Sections","WithIcons","WithDescriptions","DisabledItems","WithCheckmarks","Destructive","ComplexContent","Sizes","Alignment","UserAccountMenu","ConsentActionsMenu","Playground"];export{O as Alignment,$ as ComplexContent,U as ConsentActionsMenu,A as Default,E as Destructive,T as DisabledItems,V as Playground,R as Sections,F as Sizes,W as UserAccountMenu,P as WithCheckmarks,L as WithDescriptions,B as WithIcons,Fr as __namedExportsOrder,$r as default};
