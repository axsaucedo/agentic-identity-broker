import{j as e}from"./jsx-runtime-BYYWji4R.js";import{r as n,R as h}from"./index-ClcD9ViR.js";import{c as w,a as ie}from"./cn-JCLedEej.js";import{T as xe,p as we}from"./form-yc_mpN4r.js";import{p as Ce,r as ve}from"./bugs-Ot56VXf3.js";import{U as oe,o as y,I as le,y as de,l as je,C as F,x as Ne,c as O}from"./keyboard-DNQdDreo.js";import{T as Te}from"./use-resolve-button-type-Bw633SU1.js";import{u as ze,s as Ge}from"./hidden-BnOBa8gi.js";import{G as Ee,w as De}from"./description-Bf17jbPp.js";import"./_commonjsHelpers-Cpj98o6Y.js";let ue=n.createContext(null);function me(){let t=n.useContext(ue);if(t===null){let a=new Error("You used a <Label /> component, but it is not inside a relevant parent.");throw Error.captureStackTrace&&Error.captureStackTrace(a,me),a}return t}function Fe(){let[t,a]=n.useState([]);return[t.length>0?t.join(" "):void 0,n.useMemo(()=>function(s){let r=y(i=>(a(o=>[...o,i]),()=>a(o=>{let m=o.slice(),d=m.indexOf(i);return d!==-1&&m.splice(d,1),m}))),l=n.useMemo(()=>({register:r,slot:s.slot,name:s.name,props:s.props}),[r,s.slot,s.name,s.props]);return h.createElement(ue.Provider,{value:l},s.children)},[a])]}let Pe="label";function Ae(t,a){let s=le(),{id:r=`headlessui-label-${s}`,passive:l=!1,...i}=t,o=me(),m=de(a);je(()=>o.register(r),[r,o.register]);let d={ref:m,...o.props,id:r};return l&&("onClick"in d&&(delete d.htmlFor,delete d.onClick),"onClick"in i&&delete i.onClick),F({ourProps:d,theirProps:i,slot:o.slot||{},defaultTag:Pe,name:o.name||"Label"})}let Re=oe(Ae),Oe=Object.assign(Re,{}),P=n.createContext(null);P.displayName="GroupContext";let Le=n.Fragment;function $e(t){var a;let[s,r]=n.useState(null),[l,i]=Fe(),[o,m]=De(),d=n.useMemo(()=>({switch:s,setSwitch:r,labelledby:l,describedby:o}),[s,r,l,o]),p={},b=t;return h.createElement(m,{name:"Switch.Description"},h.createElement(i,{name:"Switch.Label",props:{htmlFor:(a=d.switch)==null?void 0:a.id,onClick(S){s&&(S.currentTarget.tagName==="LABEL"&&S.preventDefault(),s.click(),s.focus({preventScroll:!0}))}}},h.createElement(P.Provider,{value:d},F({ourProps:p,theirProps:b,defaultTag:Le,name:"Switch.Group"}))))}let qe="button";function Me(t,a){var s;let r=le(),{id:l=`headlessui-switch-${r}`,checked:i,defaultChecked:o=!1,onChange:m,disabled:d=!1,name:p,value:b,form:S,...he}=t,g=n.useContext(P),k=n.useRef(null),pe=de(k,a,g===null?null:g.setSwitch),[f,x]=xe(i,m,o),A=y(()=>x==null?void 0:x(!f)),ge=y(u=>{if(ve(u.currentTarget))return u.preventDefault();u.preventDefault(),A()}),fe=y(u=>{u.key===O.Space?(u.preventDefault(),A()):u.key===O.Enter&&we(u.currentTarget)}),be=y(u=>u.preventDefault()),ye=n.useMemo(()=>({checked:f}),[f]),Se={id:l,ref:pe,role:"switch",type:Te(t,k),tabIndex:t.tabIndex===-1?0:(s=t.tabIndex)!=null?s:0,"aria-checked":f,"aria-labelledby":g==null?void 0:g.labelledby,"aria-describedby":g==null?void 0:g.describedby,disabled:d,onClick:ge,onKeyUp:fe,onKeyPress:be},ke=Ce();return n.useEffect(()=>{var u;let R=(u=k.current)==null?void 0:u.closest("form");R&&o!==void 0&&ke.addEventListener(R,"reset",()=>{x(o)})},[k,x]),h.createElement(h.Fragment,null,p!=null&&f&&h.createElement(ze,{features:Ge.Hidden,...Ne({as:"input",type:"checkbox",hidden:!0,readOnly:!0,disabled:d,form:S,checked:f,name:p,value:b})}),F({ourProps:Se,theirProps:he,slot:ye,defaultTag:qe,name:"Switch"}))}let Ve=oe(Me),Ie=$e,D=Object.assign(Ve,{Group:Ie,Label:Oe,Description:Ee});const We=ie("relative inline-flex flex-shrink-0 items-center rounded-full transition-all duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-offset-2",{variants:{size:{sm:"h-5 w-8 focus:ring-offset-1",md:"h-6 w-11 focus:ring-offset-2",lg:"h-7 w-14 focus:ring-offset-2"},variant:{primary:"focus:ring-trust-deep data-[checked=true]:bg-trust-deep data-[checked=false]:bg-gray-300",success:"focus:ring-success-primary data-[checked=true]:bg-success-primary data-[checked=false]:bg-gray-300"}},defaultVariants:{size:"md",variant:"primary"}}),Ue=ie("inline-block transform rounded-full bg-white shadow-sm transition-transform duration-200 ease-in-out",{variants:{size:{sm:"h-4 w-4 data-[checked=true]:translate-x-3.5 data-[checked=false]:translate-x-0.5",md:"h-5 w-5 data-[checked=true]:translate-x-5 data-[checked=false]:translate-x-0.5",lg:"h-6 w-6 data-[checked=true]:translate-x-6.5 data-[checked=false]:translate-x-0.5"}},defaultVariants:{size:"md"}}),c=h.forwardRef(({size:t="md",variant:a="primary",checked:s,onChange:r,label:l,description:i,disabled:o=!1,id:m,className:d,...p},b)=>e.jsx(D.Group,{children:e.jsxs("div",{ref:b,className:w("flex items-start gap-3",d),...p,children:[e.jsx("div",{className:"flex items-center pt-0.5",children:e.jsx(D,{checked:s,onChange:r,disabled:o,id:m,"data-checked":s,className:w(We({size:t,variant:a}),o&&"opacity-50 cursor-not-allowed"),children:e.jsx("span",{className:w(Ue({size:t}),"shadow-md"),"data-checked":s})})}),(l||i)&&e.jsxs("div",{className:"flex flex-col gap-1",children:[l&&e.jsx(D.Label,{className:w("text-sm font-medium",o?"text-gray-500 cursor-not-allowed":"text-gray-900 cursor-pointer"),children:l}),i&&e.jsx("p",{className:"text-xs text-gray-600",children:i})]})]})}));c.displayName="Switch";c.__docgenInfo={description:`Switch component for binary on/off controls.
Wraps Headless UI Switch with design system styling and semantic tokens.

@example
\`\`\`tsx
<Switch
  checked={enabled}
  onChange={setEnabled}
  label="Enable notifications"
/>

<Switch
  checked={grant}
  onChange={setGrant}
  label="Grant access"
  description="Allow this agent to read your emails"
  size="lg"
/>
\`\`\``,methods:[],displayName:"Switch",props:{checked:{required:!0,tsType:{name:"boolean"},description:"Whether the switch is checked"},onChange:{required:!0,tsType:{name:"signature",type:"function",raw:"(checked: boolean) => void",signature:{arguments:[{type:{name:"boolean"},name:"checked"}],return:{name:"void"}}},description:"Callback when switch state changes"},label:{required:!1,tsType:{name:"string"},description:"Optional label text"},description:{required:!1,tsType:{name:"string"},description:"Description text displayed below label"},disabled:{required:!1,tsType:{name:"boolean"},description:"Whether the switch is disabled",defaultValue:{value:"false",computed:!1}},id:{required:!1,tsType:{name:"string"},description:"Unique identifier for form association"},size:{defaultValue:{value:"'md'",computed:!1},required:!1},variant:{defaultValue:{value:"'primary'",computed:!1},required:!1}},composes:["Omit","VariantProps"]};const tt={title:"Design System/Inputs/Switch",component:c,parameters:{layout:"centered",docs:{description:{component:"Toggle switch for binary on/off controls with labels and descriptions."}}},tags:["autodocs"]},C={render:()=>{const[t,a]=n.useState(!1);return e.jsx(c,{checked:t,onChange:a,label:"Enable feature"})}},v={render:()=>{const[t,a]=n.useState(!0),[s,r]=n.useState(!0),[l,i]=n.useState(!0);return e.jsxs("div",{className:"space-y-6",children:[e.jsx(c,{size:"sm",checked:t,onChange:a,label:"Small"}),e.jsx(c,{size:"md",checked:s,onChange:r,label:"Medium (default)"}),e.jsx(c,{size:"lg",checked:l,onChange:i,label:"Large"})]})}},j={render:()=>{const[t,a]=n.useState(!0),[s,r]=n.useState(!0);return e.jsxs("div",{className:"space-y-6",children:[e.jsx(c,{variant:"primary",checked:t,onChange:a,label:"Primary (trust)"}),e.jsx(c,{variant:"success",checked:s,onChange:r,label:"Success (emerald)"})]})}},N={render:()=>{const[t,a]=n.useState(!1),[s,r]=n.useState(!0);return e.jsxs("div",{className:"space-y-4",children:[e.jsx(c,{checked:t,onChange:a,label:"Off state"}),e.jsx(c,{checked:s,onChange:r,label:"On state"}),e.jsx(c,{checked:!1,onChange:()=>{},label:"Disabled",disabled:!0}),e.jsx(c,{checked:!0,onChange:()=>{},label:"Disabled checked",disabled:!0})]})}},T={render:()=>{const[t,a]=n.useState(!1),[s,r]=n.useState(!0);return e.jsxs("div",{className:"space-y-6 w-96",children:[e.jsx(c,{checked:t,onChange:a,label:"Grant access",description:"Allow this agent to read your emails",size:"lg"}),e.jsx(c,{checked:s,onChange:r,label:"Email notifications",description:"Receive updates about your permissions",size:"lg"})]})}},z={render:()=>{const[t,a]=n.useState(!0),[s,r]=n.useState(!1),[l,i]=n.useState(!0);return e.jsxs("div",{className:"space-y-6 p-6 bg-white border border-gray-200 rounded-lg w-96",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900",children:"Settings"}),e.jsx(c,{checked:t,onChange:a,label:"Notifications",description:"Receive email updates",size:"md"}),e.jsx(c,{checked:s,onChange:r,label:"Two-factor authentication",description:"Require 2FA for logins",size:"md"}),e.jsx(c,{checked:l,onChange:i,label:"Analytics",description:"Help us improve with usage data",size:"md"})]})}},G={render:()=>{const[t,a]=n.useState(!1);return e.jsxs("div",{className:"space-y-4",children:[e.jsx(c,{checked:t,onChange:a,label:"Try toggling me",description:"Click to see me change",size:"lg"}),e.jsxs("div",{className:"text-sm text-gray-600",children:["Current state: ",e.jsx("span",{className:"font-semibold",children:t?"ON":"OFF"})]})]})}},E={render:()=>{const[t,a]=n.useState(!0),[s,r]=n.useState(!1),[l,i]=n.useState(!0);return e.jsxs("fieldset",{className:"space-y-4 p-4 bg-white border border-gray-200 rounded-lg w-96",children:[e.jsx("legend",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Grant Permissions"}),e.jsx(c,{id:"read-emails",checked:t,onChange:a,label:"Read emails",description:"Access to email content"}),e.jsx(c,{id:"send-emails",checked:s,onChange:r,label:"Send emails",description:"Send emails on your behalf"}),e.jsx(c,{id:"manage-labels",checked:l,onChange:i,label:"Manage labels",description:"Create and organize labels"})]})},parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0}]}}}};var L,$,q;C.parameters={...C.parameters,docs:{...(L=C.parameters)==null?void 0:L.docs,source:{originalSource:`{
  render: () => {
    const [checked, setChecked] = useState(false);
    return <Switch checked={checked} onChange={setChecked} label="Enable feature" />;
  }
}`,...(q=($=C.parameters)==null?void 0:$.docs)==null?void 0:q.source}}};var M,V,I;v.parameters={...v.parameters,docs:{...(M=v.parameters)==null?void 0:M.docs,source:{originalSource:`{
  render: () => {
    const [sm, setSm] = useState(true);
    const [md, setMd] = useState(true);
    const [lg, setLg] = useState(true);
    return <div className="space-y-6">
        <Switch size="sm" checked={sm} onChange={setSm} label="Small" />
        <Switch size="md" checked={md} onChange={setMd} label="Medium (default)" />
        <Switch size="lg" checked={lg} onChange={setLg} label="Large" />
      </div>;
  }
}`,...(I=(V=v.parameters)==null?void 0:V.docs)==null?void 0:I.source}}};var W,U,_;j.parameters={...j.parameters,docs:{...(W=j.parameters)==null?void 0:W.docs,source:{originalSource:`{
  render: () => {
    const [primary, setPrimary] = useState(true);
    const [success, setSuccess] = useState(true);
    return <div className="space-y-6">
        <Switch variant="primary" checked={primary} onChange={setPrimary} label="Primary (trust)" />
        <Switch variant="success" checked={success} onChange={setSuccess} label="Success (emerald)" />
      </div>;
  }
}`,...(_=(U=j.parameters)==null?void 0:U.docs)==null?void 0:_.source}}};var H,B,K;N.parameters={...N.parameters,docs:{...(H=N.parameters)==null?void 0:H.docs,source:{originalSource:`{
  render: () => {
    const [checked, setChecked] = useState(false);
    const [checked2, setChecked2] = useState(true);
    return <div className="space-y-4">
        <Switch checked={checked} onChange={setChecked} label="Off state" />
        <Switch checked={checked2} onChange={setChecked2} label="On state" />
        <Switch checked={false} onChange={() => {}} label="Disabled" disabled />
        <Switch checked={true} onChange={() => {}} label="Disabled checked" disabled />
      </div>;
  }
}`,...(K=(B=N.parameters)==null?void 0:B.docs)==null?void 0:K.source}}};var Y,J,Q;T.parameters={...T.parameters,docs:{...(Y=T.parameters)==null?void 0:Y.docs,source:{originalSource:`{
  render: () => {
    const [grant, setGrant] = useState(false);
    const [notify, setNotify] = useState(true);
    return <div className="space-y-6 w-96">
        <Switch checked={grant} onChange={setGrant} label="Grant access" description="Allow this agent to read your emails" size="lg" />
        <Switch checked={notify} onChange={setNotify} label="Email notifications" description="Receive updates about your permissions" size="lg" />
      </div>;
  }
}`,...(Q=(J=T.parameters)==null?void 0:J.docs)==null?void 0:Q.source}}};var X,Z,ee;z.parameters={...z.parameters,docs:{...(X=z.parameters)==null?void 0:X.docs,source:{originalSource:`{
  render: () => {
    const [notifications, setNotifications] = useState(true);
    const [twoFactor, setTwoFactor] = useState(false);
    const [analytics, setAnalytics] = useState(true);
    return <div className="space-y-6 p-6 bg-white border border-gray-200 rounded-lg w-96">
        <h3 className="text-sm font-semibold text-gray-900">Settings</h3>

        <Switch checked={notifications} onChange={setNotifications} label="Notifications" description="Receive email updates" size="md" />

        <Switch checked={twoFactor} onChange={setTwoFactor} label="Two-factor authentication" description="Require 2FA for logins" size="md" />

        <Switch checked={analytics} onChange={setAnalytics} label="Analytics" description="Help us improve with usage data" size="md" />
      </div>;
  }
}`,...(ee=(Z=z.parameters)==null?void 0:Z.docs)==null?void 0:ee.source}}};var te,se,ae;G.parameters={...G.parameters,docs:{...(te=G.parameters)==null?void 0:te.docs,source:{originalSource:`{
  render: () => {
    const [checked, setChecked] = useState(false);
    return <div className="space-y-4">
        <Switch checked={checked} onChange={setChecked} label="Try toggling me" description="Click to see me change" size="lg" />
        <div className="text-sm text-gray-600">
          Current state: <span className="font-semibold">{checked ? 'ON' : 'OFF'}</span>
        </div>
      </div>;
  }
}`,...(ae=(se=G.parameters)==null?void 0:se.docs)==null?void 0:ae.source}}};var ne,re,ce;E.parameters={...E.parameters,docs:{...(ne=E.parameters)==null?void 0:ne.docs,source:{originalSource:`{
  render: () => {
    const [grant1, setGrant1] = useState(true);
    const [grant2, setGrant2] = useState(false);
    const [grant3, setGrant3] = useState(true);
    return <fieldset className="space-y-4 p-4 bg-white border border-gray-200 rounded-lg w-96">
        <legend className="text-sm font-semibold text-gray-900 mb-2">
          Grant Permissions
        </legend>

        <Switch id="read-emails" checked={grant1} onChange={setGrant1} label="Read emails" description="Access to email content" />

        <Switch id="send-emails" checked={grant2} onChange={setGrant2} label="Send emails" description="Send emails on your behalf" />

        <Switch id="manage-labels" checked={grant3} onChange={setGrant3} label="Manage labels" description="Create and organize labels" />
      </fieldset>;
  },
  parameters: {
    a11y: {
      config: {
        rules: [{
          id: 'color-contrast',
          enabled: true
        }]
      }
    }
  }
}`,...(ce=(re=E.parameters)==null?void 0:re.docs)==null?void 0:ce.source}}};const st=["Default","Sizes","Variants","States","WithDescription","RealWorldUseCases","Interactive","Accessibility"];export{E as Accessibility,C as Default,G as Interactive,z as RealWorldUseCases,v as Sizes,N as States,j as Variants,T as WithDescription,st as __namedExportsOrder,tt as default};
