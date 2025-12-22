import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as Le}from"./index-ClcD9ViR.js";import{c as t,a as Ee}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const ze=Ee("w-full px-3 py-2 rounded-md border-1.5 transition-colors font-sans",{variants:{size:{sm:"px-2.5 py-1.5 text-sm h-8",md:"px-3 py-2 text-base h-10",lg:"px-4 py-3 text-lg h-12"},variant:{default:"border-gray-300 text-gray-900 placeholder-gray-500 focus:border-trust-deep focus:ring-1 focus:ring-trust",error:"border-error-primary bg-error-light/20 text-gray-900 placeholder-gray-500 focus:border-error-primary focus:ring-1 focus:ring-error-primary",success:"border-success-primary bg-success-light/20 text-gray-900 placeholder-gray-500 focus:border-success-primary focus:ring-1 focus:ring-success-primary"}},compoundVariants:[{size:"sm",className:"text-sm"},{size:"lg",className:"text-lg"}],defaultVariants:{size:"md",variant:"default"}}),r=Le.forwardRef(({size:y="md",label:w,required:je=!1,helperText:T,errorMessage:a,successMessage:s,iconBefore:f,iconAfter:x,showCharCount:Ie=!1,className:Ne,id:j,value:g="",maxLength:b,disabled:I=!1,onChange:Se,...qe},Ce)=>{const ke=a?"error":s?"success":"default",N=typeof g=="string"?g.length:0,S=Ie&&b?`${N}/${b}`:null,v=y==="sm"?"w-4 h-4":y==="lg"?"w-5 h-5":"w-4.5 h-4.5";return e.jsxs("div",{ref:Ce,className:"w-full",children:[w&&e.jsxs("label",{htmlFor:j,className:"block text-sm font-medium text-gray-900 mb-2",children:[w,je&&e.jsx("span",{className:"ml-1 text-error-primary",children:"*"})]}),e.jsxs("div",{className:"relative flex items-center",children:[f&&e.jsx("span",{className:t("absolute left-3 text-gray-500 pointer-events-none flex items-center",v),"aria-hidden":"true",children:f}),e.jsx("input",{id:j,value:g,maxLength:b,disabled:I,onChange:Se,className:t(ze({size:y,variant:ke}),f&&"pl-10",x&&"pr-10",I&&"bg-gray-50 cursor-not-allowed opacity-60",Ne),...qe}),x&&e.jsx("span",{className:t("absolute right-3 text-gray-500 pointer-events-none flex items-center",v),"aria-hidden":"true",children:x}),(a||s)&&!x&&e.jsx("span",{className:t("absolute right-3 flex items-center",a?"text-error-primary":"text-success-primary",v),"aria-hidden":"true",children:a?e.jsx("svg",{fill:"currentColor",viewBox:"0 0 20 20",children:e.jsx("path",{fillRule:"evenodd",d:"M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z",clipRule:"evenodd"})}):e.jsx("svg",{fill:"currentColor",viewBox:"0 0 20 20",children:e.jsx("path",{fillRule:"evenodd",d:"M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z",clipRule:"evenodd"})})})]}),e.jsxs("div",{className:"mt-1.5 flex items-center justify-between",children:[a&&e.jsx("span",{className:"text-xs text-error-primary font-medium",children:a}),s&&!a&&e.jsx("span",{className:"text-xs text-success-primary font-medium",children:s}),T&&!a&&!s&&e.jsx("span",{className:"text-xs text-gray-600",children:T}),S&&e.jsx("span",{className:t("text-xs ml-auto",N>b*.8?"text-warning-primary":"text-gray-500"),children:S})]})]})});r.displayName="TextInput";r.__docgenInfo={description:`TextInput component for single-line text entry.
Supports labels, validation states, icons, and helper text.

@example
\`\`\`tsx
<TextInput
  label="Email"
  type="email"
  placeholder="you@example.com"
  helperText="We'll never share your email"
/>

<TextInput
  label="Password"
  type="password"
  errorMessage="Password must be at least 8 characters"
  variant="error"
/>

<TextInput
  label="Search"
  iconBefore={<SearchIcon />}
  placeholder="Search..."
/>

<TextInput
  label="Bio"
  maxLength={100}
  showCharCount
  helperText="Tell us about yourself"
/>
\`\`\``,methods:[],displayName:"TextInput",props:{label:{required:!1,tsType:{name:"string"},description:"Label text displayed above input"},required:{required:!1,tsType:{name:"boolean"},description:"Show required indicator on label",defaultValue:{value:"false",computed:!1}},helperText:{required:!1,tsType:{name:"string"},description:"Helper text displayed below input"},errorMessage:{required:!1,tsType:{name:"string"},description:"Error message (changes variant to error)"},successMessage:{required:!1,tsType:{name:"string"},description:"Success message"},iconBefore:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Icon element to display on the left"},iconAfter:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Icon element to display on the right"},showCharCount:{required:!1,tsType:{name:"boolean"},description:"Show character count",defaultValue:{value:"false",computed:!1}},maxLength:{required:!1,tsType:{name:"number"},description:"Maximum character count for display"},id:{required:!1,tsType:{name:"string"},description:"Unique identifier for form association"},onChange:{required:!1,tsType:{name:"signature",type:"function",raw:"(e: React.ChangeEvent<HTMLInputElement>) => void",signature:{arguments:[{type:{name:"ReactChangeEvent",raw:"React.ChangeEvent<HTMLInputElement>",elements:[{name:"HTMLInputElement"}]},name:"e"}],return:{name:"void"}}},description:"Callback when field state changes"},size:{defaultValue:{value:"'md'",computed:!1},required:!1},value:{defaultValue:{value:"''",computed:!1},required:!1},disabled:{defaultValue:{value:"false",computed:!1},required:!1}},composes:["Omit","VariantProps"]};const Pe={title:"Design System/Inputs/TextInput",component:r,parameters:{layout:"centered",docs:{description:{component:"Flexible text input with validation states, icons, helper text, and character count support."}}},tags:["autodocs"],argTypes:{size:{control:"select",options:["sm","md","lg"]},variant:{control:"select",options:["default","error","success"]},label:{control:"text"},placeholder:{control:"text"},helperText:{control:"text"},errorMessage:{control:"text"},successMessage:{control:"text"},required:{control:"boolean"},disabled:{control:"boolean"},showCharCount:{control:"boolean"}}},l={args:{label:"Email",placeholder:"you@example.com"}},o={render:()=>e.jsxs("div",{className:"w-96 space-y-4",children:[e.jsx(r,{size:"sm",label:"Small",placeholder:"Small input"}),e.jsx(r,{size:"md",label:"Medium (default)",placeholder:"Medium input"}),e.jsx(r,{size:"lg",label:"Large",placeholder:"Large input"})]})},n={render:()=>e.jsxs("div",{className:"w-96 space-y-4",children:[e.jsx(r,{label:"Default",placeholder:"Default state"}),e.jsx(r,{label:"Error",placeholder:"This field has an error",value:"invalid@",variant:"error",errorMessage:"Invalid email format"}),e.jsx(r,{label:"Success",placeholder:"This field is valid",value:"user@example.com",variant:"success",successMessage:"Email looks good!"}),e.jsx(r,{label:"Disabled",placeholder:"Disabled state",disabled:!0})]})},i={render:()=>e.jsxs("div",{className:"w-96 space-y-6",children:[e.jsx(r,{label:"Email",placeholder:"you@example.com",helperText:"We'll never share your email with anyone else"}),e.jsx(r,{label:"Username",placeholder:"choose-a-username",helperText:"3-20 characters, letters and numbers only"})]})},c={render:()=>e.jsxs("div",{className:"w-96 space-y-4",children:[e.jsx(r,{label:"Search",placeholder:"Search...",iconBefore:e.jsx("svg",{fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"})})}),e.jsx(r,{label:"Website",placeholder:"example.com",iconBefore:e.jsx("svg",{fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.658 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"})})})]})},d={render:()=>e.jsxs("div",{className:"w-96 space-y-4",children:[e.jsx(r,{label:"Bio",placeholder:"Tell us about yourself",maxLength:100,showCharCount:!0,helperText:"Make it interesting!"}),e.jsx(r,{label:"Short Title",placeholder:"Keep it short",maxLength:30,showCharCount:!0})]})},p={render:()=>e.jsxs("div",{className:"w-96 space-y-6 p-6 bg-white rounded-lg border border-gray-200",children:[e.jsx("div",{children:e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-4",children:"Sign Up Form"})}),e.jsx(r,{label:"Email",type:"email",placeholder:"you@example.com",required:!0,helperText:"Your email is required to create an account"}),e.jsx(r,{label:"Password",type:"password",placeholder:"••••••••",required:!0,errorMessage:"Password must be at least 8 characters",variant:"error"}),e.jsx(r,{label:"Confirm Password",type:"password",placeholder:"••••••••",required:!0,variant:"success",successMessage:"Passwords match!"})]})},u={render:()=>e.jsxs("div",{className:"w-96 space-y-8",children:[e.jsxs("div",{children:[e.jsx("h4",{className:"text-xs font-semibold text-gray-700 mb-3",children:"Login"}),e.jsx(r,{label:"Email or Username",placeholder:"Enter your credentials",iconBefore:e.jsx("svg",{fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 01-5 0V12m0 0V8.5A2.5 2.5 0 014 11m8-3h.01"})})})]}),e.jsxs("div",{children:[e.jsx("h4",{className:"text-xs font-semibold text-gray-700 mb-3",children:"Search"}),e.jsx(r,{placeholder:"Search agents, scopes...",iconBefore:e.jsx("svg",{fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"})})})]}),e.jsxs("div",{children:[e.jsx("h4",{className:"text-xs font-semibold text-gray-700 mb-3",children:"Required Input"}),e.jsx(r,{label:"Grant Duration",type:"number",placeholder:"Days",required:!0,helperText:"How long should this grant be valid?"})]})]})},m={args:{label:"Enter text",placeholder:"Type something...",size:"md",required:!1,disabled:!1}},h={render:()=>e.jsxs("div",{className:"w-96 space-y-6",children:[e.jsx(r,{id:"email-input",label:"Email Address",type:"email",placeholder:"your.email@example.com",required:!0,helperText:"Required for account creation","aria-describedby":"email-helper"}),e.jsx(r,{id:"password-input",label:"Password",type:"password",placeholder:"••••••••",required:!0,"aria-describedby":"password-requirements",helperText:"Minimum 8 characters, one uppercase, one number"}),e.jsx(r,{id:"confirm-input",label:"Confirm Password",type:"password",placeholder:"••••••••",required:!0})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0},{id:"label",enabled:!0}]}}}};var q,C,k,L,E;l.parameters={...l.parameters,docs:{...(q=l.parameters)==null?void 0:q.docs,source:{originalSource:`{
  args: {
    label: 'Email',
    placeholder: 'you@example.com'
  }
}`,...(k=(C=l.parameters)==null?void 0:C.docs)==null?void 0:k.source},description:{story:"Default text input",...(E=(L=l.parameters)==null?void 0:L.docs)==null?void 0:E.description}}};var z,M,R,W,B;o.parameters={...o.parameters,docs:{...(z=o.parameters)==null?void 0:z.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-4">
      <TextInput size="sm" label="Small" placeholder="Small input" />
      <TextInput size="md" label="Medium (default)" placeholder="Medium input" />
      <TextInput size="lg" label="Large" placeholder="Large input" />
    </div>
}`,...(R=(M=o.parameters)==null?void 0:M.docs)==null?void 0:R.source},description:{story:"All input sizes",...(B=(W=o.parameters)==null?void 0:W.docs)==null?void 0:B.description}}};var P,D,V,H,U;n.parameters={...n.parameters,docs:{...(P=n.parameters)==null?void 0:P.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-4">
      <TextInput label="Default" placeholder="Default state" />
      <TextInput label="Error" placeholder="This field has an error" value="invalid@" variant="error" errorMessage="Invalid email format" />
      <TextInput label="Success" placeholder="This field is valid" value="user@example.com" variant="success" successMessage="Email looks good!" />
      <TextInput label="Disabled" placeholder="Disabled state" disabled />
    </div>
}`,...(V=(D=n.parameters)==null?void 0:D.docs)==null?void 0:V.source},description:{story:"Input states: default, focus, error, success, disabled",...(U=(H=n.parameters)==null?void 0:H.docs)==null?void 0:U.description}}};var A,F,_,G,K;i.parameters={...i.parameters,docs:{...(A=i.parameters)==null?void 0:A.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-6">
      <TextInput label="Email" placeholder="you@example.com" helperText="We'll never share your email with anyone else" />
      <TextInput label="Username" placeholder="choose-a-username" helperText="3-20 characters, letters and numbers only" />
    </div>
}`,...(_=(F=i.parameters)==null?void 0:F.docs)==null?void 0:_.source},description:{story:"With helper text",...(K=(G=i.parameters)==null?void 0:G.docs)==null?void 0:K.description}}};var O,Y,$,J,Q;c.parameters={...c.parameters,docs:{...(O=c.parameters)==null?void 0:O.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-4">
      <TextInput label="Search" placeholder="Search..." iconBefore={<svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>} />
      <TextInput label="Website" placeholder="example.com" iconBefore={<svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.658 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>} />
    </div>
}`,...($=(Y=c.parameters)==null?void 0:Y.docs)==null?void 0:$.source},description:{story:"With icons",...(Q=(J=c.parameters)==null?void 0:J.docs)==null?void 0:Q.description}}};var X,Z,ee,re,ae;d.parameters={...d.parameters,docs:{...(X=d.parameters)==null?void 0:X.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-4">
      <TextInput label="Bio" placeholder="Tell us about yourself" maxLength={100} showCharCount helperText="Make it interesting!" />
      <TextInput label="Short Title" placeholder="Keep it short" maxLength={30} showCharCount />
    </div>
}`,...(ee=(Z=d.parameters)==null?void 0:Z.docs)==null?void 0:ee.source},description:{story:"With character count",...(ae=(re=d.parameters)==null?void 0:re.docs)==null?void 0:ae.description}}};var se,te,le,oe,ne;p.parameters={...p.parameters,docs:{...(se=p.parameters)==null?void 0:se.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-6 p-6 bg-white rounded-lg border border-gray-200">
      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-4">Sign Up Form</h3>
      </div>

      <TextInput label="Email" type="email" placeholder="you@example.com" required helperText="Your email is required to create an account" />

      <TextInput label="Password" type="password" placeholder="••••••••" required errorMessage="Password must be at least 8 characters" variant="error" />

      <TextInput label="Confirm Password" type="password" placeholder="••••••••" required variant="success" successMessage="Passwords match!" />
    </div>
}`,...(le=(te=p.parameters)==null?void 0:te.docs)==null?void 0:le.source},description:{story:"Form validation scenarios",...(ne=(oe=p.parameters)==null?void 0:oe.docs)==null?void 0:ne.description}}};var ie,ce,de,pe,ue;u.parameters={...u.parameters,docs:{...(ie=u.parameters)==null?void 0:ie.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-8">
      {/* Login field */}
      <div>
        <h4 className="text-xs font-semibold text-gray-700 mb-3">Login</h4>
        <TextInput label="Email or Username" placeholder="Enter your credentials" iconBefore={<svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 01-5 0V12m0 0V8.5A2.5 2.5 0 014 11m8-3h.01" />
            </svg>} />
      </div>

      {/* Search field */}
      <div>
        <h4 className="text-xs font-semibold text-gray-700 mb-3">Search</h4>
        <TextInput placeholder="Search agents, scopes..." iconBefore={<svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>} />
      </div>

      {/* Required field */}
      <div>
        <h4 className="text-xs font-semibold text-gray-700 mb-3">Required Input</h4>
        <TextInput label="Grant Duration" type="number" placeholder="Days" required helperText="How long should this grant be valid?" />
      </div>
    </div>
}`,...(de=(ce=u.parameters)==null?void 0:ce.docs)==null?void 0:de.source},description:{story:"Real-world use cases",...(ue=(pe=u.parameters)==null?void 0:pe.docs)==null?void 0:ue.description}}};var me,he,xe,be,ye;m.parameters={...m.parameters,docs:{...(me=m.parameters)==null?void 0:me.docs,source:{originalSource:`{
  args: {
    label: 'Enter text',
    placeholder: 'Type something...',
    size: 'md',
    required: false,
    disabled: false
  }
}`,...(xe=(he=m.parameters)==null?void 0:he.docs)==null?void 0:xe.source},description:{story:"Interactive playground",...(ye=(be=m.parameters)==null?void 0:be.docs)==null?void 0:ye.description}}};var fe,ge,ve,we,Te;h.parameters={...h.parameters,docs:{...(fe=h.parameters)==null?void 0:fe.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-6">
      <TextInput id="email-input" label="Email Address" type="email" placeholder="your.email@example.com" required helperText="Required for account creation" aria-describedby="email-helper" />

      <TextInput id="password-input" label="Password" type="password" placeholder="••••••••" required aria-describedby="password-requirements" helperText="Minimum 8 characters, one uppercase, one number" />

      <TextInput id="confirm-input" label="Confirm Password" type="password" placeholder="••••••••" required />
    </div>,
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
  }
}`,...(ve=(ge=h.parameters)==null?void 0:ge.docs)==null?void 0:ve.source},description:{story:"Accessibility testing",...(Te=(we=h.parameters)==null?void 0:we.docs)==null?void 0:Te.description}}};const De=["Default","Sizes","States","WithHelperText","WithIcons","WithCharCount","ValidationScenarios","UseCases","Playground","Accessibility"];export{h as Accessibility,l as Default,m as Playground,o as Sizes,n as States,u as UseCases,p as ValidationScenarios,d as WithCharCount,i as WithHelperText,c as WithIcons,De as __namedExportsOrder,Pe as default};
