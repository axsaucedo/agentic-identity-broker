import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as J}from"./index-ClcD9ViR.js";import{c as x,a as K}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const Q=K("rounded-full border-1.5 transition-all cursor-pointer flex-shrink-0",{variants:{size:{sm:"w-4 h-4",md:"w-5 h-5",lg:"w-6 h-6"},variant:{default:"border-gray-300 bg-white text-trust-deep",error:"border-error-primary bg-error-light/20 text-error-primary",success:"border-success-primary bg-success-light/20 text-success-primary"}},defaultVariants:{size:"md",variant:"default"}}),a=J.forwardRef(({size:L="md",label:u,description:p,errorMessage:s,successMessage:r,className:P,id:v,disabled:b=!1,checked:I,...Y},B)=>{const H=s?"error":r?"success":"default";return e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex items-center h-5 pt-0.5",children:e.jsx("input",{ref:B,id:v,type:"radio",checked:I,disabled:b,className:x(Q({size:L,variant:H}),"accent-current",b&&"opacity-50 cursor-not-allowed",P),...Y})}),(u||p||s||r)&&e.jsxs("div",{className:"flex flex-col gap-1",children:[u&&e.jsx("label",{htmlFor:v,className:x("text-sm font-medium",b?"text-gray-500 cursor-not-allowed":"text-gray-900 cursor-pointer"),children:u}),p&&!s&&!r&&e.jsx("p",{className:"text-xs text-gray-600",children:p}),s&&e.jsx("p",{className:"text-xs text-error-primary font-medium",children:s}),r&&!s&&e.jsx("p",{className:"text-xs text-success-primary font-medium",children:r})]})]})});a.displayName="Radio";a.__docgenInfo={description:`Radio component for single selection from multiple options.
Supports labels, descriptions, and validation states.

@example
\`\`\`tsx
<fieldset>
  <legend>Choose an option</legend>
  <Radio name="option" value="1" label="Option 1" />
  <Radio name="option" value="2" label="Option 2" />
</fieldset>

<Radio
  name="duration"
  value="7days"
  label="7 days"
  description="Grant access for one week"
  checked={true}
/>
\`\`\``,methods:[],displayName:"Radio",props:{label:{required:!1,tsType:{name:"string"},description:"Label text displayed next to radio"},description:{required:!1,tsType:{name:"string"},description:"Description text displayed below label"},errorMessage:{required:!1,tsType:{name:"string"},description:"Error message (changes variant to error)"},successMessage:{required:!1,tsType:{name:"string"},description:"Success message"},id:{required:!1,tsType:{name:"string"},description:"Unique identifier for form association"},size:{defaultValue:{value:"'md'",computed:!1},required:!1},disabled:{defaultValue:{value:"false",computed:!1},required:!1}},composes:["Omit","VariantProps"]};const ee={title:"Design System/Inputs/Radio",component:a,parameters:{layout:"centered"},tags:["autodocs"]},n={args:{label:"Select this option",name:"example",value:"option1"}},t={render:()=>e.jsxs("div",{className:"space-y-4",children:[e.jsx(a,{size:"sm",label:"Small radio",name:"size",value:"sm"}),e.jsx(a,{size:"md",label:"Medium radio (default)",name:"size",value:"md",checked:!0}),e.jsx(a,{size:"lg",label:"Large radio",name:"size",value:"lg"})]})},i={render:()=>e.jsxs("div",{className:"space-y-4",children:[e.jsx(a,{label:"Unchecked",name:"state",value:"unchecked"}),e.jsx(a,{label:"Checked",name:"state",value:"checked",checked:!0}),e.jsx(a,{label:"Disabled",name:"state",value:"disabled",disabled:!0}),e.jsx(a,{label:"Disabled checked",name:"state",value:"disabled-checked",checked:!0,disabled:!0})]})},d={render:()=>e.jsxs("div",{className:"space-y-4 w-96",children:[e.jsx(a,{label:"7 days",description:"Grant access for one week",name:"duration",value:"7days"}),e.jsx(a,{label:"30 days",description:"Grant access for one month",name:"duration",value:"30days",checked:!0}),e.jsx(a,{label:"No expiration",description:"Grant permanent access",name:"duration",value:"none"})]})},o={render:()=>e.jsxs("fieldset",{className:"space-y-3 p-4 bg-white border border-gray-200 rounded-lg w-96",children:[e.jsx("legend",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Grant Duration"}),e.jsx(a,{name:"duration",value:"24hours",label:"24 hours",description:"Shortest duration"}),e.jsx(a,{name:"duration",value:"7days",label:"7 days",description:"Standard duration",checked:!0}),e.jsx(a,{name:"duration",value:"30days",label:"30 days",description:"Extended duration"}),e.jsx(a,{name:"duration",value:"unlimited",label:"Unlimited",description:"No automatic expiration"})]})},l={render:()=>e.jsxs("div",{className:"space-y-4 w-96",children:[e.jsx(a,{label:"Approve",name:"action",value:"approve",checked:!0,variant:"success",successMessage:"Ready to proceed"}),e.jsx(a,{label:"Reject",name:"action",value:"reject",variant:"error",errorMessage:"You must select an action"})]})},c={args:{label:"Choose this",size:"md",name:"playground",value:"option",disabled:!1,checked:!1}},m={render:()=>e.jsxs("fieldset",{className:"space-y-3 w-96",children:[e.jsx("legend",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Select an access level"}),e.jsx(a,{id:"read-only",name:"access",value:"read",label:"Read-only",description:"View permissions only"}),e.jsx(a,{id:"read-write",name:"access",value:"write",label:"Read and Write",description:"Can view and modify",checked:!0}),e.jsx(a,{id:"admin",name:"access",value:"admin",label:"Admin",description:"Full permissions"})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0}]}}}};var y,h,f;n.parameters={...n.parameters,docs:{...(y=n.parameters)==null?void 0:y.docs,source:{originalSource:`{
  args: {
    label: 'Select this option',
    name: 'example',
    value: 'option1'
  }
}`,...(f=(h=n.parameters)==null?void 0:h.docs)==null?void 0:f.source}}};var g,j,R;t.parameters={...t.parameters,docs:{...(g=t.parameters)==null?void 0:g.docs,source:{originalSource:`{
  render: () => <div className="space-y-4">
      <Radio size="sm" label="Small radio" name="size" value="sm" />
      <Radio size="md" label="Medium radio (default)" name="size" value="md" checked />
      <Radio size="lg" label="Large radio" name="size" value="lg" />
    </div>
}`,...(R=(j=t.parameters)==null?void 0:j.docs)==null?void 0:R.source}}};var k,N,w;i.parameters={...i.parameters,docs:{...(k=i.parameters)==null?void 0:k.docs,source:{originalSource:`{
  render: () => <div className="space-y-4">
      <Radio label="Unchecked" name="state" value="unchecked" />
      <Radio label="Checked" name="state" value="checked" checked />
      <Radio label="Disabled" name="state" value="disabled" disabled />
      <Radio label="Disabled checked" name="state" value="disabled-checked" checked disabled />
    </div>
}`,...(w=(N=i.parameters)==null?void 0:N.docs)==null?void 0:w.source}}};var S,z,D;d.parameters={...d.parameters,docs:{...(S=d.parameters)==null?void 0:S.docs,source:{originalSource:`{
  render: () => <div className="space-y-4 w-96">
      <Radio label="7 days" description="Grant access for one week" name="duration" value="7days" />
      <Radio label="30 days" description="Grant access for one month" name="duration" value="30days" checked={true} />
      <Radio label="No expiration" description="Grant permanent access" name="duration" value="none" />
    </div>
}`,...(D=(z=d.parameters)==null?void 0:z.docs)==null?void 0:D.source}}};var G,V,q;o.parameters={...o.parameters,docs:{...(G=o.parameters)==null?void 0:G.docs,source:{originalSource:`{
  render: () => <fieldset className="space-y-3 p-4 bg-white border border-gray-200 rounded-lg w-96">
      <legend className="text-sm font-semibold text-gray-900 mb-2">
        Grant Duration
      </legend>
      <Radio name="duration" value="24hours" label="24 hours" description="Shortest duration" />
      <Radio name="duration" value="7days" label="7 days" description="Standard duration" checked={true} />
      <Radio name="duration" value="30days" label="30 days" description="Extended duration" />
      <Radio name="duration" value="unlimited" label="Unlimited" description="No automatic expiration" />
    </fieldset>
}`,...(q=(V=o.parameters)==null?void 0:V.docs)==null?void 0:q.source}}};var C,A,E;l.parameters={...l.parameters,docs:{...(C=l.parameters)==null?void 0:C.docs,source:{originalSource:`{
  render: () => <div className="space-y-4 w-96">
      <Radio label="Approve" name="action" value="approve" checked={true} variant="success" successMessage="Ready to proceed" />
      <Radio label="Reject" name="action" value="reject" variant="error" errorMessage="You must select an action" />
    </div>
}`,...(E=(A=l.parameters)==null?void 0:A.docs)==null?void 0:E.source}}};var T,U,O;c.parameters={...c.parameters,docs:{...(T=c.parameters)==null?void 0:T.docs,source:{originalSource:`{
  args: {
    label: 'Choose this',
    size: 'md',
    name: 'playground',
    value: 'option',
    disabled: false,
    checked: false
  }
}`,...(O=(U=c.parameters)==null?void 0:U.docs)==null?void 0:O.source}}};var W,_,F;m.parameters={...m.parameters,docs:{...(W=m.parameters)==null?void 0:W.docs,source:{originalSource:`{
  render: () => <fieldset className="space-y-3 w-96">
      <legend className="text-sm font-semibold text-gray-900 mb-3">
        Select an access level
      </legend>
      <Radio id="read-only" name="access" value="read" label="Read-only" description="View permissions only" />
      <Radio id="read-write" name="access" value="write" label="Read and Write" description="Can view and modify" checked={true} />
      <Radio id="admin" name="access" value="admin" label="Admin" description="Full permissions" />
    </fieldset>,
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
}`,...(F=(_=m.parameters)==null?void 0:_.docs)==null?void 0:F.source}}};const ae=["Default","Sizes","States","WithDescription","Group","ValidationStates","Playground","Accessibility"];export{m as Accessibility,n as Default,o as Group,c as Playground,t as Sizes,i as States,l as ValidationStates,d as WithDescription,ae as __namedExportsOrder,ee as default};
