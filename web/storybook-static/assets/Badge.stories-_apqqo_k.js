import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as je}from"./index-ClcD9ViR.js";import{c as g,a as ye}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const Ne=ye("inline-flex items-center justify-center font-medium border transition-colors",{variants:{variant:{success:"bg-success-light text-success-dark border-success-primary/20",error:"bg-error-light text-error-dark border-error-primary/20",warning:"bg-warning-light text-warning-dark border-warning-primary/20",info:"bg-info-light text-info-dark border-info-primary/20",neutral:"bg-gray-100 text-gray-700 border-gray-300",primary:"bg-trust-light text-trust-deep border-trust/20"},size:{sm:"px-2 py-0.5 text-xs gap-1",md:"px-2.5 py-1 text-sm gap-1.5",lg:"px-3 py-1.5 text-base gap-2"},shape:{rounded:"rounded-md",pill:"rounded-full"}},defaultVariants:{variant:"neutral",size:"md",shape:"pill"}}),a=je.forwardRef(({variant:ge,size:r,shape:he,className:ue,children:ve,iconBefore:h,iconAfter:u,showDot:xe=!1,...fe},Be)=>{const v=r==="sm"?"w-3 h-3":r==="lg"?"w-4 h-4":"w-3.5 h-3.5",we=r==="sm"?"w-1.5 h-1.5":r==="lg"?"w-2.5 h-2.5":"w-2 h-2";return e.jsxs("span",{ref:Be,className:g(Ne({variant:ge,size:r,shape:he}),ue),...fe,children:[xe&&e.jsx("span",{className:g("rounded-full bg-current",we),"aria-hidden":"true"}),h&&e.jsx("span",{className:g("flex items-center",v),"aria-hidden":"true",children:h}),ve,u&&e.jsx("span",{className:g("flex items-center",v),"aria-hidden":"true",children:u})]})});a.displayName="Badge";a.__docgenInfo={description:`Badge component for displaying status, labels, or counts.
Supports icons, dots, and multiple semantic variants.

@example
\`\`\`tsx
<Badge variant="success">Active</Badge>

<Badge variant="warning" showDot>
  Pending
</Badge>

<Badge variant="info" iconBefore={<InfoIcon />}>
  New
</Badge>

<Badge variant="neutral" size="sm" shape="rounded">
  Beta
</Badge>
\`\`\``,methods:[],displayName:"Badge",props:{iconBefore:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Icon to display before children"},iconAfter:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Icon to display after children"},showDot:{required:!1,tsType:{name:"boolean"},description:"Show a dot indicator before the content",defaultValue:{value:"false",computed:!1}}},composes:["VariantProps"]};const Le={title:"Design System/Primitives/Badge",component:a,parameters:{layout:"centered",docs:{description:{component:"Versatile badge component for status indicators, labels, and counts. Supports multiple variants, sizes, icons, and dot indicators."}}},tags:["autodocs"],argTypes:{variant:{control:"select",options:["success","error","warning","info","neutral","primary"],description:"The semantic variant of the badge"},size:{control:"select",options:["sm","md","lg"],description:"The size of the badge"},shape:{control:"select",options:["rounded","pill"],description:"The shape of the badge corners"},showDot:{control:"boolean",description:"Show a dot indicator before the content"},children:{control:"text",description:"The content of the badge"}}},s={args:{children:"Badge"}},n={render:()=>e.jsxs("div",{className:"flex flex-wrap gap-4",children:[e.jsx(a,{variant:"success",children:"Success"}),e.jsx(a,{variant:"error",children:"Error"}),e.jsx(a,{variant:"warning",children:"Warning"}),e.jsx(a,{variant:"info",children:"Info"}),e.jsx(a,{variant:"neutral",children:"Neutral"}),e.jsx(a,{variant:"primary",children:"Primary"})]})},i={render:()=>e.jsxs("div",{className:"flex flex-wrap items-center gap-4",children:[e.jsx(a,{variant:"primary",size:"sm",children:"Small"}),e.jsx(a,{variant:"primary",size:"md",children:"Medium"}),e.jsx(a,{variant:"primary",size:"lg",children:"Large"})]})},t={render:()=>e.jsxs("div",{className:"flex flex-wrap gap-4",children:[e.jsx(a,{variant:"primary",shape:"pill",children:"Pill Shape"}),e.jsx(a,{variant:"primary",shape:"rounded",children:"Rounded Shape"})]})},o={render:()=>e.jsxs("div",{className:"flex flex-wrap gap-4",children:[e.jsx(a,{variant:"success",showDot:!0,children:"Active"}),e.jsx(a,{variant:"error",showDot:!0,children:"Offline"}),e.jsx(a,{variant:"warning",showDot:!0,children:"Away"}),e.jsx(a,{variant:"info",showDot:!0,children:"Online"})]})},d={render:()=>e.jsxs("div",{className:"flex flex-wrap gap-4",children:[e.jsx(a,{variant:"success",iconBefore:e.jsx("svg",{fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",className:"w-full h-full",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"})}),children:"Verified"}),e.jsx(a,{variant:"error",iconBefore:e.jsx("svg",{fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",className:"w-full h-full",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"})}),children:"Failed"}),e.jsx(a,{variant:"warning",iconBefore:e.jsx("svg",{fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",className:"w-full h-full",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"})}),children:"Warning"})]})},l={render:()=>e.jsxs("div",{className:"flex flex-wrap gap-4",children:[e.jsx(a,{variant:"info",iconAfter:e.jsx("svg",{fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",className:"w-full h-full",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M13 7l5 5m0 0l-5 5m5-5H6"})}),children:"Next"}),e.jsx(a,{variant:"primary",iconAfter:e.jsx("svg",{fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",className:"w-full h-full",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"})}),children:"External"})]})},c={render:()=>e.jsxs("div",{className:"flex flex-col gap-6",children:[e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Status Indicators"}),e.jsxs("div",{className:"flex flex-wrap gap-2",children:[e.jsx(a,{variant:"success",showDot:!0,children:"Active"}),e.jsx(a,{variant:"error",showDot:!0,children:"Expired"}),e.jsx(a,{variant:"warning",showDot:!0,children:"Pending"}),e.jsx(a,{variant:"neutral",showDot:!0,children:"Inactive"})]})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Counts & Labels"}),e.jsxs("div",{className:"flex flex-wrap gap-2",children:[e.jsx(a,{variant:"neutral",size:"sm",shape:"rounded",children:"Beta"}),e.jsx(a,{variant:"primary",size:"sm",shape:"rounded",children:"New"}),e.jsx(a,{variant:"info",size:"sm",shape:"pill",children:"3"}),e.jsx(a,{variant:"warning",size:"sm",shape:"pill",children:"99+"})]})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Permissions"}),e.jsxs("div",{className:"flex flex-wrap gap-2",children:[e.jsx(a,{variant:"success",size:"sm",children:"Read"}),e.jsx(a,{variant:"warning",size:"sm",children:"Write"}),e.jsx(a,{variant:"error",size:"sm",children:"Delete"}),e.jsx(a,{variant:"info",size:"sm",children:"Admin"})]})]})]})},p={args:{children:"Custom Badge",variant:"primary",size:"md",shape:"pill",showDot:!1}},m={render:()=>e.jsxs("div",{className:"flex flex-col gap-4",children:[e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Badges with semantic meaning"}),e.jsxs("div",{className:"flex flex-wrap gap-2",children:[e.jsx(a,{variant:"success",role:"status","aria-label":"Status: Active",children:"Active"}),e.jsx(a,{variant:"error",role:"status","aria-label":"Status: Failed",children:"Failed"}),e.jsx(a,{variant:"warning",role:"status","aria-label":"Status: Pending",children:"Pending"})]})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Notification counts with aria-label"}),e.jsxs("div",{className:"flex flex-wrap gap-2",children:[e.jsx(a,{variant:"info",shape:"pill","aria-label":"3 unread messages",children:"3"}),e.jsx(a,{variant:"error",shape:"pill","aria-label":"99 or more notifications",children:"99+"})]})]})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0}]}}}};var x,f,B,w,j;s.parameters={...s.parameters,docs:{...(x=s.parameters)==null?void 0:x.docs,source:{originalSource:`{
  args: {
    children: 'Badge'
  }
}`,...(B=(f=s.parameters)==null?void 0:f.docs)==null?void 0:B.source},description:{story:"Default badge with neutral variant",...(j=(w=s.parameters)==null?void 0:w.docs)==null?void 0:j.description}}};var y,N,b,k,S;n.parameters={...n.parameters,docs:{...(y=n.parameters)==null?void 0:y.docs,source:{originalSource:`{
  render: () => <div className="flex flex-wrap gap-4">
      <Badge variant="success">Success</Badge>
      <Badge variant="error">Error</Badge>
      <Badge variant="warning">Warning</Badge>
      <Badge variant="info">Info</Badge>
      <Badge variant="neutral">Neutral</Badge>
      <Badge variant="primary">Primary</Badge>
    </div>
}`,...(b=(N=n.parameters)==null?void 0:N.docs)==null?void 0:b.source},description:{story:"All badge variants showcasing semantic colors",...(S=(k=n.parameters)==null?void 0:k.docs)==null?void 0:S.description}}};var z,L,D,A,W;i.parameters={...i.parameters,docs:{...(z=i.parameters)==null?void 0:z.docs,source:{originalSource:`{
  render: () => <div className="flex flex-wrap items-center gap-4">
      <Badge variant="primary" size="sm">
        Small
      </Badge>
      <Badge variant="primary" size="md">
        Medium
      </Badge>
      <Badge variant="primary" size="lg">
        Large
      </Badge>
    </div>
}`,...(D=(L=i.parameters)==null?void 0:L.docs)==null?void 0:D.source},description:{story:"Badge sizes from small to large",...(W=(A=i.parameters)==null?void 0:A.docs)==null?void 0:W.description}}};var P,C,R,I,M;t.parameters={...t.parameters,docs:{...(P=t.parameters)==null?void 0:P.docs,source:{originalSource:`{
  render: () => <div className="flex flex-wrap gap-4">
      <Badge variant="primary" shape="pill">
        Pill Shape
      </Badge>
      <Badge variant="primary" shape="rounded">
        Rounded Shape
      </Badge>
    </div>
}`,...(R=(C=t.parameters)==null?void 0:C.docs)==null?void 0:R.source},description:{story:"Badge shapes: rounded vs pill",...(M=(I=t.parameters)==null?void 0:I.docs)==null?void 0:M.description}}};var V,E,T,F,O;o.parameters={...o.parameters,docs:{...(V=o.parameters)==null?void 0:V.docs,source:{originalSource:`{
  render: () => <div className="flex flex-wrap gap-4">
      <Badge variant="success" showDot>
        Active
      </Badge>
      <Badge variant="error" showDot>
        Offline
      </Badge>
      <Badge variant="warning" showDot>
        Away
      </Badge>
      <Badge variant="info" showDot>
        Online
      </Badge>
    </div>
}`,...(T=(E=o.parameters)==null?void 0:E.docs)==null?void 0:T.source},description:{story:"Badges with dot indicators for status",...(O=(F=o.parameters)==null?void 0:F.docs)==null?void 0:O.description}}};var H,_,q,U,G;d.parameters={...d.parameters,docs:{...(H=d.parameters)==null?void 0:H.docs,source:{originalSource:`{
  render: () => <div className="flex flex-wrap gap-4">
      <Badge variant="success" iconBefore={<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" className="w-full h-full">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>}>
        Verified
      </Badge>
      <Badge variant="error" iconBefore={<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" className="w-full h-full">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>}>
        Failed
      </Badge>
      <Badge variant="warning" iconBefore={<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" className="w-full h-full">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>}>
        Warning
      </Badge>
    </div>
}`,...(q=(_=d.parameters)==null?void 0:_.docs)==null?void 0:q.source},description:{story:"Badges with icons before content",...(G=(U=d.parameters)==null?void 0:U.docs)==null?void 0:G.description}}};var J,K,Q,X,Y;l.parameters={...l.parameters,docs:{...(J=l.parameters)==null?void 0:J.docs,source:{originalSource:`{
  render: () => <div className="flex flex-wrap gap-4">
      <Badge variant="info" iconAfter={<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" className="w-full h-full">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
          </svg>}>
        Next
      </Badge>
      <Badge variant="primary" iconAfter={<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" className="w-full h-full">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
          </svg>}>
        External
      </Badge>
    </div>
}`,...(Q=(K=l.parameters)==null?void 0:K.docs)==null?void 0:Q.source},description:{story:"Badges with icons after content",...(Y=(X=l.parameters)==null?void 0:X.docs)==null?void 0:Y.description}}};var Z,$,ee,ae,re;c.parameters={...c.parameters,docs:{...(Z=c.parameters)==null?void 0:Z.docs,source:{originalSource:`{
  render: () => <div className="flex flex-col gap-6">
      {/* Status badges */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">Status Indicators</h3>
        <div className="flex flex-wrap gap-2">
          <Badge variant="success" showDot>
            Active
          </Badge>
          <Badge variant="error" showDot>
            Expired
          </Badge>
          <Badge variant="warning" showDot>
            Pending
          </Badge>
          <Badge variant="neutral" showDot>
            Inactive
          </Badge>
        </div>
      </div>

      {/* Counts and labels */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">Counts & Labels</h3>
        <div className="flex flex-wrap gap-2">
          <Badge variant="neutral" size="sm" shape="rounded">
            Beta
          </Badge>
          <Badge variant="primary" size="sm" shape="rounded">
            New
          </Badge>
          <Badge variant="info" size="sm" shape="pill">
            3
          </Badge>
          <Badge variant="warning" size="sm" shape="pill">
            99+
          </Badge>
        </div>
      </div>

      {/* Permission badges */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">Permissions</h3>
        <div className="flex flex-wrap gap-2">
          <Badge variant="success" size="sm">
            Read
          </Badge>
          <Badge variant="warning" size="sm">
            Write
          </Badge>
          <Badge variant="error" size="sm">
            Delete
          </Badge>
          <Badge variant="info" size="sm">
            Admin
          </Badge>
        </div>
      </div>
    </div>
}`,...(ee=($=c.parameters)==null?void 0:$.docs)==null?void 0:ee.source},description:{story:"Real-world use cases for badges",...(re=(ae=c.parameters)==null?void 0:ae.docs)==null?void 0:re.description}}};var se,ne,ie,te,oe;p.parameters={...p.parameters,docs:{...(se=p.parameters)==null?void 0:se.docs,source:{originalSource:`{
  args: {
    children: 'Custom Badge',
    variant: 'primary',
    size: 'md',
    shape: 'pill',
    showDot: false
  }
}`,...(ie=(ne=p.parameters)==null?void 0:ne.docs)==null?void 0:ie.source},description:{story:"Interactive playground for testing all combinations",...(oe=(te=p.parameters)==null?void 0:te.docs)==null?void 0:oe.description}}};var de,le,ce,pe,me;m.parameters={...m.parameters,docs:{...(de=m.parameters)==null?void 0:de.docs,source:{originalSource:`{
  render: () => <div className="flex flex-col gap-4">
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">
          Badges with semantic meaning
        </h3>
        <div className="flex flex-wrap gap-2">
          <Badge variant="success" role="status" aria-label="Status: Active">
            Active
          </Badge>
          <Badge variant="error" role="status" aria-label="Status: Failed">
            Failed
          </Badge>
          <Badge variant="warning" role="status" aria-label="Status: Pending">
            Pending
          </Badge>
        </div>
      </div>

      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">
          Notification counts with aria-label
        </h3>
        <div className="flex flex-wrap gap-2">
          <Badge variant="info" shape="pill" aria-label="3 unread messages">
            3
          </Badge>
          <Badge variant="error" shape="pill" aria-label="99 or more notifications">
            99+
          </Badge>
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
}`,...(ce=(le=m.parameters)==null?void 0:le.docs)==null?void 0:ce.source},description:{story:"Accessibility test - screen reader friendly",...(me=(pe=m.parameters)==null?void 0:pe.docs)==null?void 0:me.description}}};const De=["Default","Variants","Sizes","Shapes","WithDot","WithIconBefore","WithIconAfter","UseCases","Playground","Accessibility"];export{m as Accessibility,s as Default,p as Playground,t as Shapes,i as Sizes,c as UseCases,n as Variants,o as WithDot,l as WithIconAfter,d as WithIconBefore,De as __namedExportsOrder,Le as default};
