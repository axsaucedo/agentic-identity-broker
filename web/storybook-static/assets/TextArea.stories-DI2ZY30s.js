import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as y}from"./index-ClcD9ViR.js";import{c as S,a as te}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const oe=te("w-full px-3 py-2 rounded-md border-1.5 transition-colors font-sans resize-none",{variants:{variant:{default:"border-gray-300 text-gray-900 placeholder-gray-500 focus:border-trust-deep focus:ring-1 focus:ring-trust",error:"border-error-primary bg-error-light/20 text-gray-900 placeholder-gray-500 focus:border-error-primary focus:ring-1 focus:ring-error-primary",success:"border-success-primary bg-success-light/20 text-gray-900 placeholder-gray-500 focus:border-success-primary focus:ring-1 focus:ring-success-primary"}},defaultVariants:{variant:"default"}}),r=y.forwardRef(({label:w,required:Q=!1,helperText:v,errorMessage:t,successMessage:l,showCharCount:X=!1,maxLength:n,className:Z,id:T,rows:ee=4,value:f="",autoGrow:o=!1,disabled:j=!1,onChange:g,...re},i)=>{const ae=t?"error":l?"success":"default",C=typeof f=="string"?f.length:0,A=X&&n?`${C}/${n}`:null,a=y.useRef(null),se=s=>{g==null||g(s),o&&a.current&&(a.current.style.height="auto",a.current.style.height=`${a.current.scrollHeight}px`)};return y.useEffect(()=>{o&&a.current&&(a.current.style.height="auto",a.current.style.height=`${a.current.scrollHeight}px`)},[o]),e.jsxs("div",{className:"w-full",children:[w&&e.jsxs("label",{htmlFor:T,className:"block text-sm font-medium text-gray-900 mb-2",children:[w,Q&&e.jsx("span",{className:"ml-1 text-error-primary",children:"*"})]}),e.jsx("textarea",{ref:s=>{s&&(a.current=s,typeof i=="function"?i(s):i&&(i.current=s))},id:T,rows:o?1:ee,value:f,maxLength:n,disabled:j,onChange:se,className:S(oe({variant:ae}),j&&"bg-gray-50 cursor-not-allowed opacity-60",o&&"min-h-[2.5rem] overflow-hidden",Z),...re}),e.jsxs("div",{className:"mt-1.5 flex items-center justify-between",children:[t&&e.jsx("span",{className:"text-xs text-error-primary font-medium",children:t}),l&&!t&&e.jsx("span",{className:"text-xs text-success-primary font-medium",children:l}),v&&!t&&!l&&e.jsx("span",{className:"text-xs text-gray-600",children:v}),A&&e.jsx("span",{className:S("text-xs ml-auto",C>n*.8?"text-warning-primary":"text-gray-500"),children:A})]})]})});r.displayName="TextArea";r.__docgenInfo={description:`TextArea component for multi-line text entry.
Supports labels, validation states, character count, and optional auto-grow.

@example
\`\`\`tsx
<TextArea
  label="Comments"
  placeholder="Enter your feedback..."
  helperText="Your feedback helps us improve"
/>

<TextArea
  label="Description"
  maxLength={500}
  showCharCount
  autoGrow
/>
\`\`\``,methods:[],displayName:"TextArea",props:{label:{required:!1,tsType:{name:"string"},description:"Label text displayed above textarea"},required:{required:!1,tsType:{name:"boolean"},description:"Show required indicator on label",defaultValue:{value:"false",computed:!1}},helperText:{required:!1,tsType:{name:"string"},description:"Helper text displayed below textarea"},errorMessage:{required:!1,tsType:{name:"string"},description:"Error message (changes variant to error)"},successMessage:{required:!1,tsType:{name:"string"},description:"Success message"},rows:{required:!1,tsType:{name:"number"},description:"Number of visible text lines",defaultValue:{value:"4",computed:!1}},autoGrow:{required:!1,tsType:{name:"boolean"},description:"Auto-grow textarea as user types",defaultValue:{value:"false",computed:!1}},showCharCount:{required:!1,tsType:{name:"boolean"},description:"Show character count",defaultValue:{value:"false",computed:!1}},id:{required:!1,tsType:{name:"string"},description:"Unique identifier for form association"},onChange:{required:!1,tsType:{name:"signature",type:"function",raw:"(e: React.ChangeEvent<HTMLTextAreaElement>) => void",signature:{arguments:[{type:{name:"ReactChangeEvent",raw:"React.ChangeEvent<HTMLTextAreaElement>",elements:[{name:"HTMLTextAreaElement"}]},name:"e"}],return:{name:"void"}}},description:"Callback when field state changes"},value:{defaultValue:{value:"''",computed:!1},required:!1},disabled:{defaultValue:{value:"false",computed:!1},required:!1}},composes:["Omit","VariantProps"]};const de={title:"Design System/Inputs/TextArea",component:r,parameters:{layout:"centered",docs:{description:{component:"Multi-line text input with validation states, character count, and optional auto-grow support."}}},tags:["autodocs"]},c={args:{label:"Comments",placeholder:"Enter your feedback...",rows:4}},d={render:()=>e.jsxs("div",{className:"w-96 space-y-4",children:[e.jsx(r,{label:"Small (2 rows)",placeholder:"Compact",rows:2}),e.jsx(r,{label:"Medium (4 rows)",placeholder:"Default size",rows:4}),e.jsx(r,{label:"Large (6 rows)",placeholder:"Spacious",rows:6})]})},u={render:()=>e.jsxs("div",{className:"w-96 space-y-4",children:[e.jsx(r,{label:"Default",placeholder:"Default state",rows:3}),e.jsx(r,{label:"Error",value:"This message contains inappropriate...",variant:"error",errorMessage:"Please avoid using inappropriate language",rows:3}),e.jsx(r,{label:"Success",value:"Thank you for the detailed feedback!",variant:"success",successMessage:"Feedback received!",rows:3}),e.jsx(r,{label:"Disabled",placeholder:"Disabled state",rows:3,disabled:!0})]})},p={render:()=>e.jsxs("div",{className:"w-96 space-y-4",children:[e.jsx(r,{label:"Bio (200 chars)",placeholder:"Tell us about yourself",maxLength:200,showCharCount:!0,rows:4}),e.jsx(r,{label:"Comments (500 chars)",placeholder:"Share your thoughts",maxLength:500,showCharCount:!0,helperText:"Be constructive and respectful",rows:4})]})},m={render:()=>e.jsx("div",{className:"w-96 space-y-4",children:e.jsx(r,{label:"Auto-growing textarea",placeholder:"Type more to see it grow...",autoGrow:!0,helperText:"This textarea grows as you type"})})},h={render:()=>e.jsxs("div",{className:"w-96 space-y-6 p-6 bg-white rounded-lg border border-gray-200",children:[e.jsx("div",{children:e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-4",children:"Feedback Form"})}),e.jsx(r,{label:"What went wrong?",placeholder:"Describe the issue you encountered",required:!0,helperText:"Please be as detailed as possible",rows:4}),e.jsx(r,{label:"How can we improve?",placeholder:"Your suggestions",helperText:"We value your input",rows:3}),e.jsx(r,{label:"Additional comments",placeholder:"Optional",rows:2})]})},x={args:{label:"Your message",placeholder:"Type something...",rows:4,required:!1,disabled:!1}},b={render:()=>e.jsxs("div",{className:"w-96 space-y-6",children:[e.jsx(r,{id:"feedback-input",label:"Feedback",placeholder:"Share your feedback",required:!0,helperText:"Required to submit"}),e.jsx(r,{id:"description-input",label:"Description",placeholder:"Detailed description",maxLength:1e3,showCharCount:!0,helperText:"Maximum 1000 characters"})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0}]}}}};var N,q,k;c.parameters={...c.parameters,docs:{...(N=c.parameters)==null?void 0:N.docs,source:{originalSource:`{
  args: {
    label: 'Comments',
    placeholder: 'Enter your feedback...',
    rows: 4
  }
}`,...(k=(q=c.parameters)==null?void 0:q.docs)==null?void 0:k.source}}};var D,E,R;d.parameters={...d.parameters,docs:{...(D=d.parameters)==null?void 0:D.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-4">
      <TextArea label="Small (2 rows)" placeholder="Compact" rows={2} />
      <TextArea label="Medium (4 rows)" placeholder="Default size" rows={4} />
      <TextArea label="Large (6 rows)" placeholder="Spacious" rows={6} />
    </div>
}`,...(R=(E=d.parameters)==null?void 0:E.docs)==null?void 0:R.source}}};var L,F,V;u.parameters={...u.parameters,docs:{...(L=u.parameters)==null?void 0:L.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-4">
      <TextArea label="Default" placeholder="Default state" rows={3} />
      <TextArea label="Error" value="This message contains inappropriate..." variant="error" errorMessage="Please avoid using inappropriate language" rows={3} />
      <TextArea label="Success" value="Thank you for the detailed feedback!" variant="success" successMessage="Feedback received!" rows={3} />
      <TextArea label="Disabled" placeholder="Disabled state" rows={3} disabled />
    </div>
}`,...(V=(F=u.parameters)==null?void 0:F.docs)==null?void 0:V.source}}};var H,W,M;p.parameters={...p.parameters,docs:{...(H=p.parameters)==null?void 0:H.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-4">
      <TextArea label="Bio (200 chars)" placeholder="Tell us about yourself" maxLength={200} showCharCount rows={4} />
      <TextArea label="Comments (500 chars)" placeholder="Share your thoughts" maxLength={500} showCharCount helperText="Be constructive and respectful" rows={4} />
    </div>
}`,...(M=(W=p.parameters)==null?void 0:W.docs)==null?void 0:M.source}}};var P,z,Y;m.parameters={...m.parameters,docs:{...(P=m.parameters)==null?void 0:P.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-4">
      <TextArea label="Auto-growing textarea" placeholder="Type more to see it grow..." autoGrow helperText="This textarea grows as you type" />
    </div>
}`,...(Y=(z=m.parameters)==null?void 0:z.docs)==null?void 0:Y.source}}};var B,O,_;h.parameters={...h.parameters,docs:{...(B=h.parameters)==null?void 0:B.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-6 p-6 bg-white rounded-lg border border-gray-200">
      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-4">Feedback Form</h3>
      </div>

      <TextArea label="What went wrong?" placeholder="Describe the issue you encountered" required helperText="Please be as detailed as possible" rows={4} />

      <TextArea label="How can we improve?" placeholder="Your suggestions" helperText="We value your input" rows={3} />

      <TextArea label="Additional comments" placeholder="Optional" rows={2} />
    </div>
}`,...(_=(O=h.parameters)==null?void 0:O.docs)==null?void 0:_.source}}};var $,U,G;x.parameters={...x.parameters,docs:{...($=x.parameters)==null?void 0:$.docs,source:{originalSource:`{
  args: {
    label: 'Your message',
    placeholder: 'Type something...',
    rows: 4,
    required: false,
    disabled: false
  }
}`,...(G=(U=x.parameters)==null?void 0:U.docs)==null?void 0:G.source}}};var I,J,K;b.parameters={...b.parameters,docs:{...(I=b.parameters)==null?void 0:I.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-6">
      <TextArea id="feedback-input" label="Feedback" placeholder="Share your feedback" required helperText="Required to submit" />

      <TextArea id="description-input" label="Description" placeholder="Detailed description" maxLength={1000} showCharCount helperText="Maximum 1000 characters" />
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
}`,...(K=(J=b.parameters)==null?void 0:J.docs)==null?void 0:K.source}}};const ue=["Default","Sizes","States","WithCharCount","AutoGrow","RealWorldUseCases","Playground","Accessibility"];export{b as Accessibility,m as AutoGrow,c as Default,x as Playground,h as RealWorldUseCases,d as Sizes,u as States,p as WithCharCount,ue as __namedExportsOrder,de as default};
