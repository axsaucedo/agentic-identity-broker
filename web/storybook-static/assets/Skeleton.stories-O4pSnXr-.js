import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as We}from"./index-ClcD9ViR.js";import{c as k,a as Me}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const A=Me("bg-gray-200 overflow-hidden",{variants:{variant:{line:"h-4 w-full rounded",circle:"rounded-full",rectangle:"",rounded:"rounded-lg"},animate:{true:"animate-pulse",false:""}},defaultVariants:{variant:"line",animate:!0}}),i=We.forwardRef(({variant:t="line",animate:a=!0,width:r,height:d,count:j=1,gap:qe="0.5rem",className:y,style:b,...N},f)=>{const S=(()=>{const s={};return r&&(s.width=r),d&&(s.height=d),t==="circle"&&(!r&&!d?(s.width="40px",s.height="40px"):r&&!d?s.height=r:d&&!r&&(s.width=d)),(t==="rectangle"||t==="rounded")&&!d&&(s.height="100px"),(t==="rectangle"||t==="rounded")&&!r&&(s.width="100%"),s})(),Fe={...S,...b};return j===1?e.jsx("div",{ref:f,role:"status","aria-label":"Loading","aria-live":"polite",className:k(A({variant:t,animate:a}),y),style:Fe,...N,children:e.jsx("span",{className:"sr-only",children:"Loading..."})}):e.jsxs("div",{ref:f,role:"status","aria-label":"Loading","aria-live":"polite",className:y,style:b,...N,children:[e.jsx("div",{className:"flex flex-col",style:{gap:qe},children:Array.from({length:j}).map((s,Pe)=>e.jsx("div",{className:k(A({variant:t,animate:a})),style:S},Pe))}),e.jsx("span",{className:"sr-only",children:"Loading..."})]})});i.displayName="Skeleton";i.__docgenInfo={description:`Skeleton component for displaying loading state placeholders.
Supports multiple shapes, animations, and flexible sizing.

@example
\`\`\`tsx
// Basic line skeleton
<Skeleton />

// Avatar skeleton
<Skeleton variant="circle" width="48px" height="48px" />

// Card image skeleton
<Skeleton variant="rounded" width="100%" height="200px" />

// Multiple text lines
<Skeleton count={3} gap="0.5rem" />

// Without animation
<Skeleton animate={false} />
\`\`\``,methods:[],displayName:"Skeleton",props:{variant:{required:!1,tsType:{name:"union",raw:"'line' | 'circle' | 'rectangle' | 'rounded'",elements:[{name:"literal",value:"'line'"},{name:"literal",value:"'circle'"},{name:"literal",value:"'rectangle'"},{name:"literal",value:"'rounded'"}]},description:"Shape variant of the skeleton",defaultValue:{value:"'line'",computed:!1}},animate:{required:!1,tsType:{name:"boolean"},description:"Whether to animate with pulse effect",defaultValue:{value:"true",computed:!1}},width:{required:!1,tsType:{name:"string"},description:"Custom width (CSS value: px, %, rem, etc.)"},height:{required:!1,tsType:{name:"string"},description:"Custom height (CSS value: px, %, rem, etc.)"},count:{required:!1,tsType:{name:"number"},description:"Number of skeletons to render",defaultValue:{value:"1",computed:!1}},gap:{required:!1,tsType:{name:"string"},description:"Gap between multiple skeletons (CSS value: px, rem, etc.)",defaultValue:{value:"'0.5rem'",computed:!1}}},composes:["Omit"]};const He={title:"Design System/Feedback/Skeleton",component:i,parameters:{layout:"padded"},tags:["autodocs"],argTypes:{variant:{control:"select",options:["line","circle","rectangle","rounded"],description:"Shape variant of the skeleton"},animate:{control:"boolean",description:"Whether to animate with pulse effect"},width:{control:"text",description:"Custom width (CSS value: px, %, rem, etc.)"},height:{control:"text",description:"Custom height (CSS value: px, %, rem, etc.)"},count:{control:"number",description:"Number of skeletons to render"},gap:{control:"text",description:"Gap between multiple skeletons (CSS value)"}}},n={args:{variant:"line",animate:!0}},l={render:()=>e.jsxs("div",{className:"max-w-2xl space-y-6",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Article Loading"}),e.jsxs("div",{className:"space-y-3",children:[e.jsx(i,{width:"70%",height:"32px"}),e.jsxs("div",{className:"space-y-2 pt-2",children:[e.jsx(i,{}),e.jsx(i,{}),e.jsx(i,{width:"90%"}),e.jsx(i,{width:"85%"})]})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Comment Loading"}),e.jsxs("div",{className:"space-y-2",children:[e.jsx(i,{count:4,gap:"0.5rem"}),e.jsx(i,{width:"60%"})]})]})]}),args:{variant:"line",count:3}},c={render:()=>e.jsxs("div",{className:"space-y-6 max-w-2xl",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Avatar Sizes"}),e.jsxs("div",{className:"flex items-center gap-4",children:[e.jsx(i,{variant:"circle",width:"32px",height:"32px"}),e.jsx(i,{variant:"circle",width:"40px",height:"40px"}),e.jsx(i,{variant:"circle",width:"48px",height:"48px"}),e.jsx(i,{variant:"circle",width:"64px",height:"64px"}),e.jsx(i,{variant:"circle",width:"80px",height:"80px"})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"User Profile Loading"}),e.jsxs("div",{className:"flex items-center gap-3",children:[e.jsx(i,{variant:"circle",width:"48px",height:"48px"}),e.jsxs("div",{className:"flex-1 space-y-2",children:[e.jsx(i,{width:"150px",height:"16px"}),e.jsx(i,{width:"200px",height:"14px"})]})]})]})]}),args:{variant:"circle",width:"48px",height:"48px"}},o={render:()=>e.jsx("div",{className:"grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 max-w-6xl",children:Array.from({length:3}).map((t,a)=>e.jsxs("div",{className:"bg-white rounded-lg border border-gray-200 overflow-hidden",children:[e.jsx(i,{variant:"rectangle",width:"100%",height:"200px",animate:!0}),e.jsxs("div",{className:"p-4 space-y-3",children:[e.jsx(i,{width:"80%",height:"20px"}),e.jsxs("div",{className:"space-y-2",children:[e.jsx(i,{}),e.jsx(i,{}),e.jsx(i,{width:"60%"})]}),e.jsxs("div",{className:"flex items-center gap-2 pt-2",children:[e.jsx(i,{variant:"circle",width:"24px",height:"24px"}),e.jsx(i,{width:"100px",height:"14px"})]})]})]},a))}),args:{variant:"rounded"}},h={render:()=>e.jsx("div",{className:"max-w-4xl",children:e.jsxs("div",{className:"bg-white rounded-lg border border-gray-200 overflow-hidden",children:[e.jsx("div",{className:"border-b border-gray-200 bg-gray-50 p-4",children:e.jsxs("div",{className:"grid grid-cols-4 gap-4",children:[e.jsx(i,{width:"60%",height:"14px"}),e.jsx(i,{width:"70%",height:"14px"}),e.jsx(i,{width:"50%",height:"14px"}),e.jsx(i,{width:"40%",height:"14px"})]})}),e.jsx("div",{className:"divide-y divide-gray-200",children:Array.from({length:5}).map((t,a)=>e.jsx("div",{className:"p-4",children:e.jsxs("div",{className:"grid grid-cols-4 gap-4 items-center",children:[e.jsxs("div",{className:"flex items-center gap-2",children:[e.jsx(i,{variant:"circle",width:"32px",height:"32px"}),e.jsx(i,{width:"80px"})]}),e.jsx(i,{width:"120px"}),e.jsx(i,{width:"90px"}),e.jsx(i,{width:"70px"})]})},a))})]})}),args:{variant:"line"}},x={render:()=>e.jsxs("div",{className:"max-w-2xl space-y-6",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Simple List"}),e.jsx("div",{className:"bg-white rounded-lg border border-gray-200 divide-y divide-gray-200",children:Array.from({length:5}).map((t,a)=>e.jsx("div",{className:"p-4",children:e.jsx(i,{width:"70%"})},a))})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Detailed List"}),e.jsx("div",{className:"bg-white rounded-lg border border-gray-200 divide-y divide-gray-200",children:Array.from({length:4}).map((t,a)=>e.jsxs("div",{className:"p-4 flex items-start gap-3",children:[e.jsx(i,{variant:"circle",width:"40px",height:"40px"}),e.jsxs("div",{className:"flex-1 space-y-2",children:[e.jsx(i,{width:"60%",height:"18px"}),e.jsx(i,{width:"90%"}),e.jsx(i,{width:"70%"})]}),e.jsx(i,{variant:"rounded",width:"60px",height:"24px"})]},a))})]})]}),args:{variant:"line"}},m={render:()=>e.jsxs("div",{className:"space-y-6 max-w-3xl",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Line (default)"}),e.jsx(i,{variant:"line"})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Circle"}),e.jsxs("div",{className:"flex gap-4",children:[e.jsx(i,{variant:"circle",width:"60px",height:"60px"}),e.jsx(i,{variant:"circle",width:"80px",height:"80px"}),e.jsx(i,{variant:"circle",width:"100px",height:"100px"})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Rectangle"}),e.jsx(i,{variant:"rectangle",width:"100%",height:"120px"})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Rounded Rectangle"}),e.jsx(i,{variant:"rounded",width:"100%",height:"120px"})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Custom Sizes"}),e.jsxs("div",{className:"space-y-3",children:[e.jsx(i,{width:"100%",height:"8px"}),e.jsx(i,{width:"80%",height:"12px"}),e.jsx(i,{width:"60%",height:"16px"}),e.jsx(i,{width:"40%",height:"20px"})]})]})]}),args:{variant:"line"}},p={render:()=>e.jsxs("div",{className:"space-y-6 max-w-2xl",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"With Pulse Animation (default)"}),e.jsxs("div",{className:"space-y-2",children:[e.jsx(i,{animate:!0}),e.jsx(i,{animate:!0,width:"90%"}),e.jsx(i,{animate:!0,width:"80%"})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Without Animation"}),e.jsxs("div",{className:"space-y-2",children:[e.jsx(i,{animate:!1}),e.jsx(i,{animate:!1,width:"90%"}),e.jsx(i,{animate:!1,width:"80%"})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Mixed Animation States"}),e.jsxs("div",{className:"grid grid-cols-2 gap-6",children:[e.jsxs("div",{className:"space-y-4",children:[e.jsx("h4",{className:"text-xs font-medium text-gray-700",children:"Animated"}),e.jsx(i,{variant:"circle",width:"64px",height:"64px",animate:!0}),e.jsx(i,{variant:"rounded",width:"100%",height:"100px",animate:!0})]}),e.jsxs("div",{className:"space-y-4",children:[e.jsx("h4",{className:"text-xs font-medium text-gray-700",children:"Static"}),e.jsx(i,{variant:"circle",width:"64px",height:"64px",animate:!1}),e.jsx(i,{variant:"rounded",width:"100%",height:"100px",animate:!1})]})]})]})]}),args:{animate:!0}},g={args:{variant:"line",animate:!0,width:"100%",height:void 0,count:1,gap:"0.5rem"}},v={render:()=>e.jsxs("div",{className:"space-y-8 max-w-4xl",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-4",children:"Consent Request Loading"}),e.jsxs("div",{className:"bg-white rounded-lg border border-gray-200 p-6",children:[e.jsxs("div",{className:"flex items-start gap-4 mb-6",children:[e.jsx(i,{variant:"circle",width:"56px",height:"56px"}),e.jsxs("div",{className:"flex-1 space-y-3",children:[e.jsx(i,{width:"60%",height:"24px"}),e.jsx(i,{width:"90%"}),e.jsx(i,{width:"70%"})]})]}),e.jsxs("div",{className:"space-y-4 mb-6",children:[e.jsxs("div",{className:"flex items-center gap-3",children:[e.jsx(i,{variant:"circle",width:"20px",height:"20px"}),e.jsx(i,{width:"80%"})]}),e.jsxs("div",{className:"flex items-center gap-3",children:[e.jsx(i,{variant:"circle",width:"20px",height:"20px"}),e.jsx(i,{width:"70%"})]}),e.jsxs("div",{className:"flex items-center gap-3",children:[e.jsx(i,{variant:"circle",width:"20px",height:"20px"}),e.jsx(i,{width:"85%"})]})]}),e.jsxs("div",{className:"flex gap-3",children:[e.jsx(i,{variant:"rounded",width:"120px",height:"40px"}),e.jsx(i,{variant:"rounded",width:"100px",height:"40px"})]})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-4",children:"Consent History Loading"}),e.jsxs("div",{className:"bg-white rounded-lg border border-gray-200",children:[e.jsx("div",{className:"p-4 border-b border-gray-200",children:e.jsxs("div",{className:"flex items-center justify-between",children:[e.jsx(i,{width:"200px",height:"20px"}),e.jsx(i,{variant:"rounded",width:"100px",height:"32px"})]})}),e.jsx("div",{className:"divide-y divide-gray-200",children:Array.from({length:4}).map((t,a)=>e.jsxs("div",{className:"p-4 flex items-center gap-4",children:[e.jsx(i,{variant:"circle",width:"40px",height:"40px"}),e.jsxs("div",{className:"flex-1 space-y-2",children:[e.jsx(i,{width:"50%",height:"16px"}),e.jsx(i,{width:"70%",height:"14px"})]}),e.jsx(i,{variant:"rounded",width:"80px",height:"24px"})]},a))})]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-4",children:"Dashboard Loading"}),e.jsx("div",{className:"grid grid-cols-1 md:grid-cols-3 gap-4",children:Array.from({length:3}).map((t,a)=>e.jsxs("div",{className:"bg-white rounded-lg border border-gray-200 p-6",children:[e.jsx(i,{width:"40%",height:"14px",className:"mb-4"}),e.jsx(i,{width:"60%",height:"32px",className:"mb-2"}),e.jsx(i,{width:"80%",height:"12px"})]},a))})]})]}),args:{variant:"line"}},u={render:()=>e.jsx("div",{className:"max-w-2xl",children:e.jsxs("div",{className:"bg-white rounded-lg border border-gray-200 p-6",children:[e.jsx(i,{width:"40%",height:"28px",className:"mb-6"}),e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{children:[e.jsx(i,{width:"120px",height:"14px",className:"mb-2"}),e.jsx(i,{variant:"rounded",width:"100%",height:"40px"})]}),e.jsxs("div",{children:[e.jsx(i,{width:"150px",height:"14px",className:"mb-2"}),e.jsx(i,{variant:"rounded",width:"100%",height:"40px"})]}),e.jsxs("div",{children:[e.jsx(i,{width:"100px",height:"14px",className:"mb-2"}),e.jsx(i,{variant:"rounded",width:"100%",height:"120px"})]}),e.jsxs("div",{className:"space-y-3",children:[e.jsx(i,{width:"180px",height:"14px",className:"mb-3"}),e.jsxs("div",{className:"flex items-center gap-3",children:[e.jsx(i,{variant:"rounded",width:"20px",height:"20px"}),e.jsx(i,{width:"60%"})]}),e.jsxs("div",{className:"flex items-center gap-3",children:[e.jsx(i,{variant:"rounded",width:"20px",height:"20px"}),e.jsx(i,{width:"50%"})]}),e.jsxs("div",{className:"flex items-center gap-3",children:[e.jsx(i,{variant:"rounded",width:"20px",height:"20px"}),e.jsx(i,{width:"55%"})]})]}),e.jsxs("div",{className:"flex gap-3 pt-4",children:[e.jsx(i,{variant:"rounded",width:"120px",height:"44px"}),e.jsx(i,{variant:"rounded",width:"100px",height:"44px"})]})]})]})}),args:{variant:"line"}},w={render:()=>e.jsxs("div",{className:"space-y-6 max-w-3xl",children:[e.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Accessibility Features"}),e.jsxs("ul",{className:"text-sm text-gray-700 space-y-1",children:[e.jsxs("li",{children:["• Uses ",e.jsx("code",{children:'role="status"'})," for screen reader announcements"]}),e.jsxs("li",{children:["• Includes ",e.jsx("code",{children:'aria-label="Loading"'})," for context"]}),e.jsxs("li",{children:["• Uses ",e.jsx("code",{children:'aria-live="polite"'})," to avoid interruptions"]}),e.jsxs("li",{children:["• Contains ",e.jsx("code",{className:"sr-only",children:"Loading..."})," text for screen readers"]}),e.jsx("li",{children:"• Non-interactive element (no focus management needed)"}),e.jsx("li",{children:"• Sufficient color contrast for visibility"}),e.jsx("li",{children:"• Animation can be disabled via prefers-reduced-motion"})]})]}),e.jsxs("div",{className:"bg-white rounded-lg border border-gray-200 p-6",children:[e.jsx("div",{className:"space-y-4",children:e.jsx(i,{count:3})}),e.jsx("p",{className:"text-xs text-gray-500 mt-4",children:'Screen readers will announce "Loading" when this skeleton appears'})]})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0}]}}},args:{variant:"line"}};var L,C,T,_,D;n.parameters={...n.parameters,docs:{...(L=n.parameters)==null?void 0:L.docs,source:{originalSource:`{
  args: {
    variant: 'line',
    animate: true
  }
}`,...(T=(C=n.parameters)==null?void 0:C.docs)==null?void 0:T.source},description:{story:"Default skeleton with line variant",...(D=(_=n.parameters)==null?void 0:_.docs)==null?void 0:D.description}}};var R,q,F,P,W;l.parameters={...l.parameters,docs:{...(R=l.parameters)==null?void 0:R.docs,source:{originalSource:`{
  render: () => <div className="max-w-2xl space-y-6">
      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">
          Article Loading
        </h3>
        <div className="space-y-3">
          {/* Title */}
          <Skeleton width="70%" height="32px" />
          {/* Paragraph lines */}
          <div className="space-y-2 pt-2">
            <Skeleton />
            <Skeleton />
            <Skeleton width="90%" />
            <Skeleton width="85%" />
          </div>
        </div>
      </div>

      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">
          Comment Loading
        </h3>
        <div className="space-y-2">
          <Skeleton count={4} gap="0.5rem" />
          <Skeleton width="60%" />
        </div>
      </div>
    </div>,
  args: {
    variant: 'line',
    count: 3
  }
}`,...(F=(q=l.parameters)==null?void 0:q.docs)==null?void 0:F.source},description:{story:"Multiple text lines simulating a paragraph",...(W=(P=l.parameters)==null?void 0:P.docs)==null?void 0:W.description}}};var M,U,V,z,I;c.parameters={...c.parameters,docs:{...(M=c.parameters)==null?void 0:M.docs,source:{originalSource:`{
  render: () => <div className="space-y-6 max-w-2xl">
      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">Avatar Sizes</h3>
        <div className="flex items-center gap-4">
          <Skeleton variant="circle" width="32px" height="32px" />
          <Skeleton variant="circle" width="40px" height="40px" />
          <Skeleton variant="circle" width="48px" height="48px" />
          <Skeleton variant="circle" width="64px" height="64px" />
          <Skeleton variant="circle" width="80px" height="80px" />
        </div>
      </div>

      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">
          User Profile Loading
        </h3>
        <div className="flex items-center gap-3">
          <Skeleton variant="circle" width="48px" height="48px" />
          <div className="flex-1 space-y-2">
            <Skeleton width="150px" height="16px" />
            <Skeleton width="200px" height="14px" />
          </div>
        </div>
      </div>
    </div>,
  args: {
    variant: 'circle',
    width: '48px',
    height: '48px'
  }
}`,...(V=(U=c.parameters)==null?void 0:U.docs)==null?void 0:V.source},description:{story:"Circular avatar skeleton",...(I=(z=c.parameters)==null?void 0:z.docs)==null?void 0:I.description}}};var E,H,G,O,B;o.parameters={...o.parameters,docs:{...(E=o.parameters)==null?void 0:E.docs,source:{originalSource:`{
  render: () => <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 max-w-6xl">
      {Array.from({
      length: 3
    }).map((_, index) => <div key={index} className="bg-white rounded-lg border border-gray-200 overflow-hidden">
          {/* Card Image */}
          <Skeleton variant="rectangle" width="100%" height="200px" animate />

          {/* Card Content */}
          <div className="p-4 space-y-3">
            {/* Title */}
            <Skeleton width="80%" height="20px" />

            {/* Description */}
            <div className="space-y-2">
              <Skeleton />
              <Skeleton />
              <Skeleton width="60%" />
            </div>

            {/* Footer */}
            <div className="flex items-center gap-2 pt-2">
              <Skeleton variant="circle" width="24px" height="24px" />
              <Skeleton width="100px" height="14px" />
            </div>
          </div>
        </div>)}
    </div>,
  args: {
    variant: 'rounded'
  }
}`,...(G=(H=o.parameters)==null?void 0:H.docs)==null?void 0:G.source},description:{story:"Full card skeleton with image and text",...(B=(O=o.parameters)==null?void 0:O.docs)==null?void 0:B.description}}};var J,K,Q,X,Y;h.parameters={...h.parameters,docs:{...(J=h.parameters)==null?void 0:J.docs,source:{originalSource:`{
  render: () => <div className="max-w-4xl">
      <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        {/* Table Header */}
        <div className="border-b border-gray-200 bg-gray-50 p-4">
          <div className="grid grid-cols-4 gap-4">
            <Skeleton width="60%" height="14px" />
            <Skeleton width="70%" height="14px" />
            <Skeleton width="50%" height="14px" />
            <Skeleton width="40%" height="14px" />
          </div>
        </div>

        {/* Table Rows */}
        <div className="divide-y divide-gray-200">
          {Array.from({
          length: 5
        }).map((_, index) => <div key={index} className="p-4">
              <div className="grid grid-cols-4 gap-4 items-center">
                <div className="flex items-center gap-2">
                  <Skeleton variant="circle" width="32px" height="32px" />
                  <Skeleton width="80px" />
                </div>
                <Skeleton width="120px" />
                <Skeleton width="90px" />
                <Skeleton width="70px" />
              </div>
            </div>)}
        </div>
      </div>
    </div>,
  args: {
    variant: 'line'
  }
}`,...(Q=(K=h.parameters)==null?void 0:K.docs)==null?void 0:Q.source},description:{story:"Table skeleton with rows",...(Y=(X=h.parameters)==null?void 0:X.docs)==null?void 0:Y.description}}};var Z,$,ee,ie,ae;x.parameters={...x.parameters,docs:{...(Z=x.parameters)==null?void 0:Z.docs,source:{originalSource:`{
  render: () => <div className="max-w-2xl space-y-6">
      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">
          Simple List
        </h3>
        <div className="bg-white rounded-lg border border-gray-200 divide-y divide-gray-200">
          {Array.from({
          length: 5
        }).map((_, index) => <div key={index} className="p-4">
              <Skeleton width="70%" />
            </div>)}
        </div>
      </div>

      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">
          Detailed List
        </h3>
        <div className="bg-white rounded-lg border border-gray-200 divide-y divide-gray-200">
          {Array.from({
          length: 4
        }).map((_, index) => <div key={index} className="p-4 flex items-start gap-3">
              <Skeleton variant="circle" width="40px" height="40px" />
              <div className="flex-1 space-y-2">
                <Skeleton width="60%" height="18px" />
                <Skeleton width="90%" />
                <Skeleton width="70%" />
              </div>
              <Skeleton variant="rounded" width="60px" height="24px" />
            </div>)}
        </div>
      </div>
    </div>,
  args: {
    variant: 'line'
  }
}`,...(ee=($=x.parameters)==null?void 0:$.docs)==null?void 0:ee.source},description:{story:"List item skeleton repeated",...(ae=(ie=x.parameters)==null?void 0:ie.docs)==null?void 0:ae.description}}};var te,se,re,de,ne;m.parameters={...m.parameters,docs:{...(te=m.parameters)==null?void 0:te.docs,source:{originalSource:`{
  render: () => <div className="space-y-6 max-w-3xl">
      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">
          Line (default)
        </h3>
        <Skeleton variant="line" />
      </div>

      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">Circle</h3>
        <div className="flex gap-4">
          <Skeleton variant="circle" width="60px" height="60px" />
          <Skeleton variant="circle" width="80px" height="80px" />
          <Skeleton variant="circle" width="100px" height="100px" />
        </div>
      </div>

      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">Rectangle</h3>
        <Skeleton variant="rectangle" width="100%" height="120px" />
      </div>

      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">
          Rounded Rectangle
        </h3>
        <Skeleton variant="rounded" width="100%" height="120px" />
      </div>

      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">
          Custom Sizes
        </h3>
        <div className="space-y-3">
          <Skeleton width="100%" height="8px" />
          <Skeleton width="80%" height="12px" />
          <Skeleton width="60%" height="16px" />
          <Skeleton width="40%" height="20px" />
        </div>
      </div>
    </div>,
  args: {
    variant: 'line'
  }
}`,...(re=(se=m.parameters)==null?void 0:se.docs)==null?void 0:re.source},description:{story:"Different shape variants",...(ne=(de=m.parameters)==null?void 0:de.docs)==null?void 0:ne.description}}};var le,ce,oe,he,xe;p.parameters={...p.parameters,docs:{...(le=p.parameters)==null?void 0:le.docs,source:{originalSource:`{
  render: () => <div className="space-y-6 max-w-2xl">
      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">
          With Pulse Animation (default)
        </h3>
        <div className="space-y-2">
          <Skeleton animate={true} />
          <Skeleton animate={true} width="90%" />
          <Skeleton animate={true} width="80%" />
        </div>
      </div>

      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">
          Without Animation
        </h3>
        <div className="space-y-2">
          <Skeleton animate={false} />
          <Skeleton animate={false} width="90%" />
          <Skeleton animate={false} width="80%" />
        </div>
      </div>

      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-3">
          Mixed Animation States
        </h3>
        <div className="grid grid-cols-2 gap-6">
          <div className="space-y-4">
            <h4 className="text-xs font-medium text-gray-700">Animated</h4>
            <Skeleton variant="circle" width="64px" height="64px" animate />
            <Skeleton variant="rounded" width="100%" height="100px" animate />
          </div>
          <div className="space-y-4">
            <h4 className="text-xs font-medium text-gray-700">Static</h4>
            <Skeleton variant="circle" width="64px" height="64px" animate={false} />
            <Skeleton variant="rounded" width="100%" height="100px" animate={false} />
          </div>
        </div>
      </div>
    </div>,
  args: {
    animate: true
  }
}`,...(oe=(ce=p.parameters)==null?void 0:ce.docs)==null?void 0:oe.source},description:{story:"Animation variations",...(xe=(he=p.parameters)==null?void 0:he.docs)==null?void 0:xe.description}}};var me,pe,ge,ve,ue;g.parameters={...g.parameters,docs:{...(me=g.parameters)==null?void 0:me.docs,source:{originalSource:`{
  args: {
    variant: 'line',
    animate: true,
    width: '100%',
    height: undefined,
    count: 1,
    gap: '0.5rem'
  }
}`,...(ge=(pe=g.parameters)==null?void 0:pe.docs)==null?void 0:ge.source},description:{story:"Interactive playground with all controls",...(ue=(ve=g.parameters)==null?void 0:ve.docs)==null?void 0:ue.description}}};var we,je,ye,be,Ne;v.parameters={...v.parameters,docs:{...(we=v.parameters)==null?void 0:we.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 max-w-4xl">
      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-4">
          Consent Request Loading
        </h3>
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <div className="flex items-start gap-4 mb-6">
            <Skeleton variant="circle" width="56px" height="56px" />
            <div className="flex-1 space-y-3">
              <Skeleton width="60%" height="24px" />
              <Skeleton width="90%" />
              <Skeleton width="70%" />
            </div>
          </div>

          <div className="space-y-4 mb-6">
            <div className="flex items-center gap-3">
              <Skeleton variant="circle" width="20px" height="20px" />
              <Skeleton width="80%" />
            </div>
            <div className="flex items-center gap-3">
              <Skeleton variant="circle" width="20px" height="20px" />
              <Skeleton width="70%" />
            </div>
            <div className="flex items-center gap-3">
              <Skeleton variant="circle" width="20px" height="20px" />
              <Skeleton width="85%" />
            </div>
          </div>

          <div className="flex gap-3">
            <Skeleton variant="rounded" width="120px" height="40px" />
            <Skeleton variant="rounded" width="100px" height="40px" />
          </div>
        </div>
      </div>

      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-4">
          Consent History Loading
        </h3>
        <div className="bg-white rounded-lg border border-gray-200">
          <div className="p-4 border-b border-gray-200">
            <div className="flex items-center justify-between">
              <Skeleton width="200px" height="20px" />
              <Skeleton variant="rounded" width="100px" height="32px" />
            </div>
          </div>
          <div className="divide-y divide-gray-200">
            {Array.from({
            length: 4
          }).map((_, index) => <div key={index} className="p-4 flex items-center gap-4">
                <Skeleton variant="circle" width="40px" height="40px" />
                <div className="flex-1 space-y-2">
                  <Skeleton width="50%" height="16px" />
                  <Skeleton width="70%" height="14px" />
                </div>
                <Skeleton variant="rounded" width="80px" height="24px" />
              </div>)}
          </div>
        </div>
      </div>

      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-4">
          Dashboard Loading
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {Array.from({
          length: 3
        }).map((_, index) => <div key={index} className="bg-white rounded-lg border border-gray-200 p-6">
              <Skeleton width="40%" height="14px" className="mb-4" />
              <Skeleton width="60%" height="32px" className="mb-2" />
              <Skeleton width="80%" height="12px" />
            </div>)}
        </div>
      </div>
    </div>,
  args: {
    variant: 'line'
  }
}`,...(ye=(je=v.parameters)==null?void 0:je.docs)==null?void 0:ye.source},description:{story:"Real-world consent management examples",...(Ne=(be=v.parameters)==null?void 0:be.docs)==null?void 0:Ne.description}}};var fe,Se,ke,Ae,Le;u.parameters={...u.parameters,docs:{...(fe=u.parameters)==null?void 0:fe.docs,source:{originalSource:`{
  render: () => <div className="max-w-2xl">
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <Skeleton width="40%" height="28px" className="mb-6" />

        <div className="space-y-6">
          {/* Text input field */}
          <div>
            <Skeleton width="120px" height="14px" className="mb-2" />
            <Skeleton variant="rounded" width="100%" height="40px" />
          </div>

          {/* Text input field */}
          <div>
            <Skeleton width="150px" height="14px" className="mb-2" />
            <Skeleton variant="rounded" width="100%" height="40px" />
          </div>

          {/* Textarea */}
          <div>
            <Skeleton width="100px" height="14px" className="mb-2" />
            <Skeleton variant="rounded" width="100%" height="120px" />
          </div>

          {/* Checkboxes */}
          <div className="space-y-3">
            <Skeleton width="180px" height="14px" className="mb-3" />
            <div className="flex items-center gap-3">
              <Skeleton variant="rounded" width="20px" height="20px" />
              <Skeleton width="60%" />
            </div>
            <div className="flex items-center gap-3">
              <Skeleton variant="rounded" width="20px" height="20px" />
              <Skeleton width="50%" />
            </div>
            <div className="flex items-center gap-3">
              <Skeleton variant="rounded" width="20px" height="20px" />
              <Skeleton width="55%" />
            </div>
          </div>

          {/* Submit button */}
          <div className="flex gap-3 pt-4">
            <Skeleton variant="rounded" width="120px" height="44px" />
            <Skeleton variant="rounded" width="100px" height="44px" />
          </div>
        </div>
      </div>
    </div>,
  args: {
    variant: 'line'
  }
}`,...(ke=(Se=u.parameters)==null?void 0:Se.docs)==null?void 0:ke.source},description:{story:"Form loading states",...(Le=(Ae=u.parameters)==null?void 0:Ae.docs)==null?void 0:Le.description}}};var Ce,Te,_e,De,Re;w.parameters={...w.parameters,docs:{...(Ce=w.parameters)==null?void 0:Ce.docs,source:{originalSource:`{
  render: () => <div className="space-y-6 max-w-3xl">
      <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
        <h4 className="text-sm font-semibold text-gray-900 mb-2">
          Accessibility Features
        </h4>
        <ul className="text-sm text-gray-700 space-y-1">
          <li>
            • Uses <code>role="status"</code> for screen reader announcements
          </li>
          <li>
            • Includes <code>aria-label="Loading"</code> for context
          </li>
          <li>
            • Uses <code>aria-live="polite"</code> to avoid interruptions
          </li>
          <li>
            • Contains <code className="sr-only">Loading...</code> text for
            screen readers
          </li>
          <li>• Non-interactive element (no focus management needed)</li>
          <li>• Sufficient color contrast for visibility</li>
          <li>• Animation can be disabled via prefers-reduced-motion</li>
        </ul>
      </div>

      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <div className="space-y-4">
          <Skeleton count={3} />
        </div>
        <p className="text-xs text-gray-500 mt-4">
          Screen readers will announce "Loading" when this skeleton appears
        </p>
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
  },
  args: {
    variant: 'line'
  }
}`,...(_e=(Te=w.parameters)==null?void 0:Te.docs)==null?void 0:_e.source},description:{story:"Accessibility features demonstration",...(Re=(De=w.parameters)==null?void 0:De.docs)==null?void 0:Re.description}}};const Ge=["Default","Text","Avatar","Card","Table","List","Shapes","Animation","Playground","ConsentManagementExamples","FormLoading","Accessibility"];export{w as Accessibility,p as Animation,c as Avatar,o as Card,v as ConsentManagementExamples,n as Default,u as FormLoading,x as List,g as Playground,m as Shapes,h as Table,l as Text,Ge as __namedExportsOrder,He as default};
