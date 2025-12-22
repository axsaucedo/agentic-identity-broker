import{j as e}from"./jsx-runtime-BYYWji4R.js";import{B as t}from"./Button-DlbfybpJ.js";import"./index-ClcD9ViR.js";import"./_commonjsHelpers-Cpj98o6Y.js";import"./cn-JCLedEej.js";const pe={title:"Design System/Primitives/Button",component:t,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{variant:{control:"select",options:["primary","secondary","outline","ghost","danger"],description:"Visual style variant",table:{type:{summary:"string"},defaultValue:{summary:"primary"}}},size:{control:"select",options:["sm","md","lg","xl"],description:"Button size",table:{type:{summary:"string"},defaultValue:{summary:"md"}}},isLoading:{control:"boolean",description:"Show loading spinner"},disabled:{control:"boolean",description:"Disable button interactions"},fullWidth:{control:"boolean",description:"Make button full width"},children:{control:"text",description:"Button content"}}},n={args:{children:"Primary Button",variant:"primary",size:"md"}},a={render:()=>e.jsxs("div",{className:"flex flex-wrap gap-4",children:[e.jsx(t,{variant:"primary",children:"Primary"}),e.jsx(t,{variant:"secondary",children:"Secondary"}),e.jsx(t,{variant:"outline",children:"Outline"}),e.jsx(t,{variant:"ghost",children:"Ghost"}),e.jsx(t,{variant:"danger",children:"Danger"})]})},s={render:()=>e.jsxs("div",{className:"flex flex-wrap items-center gap-4",children:[e.jsx(t,{size:"sm",children:"Small"}),e.jsx(t,{size:"md",children:"Medium"}),e.jsx(t,{size:"lg",children:"Large"}),e.jsx(t,{size:"xl",children:"Extra Large"})]})},r={render:()=>e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Normal"}),e.jsxs("div",{className:"flex gap-4",children:[e.jsx(t,{variant:"primary",children:"Primary"}),e.jsx(t,{variant:"secondary",children:"Secondary"}),e.jsx(t,{variant:"outline",children:"Outline"})]})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Disabled"}),e.jsxs("div",{className:"flex gap-4",children:[e.jsx(t,{variant:"primary",disabled:!0,children:"Primary"}),e.jsx(t,{variant:"secondary",disabled:!0,children:"Secondary"}),e.jsx(t,{variant:"outline",disabled:!0,children:"Outline"})]})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Loading"}),e.jsxs("div",{className:"flex gap-4",children:[e.jsx(t,{variant:"primary",isLoading:!0,children:"Primary"}),e.jsx(t,{variant:"secondary",isLoading:!0,children:"Secondary"}),e.jsx(t,{variant:"outline",isLoading:!0,children:"Outline"})]})]})]})},i={render:()=>e.jsxs("div",{className:"flex flex-wrap gap-4",children:[e.jsx(t,{variant:"primary",iconBefore:e.jsx("svg",{className:"w-5 h-5",fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M12 4v16m8-8H4"})}),children:"Add New"}),e.jsx(t,{variant:"secondary",iconBefore:e.jsx("svg",{className:"w-5 h-5",fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M5 13l4 4L19 7"})}),children:"Approve"}),e.jsx(t,{variant:"outline",iconAfter:e.jsx("svg",{className:"w-5 h-5",fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M9 5l7 7-7 7"})}),children:"Next"}),e.jsx(t,{variant:"danger",iconBefore:e.jsx("svg",{className:"w-5 h-5",fill:"none",stroke:"currentColor",viewBox:"0 0 24 24",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"})}),children:"Delete"})]})},o={render:()=>e.jsxs("div",{className:"w-96 space-y-4",children:[e.jsx(t,{variant:"primary",fullWidth:!0,children:"Full Width Primary"}),e.jsx(t,{variant:"outline",fullWidth:!0,children:"Full Width Outline"}),e.jsx(t,{variant:"secondary",fullWidth:!0,isLoading:!0,children:"Full Width Loading"})]})},d={render:()=>e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Loading with text"}),e.jsxs("div",{className:"flex gap-4",children:[e.jsx(t,{variant:"primary",isLoading:!0,children:"Saving..."}),e.jsx(t,{variant:"secondary",isLoading:!0,children:"Processing..."}),e.jsx(t,{variant:"outline",isLoading:!0,children:"Loading..."})]})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Different sizes"}),e.jsxs("div",{className:"flex items-center gap-4",children:[e.jsx(t,{size:"sm",isLoading:!0,children:"Small"}),e.jsx(t,{size:"md",isLoading:!0,children:"Medium"}),e.jsx(t,{size:"lg",isLoading:!0,children:"Large"}),e.jsx(t,{size:"xl",isLoading:!0,children:"Extra Large"})]})]})]})},l={render:()=>e.jsxs("div",{className:"flex gap-4",children:[e.jsx(t,{variant:"danger",children:"Delete Account"}),e.jsx(t,{variant:"danger",size:"sm",children:"Remove"}),e.jsx(t,{variant:"danger",isLoading:!0,children:"Revoking..."})]})},c={args:{children:"Click me",variant:"primary",size:"md",isLoading:!1,disabled:!1,fullWidth:!1},parameters:{docs:{description:{story:"Experiment with all button props using the controls below. Try different variants, sizes, and states."}}}},u={render:()=>e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Keyboard Navigation"}),e.jsx("p",{className:"text-sm text-neutral-600 mb-4",children:"Press Tab to focus buttons, Enter or Space to activate"}),e.jsxs("div",{className:"flex gap-4",children:[e.jsx(t,{variant:"primary",children:"First Button"}),e.jsx(t,{variant:"secondary",children:"Second Button"}),e.jsx(t,{variant:"outline",children:"Third Button"})]})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Focus Indicators"}),e.jsx("p",{className:"text-sm text-neutral-600 mb-4",children:"Visible focus rings meet WCAG 2.1 AA contrast requirements"}),e.jsxs("div",{className:"flex gap-4",children:[e.jsx(t,{variant:"primary",children:"Focus Me"}),e.jsx(t,{variant:"danger",children:"Focus Me Too"})]})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Disabled State"}),e.jsx("p",{className:"text-sm text-neutral-600 mb-4",children:"Disabled buttons cannot receive focus or be activated"}),e.jsxs("div",{className:"flex gap-4",children:[e.jsx(t,{variant:"primary",disabled:!0,children:"Disabled Button"}),e.jsx(t,{variant:"secondary",disabled:!0,children:"Also Disabled"})]})]})]}),parameters:{docs:{description:{story:"Buttons are fully accessible with keyboard support, visible focus indicators, and proper ARIA attributes. All interactive states are keyboard accessible."}}}};var m,p,v,x,h;n.parameters={...n.parameters,docs:{...(m=n.parameters)==null?void 0:m.docs,source:{originalSource:`{
  args: {
    children: 'Primary Button',
    variant: 'primary',
    size: 'md'
  }
}`,...(v=(p=n.parameters)==null?void 0:p.docs)==null?void 0:v.source},description:{story:`Default button with primary variant and medium size.
This is the most common button style used for primary actions.`,...(h=(x=n.parameters)==null?void 0:x.docs)==null?void 0:h.description}}};var g,y,f,B,b;a.parameters={...a.parameters,docs:{...(g=a.parameters)==null?void 0:g.docs,source:{originalSource:`{
  render: () => <div className="flex flex-wrap gap-4">
      <Button variant="primary">Primary</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="danger">Danger</Button>
    </div>
}`,...(f=(y=a.parameters)==null?void 0:y.docs)==null?void 0:f.source},description:{story:`All button variants side-by-side for visual comparison.
Choose variants based on action importance:
- Primary: Most important actions (submit, confirm)
- Secondary: Success/positive actions (approve, grant)
- Outline: Tertiary actions (view details, learn more)
- Ghost: Subtle actions (cancel, dismiss)
- Danger: Destructive actions (delete, revoke)`,...(b=(B=a.parameters)==null?void 0:B.docs)==null?void 0:b.description}}};var j,N,L,w,k;s.parameters={...s.parameters,docs:{...(j=s.parameters)==null?void 0:j.docs,source:{originalSource:`{
  render: () => <div className="flex flex-wrap items-center gap-4">
      <Button size="sm">Small</Button>
      <Button size="md">Medium</Button>
      <Button size="lg">Large</Button>
      <Button size="xl">Extra Large</Button>
    </div>
}`,...(L=(N=s.parameters)==null?void 0:N.docs)==null?void 0:L.source},description:{story:`All button sizes from sm to xl.
- sm: Compact spaces, secondary actions
- md: Default size for most actions
- lg: Prominent CTAs, landing pages
- xl: Hero sections, high-impact actions`,...(k=(w=s.parameters)==null?void 0:w.docs)==null?void 0:k.description}}};var S,z,A,D,W;r.parameters={...r.parameters,docs:{...(S=r.parameters)==null?void 0:S.docs,source:{originalSource:`{
  render: () => <div className="space-y-6">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">Normal</p>
        <div className="flex gap-4">
          <Button variant="primary">Primary</Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="outline">Outline</Button>
        </div>
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">Disabled</p>
        <div className="flex gap-4">
          <Button variant="primary" disabled>
            Primary
          </Button>
          <Button variant="secondary" disabled>
            Secondary
          </Button>
          <Button variant="outline" disabled>
            Outline
          </Button>
        </div>
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">Loading</p>
        <div className="flex gap-4">
          <Button variant="primary" isLoading>
            Primary
          </Button>
          <Button variant="secondary" isLoading>
            Secondary
          </Button>
          <Button variant="outline" isLoading>
            Outline
          </Button>
        </div>
      </div>
    </div>
}`,...(A=(z=r.parameters)==null?void 0:z.docs)==null?void 0:A.source},description:{story:`Button states: normal, disabled, and loading.
Loading state shows spinner and disables interactions.
Disabled state reduces opacity and prevents clicks.`,...(W=(D=r.parameters)==null?void 0:D.docs)==null?void 0:W.description}}};var P,M,F,C,T;i.parameters={...i.parameters,docs:{...(P=i.parameters)==null?void 0:P.docs,source:{originalSource:`{
  render: () => <div className="flex flex-wrap gap-4">
      <Button variant="primary" iconBefore={<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>}>
        Add New
      </Button>

      <Button variant="secondary" iconBefore={<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>}>
        Approve
      </Button>

      <Button variant="outline" iconAfter={<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>}>
        Next
      </Button>

      <Button variant="danger" iconBefore={<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>}>
        Delete
      </Button>
    </div>
}`,...(F=(M=i.parameters)==null?void 0:M.docs)==null?void 0:F.source},description:{story:`Buttons with icons before or after text.
Icons add visual clarity and help users quickly identify actions.`,...(T=(C=i.parameters)==null?void 0:C.docs)==null?void 0:T.description}}};var O,E,V,I,R;o.parameters={...o.parameters,docs:{...(O=o.parameters)==null?void 0:O.docs,source:{originalSource:`{
  render: () => <div className="w-96 space-y-4">
      <Button variant="primary" fullWidth>
        Full Width Primary
      </Button>
      <Button variant="outline" fullWidth>
        Full Width Outline
      </Button>
      <Button variant="secondary" fullWidth isLoading>
        Full Width Loading
      </Button>
    </div>
}`,...(V=(E=o.parameters)==null?void 0:E.docs)==null?void 0:V.source},description:{story:`Full width buttons that span the container width.
Useful for mobile layouts and form submissions.`,...(R=(I=o.parameters)==null?void 0:I.docs)==null?void 0:R.description}}};var G,H,q,K,U;d.parameters={...d.parameters,docs:{...(G=d.parameters)==null?void 0:G.docs,source:{originalSource:`{
  render: () => <div className="space-y-6">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Loading with text
        </p>
        <div className="flex gap-4">
          <Button variant="primary" isLoading>
            Saving...
          </Button>
          <Button variant="secondary" isLoading>
            Processing...
          </Button>
          <Button variant="outline" isLoading>
            Loading...
          </Button>
        </div>
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Different sizes
        </p>
        <div className="flex items-center gap-4">
          <Button size="sm" isLoading>
            Small
          </Button>
          <Button size="md" isLoading>
            Medium
          </Button>
          <Button size="lg" isLoading>
            Large
          </Button>
          <Button size="xl" isLoading>
            Extra Large
          </Button>
        </div>
      </div>
    </div>
}`,...(q=(H=d.parameters)==null?void 0:H.docs)==null?void 0:q.source},description:{story:`Loading states showing spinner placement.
The button remains in its original size to prevent layout shift.`,...(U=(K=d.parameters)==null?void 0:K.docs)==null?void 0:U.description}}};var _,J,Q,X,Y;l.parameters={...l.parameters,docs:{...(_=l.parameters)==null?void 0:_.docs,source:{originalSource:`{
  render: () => <div className="flex gap-4">
      <Button variant="danger">Delete Account</Button>
      <Button variant="danger" size="sm">
        Remove
      </Button>
      <Button variant="danger" isLoading>
        Revoking...
      </Button>
    </div>
}`,...(Q=(J=l.parameters)==null?void 0:J.docs)==null?void 0:Q.source},description:{story:`Danger variant for destructive actions.
Use sparingly and always with confirmation dialogs.`,...(Y=(X=l.parameters)==null?void 0:X.docs)==null?void 0:Y.description}}};var Z,$,ee,te,ne;c.parameters={...c.parameters,docs:{...(Z=c.parameters)==null?void 0:Z.docs,source:{originalSource:`{
  args: {
    children: 'Click me',
    variant: 'primary',
    size: 'md',
    isLoading: false,
    disabled: false,
    fullWidth: false
  },
  parameters: {
    docs: {
      description: {
        story: 'Experiment with all button props using the controls below. Try different variants, sizes, and states.'
      }
    }
  }
}`,...(ee=($=c.parameters)==null?void 0:$.docs)==null?void 0:ee.source},description:{story:`Interactive playground to experiment with all props.
Try different combinations of variant, size, loading, and disabled states.`,...(ne=(te=c.parameters)==null?void 0:te.docs)==null?void 0:ne.description}}};var ae,se,re,ie,oe;u.parameters={...u.parameters,docs:{...(ae=u.parameters)==null?void 0:ae.docs,source:{originalSource:`{
  render: () => <div className="space-y-6">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Keyboard Navigation
        </p>
        <p className="text-sm text-neutral-600 mb-4">
          Press Tab to focus buttons, Enter or Space to activate
        </p>
        <div className="flex gap-4">
          <Button variant="primary">First Button</Button>
          <Button variant="secondary">Second Button</Button>
          <Button variant="outline">Third Button</Button>
        </div>
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Focus Indicators
        </p>
        <p className="text-sm text-neutral-600 mb-4">
          Visible focus rings meet WCAG 2.1 AA contrast requirements
        </p>
        <div className="flex gap-4">
          <Button variant="primary">Focus Me</Button>
          <Button variant="danger">Focus Me Too</Button>
        </div>
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Disabled State
        </p>
        <p className="text-sm text-neutral-600 mb-4">
          Disabled buttons cannot receive focus or be activated
        </p>
        <div className="flex gap-4">
          <Button variant="primary" disabled>
            Disabled Button
          </Button>
          <Button variant="secondary" disabled>
            Also Disabled
          </Button>
        </div>
      </div>
    </div>,
  parameters: {
    docs: {
      description: {
        story: 'Buttons are fully accessible with keyboard support, visible focus indicators, and proper ARIA attributes. All interactive states are keyboard accessible.'
      }
    }
  }
}`,...(re=(se=u.parameters)==null?void 0:se.docs)==null?void 0:re.source},description:{story:`Accessibility features demonstration.
All buttons support:
- Keyboard navigation (Tab to focus, Enter/Space to activate)
- Screen reader labels
- Visible focus indicators (ring on focus)
- Disabled state prevents interaction`,...(oe=(ie=u.parameters)==null?void 0:ie.docs)==null?void 0:oe.description}}};const ve=["Default","Variants","Sizes","States","WithIcons","FullWidth","LoadingStates","DangerActions","Playground","Accessibility"];export{u as Accessibility,l as DangerActions,n as Default,o as FullWidth,d as LoadingStates,c as Playground,s as Sizes,r as States,a as Variants,i as WithIcons,ve as __namedExportsOrder,pe as default};
