import{j as e}from"./jsx-runtime-BYYWji4R.js";import{r as l,R as N}from"./index-ClcD9ViR.js";import{c as Q,a as ye}from"./cn-JCLedEej.js";import{y as Rt,n as Tt,s as z,e as $t,N as Dt,E as Mt}from"./use-root-containers-Dht_jFDM.js";import{U as ae,y as _,o as C,C as le,I as L,b as Wt,l as zt,c as U,u as q,O as me,T as Ft,d as Ce}from"./keyboard-DNQdDreo.js";import{u as Et,d as K,y as Ut,s as Gt,q as qt}from"./transition-CF-Dw_VV.js";import{n as be}from"./use-owner-CcFJWZQm.js";import{T as Lt}from"./use-resolve-button-type-Bw633SU1.js";import{u as ge,s as fe}from"./hidden-BnOBa8gi.js";import{r as It}from"./bugs-Ot56VXf3.js";import{O as G,M as F,N as xe,f as we,o as _t,h as Kt,T as Vt}from"./use-is-mounted-62Q7yJH9.js";import{B as n}from"./Button-DlbfybpJ.js";import"./_commonjsHelpers-Cpj98o6Y.js";import"./index-BUAr5TKG.js";var Ht=(t=>(t[t.Open=0]="Open",t[t.Closed=1]="Closed",t))(Ht||{}),Qt=(t=>(t[t.TogglePopover=0]="TogglePopover",t[t.ClosePopover=1]="ClosePopover",t[t.SetButton=2]="SetButton",t[t.SetButtonId=3]="SetButtonId",t[t.SetPanel=4]="SetPanel",t[t.SetPanelId=5]="SetPanelId",t))(Qt||{});let Yt={0:t=>{let o={...t,popoverState:q(t.popoverState,{0:1,1:0})};return o.popoverState===0&&(o.__demoMode=!1),o},1(t){return t.popoverState===1?t:{...t,popoverState:1}},2(t,o){return t.button===o.button?t:{...t,button:o.button}},3(t,o){return t.buttonId===o.buttonId?t:{...t,buttonId:o.buttonId}},4(t,o){return t.panel===o.panel?t:{...t,panel:o.panel}},5(t,o){return t.panelId===o.panelId?t:{...t,panelId:o.panelId}}},je=l.createContext(null);je.displayName="PopoverContext";function he(t){let o=l.useContext(je);if(o===null){let u=new Error(`<${t} /> is missing a parent <Popover /> component.`);throw Error.captureStackTrace&&Error.captureStackTrace(u,he),u}return o}let Pe=l.createContext(null);Pe.displayName="PopoverAPIContext";function Ne(t){let o=l.useContext(Pe);if(o===null){let u=new Error(`<${t} /> is missing a parent <Popover /> component.`);throw Error.captureStackTrace&&Error.captureStackTrace(u,Ne),u}return o}let Ae=l.createContext(null);Ae.displayName="PopoverGroupContext";function kt(){return l.useContext(Ae)}let ve=l.createContext(null);ve.displayName="PopoverPanelContext";function Jt(){return l.useContext(ve)}function Xt(t,o){return q(o.type,Yt,t,o)}let Zt="div";function er(t,o){var u;let{__demoMode:w=!1,...A}=t,a=l.useRef(null),r=_(o,Ft(c=>{a.current=c})),g=l.useRef([]),f=l.useReducer(Xt,{__demoMode:w,popoverState:w?0:1,buttons:g,button:null,buttonId:null,panel:null,panelId:null,beforePanelSentinel:l.createRef(),afterPanelSentinel:l.createRef()}),[{popoverState:j,button:p,buttonId:P,panel:s,panelId:O,beforePanelSentinel:S,afterPanelSentinel:y},d]=f,x=be((u=a.current)!=null?u:p),I=l.useMemo(()=>{if(!p||!s)return!1;for(let pe of document.querySelectorAll("body > *"))if(Number(pe==null?void 0:pe.contains(p))^Number(pe==null?void 0:pe.contains(s)))return!0;let c=we(),k=c.indexOf(p),ce=(k+c.length-1)%c.length,H=(k+1)%c.length,de=c[ce],Ot=c[H];return!s.contains(de)&&!s.contains(Ot)},[p,s]),R=Ce(P),D=Ce(O),M=l.useMemo(()=>({buttonId:R,panelId:D,close:()=>d({type:1})}),[R,D,d]),$=kt(),W=$==null?void 0:$.registerPopover,h=C(()=>{var c;return(c=$==null?void 0:$.isFocusWithinPopoverGroup())!=null?c:(x==null?void 0:x.activeElement)&&((p==null?void 0:p.contains(x.activeElement))||(s==null?void 0:s.contains(x.activeElement)))});l.useEffect(()=>W==null?void 0:W(M),[W,M]);let[T,b]=$t(),i=Dt({mainTreeNodeRef:$==null?void 0:$.mainTreeNodeRef,portals:T,defaultContainers:[p,s]});Mt(x==null?void 0:x.defaultView,"focus",c=>{var k,ce,H,de;c.target!==window&&c.target instanceof HTMLElement&&j===0&&(h()||p&&s&&(i.contains(c.target)||(ce=(k=S.current)==null?void 0:k.contains)!=null&&ce.call(k,c.target)||(de=(H=y.current)==null?void 0:H.contains)!=null&&de.call(H,c.target)||d({type:1})))},!0),Ut(i.resolveContainers,(c,k)=>{d({type:1}),Kt(k,Vt.Loose)||(c.preventDefault(),p==null||p.focus())},j===0);let v=C(c=>{d({type:1});let k=c?c instanceof HTMLElement?c:"current"in c&&c.current instanceof HTMLElement?c.current:p:p;k==null||k.focus()}),E=l.useMemo(()=>({close:v,isPortalled:I}),[v,I]),B=l.useMemo(()=>({open:j===0,close:v}),[j,v]),V={ref:r};return N.createElement(ve.Provider,{value:null},N.createElement(je.Provider,{value:f},N.createElement(Pe.Provider,{value:E},N.createElement(Gt,{value:q(j,{0:K.Open,1:K.Closed})},N.createElement(b,null,le({ourProps:V,theirProps:A,slot:B,defaultTag:Zt,name:"Popover"}),N.createElement(i.MainTreeNode,null))))))}let tr="button";function rr(t,o){let u=L(),{id:w=`headlessui-popover-button-${u}`,...A}=t,[a,r]=he("Popover.Button"),{isPortalled:g}=Ne("Popover.Button"),f=l.useRef(null),j=`headlessui-focus-sentinel-${L()}`,p=kt(),P=p==null?void 0:p.closeOthers,s=Jt()!==null;l.useEffect(()=>{if(!s)return r({type:3,buttonId:w}),()=>{r({type:3,buttonId:null})}},[s,w,r]);let[O]=l.useState(()=>Symbol()),S=_(f,o,s?null:i=>{if(i)a.buttons.current.push(O);else{let v=a.buttons.current.indexOf(O);v!==-1&&a.buttons.current.splice(v,1)}a.buttons.current.length>1&&console.warn("You are already using a <Popover.Button /> but only 1 <Popover.Button /> is supported."),i&&r({type:2,button:i})}),y=_(f,o),d=be(f),x=C(i=>{var v,E,B;if(s){if(a.popoverState===1)return;switch(i.key){case U.Space:case U.Enter:i.preventDefault(),(E=(v=i.target).click)==null||E.call(v),r({type:1}),(B=a.button)==null||B.focus();break}}else switch(i.key){case U.Space:case U.Enter:i.preventDefault(),i.stopPropagation(),a.popoverState===1&&(P==null||P(a.buttonId)),r({type:0});break;case U.Escape:if(a.popoverState!==0)return P==null?void 0:P(a.buttonId);if(!f.current||d!=null&&d.activeElement&&!f.current.contains(d.activeElement))return;i.preventDefault(),i.stopPropagation(),r({type:1});break}}),I=C(i=>{s||i.key===U.Space&&i.preventDefault()}),R=C(i=>{var v,E;It(i.currentTarget)||t.disabled||(s?(r({type:1}),(v=a.button)==null||v.focus()):(i.preventDefault(),i.stopPropagation(),a.popoverState===1&&(P==null||P(a.buttonId)),r({type:0}),(E=a.button)==null||E.focus()))}),D=C(i=>{i.preventDefault(),i.stopPropagation()}),M=a.popoverState===0,$=l.useMemo(()=>({open:M}),[M]),W=Lt(t,f),h=s?{ref:y,type:W,onKeyDown:x,onClick:R}:{ref:S,id:a.buttonId,type:W,"aria-expanded":a.popoverState===0,"aria-controls":a.panel?a.panelId:void 0,onKeyDown:x,onKeyUp:I,onClick:R,onMouseDown:D},T=Tt(),b=C(()=>{let i=a.panel;if(!i)return;function v(){q(T.current,{[z.Forwards]:()=>G(i,F.First),[z.Backwards]:()=>G(i,F.Last)})===xe.Error&&G(we().filter(E=>E.dataset.headlessuiFocusGuard!=="true"),q(T.current,{[z.Forwards]:F.Next,[z.Backwards]:F.Previous}),{relativeTo:a.button})}v()});return N.createElement(N.Fragment,null,le({ourProps:h,theirProps:A,slot:$,defaultTag:tr,name:"Popover.Button"}),M&&!s&&g&&N.createElement(ge,{id:j,features:fe.Focusable,"data-headlessui-focus-guard":!0,as:"button",type:"button",onFocus:b}))}let or="div",nr=me.RenderStrategy|me.Static;function sr(t,o){let u=L(),{id:w=`headlessui-popover-overlay-${u}`,...A}=t,[{popoverState:a},r]=he("Popover.Overlay"),g=_(o),f=Et(),j=f!==null?(f&K.Open)===K.Open:a===0,p=C(s=>{if(It(s.currentTarget))return s.preventDefault();r({type:1})}),P=l.useMemo(()=>({open:a===0}),[a]);return le({ourProps:{ref:g,id:w,"aria-hidden":!0,onClick:p},theirProps:A,slot:P,defaultTag:or,features:nr,visible:j,name:"Popover.Overlay"})}let ir="div",ar=me.RenderStrategy|me.Static;function lr(t,o){let u=L(),{id:w=`headlessui-popover-panel-${u}`,focus:A=!1,...a}=t,[r,g]=he("Popover.Panel"),{close:f,isPortalled:j}=Ne("Popover.Panel"),p=`headlessui-focus-sentinel-before-${L()}`,P=`headlessui-focus-sentinel-after-${L()}`,s=l.useRef(null),O=_(s,o,h=>{g({type:4,panel:h})}),S=be(s),y=Wt();zt(()=>(g({type:5,panelId:w}),()=>{g({type:5,panelId:null})}),[w,g]);let d=Et(),x=d!==null?(d&K.Open)===K.Open:r.popoverState===0,I=C(h=>{var T;switch(h.key){case U.Escape:if(r.popoverState!==0||!s.current||S!=null&&S.activeElement&&!s.current.contains(S.activeElement))return;h.preventDefault(),h.stopPropagation(),g({type:1}),(T=r.button)==null||T.focus();break}});l.useEffect(()=>{var h;t.static||r.popoverState===1&&((h=t.unmount)==null||h)&&g({type:4,panel:null})},[r.popoverState,t.unmount,t.static,g]),l.useEffect(()=>{if(r.__demoMode||!A||r.popoverState!==0||!s.current)return;let h=S==null?void 0:S.activeElement;s.current.contains(h)||G(s.current,F.First)},[r.__demoMode,A,s,r.popoverState]);let R=l.useMemo(()=>({open:r.popoverState===0,close:f}),[r,f]),D={ref:O,id:w,onKeyDown:I,onBlur:A&&r.popoverState===0?h=>{var T,b,i,v,E;let B=h.relatedTarget;B&&s.current&&((T=s.current)!=null&&T.contains(B)||(g({type:1}),((i=(b=r.beforePanelSentinel.current)==null?void 0:b.contains)!=null&&i.call(b,B)||(E=(v=r.afterPanelSentinel.current)==null?void 0:v.contains)!=null&&E.call(v,B))&&B.focus({preventScroll:!0})))}:void 0,tabIndex:-1},M=Tt(),$=C(()=>{let h=s.current;if(!h)return;function T(){q(M.current,{[z.Forwards]:()=>{var b;G(h,F.First)===xe.Error&&((b=r.afterPanelSentinel.current)==null||b.focus())},[z.Backwards]:()=>{var b;(b=r.button)==null||b.focus({preventScroll:!0})}})}T()}),W=C(()=>{let h=s.current;if(!h)return;function T(){q(M.current,{[z.Forwards]:()=>{var b;if(!r.button)return;let i=we(),v=i.indexOf(r.button),E=i.slice(0,v+1),B=[...i.slice(v+1),...E];for(let V of B.slice())if(V.dataset.headlessuiFocusGuard==="true"||(b=r.panel)!=null&&b.contains(V)){let c=B.indexOf(V);c!==-1&&B.splice(c,1)}G(B,F.First,{sorted:!1})},[z.Backwards]:()=>{var b;G(h,F.Previous)===xe.Error&&((b=r.button)==null||b.focus())}})}T()});return N.createElement(ve.Provider,{value:w},x&&j&&N.createElement(ge,{id:p,ref:r.beforePanelSentinel,features:fe.Focusable,"data-headlessui-focus-guard":!0,as:"button",type:"button",onFocus:$}),le({mergeRefs:y,ourProps:D,theirProps:a,slot:R,defaultTag:ir,features:ar,visible:x,name:"Popover.Panel"}),x&&j&&N.createElement(ge,{id:P,ref:r.afterPanelSentinel,features:fe.Focusable,"data-headlessui-focus-guard":!0,as:"button",type:"button",onFocus:W}))}let cr="div";function dr(t,o){let u=l.useRef(null),w=_(u,o),[A,a]=l.useState([]),r=Rt(),g=C(y=>{a(d=>{let x=d.indexOf(y);if(x!==-1){let I=d.slice();return I.splice(x,1),I}return d})}),f=C(y=>(a(d=>[...d,y]),()=>g(y))),j=C(()=>{var y;let d=_t(u);if(!d)return!1;let x=d.activeElement;return(y=u.current)!=null&&y.contains(x)?!0:A.some(I=>{var R,D;return((R=d.getElementById(I.buttonId.current))==null?void 0:R.contains(x))||((D=d.getElementById(I.panelId.current))==null?void 0:D.contains(x))})}),p=C(y=>{for(let d of A)d.buttonId.current!==y&&d.close()}),P=l.useMemo(()=>({registerPopover:f,unregisterPopover:g,isFocusWithinPopoverGroup:j,closeOthers:p,mainTreeNodeRef:r.mainTreeNodeRef}),[f,g,j,p,r.mainTreeNodeRef]),s=l.useMemo(()=>({}),[]),O=t,S={ref:w};return N.createElement(Ae.Provider,{value:P},le({ourProps:S,theirProps:O,slot:s,defaultTag:cr,name:"Popover.Group"}),N.createElement(r.MainTreeNode,null))}let pr=ae(er),ur=ae(rr),mr=ae(sr),hr=ae(lr),vr=ae(dr),ue=Object.assign(pr,{Button:ur,Overlay:mr,Panel:hr,Group:vr});const gr=ye("bg-white rounded-lg shadow-xl border border-gray-200 focus:outline-none",{variants:{width:{sm:"w-[300px]",md:"w-[400px]",lg:"w-[500px]",full:"w-[90vw]"}},defaultVariants:{width:"md"}}),fr=ye("absolute z-50",{variants:{position:{top:"bottom-full mb-2 left-1/2 -translate-x-1/2",right:"left-full ml-2 top-1/2 -translate-y-1/2",bottom:"top-full mt-2 left-1/2 -translate-x-1/2",left:"right-full mr-2 top-1/2 -translate-y-1/2"}},defaultVariants:{position:"bottom"}}),xr=ye("absolute w-3 h-3 bg-white border-gray-200 rotate-45",{variants:{position:{top:"bottom-[-6px] left-1/2 -translate-x-1/2 border-b border-r",right:"left-[-6px] top-1/2 -translate-y-1/2 border-l border-b",bottom:"top-[-6px] left-1/2 -translate-x-1/2 border-t border-l",left:"right-[-6px] top-1/2 -translate-y-1/2 border-t border-r"}},defaultVariants:{position:"bottom"}}),m=N.forwardRef(({trigger:t,children:o,header:u,footer:w,position:A="bottom",width:a="md",showArrow:r=!0,isOpen:g,onOpenChange:f,className:j,triggerClassName:p,...P},s)=>{const O=g!==void 0&&f!==void 0,S=e.jsxs(e.Fragment,{children:[e.jsx(ue.Button,{className:Q("inline-flex items-center focus:outline-none focus:ring-2 focus:ring-trust-deep focus:ring-offset-2 rounded-md",p),children:t}),e.jsx(qt,{as:l.Fragment,enter:"transition ease-out duration-200",enterFrom:"opacity-0 scale-95",enterTo:"opacity-100 scale-100",leave:"transition ease-in duration-150",leaveFrom:"opacity-100 scale-100",leaveTo:"opacity-0 scale-95",children:e.jsx(ue.Panel,{className:Q(fr({position:A})),ref:s,children:e.jsxs("div",{className:Q(gr({width:a}),j),...P,children:[r&&e.jsx("div",{className:Q(xr({position:A}))}),u&&e.jsx("div",{className:"px-5 py-4 border-b border-gray-200",children:typeof u=="string"?e.jsx("h3",{className:"text-base font-semibold text-gray-900",children:u}):u}),e.jsx("div",{className:Q("px-5 py-4 text-sm text-gray-700",!u&&"pt-5",!w&&"pb-5"),children:o}),w&&e.jsx("div",{className:"px-5 py-3 border-t border-gray-200 bg-gray-50 rounded-b-lg",children:w})]})})})]});return O?e.jsx(ue,{className:"relative inline-flex",children:({open:y})=>(N.useEffect(()=>{y!==g&&f(y)},[y]),S)}):e.jsx(ue,{className:"relative inline-flex",children:S})});m.displayName="Popover";m.__docgenInfo={description:`Popover component for displaying rich interactive content panels.
Uses Headless UI Popover for full accessibility and positioning.

@example
\`\`\`tsx
<Popover trigger={<Button>Open</Button>}>
  <div>Popover content here</div>
</Popover>

<Popover
  trigger={<IconButton icon={<InfoIcon />} />}
  header="User Information"
  position="right"
  showArrow
>
  <p>Detailed user information goes here.</p>
</Popover>

<Popover
  trigger={<Button>Actions</Button>}
  header="Quick Actions"
  footer={
    <div className="flex gap-2 justify-end">
      <Button size="sm">Cancel</Button>
      <Button size="sm" variant="primary">Apply</Button>
    </div>
  }
>
  <div>Select an action to perform</div>
</Popover>
\`\`\``,methods:[],displayName:"Popover",props:{trigger:{required:!0,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Element that triggers the popover"},children:{required:!0,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Popover content"},header:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Optional header section (title or custom content)"},footer:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Optional footer section (actions or custom content)"},position:{required:!1,tsType:{name:"union",raw:"'top' | 'right' | 'bottom' | 'left'",elements:[{name:"literal",value:"'top'"},{name:"literal",value:"'right'"},{name:"literal",value:"'bottom'"},{name:"literal",value:"'left'"}]},description:"Position of popover relative to trigger",defaultValue:{value:"'bottom'",computed:!1}},width:{required:!1,tsType:{name:"union",raw:"'sm' | 'md' | 'lg' | 'full'",elements:[{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"},{name:"literal",value:"'lg'"},{name:"literal",value:"'full'"}]},description:"Width variant",defaultValue:{value:"'md'",computed:!1}},showArrow:{required:!1,tsType:{name:"boolean"},description:"Whether to show arrow indicator",defaultValue:{value:"true",computed:!1}},isOpen:{required:!1,tsType:{name:"boolean"},description:"Controlled open state (optional)"},onOpenChange:{required:!1,tsType:{name:"signature",type:"function",raw:"(open: boolean) => void",signature:{arguments:[{type:{name:"boolean"},name:"open"}],return:{name:"void"}}},description:"Callback when open state changes (optional)"},className:{required:!1,tsType:{name:"string"},description:"Additional CSS classes for the panel"},triggerClassName:{required:!1,tsType:{name:"string"},description:"Additional CSS classes for the trigger wrapper"}}};const Be=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"})}),yr=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"})}),Rr={title:"Design System/Overlays/Popover",component:m,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{trigger:{control:!1,description:"Element that triggers the popover"},children:{control:"text",description:"Popover content"},header:{control:"text",description:"Optional header section (title or custom content)"},footer:{control:!1,description:"Optional footer section (actions or custom content)"},position:{control:"select",options:["top","right","bottom","left"],description:"Position of popover relative to trigger"},width:{control:"select",options:["sm","md","lg","full"],description:"Width variant"},showArrow:{control:"boolean",description:"Whether to show arrow indicator"},isOpen:{control:"boolean",description:"Controlled open state (optional)"}}},Y={render:t=>e.jsx("div",{className:"flex items-center justify-center min-h-[400px]",children:e.jsx(m,{...t,trigger:e.jsx(n,{variant:"outline",children:"Open Popover"}),children:e.jsx("p",{children:"This is a basic popover with simple content. Click outside or press ESC to close."})})}),args:{trigger:e.jsx(n,{variant:"outline",children:"Open Popover"}),children:"Default popover content",position:"bottom",width:"md",showArrow:!0}},J={render:()=>e.jsx("div",{className:"flex items-center justify-center min-h-[600px] gap-20",children:e.jsxs("div",{className:"grid grid-cols-3 gap-20 items-center",children:[e.jsx("div",{className:"col-start-2 flex justify-center",children:e.jsx(m,{trigger:e.jsx(n,{variant:"outline",children:"Top"}),position:"top",showArrow:!0,children:e.jsx("p",{children:"Popover positioned at the top"})})}),e.jsx("div",{className:"col-start-1 flex justify-center",children:e.jsx(m,{trigger:e.jsx(n,{variant:"outline",children:"Left"}),position:"left",showArrow:!0,children:e.jsx("p",{children:"Popover positioned on the left"})})}),e.jsx("div",{className:"col-start-2 flex justify-center",children:e.jsx("div",{className:"w-24 h-24 rounded-lg border-2 border-dashed border-gray-300 flex items-center justify-center text-sm text-gray-500",children:"Trigger"})}),e.jsx("div",{className:"col-start-3 flex justify-center",children:e.jsx(m,{trigger:e.jsx(n,{variant:"outline",children:"Right"}),position:"right",showArrow:!0,children:e.jsx("p",{children:"Popover positioned on the right"})})}),e.jsx("div",{className:"col-start-2 flex justify-center",children:e.jsx(m,{trigger:e.jsx(n,{variant:"outline",children:"Bottom"}),position:"bottom",showArrow:!0,children:e.jsx("p",{children:"Popover positioned at the bottom"})})})]})}),args:{trigger:e.jsx(n,{variant:"outline",children:"Position"}),children:"Popover content",position:"bottom",showArrow:!0}},X={render:()=>e.jsx("div",{className:"flex items-center justify-center min-h-[400px]",children:e.jsx(m,{trigger:e.jsx(n,{variant:"outline",iconBefore:e.jsx(Be,{}),children:"User Information"}),header:"Profile Details",position:"bottom",showArrow:!0,children:e.jsxs("div",{className:"space-y-3",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-xs font-medium text-gray-500 mb-1",children:"Name"}),e.jsx("p",{className:"font-medium text-gray-900",children:"John Doe"})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-xs font-medium text-gray-500 mb-1",children:"Email"}),e.jsx("p",{className:"text-gray-700",children:"john.doe@example.com"})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-xs font-medium text-gray-500 mb-1",children:"Role"}),e.jsx("p",{className:"text-gray-700",children:"Administrator"})]})]})})}),args:{trigger:e.jsx(n,{variant:"outline",children:"User Info"}),children:"User information content",header:"Profile Details",position:"bottom",showArrow:!0}},Z={render:()=>e.jsx("div",{className:"flex items-center justify-center min-h-[400px]",children:e.jsx(m,{trigger:e.jsx(n,{variant:"outline",children:"Quick Actions"}),position:"bottom",showArrow:!0,footer:e.jsxs("div",{className:"flex gap-2 justify-end",children:[e.jsx(n,{size:"sm",variant:"outline",children:"Cancel"}),e.jsx(n,{size:"sm",variant:"primary",children:"Apply"})]}),children:e.jsx("p",{children:"Select an action to perform. Use the buttons below to confirm or cancel."})})}),args:{trigger:e.jsx(n,{variant:"outline",children:"Quick Actions"}),children:"Action content",position:"bottom",showArrow:!0}},ee={render:()=>e.jsx("div",{className:"flex items-center justify-center min-h-[500px]",children:e.jsx(m,{trigger:e.jsx(n,{variant:"primary",children:"Confirm Consent"}),header:"Grant Access",position:"bottom",width:"md",showArrow:!0,footer:e.jsxs("div",{className:"flex gap-2 justify-end",children:[e.jsx(n,{size:"sm",variant:"outline",children:"Deny"}),e.jsx(n,{size:"sm",variant:"primary",children:"Grant Access"})]}),children:e.jsxs("div",{className:"space-y-3",children:[e.jsxs("p",{children:[e.jsx("strong",{children:"Analytics Dashboard"})," is requesting access to:"]}),e.jsxs("ul",{className:"list-disc list-inside space-y-1 text-sm pl-2",children:[e.jsx("li",{children:"Basic profile information"}),e.jsx("li",{children:"Usage statistics"}),e.jsx("li",{children:"Preference settings"})]}),e.jsx("div",{className:"mt-3 p-3 bg-info-light border border-info-primary/20 rounded-md",children:e.jsx("p",{className:"text-xs text-info-dark",children:"This permission will be valid for 30 days and can be revoked at any time."})})]})})}),args:{trigger:e.jsx(n,{variant:"primary",children:"Confirm Consent"}),children:"Consent content",header:"Grant Access",position:"bottom",width:"md",showArrow:!0}},te={render:()=>e.jsxs("div",{className:"flex items-center justify-center min-h-[400px] gap-8",children:[e.jsx(m,{trigger:e.jsx(n,{variant:"outline",children:"With Arrow"}),position:"bottom",showArrow:!0,children:e.jsx("p",{children:"This popover has an arrow indicator pointing to the trigger."})}),e.jsx(m,{trigger:e.jsx(n,{variant:"outline",children:"Without Arrow"}),position:"bottom",showArrow:!1,children:e.jsx("p",{children:"This popover has no arrow indicator."})})]}),args:{trigger:e.jsx(n,{variant:"outline",children:"Arrow Toggle"}),children:"Arrow content",position:"bottom",showArrow:!0}},re={render:()=>e.jsxs("div",{className:"flex items-center justify-center min-h-[500px] gap-6",children:[e.jsx(m,{trigger:e.jsx(n,{variant:"outline",children:"Small (300px)"}),header:"Small Width",position:"bottom",width:"sm",showArrow:!0,children:e.jsx("p",{children:"This is a small popover at 300px width."})}),e.jsx(m,{trigger:e.jsx(n,{variant:"outline",children:"Medium (400px)"}),header:"Medium Width",position:"bottom",width:"md",showArrow:!0,children:e.jsx("p",{children:"This is a medium popover at 400px width (default)."})}),e.jsx(m,{trigger:e.jsx(n,{variant:"outline",children:"Large (500px)"}),header:"Large Width",position:"bottom",width:"lg",showArrow:!0,children:e.jsxs("div",{className:"space-y-3",children:[e.jsx("p",{children:"This is a large popover at 500px width."}),e.jsx("p",{className:"text-sm text-gray-600",children:"Perfect for more detailed content or data tables."}),e.jsx("div",{className:"p-3 bg-gray-50 rounded-md border border-gray-200",children:e.jsx("p",{className:"text-xs font-medium text-gray-700",children:"Example content area with more space"})})]})})]}),args:{trigger:e.jsx(n,{variant:"outline",children:"Width Demo"}),children:"Width content",position:"bottom",width:"md",showArrow:!0}},oe={render:()=>e.jsxs("div",{className:"flex flex-col items-center justify-center min-h-[400px] gap-8",children:[e.jsxs("div",{className:"text-center max-w-md mb-4",children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-2",children:"Keyboard Navigation"}),e.jsx("p",{className:"text-sm text-gray-600",children:"Try using Tab to navigate between buttons, then press Enter or Space to open the popover. Press Escape to close."})]}),e.jsxs("div",{className:"flex gap-4",children:[e.jsx(m,{trigger:e.jsx("button",{className:"px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-trust-deep focus:ring-offset-2",tabIndex:0,children:e.jsxs("span",{className:"flex items-center gap-2",children:[e.jsx(Be,{}),"Focusable Button 1"]})}),header:"Accessible Popover",position:"bottom",showArrow:!0,children:e.jsx("p",{children:"This popover can be triggered with keyboard navigation. Press Tab to focus, Enter to open, and Escape to close."})}),e.jsx(m,{trigger:e.jsx("button",{className:"px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-trust-deep focus:ring-offset-2",tabIndex:0,children:e.jsxs("span",{className:"flex items-center gap-2",children:[e.jsx(yr,{}),"Focusable Button 2"]})}),header:"WCAG 2.1 AA Compliant",position:"bottom",showArrow:!0,children:e.jsx("p",{children:"All popovers support full keyboard accessibility and screen reader announcements."})})]})]}),args:{trigger:e.jsx("button",{children:"Keyboard Focus"}),children:"Keyboard focus content",position:"bottom",showArrow:!0}},ne={render:t=>e.jsx("div",{className:"flex items-center justify-center min-h-[600px]",children:e.jsx(m,{...t,trigger:e.jsx(n,{variant:"primary",children:"Open Playground"}),children:t.children||e.jsxs("div",{children:[e.jsx("p",{className:"mb-3",children:"Use the controls below to customize this popover's appearance and behavior."}),e.jsxs("ul",{className:"list-disc list-inside text-sm space-y-1",children:[e.jsx("li",{children:"Change position (top, right, bottom, left)"}),e.jsx("li",{children:"Adjust width (sm, md, lg, full)"}),e.jsx("li",{children:"Toggle arrow indicator"}),e.jsx("li",{children:"Add header or footer sections"})]})]})})}),args:{trigger:e.jsx(n,{variant:"primary",children:"Open Playground"}),children:"Playground content",position:"bottom",width:"md",showArrow:!0,header:"Playground Popover"}},se={render:()=>{const[t,o]=l.useState(!1);return e.jsxs("div",{className:"flex flex-col items-center justify-center min-h-[400px] gap-6",children:[e.jsxs("div",{className:"text-center",children:[e.jsx("p",{className:"text-sm text-gray-600 mb-3",children:"External controls (controlled mode):"}),e.jsxs("div",{className:"flex gap-2 justify-center",children:[e.jsx(n,{size:"sm",variant:"outline",onClick:()=>o(!0),children:"Open Popover"}),e.jsx(n,{size:"sm",variant:"outline",onClick:()=>o(!1),children:"Close Popover"})]}),e.jsxs("p",{className:"text-xs text-gray-500 mt-2",children:["Current state: ",t?"Open":"Closed"]})]}),e.jsx(m,{trigger:e.jsx(n,{variant:"primary",children:"Controlled Popover"}),header:"Controlled Mode",position:"bottom",showArrow:!0,isOpen:t,onOpenChange:o,children:e.jsx("p",{children:"This popover's state is controlled externally. You can open/close it using the buttons above or by clicking the trigger."})})]})},args:{trigger:e.jsx(n,{variant:"primary",children:"Controlled"}),children:"Controlled content",position:"bottom",showArrow:!0}},ie={render:()=>e.jsxs("div",{className:"flex flex-col items-center justify-center min-h-[500px] gap-6",children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900",children:"Consent Management Use Cases"}),e.jsxs("div",{className:"flex flex-wrap gap-4 justify-center max-w-2xl",children:[e.jsx(m,{trigger:e.jsx(n,{variant:"outline",size:"sm",iconBefore:e.jsx(Be,{}),children:"Consent Info"}),header:"What is Consent?",position:"bottom",width:"sm",showArrow:!0,children:e.jsx("p",{className:"text-sm",children:"Consent allows applications to access your data with your permission. You can revoke access at any time."})}),e.jsx(m,{trigger:e.jsx(n,{variant:"primary",size:"sm",children:"Grant Access"}),header:"Grant Consent",position:"bottom",width:"md",showArrow:!0,footer:e.jsxs("div",{className:"flex gap-2 justify-end",children:[e.jsx(n,{size:"sm",variant:"outline",children:"Deny"}),e.jsx(n,{size:"sm",variant:"primary",children:"Allow"})]}),children:e.jsxs("div",{className:"space-y-2",children:[e.jsxs("p",{className:"text-sm font-medium",children:[e.jsx("strong",{children:"Analytics Platform"})," requests:"]}),e.jsxs("ul",{className:"list-disc list-inside text-sm space-y-1 pl-2",children:[e.jsx("li",{children:"Read profile data"}),e.jsx("li",{children:"Access usage statistics"})]})]})}),e.jsx(m,{trigger:e.jsx(n,{variant:"outline",size:"sm",children:"View Details"}),header:"Active Consent",position:"bottom",width:"md",showArrow:!0,children:e.jsxs("div",{className:"space-y-3 text-sm",children:[e.jsxs("div",{children:[e.jsx("p",{className:"font-medium text-gray-900 mb-1",children:"Marketing Dashboard"}),e.jsx("p",{className:"text-xs text-gray-600",children:"Granted on Dec 1, 2025 - Expires in 15 days"})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-xs font-medium text-gray-700 mb-1",children:"Permissions:"}),e.jsxs("ul",{className:"text-xs text-gray-600 space-y-0.5 pl-3",children:[e.jsx("li",{children:"• Basic profile access"}),e.jsx("li",{children:"• Usage data (read-only)"})]})]})]})}),e.jsx(m,{trigger:e.jsx(n,{variant:"danger",size:"sm",children:"Revoke"}),header:"Revoke Consent?",position:"top",width:"sm",showArrow:!0,footer:e.jsxs("div",{className:"flex gap-2 justify-end",children:[e.jsx(n,{size:"sm",variant:"outline",children:"Cancel"}),e.jsx(n,{size:"sm",variant:"danger",children:"Revoke"})]}),children:e.jsx("p",{className:"text-sm",children:"This will immediately remove the application's access to your data."})})]})]}),args:{trigger:e.jsx(n,{variant:"outline",children:"Consent"}),children:"Consent content",position:"bottom",showArrow:!0}};var Se,Te,Ee,Ie,ke;Y.parameters={...Y.parameters,docs:{...(Se=Y.parameters)==null?void 0:Se.docs,source:{originalSource:`{
  render: args => {
    return <div className="flex items-center justify-center min-h-[400px]">
        <Popover {...args} trigger={<Button variant="outline">Open Popover</Button>}>
          <p>
            This is a basic popover with simple content. Click outside or press
            ESC to close.
          </p>
        </Popover>
      </div>;
  },
  args: {
    trigger: <Button variant="outline">Open Popover</Button>,
    children: 'Default popover content',
    position: 'bottom',
    width: 'md',
    showArrow: true
  }
}`,...(Ee=(Te=Y.parameters)==null?void 0:Te.docs)==null?void 0:Ee.source},description:{story:"Default popover with basic content",...(ke=(Ie=Y.parameters)==null?void 0:Ie.docs)==null?void 0:ke.description}}};var Oe,Re,$e,De,Me;J.parameters={...J.parameters,docs:{...(Oe=J.parameters)==null?void 0:Oe.docs,source:{originalSource:`{
  render: () => {
    return <div className="flex items-center justify-center min-h-[600px] gap-20">
        <div className="grid grid-cols-3 gap-20 items-center">
          {/* Top */}
          <div className="col-start-2 flex justify-center">
            <Popover trigger={<Button variant="outline">Top</Button>} position="top" showArrow>
              <p>Popover positioned at the top</p>
            </Popover>
          </div>

          {/* Left */}
          <div className="col-start-1 flex justify-center">
            <Popover trigger={<Button variant="outline">Left</Button>} position="left" showArrow>
              <p>Popover positioned on the left</p>
            </Popover>
          </div>

          {/* Center placeholder */}
          <div className="col-start-2 flex justify-center">
            <div className="w-24 h-24 rounded-lg border-2 border-dashed border-gray-300 flex items-center justify-center text-sm text-gray-500">
              Trigger
            </div>
          </div>

          {/* Right */}
          <div className="col-start-3 flex justify-center">
            <Popover trigger={<Button variant="outline">Right</Button>} position="right" showArrow>
              <p>Popover positioned on the right</p>
            </Popover>
          </div>

          {/* Bottom */}
          <div className="col-start-2 flex justify-center">
            <Popover trigger={<Button variant="outline">Bottom</Button>} position="bottom" showArrow>
              <p>Popover positioned at the bottom</p>
            </Popover>
          </div>
        </div>
      </div>;
  },
  args: {
    trigger: <Button variant="outline">Position</Button>,
    children: 'Popover content',
    position: 'bottom',
    showArrow: true
  }
}`,...($e=(Re=J.parameters)==null?void 0:Re.docs)==null?void 0:$e.source},description:{story:"All position variants: top, right, bottom, left",...(Me=(De=J.parameters)==null?void 0:De.docs)==null?void 0:Me.description}}};var We,ze,Fe,Ue,Ge;X.parameters={...X.parameters,docs:{...(We=X.parameters)==null?void 0:We.docs,source:{originalSource:`{
  render: () => {
    return <div className="flex items-center justify-center min-h-[400px]">
        <Popover trigger={<Button variant="outline" iconBefore={<InfoIcon />}>
              User Information
            </Button>} header="Profile Details" position="bottom" showArrow>
          <div className="space-y-3">
            <div>
              <p className="text-xs font-medium text-gray-500 mb-1">Name</p>
              <p className="font-medium text-gray-900">John Doe</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 mb-1">Email</p>
              <p className="text-gray-700">john.doe@example.com</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 mb-1">Role</p>
              <p className="text-gray-700">Administrator</p>
            </div>
          </div>
        </Popover>
      </div>;
  },
  args: {
    trigger: <Button variant="outline">User Info</Button>,
    children: 'User information content',
    header: 'Profile Details',
    position: 'bottom',
    showArrow: true
  }
}`,...(Fe=(ze=X.parameters)==null?void 0:ze.docs)==null?void 0:Fe.source},description:{story:"Popover with header section",...(Ge=(Ue=X.parameters)==null?void 0:Ue.docs)==null?void 0:Ge.description}}};var qe,Le,_e,Ke,Ve;Z.parameters={...Z.parameters,docs:{...(qe=Z.parameters)==null?void 0:qe.docs,source:{originalSource:`{
  render: () => {
    return <div className="flex items-center justify-center min-h-[400px]">
        <Popover trigger={<Button variant="outline">Quick Actions</Button>} position="bottom" showArrow footer={<div className="flex gap-2 justify-end">
              <Button size="sm" variant="outline">
                Cancel
              </Button>
              <Button size="sm" variant="primary">
                Apply
              </Button>
            </div>}>
          <p>
            Select an action to perform. Use the buttons below to confirm or
            cancel.
          </p>
        </Popover>
      </div>;
  },
  args: {
    trigger: <Button variant="outline">Quick Actions</Button>,
    children: 'Action content',
    position: 'bottom',
    showArrow: true
  }
}`,...(_e=(Le=Z.parameters)==null?void 0:Le.docs)==null?void 0:_e.source},description:{story:"Popover with footer actions",...(Ve=(Ke=Z.parameters)==null?void 0:Ke.docs)==null?void 0:Ve.description}}};var He,Qe,Ye,Je,Xe;ee.parameters={...ee.parameters,docs:{...(He=ee.parameters)==null?void 0:He.docs,source:{originalSource:`{
  render: () => {
    return <div className="flex items-center justify-center min-h-[500px]">
        <Popover trigger={<Button variant="primary">Confirm Consent</Button>} header="Grant Access" position="bottom" width="md" showArrow footer={<div className="flex gap-2 justify-end">
              <Button size="sm" variant="outline">
                Deny
              </Button>
              <Button size="sm" variant="primary">
                Grant Access
              </Button>
            </div>}>
          <div className="space-y-3">
            <p>
              <strong>Analytics Dashboard</strong> is requesting access to:
            </p>
            <ul className="list-disc list-inside space-y-1 text-sm pl-2">
              <li>Basic profile information</li>
              <li>Usage statistics</li>
              <li>Preference settings</li>
            </ul>
            <div className="mt-3 p-3 bg-info-light border border-info-primary/20 rounded-md">
              <p className="text-xs text-info-dark">
                This permission will be valid for 30 days and can be revoked at
                any time.
              </p>
            </div>
          </div>
        </Popover>
      </div>;
  },
  args: {
    trigger: <Button variant="primary">Confirm Consent</Button>,
    children: 'Consent content',
    header: 'Grant Access',
    position: 'bottom',
    width: 'md',
    showArrow: true
  }
}`,...(Ye=(Qe=ee.parameters)==null?void 0:Qe.docs)==null?void 0:Ye.source},description:{story:"Popover with both header and footer",...(Xe=(Je=ee.parameters)==null?void 0:Je.docs)==null?void 0:Xe.description}}};var Ze,et,tt,rt,ot;te.parameters={...te.parameters,docs:{...(Ze=te.parameters)==null?void 0:Ze.docs,source:{originalSource:`{
  render: () => {
    return <div className="flex items-center justify-center min-h-[400px] gap-8">
        <Popover trigger={<Button variant="outline">With Arrow</Button>} position="bottom" showArrow={true}>
          <p>This popover has an arrow indicator pointing to the trigger.</p>
        </Popover>

        <Popover trigger={<Button variant="outline">Without Arrow</Button>} position="bottom" showArrow={false}>
          <p>This popover has no arrow indicator.</p>
        </Popover>
      </div>;
  },
  args: {
    trigger: <Button variant="outline">Arrow Toggle</Button>,
    children: 'Arrow content',
    position: 'bottom',
    showArrow: true
  }
}`,...(tt=(et=te.parameters)==null?void 0:et.docs)==null?void 0:tt.source},description:{story:"Arrow indicator toggle",...(ot=(rt=te.parameters)==null?void 0:rt.docs)==null?void 0:ot.description}}};var nt,st,it,at,lt;re.parameters={...re.parameters,docs:{...(nt=re.parameters)==null?void 0:nt.docs,source:{originalSource:`{
  render: () => {
    return <div className="flex items-center justify-center min-h-[500px] gap-6">
        <Popover trigger={<Button variant="outline">Small (300px)</Button>} header="Small Width" position="bottom" width="sm" showArrow>
          <p>This is a small popover at 300px width.</p>
        </Popover>

        <Popover trigger={<Button variant="outline">Medium (400px)</Button>} header="Medium Width" position="bottom" width="md" showArrow>
          <p>This is a medium popover at 400px width (default).</p>
        </Popover>

        <Popover trigger={<Button variant="outline">Large (500px)</Button>} header="Large Width" position="bottom" width="lg" showArrow>
          <div className="space-y-3">
            <p>This is a large popover at 500px width.</p>
            <p className="text-sm text-gray-600">
              Perfect for more detailed content or data tables.
            </p>
            <div className="p-3 bg-gray-50 rounded-md border border-gray-200">
              <p className="text-xs font-medium text-gray-700">
                Example content area with more space
              </p>
            </div>
          </div>
        </Popover>
      </div>;
  },
  args: {
    trigger: <Button variant="outline">Width Demo</Button>,
    children: 'Width content',
    position: 'bottom',
    width: 'md',
    showArrow: true
  }
}`,...(it=(st=re.parameters)==null?void 0:st.docs)==null?void 0:it.source},description:{story:"Wide content demonstration",...(lt=(at=re.parameters)==null?void 0:at.docs)==null?void 0:lt.description}}};var ct,dt,pt,ut,mt;oe.parameters={...oe.parameters,docs:{...(ct=oe.parameters)==null?void 0:ct.docs,source:{originalSource:`{
  render: () => {
    return <div className="flex flex-col items-center justify-center min-h-[400px] gap-8">
        <div className="text-center max-w-md mb-4">
          <h3 className="text-lg font-semibold text-gray-900 mb-2">
            Keyboard Navigation
          </h3>
          <p className="text-sm text-gray-600">
            Try using Tab to navigate between buttons, then press Enter or Space
            to open the popover. Press Escape to close.
          </p>
        </div>

        <div className="flex gap-4">
          <Popover trigger={<button className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-trust-deep focus:ring-offset-2" tabIndex={0}>
                <span className="flex items-center gap-2">
                  <InfoIcon />
                  Focusable Button 1
                </span>
              </button>} header="Accessible Popover" position="bottom" showArrow>
            <p>
              This popover can be triggered with keyboard navigation. Press Tab
              to focus, Enter to open, and Escape to close.
            </p>
          </Popover>

          <Popover trigger={<button className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-trust-deep focus:ring-offset-2" tabIndex={0}>
                <span className="flex items-center gap-2">
                  <UserIcon />
                  Focusable Button 2
                </span>
              </button>} header="WCAG 2.1 AA Compliant" position="bottom" showArrow>
            <p>
              All popovers support full keyboard accessibility and screen reader
              announcements.
            </p>
          </Popover>
        </div>
      </div>;
  },
  args: {
    trigger: <button>Keyboard Focus</button>,
    children: 'Keyboard focus content',
    position: 'bottom',
    showArrow: true
  }
}`,...(pt=(dt=oe.parameters)==null?void 0:dt.docs)==null?void 0:pt.source},description:{story:"Keyboard focus accessibility",...(mt=(ut=oe.parameters)==null?void 0:ut.docs)==null?void 0:mt.description}}};var ht,vt,gt,ft,xt;ne.parameters={...ne.parameters,docs:{...(ht=ne.parameters)==null?void 0:ht.docs,source:{originalSource:`{
  render: args => {
    return <div className="flex items-center justify-center min-h-[600px]">
        <Popover {...args} trigger={<Button variant="primary">Open Playground</Button>}>
          {args.children || <div>
              <p className="mb-3">
                Use the controls below to customize this popover's appearance and
                behavior.
              </p>
              <ul className="list-disc list-inside text-sm space-y-1">
                <li>Change position (top, right, bottom, left)</li>
                <li>Adjust width (sm, md, lg, full)</li>
                <li>Toggle arrow indicator</li>
                <li>Add header or footer sections</li>
              </ul>
            </div>}
        </Popover>
      </div>;
  },
  args: {
    trigger: <Button variant="primary">Open Playground</Button>,
    children: 'Playground content',
    position: 'bottom',
    width: 'md',
    showArrow: true,
    header: 'Playground Popover'
  }
}`,...(gt=(vt=ne.parameters)==null?void 0:vt.docs)==null?void 0:gt.source},description:{story:"Interactive playground with all controls",...(xt=(ft=ne.parameters)==null?void 0:ft.docs)==null?void 0:xt.description}}};var yt,bt,wt,jt,Pt;se.parameters={...se.parameters,docs:{...(yt=se.parameters)==null?void 0:yt.docs,source:{originalSource:`{
  render: () => {
    const [isOpen, setIsOpen] = useState(false);
    return <div className="flex flex-col items-center justify-center min-h-[400px] gap-6">
        <div className="text-center">
          <p className="text-sm text-gray-600 mb-3">
            External controls (controlled mode):
          </p>
          <div className="flex gap-2 justify-center">
            <Button size="sm" variant="outline" onClick={() => setIsOpen(true)}>
              Open Popover
            </Button>
            <Button size="sm" variant="outline" onClick={() => setIsOpen(false)}>
              Close Popover
            </Button>
          </div>
          <p className="text-xs text-gray-500 mt-2">
            Current state: {isOpen ? 'Open' : 'Closed'}
          </p>
        </div>

        <Popover trigger={<Button variant="primary">Controlled Popover</Button>} header="Controlled Mode" position="bottom" showArrow isOpen={isOpen} onOpenChange={setIsOpen}>
          <p>
            This popover's state is controlled externally. You can open/close it
            using the buttons above or by clicking the trigger.
          </p>
        </Popover>
      </div>;
  },
  args: {
    trigger: <Button variant="primary">Controlled</Button>,
    children: 'Controlled content',
    position: 'bottom',
    showArrow: true
  }
}`,...(wt=(bt=se.parameters)==null?void 0:bt.docs)==null?void 0:wt.source},description:{story:"Controlled mode with custom state management",...(Pt=(jt=se.parameters)==null?void 0:jt.docs)==null?void 0:Pt.description}}};var Nt,At,Bt,Ct,St;ie.parameters={...ie.parameters,docs:{...(Nt=ie.parameters)==null?void 0:Nt.docs,source:{originalSource:`{
  render: () => {
    return <div className="flex flex-col items-center justify-center min-h-[500px] gap-6">
        <h3 className="text-lg font-semibold text-gray-900">
          Consent Management Use Cases
        </h3>

        <div className="flex flex-wrap gap-4 justify-center max-w-2xl">
          {/* Quick Info Popover */}
          <Popover trigger={<Button variant="outline" size="sm" iconBefore={<InfoIcon />}>
                Consent Info
              </Button>} header="What is Consent?" position="bottom" width="sm" showArrow>
            <p className="text-sm">
              Consent allows applications to access your data with your
              permission. You can revoke access at any time.
            </p>
          </Popover>

          {/* Grant Consent Popover */}
          <Popover trigger={<Button variant="primary" size="sm">Grant Access</Button>} header="Grant Consent" position="bottom" width="md" showArrow footer={<div className="flex gap-2 justify-end">
                <Button size="sm" variant="outline">
                  Deny
                </Button>
                <Button size="sm" variant="primary">
                  Allow
                </Button>
              </div>}>
            <div className="space-y-2">
              <p className="text-sm font-medium">
                <strong>Analytics Platform</strong> requests:
              </p>
              <ul className="list-disc list-inside text-sm space-y-1 pl-2">
                <li>Read profile data</li>
                <li>Access usage statistics</li>
              </ul>
            </div>
          </Popover>

          {/* View Details Popover */}
          <Popover trigger={<Button variant="outline" size="sm">
                View Details
              </Button>} header="Active Consent" position="bottom" width="md" showArrow>
            <div className="space-y-3 text-sm">
              <div>
                <p className="font-medium text-gray-900 mb-1">
                  Marketing Dashboard
                </p>
                <p className="text-xs text-gray-600">
                  Granted on Dec 1, 2025 - Expires in 15 days
                </p>
              </div>
              <div>
                <p className="text-xs font-medium text-gray-700 mb-1">
                  Permissions:
                </p>
                <ul className="text-xs text-gray-600 space-y-0.5 pl-3">
                  <li>• Basic profile access</li>
                  <li>• Usage data (read-only)</li>
                </ul>
              </div>
            </div>
          </Popover>

          {/* Revoke Warning Popover */}
          <Popover trigger={<Button variant="danger" size="sm">
                Revoke
              </Button>} header="Revoke Consent?" position="top" width="sm" showArrow footer={<div className="flex gap-2 justify-end">
                <Button size="sm" variant="outline">
                  Cancel
                </Button>
                <Button size="sm" variant="danger">
                  Revoke
                </Button>
              </div>}>
            <p className="text-sm">
              This will immediately remove the application's access to your data.
            </p>
          </Popover>
        </div>
      </div>;
  },
  args: {
    trigger: <Button variant="outline">Consent</Button>,
    children: 'Consent content',
    position: 'bottom',
    showArrow: true
  }
}`,...(Bt=(At=ie.parameters)==null?void 0:At.docs)==null?void 0:Bt.source},description:{story:"Real-world consent management examples",...(St=(Ct=ie.parameters)==null?void 0:Ct.docs)==null?void 0:St.description}}};const $r=["Default","Positions","WithHeader","WithFooter","WithHeaderAndFooter","WithArrow","WideContent","KeyboardFocus","Playground","Controlled","ConsentExamples"];export{ie as ConsentExamples,se as Controlled,Y as Default,oe as KeyboardFocus,ne as Playground,J as Positions,re as WideContent,te as WithArrow,Z as WithFooter,X as WithHeader,ee as WithHeaderAndFooter,$r as __namedExportsOrder,Rr as default};
