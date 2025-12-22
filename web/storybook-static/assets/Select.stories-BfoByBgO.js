import{j as a}from"./jsx-runtime-BYYWji4R.js";import{r as i,R as q}from"./index-ClcD9ViR.js";import{c as ge,a as xe}from"./cn-JCLedEej.js";import{d as ot,l as z,U as V,y as F,o as h,u as U,x as vt,C as G,I as ce,a as fe,O as Ce,c as w}from"./keyboard-DNQdDreo.js";import{T as ht,e as ft}from"./form-yc_mpN4r.js";import{p as oe,r as xt}from"./bugs-Ot56VXf3.js";import{y as St,s as yt,u as Ot,d as ie,q as Ct}from"./transition-CF-Dw_VV.js";import{T as wt}from"./use-resolve-button-type-Bw633SU1.js";import{s as Nt,u as Rt,f as jt,c as I}from"./use-text-value-DhDQuHs4.js";import{u as Tt,s as Lt}from"./hidden-BnOBa8gi.js";import{h as At,T as Pt,o as It,I as Et}from"./use-is-mounted-62Q7yJH9.js";import"./_commonjsHelpers-Cpj98o6Y.js";function it(e,t){let[l,s]=i.useState(e),r=ot(e);return z(()=>s(r.current),[r,s,...t]),l}var kt=(e=>(e[e.Open=0]="Open",e[e.Closed=1]="Closed",e))(kt||{}),$t=(e=>(e[e.Single=0]="Single",e[e.Multi=1]="Multi",e))($t||{}),Mt=(e=>(e[e.Pointer=0]="Pointer",e[e.Other=1]="Other",e))(Mt||{}),Dt=(e=>(e[e.OpenListbox=0]="OpenListbox",e[e.CloseListbox=1]="CloseListbox",e[e.GoToOption=2]="GoToOption",e[e.Search=3]="Search",e[e.ClearSearch=4]="ClearSearch",e[e.RegisterOption=5]="RegisterOption",e[e.UnregisterOption=6]="UnregisterOption",e[e.RegisterLabel=7]="RegisterLabel",e))(Dt||{});function ve(e,t=l=>l){let l=e.activeOptionIndex!==null?e.options[e.activeOptionIndex]:null,s=Et(t(e.options.slice()),u=>u.dataRef.current.domRef.current),r=l?s.indexOf(l):null;return r===-1&&(r=null),{options:s,activeOptionIndex:r}}let Ut={1(e){return e.dataRef.current.disabled||e.listboxState===1?e:{...e,activeOptionIndex:null,listboxState:1}},0(e){if(e.dataRef.current.disabled||e.listboxState===0)return e;let t=e.activeOptionIndex,{isSelected:l}=e.dataRef.current,s=e.options.findIndex(r=>l(r.dataRef.current.value));return s!==-1&&(t=s),{...e,listboxState:0,activeOptionIndex:t}},2(e,t){var l;if(e.dataRef.current.disabled||e.listboxState===1)return e;let s=ve(e),r=jt(t,{resolveItems:()=>s.options,resolveActiveIndex:()=>s.activeOptionIndex,resolveId:u=>u.id,resolveDisabled:u=>u.dataRef.current.disabled});return{...e,...s,searchQuery:"",activeOptionIndex:r,activationTrigger:(l=t.trigger)!=null?l:1}},3:(e,t)=>{if(e.dataRef.current.disabled||e.listboxState===1)return e;let l=e.searchQuery!==""?0:1,s=e.searchQuery+t.value.toLowerCase(),r=(e.activeOptionIndex!==null?e.options.slice(e.activeOptionIndex+l).concat(e.options.slice(0,e.activeOptionIndex+l)):e.options).find(n=>{var c;return!n.dataRef.current.disabled&&((c=n.dataRef.current.textValue)==null?void 0:c.startsWith(s))}),u=r?e.options.indexOf(r):-1;return u===-1||u===e.activeOptionIndex?{...e,searchQuery:s}:{...e,searchQuery:s,activeOptionIndex:u,activationTrigger:1}},4(e){return e.dataRef.current.disabled||e.listboxState===1||e.searchQuery===""?e:{...e,searchQuery:""}},5:(e,t)=>{let l={id:t.id,dataRef:t.dataRef},s=ve(e,r=>[...r,l]);return e.activeOptionIndex===null&&e.dataRef.current.isSelected(t.dataRef.current.value)&&(s.activeOptionIndex=s.options.indexOf(l)),{...e,...s}},6:(e,t)=>{let l=ve(e,s=>{let r=s.findIndex(u=>u.id===t.id);return r!==-1&&s.splice(r,1),s});return{...e,...l,activationTrigger:1}},7:(e,t)=>({...e,labelId:t.id})},Se=i.createContext(null);Se.displayName="ListboxActionsContext";function H(e){let t=i.useContext(Se);if(t===null){let l=new Error(`<${e} /> is missing a parent <Listbox /> component.`);throw Error.captureStackTrace&&Error.captureStackTrace(l,H),l}return t}let ye=i.createContext(null);ye.displayName="ListboxDataContext";function B(e){let t=i.useContext(ye);if(t===null){let l=new Error(`<${e} /> is missing a parent <Listbox /> component.`);throw Error.captureStackTrace&&Error.captureStackTrace(l,B),l}return t}function qt(e,t){return U(t.type,Ut,e,t)}let zt=i.Fragment;function Vt(e,t){let{value:l,defaultValue:s,form:r,name:u,onChange:n,by:c=(p,b)=>p===b,disabled:m=!1,horizontal:f=!1,multiple:C=!1,...L}=e;const k=f?"horizontal":"vertical";let E=F(t),[S=C?[]:void 0,R]=ht(l,n,s),[x,o]=i.useReducer(qt,{dataRef:i.createRef(),listboxState:1,options:[],searchQuery:"",labelId:null,activeOptionIndex:null,activationTrigger:1}),y=i.useRef({static:!1,hold:!1}),$=i.useRef(null),M=i.useRef(null),D=i.useRef(null),O=h(typeof c=="string"?(p,b)=>{let T=c;return(p==null?void 0:p[T])===(b==null?void 0:b[T])}:c),j=i.useCallback(p=>U(g.mode,{1:()=>S.some(b=>O(b,p)),0:()=>O(S,p)}),[S]),g=i.useMemo(()=>({...x,value:S,disabled:m,mode:C?1:0,orientation:k,compare:O,isSelected:j,optionsPropsRef:y,labelRef:$,buttonRef:M,optionsRef:D}),[S,m,C,x]);z(()=>{x.dataRef.current=g},[g]),St([g.buttonRef,g.optionsRef],(p,b)=>{var T;o({type:1}),At(b,Pt.Loose)||(p.preventDefault(),(T=g.buttonRef.current)==null||T.focus())},g.listboxState===0);let ue=i.useMemo(()=>({open:g.listboxState===0,disabled:m,value:S}),[g,m,S]),de=h(p=>{let b=g.options.find(T=>T.id===p);b&&be(b.dataRef.current.value)}),pe=h(()=>{if(g.activeOptionIndex!==null){let{dataRef:p,id:b}=g.options[g.activeOptionIndex];be(p.current.value),o({type:2,focus:I.Specific,id:b})}}),d=h(()=>o({type:0})),P=h(()=>o({type:1})),me=h((p,b,T)=>p===I.Specific?o({type:2,focus:I.Specific,id:b,trigger:T}):o({type:2,focus:p,trigger:T})),ct=h((p,b)=>(o({type:5,id:p,dataRef:b}),()=>o({type:6,id:p}))),ut=h(p=>(o({type:7,id:p}),()=>o({type:7,id:null}))),be=h(p=>U(g.mode,{0(){return R==null?void 0:R(p)},1(){let b=g.value.slice(),T=b.findIndex(Q=>O(Q,p));return T===-1?b.push(p):b.splice(T,1),R==null?void 0:R(b)}})),dt=h(p=>o({type:3,value:p})),pt=h(()=>o({type:4})),mt=i.useMemo(()=>({onChange:be,registerOption:ct,registerLabel:ut,goToOption:me,closeListbox:P,openListbox:d,selectActiveOption:pe,selectOption:de,search:dt,clearSearch:pt}),[]),bt={ref:E},W=i.useRef(null),gt=oe();return i.useEffect(()=>{W.current&&s!==void 0&&gt.addEventListener(W.current,"reset",()=>{R==null||R(s)})},[W,R]),q.createElement(Se.Provider,{value:mt},q.createElement(ye.Provider,{value:g},q.createElement(yt,{value:U(g.listboxState,{0:ie.Open,1:ie.Closed})},u!=null&&S!=null&&ft({[u]:S}).map(([p,b],T)=>q.createElement(Tt,{features:Lt.Hidden,ref:T===0?Q=>{var Oe;W.current=(Oe=Q==null?void 0:Q.closest("form"))!=null?Oe:null}:void 0,...vt({key:p,as:"input",type:"hidden",hidden:!0,readOnly:!0,form:r,disabled:m,name:p,value:b})})),G({ourProps:bt,theirProps:L,slot:ue,defaultTag:zt,name:"Listbox"}))))}let Ft="button";function Gt(e,t){var l;let s=ce(),{id:r=`headlessui-listbox-button-${s}`,...u}=e,n=B("Listbox.Button"),c=H("Listbox.Button"),m=F(n.buttonRef,t),f=oe(),C=h(x=>{switch(x.key){case w.Space:case w.Enter:case w.ArrowDown:x.preventDefault(),c.openListbox(),f.nextFrame(()=>{n.value||c.goToOption(I.First)});break;case w.ArrowUp:x.preventDefault(),c.openListbox(),f.nextFrame(()=>{n.value||c.goToOption(I.Last)});break}}),L=h(x=>{switch(x.key){case w.Space:x.preventDefault();break}}),k=h(x=>{if(xt(x.currentTarget))return x.preventDefault();n.listboxState===0?(c.closeListbox(),f.nextFrame(()=>{var o;return(o=n.buttonRef.current)==null?void 0:o.focus({preventScroll:!0})})):(x.preventDefault(),c.openListbox())}),E=it(()=>{if(n.labelId)return[n.labelId,r].join(" ")},[n.labelId,r]),S=i.useMemo(()=>({open:n.listboxState===0,disabled:n.disabled,value:n.value}),[n]),R={ref:m,id:r,type:wt(e,n.buttonRef),"aria-haspopup":"listbox","aria-controls":(l=n.optionsRef.current)==null?void 0:l.id,"aria-expanded":n.listboxState===0,"aria-labelledby":E,disabled:n.disabled,onKeyDown:C,onKeyUp:L,onClick:k};return G({ourProps:R,theirProps:u,slot:S,defaultTag:Ft,name:"Listbox.Button"})}let Ht="label";function Bt(e,t){let l=ce(),{id:s=`headlessui-listbox-label-${l}`,...r}=e,u=B("Listbox.Label"),n=H("Listbox.Label"),c=F(u.labelRef,t);z(()=>n.registerLabel(s),[s]);let m=h(()=>{var C;return(C=u.buttonRef.current)==null?void 0:C.focus({preventScroll:!0})}),f=i.useMemo(()=>({open:u.listboxState===0,disabled:u.disabled}),[u]);return G({ourProps:{ref:c,id:s,onClick:m},theirProps:r,slot:f,defaultTag:Ht,name:"Listbox.Label"})}let Wt="ul",Qt=Ce.RenderStrategy|Ce.Static;function Yt(e,t){var l;let s=ce(),{id:r=`headlessui-listbox-options-${s}`,...u}=e,n=B("Listbox.Options"),c=H("Listbox.Options"),m=F(n.optionsRef,t),f=oe(),C=oe(),L=Ot(),k=L!==null?(L&ie.Open)===ie.Open:n.listboxState===0;i.useEffect(()=>{var o;let y=n.optionsRef.current;y&&n.listboxState===0&&y!==((o=It(y))==null?void 0:o.activeElement)&&y.focus({preventScroll:!0})},[n.listboxState,n.optionsRef]);let E=h(o=>{switch(C.dispose(),o.key){case w.Space:if(n.searchQuery!=="")return o.preventDefault(),o.stopPropagation(),c.search(o.key);case w.Enter:if(o.preventDefault(),o.stopPropagation(),n.activeOptionIndex!==null){let{dataRef:y}=n.options[n.activeOptionIndex];c.onChange(y.current.value)}n.mode===0&&(c.closeListbox(),fe().nextFrame(()=>{var y;return(y=n.buttonRef.current)==null?void 0:y.focus({preventScroll:!0})}));break;case U(n.orientation,{vertical:w.ArrowDown,horizontal:w.ArrowRight}):return o.preventDefault(),o.stopPropagation(),c.goToOption(I.Next);case U(n.orientation,{vertical:w.ArrowUp,horizontal:w.ArrowLeft}):return o.preventDefault(),o.stopPropagation(),c.goToOption(I.Previous);case w.Home:case w.PageUp:return o.preventDefault(),o.stopPropagation(),c.goToOption(I.First);case w.End:case w.PageDown:return o.preventDefault(),o.stopPropagation(),c.goToOption(I.Last);case w.Escape:return o.preventDefault(),o.stopPropagation(),c.closeListbox(),f.nextFrame(()=>{var y;return(y=n.buttonRef.current)==null?void 0:y.focus({preventScroll:!0})});case w.Tab:o.preventDefault(),o.stopPropagation();break;default:o.key.length===1&&(c.search(o.key),C.setTimeout(()=>c.clearSearch(),350));break}}),S=it(()=>{var o;return(o=n.buttonRef.current)==null?void 0:o.id},[n.buttonRef.current]),R=i.useMemo(()=>({open:n.listboxState===0}),[n]),x={"aria-activedescendant":n.activeOptionIndex===null||(l=n.options[n.activeOptionIndex])==null?void 0:l.id,"aria-multiselectable":n.mode===1?!0:void 0,"aria-labelledby":S,"aria-orientation":n.orientation,id:r,onKeyDown:E,role:"listbox",tabIndex:0,ref:m};return G({ourProps:x,theirProps:u,slot:R,defaultTag:Wt,features:Qt,visible:k,name:"Listbox.Options"})}let Kt="li";function Jt(e,t){let l=ce(),{id:s=`headlessui-listbox-option-${l}`,disabled:r=!1,value:u,...n}=e,c=B("Listbox.Option"),m=H("Listbox.Option"),f=c.activeOptionIndex!==null?c.options[c.activeOptionIndex].id===s:!1,C=c.isSelected(u),L=i.useRef(null),k=Nt(L),E=ot({disabled:r,value:u,domRef:L,get textValue(){return k()}}),S=F(t,L);z(()=>{if(c.listboxState!==0||!f||c.activationTrigger===0)return;let O=fe();return O.requestAnimationFrame(()=>{var j,g;(g=(j=L.current)==null?void 0:j.scrollIntoView)==null||g.call(j,{block:"nearest"})}),O.dispose},[L,f,c.listboxState,c.activationTrigger,c.activeOptionIndex]),z(()=>m.registerOption(s,E),[E,s]);let R=h(O=>{if(r)return O.preventDefault();m.onChange(u),c.mode===0&&(m.closeListbox(),fe().nextFrame(()=>{var j;return(j=c.buttonRef.current)==null?void 0:j.focus({preventScroll:!0})}))}),x=h(()=>{if(r)return m.goToOption(I.Nothing);m.goToOption(I.Specific,s)}),o=Rt(),y=h(O=>o.update(O)),$=h(O=>{o.wasMoved(O)&&(r||f||m.goToOption(I.Specific,s,0))}),M=h(O=>{o.wasMoved(O)&&(r||f&&m.goToOption(I.Nothing))}),D=i.useMemo(()=>({active:f,selected:C,disabled:r}),[f,C,r]);return G({ourProps:{id:s,ref:S,role:"option",tabIndex:r===!0?void 0:-1,"aria-disabled":r===!0?!0:void 0,"aria-selected":C,disabled:void 0,onClick:R,onFocus:x,onPointerEnter:y,onMouseEnter:y,onPointerMove:$,onMouseMove:$,onPointerLeave:M,onMouseLeave:M},theirProps:n,slot:D,defaultTag:Kt,name:"Listbox.Option"})}let _t=V(Vt),Xt=V(Gt),Zt=V(Bt),ea=V(Yt),ta=V(Jt),Y=Object.assign(_t,{Button:Xt,Label:Zt,Options:ea,Option:ta});const aa=xe("relative w-full text-left bg-white border-1.5 rounded-lg transition-all focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-trust-deep disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-between px-3 py-2",{variants:{size:{sm:"text-sm py-1.5 px-2.5",md:"text-base py-2 px-3",lg:"text-lg py-2.5 px-3.5"},variant:{default:"border-gray-300 hover:border-gray-400",error:"border-error-primary bg-error-light/20",success:"border-success-primary bg-success-light/20"},open:{true:"rounded-b-none",false:""}},defaultVariants:{size:"md",variant:"default",open:!1}}),la=xe("absolute top-full left-0 right-0 z-50 w-full bg-white border-1.5 border-t-0 border-gray-300 rounded-b-lg shadow-lg focus:outline-none max-h-60 overflow-y-auto",{variants:{size:{sm:"text-sm",md:"text-base",lg:"text-lg"}},defaultVariants:{size:"md"}}),sa=xe("relative cursor-pointer select-none py-2 px-3 flex items-center justify-between transition-colors",{variants:{selected:{true:"bg-trust-light text-trust-deep font-medium",false:"text-gray-900 hover:bg-gray-50"},disabled:{true:"opacity-50 cursor-not-allowed",false:""}},defaultVariants:{selected:!1,disabled:!1}}),v=q.forwardRef(({options:e,value:t,onChange:l,multiselect:s=!1,label:r,helperText:u,errorMessage:n,successMessage:c,placeholder:m="Select...",disabled:f=!1,searchable:C=!1,required:L=!1,size:k="md",renderOption:E,renderLabel:S,id:R,className:x,...o},y)=>{const[$,M]=i.useState(!1),[D,O]=i.useState(""),j=i.useMemo(()=>e.flatMap(d=>"options"in d?d.options:d),[e]),g=i.useMemo(()=>D?j.filter(d=>d.label.toLowerCase().includes(D.toLowerCase())):j,[j,D]),ue=i.useMemo(()=>{if(S)return S(t);if(s&&Array.isArray(t)){if(t.length===0)return m;if(t.length===1){const d=j.find(P=>P.value===t[0]);return(d==null?void 0:d.label)||m}return`${t.length} selected`}if(t){const d=j.find(P=>P.value===t);return(d==null?void 0:d.label)||m}return m},[t,j,s,m,S]),de=n?"error":c?"success":"default",pe=d=>{if(s)if(Array.isArray(d))l(d);else{const P=Array.isArray(t)?t:[];P.includes(d)?l(P.filter(me=>me!==d)):l([...P,d])}else l(d),M(!1),O("")};return a.jsxs("div",{ref:y,className:ge("flex flex-col gap-1.5",x),...o,children:[r&&a.jsx("label",{htmlFor:R,className:ge("text-sm font-medium",f?"text-gray-500":"text-gray-900",L&&"after:content-['*'] after:ml-1 after:text-error-primary"),children:r}),a.jsx("div",{className:"relative w-full",children:a.jsxs(Y,{value:t,onChange:pe,disabled:f,multiple:s,as:"div",className:"relative",children:[a.jsxs(Y.Button,{className:aa({size:k,variant:de,open:$}),onClick:()=>M(!$),onFocus:()=>C&&M(!0),children:[a.jsx("span",{className:"block truncate",children:ue}),a.jsx("svg",{className:ge("w-5 h-5 transition-transform flex-shrink-0",$&&"rotate-180"),fill:"none",viewBox:"0 0 20 20",stroke:"currentColor","aria-hidden":"true",children:a.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:1.5,d:"M7 8l3 3 3-3m-3 3V4"})})]}),a.jsx(Ct,{show:$&&!f,enter:"transition ease-out duration-100",enterFrom:"transform opacity-0 scale-95",enterTo:"transform opacity-100 scale-100",leave:"transition ease-in duration-75",leaveFrom:"transform opacity-100 scale-100",leaveTo:"transform opacity-0 scale-95",className:"relative",children:a.jsxs(Y.Options,{className:la({size:k}),static:!0,children:[C&&a.jsx("div",{className:"sticky top-0 bg-white border-b border-gray-200 p-2",children:a.jsx("input",{type:"text",placeholder:"Search...",value:D,onChange:d=>O(d.target.value),className:"w-full px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:border-trust-deep",onClick:d=>d.stopPropagation()})}),g.length>0?g.map(d=>{const P=s?Array.isArray(t)&&t.includes(d.value):t===d.value;return a.jsx(Y.Option,{value:d.value,disabled:d.disabled,as:"div",className:sa({selected:P,disabled:d.disabled}),children:E?E(d,P):a.jsxs(a.Fragment,{children:[a.jsx("span",{children:d.label}),P&&a.jsx("svg",{className:"w-5 h-5 flex-shrink-0",fill:"currentColor",viewBox:"0 0 20 20","aria-hidden":"true",children:a.jsx("path",{fillRule:"evenodd",d:"M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z",clipRule:"evenodd"})})]})},d.value)}):a.jsx("div",{className:"px-3 py-2 text-sm text-gray-500",children:"No options found"})]})}),$&&a.jsx("div",{className:"fixed inset-0 z-40",onClick:()=>{M(!1),O("")}})]})}),a.jsxs("div",{className:"flex items-center gap-1.5 min-h-5",children:[n&&a.jsx("p",{className:"text-xs text-error-primary font-medium",children:n}),c&&!n&&a.jsx("p",{className:"text-xs text-success-primary font-medium",children:c}),u&&!n&&!c&&a.jsx("p",{className:"text-xs text-gray-600",children:u})]})]})});v.displayName="Select";v.__docgenInfo={description:`Select component for choosing from a list of options.
Supports single and multi-select with search capability.

@example
\`\`\`tsx
<Select
  options={[
    { value: 'option1', label: 'Option 1' },
    { value: 'option2', label: 'Option 2' },
  ]}
  value={selected}
  onChange={setSelected}
  label="Choose an option"
  placeholder="Select one..."
/>
\`\`\``,methods:[],displayName:"Select",props:{options:{required:!0,tsType:{name:"Array",elements:[{name:"unknown"}],raw:"(SelectOption | SelectOptionGroup)[]"},description:"Array of options or grouped options"},value:{required:!0,tsType:{name:"union",raw:"string | number | (string | number)[] | null",elements:[{name:"string"},{name:"number"},{name:"Array",elements:[{name:"unknown"}],raw:"(string | number)[]"},{name:"null"}]},description:"Selected value(s)"},onChange:{required:!0,tsType:{name:"signature",type:"function",raw:"(value: string | number | (string | number)[] | null) => void",signature:{arguments:[{type:{name:"union",raw:"string | number | (string | number)[] | null",elements:[{name:"string"},{name:"number"},{name:"Array",elements:[{name:"unknown"}],raw:"(string | number)[]"},{name:"null"}]},name:"value"}],return:{name:"void"}}},description:"Callback when value changes"},multiselect:{required:!1,tsType:{name:"boolean"},description:"Enable multi-select mode",defaultValue:{value:"false",computed:!1}},label:{required:!1,tsType:{name:"string"},description:"Label text"},helperText:{required:!1,tsType:{name:"string"},description:"Helper text"},errorMessage:{required:!1,tsType:{name:"string"},description:"Error message"},successMessage:{required:!1,tsType:{name:"string"},description:"Success message"},placeholder:{required:!1,tsType:{name:"string"},description:"Placeholder text",defaultValue:{value:"'Select...'",computed:!1}},disabled:{required:!1,tsType:{name:"boolean"},description:"Disable the select",defaultValue:{value:"false",computed:!1}},searchable:{required:!1,tsType:{name:"boolean"},description:"Enable search/filter",defaultValue:{value:"false",computed:!1}},required:{required:!1,tsType:{name:"boolean"},description:"Required field indicator",defaultValue:{value:"false",computed:!1}},size:{required:!1,tsType:{name:"union",raw:"'sm' | 'md' | 'lg'",elements:[{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"},{name:"literal",value:"'lg'"}]},description:"Component size",defaultValue:{value:"'md'",computed:!1}},renderOption:{required:!1,tsType:{name:"signature",type:"function",raw:"(option: SelectOption, isSelected: boolean) => React.ReactNode",signature:{arguments:[{type:{name:"SelectOption"},name:"option"},{type:{name:"boolean"},name:"isSelected"}],return:{name:"ReactReactNode",raw:"React.ReactNode"}}},description:"Custom option renderer"},renderLabel:{required:!1,tsType:{name:"signature",type:"function",raw:"(value: string | number | (string | number)[] | null) => React.ReactNode",signature:{arguments:[{type:{name:"union",raw:"string | number | (string | number)[] | null",elements:[{name:"string"},{name:"number"},{name:"Array",elements:[{name:"unknown"}],raw:"(string | number)[]"},{name:"null"}]},name:"value"}],return:{name:"ReactReactNode",raw:"React.ReactNode"}}},description:"Custom label renderer"},id:{required:!1,tsType:{name:"string"},description:"Unique identifier"}},composes:["Omit"]};const ha={title:"Design System/Inputs/Select",component:v,parameters:{layout:"centered"},tags:["autodocs"]},A=[{value:"option1",label:"Option 1"},{value:"option2",label:"Option 2"},{value:"option3",label:"Option 3"},{value:"option4",label:"Option 4",disabled:!0}],he=[{value:"viewer",label:"Viewer - Read-only access"},{value:"editor",label:"Editor - Can modify content"},{value:"admin",label:"Admin - Full permissions"},{value:"owner",label:"Owner - Ownership control"}],we=[{value:"us",label:"United States"},{value:"ca",label:"Canada"},{value:"uk",label:"United Kingdom"},{value:"au",label:"Australia"},{value:"de",label:"Germany"},{value:"fr",label:"France"},{value:"jp",label:"Japan"},{value:"sg",label:"Singapore"}],N=e=>t=>{(typeof t=="string"||typeof t=="number"||t===null)&&e(t)},K={render:()=>{const[e,t]=i.useState(null);return a.jsx(v,{options:A,value:e,onChange:N(t),label:"Choose an option",placeholder:"Select one...",helperText:"This is a basic select dropdown"})},args:{options:A,value:null,onChange:()=>{}}},J={render:()=>{const[e,t]=i.useState(null),[l,s]=i.useState(null),[r,u]=i.useState(null);return a.jsxs("div",{className:"space-y-6 w-96",children:[a.jsx(v,{size:"sm",options:A,value:e,onChange:N(t),label:"Small Select",placeholder:"Select..."}),a.jsx(v,{size:"md",options:A,value:l,onChange:N(s),label:"Medium Select (default)",placeholder:"Select..."}),a.jsx(v,{size:"lg",options:A,value:r,onChange:N(u),label:"Large Select",placeholder:"Select..."})]})},args:{options:A,value:null,onChange:()=>{}}},_={render:()=>{const[e,t]=i.useState(null),[l,s]=i.useState(null),[r,u]=i.useState("option2"),[n]=i.useState(null);return a.jsxs("div",{className:"space-y-6 w-96",children:[a.jsx(v,{options:A,value:e,onChange:N(t),label:"Normal State",placeholder:"Select an option..."}),a.jsx(v,{options:A,value:l,onChange:N(s),label:"Error State",placeholder:"Select an option...",errorMessage:"Please select a valid option"}),a.jsx(v,{options:A,value:r,onChange:N(u),label:"Success State",placeholder:"Select an option...",successMessage:"Option selected successfully"}),a.jsx(v,{options:A,value:n,onChange:()=>{},label:"Disabled Select",placeholder:"Select an option...",disabled:!0,helperText:"This select is disabled"})]})},args:{options:A,value:null,onChange:()=>{}}},X={render:()=>{const[e,t]=i.useState([]);return a.jsxs("div",{className:"w-96",children:[a.jsx(v,{multiselect:!0,options:he,value:e,onChange:l=>t(Array.isArray(l)?l:[]),label:"Select access levels",placeholder:"Choose one or more...",helperText:"You can select multiple options"}),e.length>0&&a.jsxs("div",{className:"mt-4 p-3 bg-blue-50 border border-blue-200 rounded",children:[a.jsx("p",{className:"text-sm font-medium text-blue-900 mb-2",children:"Selected:"}),a.jsx("ul",{className:"text-sm text-blue-800 space-y-1",children:e.map((l,s)=>{const r=he.find(u=>u.value===l);return a.jsxs("li",{children:["• ",r==null?void 0:r.label]},`${l}-${s}`)})})]})]})},args:{options:he,value:[],onChange:()=>{}}},Z={render:()=>{const[e,t]=i.useState(null);return a.jsx(v,{options:we,value:e,onChange:N(t),label:"Select a country",placeholder:"Search or select...",helperText:"Type to filter the list",searchable:!0,className:"w-96"})},args:{options:we,value:null,onChange:()=>{}}},ee={render:()=>{const[e,t]=i.useState(null),l=[{value:"us",label:"United States"},{value:"ca",label:"Canada"},{value:"mx",label:"Mexico"}];return a.jsx(v,{options:l,value:e,onChange:N(t),label:"Select a country",placeholder:"Choose...",renderOption:(s,r)=>a.jsxs("div",{className:"flex items-center justify-between w-full",children:[a.jsxs("div",{className:"flex flex-col",children:[a.jsx("span",{className:"font-medium",children:s.label}),a.jsx("span",{className:"text-xs text-gray-500",children:String(s.value).toUpperCase()})]}),r&&a.jsx("span",{className:"text-trust-deep font-bold",children:"✓"})]}),className:"w-96"})},args:{options:[{value:"us",label:"United States"},{value:"ca",label:"Canada"},{value:"mx",label:"Mexico"}],value:null,onChange:()=>{}}},te={render:()=>{const[e,t]=i.useState(null),l=[{label:"North America",options:[{value:"us",label:"United States"},{value:"ca",label:"Canada"},{value:"mx",label:"Mexico"}]},{label:"Europe",options:[{value:"uk",label:"United Kingdom"},{value:"de",label:"Germany"},{value:"fr",label:"France"}]},{label:"Asia Pacific",options:[{value:"au",label:"Australia"},{value:"jp",label:"Japan"},{value:"sg",label:"Singapore"}]}];return a.jsx(v,{options:l,value:e,onChange:N(t),label:"Select a region",placeholder:"Choose a country...",className:"w-96"})},args:{options:[{value:"us",label:"United States"},{value:"ca",label:"Canada"}],value:null,onChange:()=>{}}},ae={render:()=>{const[e,t]=i.useState(null),l=[{value:"read",label:"Read - View documents"},{value:"comment",label:"Comment - Add comments"},{value:"edit",label:"Edit - Modify documents"}];return a.jsxs("div",{className:"space-y-8 w-96",children:[a.jsx(v,{options:l,value:e,onChange:N(t),label:"Permission Level",placeholder:"Select permission...",helperText:"Higher permissions allow more actions on shared documents",required:!0}),a.jsx(v,{options:l,value:e,onChange:N(t),label:"Grant Duration",placeholder:"Select duration...",errorMessage:"You must select a duration"}),a.jsx(v,{options:l,value:e,onChange:N(t),label:"Access Level",placeholder:"Select level...",successMessage:"Access level configured"})]})},args:{options:[{value:"read",label:"Read"},{value:"edit",label:"Edit"}],value:null,onChange:()=>{}}},le={render:()=>{const[e,t]=i.useState("7days"),[l,s]=i.useState("viewer"),r=[{value:"24hours",label:"24 hours - Limited time access"},{value:"7days",label:"7 days - One week duration"},{value:"30days",label:"30 days - One month access"},{value:"unlimited",label:"Unlimited - No expiration"}],u=[{value:"viewer",label:"Viewer"},{value:"commenter",label:"Commenter"},{value:"editor",label:"Editor"},{value:"admin",label:"Administrator"}];return a.jsx("div",{className:"space-y-8 w-96 p-4 bg-white border border-gray-200 rounded-lg",children:a.jsxs("div",{children:[a.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-4",children:"Grant Temporary Access"}),a.jsxs("div",{className:"space-y-6",children:[a.jsx(v,{options:r,value:e,onChange:N(t),label:"Access Duration",placeholder:"Choose duration...",helperText:"How long should access be granted?"}),a.jsx(v,{options:u,value:l,onChange:N(s),label:"Permission Level",placeholder:"Choose level...",helperText:"What can they do with this access?"}),a.jsx("button",{className:"w-full px-4 py-2 bg-trust-deep text-white rounded-lg font-medium hover:bg-trust-hover transition-colors",children:"Grant Access"})]})]})})},args:{options:[{value:"7days",label:"7 days - One week duration"},{value:"30days",label:"30 days - One month access"}],value:"7days",onChange:()=>{}}},se={render:e=>{const[t,l]=i.useState(null);return a.jsx(v,{...e,value:t,onChange:l,className:"w-96"})},args:{options:A,value:null,onChange:()=>{},label:"Select an option",placeholder:"Choose...",size:"md",disabled:!1,searchable:!1,multiselect:!1,required:!1}},ne={render:()=>{const[e,t]=i.useState(null),l=[{value:"viewer",label:"Viewer"},{value:"editor",label:"Editor"},{value:"admin",label:"Administrator"}];return a.jsxs("fieldset",{className:"border border-gray-200 rounded-lg p-6 w-96",children:[a.jsx("legend",{className:"text-lg font-semibold text-gray-900 mb-4",children:"User Role Assignment"}),a.jsx(v,{id:"user-role",options:l,value:e,onChange:N(t),label:"Assign role",placeholder:"Select a role...",helperText:"Select the role that best describes this user's responsibilities",required:!0}),a.jsxs("div",{className:"mt-6 p-3 bg-gray-50 rounded border border-gray-200",children:[a.jsx("p",{className:"text-sm font-medium text-gray-900",children:"Accessibility notes:"}),a.jsxs("ul",{className:"text-xs text-gray-700 space-y-1 mt-2",children:[a.jsx("li",{children:"• Use Tab to navigate to the select"}),a.jsx("li",{children:"• Press Space or Enter to open options"}),a.jsx("li",{children:"• Use arrow keys to navigate options"}),a.jsx("li",{children:"• Press Enter to select option"}),a.jsx("li",{children:"• Press Escape to close dropdown"})]})]})]})},parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0},{id:"label",enabled:!0}]}}},args:{options:[{value:"viewer",label:"Viewer"},{value:"editor",label:"Editor"}],value:null,onChange:()=>{}}},re={render:()=>{const[e,t]=i.useState(null);return a.jsxs("div",{className:"space-y-4 w-96",children:[a.jsx(v,{options:A,value:e,onChange:N(t),label:"Simulated Loading",placeholder:"Options loaded...",helperText:"In a real app, these would be fetched from an API"}),a.jsxs("div",{className:"p-3 bg-blue-50 border border-blue-200 rounded text-sm text-blue-800",children:[a.jsx("strong",{children:"Note:"})," Loading states are typically handled by your application layer. You can set ",a.jsx("code",{children:"disabled"})," and show a loading state externally while fetching options."]})]})},args:{options:A,value:null,onChange:()=>{}}};var Ne,Re,je;K.parameters={...K.parameters,docs:{...(Ne=K.parameters)==null?void 0:Ne.docs,source:{originalSource:`{
  render: () => {
    const [selected, setSelected] = useState<string | number | null>(null);
    return <Select options={basicOptions} value={selected} onChange={handleSingleChange(setSelected)} label="Choose an option" placeholder="Select one..." helperText="This is a basic select dropdown" />;
  },
  args: {
    options: basicOptions,
    value: null,
    onChange: () => {}
  }
}`,...(je=(Re=K.parameters)==null?void 0:Re.docs)==null?void 0:je.source}}};var Te,Le,Ae;J.parameters={...J.parameters,docs:{...(Te=J.parameters)==null?void 0:Te.docs,source:{originalSource:`{
  render: () => {
    const [small, setSmall] = useState<string | number | null>(null);
    const [medium, setMedium] = useState<string | number | null>(null);
    const [large, setLarge] = useState<string | number | null>(null);
    return <div className="space-y-6 w-96">
        <Select size="sm" options={basicOptions} value={small} onChange={handleSingleChange(setSmall)} label="Small Select" placeholder="Select..." />
        <Select size="md" options={basicOptions} value={medium} onChange={handleSingleChange(setMedium)} label="Medium Select (default)" placeholder="Select..." />
        <Select size="lg" options={basicOptions} value={large} onChange={handleSingleChange(setLarge)} label="Large Select" placeholder="Select..." />
      </div>;
  },
  args: {
    options: basicOptions,
    value: null,
    onChange: () => {}
  }
}`,...(Ae=(Le=J.parameters)==null?void 0:Le.docs)==null?void 0:Ae.source}}};var Pe,Ie,Ee;_.parameters={..._.parameters,docs:{...(Pe=_.parameters)==null?void 0:Pe.docs,source:{originalSource:`{
  render: () => {
    const [normal, setNormal] = useState<string | number | null>(null);
    const [error, setError] = useState<string | number | null>(null);
    const [success, setSuccess] = useState<string | number | null>('option2');
    const [disabled] = useState<string | number | null>(null);
    return <div className="space-y-6 w-96">
        <Select options={basicOptions} value={normal} onChange={handleSingleChange(setNormal)} label="Normal State" placeholder="Select an option..." />
        <Select options={basicOptions} value={error} onChange={handleSingleChange(setError)} label="Error State" placeholder="Select an option..." errorMessage="Please select a valid option" />
        <Select options={basicOptions} value={success} onChange={handleSingleChange(setSuccess)} label="Success State" placeholder="Select an option..." successMessage="Option selected successfully" />
        <Select options={basicOptions} value={disabled} onChange={() => {}} label="Disabled Select" placeholder="Select an option..." disabled helperText="This select is disabled" />
      </div>;
  },
  args: {
    options: basicOptions,
    value: null,
    onChange: () => {}
  }
}`,...(Ee=(Ie=_.parameters)==null?void 0:Ie.docs)==null?void 0:Ee.source}}};var ke,$e,Me;X.parameters={...X.parameters,docs:{...(ke=X.parameters)==null?void 0:ke.docs,source:{originalSource:`{
  render: () => {
    const [selected, setSelected] = useState<(string | number)[]>([]);
    return <div className="w-96">
        <Select multiselect options={accessLevelOptions} value={selected} onChange={val => setSelected(Array.isArray(val) ? val : [])} label="Select access levels" placeholder="Choose one or more..." helperText="You can select multiple options" />
        {selected.length > 0 && <div className="mt-4 p-3 bg-blue-50 border border-blue-200 rounded">
            <p className="text-sm font-medium text-blue-900 mb-2">Selected:</p>
            <ul className="text-sm text-blue-800 space-y-1">
              {selected.map((val, idx) => {
            const opt = accessLevelOptions.find(o => o.value === val);
            return <li key={\`\${val}-\${idx}\`}>• {opt?.label}</li>;
          })}
            </ul>
          </div>}
      </div>;
  },
  args: {
    options: accessLevelOptions,
    value: [],
    onChange: () => {}
  }
}`,...(Me=($e=X.parameters)==null?void 0:$e.docs)==null?void 0:Me.source}}};var De,Ue,qe;Z.parameters={...Z.parameters,docs:{...(De=Z.parameters)==null?void 0:De.docs,source:{originalSource:`{
  render: () => {
    const [selected, setSelected] = useState<string | number | null>(null);
    return <Select options={countryOptions} value={selected} onChange={handleSingleChange(setSelected)} label="Select a country" placeholder="Search or select..." helperText="Type to filter the list" searchable className="w-96" />;
  },
  args: {
    options: countryOptions,
    value: null,
    onChange: () => {}
  }
}`,...(qe=(Ue=Z.parameters)==null?void 0:Ue.docs)==null?void 0:qe.source}}};var ze,Ve,Fe;ee.parameters={...ee.parameters,docs:{...(ze=ee.parameters)==null?void 0:ze.docs,source:{originalSource:`{
  render: () => {
    const [selected, setSelected] = useState<string | number | null>(null);
    const customOptions: SelectOption[] = [{
      value: 'us',
      label: 'United States'
    }, {
      value: 'ca',
      label: 'Canada'
    }, {
      value: 'mx',
      label: 'Mexico'
    }];
    return <Select options={customOptions} value={selected} onChange={handleSingleChange(setSelected)} label="Select a country" placeholder="Choose..." renderOption={(option, isSelected) => <div className="flex items-center justify-between w-full">
            <div className="flex flex-col">
              <span className="font-medium">{option.label}</span>
              <span className="text-xs text-gray-500">
                {String(option.value).toUpperCase()}
              </span>
            </div>
            {isSelected && <span className="text-trust-deep font-bold">✓</span>}
          </div>} className="w-96" />;
  },
  args: {
    options: [{
      value: 'us',
      label: 'United States'
    }, {
      value: 'ca',
      label: 'Canada'
    }, {
      value: 'mx',
      label: 'Mexico'
    }],
    value: null,
    onChange: () => {}
  }
}`,...(Fe=(Ve=ee.parameters)==null?void 0:Ve.docs)==null?void 0:Fe.source}}};var Ge,He,Be;te.parameters={...te.parameters,docs:{...(Ge=te.parameters)==null?void 0:Ge.docs,source:{originalSource:`{
  render: () => {
    const [selected, setSelected] = useState<string | number | null>(null);
    const groupedOptions = [{
      label: 'North America',
      options: [{
        value: 'us',
        label: 'United States'
      }, {
        value: 'ca',
        label: 'Canada'
      }, {
        value: 'mx',
        label: 'Mexico'
      }]
    }, {
      label: 'Europe',
      options: [{
        value: 'uk',
        label: 'United Kingdom'
      }, {
        value: 'de',
        label: 'Germany'
      }, {
        value: 'fr',
        label: 'France'
      }]
    }, {
      label: 'Asia Pacific',
      options: [{
        value: 'au',
        label: 'Australia'
      }, {
        value: 'jp',
        label: 'Japan'
      }, {
        value: 'sg',
        label: 'Singapore'
      }]
    }];
    return <Select options={groupedOptions} value={selected} onChange={handleSingleChange(setSelected)} label="Select a region" placeholder="Choose a country..." className="w-96" />;
  },
  args: {
    options: [{
      value: 'us',
      label: 'United States'
    }, {
      value: 'ca',
      label: 'Canada'
    }],
    value: null,
    onChange: () => {}
  }
}`,...(Be=(He=te.parameters)==null?void 0:He.docs)==null?void 0:Be.source}}};var We,Qe,Ye;ae.parameters={...ae.parameters,docs:{...(We=ae.parameters)==null?void 0:We.docs,source:{originalSource:`{
  render: () => {
    const [selected, setSelected] = useState<string | number | null>(null);
    const permissionOptions: SelectOption[] = [{
      value: 'read',
      label: 'Read - View documents'
    }, {
      value: 'comment',
      label: 'Comment - Add comments'
    }, {
      value: 'edit',
      label: 'Edit - Modify documents'
    }];
    return <div className="space-y-8 w-96">
        <Select options={permissionOptions} value={selected} onChange={handleSingleChange(setSelected)} label="Permission Level" placeholder="Select permission..." helperText="Higher permissions allow more actions on shared documents" required />
        <Select options={permissionOptions} value={selected} onChange={handleSingleChange(setSelected)} label="Grant Duration" placeholder="Select duration..." errorMessage="You must select a duration" />
        <Select options={permissionOptions} value={selected} onChange={handleSingleChange(setSelected)} label="Access Level" placeholder="Select level..." successMessage="Access level configured" />
      </div>;
  },
  args: {
    options: [{
      value: 'read',
      label: 'Read'
    }, {
      value: 'edit',
      label: 'Edit'
    }],
    value: null,
    onChange: () => {}
  }
}`,...(Ye=(Qe=ae.parameters)==null?void 0:Qe.docs)==null?void 0:Ye.source}}};var Ke,Je,_e;le.parameters={...le.parameters,docs:{...(Ke=le.parameters)==null?void 0:Ke.docs,source:{originalSource:`{
  render: () => {
    const [duration, setDuration] = useState<string | number | null>('7days');
    const [permission, setPermission] = useState<string | number | null>('viewer');
    const durationOptions: SelectOption[] = [{
      value: '24hours',
      label: '24 hours - Limited time access'
    }, {
      value: '7days',
      label: '7 days - One week duration'
    }, {
      value: '30days',
      label: '30 days - One month access'
    }, {
      value: 'unlimited',
      label: 'Unlimited - No expiration'
    }];
    const permissionOptions: SelectOption[] = [{
      value: 'viewer',
      label: 'Viewer'
    }, {
      value: 'commenter',
      label: 'Commenter'
    }, {
      value: 'editor',
      label: 'Editor'
    }, {
      value: 'admin',
      label: 'Administrator'
    }];
    return <div className="space-y-8 w-96 p-4 bg-white border border-gray-200 rounded-lg">
        <div>
          <h3 className="text-lg font-semibold text-gray-900 mb-4">
            Grant Temporary Access
          </h3>

          <div className="space-y-6">
            <Select options={durationOptions} value={duration} onChange={handleSingleChange(setDuration)} label="Access Duration" placeholder="Choose duration..." helperText="How long should access be granted?" />

            <Select options={permissionOptions} value={permission} onChange={handleSingleChange(setPermission)} label="Permission Level" placeholder="Choose level..." helperText="What can they do with this access?" />

            <button className="w-full px-4 py-2 bg-trust-deep text-white rounded-lg font-medium hover:bg-trust-hover transition-colors">
              Grant Access
            </button>
          </div>
        </div>
      </div>;
  },
  args: {
    options: [{
      value: '7days',
      label: '7 days - One week duration'
    }, {
      value: '30days',
      label: '30 days - One month access'
    }],
    value: '7days',
    onChange: () => {}
  }
}`,...(_e=(Je=le.parameters)==null?void 0:Je.docs)==null?void 0:_e.source}}};var Xe,Ze,et;se.parameters={...se.parameters,docs:{...(Xe=se.parameters)==null?void 0:Xe.docs,source:{originalSource:`{
  render: (args: any) => {
    const [selected, setSelected] = useState<string | number | (string | number)[] | null>(null);
    return <Select {...args} value={selected} onChange={setSelected} className="w-96" />;
  },
  args: {
    options: basicOptions,
    value: null,
    onChange: () => {},
    label: 'Select an option',
    placeholder: 'Choose...',
    size: 'md',
    disabled: false,
    searchable: false,
    multiselect: false,
    required: false
  }
} satisfies Story`,...(et=(Ze=se.parameters)==null?void 0:Ze.docs)==null?void 0:et.source}}};var tt,at,lt;ne.parameters={...ne.parameters,docs:{...(tt=ne.parameters)==null?void 0:tt.docs,source:{originalSource:`{
  render: () => {
    const [selected, setSelected] = useState<string | number | null>(null);
    const roleOptions: SelectOption[] = [{
      value: 'viewer',
      label: 'Viewer'
    }, {
      value: 'editor',
      label: 'Editor'
    }, {
      value: 'admin',
      label: 'Administrator'
    }];
    return <fieldset className="border border-gray-200 rounded-lg p-6 w-96">
        <legend className="text-lg font-semibold text-gray-900 mb-4">
          User Role Assignment
        </legend>

        <Select id="user-role" options={roleOptions} value={selected} onChange={handleSingleChange(setSelected)} label="Assign role" placeholder="Select a role..." helperText="Select the role that best describes this user's responsibilities" required />

        <div className="mt-6 p-3 bg-gray-50 rounded border border-gray-200">
          <p className="text-sm font-medium text-gray-900">Accessibility notes:</p>
          <ul className="text-xs text-gray-700 space-y-1 mt-2">
            <li>• Use Tab to navigate to the select</li>
            <li>• Press Space or Enter to open options</li>
            <li>• Use arrow keys to navigate options</li>
            <li>• Press Enter to select option</li>
            <li>• Press Escape to close dropdown</li>
          </ul>
        </div>
      </fieldset>;
  },
  parameters: {
    a11y: {
      config: {
        rules: [{
          id: 'color-contrast',
          enabled: true
        }, {
          id: 'label',
          enabled: true
        }]
      }
    }
  },
  args: {
    options: [{
      value: 'viewer',
      label: 'Viewer'
    }, {
      value: 'editor',
      label: 'Editor'
    }],
    value: null,
    onChange: () => {}
  }
}`,...(lt=(at=ne.parameters)==null?void 0:at.docs)==null?void 0:lt.source}}};var st,nt,rt;re.parameters={...re.parameters,docs:{...(st=re.parameters)==null?void 0:st.docs,source:{originalSource:`{
  render: () => {
    const [selected, setSelected] = useState<string | number | null>(null);
    return <div className="space-y-4 w-96">
        <Select options={basicOptions} value={selected} onChange={handleSingleChange(setSelected)} label="Simulated Loading" placeholder="Options loaded..." helperText="In a real app, these would be fetched from an API" />
        <div className="p-3 bg-blue-50 border border-blue-200 rounded text-sm text-blue-800">
          <strong>Note:</strong> Loading states are typically handled by your application layer.
          You can set <code>disabled</code> and show a loading state externally while fetching options.
        </div>
      </div>;
  },
  args: {
    options: basicOptions,
    value: null,
    onChange: () => {}
  }
}`,...(rt=(nt=re.parameters)==null?void 0:nt.docs)==null?void 0:rt.source}}};const fa=["Default","Sizes","States","MultiSelect","Searchable","CustomRendering","GroupedOptions","WithHelperText","RealWorldUseCases","Playground","Accessibility","LoadingState"];export{ne as Accessibility,ee as CustomRendering,K as Default,te as GroupedOptions,re as LoadingState,X as MultiSelect,se as Playground,le as RealWorldUseCases,Z as Searchable,J as Sizes,_ as States,ae as WithHelperText,fa as __namedExportsOrder,ha as default};
