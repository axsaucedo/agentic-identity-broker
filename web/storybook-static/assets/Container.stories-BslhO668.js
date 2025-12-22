import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as me}from"./index-ClcD9ViR.js";import{c as pe,a as xe}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const he=xe("w-full",{variants:{size:{sm:"max-w-screen-sm",md:"max-w-screen-md",lg:"max-w-screen-lg",xl:"max-w-screen-xl",full:"max-w-full"},centered:{true:"mx-auto",false:""},padding:{false:"",true:"px-4 py-4",sm:"px-3 py-2",md:"px-4 py-4",lg:"px-6 py-6"}},defaultVariants:{size:"lg",centered:!0,padding:!1}}),n=me.forwardRef(({size:t,centered:h,padding:re,className:de,children:le,...oe},ce)=>e.jsx("div",{ref:ce,className:pe(he({size:t,centered:h,padding:re}),de),...oe,children:le}));n.displayName="Container";n.__docgenInfo={description:`Container component for constraining content width.
Provides consistent max-width constraints and centering.

@example
\`\`\`tsx
// Default centered container with large max-width
<Container>
  <h1>Page Title</h1>
  <p>Content goes here...</p>
</Container>

// Small container with padding
<Container size="sm" padding="md">
  <p>Compact content</p>
</Container>

// Full width, left-aligned
<Container size="full" centered={false}>
  <nav>Navigation items</nav>
</Container>

// Nested containers for complex layouts
<Container size="xl">
  <Container size="md" padding="lg">
    <article>Nested content</article>
  </Container>
</Container>
\`\`\``,methods:[],displayName:"Container",props:{size:{required:!1,tsType:{name:"union",raw:"'sm' | 'md' | 'lg' | 'xl' | 'full'",elements:[{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"},{name:"literal",value:"'lg'"},{name:"literal",value:"'xl'"},{name:"literal",value:"'full'"}]},description:"Maximum width constraint"},centered:{required:!1,tsType:{name:"boolean"},description:"Center content horizontally"},padding:{required:!1,tsType:{name:"union",raw:"boolean | 'sm' | 'md' | 'lg'",elements:[{name:"boolean"},{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"},{name:"literal",value:"'lg'"}]},description:"Internal padding - boolean or size variant"},children:{required:!0,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Content to render inside container"}},composes:["VariantProps"]};const ye={title:"Design System/Layout/Container",component:n,parameters:{layout:"fullscreen"},tags:["autodocs"],argTypes:{size:{control:"select",options:["sm","md","lg","xl","full"],description:"Maximum width constraint",table:{type:{summary:"string"},defaultValue:{summary:"lg"}}},centered:{control:"boolean",description:"Center content horizontally",table:{type:{summary:"boolean"},defaultValue:{summary:"true"}}},padding:{control:"select",options:[!1,!0,"sm","md","lg"],description:"Internal padding",table:{type:{summary:'boolean | "sm" | "md" | "lg"'},defaultValue:{summary:"false"}}},children:{control:"text",description:"Container content"}}},a=({children:t,label:h})=>e.jsxs("div",{className:"bg-cream min-h-[200px] py-8",children:[h&&e.jsx("div",{className:"text-center mb-4",children:e.jsx("span",{className:"inline-block px-3 py-1 text-xs font-medium bg-navy-100 text-navy-900 rounded-full",children:h})}),e.jsxs("div",{className:"relative",children:[e.jsx("div",{className:"absolute inset-0 bg-grid-pattern opacity-5 pointer-events-none"}),t]})]}),s=()=>e.jsxs("div",{className:"bg-white rounded-lg shadow-md-premium p-6 border border-slate",children:[e.jsx("h2",{className:"text-2xl font-display font-semibold text-navy-900 mb-4",children:"Sample Content"}),e.jsx("p",{className:"text-base text-secondary-600 leading-relaxed mb-4",children:"This is sample content to demonstrate the container behavior. The container constrains the maximum width and can optionally center content horizontally."}),e.jsx("p",{className:"text-base text-secondary-600 leading-relaxed",children:"Containers are essential for creating readable layouts and maintaining consistent spacing across different screen sizes."})]}),i={args:{size:"lg",centered:!0,padding:!1,children:e.jsx(s,{})},render:t=>e.jsx(a,{label:"Default Container (lg, centered)",children:e.jsx(n,{...t})})},r={render:()=>e.jsxs("div",{className:"space-y-8 bg-cream py-8",children:[e.jsx(a,{label:"Small (640px)",children:e.jsx(n,{size:"sm",children:e.jsx(s,{})})}),e.jsx(a,{label:"Medium (768px)",children:e.jsx(n,{size:"md",children:e.jsx(s,{})})}),e.jsx(a,{label:"Large (1024px) - Default",children:e.jsx(n,{size:"lg",children:e.jsx(s,{})})}),e.jsx(a,{label:"Extra Large (1280px)",children:e.jsx(n,{size:"xl",children:e.jsx(s,{})})}),e.jsx(a,{label:"Full Width (100%)",children:e.jsx(n,{size:"full",children:e.jsx(s,{})})})]})},d={render:()=>e.jsxs("div",{className:"space-y-8",children:[e.jsx(a,{label:"Centered (default)",children:e.jsx(n,{size:"md",centered:!0,children:e.jsx(s,{})})}),e.jsx(a,{label:"Left-aligned",children:e.jsx(n,{size:"md",centered:!1,children:e.jsx(s,{})})})]})},l={render:()=>e.jsxs("div",{className:"space-y-8",children:[e.jsx(a,{label:"No Padding (default)",children:e.jsx(n,{size:"md",children:e.jsx(s,{})})}),e.jsx(a,{label:"Small Padding (px-3 py-2)",children:e.jsx(n,{size:"md",padding:"sm",children:e.jsx("div",{className:"bg-emerald-50 rounded-lg border-2 border-dashed border-emerald-300",children:e.jsx(s,{})})})}),e.jsx(a,{label:"Medium Padding (px-4 py-4)",children:e.jsx(n,{size:"md",padding:"md",children:e.jsx("div",{className:"bg-amber-50 rounded-lg border-2 border-dashed border-amber-300",children:e.jsx(s,{})})})}),e.jsx(a,{label:"Large Padding (px-6 py-6)",children:e.jsx(n,{size:"md",padding:"lg",children:e.jsx("div",{className:"bg-navy-50 rounded-lg border-2 border-dashed border-navy-300",children:e.jsx(s,{})})})})]})},o={render:()=>e.jsx(a,{label:"Outer Container (xl) with Inner Container (md)",children:e.jsxs(n,{size:"xl",padding:"lg",className:"bg-sand rounded-xl",children:[e.jsxs("div",{className:"mb-6",children:[e.jsx("h1",{className:"text-3xl font-display font-bold text-navy-900 mb-2",children:"Page with Nested Container"}),e.jsx("p",{className:"text-secondary-600",children:"This outer container spans 1280px max width"})]}),e.jsxs(n,{size:"md",padding:"md",className:"bg-white rounded-lg shadow-md-premium border border-slate",children:[e.jsx("h2",{className:"text-xl font-display font-semibold text-navy-900 mb-3",children:"Nested Inner Container"}),e.jsx("p",{className:"text-secondary-600 mb-3",children:"This inner container has a medium width constraint (768px), creating a narrower reading area within the wider page."}),e.jsx("p",{className:"text-secondary-600",children:"This pattern is useful for:"}),e.jsxs("ul",{className:"list-disc list-inside text-secondary-600 mt-2 space-y-1",children:[e.jsx("li",{children:"Long-form articles within wide layouts"}),e.jsx("li",{children:"Forms within dashboard pages"}),e.jsx("li",{children:"Focused content within marketing pages"})]})]})]})})},c={render:()=>e.jsxs("div",{className:"space-y-8 bg-cream p-4",children:[e.jsxs("div",{className:"bg-white rounded-lg p-6 shadow-md-premium",children:[e.jsx("h2",{className:"text-xl font-display font-semibold text-navy-900 mb-4",children:"Responsive Behavior"}),e.jsx("p",{className:"text-secondary-600 mb-4",children:"Try resizing your browser window to see how containers adapt:"}),e.jsxs("ul",{className:"list-disc list-inside text-secondary-600 space-y-2",children:[e.jsx("li",{children:"On wide screens, container constrains to max-width"}),e.jsx("li",{children:"On narrow screens, container fills available space"}),e.jsx("li",{children:"Padding helps prevent edge-to-edge content on mobile"}),e.jsx("li",{children:"Content remains readable across all viewport sizes"})]})]}),e.jsx(n,{size:"lg",padding:"md",className:"bg-navy-50 rounded-lg",children:e.jsxs("div",{className:"bg-white rounded-lg p-6 shadow-sm",children:[e.jsx("h3",{className:"text-lg font-semibold text-navy-900 mb-2",children:"Large Container with Padding"}),e.jsx("p",{className:"text-secondary-600",children:"Max-width: 1024px on desktop, 100% on mobile with padding"})]})}),e.jsx(n,{size:"md",padding:"sm",className:"bg-emerald-50 rounded-lg",children:e.jsxs("div",{className:"bg-white rounded-lg p-6 shadow-sm",children:[e.jsx("h3",{className:"text-lg font-semibold text-navy-900 mb-2",children:"Medium Container with Small Padding"}),e.jsx("p",{className:"text-secondary-600",children:"Max-width: 768px on desktop, 100% on mobile with padding"})]})})]})},m={render:()=>e.jsxs("div",{className:"space-y-0",children:[e.jsx("div",{className:"bg-navy-900 text-white py-4",children:e.jsx(n,{size:"xl",padding:"md",children:e.jsxs("div",{className:"flex items-center justify-between",children:[e.jsx("h1",{className:"text-xl font-display font-bold",children:"Application Name"}),e.jsxs("nav",{className:"flex gap-4 text-sm",children:[e.jsx("a",{href:"#",className:"hover:text-navy-200",children:"Home"}),e.jsx("a",{href:"#",className:"hover:text-navy-200",children:"About"}),e.jsx("a",{href:"#",className:"hover:text-navy-200",children:"Contact"})]})]})})}),e.jsx("div",{className:"bg-gradient-to-br from-navy-50 to-emerald-50 py-16",children:e.jsx(n,{size:"lg",children:e.jsxs("div",{className:"text-center",children:[e.jsx("h2",{className:"text-4xl font-display font-bold text-navy-900 mb-4",children:"Welcome to Our Service"}),e.jsx("p",{className:"text-lg text-secondary-600 max-w-2xl mx-auto",children:"This hero section uses a large container to constrain the width while allowing the background to span full width."})]})})}),e.jsx("div",{className:"py-12 bg-white",children:e.jsx(n,{size:"md",padding:"md",children:e.jsxs("article",{className:"prose prose-lg max-w-none",children:[e.jsx("h2",{className:"text-2xl font-display font-semibold text-navy-900 mb-4",children:"Article Content"}),e.jsx("p",{className:"text-secondary-600 leading-relaxed mb-4",children:"Article content uses a medium container for optimal reading width. Research shows that line lengths of 60-75 characters improve readability."}),e.jsx("p",{className:"text-secondary-600 leading-relaxed",children:"The medium container (768px) naturally creates comfortable line lengths for reading, reducing eye strain and improving comprehension."})]})})}),e.jsx("div",{className:"bg-secondary-800 text-secondary-100 py-8",children:e.jsx(n,{size:"xl",padding:"md",children:e.jsx("div",{className:"text-center text-sm",children:e.jsx("p",{children:"© 2024 Your Company. All rights reserved."})})})})]})},p={args:{size:"lg",centered:!0,padding:!1,children:e.jsx(s,{})},render:t=>e.jsx(a,{label:"Interactive Playground",children:e.jsx(n,{...t})}),parameters:{docs:{description:{story:"Experiment with all container props using the controls below. Try different sizes, toggle centering, and adjust padding."}}}},x={render:()=>e.jsx("div",{className:"bg-cream p-8",children:e.jsx(n,{size:"md",padding:"md",className:"bg-white rounded-lg shadow-md-premium",children:e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{children:[e.jsx("h2",{className:"text-xl font-display font-semibold text-navy-900 mb-3",children:"Accessibility Features"}),e.jsx("p",{className:"text-secondary-600",children:"Containers are designed with accessibility in mind:"})]}),e.jsxs("div",{className:"space-y-4",children:[e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"Semantic HTML"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Uses standard div elements with no ARIA required"})]})]}),e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"Responsive Design"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Adapts to all screen sizes, from mobile to ultra-wide"})]})]}),e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"Touch-Friendly"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Padding options ensure content doesn't touch viewport edges"})]})]}),e.jsxs("div",{className:"flex items-start gap-3",children:[e.jsx("div",{className:"flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold",children:"✓"}),e.jsxs("div",{children:[e.jsx("h3",{className:"font-semibold text-navy-900",children:"No Motion Dependencies"}),e.jsx("p",{className:"text-sm text-secondary-600",children:"Pure layout component with no animations (respects prefers-reduced-motion)"})]})]})]})]})})}),parameters:{docs:{description:{story:"Containers follow accessibility best practices with semantic HTML, responsive design, and consideration for all users regardless of device or ability."}}}};var g,u,b,v,y;i.parameters={...i.parameters,docs:{...(g=i.parameters)==null?void 0:g.docs,source:{originalSource:`{
  args: {
    size: 'lg',
    centered: true,
    padding: false,
    children: <SampleContent />
  },
  render: args => <ContainerDemo label="Default Container (lg, centered)">
      <Container {...args} />
    </ContainerDemo>
}`,...(b=(u=i.parameters)==null?void 0:u.docs)==null?void 0:b.source},description:{story:`Default container with large max-width (1024px) and centered content.
This is the most common container configuration for page content.`,...(y=(v=i.parameters)==null?void 0:v.docs)==null?void 0:y.description}}};var f,N,w,j,C;r.parameters={...r.parameters,docs:{...(f=r.parameters)==null?void 0:f.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 bg-cream py-8">
      <ContainerDemo label="Small (640px)">
        <Container size="sm">
          <SampleContent />
        </Container>
      </ContainerDemo>

      <ContainerDemo label="Medium (768px)">
        <Container size="md">
          <SampleContent />
        </Container>
      </ContainerDemo>

      <ContainerDemo label="Large (1024px) - Default">
        <Container size="lg">
          <SampleContent />
        </Container>
      </ContainerDemo>

      <ContainerDemo label="Extra Large (1280px)">
        <Container size="xl">
          <SampleContent />
        </Container>
      </ContainerDemo>

      <ContainerDemo label="Full Width (100%)">
        <Container size="full">
          <SampleContent />
        </Container>
      </ContainerDemo>
    </div>
}`,...(w=(N=r.parameters)==null?void 0:N.docs)==null?void 0:w.source},description:{story:`All container sizes side-by-side for visual comparison.
Choose sizes based on content type:
- sm (640px): Compact content, forms, sidebars
- md (768px): Articles, blog posts, narrow content
- lg (1024px): Standard page content (default)
- xl (1280px): Wide dashboards, marketing pages
- full: No constraint, spans full viewport`,...(C=(j=r.parameters)==null?void 0:j.docs)==null?void 0:C.description}}};var z,D,S,T,P;d.parameters={...d.parameters,docs:{...(z=d.parameters)==null?void 0:z.docs,source:{originalSource:`{
  render: () => <div className="space-y-8">
      <ContainerDemo label="Centered (default)">
        <Container size="md" centered={true}>
          <SampleContent />
        </Container>
      </ContainerDemo>

      <ContainerDemo label="Left-aligned">
        <Container size="md" centered={false}>
          <SampleContent />
        </Container>
      </ContainerDemo>
    </div>
}`,...(S=(D=d.parameters)==null?void 0:D.docs)==null?void 0:S.source},description:{story:`Comparison of centered vs left-aligned containers.
Centered is default and recommended for most content.
Left-aligned is useful for navigation bars and full-width sections.`,...(P=(T=d.parameters)==null?void 0:T.docs)==null?void 0:P.description}}};var A,R,k,M,L;l.parameters={...l.parameters,docs:{...(A=l.parameters)==null?void 0:A.docs,source:{originalSource:`{
  render: () => <div className="space-y-8">
      <ContainerDemo label="No Padding (default)">
        <Container size="md">
          <SampleContent />
        </Container>
      </ContainerDemo>

      <ContainerDemo label="Small Padding (px-3 py-2)">
        <Container size="md" padding="sm">
          <div className="bg-emerald-50 rounded-lg border-2 border-dashed border-emerald-300">
            <SampleContent />
          </div>
        </Container>
      </ContainerDemo>

      <ContainerDemo label="Medium Padding (px-4 py-4)">
        <Container size="md" padding="md">
          <div className="bg-amber-50 rounded-lg border-2 border-dashed border-amber-300">
            <SampleContent />
          </div>
        </Container>
      </ContainerDemo>

      <ContainerDemo label="Large Padding (px-6 py-6)">
        <Container size="md" padding="lg">
          <div className="bg-navy-50 rounded-lg border-2 border-dashed border-navy-300">
            <SampleContent />
          </div>
        </Container>
      </ContainerDemo>
    </div>
}`,...(k=(R=l.parameters)==null?void 0:R.docs)==null?void 0:k.source},description:{story:`Container padding variants.
Padding adds internal spacing and is useful for:
- Mobile responsiveness (prevents content from touching edges)
- Creating visual separation
- Section backgrounds`,...(L=(M=l.parameters)==null?void 0:M.docs)==null?void 0:L.description}}};var F,I,O,H,E;o.parameters={...o.parameters,docs:{...(F=o.parameters)==null?void 0:F.docs,source:{originalSource:`{
  render: () => <ContainerDemo label="Outer Container (xl) with Inner Container (md)">
      <Container size="xl" padding="lg" className="bg-sand rounded-xl">
        <div className="mb-6">
          <h1 className="text-3xl font-display font-bold text-navy-900 mb-2">
            Page with Nested Container
          </h1>
          <p className="text-secondary-600">
            This outer container spans 1280px max width
          </p>
        </div>

        <Container size="md" padding="md" className="bg-white rounded-lg shadow-md-premium border border-slate">
          <h2 className="text-xl font-display font-semibold text-navy-900 mb-3">
            Nested Inner Container
          </h2>
          <p className="text-secondary-600 mb-3">
            This inner container has a medium width constraint (768px),
            creating a narrower reading area within the wider page.
          </p>
          <p className="text-secondary-600">
            This pattern is useful for:
          </p>
          <ul className="list-disc list-inside text-secondary-600 mt-2 space-y-1">
            <li>Long-form articles within wide layouts</li>
            <li>Forms within dashboard pages</li>
            <li>Focused content within marketing pages</li>
          </ul>
        </Container>
      </Container>
    </ContainerDemo>
}`,...(O=(I=o.parameters)==null?void 0:I.docs)==null?void 0:O.source},description:{story:`Nested containers for complex layouts.
Outer container provides overall page width constraint,
inner containers can further constrain specific sections.`,...(E=(H=o.parameters)==null?void 0:H.docs)==null?void 0:E.description}}};var W,q,V,_,B;c.parameters={...c.parameters,docs:{...(W=c.parameters)==null?void 0:W.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 bg-cream p-4">
      <div className="bg-white rounded-lg p-6 shadow-md-premium">
        <h2 className="text-xl font-display font-semibold text-navy-900 mb-4">
          Responsive Behavior
        </h2>
        <p className="text-secondary-600 mb-4">
          Try resizing your browser window to see how containers adapt:
        </p>
        <ul className="list-disc list-inside text-secondary-600 space-y-2">
          <li>On wide screens, container constrains to max-width</li>
          <li>On narrow screens, container fills available space</li>
          <li>Padding helps prevent edge-to-edge content on mobile</li>
          <li>Content remains readable across all viewport sizes</li>
        </ul>
      </div>

      <Container size="lg" padding="md" className="bg-navy-50 rounded-lg">
        <div className="bg-white rounded-lg p-6 shadow-sm">
          <h3 className="text-lg font-semibold text-navy-900 mb-2">
            Large Container with Padding
          </h3>
          <p className="text-secondary-600">
            Max-width: 1024px on desktop, 100% on mobile with padding
          </p>
        </div>
      </Container>

      <Container size="md" padding="sm" className="bg-emerald-50 rounded-lg">
        <div className="bg-white rounded-lg p-6 shadow-sm">
          <h3 className="text-lg font-semibold text-navy-900 mb-2">
            Medium Container with Small Padding
          </h3>
          <p className="text-secondary-600">
            Max-width: 768px on desktop, 100% on mobile with padding
          </p>
        </div>
      </Container>
    </div>
}`,...(V=(q=c.parameters)==null?void 0:q.docs)==null?void 0:V.source},description:{story:`Responsive behavior demonstration.
Containers automatically adapt to smaller screens while maintaining padding
and centering. On mobile, content gracefully fills available space.`,...(B=(_=c.parameters)==null?void 0:_.docs)==null?void 0:B.description}}};var U,Y,G,J,K;m.parameters={...m.parameters,docs:{...(U=m.parameters)==null?void 0:U.docs,source:{originalSource:`{
  render: () => <div className="space-y-0">
      {/* Header with full-width background, constrained content */}
      <div className="bg-navy-900 text-white py-4">
        <Container size="xl" padding="md">
          <div className="flex items-center justify-between">
            <h1 className="text-xl font-display font-bold">Application Name</h1>
            <nav className="flex gap-4 text-sm">
              <a href="#" className="hover:text-navy-200">Home</a>
              <a href="#" className="hover:text-navy-200">About</a>
              <a href="#" className="hover:text-navy-200">Contact</a>
            </nav>
          </div>
        </Container>
      </div>

      {/* Hero section with large container */}
      <div className="bg-gradient-to-br from-navy-50 to-emerald-50 py-16">
        <Container size="lg">
          <div className="text-center">
            <h2 className="text-4xl font-display font-bold text-navy-900 mb-4">
              Welcome to Our Service
            </h2>
            <p className="text-lg text-secondary-600 max-w-2xl mx-auto">
              This hero section uses a large container to constrain the width
              while allowing the background to span full width.
            </p>
          </div>
        </Container>
      </div>

      {/* Content section with medium container */}
      <div className="py-12 bg-white">
        <Container size="md" padding="md">
          <article className="prose prose-lg max-w-none">
            <h2 className="text-2xl font-display font-semibold text-navy-900 mb-4">
              Article Content
            </h2>
            <p className="text-secondary-600 leading-relaxed mb-4">
              Article content uses a medium container for optimal reading width.
              Research shows that line lengths of 60-75 characters improve
              readability.
            </p>
            <p className="text-secondary-600 leading-relaxed">
              The medium container (768px) naturally creates comfortable line
              lengths for reading, reducing eye strain and improving comprehension.
            </p>
          </article>
        </Container>
      </div>

      {/* Footer with full-width background */}
      <div className="bg-secondary-800 text-secondary-100 py-8">
        <Container size="xl" padding="md">
          <div className="text-center text-sm">
            <p>&copy; 2024 Your Company. All rights reserved.</p>
          </div>
        </Container>
      </div>
    </div>
}`,...(G=(Y=m.parameters)==null?void 0:Y.docs)==null?void 0:G.source},description:{story:`Real-world usage examples showing common patterns.
These demonstrate how containers are typically used in applications.`,...(K=(J=m.parameters)==null?void 0:J.docs)==null?void 0:K.description}}};var Q,X,Z,$,ee;p.parameters={...p.parameters,docs:{...(Q=p.parameters)==null?void 0:Q.docs,source:{originalSource:`{
  args: {
    size: 'lg',
    centered: true,
    padding: false,
    children: <SampleContent />
  },
  render: args => <ContainerDemo label="Interactive Playground">
      <Container {...args} />
    </ContainerDemo>,
  parameters: {
    docs: {
      description: {
        story: 'Experiment with all container props using the controls below. Try different sizes, toggle centering, and adjust padding.'
      }
    }
  }
}`,...(Z=(X=p.parameters)==null?void 0:X.docs)==null?void 0:Z.source},description:{story:`Interactive playground to experiment with all props.
Try different combinations of size, centered, and padding.`,...(ee=($=p.parameters)==null?void 0:$.docs)==null?void 0:ee.description}}};var ne,ae,se,te,ie;x.parameters={...x.parameters,docs:{...(ne=x.parameters)==null?void 0:ne.docs,source:{originalSource:`{
  render: () => <div className="bg-cream p-8">
      <Container size="md" padding="md" className="bg-white rounded-lg shadow-md-premium">
        <div className="space-y-6">
          <div>
            <h2 className="text-xl font-display font-semibold text-navy-900 mb-3">
              Accessibility Features
            </h2>
            <p className="text-secondary-600">
              Containers are designed with accessibility in mind:
            </p>
          </div>

          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold">
                ✓
              </div>
              <div>
                <h3 className="font-semibold text-navy-900">Semantic HTML</h3>
                <p className="text-sm text-secondary-600">
                  Uses standard div elements with no ARIA required
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold">
                ✓
              </div>
              <div>
                <h3 className="font-semibold text-navy-900">Responsive Design</h3>
                <p className="text-sm text-secondary-600">
                  Adapts to all screen sizes, from mobile to ultra-wide
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-sm font-semibold">
                ✓
              </div>
              <div>
                <h3 className="font-semibold text-navy-900">Touch-Friendly</h3>
                <p className="text-sm text-secondary-600">
                  Padding options ensure content doesn't touch viewport edges
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
          </div>
        </div>
      </Container>
    </div>,
  parameters: {
    docs: {
      description: {
        story: 'Containers follow accessibility best practices with semantic HTML, responsive design, and consideration for all users regardless of device or ability.'
      }
    }
  }
}`,...(se=(ae=x.parameters)==null?void 0:ae.docs)==null?void 0:se.source},description:{story:`Accessibility features demonstration.
Containers use semantic HTML and support:
- Proper document structure
- Responsive design for all devices
- Sufficient spacing for touch targets
- No reliance on specific viewport sizes`,...(ie=(te=x.parameters)==null?void 0:te.docs)==null?void 0:ie.description}}};const fe=["Default","Sizes","Centered","Padding","Nested","Responsive","RealWorldExamples","Playground","Accessibility"];export{x as Accessibility,d as Centered,i as Default,o as Nested,l as Padding,p as Playground,m as RealWorldExamples,c as Responsive,r as Sizes,fe as __namedExportsOrder,ye as default};
