import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as ne}from"./index-ClcD9ViR.js";import{c as re,a as de}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const ae=de("grid",{variants:{columns:{1:"grid-cols-1",2:"grid-cols-2",3:"grid-cols-3",4:"grid-cols-4",5:"grid-cols-5",6:"grid-cols-6"},gap:{sm:"gap-3",md:"gap-4",lg:"gap-6"}},defaultVariants:{columns:3,gap:"md"}}),t=ne.forwardRef(({columns:n,gap:r,className:x,children:u,...te},se)=>e.jsx("div",{ref:se,className:re(ae({columns:n,gap:r}),x),...te,children:u}));t.displayName="Grid";t.__docgenInfo={description:`Grid component for CSS Grid layouts.
Provides consistent column structure and spacing for child elements.

@example
\`\`\`tsx
// Default 3-column grid with medium gap
<Grid>
  <div>Item 1</div>
  <div>Item 2</div>
  <div>Item 3</div>
</Grid>

// 4-column grid with large gap
<Grid columns={4} gap="lg">
  <Card>Card 1</Card>
  <Card>Card 2</Card>
  <Card>Card 3</Card>
  <Card>Card 4</Card>
</Grid>

// Responsive grid with breakpoint classes
<Grid columns={1} className="sm:grid-cols-2 lg:grid-cols-4">
  <div>Responsive Item 1</div>
  <div>Responsive Item 2</div>
  <div>Responsive Item 3</div>
  <div>Responsive Item 4</div>
</Grid>

// Auto-fit columns with minimum width
<Grid className="grid-cols-[repeat(auto-fit,minmax(200px,1fr))]" gap="sm">
  <div>Auto Item 1</div>
  <div>Auto Item 2</div>
  <div>Auto Item 3</div>
</Grid>

// Card grid example
<Grid columns={3} gap="md">
  <Card title="Feature 1" />
  <Card title="Feature 2" />
  <Card title="Feature 3" />
</Grid>
\`\`\``,methods:[],displayName:"Grid",props:{columns:{required:!1,tsType:{name:"union",raw:"1 | 2 | 3 | 4 | 5 | 6",elements:[{name:"literal",value:"1"},{name:"literal",value:"2"},{name:"literal",value:"3"},{name:"literal",value:"4"},{name:"literal",value:"5"},{name:"literal",value:"6"}]},description:"Number of columns (1-6)"},gap:{required:!1,tsType:{name:"union",raw:"'sm' | 'md' | 'lg'",elements:[{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"},{name:"literal",value:"'lg'"}]},description:"Gap size between items"},children:{required:!0,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Grid content"}},composes:["VariantProps"]};const pe={title:"Design System/Layout/Grid",component:t,parameters:{layout:"padded"},tags:["autodocs"],argTypes:{columns:{control:"select",options:[1,2,3,4,5,6],description:"Number of columns (1-6)",table:{type:{summary:"number"},defaultValue:{summary:"3"}}},gap:{control:"select",options:["sm","md","lg"],description:"Gap size between items",table:{type:{summary:"string"},defaultValue:{summary:"md"}}},children:{control:"text",description:"Grid content"}}},i=({children:n,variant:r="default",tall:x=!1})=>{const u={default:"bg-navy-100 text-navy-900 border-navy-200",highlight:"bg-emerald-100 text-emerald-900 border-emerald-200",accent:"bg-sand text-secondary-900 border-slate"};return e.jsx("div",{className:`px-4 py-6 rounded-lg border-2 font-medium text-sm flex items-center justify-center text-center ${u[r]} ${x?"min-h-[120px]":"min-h-[80px]"}`,children:n})},s=({title:n,description:r})=>e.jsxs("div",{className:"bg-white rounded-lg shadow-card p-6 border-2 border-navy-100 hover:shadow-card-hover transition-shadow",children:[e.jsx("h3",{className:"text-lg font-display font-semibold text-navy-900 mb-2",children:n}),e.jsx("p",{className:"text-sm text-secondary-600",children:r})]}),d={args:{columns:3,gap:"md",children:e.jsxs(e.Fragment,{children:[e.jsx(i,{children:"Item 1"}),e.jsx(i,{children:"Item 2"}),e.jsx(i,{children:"Item 3"}),e.jsx(i,{children:"Item 4"}),e.jsx(i,{children:"Item 5"}),e.jsx(i,{children:"Item 6"})]})}},a={render:()=>e.jsxs("div",{className:"space-y-8 bg-cream p-6 rounded-xl",children:[e.jsxs("div",{children:[e.jsx("div",{className:"mb-3 text-sm font-semibold text-secondary-600",children:"1 Column - Full Width"}),e.jsxs(t,{columns:1,children:[e.jsx(i,{children:"Item 1"}),e.jsx(i,{children:"Item 2"}),e.jsx(i,{children:"Item 3"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-3 text-sm font-semibold text-secondary-600",children:"2 Columns - Split Layout"}),e.jsxs(t,{columns:2,children:[e.jsx(i,{children:"Item 1"}),e.jsx(i,{children:"Item 2"}),e.jsx(i,{children:"Item 3"}),e.jsx(i,{children:"Item 4"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-3 text-sm font-semibold text-secondary-600",children:"3 Columns - Balanced (Default)"}),e.jsxs(t,{columns:3,children:[e.jsx(i,{children:"Item 1"}),e.jsx(i,{children:"Item 2"}),e.jsx(i,{children:"Item 3"}),e.jsx(i,{children:"Item 4"}),e.jsx(i,{children:"Item 5"}),e.jsx(i,{children:"Item 6"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-3 text-sm font-semibold text-secondary-600",children:"4 Columns - Dense Grid"}),e.jsxs(t,{columns:4,children:[e.jsx(i,{children:"Item 1"}),e.jsx(i,{children:"Item 2"}),e.jsx(i,{children:"Item 3"}),e.jsx(i,{children:"Item 4"}),e.jsx(i,{children:"Item 5"}),e.jsx(i,{children:"Item 6"}),e.jsx(i,{children:"Item 7"}),e.jsx(i,{children:"Item 8"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-3 text-sm font-semibold text-secondary-600",children:"6 Columns - Maximum Density"}),e.jsxs(t,{columns:6,children:[e.jsx(i,{children:"1"}),e.jsx(i,{children:"2"}),e.jsx(i,{children:"3"}),e.jsx(i,{children:"4"}),e.jsx(i,{children:"5"}),e.jsx(i,{children:"6"}),e.jsx(i,{children:"7"}),e.jsx(i,{children:"8"}),e.jsx(i,{children:"9"}),e.jsx(i,{children:"10"}),e.jsx(i,{children:"11"}),e.jsx(i,{children:"12"})]})]})]})},l={render:()=>e.jsxs("div",{className:"space-y-8 bg-cream p-6 rounded-xl",children:[e.jsxs("div",{children:[e.jsx("div",{className:"mb-3 text-sm font-semibold text-secondary-600",children:"Small Gap (sm - 12px)"}),e.jsxs(t,{columns:3,gap:"sm",children:[e.jsx(i,{children:"Item 1"}),e.jsx(i,{children:"Item 2"}),e.jsx(i,{children:"Item 3"}),e.jsx(i,{children:"Item 4"}),e.jsx(i,{children:"Item 5"}),e.jsx(i,{children:"Item 6"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-3 text-sm font-semibold text-secondary-600",children:"Medium Gap (md - 16px) - Default"}),e.jsxs(t,{columns:3,gap:"md",children:[e.jsx(i,{children:"Item 1"}),e.jsx(i,{children:"Item 2"}),e.jsx(i,{children:"Item 3"}),e.jsx(i,{children:"Item 4"}),e.jsx(i,{children:"Item 5"}),e.jsx(i,{children:"Item 6"})]})]}),e.jsxs("div",{children:[e.jsx("div",{className:"mb-3 text-sm font-semibold text-secondary-600",children:"Large Gap (lg - 24px)"}),e.jsxs(t,{columns:3,gap:"lg",children:[e.jsx(i,{children:"Item 1"}),e.jsx(i,{children:"Item 2"}),e.jsx(i,{children:"Item 3"}),e.jsx(i,{children:"Item 4"}),e.jsx(i,{children:"Item 5"}),e.jsx(i,{children:"Item 6"})]})]})]})},m={render:()=>e.jsxs("div",{className:"space-y-8 bg-cream p-6 rounded-xl",children:[e.jsxs("div",{className:"bg-white rounded-lg p-6 shadow-md-premium",children:[e.jsx("h2",{className:"text-xl font-display font-semibold text-navy-900 mb-4",children:"Responsive Grid Behavior"}),e.jsx("p",{className:"text-secondary-600 mb-4",children:"Try resizing your browser window to see how the grid adapts:"}),e.jsxs("ul",{className:"list-disc list-inside text-secondary-600 space-y-2 mb-6",children:[e.jsx("li",{children:"Mobile (default): 1 column"}),e.jsx("li",{children:"Tablet (sm: 640px+): 2 columns"}),e.jsx("li",{children:"Desktop (lg: 1024px+): 4 columns"})]}),e.jsxs(t,{columns:1,gap:"md",className:"sm:grid-cols-2 lg:grid-cols-4",children:[e.jsx(i,{variant:"highlight",children:"Responsive 1"}),e.jsx(i,{variant:"highlight",children:"Responsive 2"}),e.jsx(i,{variant:"highlight",children:"Responsive 3"}),e.jsx(i,{variant:"highlight",children:"Responsive 4"}),e.jsx(i,{variant:"highlight",children:"Responsive 5"}),e.jsx(i,{variant:"highlight",children:"Responsive 6"}),e.jsx(i,{variant:"highlight",children:"Responsive 7"}),e.jsx(i,{variant:"highlight",children:"Responsive 8"})]})]}),e.jsxs("div",{className:"bg-white rounded-lg p-6 shadow-md-premium",children:[e.jsx("h3",{className:"text-lg font-display font-semibold text-navy-900 mb-3",children:"Mobile-First Card Grid"}),e.jsx("p",{className:"text-secondary-600 mb-4",children:"1 column on mobile, 2 on tablet, 3 on desktop:"}),e.jsxs(t,{columns:1,gap:"lg",className:"sm:grid-cols-2 md:grid-cols-3",children:[e.jsxs(i,{variant:"accent",tall:!0,children:["Card 1",e.jsx("br",{}),"Mobile First"]}),e.jsxs(i,{variant:"accent",tall:!0,children:["Card 2",e.jsx("br",{}),"Mobile First"]}),e.jsxs(i,{variant:"accent",tall:!0,children:["Card 3",e.jsx("br",{}),"Mobile First"]})]})]})]})},o={render:()=>e.jsxs("div",{className:"space-y-8 bg-cream p-6 rounded-xl",children:[e.jsxs("div",{className:"bg-white rounded-lg p-6 shadow-md-premium",children:[e.jsx("h2",{className:"text-xl font-display font-semibold text-navy-900 mb-4",children:"Auto-Fit Grid"}),e.jsx("p",{className:"text-secondary-600 mb-4",children:"Columns automatically fit based on container width. Minimum 200px, maximum 1fr."}),e.jsx("p",{className:"text-sm text-secondary-500 mb-6",children:"Try resizing to see columns adjust without media queries."}),e.jsxs(t,{gap:"md",className:"grid-cols-[repeat(auto-fit,minmax(200px,1fr))]",children:[e.jsx(i,{variant:"highlight",children:"Auto 1"}),e.jsx(i,{variant:"highlight",children:"Auto 2"}),e.jsx(i,{variant:"highlight",children:"Auto 3"}),e.jsx(i,{variant:"highlight",children:"Auto 4"}),e.jsx(i,{variant:"highlight",children:"Auto 5"}),e.jsx(i,{variant:"highlight",children:"Auto 6"})]})]}),e.jsxs("div",{className:"bg-white rounded-lg p-6 shadow-md-premium",children:[e.jsx("h3",{className:"text-lg font-display font-semibold text-navy-900 mb-3",children:"Auto-Fill Grid (Larger Minimum)"}),e.jsx("p",{className:"text-secondary-600 mb-4",children:"Minimum 250px per column - fewer columns, more breathing room."}),e.jsxs(t,{gap:"lg",className:"grid-cols-[repeat(auto-fill,minmax(250px,1fr))]",children:[e.jsx(i,{variant:"accent",children:"Auto-Fill 1"}),e.jsx(i,{variant:"accent",children:"Auto-Fill 2"}),e.jsx(i,{variant:"accent",children:"Auto-Fill 3"}),e.jsx(i,{variant:"accent",children:"Auto-Fill 4"})]})]})]})},c={render:()=>e.jsxs("div",{className:"space-y-8 bg-cream p-6 rounded-xl",children:[e.jsxs("div",{children:[e.jsx("h2",{className:"text-2xl font-display font-bold text-navy-900 mb-2",children:"Features Overview"}),e.jsx("p",{className:"text-secondary-600 mb-6",children:"Responsive card grid showcasing product features"}),e.jsxs(t,{columns:1,gap:"lg",className:"sm:grid-cols-2 lg:grid-cols-3",children:[e.jsx(s,{title:"Secure Authentication",description:"Industry-standard OAuth 2.0 and OpenID Connect protocols with enterprise-grade security."}),e.jsx(s,{title:"Fine-Grained Consent",description:"Granular permission management allowing users to control exactly what data they share."}),e.jsx(s,{title:"Multi-Provider Support",description:"Connect with multiple identity providers seamlessly with unified consent flows."}),e.jsx(s,{title:"Real-Time Monitoring",description:"Track authentication events and consent decisions in real-time with detailed analytics."}),e.jsx(s,{title:"Custom Branding",description:"White-label consent screens to match your brand identity and user experience."}),e.jsx(s,{title:"Compliance Ready",description:"Built-in GDPR, CCPA, and privacy regulation compliance with audit trails."})]})]}),e.jsxs("div",{children:[e.jsx("h2",{className:"text-2xl font-display font-bold text-navy-900 mb-2",children:"Dense Layout (4 Columns)"}),e.jsx("p",{className:"text-secondary-600 mb-6",children:"Higher density for desktop viewing"}),e.jsxs(t,{columns:4,gap:"sm",children:[e.jsx(s,{title:"Feature 1",description:"Quick overview of capability"}),e.jsx(s,{title:"Feature 2",description:"Quick overview of capability"}),e.jsx(s,{title:"Feature 3",description:"Quick overview of capability"}),e.jsx(s,{title:"Feature 4",description:"Quick overview of capability"}),e.jsx(s,{title:"Feature 5",description:"Quick overview of capability"}),e.jsx(s,{title:"Feature 6",description:"Quick overview of capability"}),e.jsx(s,{title:"Feature 7",description:"Quick overview of capability"}),e.jsx(s,{title:"Feature 8",description:"Quick overview of capability"})]})]})]})},p={args:{columns:3,gap:"md",children:e.jsxs(e.Fragment,{children:[e.jsx(i,{children:"Item 1"}),e.jsx(i,{children:"Item 2"}),e.jsx(i,{children:"Item 3"}),e.jsx(i,{children:"Item 4"}),e.jsx(i,{children:"Item 5"}),e.jsx(i,{children:"Item 6"}),e.jsx(i,{children:"Item 7"}),e.jsx(i,{children:"Item 8"}),e.jsx(i,{children:"Item 9"})]})},parameters:{docs:{description:{story:"Experiment with all grid props using the controls below. Try different column counts and gap sizes to see how they affect the layout."}}}},h={render:()=>e.jsx("div",{className:"bg-cream p-8",children:e.jsxs("div",{className:"bg-white rounded-lg shadow-md-premium p-6",children:[e.jsx("h2",{className:"text-xl font-display font-semibold text-navy-900 mb-3",children:"Accessibility Features"}),e.jsx("p",{className:"text-secondary-600 mb-6",children:"Grids are designed with accessibility in mind:"}),e.jsxs(t,{columns:1,gap:"md",className:"sm:grid-cols-2",children:[e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"Semantic HTML"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Uses standard div elements with CSS Grid, no ARIA required"})]})]}),e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"Logical Reading Order"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Content flows naturally left-to-right, top-to-bottom for screen readers"})]})]}),e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"Touch-Friendly Spacing"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Gap options ensure adequate spacing for touch targets (minimum 44x44px)"})]})]}),e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"Fully Responsive"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Adapts gracefully across all viewport sizes and zoom levels"})]})]}),e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"Keyboard Navigation"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Interactive grid children fully support keyboard navigation"})]})]}),e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"No Motion Dependencies"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Pure layout component with no animations (respects prefers-reduced-motion)"})]})]})]})]})}),parameters:{docs:{description:{story:"Grids follow accessibility best practices with semantic HTML, logical reading order, and consideration for all users regardless of device or ability."}}}};var v,g,I,G,f;d.parameters={...d.parameters,docs:{...(v=d.parameters)==null?void 0:v.docs,source:{originalSource:`{
  args: {
    columns: 3,
    gap: 'md',
    children: <>
        <GridItem>Item 1</GridItem>
        <GridItem>Item 2</GridItem>
        <GridItem>Item 3</GridItem>
        <GridItem>Item 4</GridItem>
        <GridItem>Item 5</GridItem>
        <GridItem>Item 6</GridItem>
      </>
  }
}`,...(I=(g=d.parameters)==null?void 0:g.docs)==null?void 0:I.source},description:{story:`Default 3-column grid with medium gap.
This is the most common grid configuration for balanced layouts.`,...(f=(G=d.parameters)==null?void 0:G.docs)==null?void 0:f.description}}};var y,b,j,N,w;a.parameters={...a.parameters,docs:{...(y=a.parameters)==null?void 0:y.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 bg-cream p-6 rounded-xl">
      <div>
        <div className="mb-3 text-sm font-semibold text-secondary-600">
          1 Column - Full Width
        </div>
        <Grid columns={1}>
          <GridItem>Item 1</GridItem>
          <GridItem>Item 2</GridItem>
          <GridItem>Item 3</GridItem>
        </Grid>
      </div>

      <div>
        <div className="mb-3 text-sm font-semibold text-secondary-600">
          2 Columns - Split Layout
        </div>
        <Grid columns={2}>
          <GridItem>Item 1</GridItem>
          <GridItem>Item 2</GridItem>
          <GridItem>Item 3</GridItem>
          <GridItem>Item 4</GridItem>
        </Grid>
      </div>

      <div>
        <div className="mb-3 text-sm font-semibold text-secondary-600">
          3 Columns - Balanced (Default)
        </div>
        <Grid columns={3}>
          <GridItem>Item 1</GridItem>
          <GridItem>Item 2</GridItem>
          <GridItem>Item 3</GridItem>
          <GridItem>Item 4</GridItem>
          <GridItem>Item 5</GridItem>
          <GridItem>Item 6</GridItem>
        </Grid>
      </div>

      <div>
        <div className="mb-3 text-sm font-semibold text-secondary-600">
          4 Columns - Dense Grid
        </div>
        <Grid columns={4}>
          <GridItem>Item 1</GridItem>
          <GridItem>Item 2</GridItem>
          <GridItem>Item 3</GridItem>
          <GridItem>Item 4</GridItem>
          <GridItem>Item 5</GridItem>
          <GridItem>Item 6</GridItem>
          <GridItem>Item 7</GridItem>
          <GridItem>Item 8</GridItem>
        </Grid>
      </div>

      <div>
        <div className="mb-3 text-sm font-semibold text-secondary-600">
          6 Columns - Maximum Density
        </div>
        <Grid columns={6}>
          <GridItem>1</GridItem>
          <GridItem>2</GridItem>
          <GridItem>3</GridItem>
          <GridItem>4</GridItem>
          <GridItem>5</GridItem>
          <GridItem>6</GridItem>
          <GridItem>7</GridItem>
          <GridItem>8</GridItem>
          <GridItem>9</GridItem>
          <GridItem>10</GridItem>
          <GridItem>11</GridItem>
          <GridItem>12</GridItem>
        </Grid>
      </div>
    </div>
}`,...(j=(b=a.parameters)==null?void 0:b.docs)==null?void 0:j.source},description:{story:`All column configurations for visual comparison.
Choose column count based on content density and viewport:
- 1 column: Mobile, full-width content
- 2 columns: Split layouts, tablet
- 3 columns: Balanced layouts, desktop (default)
- 4 columns: Dense grids, wide viewports
- 5 columns: Specialized layouts
- 6 columns: Maximum density, extra-wide viewports`,...(w=(N=a.parameters)==null?void 0:N.docs)==null?void 0:w.description}}};var C,F,A,R,k;l.parameters={...l.parameters,docs:{...(C=l.parameters)==null?void 0:C.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 bg-cream p-6 rounded-xl">
      <div>
        <div className="mb-3 text-sm font-semibold text-secondary-600">
          Small Gap (sm - 12px)
        </div>
        <Grid columns={3} gap="sm">
          <GridItem>Item 1</GridItem>
          <GridItem>Item 2</GridItem>
          <GridItem>Item 3</GridItem>
          <GridItem>Item 4</GridItem>
          <GridItem>Item 5</GridItem>
          <GridItem>Item 6</GridItem>
        </Grid>
      </div>

      <div>
        <div className="mb-3 text-sm font-semibold text-secondary-600">
          Medium Gap (md - 16px) - Default
        </div>
        <Grid columns={3} gap="md">
          <GridItem>Item 1</GridItem>
          <GridItem>Item 2</GridItem>
          <GridItem>Item 3</GridItem>
          <GridItem>Item 4</GridItem>
          <GridItem>Item 5</GridItem>
          <GridItem>Item 6</GridItem>
        </Grid>
      </div>

      <div>
        <div className="mb-3 text-sm font-semibold text-secondary-600">
          Large Gap (lg - 24px)
        </div>
        <Grid columns={3} gap="lg">
          <GridItem>Item 1</GridItem>
          <GridItem>Item 2</GridItem>
          <GridItem>Item 3</GridItem>
          <GridItem>Item 4</GridItem>
          <GridItem>Item 5</GridItem>
          <GridItem>Item 6</GridItem>
        </Grid>
      </div>
    </div>
}`,...(A=(F=l.parameters)==null?void 0:F.docs)==null?void 0:A.source},description:{story:`All gap sizes for visual comparison.
Choose gap sizes based on visual hierarchy and spacing needs:
- sm (12px): Tight grouping, dense layouts
- md (16px): Standard spacing (default)
- lg (24px): Generous spacing, distinct items`,...(k=(R=l.parameters)==null?void 0:R.docs)==null?void 0:k.description}}};var D,M,S,T,z;m.parameters={...m.parameters,docs:{...(D=m.parameters)==null?void 0:D.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 bg-cream p-6 rounded-xl">
      <div className="bg-white rounded-lg p-6 shadow-md-premium">
        <h2 className="text-xl font-display font-semibold text-navy-900 mb-4">
          Responsive Grid Behavior
        </h2>
        <p className="text-secondary-600 mb-4">
          Try resizing your browser window to see how the grid adapts:
        </p>
        <ul className="list-disc list-inside text-secondary-600 space-y-2 mb-6">
          <li>Mobile (default): 1 column</li>
          <li>Tablet (sm: 640px+): 2 columns</li>
          <li>Desktop (lg: 1024px+): 4 columns</li>
        </ul>

        <Grid columns={1} gap="md" className="sm:grid-cols-2 lg:grid-cols-4">
          <GridItem variant="highlight">Responsive 1</GridItem>
          <GridItem variant="highlight">Responsive 2</GridItem>
          <GridItem variant="highlight">Responsive 3</GridItem>
          <GridItem variant="highlight">Responsive 4</GridItem>
          <GridItem variant="highlight">Responsive 5</GridItem>
          <GridItem variant="highlight">Responsive 6</GridItem>
          <GridItem variant="highlight">Responsive 7</GridItem>
          <GridItem variant="highlight">Responsive 8</GridItem>
        </Grid>
      </div>

      <div className="bg-white rounded-lg p-6 shadow-md-premium">
        <h3 className="text-lg font-display font-semibold text-navy-900 mb-3">
          Mobile-First Card Grid
        </h3>
        <p className="text-secondary-600 mb-4">
          1 column on mobile, 2 on tablet, 3 on desktop:
        </p>

        <Grid columns={1} gap="lg" className="sm:grid-cols-2 md:grid-cols-3">
          <GridItem variant="accent" tall>
            Card 1<br />Mobile First
          </GridItem>
          <GridItem variant="accent" tall>
            Card 2<br />Mobile First
          </GridItem>
          <GridItem variant="accent" tall>
            Card 3<br />Mobile First
          </GridItem>
        </Grid>
      </div>
    </div>
}`,...(S=(M=m.parameters)==null?void 0:M.docs)==null?void 0:S.source},description:{story:`Responsive grid with column changes at breakpoints.
Grid adapts from 1 column on mobile to 2 on tablet to 4 on desktop.
This pattern is essential for mobile-first responsive design.`,...(z=(T=m.parameters)==null?void 0:T.docs)==null?void 0:z.description}}};var L,Q,P,q,B;o.parameters={...o.parameters,docs:{...(L=o.parameters)==null?void 0:L.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 bg-cream p-6 rounded-xl">
      <div className="bg-white rounded-lg p-6 shadow-md-premium">
        <h2 className="text-xl font-display font-semibold text-navy-900 mb-4">
          Auto-Fit Grid
        </h2>
        <p className="text-secondary-600 mb-4">
          Columns automatically fit based on container width. Minimum 200px, maximum 1fr.
        </p>
        <p className="text-sm text-secondary-500 mb-6">
          Try resizing to see columns adjust without media queries.
        </p>

        <Grid gap="md" className="grid-cols-[repeat(auto-fit,minmax(200px,1fr))]">
          <GridItem variant="highlight">Auto 1</GridItem>
          <GridItem variant="highlight">Auto 2</GridItem>
          <GridItem variant="highlight">Auto 3</GridItem>
          <GridItem variant="highlight">Auto 4</GridItem>
          <GridItem variant="highlight">Auto 5</GridItem>
          <GridItem variant="highlight">Auto 6</GridItem>
        </Grid>
      </div>

      <div className="bg-white rounded-lg p-6 shadow-md-premium">
        <h3 className="text-lg font-display font-semibold text-navy-900 mb-3">
          Auto-Fill Grid (Larger Minimum)
        </h3>
        <p className="text-secondary-600 mb-4">
          Minimum 250px per column - fewer columns, more breathing room.
        </p>

        <Grid gap="lg" className="grid-cols-[repeat(auto-fill,minmax(250px,1fr))]">
          <GridItem variant="accent">Auto-Fill 1</GridItem>
          <GridItem variant="accent">Auto-Fill 2</GridItem>
          <GridItem variant="accent">Auto-Fill 3</GridItem>
          <GridItem variant="accent">Auto-Fill 4</GridItem>
        </Grid>
      </div>
    </div>
}`,...(P=(Q=o.parameters)==null?void 0:Q.docs)==null?void 0:P.source},description:{story:`Auto-fit columns with minimum and maximum widths.
Columns automatically adjust based on available space and content.
Useful for responsive card grids without breakpoints.`,...(B=(q=o.parameters)==null?void 0:q.docs)==null?void 0:B.description}}};var O,H,V,E,W;c.parameters={...c.parameters,docs:{...(O=c.parameters)==null?void 0:O.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 bg-cream p-6 rounded-xl">
      <div>
        <h2 className="text-2xl font-display font-bold text-navy-900 mb-2">
          Features Overview
        </h2>
        <p className="text-secondary-600 mb-6">
          Responsive card grid showcasing product features
        </p>

        <Grid columns={1} gap="lg" className="sm:grid-cols-2 lg:grid-cols-3">
          <DemoCard title="Secure Authentication" description="Industry-standard OAuth 2.0 and OpenID Connect protocols with enterprise-grade security." />
          <DemoCard title="Fine-Grained Consent" description="Granular permission management allowing users to control exactly what data they share." />
          <DemoCard title="Multi-Provider Support" description="Connect with multiple identity providers seamlessly with unified consent flows." />
          <DemoCard title="Real-Time Monitoring" description="Track authentication events and consent decisions in real-time with detailed analytics." />
          <DemoCard title="Custom Branding" description="White-label consent screens to match your brand identity and user experience." />
          <DemoCard title="Compliance Ready" description="Built-in GDPR, CCPA, and privacy regulation compliance with audit trails." />
        </Grid>
      </div>

      <div>
        <h2 className="text-2xl font-display font-bold text-navy-900 mb-2">
          Dense Layout (4 Columns)
        </h2>
        <p className="text-secondary-600 mb-6">
          Higher density for desktop viewing
        </p>

        <Grid columns={4} gap="sm">
          <DemoCard title="Feature 1" description="Quick overview of capability" />
          <DemoCard title="Feature 2" description="Quick overview of capability" />
          <DemoCard title="Feature 3" description="Quick overview of capability" />
          <DemoCard title="Feature 4" description="Quick overview of capability" />
          <DemoCard title="Feature 5" description="Quick overview of capability" />
          <DemoCard title="Feature 6" description="Quick overview of capability" />
          <DemoCard title="Feature 7" description="Quick overview of capability" />
          <DemoCard title="Feature 8" description="Quick overview of capability" />
        </Grid>
      </div>
    </div>
}`,...(V=(H=c.parameters)==null?void 0:H.docs)==null?void 0:V.source},description:{story:`Real-world card grid implementation.
Demonstrates how to use Grid with actual card components
for features, products, or content listings.`,...(W=(E=c.parameters)==null?void 0:E.docs)==null?void 0:W.description}}};var _,K,U,$,J;p.parameters={...p.parameters,docs:{...(_=p.parameters)==null?void 0:_.docs,source:{originalSource:`{
  args: {
    columns: 3,
    gap: 'md',
    children: <>
        <GridItem>Item 1</GridItem>
        <GridItem>Item 2</GridItem>
        <GridItem>Item 3</GridItem>
        <GridItem>Item 4</GridItem>
        <GridItem>Item 5</GridItem>
        <GridItem>Item 6</GridItem>
        <GridItem>Item 7</GridItem>
        <GridItem>Item 8</GridItem>
        <GridItem>Item 9</GridItem>
      </>
  },
  parameters: {
    docs: {
      description: {
        story: 'Experiment with all grid props using the controls below. Try different column counts and gap sizes to see how they affect the layout.'
      }
    }
  }
}`,...(U=(K=p.parameters)==null?void 0:K.docs)==null?void 0:U.source},description:{story:`Interactive playground to experiment with all props.
Try different combinations of columns and gap sizes.`,...(J=($=p.parameters)==null?void 0:$.docs)==null?void 0:J.description}}};var X,Y,Z,ee,ie;h.parameters={...h.parameters,docs:{...(X=h.parameters)==null?void 0:X.docs,source:{originalSource:`{
  render: () => <div className="bg-cream p-8">
      <div className="bg-white rounded-lg shadow-md-premium p-6">
        <h2 className="text-xl font-display font-semibold text-navy-900 mb-3">
          Accessibility Features
        </h2>
        <p className="text-secondary-600 mb-6">
          Grids are designed with accessibility in mind:
        </p>

        <Grid columns={1} gap="md" className="sm:grid-cols-2">
          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold">
              ✓
            </div>
            <div>
              <h3 className="font-semibold text-navy-900">Semantic HTML</h3>
              <p className="text-sm text-secondary-600">
                Uses standard div elements with CSS Grid, no ARIA required
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold">
              ✓
            </div>
            <div>
              <h3 className="font-semibold text-navy-900">Logical Reading Order</h3>
              <p className="text-sm text-secondary-600">
                Content flows naturally left-to-right, top-to-bottom for screen readers
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
              <h3 className="font-semibold text-navy-900">Fully Responsive</h3>
              <p className="text-sm text-secondary-600">
                Adapts gracefully across all viewport sizes and zoom levels
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold">
              ✓
            </div>
            <div>
              <h3 className="font-semibold text-navy-900">Keyboard Navigation</h3>
              <p className="text-sm text-secondary-600">
                Interactive grid children fully support keyboard navigation
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
        </Grid>
      </div>
    </div>,
  parameters: {
    docs: {
      description: {
        story: 'Grids follow accessibility best practices with semantic HTML, logical reading order, and consideration for all users regardless of device or ability.'
      }
    }
  }
}`,...(Z=(Y=h.parameters)==null?void 0:Y.docs)==null?void 0:Z.source},description:{story:`Accessibility features demonstration.
Grids use semantic HTML and support:
- Proper document structure with logical reading order
- Keyboard navigation for interactive children
- Sufficient spacing for touch targets
- Responsive behavior for all screen sizes
- Screen reader friendly markup`,...(ie=(ee=h.parameters)==null?void 0:ee.docs)==null?void 0:ie.description}}};const he=["Default","Columns","Gaps","Responsive","AutoFit","CardGrid","Playground","Accessibility"];export{h as Accessibility,o as AutoFit,c as CardGrid,a as Columns,d as Default,l as Gaps,p as Playground,m as Responsive,he as __namedExportsOrder,pe as default};
