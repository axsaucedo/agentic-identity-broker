import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as ge}from"./index-ClcD9ViR.js";import{c as ve,a as fe}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const he=fe("animate-spin",{variants:{size:{xs:"w-3 h-3",sm:"w-4 h-4",md:"w-6 h-6",lg:"w-8 h-8"},variant:{primary:"text-trust-deep",success:"text-success-primary",error:"text-error-primary",warning:"text-warning-primary",info:"text-info-primary",neutral:"text-gray-500",white:"text-white"}},defaultVariants:{size:"md",variant:"primary"}}),s=ge.forwardRef(({size:de,variant:me,className:xe,label:o="Loading...",...oe},pe)=>e.jsxs("svg",{ref:pe,className:ve(he({size:de,variant:me}),xe),xmlns:"http://www.w3.org/2000/svg",fill:"none",viewBox:"0 0 24 24",role:"status","aria-label":o,...oe,children:[e.jsx("circle",{className:"opacity-25",cx:"12",cy:"12",r:"10",stroke:"currentColor",strokeWidth:"4"}),e.jsx("path",{className:"opacity-75",fill:"currentColor",d:"M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"}),e.jsx("title",{children:o})]}));s.displayName="Spinner";s.__docgenInfo={description:`Spinner component for loading states.
Uses SVG animation for smooth rendering across all devices.

@example
\`\`\`tsx
<Spinner />

<Spinner size="sm" variant="success" label="Loading data..." />

<Spinner size="lg" variant="primary" />
\`\`\``,methods:[],displayName:"Spinner",props:{label:{required:!1,tsType:{name:"string"},description:"Accessible label for screen readers",defaultValue:{value:"'Loading...'",computed:!1}}},composes:["Omit","VariantProps"]};const be={title:"Design System/Primitives/Spinner",component:s,parameters:{layout:"centered",docs:{description:{component:"Loading indicator component with multiple sizes and semantic color variants. Uses smooth CSS animations and includes proper ARIA attributes for accessibility."}}},tags:["autodocs"],argTypes:{size:{control:"select",options:["xs","sm","md","lg"],description:"The size of the spinner"},variant:{control:"select",options:["primary","success","error","warning","info","neutral","white"],description:"The semantic color variant"},label:{control:"text",description:"Accessible label for screen readers"}}},a={args:{label:"Loading..."}},r={render:()=>e.jsxs("div",{className:"flex items-center gap-8",children:[e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{size:"xs"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Extra Small"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{size:"sm"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Small"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{size:"md"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Medium"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{size:"lg"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Large"})]})]})},n={render:()=>e.jsxs("div",{className:"flex flex-wrap items-center gap-8",children:[e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{variant:"primary"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Primary"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{variant:"success"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Success"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{variant:"error"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Error"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{variant:"warning"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Warning"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{variant:"info"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Info"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{variant:"neutral"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Neutral"})]})]})},t={render:()=>e.jsxs("div",{className:"flex items-center justify-center gap-8 p-8 bg-trust-deep rounded-lg",children:[e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{variant:"white",size:"sm"}),e.jsx("span",{className:"text-xs text-white",children:"Small"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{variant:"white",size:"md"}),e.jsx("span",{className:"text-xs text-white",children:"Medium"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{variant:"white",size:"lg"}),e.jsx("span",{className:"text-xs text-white",children:"Large"})]})]})},i={render:()=>e.jsxs("div",{className:"flex flex-col gap-6",children:[e.jsxs("div",{className:"flex items-center gap-3",children:[e.jsx(s,{size:"sm"}),e.jsx("span",{className:"text-sm text-gray-700",children:"Loading data..."})]}),e.jsxs("div",{className:"flex items-center gap-3",children:[e.jsx(s,{size:"md",variant:"success"}),e.jsx("span",{className:"text-base text-gray-700",children:"Processing request..."})]}),e.jsxs("div",{className:"flex items-center gap-3",children:[e.jsx(s,{size:"lg",variant:"primary"}),e.jsx("span",{className:"text-lg text-gray-700",children:"Please wait..."})]})]})},l={render:()=>e.jsx("div",{className:"flex items-center justify-center w-96 h-64 bg-gray-50 rounded-lg border border-gray-200",children:e.jsxs("div",{className:"flex flex-col items-center gap-4",children:[e.jsx(s,{size:"lg"}),e.jsx("p",{className:"text-sm text-gray-600",children:"Loading your data..."})]})})},c={render:()=>e.jsxs("div",{className:"flex flex-col gap-6 p-6",children:[e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Button Loading State"}),e.jsxs("button",{className:"inline-flex items-center gap-2 px-4 py-2 bg-trust-deep text-white rounded-md",disabled:!0,children:[e.jsx(s,{size:"sm",variant:"white"}),"Processing..."]})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Card Loading State"}),e.jsxs("div",{className:"p-4 bg-white border border-gray-200 rounded-lg",children:[e.jsxs("div",{className:"flex items-center gap-2 mb-3",children:[e.jsx(s,{size:"xs"}),e.jsx("h4",{className:"text-sm font-medium text-gray-900",children:"Fetching updates..."})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("div",{className:"h-4 bg-gray-100 rounded animate-pulse"}),e.jsx("div",{className:"h-4 bg-gray-100 rounded animate-pulse w-5/6"}),e.jsx("div",{className:"h-4 bg-gray-100 rounded animate-pulse w-4/6"})]})]})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"List Item Loading"}),e.jsxs("div",{className:"space-y-2",children:[e.jsxs("div",{className:"flex items-center gap-3 p-3 bg-white border border-gray-200 rounded-lg",children:[e.jsx(s,{size:"xs",variant:"info"}),e.jsx("span",{className:"text-sm text-gray-700",children:"Syncing permissions..."})]}),e.jsxs("div",{className:"flex items-center gap-3 p-3 bg-white border border-gray-200 rounded-lg",children:[e.jsx(s,{size:"xs",variant:"success"}),e.jsx("span",{className:"text-sm text-gray-700",children:"Updating grants..."})]})]})]})]})},d={render:()=>e.jsxs("div",{className:"flex flex-wrap gap-4",children:[e.jsxs("div",{className:"flex items-center gap-2 px-3 py-2 bg-blue-50 rounded-md",children:[e.jsx(s,{size:"xs",variant:"info"}),e.jsx("span",{className:"text-sm text-blue-900",children:"Fetching..."})]}),e.jsxs("div",{className:"flex items-center gap-2 px-3 py-2 bg-green-50 rounded-md",children:[e.jsx(s,{size:"xs",variant:"success"}),e.jsx("span",{className:"text-sm text-green-900",children:"Saving..."})]}),e.jsxs("div",{className:"flex items-center gap-2 px-3 py-2 bg-amber-50 rounded-md",children:[e.jsx(s,{size:"xs",variant:"warning"}),e.jsx("span",{className:"text-sm text-amber-900",children:"Processing..."})]}),e.jsxs("div",{className:"flex items-center gap-2 px-3 py-2 bg-red-50 rounded-md",children:[e.jsx(s,{size:"xs",variant:"error"}),e.jsx("span",{className:"text-sm text-red-900",children:"Retrying..."})]})]})},m={args:{size:"md",variant:"primary",label:"Loading..."}},x={render:()=>e.jsxs("div",{className:"flex flex-col gap-6",children:[e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Spinners with descriptive labels"}),e.jsxs("div",{className:"flex flex-wrap gap-4",children:[e.jsx(s,{label:"Loading user profile data"}),e.jsx(s,{variant:"success",label:"Saving changes to the database"}),e.jsx(s,{variant:"error",label:"Retrying failed request"})]})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Live region announcement (for dynamic loading)"}),e.jsx("div",{className:"p-4 bg-gray-50 rounded-lg",children:e.jsxs("div",{className:"flex items-center gap-3",children:[e.jsx(s,{size:"sm"}),e.jsx("div",{"aria-live":"polite","aria-atomic":"true",children:e.jsx("p",{className:"text-sm text-gray-700",children:"Loading... Please wait while we fetch your data."})})]})})]})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0}]}}}};var p,g,v,f,h;a.parameters={...a.parameters,docs:{...(p=a.parameters)==null?void 0:p.docs,source:{originalSource:`{
  args: {
    label: 'Loading...'
  }
}`,...(v=(g=a.parameters)==null?void 0:g.docs)==null?void 0:v.source},description:{story:"Default spinner with primary variant",...(h=(f=a.parameters)==null?void 0:f.docs)==null?void 0:h.description}}};var N,u,y,j,b;r.parameters={...r.parameters,docs:{...(N=r.parameters)==null?void 0:N.docs,source:{originalSource:`{
  render: () => <div className="flex items-center gap-8">
      <div className="flex flex-col items-center gap-2">
        <Spinner size="xs" />
        <span className="text-xs text-gray-600">Extra Small</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Spinner size="sm" />
        <span className="text-xs text-gray-600">Small</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Spinner size="md" />
        <span className="text-xs text-gray-600">Medium</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Spinner size="lg" />
        <span className="text-xs text-gray-600">Large</span>
      </div>
    </div>
}`,...(y=(u=r.parameters)==null?void 0:u.docs)==null?void 0:y.source},description:{story:"All spinner sizes from extra small to large",...(b=(j=r.parameters)==null?void 0:j.docs)==null?void 0:b.description}}};var S,w,z,L,P;n.parameters={...n.parameters,docs:{...(S=n.parameters)==null?void 0:S.docs,source:{originalSource:`{
  render: () => <div className="flex flex-wrap items-center gap-8">
      <div className="flex flex-col items-center gap-2">
        <Spinner variant="primary" />
        <span className="text-xs text-gray-600">Primary</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Spinner variant="success" />
        <span className="text-xs text-gray-600">Success</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Spinner variant="error" />
        <span className="text-xs text-gray-600">Error</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Spinner variant="warning" />
        <span className="text-xs text-gray-600">Warning</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Spinner variant="info" />
        <span className="text-xs text-gray-600">Info</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Spinner variant="neutral" />
        <span className="text-xs text-gray-600">Neutral</span>
      </div>
    </div>
}`,...(z=(w=n.parameters)==null?void 0:w.docs)==null?void 0:z.source},description:{story:"All semantic color variants",...(P=(L=n.parameters)==null?void 0:L.docs)==null?void 0:P.description}}};var I,C,A,R,k;t.parameters={...t.parameters,docs:{...(I=t.parameters)==null?void 0:I.docs,source:{originalSource:`{
  render: () => <div className="flex items-center justify-center gap-8 p-8 bg-trust-deep rounded-lg">
      <div className="flex flex-col items-center gap-2">
        <Spinner variant="white" size="sm" />
        <span className="text-xs text-white">Small</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Spinner variant="white" size="md" />
        <span className="text-xs text-white">Medium</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Spinner variant="white" size="lg" />
        <span className="text-xs text-white">Large</span>
      </div>
    </div>
}`,...(A=(C=t.parameters)==null?void 0:C.docs)==null?void 0:A.source},description:{story:"White spinner for dark backgrounds",...(k=(R=t.parameters)==null?void 0:R.docs)==null?void 0:k.description}}};var V,D,E,T,U;i.parameters={...i.parameters,docs:{...(V=i.parameters)==null?void 0:V.docs,source:{originalSource:`{
  render: () => <div className="flex flex-col gap-6">
      <div className="flex items-center gap-3">
        <Spinner size="sm" />
        <span className="text-sm text-gray-700">Loading data...</span>
      </div>
      <div className="flex items-center gap-3">
        <Spinner size="md" variant="success" />
        <span className="text-base text-gray-700">Processing request...</span>
      </div>
      <div className="flex items-center gap-3">
        <Spinner size="lg" variant="primary" />
        <span className="text-lg text-gray-700">Please wait...</span>
      </div>
    </div>
}`,...(E=(D=i.parameters)==null?void 0:D.docs)==null?void 0:E.source},description:{story:"Spinner with text labels",...(U=(T=i.parameters)==null?void 0:T.docs)==null?void 0:U.description}}};var W,q,B,M,F;l.parameters={...l.parameters,docs:{...(W=l.parameters)==null?void 0:W.docs,source:{originalSource:`{
  render: () => <div className="flex items-center justify-center w-96 h-64 bg-gray-50 rounded-lg border border-gray-200">
      <div className="flex flex-col items-center gap-4">
        <Spinner size="lg" />
        <p className="text-sm text-gray-600">Loading your data...</p>
      </div>
    </div>
}`,...(B=(q=l.parameters)==null?void 0:q.docs)==null?void 0:B.source},description:{story:"Centered spinner for full-page loading",...(F=(M=l.parameters)==null?void 0:M.docs)==null?void 0:F.description}}};var O,_,G,H,J;c.parameters={...c.parameters,docs:{...(O=c.parameters)==null?void 0:O.docs,source:{originalSource:`{
  render: () => <div className="flex flex-col gap-6 p-6">
      {/* In a button */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">Button Loading State</h3>
        <button className="inline-flex items-center gap-2 px-4 py-2 bg-trust-deep text-white rounded-md" disabled>
          <Spinner size="sm" variant="white" />
          Processing...
        </button>
      </div>

      {/* In a card header */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">Card Loading State</h3>
        <div className="p-4 bg-white border border-gray-200 rounded-lg">
          <div className="flex items-center gap-2 mb-3">
            <Spinner size="xs" />
            <h4 className="text-sm font-medium text-gray-900">Fetching updates...</h4>
          </div>
          <div className="space-y-2">
            <div className="h-4 bg-gray-100 rounded animate-pulse" />
            <div className="h-4 bg-gray-100 rounded animate-pulse w-5/6" />
            <div className="h-4 bg-gray-100 rounded animate-pulse w-4/6" />
          </div>
        </div>
      </div>

      {/* In a list item */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">List Item Loading</h3>
        <div className="space-y-2">
          <div className="flex items-center gap-3 p-3 bg-white border border-gray-200 rounded-lg">
            <Spinner size="xs" variant="info" />
            <span className="text-sm text-gray-700">Syncing permissions...</span>
          </div>
          <div className="flex items-center gap-3 p-3 bg-white border border-gray-200 rounded-lg">
            <Spinner size="xs" variant="success" />
            <span className="text-sm text-gray-700">Updating grants...</span>
          </div>
        </div>
      </div>
    </div>
}`,...(G=(_=c.parameters)==null?void 0:_.docs)==null?void 0:G.source},description:{story:"Inline spinners in different contexts",...(J=(H=c.parameters)==null?void 0:H.docs)==null?void 0:J.description}}};var K,Q,X,Y,Z;d.parameters={...d.parameters,docs:{...(K=d.parameters)==null?void 0:K.docs,source:{originalSource:`{
  render: () => <div className="flex flex-wrap gap-4">
      <div className="flex items-center gap-2 px-3 py-2 bg-blue-50 rounded-md">
        <Spinner size="xs" variant="info" />
        <span className="text-sm text-blue-900">Fetching...</span>
      </div>
      <div className="flex items-center gap-2 px-3 py-2 bg-green-50 rounded-md">
        <Spinner size="xs" variant="success" />
        <span className="text-sm text-green-900">Saving...</span>
      </div>
      <div className="flex items-center gap-2 px-3 py-2 bg-amber-50 rounded-md">
        <Spinner size="xs" variant="warning" />
        <span className="text-sm text-amber-900">Processing...</span>
      </div>
      <div className="flex items-center gap-2 px-3 py-2 bg-red-50 rounded-md">
        <Spinner size="xs" variant="error" />
        <span className="text-sm text-red-900">Retrying...</span>
      </div>
    </div>
}`,...(X=(Q=d.parameters)==null?void 0:Q.docs)==null?void 0:X.source},description:{story:"Different loading states for operations",...(Z=(Y=d.parameters)==null?void 0:Y.docs)==null?void 0:Z.description}}};var $,ee,se,ae,re;m.parameters={...m.parameters,docs:{...($=m.parameters)==null?void 0:$.docs,source:{originalSource:`{
  args: {
    size: 'md',
    variant: 'primary',
    label: 'Loading...'
  }
}`,...(se=(ee=m.parameters)==null?void 0:ee.docs)==null?void 0:se.source},description:{story:"Interactive playground for testing all combinations",...(re=(ae=m.parameters)==null?void 0:ae.docs)==null?void 0:re.description}}};var ne,te,ie,le,ce;x.parameters={...x.parameters,docs:{...(ne=x.parameters)==null?void 0:ne.docs,source:{originalSource:`{
  render: () => <div className="flex flex-col gap-6">
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">
          Spinners with descriptive labels
        </h3>
        <div className="flex flex-wrap gap-4">
          <Spinner label="Loading user profile data" />
          <Spinner variant="success" label="Saving changes to the database" />
          <Spinner variant="error" label="Retrying failed request" />
        </div>
      </div>

      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">
          Live region announcement (for dynamic loading)
        </h3>
        <div className="p-4 bg-gray-50 rounded-lg">
          <div className="flex items-center gap-3">
            <Spinner size="sm" />
            <div aria-live="polite" aria-atomic="true">
              <p className="text-sm text-gray-700">
                Loading... Please wait while we fetch your data.
              </p>
            </div>
          </div>
        </div>
      </div>
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
}`,...(ie=(te=x.parameters)==null?void 0:te.docs)==null?void 0:ie.source},description:{story:"Accessibility test - screen reader friendly",...(ce=(le=x.parameters)==null?void 0:le.docs)==null?void 0:ce.description}}};const Se=["Default","Sizes","Variants","OnDarkBackground","WithText","CenteredLoading","InlineUseCases","LoadingStates","Playground","Accessibility"];export{x as Accessibility,l as CenteredLoading,a as Default,c as InlineUseCases,d as LoadingStates,t as OnDarkBackground,m as Playground,r as Sizes,n as Variants,i as WithText,Se as __namedExportsOrder,be as default};
