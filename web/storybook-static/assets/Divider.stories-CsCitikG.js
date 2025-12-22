import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as ze}from"./index-ClcD9ViR.js";import{c as r,a as we}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const N=we("",{variants:{orientation:{horizontal:"w-full",vertical:"h-full"},variant:{default:"",subtle:"",muted:""}},defaultVariants:{orientation:"horizontal",variant:"default"}}),g=we("",{variants:{orientation:{horizontal:"h-px w-full",vertical:"w-px h-full"},style:{solid:"",dashed:""},variant:{default:"bg-gray-300",subtle:"bg-gray-200",muted:"bg-gray-100"}},compoundVariants:[{style:"dashed",orientation:"horizontal",className:"border-t border-dashed border-gray-300"},{style:"dashed",orientation:"vertical",className:"border-l border-dashed border-gray-300"}],defaultVariants:{orientation:"horizontal",style:"solid",variant:"default"}}),s=ze.forwardRef(({orientation:a="horizontal",variant:i="default",className:Se,style:t="solid",label:h,labelSpacing:De="md",...f},b)=>{const Oe={sm:"px-2",md:"px-3",lg:"px-4"}[De];return a==="vertical"?e.jsx("div",{ref:b,className:r(N({orientation:a,variant:i}),"flex items-center"),role:"separator","aria-orientation":"vertical",...f,children:e.jsx("div",{className:r(g({orientation:a,style:t,variant:i})),style:t==="solid"?{}:void 0})}):h?e.jsxs("div",{ref:b,className:r(N({orientation:a,variant:i}),"flex items-center gap-0",Se),role:"separator","aria-label":typeof h=="string"?h:void 0,...f,children:[e.jsx("div",{className:r(g({orientation:a,style:t,variant:i}),"flex-1"),style:t==="solid"?{}:void 0}),e.jsx("span",{className:r(Oe,"text-sm text-gray-600 font-medium whitespace-nowrap"),children:h}),e.jsx("div",{className:r(g({orientation:a,style:t,variant:i}),"flex-1"),style:t==="solid"?{}:void 0})]}):e.jsx("div",{ref:b,className:r(N({orientation:a,variant:i}),"flex items-center"),role:"separator","aria-orientation":"horizontal",...f,children:e.jsx("div",{className:r(g({orientation:a,style:t,variant:i})),style:t==="solid"?{}:void 0})})});s.displayName="Divider";s.__docgenInfo={description:`Divider component for visual content separation.
Can display horizontally or vertically with optional centered text.

@example
\`\`\`tsx
<Divider />

<Divider label="Or" />

<Divider variant="subtle" />

<Divider orientation="vertical" />

<Divider style="dashed" labelSpacing="lg" label="Section Break" />
\`\`\``,methods:[],displayName:"Divider",props:{style:{required:!1,tsType:{name:"union",raw:"'solid' | 'dashed'",elements:[{name:"literal",value:"'solid'"},{name:"literal",value:"'dashed'"}]},description:"Style of the divider: solid or dashed",defaultValue:{value:"'solid'",computed:!1}},label:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Optional text label to display in the middle (horizontal only)"},labelSpacing:{required:!1,tsType:{name:"union",raw:"'sm' | 'md' | 'lg'",elements:[{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"},{name:"literal",value:"'lg'"}]},description:"Spacing around the label",defaultValue:{value:"'md'",computed:!1}},orientation:{defaultValue:{value:"'horizontal'",computed:!1},required:!1},variant:{defaultValue:{value:"'default'",computed:!1},required:!1}},composes:["VariantProps"]};const Ae={title:"Design System/Primitives/Divider",component:s,parameters:{layout:"centered",docs:{description:{component:"Divider component for visual content separation. Supports horizontal and vertical orientations with optional centered text labels."}}},tags:["autodocs"],argTypes:{orientation:{control:"select",options:["horizontal","vertical"],description:"The orientation of the divider"},variant:{control:"select",options:["default","subtle","muted"],description:"The visual prominence variant"},style:{control:"select",options:["solid","dashed"],description:"The line style"},label:{control:"text",description:"Optional text label to display in the center (horizontal only)"},labelSpacing:{control:"select",options:["sm","md","lg"],description:"Spacing around the label"}}},n={args:{}},l={render:()=>e.jsxs("div",{className:"w-full max-w-md space-y-8",children:[e.jsxs("div",{className:"space-y-2",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-700",children:"Default"}),e.jsx(s,{variant:"default"})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-700",children:"Subtle"}),e.jsx(s,{variant:"subtle"})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-700",children:"Muted"}),e.jsx(s,{variant:"muted"})]})]})},d={render:()=>e.jsxs("div",{className:"w-full max-w-md space-y-8",children:[e.jsxs("div",{className:"space-y-2",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-700",children:"Solid"}),e.jsx(s,{style:"solid"})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-700",children:"Dashed"}),e.jsx(s,{style:"dashed"})]})]})},o={render:()=>e.jsxs("div",{className:"w-full max-w-md space-y-8",children:[e.jsx(s,{label:"Or"}),e.jsx(s,{label:"Section Break"}),e.jsx(s,{label:"More Options",style:"dashed"}),e.jsx(s,{label:"End of Content",variant:"subtle"})]})},c={render:()=>e.jsxs("div",{className:"w-full max-w-md space-y-8",children:[e.jsxs("div",{className:"space-y-2",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-700",children:"Small Spacing"}),e.jsx(s,{label:"Or",labelSpacing:"sm"})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-700",children:"Medium Spacing"}),e.jsx(s,{label:"Or",labelSpacing:"md"})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-700",children:"Large Spacing"}),e.jsx(s,{label:"Or",labelSpacing:"lg"})]})]})},m={render:()=>e.jsxs("div",{className:"flex items-center gap-4 h-32",children:[e.jsx("div",{className:"flex-1 p-4 bg-gray-50 rounded-lg",children:e.jsx("p",{className:"text-sm text-gray-600",children:"Left Section"})}),e.jsx(s,{orientation:"vertical"}),e.jsx("div",{className:"flex-1 p-4 bg-gray-50 rounded-lg",children:e.jsx("p",{className:"text-sm text-gray-600",children:"Right Section"})})]})},p={render:()=>e.jsxs("div",{className:"w-full max-w-md space-y-6 p-6 bg-white rounded-lg border border-gray-200",children:[e.jsxs("div",{className:"space-y-4",children:[e.jsxs("div",{children:[e.jsx("label",{className:"text-sm font-medium text-gray-700",children:"Email"}),e.jsx("input",{type:"email",className:"mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm",placeholder:"you@example.com"})]}),e.jsxs("div",{children:[e.jsx("label",{className:"text-sm font-medium text-gray-700",children:"Password"}),e.jsx("input",{type:"password",className:"mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm",placeholder:"••••••••"})]})]}),e.jsx(s,{label:"Or"}),e.jsx("button",{className:"w-full py-2 px-3 bg-gray-100 text-gray-900 rounded-md text-sm font-medium hover:bg-gray-200 transition",children:"Sign in with Google"})]})},x={render:()=>e.jsxs("div",{className:"w-full max-w-lg p-6 space-y-6",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-2",children:"Overview"}),e.jsx("p",{className:"text-sm text-gray-600",children:"This is the overview section with important information about your account."})]}),e.jsx(s,{variant:"subtle"}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-2",children:"Settings"}),e.jsx("p",{className:"text-sm text-gray-600",children:"Configure your preferences and notification settings here."})]}),e.jsx(s,{variant:"subtle"}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-2",children:"Security"}),e.jsx("p",{className:"text-sm text-gray-600",children:"Manage your password and connected devices."})]}),e.jsx(s,{variant:"subtle"}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-2",children:"Privacy"}),e.jsx("p",{className:"text-sm text-gray-600",children:"Control how your data is shared and used."})]})]})},v={render:()=>e.jsxs("div",{className:"w-full max-w-md p-6 space-y-4",children:[e.jsxs("div",{className:"flex gap-3",children:[e.jsx("div",{className:"w-8 h-8 rounded-full bg-success-primary text-white flex items-center justify-center text-sm font-bold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h4",{className:"text-sm font-medium text-gray-900",children:"Step 1: Review"}),e.jsx("p",{className:"text-xs text-gray-600",children:"Completed"})]})]}),e.jsx("div",{className:"ml-4",children:e.jsx(s,{orientation:"vertical",style:"dashed"})}),e.jsxs("div",{className:"flex gap-3",children:[e.jsx("div",{className:"w-8 h-8 rounded-full bg-info-primary text-white flex items-center justify-center text-sm font-bold",children:"2"}),e.jsxs("div",{children:[e.jsx("h4",{className:"text-sm font-medium text-gray-900",children:"Step 2: Approve"}),e.jsx("p",{className:"text-xs text-gray-600",children:"In progress"})]})]}),e.jsx("div",{className:"ml-4",children:e.jsx(s,{orientation:"vertical",style:"dashed"})}),e.jsxs("div",{className:"flex gap-3",children:[e.jsx("div",{className:"w-8 h-8 rounded-full bg-gray-300 text-gray-600 flex items-center justify-center text-sm font-bold",children:"3"}),e.jsxs("div",{children:[e.jsx("h4",{className:"text-sm font-medium text-gray-900",children:"Step 3: Complete"}),e.jsx("p",{className:"text-xs text-gray-600",children:"Pending"})]})]})]})},u={args:{orientation:"horizontal",variant:"default",style:"solid",label:"Or",labelSpacing:"md"}},y={render:()=>e.jsxs("div",{className:"w-full max-w-md space-y-6",children:[e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Dividers with semantic meaning"}),e.jsxs("div",{className:"space-y-4",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm text-gray-600 mb-2",children:"Account Information"}),e.jsx(s,{})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm text-gray-600 mb-2",children:"Security Settings"}),e.jsx(s,{})]})]})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Labeled dividers"}),e.jsx("div",{className:"space-y-4",children:e.jsxs("div",{children:[e.jsx("p",{className:"text-sm text-gray-600 mb-2",children:"Alternative Sign-In"}),e.jsx(s,{label:"Or continue with"})]})})]})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0}]}}}};var j,w,S,D,O;n.parameters={...n.parameters,docs:{...(j=n.parameters)==null?void 0:j.docs,source:{originalSource:`{
  args: {}
}`,...(S=(w=n.parameters)==null?void 0:w.docs)==null?void 0:S.source},description:{story:"Default horizontal divider",...(O=(D=n.parameters)==null?void 0:D.docs)==null?void 0:O.description}}};var z,C,V,R,L;l.parameters={...l.parameters,docs:{...(z=l.parameters)==null?void 0:z.docs,source:{originalSource:`{
  render: () => <div className="w-full max-w-md space-y-8">
      <div className="space-y-2">
        <h4 className="text-sm font-semibold text-gray-700">Default</h4>
        <Divider variant="default" />
      </div>

      <div className="space-y-2">
        <h4 className="text-sm font-semibold text-gray-700">Subtle</h4>
        <Divider variant="subtle" />
      </div>

      <div className="space-y-2">
        <h4 className="text-sm font-semibold text-gray-700">Muted</h4>
        <Divider variant="muted" />
      </div>
    </div>
}`,...(V=(C=l.parameters)==null?void 0:C.docs)==null?void 0:V.source},description:{story:"All divider variants from prominent to subtle",...(L=(R=l.parameters)==null?void 0:R.docs)==null?void 0:L.description}}};var A,T,P,I,M;d.parameters={...d.parameters,docs:{...(A=d.parameters)==null?void 0:A.docs,source:{originalSource:`{
  render: () => <div className="w-full max-w-md space-y-8">
      <div className="space-y-2">
        <h4 className="text-sm font-semibold text-gray-700">Solid</h4>
        <Divider style="solid" />
      </div>

      <div className="space-y-2">
        <h4 className="text-sm font-semibold text-gray-700">Dashed</h4>
        <Divider style="dashed" />
      </div>
    </div>
}`,...(P=(T=d.parameters)==null?void 0:T.docs)==null?void 0:P.source},description:{story:"Divider line styles: solid and dashed",...(M=(I=d.parameters)==null?void 0:I.docs)==null?void 0:M.description}}};var k,E,q,_,B;o.parameters={...o.parameters,docs:{...(k=o.parameters)==null?void 0:k.docs,source:{originalSource:`{
  render: () => <div className="w-full max-w-md space-y-8">
      <Divider label="Or" />
      <Divider label="Section Break" />
      <Divider label="More Options" style="dashed" />
      <Divider label="End of Content" variant="subtle" />
    </div>
}`,...(q=(E=o.parameters)==null?void 0:E.docs)==null?void 0:q.source},description:{story:"Divider with centered label",...(B=(_=o.parameters)==null?void 0:_.docs)==null?void 0:B.description}}};var G,U,W,F,H;c.parameters={...c.parameters,docs:{...(G=c.parameters)==null?void 0:G.docs,source:{originalSource:`{
  render: () => <div className="w-full max-w-md space-y-8">
      <div className="space-y-2">
        <h4 className="text-sm font-semibold text-gray-700">Small Spacing</h4>
        <Divider label="Or" labelSpacing="sm" />
      </div>

      <div className="space-y-2">
        <h4 className="text-sm font-semibold text-gray-700">Medium Spacing</h4>
        <Divider label="Or" labelSpacing="md" />
      </div>

      <div className="space-y-2">
        <h4 className="text-sm font-semibold text-gray-700">Large Spacing</h4>
        <Divider label="Or" labelSpacing="lg" />
      </div>
    </div>
}`,...(W=(U=c.parameters)==null?void 0:U.docs)==null?void 0:W.source},description:{story:"Label spacing variations",...(H=(F=c.parameters)==null?void 0:F.docs)==null?void 0:H.description}}};var J,K,Q,X,Y;m.parameters={...m.parameters,docs:{...(J=m.parameters)==null?void 0:J.docs,source:{originalSource:`{
  render: () => <div className="flex items-center gap-4 h-32">
      <div className="flex-1 p-4 bg-gray-50 rounded-lg">
        <p className="text-sm text-gray-600">Left Section</p>
      </div>
      <Divider orientation="vertical" />
      <div className="flex-1 p-4 bg-gray-50 rounded-lg">
        <p className="text-sm text-gray-600">Right Section</p>
      </div>
    </div>
}`,...(Q=(K=m.parameters)==null?void 0:K.docs)==null?void 0:Q.source},description:{story:"Vertical divider for side-by-side layouts",...(Y=(X=m.parameters)==null?void 0:X.docs)==null?void 0:Y.description}}};var Z,$,ee,se,ae;p.parameters={...p.parameters,docs:{...(Z=p.parameters)==null?void 0:Z.docs,source:{originalSource:`{
  render: () => <div className="w-full max-w-md space-y-6 p-6 bg-white rounded-lg border border-gray-200">
      {/* Form sections */}
      <div className="space-y-4">
        <div>
          <label className="text-sm font-medium text-gray-700">Email</label>
          <input type="email" className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" placeholder="you@example.com" />
        </div>
        <div>
          <label className="text-sm font-medium text-gray-700">Password</label>
          <input type="password" className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" placeholder="••••••••" />
        </div>
      </div>

      <Divider label="Or" />

      {/* Alternative method */}
      <button className="w-full py-2 px-3 bg-gray-100 text-gray-900 rounded-md text-sm font-medium hover:bg-gray-200 transition">
        Sign in with Google
      </button>
    </div>
}`,...(ee=($=p.parameters)==null?void 0:$.docs)==null?void 0:ee.source},description:{story:"Real-world use cases",...(ae=(se=p.parameters)==null?void 0:se.docs)==null?void 0:ae.description}}};var te,re,ie,ne,le;x.parameters={...x.parameters,docs:{...(te=x.parameters)==null?void 0:te.docs,source:{originalSource:`{
  render: () => <div className="w-full max-w-lg p-6 space-y-6">
      {/* Section 1 */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-2">Overview</h3>
        <p className="text-sm text-gray-600">
          This is the overview section with important information about your account.
        </p>
      </div>

      <Divider variant="subtle" />

      {/* Section 2 */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-2">Settings</h3>
        <p className="text-sm text-gray-600">
          Configure your preferences and notification settings here.
        </p>
      </div>

      <Divider variant="subtle" />

      {/* Section 3 */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-2">Security</h3>
        <p className="text-sm text-gray-600">
          Manage your password and connected devices.
        </p>
      </div>

      <Divider variant="subtle" />

      {/* Section 4 */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-2">Privacy</h3>
        <p className="text-sm text-gray-600">
          Control how your data is shared and used.
        </p>
      </div>
    </div>
}`,...(ie=(re=x.parameters)==null?void 0:re.docs)==null?void 0:ie.source},description:{story:"Content sections separated by dividers",...(le=(ne=x.parameters)==null?void 0:ne.docs)==null?void 0:le.description}}};var de,oe,ce,me,pe;v.parameters={...v.parameters,docs:{...(de=v.parameters)==null?void 0:de.docs,source:{originalSource:`{
  render: () => <div className="w-full max-w-md p-6 space-y-4">
      <div className="flex gap-3">
        <div className="w-8 h-8 rounded-full bg-success-primary text-white flex items-center justify-center text-sm font-bold">
          ✓
        </div>
        <div>
          <h4 className="text-sm font-medium text-gray-900">Step 1: Review</h4>
          <p className="text-xs text-gray-600">Completed</p>
        </div>
      </div>

      <div className="ml-4">
        <Divider orientation="vertical" style="dashed" />
      </div>

      <div className="flex gap-3">
        <div className="w-8 h-8 rounded-full bg-info-primary text-white flex items-center justify-center text-sm font-bold">
          2
        </div>
        <div>
          <h4 className="text-sm font-medium text-gray-900">Step 2: Approve</h4>
          <p className="text-xs text-gray-600">In progress</p>
        </div>
      </div>

      <div className="ml-4">
        <Divider orientation="vertical" style="dashed" />
      </div>

      <div className="flex gap-3">
        <div className="w-8 h-8 rounded-full bg-gray-300 text-gray-600 flex items-center justify-center text-sm font-bold">
          3
        </div>
        <div>
          <h4 className="text-sm font-medium text-gray-900">Step 3: Complete</h4>
          <p className="text-xs text-gray-600">Pending</p>
        </div>
      </div>
    </div>
}`,...(ce=(oe=v.parameters)==null?void 0:oe.docs)==null?void 0:ce.source},description:{story:"Steps or timeline with dividers",...(pe=(me=v.parameters)==null?void 0:me.docs)==null?void 0:pe.description}}};var xe,ve,ue,ye,he;u.parameters={...u.parameters,docs:{...(xe=u.parameters)==null?void 0:xe.docs,source:{originalSource:`{
  args: {
    orientation: 'horizontal',
    variant: 'default',
    style: 'solid',
    label: 'Or',
    labelSpacing: 'md'
  }
}`,...(ue=(ve=u.parameters)==null?void 0:ve.docs)==null?void 0:ue.source},description:{story:"Interactive playground for testing all combinations",...(he=(ye=u.parameters)==null?void 0:ye.docs)==null?void 0:he.description}}};var ge,fe,be,Ne,je;y.parameters={...y.parameters,docs:{...(ge=y.parameters)==null?void 0:ge.docs,source:{originalSource:`{
  render: () => <div className="w-full max-w-md space-y-6">
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">
          Dividers with semantic meaning
        </h3>
        <div className="space-y-4">
          <div>
            <p className="text-sm text-gray-600 mb-2">Account Information</p>
            <Divider />
          </div>
          <div>
            <p className="text-sm text-gray-600 mb-2">Security Settings</p>
            <Divider />
          </div>
        </div>
      </div>

      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">
          Labeled dividers
        </h3>
        <div className="space-y-4">
          <div>
            <p className="text-sm text-gray-600 mb-2">Alternative Sign-In</p>
            <Divider label="Or continue with" />
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
}`,...(be=(fe=y.parameters)==null?void 0:fe.docs)==null?void 0:be.source},description:{story:"Accessibility test - screen reader friendly",...(je=(Ne=y.parameters)==null?void 0:Ne.docs)==null?void 0:je.description}}};const Te=["Default","Variants","Styles","WithLabel","LabelSpacing","Vertical","UseCases","ContentSeparation","Timeline","Playground","Accessibility"];export{y as Accessibility,x as ContentSeparation,n as Default,c as LabelSpacing,u as Playground,d as Styles,v as Timeline,p as UseCases,l as Variants,m as Vertical,o as WithLabel,Te as __namedExportsOrder,Ae as default};
