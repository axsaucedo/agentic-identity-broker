import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as We}from"./index-ClcD9ViR.js";import{c as Te,a as h}from"./cn-JCLedEej.js";import{B as C}from"./Button-DlbfybpJ.js";import"./_commonjsHelpers-Cpj98o6Y.js";const Be=h("flex flex-col items-center justify-center text-center max-w-md mx-auto",{variants:{size:{compact:"py-6 px-4 space-y-2",default:"py-10 px-6 space-y-4",expanded:"py-16 px-8 space-y-6"}},defaultVariants:{size:"default"}}),Ve=h("flex items-center justify-center rounded-full",{variants:{size:{compact:"w-12 h-12 mb-2",default:"w-16 h-16 mb-3",expanded:"w-20 h-20 mb-4"}},defaultVariants:{size:"default"}}),De=h("font-semibold text-gray-900",{variants:{size:{compact:"text-base",default:"text-lg",expanded:"text-xl"}},defaultVariants:{size:"default"}}),Me=h("text-gray-600 leading-relaxed",{variants:{size:{compact:"text-sm",default:"text-base",expanded:"text-base"}},defaultVariants:{size:"default"}}),Ue=h("flex gap-3",{variants:{size:{compact:"flex-col w-full mt-3",default:"flex-row items-center justify-center mt-4",expanded:"flex-row items-center justify-center mt-6"}},defaultVariants:{size:"default"}}),Ye=()=>e.jsx("svg",{className:"w-full h-full text-gray-400",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:1.5,d:"M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"})}),t=We.forwardRef(({icon:b,title:qe,description:v,primaryAction:x,secondaryAction:g,size:a="default",className:Re,...Ee},Le)=>e.jsxs("div",{ref:Le,role:"status","aria-live":"polite",className:Te(Be({size:a}),Re),...Ee,children:[(b||!b)&&e.jsx("div",{className:Ve({size:a}),children:b||e.jsx(Ye,{})}),e.jsx("h3",{className:De({size:a}),children:qe}),v&&e.jsx("p",{className:Me({size:a}),children:v}),(x||g)&&e.jsxs("div",{className:Ue({size:a}),children:[x&&e.jsx(C,{variant:"primary",size:a==="compact"?"sm":"md",onClick:x.onClick,fullWidth:a==="compact",children:x.label}),g&&e.jsx(C,{variant:"outline",size:a==="compact"?"sm":"md",onClick:g.onClick,fullWidth:a==="compact",children:g.label})]})]}));t.displayName="EmptyState";t.__docgenInfo={description:`EmptyState component for displaying friendly empty state UIs.
Shows when there's no data, no search results, or user lacks permissions.

@example
\`\`\`tsx
<EmptyState
  title="No consents yet"
  description="You haven't granted any permissions to applications."
/>

<EmptyState
  icon={<SearchIcon />}
  title="No results found"
  description="Try adjusting your search or filters."
  primaryAction={{
    label: "Clear filters",
    onClick: handleClear
  }}
/>

<EmptyState
  size="expanded"
  icon={<IllustrationComponent />}
  title="Get started"
  description="Connect your first application to begin."
  primaryAction={{
    label: "Add application",
    onClick: handleAdd
  }}
  secondaryAction={{
    label: "Learn more",
    onClick: handleLearnMore
  }}
/>
\`\`\``,methods:[],displayName:"EmptyState",props:{icon:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Icon or illustration to display (decorative)"},title:{required:!0,tsType:{name:"string"},description:"Main heading text (required)"},description:{required:!1,tsType:{name:"string"},description:"Supporting description text (optional)"},primaryAction:{required:!1,tsType:{name:"signature",type:"object",raw:`{
  label: string;
  onClick: () => void;
}`,signature:{properties:[{key:"label",value:{name:"string",required:!0}},{key:"onClick",value:{name:"signature",type:"function",raw:"() => void",signature:{arguments:[],return:{name:"void"}},required:!0}}]}},description:"Primary call-to-action button"},secondaryAction:{required:!1,tsType:{name:"signature",type:"object",raw:`{
  label: string;
  onClick: () => void;
}`,signature:{properties:[{key:"label",value:{name:"string",required:!0}},{key:"onClick",value:{name:"signature",type:"function",raw:"() => void",signature:{arguments:[],return:{name:"void"}},required:!0}}]}},description:"Secondary action button"},size:{required:!1,tsType:{name:"union",raw:"'compact' | 'default' | 'expanded'",elements:[{name:"literal",value:"'compact'"},{name:"literal",value:"'default'"},{name:"literal",value:"'expanded'"}]},description:"Size variant for spacing and typography",defaultValue:{value:"'default'",computed:!1}}},composes:["Omit"]};const k=()=>e.jsx("svg",{className:"w-full h-full text-gray-400",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:1.5,d:"M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"})}),Fe=()=>e.jsx("svg",{className:"w-full h-full text-gray-400",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:1.5,d:"M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"})}),f=()=>e.jsx("svg",{className:"w-full h-full text-gray-400",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:1.5,d:"M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"})}),Ie=()=>e.jsx("svg",{className:"w-full h-full text-gray-400",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:1.5,d:"M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"})}),Ge=()=>e.jsx("svg",{className:"w-full h-full text-gray-400",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:1.5,d:"M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"})}),Pe=()=>e.jsx("svg",{className:"w-full h-full text-gray-400",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:1.5,d:"M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"})}),He=()=>e.jsxs("svg",{className:"w-full h-full text-trust-soft",viewBox:"0 0 200 200",fill:"none",xmlns:"http://www.w3.org/2000/svg","aria-hidden":"true",children:[e.jsx("circle",{cx:"100",cy:"100",r:"80",fill:"currentColor",opacity:"0.1"}),e.jsx("circle",{cx:"100",cy:"100",r:"60",fill:"currentColor",opacity:"0.2"}),e.jsx("path",{d:"M100 60v80M60 100h80",stroke:"currentColor",strokeWidth:"8",strokeLinecap:"round",opacity:"0.4"})]}),Qe={title:"Design System/Feedback/EmptyState",component:t,parameters:{layout:"padded"},tags:["autodocs"],argTypes:{title:{control:"text",description:"Main heading text (required)"},description:{control:"text",description:"Supporting description text (optional)"},size:{control:"select",options:["compact","default","expanded"],description:"Size variant for spacing and typography"},icon:{control:!1,description:"Icon or illustration to display (decorative)"},primaryAction:{control:!1,description:"Primary call-to-action button"},secondaryAction:{control:!1,description:"Secondary action button"}}},r={args:{title:"No items yet",description:"Get started by creating your first item."}},o={args:{title:"No consents found",description:"You haven't granted any permissions yet. Connect your first application to get started.",primaryAction:{label:"Add Application",onClick:()=>alert("Add application clicked")}}},s={args:{title:"Build your first workflow",description:"Workflows help automate consent management tasks. Start from scratch or use a template.",primaryAction:{label:"Create Workflow",onClick:()=>alert("Create workflow clicked")},secondaryAction:{label:"Browse Templates",onClick:()=>alert("Browse templates clicked")}}},n={args:{icon:e.jsx(k,{}),title:"No results found",description:"We couldn't find any matches for your search. Try adjusting your filters or search terms.",primaryAction:{label:"Clear Filters",onClick:()=>alert("Clear filters clicked")}}},i={args:{icon:e.jsx(Fe,{}),title:"Access restricted",description:"You don't have permission to view this content. Contact your administrator to request access.",primaryAction:{label:"Request Access",onClick:()=>alert("Request access clicked")},secondaryAction:{label:"Go Back",onClick:()=>alert("Go back clicked")}}},c={render:()=>e.jsx("div",{className:"max-w-xs border border-gray-200 rounded-lg p-4",children:e.jsx(t,{size:"compact",icon:e.jsx(Ie,{}),title:"No messages",description:"Your inbox is empty.",primaryAction:{label:"Compose",onClick:()=>alert("Compose clicked")}})}),args:{size:"compact",title:"No messages",description:"Your inbox is empty."}},l={args:{size:"expanded",icon:e.jsx(Ge,{}),title:"Welcome to Consent Manager",description:"Take control of your data privacy. Start by connecting your first application and managing consent preferences.",primaryAction:{label:"Get Started",onClick:()=>alert("Get started clicked")},secondaryAction:{label:"Learn More",onClick:()=>alert("Learn more clicked")}}},d={args:{size:"expanded",icon:e.jsx(He,{}),title:"Start managing consent requests",description:"Centralize all your consent management in one place. Review requests, grant permissions, and track usage across all connected applications.",primaryAction:{label:"Connect First App",onClick:()=>alert("Connect app clicked")},secondaryAction:{label:"View Documentation",onClick:()=>alert("View docs clicked")}}},p={args:{title:"Customize this empty state",description:"Use the controls below to experiment with different props and configurations.",size:"default",primaryAction:{label:"Primary Action",onClick:()=>alert("Primary action clicked")},secondaryAction:{label:"Secondary Action",onClick:()=>alert("Secondary action clicked")}}},m={render:()=>e.jsxs("div",{className:"space-y-12 max-w-4xl mx-auto",children:[e.jsxs("div",{className:"border border-gray-200 rounded-lg p-8",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-700 mb-6",children:"Active Consents Tab"}),e.jsx(t,{icon:e.jsx(Pe,{}),title:"No active consents",description:"You don't have any active consent grants. Applications you authorize will appear here.",primaryAction:{label:"Browse Applications",onClick:()=>alert("Browse applications")}})]}),e.jsxs("div",{className:"border border-gray-200 rounded-lg p-8",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-700 mb-6",children:"Pending Requests Tab"}),e.jsx(t,{icon:e.jsx(Ie,{}),title:"All caught up!",description:"You have no pending consent requests at the moment. New requests will appear here."})]}),e.jsxs("div",{className:"border border-gray-200 rounded-lg p-8",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-700 mb-6",children:"Revoked History Tab"}),e.jsx(t,{icon:e.jsx(f,{}),title:"No revoked consents",description:"You haven't revoked any permissions yet. When you revoke access to an application, it will appear in this history.",secondaryAction:{label:"View Active Consents",onClick:()=>alert("View active")}})]}),e.jsxs("div",{className:"border border-gray-200 rounded-lg p-8",children:[e.jsxs("div",{className:"mb-6",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-700 mb-2",children:"Search Results"}),e.jsx("input",{type:"text",placeholder:"Search applications...",value:"xyzabc123",readOnly:!0,className:"w-full px-3 py-2 border border-gray-300 rounded-md text-sm"})]}),e.jsx(t,{size:"compact",icon:e.jsx(k,{}),title:"No matching applications",description:"No applications found matching 'xyzabc123'. Try a different search term.",primaryAction:{label:"Clear Search",onClick:()=>alert("Clear search")}})]})]}),args:{title:"Consent scenarios"}},u={render:()=>e.jsxs("div",{className:"space-y-8",children:[e.jsxs("div",{className:"border border-gray-200 rounded-lg p-4",children:[e.jsx("h4",{className:"text-xs font-semibold text-gray-500 uppercase mb-4",children:"Compact Size"}),e.jsx(t,{size:"compact",icon:e.jsx(f,{}),title:"No documents",description:"Upload your first document to get started.",primaryAction:{label:"Upload",onClick:()=>alert("Upload clicked")}})]}),e.jsxs("div",{className:"border border-gray-200 rounded-lg p-6",children:[e.jsx("h4",{className:"text-xs font-semibold text-gray-500 uppercase mb-6",children:"Default Size"}),e.jsx(t,{size:"default",icon:e.jsx(f,{}),title:"No documents",description:"Upload your first document to get started. You can drag and drop files or use the upload button.",primaryAction:{label:"Upload Document",onClick:()=>alert("Upload clicked")},secondaryAction:{label:"Learn More",onClick:()=>alert("Learn more clicked")}})]}),e.jsxs("div",{className:"border border-gray-200 rounded-lg p-8",children:[e.jsx("h4",{className:"text-xs font-semibold text-gray-500 uppercase mb-8",children:"Expanded Size"}),e.jsx(t,{size:"expanded",icon:e.jsx(f,{}),title:"Get started with document management",description:"Securely store and manage all your important documents in one place. Upload files up to 10MB in PDF, DOC, or TXT format.",primaryAction:{label:"Upload Your First Document",onClick:()=>alert("Upload clicked")},secondaryAction:{label:"View Documentation",onClick:()=>alert("Learn more clicked")}})]})]}),args:{title:"Size comparison"}},y={render:()=>e.jsxs("div",{className:"space-y-6 max-w-3xl",children:[e.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Accessibility Features"}),e.jsxs("ul",{className:"text-sm text-gray-700 space-y-1",children:[e.jsxs("li",{children:["• Uses ",e.jsx("code",{children:'role="status"'})," and ",e.jsx("code",{children:'aria-live="polite"'})]}),e.jsxs("li",{children:["• Icons are decorative with ",e.jsx("code",{children:'aria-hidden="true"'})]}),e.jsx("li",{children:"• All interactive elements are keyboard accessible"}),e.jsx("li",{children:"• Semantic heading hierarchy (h3 for title)"}),e.jsx("li",{children:"• Focus indicators meet WCAG 2.1 AA requirements"}),e.jsx("li",{children:"• Color contrast ratios comply with AA standards"}),e.jsx("li",{children:"• Button labels are clear and descriptive"})]})]}),e.jsx(t,{icon:e.jsx(k,{}),title:"Search complete",description:"No results found for your query. The empty state is announced to screen readers via aria-live.",primaryAction:{label:"Refine Search",onClick:()=>alert("Refine search")},secondaryAction:{label:"Reset Filters",onClick:()=>alert("Reset filters")}})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0},{id:"button-name",enabled:!0}]}}},args:{title:"Accessible empty state"}};var w,A,j,N,S;r.parameters={...r.parameters,docs:{...(w=r.parameters)==null?void 0:w.docs,source:{originalSource:`{
  args: {
    title: 'No items yet',
    description: 'Get started by creating your first item.'
  }
}`,...(j=(A=r.parameters)==null?void 0:A.docs)==null?void 0:j.source},description:{story:`Default empty state with icon and message.
Basic usage for simple empty scenarios.`,...(S=(N=r.parameters)==null?void 0:N.docs)==null?void 0:S.description}}};var z,I,q,R,E;o.parameters={...o.parameters,docs:{...(z=o.parameters)==null?void 0:z.docs,source:{originalSource:`{
  args: {
    title: 'No consents found',
    description: 'You haven\\'t granted any permissions yet. Connect your first application to get started.',
    primaryAction: {
      label: 'Add Application',
      onClick: () => alert('Add application clicked')
    }
  }
}`,...(q=(I=o.parameters)==null?void 0:I.docs)==null?void 0:q.source},description:{story:`Empty state with a primary action button.
Encourages users to take action to populate the empty state.`,...(E=(R=o.parameters)==null?void 0:R.docs)==null?void 0:E.description}}};var L,W,T,B,V;s.parameters={...s.parameters,docs:{...(L=s.parameters)==null?void 0:L.docs,source:{originalSource:`{
  args: {
    title: 'Build your first workflow',
    description: 'Workflows help automate consent management tasks. Start from scratch or use a template.',
    primaryAction: {
      label: 'Create Workflow',
      onClick: () => alert('Create workflow clicked')
    },
    secondaryAction: {
      label: 'Browse Templates',
      onClick: () => alert('Browse templates clicked')
    }
  }
}`,...(T=(W=s.parameters)==null?void 0:W.docs)==null?void 0:T.source},description:{story:`Empty state with both primary and secondary actions.
Provides multiple pathways for users to proceed.`,...(V=(B=s.parameters)==null?void 0:B.docs)==null?void 0:V.description}}};var D,M,U,Y,F;n.parameters={...n.parameters,docs:{...(D=n.parameters)==null?void 0:D.docs,source:{originalSource:`{
  args: {
    icon: <SearchIcon />,
    title: 'No results found',
    description: 'We couldn\\'t find any matches for your search. Try adjusting your filters or search terms.',
    primaryAction: {
      label: 'Clear Filters',
      onClick: () => alert('Clear filters clicked')
    }
  }
}`,...(U=(M=n.parameters)==null?void 0:M.docs)==null?void 0:U.source},description:{story:`Empty state for search/filter results.
Shows when user searches or filters but no results match.`,...(F=(Y=n.parameters)==null?void 0:Y.docs)==null?void 0:F.description}}};var G,P,H,O,_;i.parameters={...i.parameters,docs:{...(G=i.parameters)==null?void 0:G.docs,source:{originalSource:`{
  args: {
    icon: <LockIcon />,
    title: 'Access restricted',
    description: 'You don\\'t have permission to view this content. Contact your administrator to request access.',
    primaryAction: {
      label: 'Request Access',
      onClick: () => alert('Request access clicked')
    },
    secondaryAction: {
      label: 'Go Back',
      onClick: () => alert('Go back clicked')
    }
  }
}`,...(H=(P=i.parameters)==null?void 0:P.docs)==null?void 0:H.source},description:{story:`Empty state for permission/access restrictions.
Shows when user lacks required permissions to view content.`,...(_=(O=i.parameters)==null?void 0:O.docs)==null?void 0:_.description}}};var X,J,K,Q,Z;c.parameters={...c.parameters,docs:{...(X=c.parameters)==null?void 0:X.docs,source:{originalSource:`{
  render: () => <div className="max-w-xs border border-gray-200 rounded-lg p-4">
      <EmptyState size="compact" icon={<InboxIcon />} title="No messages" description="Your inbox is empty." primaryAction={{
      label: 'Compose',
      onClick: () => alert('Compose clicked')
    }} />
    </div>,
  args: {
    size: 'compact',
    title: 'No messages',
    description: 'Your inbox is empty.'
  }
}`,...(K=(J=c.parameters)==null?void 0:J.docs)==null?void 0:K.source},description:{story:`Compact size variant for inline or sidebar use.
Reduces spacing and text size for constrained layouts.`,...(Z=(Q=c.parameters)==null?void 0:Q.docs)==null?void 0:Z.description}}};var $,ee,te,ae,re;l.parameters={...l.parameters,docs:{...($=l.parameters)==null?void 0:$.docs,source:{originalSource:`{
  args: {
    size: 'expanded',
    icon: <StarIcon />,
    title: 'Welcome to Consent Manager',
    description: 'Take control of your data privacy. Start by connecting your first application and managing consent preferences.',
    primaryAction: {
      label: 'Get Started',
      onClick: () => alert('Get started clicked')
    },
    secondaryAction: {
      label: 'Learn More',
      onClick: () => alert('Learn more clicked')
    }
  }
}`,...(te=(ee=l.parameters)==null?void 0:ee.docs)==null?void 0:te.source},description:{story:`Expanded size variant for prominent display.
Increases spacing and text size for onboarding or hero sections.`,...(re=(ae=l.parameters)==null?void 0:ae.docs)==null?void 0:re.description}}};var oe,se,ne,ie,ce;d.parameters={...d.parameters,docs:{...(oe=d.parameters)==null?void 0:oe.docs,source:{originalSource:`{
  args: {
    size: 'expanded',
    icon: <WelcomeIllustration />,
    title: 'Start managing consent requests',
    description: 'Centralize all your consent management in one place. Review requests, grant permissions, and track usage across all connected applications.',
    primaryAction: {
      label: 'Connect First App',
      onClick: () => alert('Connect app clicked')
    },
    secondaryAction: {
      label: 'View Documentation',
      onClick: () => alert('View docs clicked')
    }
  }
}`,...(ne=(se=d.parameters)==null?void 0:se.docs)==null?void 0:ne.source},description:{story:`Empty state with custom illustration.
Large hero-style empty state for onboarding flows.`,...(ce=(ie=d.parameters)==null?void 0:ie.docs)==null?void 0:ce.description}}};var le,de,pe,me,ue;p.parameters={...p.parameters,docs:{...(le=p.parameters)==null?void 0:le.docs,source:{originalSource:`{
  args: {
    title: 'Customize this empty state',
    description: 'Use the controls below to experiment with different props and configurations.',
    size: 'default',
    primaryAction: {
      label: 'Primary Action',
      onClick: () => alert('Primary action clicked')
    },
    secondaryAction: {
      label: 'Secondary Action',
      onClick: () => alert('Secondary action clicked')
    }
  }
}`,...(pe=(de=p.parameters)==null?void 0:de.docs)==null?void 0:pe.source},description:{story:`Interactive playground with all controls.
Experiment with different combinations of props.`,...(ue=(me=p.parameters)==null?void 0:me.docs)==null?void 0:ue.description}}};var ye,he,xe,ge,fe;m.parameters={...m.parameters,docs:{...(ye=m.parameters)==null?void 0:ye.docs,source:{originalSource:`{
  render: () => <div className="space-y-12 max-w-4xl mx-auto">
      {/* No active consents */}
      <div className="border border-gray-200 rounded-lg p-8">
        <h4 className="text-sm font-semibold text-gray-700 mb-6">
          Active Consents Tab
        </h4>
        <EmptyState icon={<ClipboardIcon />} title="No active consents" description="You don't have any active consent grants. Applications you authorize will appear here." primaryAction={{
        label: 'Browse Applications',
        onClick: () => alert('Browse applications')
      }} />
      </div>

      {/* No pending requests */}
      <div className="border border-gray-200 rounded-lg p-8">
        <h4 className="text-sm font-semibold text-gray-700 mb-6">
          Pending Requests Tab
        </h4>
        <EmptyState icon={<InboxIcon />} title="All caught up!" description="You have no pending consent requests at the moment. New requests will appear here." />
      </div>

      {/* No revoked consents */}
      <div className="border border-gray-200 rounded-lg p-8">
        <h4 className="text-sm font-semibold text-gray-700 mb-6">
          Revoked History Tab
        </h4>
        <EmptyState icon={<DocumentIcon />} title="No revoked consents" description="You haven't revoked any permissions yet. When you revoke access to an application, it will appear in this history." secondaryAction={{
        label: 'View Active Consents',
        onClick: () => alert('View active')
      }} />
      </div>

      {/* Search with no results */}
      <div className="border border-gray-200 rounded-lg p-8">
        <div className="mb-6">
          <h4 className="text-sm font-semibold text-gray-700 mb-2">
            Search Results
          </h4>
          <input type="text" placeholder="Search applications..." value="xyzabc123" readOnly className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm" />
        </div>
        <EmptyState size="compact" icon={<SearchIcon />} title="No matching applications" description="No applications found matching 'xyzabc123'. Try a different search term." primaryAction={{
        label: 'Clear Search',
        onClick: () => alert('Clear search')
      }} />
      </div>
    </div>,
  args: {
    title: 'Consent scenarios'
  }
}`,...(xe=(he=m.parameters)==null?void 0:he.docs)==null?void 0:xe.source},description:{story:`Real-world consent management scenarios.
Demonstrates various empty states in context.`,...(fe=(ge=m.parameters)==null?void 0:ge.docs)==null?void 0:fe.description}}};var be,ke,ve,Ce,we;u.parameters={...u.parameters,docs:{...(be=u.parameters)==null?void 0:be.docs,source:{originalSource:`{
  render: () => <div className="space-y-8">
      <div className="border border-gray-200 rounded-lg p-4">
        <h4 className="text-xs font-semibold text-gray-500 uppercase mb-4">
          Compact Size
        </h4>
        <EmptyState size="compact" icon={<DocumentIcon />} title="No documents" description="Upload your first document to get started." primaryAction={{
        label: 'Upload',
        onClick: () => alert('Upload clicked')
      }} />
      </div>

      <div className="border border-gray-200 rounded-lg p-6">
        <h4 className="text-xs font-semibold text-gray-500 uppercase mb-6">
          Default Size
        </h4>
        <EmptyState size="default" icon={<DocumentIcon />} title="No documents" description="Upload your first document to get started. You can drag and drop files or use the upload button." primaryAction={{
        label: 'Upload Document',
        onClick: () => alert('Upload clicked')
      }} secondaryAction={{
        label: 'Learn More',
        onClick: () => alert('Learn more clicked')
      }} />
      </div>

      <div className="border border-gray-200 rounded-lg p-8">
        <h4 className="text-xs font-semibold text-gray-500 uppercase mb-8">
          Expanded Size
        </h4>
        <EmptyState size="expanded" icon={<DocumentIcon />} title="Get started with document management" description="Securely store and manage all your important documents in one place. Upload files up to 10MB in PDF, DOC, or TXT format." primaryAction={{
        label: 'Upload Your First Document',
        onClick: () => alert('Upload clicked')
      }} secondaryAction={{
        label: 'View Documentation',
        onClick: () => alert('Learn more clicked')
      }} />
      </div>
    </div>,
  args: {
    title: 'Size comparison'
  }
}`,...(ve=(ke=u.parameters)==null?void 0:ke.docs)==null?void 0:ve.source},description:{story:`Size comparison showing all variants together.
Helps understand spacing and typography differences.`,...(we=(Ce=u.parameters)==null?void 0:Ce.docs)==null?void 0:we.description}}};var Ae,je,Ne,Se,ze;y.parameters={...y.parameters,docs:{...(Ae=y.parameters)==null?void 0:Ae.docs,source:{originalSource:`{
  render: () => <div className="space-y-6 max-w-3xl">
      <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
        <h4 className="text-sm font-semibold text-gray-900 mb-2">
          Accessibility Features
        </h4>
        <ul className="text-sm text-gray-700 space-y-1">
          <li>• Uses <code>role="status"</code> and <code>aria-live="polite"</code></li>
          <li>• Icons are decorative with <code>aria-hidden="true"</code></li>
          <li>• All interactive elements are keyboard accessible</li>
          <li>• Semantic heading hierarchy (h3 for title)</li>
          <li>• Focus indicators meet WCAG 2.1 AA requirements</li>
          <li>• Color contrast ratios comply with AA standards</li>
          <li>• Button labels are clear and descriptive</li>
        </ul>
      </div>

      <EmptyState icon={<SearchIcon />} title="Search complete" description="No results found for your query. The empty state is announced to screen readers via aria-live." primaryAction={{
      label: 'Refine Search',
      onClick: () => alert('Refine search')
    }} secondaryAction={{
      label: 'Reset Filters',
      onClick: () => alert('Reset filters')
    }} />
    </div>,
  parameters: {
    a11y: {
      config: {
        rules: [{
          id: 'color-contrast',
          enabled: true
        }, {
          id: 'button-name',
          enabled: true
        }]
      }
    }
  },
  args: {
    title: 'Accessible empty state'
  }
}`,...(Ne=(je=y.parameters)==null?void 0:je.docs)==null?void 0:Ne.source},description:{story:`Accessibility features demonstration.
Shows semantic HTML and ARIA attributes in action.`,...(ze=(Se=y.parameters)==null?void 0:Se.docs)==null?void 0:ze.description}}};const Ze=["Default","WithAction","WithMultipleActions","NoResults","NoPermissions","Compact","Expanded","WithIllustration","Playground","ConsentScenarios","SizeComparison","Accessibility"];export{y as Accessibility,c as Compact,m as ConsentScenarios,r as Default,l as Expanded,i as NoPermissions,n as NoResults,p as Playground,u as SizeComparison,o as WithAction,d as WithIllustration,s as WithMultipleActions,Ze as __namedExportsOrder,Qe as default};
