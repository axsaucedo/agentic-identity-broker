import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as be}from"./index-ClcD9ViR.js";import{c as ye,a as fe}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const Se=fe("flex",{variants:{direction:{row:"flex-row",column:"flex-col"},gap:{xs:"gap-1",sm:"gap-2",md:"gap-4",lg:"gap-6",xl:"gap-8"},align:{start:"items-start",center:"items-center",end:"items-end",stretch:"items-stretch"},justify:{start:"justify-start",center:"justify-center",end:"justify-end","space-between":"justify-between","space-around":"justify-around"},wrap:{true:"flex-wrap",false:"flex-nowrap"}},defaultVariants:{direction:"column",gap:"md",align:"stretch",justify:"start",wrap:!1}}),a=be.forwardRef(({direction:h,gap:u,align:x,justify:g,wrap:xe,className:he,children:ue,...ge},ve)=>e.jsx("div",{ref:ve,className:ye(Se({direction:h,gap:u,align:x,justify:g,wrap:xe}),he),...ge,children:ue}));a.displayName="Stack";a.__docgenInfo={description:`Stack component for flexible row and column layouts.
Provides consistent spacing and alignment for child elements.

@example
\`\`\`tsx
// Default vertical stack with medium gap
<Stack>
  <div>Item 1</div>
  <div>Item 2</div>
  <div>Item 3</div>
</Stack>

// Horizontal stack with large gap and centered items
<Stack direction="row" gap="lg" align="center">
  <button>Action 1</button>
  <button>Action 2</button>
  <button>Action 3</button>
</Stack>

// Responsive stack with wrapping
<Stack direction="row" wrap={true} gap="sm">
  <div>Tag 1</div>
  <div>Tag 2</div>
  <div>Tag 3</div>
</Stack>

// Space-between layout for header/footer patterns
<Stack direction="row" justify="space-between" align="center">
  <h1>Logo</h1>
  <nav>Navigation</nav>
</Stack>

// Nested stacks for complex layouts
<Stack gap="xl">
  <Stack direction="row" justify="space-between">
    <h1>Title</h1>
    <button>Action</button>
  </Stack>
  <Stack gap="sm">
    <p>Content paragraph 1</p>
    <p>Content paragraph 2</p>
  </Stack>
</Stack>
\`\`\``,methods:[],displayName:"Stack",props:{direction:{required:!1,tsType:{name:"union",raw:"'row' | 'column'",elements:[{name:"literal",value:"'row'"},{name:"literal",value:"'column'"}]},description:"Direction of stack - row (horizontal) or column (vertical)"},gap:{required:!1,tsType:{name:"union",raw:"'xs' | 'sm' | 'md' | 'lg' | 'xl'",elements:[{name:"literal",value:"'xs'"},{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"},{name:"literal",value:"'lg'"},{name:"literal",value:"'xl'"}]},description:"Gap size between items"},align:{required:!1,tsType:{name:"union",raw:"'start' | 'center' | 'end' | 'stretch'",elements:[{name:"literal",value:"'start'"},{name:"literal",value:"'center'"},{name:"literal",value:"'end'"},{name:"literal",value:"'stretch'"}]},description:"Alignment on cross axis"},justify:{required:!1,tsType:{name:"union",raw:"'start' | 'center' | 'end' | 'space-between' | 'space-around'",elements:[{name:"literal",value:"'start'"},{name:"literal",value:"'center'"},{name:"literal",value:"'end'"},{name:"literal",value:"'space-between'"},{name:"literal",value:"'space-around'"}]},description:"Justification on main axis"},wrap:{required:!1,tsType:{name:"boolean"},description:"Enable wrapping for multi-line layouts"},children:{required:!0,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Content to stack"}},composes:["VariantProps"]};const Ne={title:"Design System/Layout/Stack",component:a,parameters:{layout:"padded"},tags:["autodocs"],argTypes:{direction:{control:"select",options:["row","column"],description:"Direction of stack - row (horizontal) or column (vertical)",table:{type:{summary:"string"},defaultValue:{summary:"column"}}},gap:{control:"select",options:["xs","sm","md","lg","xl"],description:"Gap size between items",table:{type:{summary:"string"},defaultValue:{summary:"md"}}},align:{control:"select",options:["start","center","end","stretch"],description:"Alignment on cross axis",table:{type:{summary:"string"},defaultValue:{summary:"stretch"}}},justify:{control:"select",options:["start","center","end","space-between","space-around"],description:"Justification on main axis",table:{type:{summary:"string"},defaultValue:{summary:"start"}}},wrap:{control:"boolean",description:"Enable wrapping for multi-line layouts",table:{type:{summary:"boolean"},defaultValue:{summary:"false"}}},children:{control:"text",description:"Stack content"}}},t=({children:h,variant:u="default",height:x})=>{const g={default:"bg-navy-100 text-navy-900 border-navy-200",highlight:"bg-emerald-100 text-emerald-900 border-emerald-200",accent:"bg-sand text-secondary-900 border-slate"};return e.jsx("div",{className:`px-4 py-3 rounded-lg border-2 font-medium text-sm ${g[u]}`,style:x?{height:x}:void 0,children:h})},n={args:{direction:"column",gap:"md",align:"stretch",justify:"start",wrap:!1,children:e.jsxs(e.Fragment,{children:[e.jsx(t,{children:"Item 1"}),e.jsx(t,{children:"Item 2"}),e.jsx(t,{children:"Item 3"})]})}},s={args:{direction:"row",gap:"md",align:"center",justify:"start",wrap:!1,children:e.jsxs(e.Fragment,{children:[e.jsx(t,{children:"Action 1"}),e.jsx(t,{children:"Action 2"}),e.jsx(t,{children:"Action 3"})]})}},r={render:()=>e.jsxs("div",{className:"space-y-8 bg-cream p-6 rounded-xl",children:[e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Extra Small Gap (xs - 4px)"}),e.jsxs(a,{gap:"xs",children:[e.jsx(t,{children:"Item 1"}),e.jsx(t,{children:"Item 2"}),e.jsx(t,{children:"Item 3"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Small Gap (sm - 8px)"}),e.jsxs(a,{gap:"sm",children:[e.jsx(t,{children:"Item 1"}),e.jsx(t,{children:"Item 2"}),e.jsx(t,{children:"Item 3"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Medium Gap (md - 16px) - Default"}),e.jsxs(a,{gap:"md",children:[e.jsx(t,{children:"Item 1"}),e.jsx(t,{children:"Item 2"}),e.jsx(t,{children:"Item 3"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Large Gap (lg - 24px)"}),e.jsxs(a,{gap:"lg",children:[e.jsx(t,{children:"Item 1"}),e.jsx(t,{children:"Item 2"}),e.jsx(t,{children:"Item 3"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Extra Large Gap (xl - 32px)"}),e.jsxs(a,{gap:"xl",children:[e.jsx(t,{children:"Item 1"}),e.jsx(t,{children:"Item 2"}),e.jsx(t,{children:"Item 3"})]})]})]})},i={render:()=>e.jsxs("div",{className:"space-y-8 bg-cream p-6 rounded-xl",children:[e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Align Start"}),e.jsxs(a,{direction:"row",gap:"md",align:"start",className:"bg-white p-4 rounded-lg border-2 border-dashed border-navy-200 min-h-[120px]",children:[e.jsx(t,{height:"40px",children:"Short"}),e.jsx(t,{height:"60px",children:"Medium"}),e.jsx(t,{height:"80px",children:"Tall"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Align Center"}),e.jsxs(a,{direction:"row",gap:"md",align:"center",className:"bg-white p-4 rounded-lg border-2 border-dashed border-navy-200 min-h-[120px]",children:[e.jsx(t,{height:"40px",children:"Short"}),e.jsx(t,{height:"60px",children:"Medium"}),e.jsx(t,{height:"80px",children:"Tall"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Align End"}),e.jsxs(a,{direction:"row",gap:"md",align:"end",className:"bg-white p-4 rounded-lg border-2 border-dashed border-navy-200 min-h-[120px]",children:[e.jsx(t,{height:"40px",children:"Short"}),e.jsx(t,{height:"60px",children:"Medium"}),e.jsx(t,{height:"80px",children:"Tall"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Align Stretch (Default)"}),e.jsxs(a,{direction:"row",gap:"md",align:"stretch",className:"bg-white p-4 rounded-lg border-2 border-dashed border-navy-200 min-h-[120px]",children:[e.jsx(t,{children:"Stretch 1"}),e.jsx(t,{children:"Stretch 2"}),e.jsx(t,{children:"Stretch 3"})]})]})]})},d={render:()=>e.jsxs("div",{className:"space-y-8 bg-cream p-6 rounded-xl",children:[e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Justify Start (Default)"}),e.jsxs(a,{direction:"row",gap:"md",justify:"start",className:"bg-white p-4 rounded-lg border-2 border-dashed border-navy-200",children:[e.jsx(t,{children:"Item 1"}),e.jsx(t,{children:"Item 2"}),e.jsx(t,{children:"Item 3"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Justify Center"}),e.jsxs(a,{direction:"row",gap:"md",justify:"center",className:"bg-white p-4 rounded-lg border-2 border-dashed border-navy-200",children:[e.jsx(t,{children:"Item 1"}),e.jsx(t,{children:"Item 2"}),e.jsx(t,{children:"Item 3"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Justify End"}),e.jsxs(a,{direction:"row",gap:"md",justify:"end",className:"bg-white p-4 rounded-lg border-2 border-dashed border-navy-200",children:[e.jsx(t,{children:"Item 1"}),e.jsx(t,{children:"Item 2"}),e.jsx(t,{children:"Item 3"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Justify Space Between"}),e.jsxs(a,{direction:"row",gap:"md",justify:"space-between",className:"bg-white p-4 rounded-lg border-2 border-dashed border-navy-200",children:[e.jsx(t,{children:"Item 1"}),e.jsx(t,{children:"Item 2"}),e.jsx(t,{children:"Item 3"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-2 text-sm font-semibold text-secondary-600",children:"Justify Space Around"}),e.jsxs(a,{direction:"row",gap:"md",justify:"space-around",className:"bg-white p-4 rounded-lg border-2 border-dashed border-navy-200",children:[e.jsx(t,{children:"Item 1"}),e.jsx(t,{children:"Item 2"}),e.jsx(t,{children:"Item 3"})]})]})]})},c={render:()=>e.jsxs("div",{className:"space-y-8 bg-cream p-6 rounded-xl",children:[e.jsxs("div",{className:"bg-white rounded-lg p-6 shadow-md-premium",children:[e.jsx("h2",{className:"text-xl font-display font-semibold text-navy-900 mb-4",children:"Responsive Stack Behavior"}),e.jsx("p",{className:"text-secondary-600 mb-4",children:"Try resizing your browser window to see how the stack adapts:"}),e.jsxs("ul",{className:"list-disc list-inside text-secondary-600 space-y-2 mb-6",children:[e.jsx("li",{children:"On wide screens: horizontal layout (row)"}),e.jsx("li",{children:"On narrow screens: vertical layout (column)"}),e.jsx("li",{children:"Breakpoint: 768px (Tailwind md: breakpoint)"})]}),e.jsxs(a,{direction:"row",gap:"md",className:"md:flex-row flex-col",children:[e.jsx(t,{variant:"highlight",children:"Responsive Item 1"}),e.jsx(t,{variant:"highlight",children:"Responsive Item 2"}),e.jsx(t,{variant:"highlight",children:"Responsive Item 3"})]})]}),e.jsxs("div",{className:"bg-white rounded-lg p-6 shadow-md-premium",children:[e.jsx("h3",{className:"text-lg font-display font-semibold text-navy-900 mb-3",children:"Mobile-First Approach"}),e.jsx("p",{className:"text-secondary-600 mb-4",children:"Stack defaults to column on mobile, switches to row on tablet+:"}),e.jsxs(a,{direction:"column",gap:"sm",className:"sm:flex-row",children:[e.jsx(t,{variant:"accent",children:"Mobile: Column"}),e.jsx(t,{variant:"accent",children:"Tablet+: Row"})]})]})]})},o={render:()=>e.jsxs(a,{gap:"lg",className:"bg-cream p-6 rounded-xl",children:[e.jsxs(a,{direction:"row",justify:"space-between",align:"center",className:"bg-white p-4 rounded-lg shadow-sm",children:[e.jsx("div",{className:"text-lg font-display font-bold text-navy-900",children:"Application Header"}),e.jsxs(a,{direction:"row",gap:"sm",children:[e.jsx(t,{variant:"default",children:"Profile"}),e.jsx(t,{variant:"default",children:"Settings"})]})]}),e.jsxs(a,{gap:"md",className:"bg-white p-6 rounded-lg shadow-sm",children:[e.jsx("h2",{className:"text-xl font-display font-semibold text-navy-900",children:"Main Content Section"}),e.jsxs(a,{direction:"row",gap:"md",wrap:!0,children:[e.jsx(t,{variant:"highlight",children:"Card 1"}),e.jsx(t,{variant:"highlight",children:"Card 2"}),e.jsx(t,{variant:"highlight",children:"Card 3"})]}),e.jsxs(a,{gap:"sm",children:[e.jsx("div",{className:"text-sm font-semibold text-secondary-600",children:"List Items:"}),e.jsx(t,{variant:"accent",children:"List item A"}),e.jsx(t,{variant:"accent",children:"List item B"}),e.jsx(t,{variant:"accent",children:"List item C"})]})]}),e.jsxs(a,{direction:"row",justify:"space-between",align:"center",className:"bg-white p-4 rounded-lg shadow-sm",children:[e.jsx("div",{className:"text-sm text-secondary-600",children:"Copyright 2024"}),e.jsxs(a,{direction:"row",gap:"md",children:[e.jsx(t,{variant:"default",children:"Terms"}),e.jsx(t,{variant:"default",children:"Privacy"}),e.jsx(t,{variant:"default",children:"Contact"})]})]})]})},l={args:{direction:"column",gap:"md",align:"stretch",justify:"start",wrap:!1,children:e.jsxs(e.Fragment,{children:[e.jsx(t,{children:"Item 1"}),e.jsx(t,{children:"Item 2"}),e.jsx(t,{children:"Item 3"}),e.jsx(t,{children:"Item 4"})]})},parameters:{docs:{description:{story:"Experiment with all stack props using the controls below. Try different directions, gaps, alignments, and justification options."}}}},m={render:()=>e.jsxs(a,{gap:"xl",className:"bg-cream p-6 rounded-xl",children:[e.jsxs("div",{className:"bg-white p-6 rounded-lg shadow-md-premium",children:[e.jsx("h3",{className:"text-lg font-display font-semibold text-navy-900 mb-4",children:"Button Group Pattern"}),e.jsxs(a,{direction:"row",gap:"sm",justify:"end",children:[e.jsx("button",{className:"px-4 py-2 rounded-lg border-2 border-navy-200 text-navy-900 hover:bg-navy-50",children:"Cancel"}),e.jsx("button",{className:"px-4 py-2 rounded-lg bg-emerald-600 text-white hover:bg-emerald-700",children:"Confirm"})]})]}),e.jsxs("div",{className:"bg-white p-6 rounded-lg shadow-md-premium",children:[e.jsx("h3",{className:"text-lg font-display font-semibold text-navy-900 mb-4",children:"Form Layout Pattern"}),e.jsxs(a,{gap:"md",children:[e.jsxs("div",{children:[e.jsx("label",{className:"block text-sm font-medium text-secondary-700 mb-1",children:"Name"}),e.jsx("input",{type:"text",className:"w-full px-3 py-2 border-2 border-slate rounded-lg",placeholder:"Enter your name"})]}),e.jsxs("div",{children:[e.jsx("label",{className:"block text-sm font-medium text-secondary-700 mb-1",children:"Email"}),e.jsx("input",{type:"email",className:"w-full px-3 py-2 border-2 border-slate rounded-lg",placeholder:"Enter your email"})]}),e.jsxs(a,{direction:"row",gap:"sm",justify:"end",children:[e.jsx("button",{className:"px-4 py-2 rounded-lg border-2 border-navy-200 text-navy-900",children:"Cancel"}),e.jsx("button",{className:"px-4 py-2 rounded-lg bg-emerald-600 text-white",children:"Submit"})]})]})]}),e.jsx("div",{className:"bg-white p-4 rounded-lg shadow-md-premium",children:e.jsxs(a,{direction:"row",justify:"space-between",align:"center",children:[e.jsx("div",{className:"text-xl font-display font-bold text-navy-900",children:"Logo"}),e.jsxs(a,{direction:"row",gap:"md",align:"center",children:[e.jsx("a",{href:"#",className:"text-navy-900 hover:text-emerald-600",children:"Home"}),e.jsx("a",{href:"#",className:"text-navy-900 hover:text-emerald-600",children:"About"}),e.jsx("a",{href:"#",className:"text-navy-900 hover:text-emerald-600",children:"Contact"}),e.jsx("button",{className:"px-4 py-2 rounded-lg bg-emerald-600 text-white",children:"Sign In"})]})]})}),e.jsxs("div",{className:"bg-white p-6 rounded-lg shadow-md-premium",children:[e.jsx("h3",{className:"text-lg font-display font-semibold text-navy-900 mb-4",children:"Card Grid with Wrapping"}),e.jsxs(a,{direction:"row",gap:"md",wrap:!0,children:[e.jsxs("div",{className:"bg-navy-50 p-4 rounded-lg border-2 border-navy-200 min-w-[150px]",children:[e.jsx("div",{className:"font-semibold text-navy-900 mb-2",children:"Feature 1"}),e.jsx("div",{className:"text-sm text-secondary-600",children:"Description"})]}),e.jsxs("div",{className:"bg-navy-50 p-4 rounded-lg border-2 border-navy-200 min-w-[150px]",children:[e.jsx("div",{className:"font-semibold text-navy-900 mb-2",children:"Feature 2"}),e.jsx("div",{className:"text-sm text-secondary-600",children:"Description"})]}),e.jsxs("div",{className:"bg-navy-50 p-4 rounded-lg border-2 border-navy-200 min-w-[150px]",children:[e.jsx("div",{className:"font-semibold text-navy-900 mb-2",children:"Feature 3"}),e.jsx("div",{className:"text-sm text-secondary-600",children:"Description"})]}),e.jsxs("div",{className:"bg-navy-50 p-4 rounded-lg border-2 border-navy-200 min-w-[150px]",children:[e.jsx("div",{className:"font-semibold text-navy-900 mb-2",children:"Feature 4"}),e.jsx("div",{className:"text-sm text-secondary-600",children:"Description"})]})]})]})]})},p={render:()=>e.jsx("div",{className:"bg-cream p-8",children:e.jsxs(a,{gap:"md",className:"bg-white rounded-lg shadow-md-premium p-6",children:[e.jsxs("div",{children:[e.jsx("h2",{className:"text-xl font-display font-semibold text-navy-900 mb-3",children:"Accessibility Features"}),e.jsx("p",{className:"text-secondary-600",children:"Stacks are designed with accessibility in mind:"})]}),e.jsxs(a,{gap:"md",children:[e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"Semantic HTML"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Uses standard div elements with flexbox, no ARIA required"})]})]}),e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"Logical Source Order"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Visual order matches DOM order for screen readers"})]})]}),e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"Touch-Friendly Spacing"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Gap options ensure adequate spacing for touch targets (minimum 44x44px)"})]})]}),e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"Responsive & Flexible"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Works across all viewport sizes and zoom levels"})]})]}),e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"No Motion Dependencies"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Pure layout component with no animations (respects prefers-reduced-motion)"})]})]})]})]})}),parameters:{docs:{description:{story:"Stacks follow accessibility best practices with semantic HTML, logical source order, and consideration for all users regardless of device or ability."}}}};var v,b,y,f,S;n.parameters={...n.parameters,docs:{...(v=n.parameters)==null?void 0:v.docs,source:{originalSource:`{
  args: {
    direction: 'column',
    gap: 'md',
    align: 'stretch',
    justify: 'start',
    wrap: false,
    children: <>
        <StackItem>Item 1</StackItem>
        <StackItem>Item 2</StackItem>
        <StackItem>Item 3</StackItem>
      </>
  }
}`,...(y=(b=n.parameters)==null?void 0:b.docs)==null?void 0:y.source},description:{story:`Default vertical stack with medium gap.
This is the most common stack configuration for vertical layouts.`,...(S=(f=n.parameters)==null?void 0:f.docs)==null?void 0:S.description}}};var k,j,w,I,N;s.parameters={...s.parameters,docs:{...(k=s.parameters)==null?void 0:k.docs,source:{originalSource:`{
  args: {
    direction: 'row',
    gap: 'md',
    align: 'center',
    justify: 'start',
    wrap: false,
    children: <>
        <StackItem>Action 1</StackItem>
        <StackItem>Action 2</StackItem>
        <StackItem>Action 3</StackItem>
      </>
  }
}`,...(w=(j=s.parameters)==null?void 0:j.docs)==null?void 0:w.source},description:{story:`Horizontal stack (row direction).
Useful for navigation bars, button groups, and horizontal layouts.`,...(N=(I=s.parameters)==null?void 0:I.docs)==null?void 0:N.description}}};var A,C,T,R,M;r.parameters={...r.parameters,docs:{...(A=r.parameters)==null?void 0:A.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 bg-cream p-6 rounded-xl">
      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Extra Small Gap (xs - 4px)
        </div>
        <Stack gap="xs">
          <StackItem>Item 1</StackItem>
          <StackItem>Item 2</StackItem>
          <StackItem>Item 3</StackItem>
        </Stack>
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Small Gap (sm - 8px)
        </div>
        <Stack gap="sm">
          <StackItem>Item 1</StackItem>
          <StackItem>Item 2</StackItem>
          <StackItem>Item 3</StackItem>
        </Stack>
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Medium Gap (md - 16px) - Default
        </div>
        <Stack gap="md">
          <StackItem>Item 1</StackItem>
          <StackItem>Item 2</StackItem>
          <StackItem>Item 3</StackItem>
        </Stack>
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Large Gap (lg - 24px)
        </div>
        <Stack gap="lg">
          <StackItem>Item 1</StackItem>
          <StackItem>Item 2</StackItem>
          <StackItem>Item 3</StackItem>
        </Stack>
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Extra Large Gap (xl - 32px)
        </div>
        <Stack gap="xl">
          <StackItem>Item 1</StackItem>
          <StackItem>Item 2</StackItem>
          <StackItem>Item 3</StackItem>
        </Stack>
      </div>
    </div>
}`,...(T=(C=r.parameters)==null?void 0:C.docs)==null?void 0:T.source},description:{story:`All gap sizes for visual comparison.
Choose gap sizes based on visual hierarchy and spacing needs:
- xs (4px): Minimal spacing, tight grouping
- sm (8px): Small spacing, related items
- md (16px): Standard spacing (default)
- lg (24px): Generous spacing, distinct sections
- xl (32px): Maximum spacing, strong separation`,...(M=(R=r.parameters)==null?void 0:R.docs)==null?void 0:M.description}}};var L,D,F,z,E;i.parameters={...i.parameters,docs:{...(L=i.parameters)==null?void 0:L.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 bg-cream p-6 rounded-xl">
      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Align Start
        </div>
        <Stack direction="row" gap="md" align="start" className="bg-white p-4 rounded-lg border-2 border-dashed border-navy-200 min-h-[120px]">
          <StackItem height="40px">Short</StackItem>
          <StackItem height="60px">Medium</StackItem>
          <StackItem height="80px">Tall</StackItem>
        </Stack>
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Align Center
        </div>
        <Stack direction="row" gap="md" align="center" className="bg-white p-4 rounded-lg border-2 border-dashed border-navy-200 min-h-[120px]">
          <StackItem height="40px">Short</StackItem>
          <StackItem height="60px">Medium</StackItem>
          <StackItem height="80px">Tall</StackItem>
        </Stack>
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Align End
        </div>
        <Stack direction="row" gap="md" align="end" className="bg-white p-4 rounded-lg border-2 border-dashed border-navy-200 min-h-[120px]">
          <StackItem height="40px">Short</StackItem>
          <StackItem height="60px">Medium</StackItem>
          <StackItem height="80px">Tall</StackItem>
        </Stack>
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Align Stretch (Default)
        </div>
        <Stack direction="row" gap="md" align="stretch" className="bg-white p-4 rounded-lg border-2 border-dashed border-navy-200 min-h-[120px]">
          <StackItem>Stretch 1</StackItem>
          <StackItem>Stretch 2</StackItem>
          <StackItem>Stretch 3</StackItem>
        </Stack>
      </div>
    </div>
}`,...(F=(D=i.parameters)==null?void 0:D.docs)==null?void 0:F.source},description:{story:`Cross-axis alignment options.
Controls how items align perpendicular to the stack direction:
- start: Align to start of cross axis
- center: Center items on cross axis
- end: Align to end of cross axis
- stretch: Fill cross axis (default)`,...(E=(z=i.parameters)==null?void 0:z.docs)==null?void 0:E.description}}};var G,P,H,J,q;d.parameters={...d.parameters,docs:{...(G=d.parameters)==null?void 0:G.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 bg-cream p-6 rounded-xl">
      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Justify Start (Default)
        </div>
        <Stack direction="row" gap="md" justify="start" className="bg-white p-4 rounded-lg border-2 border-dashed border-navy-200">
          <StackItem>Item 1</StackItem>
          <StackItem>Item 2</StackItem>
          <StackItem>Item 3</StackItem>
        </Stack>
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Justify Center
        </div>
        <Stack direction="row" gap="md" justify="center" className="bg-white p-4 rounded-lg border-2 border-dashed border-navy-200">
          <StackItem>Item 1</StackItem>
          <StackItem>Item 2</StackItem>
          <StackItem>Item 3</StackItem>
        </Stack>
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Justify End
        </div>
        <Stack direction="row" gap="md" justify="end" className="bg-white p-4 rounded-lg border-2 border-dashed border-navy-200">
          <StackItem>Item 1</StackItem>
          <StackItem>Item 2</StackItem>
          <StackItem>Item 3</StackItem>
        </Stack>
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Justify Space Between
        </div>
        <Stack direction="row" gap="md" justify="space-between" className="bg-white p-4 rounded-lg border-2 border-dashed border-navy-200">
          <StackItem>Item 1</StackItem>
          <StackItem>Item 2</StackItem>
          <StackItem>Item 3</StackItem>
        </Stack>
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold text-secondary-600">
          Justify Space Around
        </div>
        <Stack direction="row" gap="md" justify="space-around" className="bg-white p-4 rounded-lg border-2 border-dashed border-navy-200">
          <StackItem>Item 1</StackItem>
          <StackItem>Item 2</StackItem>
          <StackItem>Item 3</StackItem>
        </Stack>
      </div>
    </div>
}`,...(H=(P=d.parameters)==null?void 0:P.docs)==null?void 0:H.source},description:{story:`Main-axis justification options.
Controls how items are distributed along the stack direction:
- start: Pack items to start (default)
- center: Pack items to center
- end: Pack items to end
- space-between: First item at start, last at end, equal spacing
- space-around: Equal space around each item`,...(q=(J=d.parameters)==null?void 0:J.docs)==null?void 0:q.description}}};var B,O,V,W,_;c.parameters={...c.parameters,docs:{...(B=c.parameters)==null?void 0:B.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 bg-cream p-6 rounded-xl">
      <div className="bg-white rounded-lg p-6 shadow-md-premium">
        <h2 className="text-xl font-display font-semibold text-navy-900 mb-4">
          Responsive Stack Behavior
        </h2>
        <p className="text-secondary-600 mb-4">
          Try resizing your browser window to see how the stack adapts:
        </p>
        <ul className="list-disc list-inside text-secondary-600 space-y-2 mb-6">
          <li>On wide screens: horizontal layout (row)</li>
          <li>On narrow screens: vertical layout (column)</li>
          <li>Breakpoint: 768px (Tailwind md: breakpoint)</li>
        </ul>

        <Stack direction="row" gap="md" className="md:flex-row flex-col">
          <StackItem variant="highlight">Responsive Item 1</StackItem>
          <StackItem variant="highlight">Responsive Item 2</StackItem>
          <StackItem variant="highlight">Responsive Item 3</StackItem>
        </Stack>
      </div>

      <div className="bg-white rounded-lg p-6 shadow-md-premium">
        <h3 className="text-lg font-display font-semibold text-navy-900 mb-3">
          Mobile-First Approach
        </h3>
        <p className="text-secondary-600 mb-4">
          Stack defaults to column on mobile, switches to row on tablet+:
        </p>

        <Stack direction="column" gap="sm" className="sm:flex-row">
          <StackItem variant="accent">Mobile: Column</StackItem>
          <StackItem variant="accent">Tablet+: Row</StackItem>
        </Stack>
      </div>
    </div>
}`,...(V=(O=c.parameters)==null?void 0:O.docs)==null?void 0:V.source},description:{story:`Responsive stack with direction changes.
Stack changes from row (horizontal) on desktop to column (vertical) on mobile.
This pattern is useful for responsive navigation, card grids, and adaptive layouts.`,...(_=(W=c.parameters)==null?void 0:W.docs)==null?void 0:_.description}}};var U,K,$,Q,X;o.parameters={...o.parameters,docs:{...(U=o.parameters)==null?void 0:U.docs,source:{originalSource:`{
  render: () => <Stack gap="lg" className="bg-cream p-6 rounded-xl">
      {/* Header with horizontal stack */}
      <Stack direction="row" justify="space-between" align="center" className="bg-white p-4 rounded-lg shadow-sm">
        <div className="text-lg font-display font-bold text-navy-900">
          Application Header
        </div>
        <Stack direction="row" gap="sm">
          <StackItem variant="default">Profile</StackItem>
          <StackItem variant="default">Settings</StackItem>
        </Stack>
      </Stack>

      {/* Main content area */}
      <Stack gap="md" className="bg-white p-6 rounded-lg shadow-sm">
        <h2 className="text-xl font-display font-semibold text-navy-900">
          Main Content Section
        </h2>

        {/* Nested horizontal stack for cards */}
        <Stack direction="row" gap="md" wrap={true}>
          <StackItem variant="highlight">Card 1</StackItem>
          <StackItem variant="highlight">Card 2</StackItem>
          <StackItem variant="highlight">Card 3</StackItem>
        </Stack>

        {/* Nested vertical stack for list */}
        <Stack gap="sm">
          <div className="text-sm font-semibold text-secondary-600">List Items:</div>
          <StackItem variant="accent">List item A</StackItem>
          <StackItem variant="accent">List item B</StackItem>
          <StackItem variant="accent">List item C</StackItem>
        </Stack>
      </Stack>

      {/* Footer with horizontal stack */}
      <Stack direction="row" justify="space-between" align="center" className="bg-white p-4 rounded-lg shadow-sm">
        <div className="text-sm text-secondary-600">
          Copyright 2024
        </div>
        <Stack direction="row" gap="md">
          <StackItem variant="default">Terms</StackItem>
          <StackItem variant="default">Privacy</StackItem>
          <StackItem variant="default">Contact</StackItem>
        </Stack>
      </Stack>
    </Stack>
}`,...($=(K=o.parameters)==null?void 0:K.docs)==null?void 0:$.source},description:{story:`Nested stacks for complex layouts.
Outer stack provides overall structure,
inner stacks handle specific sections.`,...(X=(Q=o.parameters)==null?void 0:Q.docs)==null?void 0:X.description}}};var Y,Z,ee,te,ae;l.parameters={...l.parameters,docs:{...(Y=l.parameters)==null?void 0:Y.docs,source:{originalSource:`{
  args: {
    direction: 'column',
    gap: 'md',
    align: 'stretch',
    justify: 'start',
    wrap: false,
    children: <>
        <StackItem>Item 1</StackItem>
        <StackItem>Item 2</StackItem>
        <StackItem>Item 3</StackItem>
        <StackItem>Item 4</StackItem>
      </>
  },
  parameters: {
    docs: {
      description: {
        story: 'Experiment with all stack props using the controls below. Try different directions, gaps, alignments, and justification options.'
      }
    }
  }
}`,...(ee=(Z=l.parameters)==null?void 0:Z.docs)==null?void 0:ee.source},description:{story:`Interactive playground to experiment with all props.
Try different combinations of direction, gap, alignment, and justification.`,...(ae=(te=l.parameters)==null?void 0:te.docs)==null?void 0:ae.description}}};var ne,se,re,ie,de;m.parameters={...m.parameters,docs:{...(ne=m.parameters)==null?void 0:ne.docs,source:{originalSource:`{
  render: () => <Stack gap="xl" className="bg-cream p-6 rounded-xl">
      {/* Button Group Pattern */}
      <div className="bg-white p-6 rounded-lg shadow-md-premium">
        <h3 className="text-lg font-display font-semibold text-navy-900 mb-4">
          Button Group Pattern
        </h3>
        <Stack direction="row" gap="sm" justify="end">
          <button className="px-4 py-2 rounded-lg border-2 border-navy-200 text-navy-900 hover:bg-navy-50">
            Cancel
          </button>
          <button className="px-4 py-2 rounded-lg bg-emerald-600 text-white hover:bg-emerald-700">
            Confirm
          </button>
        </Stack>
      </div>

      {/* Form Layout Pattern */}
      <div className="bg-white p-6 rounded-lg shadow-md-premium">
        <h3 className="text-lg font-display font-semibold text-navy-900 mb-4">
          Form Layout Pattern
        </h3>
        <Stack gap="md">
          <div>
            <label className="block text-sm font-medium text-secondary-700 mb-1">
              Name
            </label>
            <input type="text" className="w-full px-3 py-2 border-2 border-slate rounded-lg" placeholder="Enter your name" />
          </div>
          <div>
            <label className="block text-sm font-medium text-secondary-700 mb-1">
              Email
            </label>
            <input type="email" className="w-full px-3 py-2 border-2 border-slate rounded-lg" placeholder="Enter your email" />
          </div>
          <Stack direction="row" gap="sm" justify="end">
            <button className="px-4 py-2 rounded-lg border-2 border-navy-200 text-navy-900">
              Cancel
            </button>
            <button className="px-4 py-2 rounded-lg bg-emerald-600 text-white">
              Submit
            </button>
          </Stack>
        </Stack>
      </div>

      {/* Navigation Pattern */}
      <div className="bg-white p-4 rounded-lg shadow-md-premium">
        <Stack direction="row" justify="space-between" align="center">
          <div className="text-xl font-display font-bold text-navy-900">
            Logo
          </div>
          <Stack direction="row" gap="md" align="center">
            <a href="#" className="text-navy-900 hover:text-emerald-600">Home</a>
            <a href="#" className="text-navy-900 hover:text-emerald-600">About</a>
            <a href="#" className="text-navy-900 hover:text-emerald-600">Contact</a>
            <button className="px-4 py-2 rounded-lg bg-emerald-600 text-white">
              Sign In
            </button>
          </Stack>
        </Stack>
      </div>

      {/* Card Grid with Wrapping */}
      <div className="bg-white p-6 rounded-lg shadow-md-premium">
        <h3 className="text-lg font-display font-semibold text-navy-900 mb-4">
          Card Grid with Wrapping
        </h3>
        <Stack direction="row" gap="md" wrap={true}>
          <div className="bg-navy-50 p-4 rounded-lg border-2 border-navy-200 min-w-[150px]">
            <div className="font-semibold text-navy-900 mb-2">Feature 1</div>
            <div className="text-sm text-secondary-600">Description</div>
          </div>
          <div className="bg-navy-50 p-4 rounded-lg border-2 border-navy-200 min-w-[150px]">
            <div className="font-semibold text-navy-900 mb-2">Feature 2</div>
            <div className="text-sm text-secondary-600">Description</div>
          </div>
          <div className="bg-navy-50 p-4 rounded-lg border-2 border-navy-200 min-w-[150px]">
            <div className="font-semibold text-navy-900 mb-2">Feature 3</div>
            <div className="text-sm text-secondary-600">Description</div>
          </div>
          <div className="bg-navy-50 p-4 rounded-lg border-2 border-navy-200 min-w-[150px]">
            <div className="font-semibold text-navy-900 mb-2">Feature 4</div>
            <div className="text-sm text-secondary-600">Description</div>
          </div>
        </Stack>
      </div>
    </Stack>
}`,...(re=(se=m.parameters)==null?void 0:se.docs)==null?void 0:re.source},description:{story:`Real-world usage examples showing common patterns.
These demonstrate how stacks are typically used in applications.`,...(de=(ie=m.parameters)==null?void 0:ie.docs)==null?void 0:de.description}}};var ce,oe,le,me,pe;p.parameters={...p.parameters,docs:{...(ce=p.parameters)==null?void 0:ce.docs,source:{originalSource:`{
  render: () => <div className="bg-cream p-8">
      <Stack gap="md" className="bg-white rounded-lg shadow-md-premium p-6">
        <div>
          <h2 className="text-xl font-display font-semibold text-navy-900 mb-3">
            Accessibility Features
          </h2>
          <p className="text-secondary-600">
            Stacks are designed with accessibility in mind:
          </p>
        </div>

        <Stack gap="md">
          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold">
              ✓
            </div>
            <div>
              <h3 className="font-semibold text-navy-900">Semantic HTML</h3>
              <p className="text-sm text-secondary-600">
                Uses standard div elements with flexbox, no ARIA required
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold">
              ✓
            </div>
            <div>
              <h3 className="font-semibold text-navy-900">Logical Source Order</h3>
              <p className="text-sm text-secondary-600">
                Visual order matches DOM order for screen readers
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold">
              ✓
            </div>
            <div>
              <h3 className="font-semibold text-navy-900">Touch-Friendly Spacing</h3>
              <p className="text-sm text-secondary-600">
                Gap options ensure adequate spacing for touch targets (minimum 44x44px)
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold">
              ✓
            </div>
            <div>
              <h3 className="font-semibold text-navy-900">Responsive & Flexible</h3>
              <p className="text-sm text-secondary-600">
                Works across all viewport sizes and zoom levels
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold">
              ✓
            </div>
            <div>
              <h3 className="font-semibold text-navy-900">No Motion Dependencies</h3>
              <p className="text-sm text-secondary-600">
                Pure layout component with no animations (respects prefers-reduced-motion)
              </p>
            </div>
          </div>
        </Stack>
      </Stack>
    </div>,
  parameters: {
    docs: {
      description: {
        story: 'Stacks follow accessibility best practices with semantic HTML, logical source order, and consideration for all users regardless of device or ability.'
      }
    }
  }
}`,...(le=(oe=p.parameters)==null?void 0:oe.docs)==null?void 0:le.source},description:{story:`Accessibility features demonstration.
Stacks use semantic HTML and support:
- Proper document structure
- Keyboard navigation for interactive children
- Sufficient spacing for touch targets
- No reliance on specific visual order`,...(pe=(me=p.parameters)==null?void 0:me.docs)==null?void 0:pe.description}}};const Ae=["Default","Horizontal","Gaps","Alignment","Justify","Responsive","Nested","Playground","RealWorldExamples","Accessibility"];export{p as Accessibility,i as Alignment,n as Default,r as Gaps,s as Horizontal,d as Justify,o as Nested,l as Playground,m as RealWorldExamples,c as Responsive,Ae as __namedExportsOrder,Ne as default};
