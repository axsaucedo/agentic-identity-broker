import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as Ue,r}from"./index-ClcD9ViR.js";import{c as M,a as k}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const Ge=k("absolute z-50 px-3 py-2 text-sm font-medium rounded-md shadow-lg pointer-events-none transition-opacity duration-150",{variants:{theme:{dark:"bg-gray-900 text-white",light:"bg-white text-gray-900 border border-gray-300 shadow-md"},position:{top:"",right:"",bottom:"",left:""}},defaultVariants:{theme:"dark",position:"top"}}),Ye=k("absolute w-2 h-2 rotate-45",{variants:{theme:{dark:"bg-gray-900",light:"bg-white border-gray-300"},position:{top:"bottom-[-4px] left-1/2 -translate-x-1/2",right:"left-[-4px] top-1/2 -translate-y-1/2",bottom:"top-[-4px] left-1/2 -translate-x-1/2",left:"right-[-4px] top-1/2 -translate-y-1/2"}},defaultVariants:{theme:"dark",position:"top"}}),_e=k("",{variants:{theme:{dark:"",light:"border-l border-t"},position:{top:"",right:"border-l border-t",bottom:"",left:"border-l border-t"}}}),t=Ue.forwardRef(({content:o,children:b,position:v="top",theme:f="dark",showArrow:Ee=!0,delay:A=200,className:Pe,disabled:C=!1,...He},Fe)=>{const[L,w]=r.useState(!1),[Ke,j]=r.useState(!1),s=r.useRef(null),N=r.useRef(null),T=r.useRef(null);r.useEffect(()=>()=>{s.current&&window.clearTimeout(s.current)},[]);const R=()=>{C||(A>0?s.current=window.setTimeout(()=>{w(!0),setTimeout(()=>j(!0),10)},A):(w(!0),setTimeout(()=>j(!0),10)))},W=()=>{s.current&&(window.clearTimeout(s.current),s.current=null),j(!1),setTimeout(()=>w(!1),150)},qe=()=>{if(!T.current||!N.current)return{};T.current.getBoundingClientRect(),N.current.getBoundingClientRect();const y=8;switch(v){case"top":return{bottom:`calc(100% + ${y}px)`,left:"50%",transform:"translateX(-50%)"};case"bottom":return{top:`calc(100% + ${y}px)`,left:"50%",transform:"translateX(-50%)"};case"left":return{right:`calc(100% + ${y}px)`,top:"50%",transform:"translateY(-50%)"};case"right":return{left:`calc(100% + ${y}px)`,top:"50%",transform:"translateY(-50%)"};default:return{}}};return C?e.jsx(e.Fragment,{children:b}):e.jsxs("div",{ref:Fe,className:"relative inline-flex",onMouseEnter:R,onMouseLeave:W,onFocus:R,onBlur:W,...He,children:[e.jsx("div",{ref:T,className:"inline-flex",tabIndex:0,role:"button","aria-describedby":L?"tooltip":void 0,children:b}),L&&e.jsxs("div",{ref:N,id:"tooltip",role:"tooltip",className:M(Ge({theme:f,position:v}),Ke?"opacity-100":"opacity-0",Pe),style:qe(),children:[Ee&&e.jsx("div",{className:M(Ye({theme:f,position:v}),_e({theme:f,position:v}))}),e.jsx("div",{className:"relative z-10 whitespace-nowrap",children:o})]})]})});t.displayName="Tooltip";t.__docgenInfo={description:`Tooltip component for displaying contextual help on hover or focus.
Automatically positions itself based on the position prop.

@example
\`\`\`tsx
<Tooltip content="This is a helpful tip">
  <button>Hover me</button>
</Tooltip>

<Tooltip content="Info about this field" position="right" theme="light">
  <InfoIcon />
</Tooltip>

<Tooltip content="Multi-line content\\nSupported here" showArrow delay={500}>
  <span>Long delay tooltip</span>
</Tooltip>
\`\`\``,methods:[],displayName:"Tooltip",props:{content:{required:!0,tsType:{name:"union",raw:"string | React.ReactNode",elements:[{name:"string"},{name:"ReactReactNode",raw:"React.ReactNode"}]},description:"Content to display in the tooltip"},children:{required:!0,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Element that triggers the tooltip"},position:{required:!1,tsType:{name:"union",raw:"'top' | 'right' | 'bottom' | 'left'",elements:[{name:"literal",value:"'top'"},{name:"literal",value:"'right'"},{name:"literal",value:"'bottom'"},{name:"literal",value:"'left'"}]},description:"Position of tooltip relative to trigger",defaultValue:{value:"'top'",computed:!1}},theme:{required:!1,tsType:{name:"union",raw:"'dark' | 'light'",elements:[{name:"literal",value:"'dark'"},{name:"literal",value:"'light'"}]},description:"Visual theme variant",defaultValue:{value:"'dark'",computed:!1}},showArrow:{required:!1,tsType:{name:"boolean"},description:"Whether to show arrow indicator",defaultValue:{value:"true",computed:!1}},delay:{required:!1,tsType:{name:"number"},description:"Delay in milliseconds before showing tooltip",defaultValue:{value:"200",computed:!1}},className:{required:!1,tsType:{name:"string"},description:"Additional CSS classes"},disabled:{required:!1,tsType:{name:"boolean"},description:"Whether tooltip is disabled",defaultValue:{value:"false",computed:!1}}}};const Ze={title:"Design System/Overlays/Tooltip",component:t,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{content:{control:"text",description:"Content to display in the tooltip"},position:{control:"select",options:["top","right","bottom","left"],description:"Position of tooltip relative to trigger"},theme:{control:"select",options:["dark","light"],description:"Visual theme variant"},showArrow:{control:"boolean",description:"Whether to show arrow indicator"},delay:{control:"number",description:"Delay in milliseconds before showing tooltip"},disabled:{control:"boolean",description:"Whether tooltip is disabled"}}},n={args:{content:"This is a helpful tooltip",position:"top",theme:"dark",showArrow:!0,delay:200,children:e.jsx("button",{className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors",children:"Hover me"})}},i={render:()=>e.jsxs("div",{className:"flex gap-12 items-center justify-center p-24",children:[e.jsx(t,{content:"Top tooltip",position:"top",children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"Top"})}),e.jsx(t,{content:"Right tooltip",position:"right",children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"Right"})}),e.jsx(t,{content:"Bottom tooltip",position:"bottom",children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"Bottom"})}),e.jsx(t,{content:"Left tooltip",position:"left",children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"Left"})})]}),args:{content:"Tooltip content",position:"top",children:e.jsx("button",{children:"Trigger"})}},a={render:()=>e.jsxs("div",{className:"flex gap-12 items-center justify-center p-12",children:[e.jsx(t,{content:"Dark theme tooltip (default)",theme:"dark",children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"Dark Theme"})}),e.jsx(t,{content:"Light theme tooltip with border",theme:"light",children:e.jsx("button",{className:"px-4 py-2 bg-white border border-gray-300 text-gray-900 rounded-md hover:bg-gray-50 transition-colors",children:"Light Theme"})})]}),args:{content:"Tooltip content",theme:"dark",children:e.jsx("button",{children:"Trigger"})}},l={render:()=>e.jsxs("div",{className:"flex gap-12 items-center justify-center p-12",children:[e.jsx(t,{content:"Tooltip with arrow",showArrow:!0,children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"With Arrow"})}),e.jsx(t,{content:"Tooltip without arrow",showArrow:!1,children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"Without Arrow"})})]}),args:{content:"Tooltip content",showArrow:!0,children:e.jsx("button",{children:"Trigger"})}},c={render:()=>e.jsxs("div",{className:"flex gap-8 items-center justify-center p-12",children:[e.jsx(t,{content:"No delay (instant)",delay:0,children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"0ms delay"})}),e.jsx(t,{content:"Default delay",delay:200,children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"200ms delay"})}),e.jsx(t,{content:"Long delay",delay:500,children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"500ms delay"})}),e.jsx(t,{content:"Very long delay",delay:1e3,children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"1000ms delay"})})]}),args:{content:"Tooltip content",delay:200,children:e.jsx("button",{children:"Trigger"})}},d={render:()=>e.jsxs("div",{className:"flex gap-8 items-center justify-center p-12",children:[e.jsx(t,{content:e.jsxs("div",{className:"max-w-xs",children:[e.jsx("p",{className:"mb-2 font-semibold",children:"Enhanced Privacy Controls"}),e.jsx("p",{className:"text-sm leading-relaxed",children:"This feature allows you to manage consent preferences with granular control over data sharing and third-party access."})]}),position:"top",children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"Hover for details"})}),e.jsx(t,{content:e.jsxs("div",{className:"space-y-1 text-xs",children:[e.jsx("div",{className:"font-semibold mb-1",children:"Keyboard Shortcuts:"}),e.jsx("div",{children:"Ctrl + S - Save"}),e.jsx("div",{children:"Ctrl + Z - Undo"}),e.jsx("div",{children:"Ctrl + Y - Redo"}),e.jsx("div",{children:"Esc - Close"})]}),position:"bottom",theme:"light",children:e.jsx("button",{className:"px-4 py-2 bg-white border border-gray-300 text-gray-900 rounded-md hover:bg-gray-50 transition-colors",children:"Shortcuts"})})]}),args:{content:"Long content tooltip",children:e.jsx("button",{children:"Trigger"})}},p={render:()=>{const o=()=>e.jsx("svg",{className:"w-5 h-5 text-gray-500 hover:text-gray-700 transition-colors cursor-help",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"})}),b=()=>e.jsx("svg",{className:"w-5 h-5 text-gray-500 hover:text-gray-700 transition-colors cursor-help",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"})});return e.jsxs("div",{className:"flex flex-col gap-8 items-start p-12 max-w-lg",children:[e.jsxs("div",{className:"flex items-center gap-2",children:[e.jsx("label",{className:"text-sm font-medium text-gray-900",children:"Email Address"}),e.jsx(t,{content:"We'll never share your email with anyone else",position:"right",children:e.jsx(o,{})})]}),e.jsxs("div",{className:"flex items-center gap-2",children:[e.jsx("label",{className:"text-sm font-medium text-gray-900",children:"Data Retention Period"}),e.jsx(t,{content:"How long we keep your data before automatic deletion (30-90 days)",position:"right",theme:"light",children:e.jsx(b,{})})]}),e.jsxs("div",{className:"flex items-center gap-2",children:[e.jsx("label",{className:"text-sm font-medium text-gray-900",children:"Third-Party Access"}),e.jsx(t,{content:e.jsxs("div",{className:"text-xs",children:[e.jsx("div",{className:"font-semibold mb-1",children:"Authorized Services:"}),e.jsx("div",{children:"• Analytics Dashboard"}),e.jsx("div",{children:"• Marketing Platform"}),e.jsx("div",{children:"• CRM Integration"})]}),position:"right",children:e.jsx(o,{})})]})]})},args:{content:"Helpful information",position:"right",children:e.jsx("svg",{className:"w-5 h-5 text-gray-500",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"})})}},m={render:()=>e.jsxs("div",{className:"p-12 space-y-6 max-w-2xl",children:[e.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Keyboard Accessibility"}),e.jsxs("p",{className:"text-sm text-gray-700 mb-3",children:["Press ",e.jsx("kbd",{className:"px-2 py-1 bg-white border border-gray-300 rounded text-xs font-mono",children:"Tab"})," to navigate between elements. Tooltips appear on focus for keyboard users."]})]}),e.jsxs("div",{className:"flex flex-wrap gap-4",children:[e.jsx(t,{content:"First tooltip - accessible via Tab",position:"top",children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 transition-colors",children:"Tab to focus #1"})}),e.jsx(t,{content:"Second tooltip - keyboard accessible",position:"top",children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 transition-colors",children:"Tab to focus #2"})}),e.jsx(t,{content:"Third tooltip with light theme",position:"top",theme:"light",children:e.jsx("button",{className:"px-4 py-2 bg-white border border-gray-300 text-gray-900 rounded-md hover:bg-gray-50 focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 transition-colors",children:"Tab to focus #3"})}),e.jsx(t,{content:"Works with links too",position:"bottom",children:e.jsx("a",{href:"#",className:"inline-flex items-center px-4 py-2 text-navy-700 hover:text-navy-900 underline focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 rounded transition-colors",onClick:o=>o.preventDefault(),children:"Focusable link"})}),e.jsx(t,{content:e.jsxs("div",{className:"text-xs",children:[e.jsx("div",{children:"Shift + Tab to go back"}),e.jsx("div",{children:"Tab to move forward"})]}),position:"bottom",children:e.jsx("button",{className:"px-4 py-2 bg-emerald-600 text-white rounded-md hover:bg-emerald-700 focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2 transition-colors",children:"Interactive element"})})]})]}),args:{content:"Keyboard accessible tooltip",children:e.jsx("button",{children:"Focus me"})}},h={args:{content:"Customize this tooltip using the controls below",position:"top",theme:"dark",showArrow:!0,delay:200,disabled:!1,children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"Hover or focus me"})}},u={render:()=>{const o=()=>e.jsx("svg",{className:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"})});return e.jsx("div",{className:"space-y-8 max-w-3xl p-8",children:e.jsxs("div",{children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-4",children:"Consent Management UI Examples"}),e.jsxs("div",{className:"bg-white border border-gray-200 rounded-lg p-6 space-y-4",children:[e.jsxs("div",{className:"flex items-start justify-between",children:[e.jsxs("div",{className:"flex-1",children:[e.jsxs("div",{className:"flex items-center gap-2 mb-2",children:[e.jsx("h4",{className:"text-base font-semibold text-gray-900",children:"Analytics Dashboard"}),e.jsx(t,{content:"Third-party analytics service for usage insights",position:"right",children:e.jsx("svg",{className:"w-4 h-4 text-gray-400 cursor-help",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"})})})]}),e.jsx("p",{className:"text-sm text-gray-600",children:"Access to usage statistics and behavior patterns"})]}),e.jsx(t,{content:e.jsxs("div",{className:"text-xs space-y-1",children:[e.jsx("div",{className:"font-semibold",children:"Permissions:"}),e.jsx("div",{children:"• View usage data"}),e.jsx("div",{children:"• Read profile info"}),e.jsx("div",{children:"• Anonymous analytics"})]}),position:"left",theme:"light",children:e.jsx("button",{className:"px-3 py-1.5 text-xs font-medium text-navy-700 hover:text-navy-900 border border-navy-300 rounded-md hover:bg-navy-50 transition-colors",children:"View Permissions"})})]}),e.jsxs("div",{className:"flex items-center gap-4 text-xs text-gray-500",children:[e.jsx(t,{content:"Data is encrypted in transit and at rest",position:"bottom",children:e.jsxs("div",{className:"flex items-center gap-1 cursor-help",children:[e.jsx(o,{}),e.jsx("span",{children:"Encrypted"})]})}),e.jsx(t,{content:"Consent expires on Jan 15, 2026",position:"bottom",children:e.jsx("span",{className:"cursor-help",children:"Valid for 30 days"})}),e.jsx(t,{content:"Last accessed 2 hours ago",position:"bottom",children:e.jsx("span",{className:"cursor-help",children:"Active"})})]})]}),e.jsx("div",{className:"mt-6 bg-white border border-gray-200 rounded-lg p-6",children:e.jsxs("div",{className:"space-y-4",children:[e.jsxs("div",{className:"flex items-center justify-between",children:[e.jsxs("div",{className:"flex items-center gap-2",children:[e.jsx("span",{className:"text-sm font-medium text-gray-900",children:"Marketing Communications"}),e.jsx(t,{content:"Receive promotional emails and product updates",position:"right",children:e.jsx("svg",{className:"w-4 h-4 text-gray-400 cursor-help",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"})})})]}),e.jsx(t,{content:"Click to enable marketing emails",position:"left",children:e.jsx("button",{className:"relative inline-flex h-6 w-11 items-center rounded-full bg-gray-200 transition-colors hover:bg-gray-300",children:e.jsx("span",{className:"inline-block h-4 w-4 transform rounded-full bg-white transition-transform translate-x-1"})})})]}),e.jsxs("div",{className:"flex items-center justify-between",children:[e.jsxs("div",{className:"flex items-center gap-2",children:[e.jsx("span",{className:"text-sm font-medium text-gray-900",children:"Usage Analytics"}),e.jsx(t,{content:"Help improve our service by sharing usage data",position:"right",children:e.jsx("svg",{className:"w-4 h-4 text-gray-400 cursor-help",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"})})})]}),e.jsx(t,{content:"Analytics enabled",position:"left",children:e.jsx("button",{className:"relative inline-flex h-6 w-11 items-center rounded-full bg-emerald-600 transition-colors hover:bg-emerald-700",children:e.jsx("span",{className:"inline-block h-4 w-4 transform rounded-full bg-white transition-transform translate-x-6"})})})]})]})})]})})},args:{content:"Real-world example",children:e.jsx("button",{children:"Trigger"})}},g={render:()=>e.jsxs("div",{className:"flex gap-8 items-center justify-center p-12",children:[e.jsx(t,{content:"This tooltip is enabled",disabled:!1,children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors",children:"Enabled Tooltip"})}),e.jsx(t,{content:"This tooltip won't show",disabled:!0,children:e.jsx("button",{className:"px-4 py-2 bg-gray-400 text-white rounded-md cursor-not-allowed",children:"Disabled Tooltip"})})]}),args:{content:"This tooltip is disabled",disabled:!0,children:e.jsx("button",{children:"No tooltip appears"})}},x={render:()=>e.jsxs("div",{className:"space-y-6 max-w-3xl p-8",children:[e.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Accessibility Features"}),e.jsxs("ul",{className:"text-sm text-gray-700 space-y-1",children:[e.jsxs("li",{children:["• Uses ",e.jsx("code",{children:'role="tooltip"'})," for screen readers"]}),e.jsxs("li",{children:["• Trigger has ",e.jsx("code",{children:"aria-describedby"})," pointing to tooltip"]}),e.jsx("li",{children:"• Shows on both hover and keyboard focus (Tab key)"}),e.jsx("li",{children:"• Configurable delay prevents accidental triggers"}),e.jsx("li",{children:"• High contrast themes meet WCAG AA standards"}),e.jsx("li",{children:"• Keyboard navigable with proper focus indicators"}),e.jsx("li",{children:"• Does not trap focus or interfere with navigation"})]})]}),e.jsxs("div",{className:"flex flex-wrap gap-4",children:[e.jsx(t,{content:"WCAG 2.1 AA compliant dark tooltip",position:"top",theme:"dark",children:e.jsx("button",{className:"px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 transition-colors",children:"Dark (AA Compliant)"})}),e.jsx(t,{content:"WCAG 2.1 AA compliant light tooltip",position:"top",theme:"light",children:e.jsx("button",{className:"px-4 py-2 bg-white border border-gray-300 text-gray-900 rounded-md hover:bg-gray-50 focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 transition-colors",children:"Light (AA Compliant)"})})]})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0},{id:"aria-allowed-attr",enabled:!0}]}}},args:{content:"Accessible tooltip",children:e.jsx("button",{children:"Trigger"})}};var S,D,I,V,z;n.parameters={...n.parameters,docs:{...(S=n.parameters)==null?void 0:S.docs,source:{originalSource:`{
  args: {
    content: 'This is a helpful tooltip',
    position: 'top',
    theme: 'dark',
    showArrow: true,
    delay: 200,
    children: <button className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors">
        Hover me
      </button>
  }
}`,...(I=(D=n.parameters)==null?void 0:D.docs)==null?void 0:I.source},description:{story:"Default tooltip with dark theme and top position",...(z=(V=n.parameters)==null?void 0:V.docs)==null?void 0:z.description}}};var B,E,P,H,F;i.parameters={...i.parameters,docs:{...(B=i.parameters)==null?void 0:B.docs,source:{originalSource:`{
  render: () => <div className="flex gap-12 items-center justify-center p-24">
      <Tooltip content="Top tooltip" position="top">
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          Top
        </button>
      </Tooltip>

      <Tooltip content="Right tooltip" position="right">
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          Right
        </button>
      </Tooltip>

      <Tooltip content="Bottom tooltip" position="bottom">
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          Bottom
        </button>
      </Tooltip>

      <Tooltip content="Left tooltip" position="left">
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          Left
        </button>
      </Tooltip>
    </div>,
  args: {
    content: 'Tooltip content',
    position: 'top',
    children: <button>Trigger</button>
  }
}`,...(P=(E=i.parameters)==null?void 0:E.docs)==null?void 0:P.source},description:{story:"All position variants displayed together",...(F=(H=i.parameters)==null?void 0:H.docs)==null?void 0:F.description}}};var K,q,U,G,Y;a.parameters={...a.parameters,docs:{...(K=a.parameters)==null?void 0:K.docs,source:{originalSource:`{
  render: () => <div className="flex gap-12 items-center justify-center p-12">
      <Tooltip content="Dark theme tooltip (default)" theme="dark">
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          Dark Theme
        </button>
      </Tooltip>

      <Tooltip content="Light theme tooltip with border" theme="light">
        <button className="px-4 py-2 bg-white border border-gray-300 text-gray-900 rounded-md hover:bg-gray-50 transition-colors">
          Light Theme
        </button>
      </Tooltip>
    </div>,
  args: {
    content: 'Tooltip content',
    theme: 'dark',
    children: <button>Trigger</button>
  }
}`,...(U=(q=a.parameters)==null?void 0:q.docs)==null?void 0:U.source},description:{story:"Dark and light theme variants",...(Y=(G=a.parameters)==null?void 0:G.docs)==null?void 0:Y.description}}};var _,$,J,O,X;l.parameters={...l.parameters,docs:{...(_=l.parameters)==null?void 0:_.docs,source:{originalSource:`{
  render: () => <div className="flex gap-12 items-center justify-center p-12">
      <Tooltip content="Tooltip with arrow" showArrow={true}>
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          With Arrow
        </button>
      </Tooltip>

      <Tooltip content="Tooltip without arrow" showArrow={false}>
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          Without Arrow
        </button>
      </Tooltip>
    </div>,
  args: {
    content: 'Tooltip content',
    showArrow: true,
    children: <button>Trigger</button>
  }
}`,...(J=($=l.parameters)==null?void 0:$.docs)==null?void 0:J.source},description:{story:"Tooltips with and without arrow indicators",...(X=(O=l.parameters)==null?void 0:O.docs)==null?void 0:X.description}}};var Z,Q,ee,te,oe;c.parameters={...c.parameters,docs:{...(Z=c.parameters)==null?void 0:Z.docs,source:{originalSource:`{
  render: () => <div className="flex gap-8 items-center justify-center p-12">
      <Tooltip content="No delay (instant)" delay={0}>
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          0ms delay
        </button>
      </Tooltip>

      <Tooltip content="Default delay" delay={200}>
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          200ms delay
        </button>
      </Tooltip>

      <Tooltip content="Long delay" delay={500}>
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          500ms delay
        </button>
      </Tooltip>

      <Tooltip content="Very long delay" delay={1000}>
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          1000ms delay
        </button>
      </Tooltip>
    </div>,
  args: {
    content: 'Tooltip content',
    delay: 200,
    children: <button>Trigger</button>
  }
}`,...(ee=(Q=c.parameters)==null?void 0:Q.docs)==null?void 0:ee.source},description:{story:"Tooltips with different hover delays",...(oe=(te=c.parameters)==null?void 0:te.docs)==null?void 0:oe.description}}};var se,re,ne,ie,ae;d.parameters={...d.parameters,docs:{...(se=d.parameters)==null?void 0:se.docs,source:{originalSource:`{
  render: () => <div className="flex gap-8 items-center justify-center p-12">
      <Tooltip content={<div className="max-w-xs">
            <p className="mb-2 font-semibold">Enhanced Privacy Controls</p>
            <p className="text-sm leading-relaxed">
              This feature allows you to manage consent preferences with
              granular control over data sharing and third-party access.
            </p>
          </div>} position="top">
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          Hover for details
        </button>
      </Tooltip>

      <Tooltip content={<div className="space-y-1 text-xs">
            <div className="font-semibold mb-1">Keyboard Shortcuts:</div>
            <div>Ctrl + S - Save</div>
            <div>Ctrl + Z - Undo</div>
            <div>Ctrl + Y - Redo</div>
            <div>Esc - Close</div>
          </div>} position="bottom" theme="light">
        <button className="px-4 py-2 bg-white border border-gray-300 text-gray-900 rounded-md hover:bg-gray-50 transition-colors">
          Shortcuts
        </button>
      </Tooltip>
    </div>,
  args: {
    content: 'Long content tooltip',
    children: <button>Trigger</button>
  }
}`,...(ne=(re=d.parameters)==null?void 0:re.docs)==null?void 0:ne.source},description:{story:"Tooltip with longer multi-line content",...(ae=(ie=d.parameters)==null?void 0:ie.docs)==null?void 0:ae.description}}};var le,ce,de,pe,me;p.parameters={...p.parameters,docs:{...(le=p.parameters)==null?void 0:le.docs,source:{originalSource:`{
  render: () => {
    const InfoIcon = () => <svg className="w-5 h-5 text-gray-500 hover:text-gray-700 transition-colors cursor-help" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>;
    const HelpIcon = () => <svg className="w-5 h-5 text-gray-500 hover:text-gray-700 transition-colors cursor-help" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>;
    return <div className="flex flex-col gap-8 items-start p-12 max-w-lg">
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-gray-900">
            Email Address
          </label>
          <Tooltip content="We'll never share your email with anyone else" position="right">
            <InfoIcon />
          </Tooltip>
        </div>

        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-gray-900">
            Data Retention Period
          </label>
          <Tooltip content="How long we keep your data before automatic deletion (30-90 days)" position="right" theme="light">
            <HelpIcon />
          </Tooltip>
        </div>

        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-gray-900">
            Third-Party Access
          </label>
          <Tooltip content={<div className="text-xs">
                <div className="font-semibold mb-1">Authorized Services:</div>
                <div>• Analytics Dashboard</div>
                <div>• Marketing Platform</div>
                <div>• CRM Integration</div>
              </div>} position="right">
            <InfoIcon />
          </Tooltip>
        </div>
      </div>;
  },
  args: {
    content: 'Helpful information',
    position: 'right',
    children: <svg className="w-5 h-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
  }
}`,...(de=(ce=p.parameters)==null?void 0:ce.docs)==null?void 0:de.source},description:{story:"Tooltip on icon trigger (common use case)",...(me=(pe=p.parameters)==null?void 0:pe.docs)==null?void 0:me.description}}};var he,ue,ge,xe,be;m.parameters={...m.parameters,docs:{...(he=m.parameters)==null?void 0:he.docs,source:{originalSource:`{
  render: () => <div className="p-12 space-y-6 max-w-2xl">
      <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
        <h4 className="text-sm font-semibold text-gray-900 mb-2">
          Keyboard Accessibility
        </h4>
        <p className="text-sm text-gray-700 mb-3">
          Press <kbd className="px-2 py-1 bg-white border border-gray-300 rounded text-xs font-mono">Tab</kbd> to navigate
          between elements. Tooltips appear on focus for keyboard users.
        </p>
      </div>

      <div className="flex flex-wrap gap-4">
        <Tooltip content="First tooltip - accessible via Tab" position="top">
          <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 transition-colors">
            Tab to focus #1
          </button>
        </Tooltip>

        <Tooltip content="Second tooltip - keyboard accessible" position="top">
          <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 transition-colors">
            Tab to focus #2
          </button>
        </Tooltip>

        <Tooltip content="Third tooltip with light theme" position="top" theme="light">
          <button className="px-4 py-2 bg-white border border-gray-300 text-gray-900 rounded-md hover:bg-gray-50 focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 transition-colors">
            Tab to focus #3
          </button>
        </Tooltip>

        <Tooltip content="Works with links too" position="bottom">
          <a href="#" className="inline-flex items-center px-4 py-2 text-navy-700 hover:text-navy-900 underline focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 rounded transition-colors" onClick={e => e.preventDefault()}>
            Focusable link
          </a>
        </Tooltip>

        <Tooltip content={<div className="text-xs">
              <div>Shift + Tab to go back</div>
              <div>Tab to move forward</div>
            </div>} position="bottom">
          <button className="px-4 py-2 bg-emerald-600 text-white rounded-md hover:bg-emerald-700 focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2 transition-colors">
            Interactive element
          </button>
        </Tooltip>
      </div>
    </div>,
  args: {
    content: 'Keyboard accessible tooltip',
    children: <button>Focus me</button>
  }
}`,...(ge=(ue=m.parameters)==null?void 0:ue.docs)==null?void 0:ge.source},description:{story:"Tooltip keyboard focus support for accessibility",...(be=(xe=m.parameters)==null?void 0:xe.docs)==null?void 0:be.description}}};var ve,ye,fe,we,je;h.parameters={...h.parameters,docs:{...(ve=h.parameters)==null?void 0:ve.docs,source:{originalSource:`{
  args: {
    content: 'Customize this tooltip using the controls below',
    position: 'top',
    theme: 'dark',
    showArrow: true,
    delay: 200,
    disabled: false,
    children: <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
        Hover or focus me
      </button>
  }
}`,...(fe=(ye=h.parameters)==null?void 0:ye.docs)==null?void 0:fe.source},description:{story:"Interactive playground with all controls",...(je=(we=h.parameters)==null?void 0:we.docs)==null?void 0:je.description}}};var Ne,Te,ke,Ae,Ce;u.parameters={...u.parameters,docs:{...(Ne=u.parameters)==null?void 0:Ne.docs,source:{originalSource:`{
  render: () => {
    const LockIcon = () => <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
      </svg>;
    return <div className="space-y-8 max-w-3xl p-8">
        <div>
          <h3 className="text-lg font-semibold text-gray-900 mb-4">
            Consent Management UI Examples
          </h3>

          {/* Consent Card Example */}
          <div className="bg-white border border-gray-200 rounded-lg p-6 space-y-4">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <h4 className="text-base font-semibold text-gray-900">
                    Analytics Dashboard
                  </h4>
                  <Tooltip content="Third-party analytics service for usage insights" position="right">
                    <svg className="w-4 h-4 text-gray-400 cursor-help" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </Tooltip>
                </div>
                <p className="text-sm text-gray-600">
                  Access to usage statistics and behavior patterns
                </p>
              </div>
              <Tooltip content={<div className="text-xs space-y-1">
                    <div className="font-semibold">Permissions:</div>
                    <div>• View usage data</div>
                    <div>• Read profile info</div>
                    <div>• Anonymous analytics</div>
                  </div>} position="left" theme="light">
                <button className="px-3 py-1.5 text-xs font-medium text-navy-700 hover:text-navy-900 border border-navy-300 rounded-md hover:bg-navy-50 transition-colors">
                  View Permissions
                </button>
              </Tooltip>
            </div>

            <div className="flex items-center gap-4 text-xs text-gray-500">
              <Tooltip content="Data is encrypted in transit and at rest" position="bottom">
                <div className="flex items-center gap-1 cursor-help">
                  <LockIcon />
                  <span>Encrypted</span>
                </div>
              </Tooltip>
              <Tooltip content="Consent expires on Jan 15, 2026" position="bottom">
                <span className="cursor-help">Valid for 30 days</span>
              </Tooltip>
              <Tooltip content="Last accessed 2 hours ago" position="bottom">
                <span className="cursor-help">Active</span>
              </Tooltip>
            </div>
          </div>

          {/* Permission Toggle Example */}
          <div className="mt-6 bg-white border border-gray-200 rounded-lg p-6">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-gray-900">
                    Marketing Communications
                  </span>
                  <Tooltip content="Receive promotional emails and product updates" position="right">
                    <svg className="w-4 h-4 text-gray-400 cursor-help" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </Tooltip>
                </div>
                <Tooltip content="Click to enable marketing emails" position="left">
                  <button className="relative inline-flex h-6 w-11 items-center rounded-full bg-gray-200 transition-colors hover:bg-gray-300">
                    <span className="inline-block h-4 w-4 transform rounded-full bg-white transition-transform translate-x-1" />
                  </button>
                </Tooltip>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-gray-900">
                    Usage Analytics
                  </span>
                  <Tooltip content="Help improve our service by sharing usage data" position="right">
                    <svg className="w-4 h-4 text-gray-400 cursor-help" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </Tooltip>
                </div>
                <Tooltip content="Analytics enabled" position="left">
                  <button className="relative inline-flex h-6 w-11 items-center rounded-full bg-emerald-600 transition-colors hover:bg-emerald-700">
                    <span className="inline-block h-4 w-4 transform rounded-full bg-white transition-transform translate-x-6" />
                  </button>
                </Tooltip>
              </div>
            </div>
          </div>
        </div>
      </div>;
  },
  args: {
    content: 'Real-world example',
    children: <button>Trigger</button>
  }
}`,...(ke=(Te=u.parameters)==null?void 0:Te.docs)==null?void 0:ke.source},description:{story:"Real-world consent management use cases",...(Ce=(Ae=u.parameters)==null?void 0:Ae.docs)==null?void 0:Ce.description}}};var Le,Re,We,Me,Se;g.parameters={...g.parameters,docs:{...(Le=g.parameters)==null?void 0:Le.docs,source:{originalSource:`{
  render: () => <div className="flex gap-8 items-center justify-center p-12">
      <Tooltip content="This tooltip is enabled" disabled={false}>
        <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 transition-colors">
          Enabled Tooltip
        </button>
      </Tooltip>

      <Tooltip content="This tooltip won't show" disabled={true}>
        <button className="px-4 py-2 bg-gray-400 text-white rounded-md cursor-not-allowed">
          Disabled Tooltip
        </button>
      </Tooltip>
    </div>,
  args: {
    content: 'This tooltip is disabled',
    disabled: true,
    children: <button>No tooltip appears</button>
  }
}`,...(We=(Re=g.parameters)==null?void 0:Re.docs)==null?void 0:We.source},description:{story:"Disabled state - tooltip won't show",...(Se=(Me=g.parameters)==null?void 0:Me.docs)==null?void 0:Se.description}}};var De,Ie,Ve,ze,Be;x.parameters={...x.parameters,docs:{...(De=x.parameters)==null?void 0:De.docs,source:{originalSource:`{
  render: () => <div className="space-y-6 max-w-3xl p-8">
      <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
        <h4 className="text-sm font-semibold text-gray-900 mb-2">
          Accessibility Features
        </h4>
        <ul className="text-sm text-gray-700 space-y-1">
          <li>• Uses <code>role="tooltip"</code> for screen readers</li>
          <li>• Trigger has <code>aria-describedby</code> pointing to tooltip</li>
          <li>• Shows on both hover and keyboard focus (Tab key)</li>
          <li>• Configurable delay prevents accidental triggers</li>
          <li>• High contrast themes meet WCAG AA standards</li>
          <li>• Keyboard navigable with proper focus indicators</li>
          <li>• Does not trap focus or interfere with navigation</li>
        </ul>
      </div>

      <div className="flex flex-wrap gap-4">
        <Tooltip content="WCAG 2.1 AA compliant dark tooltip" position="top" theme="dark">
          <button className="px-4 py-2 bg-navy-600 text-white rounded-md hover:bg-navy-700 focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 transition-colors">
            Dark (AA Compliant)
          </button>
        </Tooltip>

        <Tooltip content="WCAG 2.1 AA compliant light tooltip" position="top" theme="light">
          <button className="px-4 py-2 bg-white border border-gray-300 text-gray-900 rounded-md hover:bg-gray-50 focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 transition-colors">
            Light (AA Compliant)
          </button>
        </Tooltip>
      </div>
    </div>,
  parameters: {
    a11y: {
      config: {
        rules: [{
          id: 'color-contrast',
          enabled: true
        }, {
          id: 'aria-allowed-attr',
          enabled: true
        }]
      }
    }
  },
  args: {
    content: 'Accessible tooltip',
    children: <button>Trigger</button>
  }
}`,...(Ve=(Ie=x.parameters)==null?void 0:Ie.docs)==null?void 0:Ve.source},description:{story:"Accessibility features demonstration",...(Be=(ze=x.parameters)==null?void 0:ze.docs)==null?void 0:Be.description}}};const Qe=["Default","Positions","Themes","WithArrow","WithDelay","LongContent","IconTrigger","KeyboardFocus","Playground","RealWorldExamples","Disabled","Accessibility"];export{x as Accessibility,n as Default,g as Disabled,p as IconTrigger,m as KeyboardFocus,d as LongContent,h as Playground,i as Positions,u as RealWorldExamples,a as Themes,l as WithArrow,c as WithDelay,Qe as __namedExportsOrder,Ze as default};
