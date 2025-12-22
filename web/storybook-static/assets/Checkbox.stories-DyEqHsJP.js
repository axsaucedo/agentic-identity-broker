import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as B}from"./index-ClcD9ViR.js";import{c as x,a as J}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const K=J("rounded border-1.5 transition-all cursor-pointer flex-shrink-0",{variants:{size:{sm:"w-4 h-4",md:"w-5 h-5",lg:"w-6 h-6"},variant:{default:"border-gray-300 bg-white text-trust-deep",error:"border-error-primary bg-error-light/20 text-error-primary",success:"border-success-primary bg-success-light/20 text-success-primary"}},defaultVariants:{size:"md",variant:"default"}}),s=B.forwardRef(({size:G="md",label:p,description:u,errorMessage:a,successMessage:r,className:H,id:h,disabled:b=!1,checked:M,...O},W)=>{const F=a?"error":r?"success":"default";return e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex items-center h-5 pt-0.5",children:e.jsx("input",{ref:W,id:h,type:"checkbox",checked:M,disabled:b,className:x(K({size:G,variant:F}),"accent-current",b&&"opacity-50 cursor-not-allowed",H),...O})}),(p||u||a||r)&&e.jsxs("div",{className:"flex flex-col gap-1",children:[p&&e.jsx("label",{htmlFor:h,className:x("text-sm font-medium",b?"text-gray-500 cursor-not-allowed":"text-gray-900 cursor-pointer"),children:p}),u&&!a&&!r&&e.jsx("p",{className:"text-xs text-gray-600",children:u}),a&&e.jsx("p",{className:"text-xs text-error-primary font-medium",children:a}),r&&!a&&e.jsx("p",{className:"text-xs text-success-primary font-medium",children:r})]})]})});s.displayName="Checkbox";s.__docgenInfo={description:`Checkbox component for boolean selections.
Supports labels, descriptions, and validation states.

@example
\`\`\`tsx
<Checkbox label="I agree to the terms" />

<Checkbox
  label="Enable notifications"
  description="Receive updates about your grants"
/>

<Checkbox
  label="Confirm deletion"
  errorMessage="You must confirm before deleting"
  variant="error"
/>
\`\`\``,methods:[],displayName:"Checkbox",props:{label:{required:!1,tsType:{name:"string"},description:"Label text displayed next to checkbox"},description:{required:!1,tsType:{name:"string"},description:"Description text displayed below label"},errorMessage:{required:!1,tsType:{name:"string"},description:"Error message (changes variant to error)"},successMessage:{required:!1,tsType:{name:"string"},description:"Success message"},id:{required:!1,tsType:{name:"string"},description:"Unique identifier for form association"},size:{defaultValue:{value:"'md'",computed:!1},required:!1},disabled:{defaultValue:{value:"false",computed:!1},required:!1}},composes:["Omit","VariantProps"]};const ee={title:"Design System/Inputs/Checkbox",component:s,parameters:{layout:"centered"},tags:["autodocs"]},c={args:{label:"Accept terms and conditions"}},t={render:()=>e.jsxs("div",{className:"space-y-4",children:[e.jsx(s,{size:"sm",label:"Small checkbox"}),e.jsx(s,{size:"md",label:"Medium checkbox (default)"}),e.jsx(s,{size:"lg",label:"Large checkbox"})]})},o={render:()=>e.jsxs("div",{className:"space-y-4",children:[e.jsx(s,{label:"Unchecked",checked:!1}),e.jsx(s,{label:"Checked",checked:!0}),e.jsx(s,{label:"Disabled",disabled:!0}),e.jsx(s,{label:"Disabled checked",checked:!0,disabled:!0})]})},i={render:()=>e.jsxs("div",{className:"space-y-4 w-96",children:[e.jsx(s,{label:"Email notifications",description:"Receive updates about your grants"}),e.jsx(s,{label:"Share analytics",description:"Help us improve by sharing usage data",checked:!0})]})},l={render:()=>e.jsxs("div",{className:"space-y-4 w-96",children:[e.jsx(s,{label:"I agree",checked:!0,variant:"success",successMessage:"Thank you!"}),e.jsx(s,{label:"Confirm deletion",variant:"error",errorMessage:"You must confirm before deleting"})]})},n={render:()=>e.jsxs("div",{className:"space-y-3 p-4 bg-white border border-gray-200 rounded-lg w-96",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Permissions"}),e.jsx(s,{label:"Read emails",description:"View your email messages",checked:!0}),e.jsx(s,{label:"Send emails",description:"Send emails on your behalf",checked:!0}),e.jsx(s,{label:"Delete emails",description:"Delete your email messages"}),e.jsx(s,{label:"Manage labels",description:"Create and manage email labels",checked:!0})]})},d={args:{label:"Accept",size:"md",disabled:!1,checked:!1}},m={render:()=>e.jsxs("div",{className:"space-y-4 w-96",children:[e.jsx(s,{id:"terms-checkbox",label:"I have read and agree to the terms of service",description:"This is required to continue"}),e.jsx(s,{id:"privacy-checkbox",label:"I consent to the privacy policy",checked:!0})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0}]}}}};var g,k,f;c.parameters={...c.parameters,docs:{...(g=c.parameters)==null?void 0:g.docs,source:{originalSource:`{
  args: {
    label: 'Accept terms and conditions'
  }
}`,...(f=(k=c.parameters)==null?void 0:k.docs)==null?void 0:f.source}}};var y,v,j;t.parameters={...t.parameters,docs:{...(y=t.parameters)==null?void 0:y.docs,source:{originalSource:`{
  render: () => <div className="space-y-4">
      <Checkbox size="sm" label="Small checkbox" />
      <Checkbox size="md" label="Medium checkbox (default)" />
      <Checkbox size="lg" label="Large checkbox" />
    </div>
}`,...(j=(v=t.parameters)==null?void 0:v.docs)==null?void 0:j.source}}};var C,S,N;o.parameters={...o.parameters,docs:{...(C=o.parameters)==null?void 0:C.docs,source:{originalSource:`{
  render: () => <div className="space-y-4">
      <Checkbox label="Unchecked" checked={false} />
      <Checkbox label="Checked" checked={true} />
      <Checkbox label="Disabled" disabled />
      <Checkbox label="Disabled checked" checked={true} disabled />
    </div>
}`,...(N=(S=o.parameters)==null?void 0:S.docs)==null?void 0:N.source}}};var w,D,z;i.parameters={...i.parameters,docs:{...(w=i.parameters)==null?void 0:w.docs,source:{originalSource:`{
  render: () => <div className="space-y-4 w-96">
      <Checkbox label="Email notifications" description="Receive updates about your grants" />
      <Checkbox label="Share analytics" description="Help us improve by sharing usage data" checked={true} />
    </div>
}`,...(z=(D=i.parameters)==null?void 0:D.docs)==null?void 0:z.source}}};var q,I,R;l.parameters={...l.parameters,docs:{...(q=l.parameters)==null?void 0:q.docs,source:{originalSource:`{
  render: () => <div className="space-y-4 w-96">
      <Checkbox label="I agree" checked={true} variant="success" successMessage="Thank you!" />
      <Checkbox label="Confirm deletion" variant="error" errorMessage="You must confirm before deleting" />
    </div>
}`,...(R=(I=l.parameters)==null?void 0:I.docs)==null?void 0:R.source}}};var T,V,A;n.parameters={...n.parameters,docs:{...(T=n.parameters)==null?void 0:T.docs,source:{originalSource:`{
  render: () => <div className="space-y-3 p-4 bg-white border border-gray-200 rounded-lg w-96">
      <h4 className="text-sm font-semibold text-gray-900 mb-2">Permissions</h4>
      <Checkbox label="Read emails" description="View your email messages" checked={true} />
      <Checkbox label="Send emails" description="Send emails on your behalf" checked={true} />
      <Checkbox label="Delete emails" description="Delete your email messages" />
      <Checkbox label="Manage labels" description="Create and manage email labels" checked={true} />
    </div>
}`,...(A=(V=n.parameters)==null?void 0:V.docs)==null?void 0:A.source}}};var E,P,_;d.parameters={...d.parameters,docs:{...(E=d.parameters)==null?void 0:E.docs,source:{originalSource:`{
  args: {
    label: 'Accept',
    size: 'md',
    disabled: false,
    checked: false
  }
}`,...(_=(P=d.parameters)==null?void 0:P.docs)==null?void 0:_.source}}};var L,U,Y;m.parameters={...m.parameters,docs:{...(L=m.parameters)==null?void 0:L.docs,source:{originalSource:`{
  render: () => <div className="space-y-4 w-96">
      <Checkbox id="terms-checkbox" label="I have read and agree to the terms of service" description="This is required to continue" />
      <Checkbox id="privacy-checkbox" label="I consent to the privacy policy" checked={true} />
    </div>,
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
}`,...(Y=(U=m.parameters)==null?void 0:U.docs)==null?void 0:Y.source}}};const se=["Default","Sizes","States","WithDescription","ValidationStates","Group","Playground","Accessibility"];export{m as Accessibility,c as Default,n as Group,d as Playground,t as Sizes,o as States,l as ValidationStates,i as WithDescription,se as __namedExportsOrder,ee as default};
