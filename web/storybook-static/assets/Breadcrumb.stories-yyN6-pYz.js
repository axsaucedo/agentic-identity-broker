import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as B}from"./index-ClcD9ViR.js";import{c as o,a as S}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const _e=S("inline-flex items-center",{variants:{size:{sm:"text-xs gap-1",md:"text-sm gap-2",lg:"text-base gap-3"}},defaultVariants:{size:"md"}}),Ke=S("inline-flex items-center transition-colors duration-200",{variants:{size:{sm:"gap-1",md:"gap-1.5",lg:"gap-2"},isCurrent:{true:"font-medium text-navy-900",false:""}},defaultVariants:{size:"md",isCurrent:!1}}),Oe=S("hover:underline focus:outline-none focus:ring-2 focus:ring-navy-700 focus:ring-offset-1 rounded-sm transition-colors duration-200",{variants:{disabled:{true:"opacity-50 cursor-not-allowed pointer-events-none",false:"text-navy-600 hover:text-navy-900"}},defaultVariants:{disabled:!1}}),$e=S("text-gray-400 select-none",{variants:{size:{sm:"text-xs",md:"text-sm",lg:"text-base"}},defaultVariants:{size:"md"}}),Qe=({className:t})=>e.jsx("svg",{className:t,fill:"none",stroke:"currentColor",viewBox:"0 0 24 24","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M9 5l7 7-7 7"})}),Ue=({className:t})=>e.jsx("svg",{className:t,fill:"currentColor",viewBox:"0 0 24 24","aria-hidden":"true",children:e.jsx("path",{d:"M12 8c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z"})}),a=B.forwardRef(({items:t,size:r="md",separator:Te,maxItems:c,showIcon:H=!0,className:Me,"aria-label":Ee="Breadcrumb",...Fe},Ve)=>{const k=B.useMemo(()=>{if(!c||t.length<=c)return t;const s=t[0],N=t.slice(-(c-1)),w=t.length-c;return[s,{label:`${w} more`,disabled:!0,icon:e.jsx(Ue,{className:"w-4 h-4"})},...N]},[t,c]),We=e.jsx(Qe,{className:o("w-3 h-3",r==="sm"&&"w-2.5 h-2.5",r==="lg"&&"w-3.5 h-3.5")}),qe=Te??We;return e.jsx("nav",{ref:Ve,"aria-label":Ee,className:o(_e({size:r}),Me),...Fe,children:e.jsx("ol",{className:"inline-flex items-center list-none m-0 p-0",children:k.map((s,N)=>{const w=N===k.length-1,P=w&&!s.href,Ge=s.href&&!s.disabled;return e.jsxs("li",{className:o(Ke({size:r,isCurrent:P})),children:[Ge?e.jsxs("a",{href:s.href,className:o(Oe({disabled:s.disabled}),"inline-flex items-center",r==="sm"&&"gap-1",r==="md"&&"gap-1.5",r==="lg"&&"gap-2"),"aria-current":P?"page":void 0,"aria-disabled":s.disabled,children:[H&&s.icon&&e.jsx("span",{className:"inline-flex items-center flex-shrink-0","aria-hidden":"true",children:s.icon}),e.jsx("span",{children:s.label})]}):e.jsxs("span",{className:o("inline-flex items-center",r==="sm"&&"gap-1",r==="md"&&"gap-1.5",r==="lg"&&"gap-2",s.disabled&&"opacity-50 cursor-not-allowed"),"aria-current":P?"page":void 0,"aria-disabled":s.disabled,children:[H&&s.icon&&e.jsx("span",{className:"inline-flex items-center flex-shrink-0","aria-hidden":"true",children:s.icon}),e.jsx("span",{children:s.label})]}),!w&&e.jsx("span",{className:o($e({size:r}),"mx-1",r==="sm"&&"mx-0.5",r==="lg"&&"mx-2"),"aria-hidden":"true",children:qe})]},`${s.label}-${N}`)})})})});a.displayName="Breadcrumb";a.__docgenInfo={description:`Breadcrumb navigation component for showing location hierarchy.
Automatically marks the last item as current page (non-interactive).

@example
\`\`\`tsx
<Breadcrumb
  items={[
    { label: 'Home', href: '/' },
    { label: 'Settings', href: '/settings' },
    { label: 'Profile' },
  ]}
/>

<Breadcrumb
  items={items}
  size="lg"
  separator={<span>→</span>}
  maxItems={3}
/>
\`\`\``,methods:[],displayName:"Breadcrumb",props:{items:{required:!0,tsType:{name:"Array",elements:[{name:"BreadcrumbItem"}],raw:"BreadcrumbItem[]"},description:"Array of breadcrumb items to display"},size:{required:!1,tsType:{name:"union",raw:"'sm' | 'md' | 'lg'",elements:[{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"},{name:"literal",value:"'lg'"}]},description:"Size variant",defaultValue:{value:"'md'",computed:!1}},separator:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Custom separator element (defaults to /)"},maxItems:{required:!1,tsType:{name:"number"},description:"Maximum number of items to show (truncates middle items if exceeded)"},showIcon:{required:!1,tsType:{name:"boolean"},description:"Whether to show icons for items that have them",defaultValue:{value:"true",computed:!1}},"aria-label":{required:!1,tsType:{name:"string"},description:"Custom aria-label for the navigation",defaultValue:{value:"'Breadcrumb'",computed:!1}}}};const ea={title:"Design System/Navigation/Breadcrumb",component:a,parameters:{layout:"padded"},tags:["autodocs"],argTypes:{items:{control:"object",description:"Array of breadcrumb items",table:{type:{summary:"BreadcrumbItem[]"}}},size:{control:"select",options:["sm","md","lg"],description:"Breadcrumb size",table:{type:{summary:"string"},defaultValue:{summary:"md"}}},separator:{control:"text",description:"Custom separator element",table:{type:{summary:"React.ReactNode"}}},maxItems:{control:"number",description:"Maximum number of items before truncation",table:{type:{summary:"number"}}},showIcon:{control:"boolean",description:"Whether to show icons for items",table:{defaultValue:{summary:"true"}}}}},n=()=>e.jsx("svg",{className:"w-4 h-4",fill:"currentColor",viewBox:"0 0 20 20",children:e.jsx("path",{d:"M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z"})}),l=()=>e.jsx("svg",{className:"w-4 h-4",fill:"currentColor",viewBox:"0 0 20 20",children:e.jsx("path",{d:"M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"})}),i=()=>e.jsx("svg",{className:"w-4 h-4",fill:"currentColor",viewBox:"0 0 20 20",children:e.jsx("path",{fillRule:"evenodd",d:"M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z",clipRule:"evenodd"})}),I=()=>e.jsx("svg",{className:"w-4 h-4",fill:"currentColor",viewBox:"0 0 20 20",children:e.jsx("path",{fillRule:"evenodd",d:"M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z",clipRule:"evenodd"})}),m={args:{items:[{label:"Home",href:"/"},{label:"Settings",href:"/settings"},{label:"Profile"}],size:"md"}},d={args:{items:[{label:"Home",href:"/",icon:e.jsx(n,{})},{label:"Projects",href:"/projects",icon:e.jsx(l,{})},{label:"Document.pdf",icon:e.jsx(i,{})}],size:"md",showIcon:!0}},p={args:{items:[{label:"Home",href:"/"},{label:"Dashboard",href:"/dashboard"},{label:"Analytics",href:"/dashboard/analytics"},{label:"Reports"}],size:"md"}},b={args:{items:[{label:"Home",href:"/"},{label:"Documents",href:"/documents"},{label:"Work",href:"/documents/work"},{label:"2024",href:"/documents/work/2024"},{label:"Q4",href:"/documents/work/2024/q4"},{label:"Reports",href:"/documents/work/2024/q4/reports"},{label:"Final Report.pdf"}],maxItems:4,size:"md"}},u={render:()=>e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"7 levels collapsed to 3 items"}),e.jsx(a,{items:[{label:"Root",href:"/"},{label:"Level 1",href:"/1"},{label:"Level 2",href:"/1/2"},{label:"Level 3",href:"/1/2/3"},{label:"Level 4",href:"/1/2/3/4"},{label:"Level 5",href:"/1/2/3/4/5"},{label:"Current Page"}],maxItems:3})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"10 levels collapsed to 5 items"}),e.jsx(a,{items:[{label:"Home",href:"/",icon:e.jsx(n,{})},{label:"Organizations",href:"/orgs"},{label:"Acme Corp",href:"/orgs/acme"},{label:"Teams",href:"/orgs/acme/teams"},{label:"Engineering",href:"/orgs/acme/teams/eng"},{label:"Backend",href:"/orgs/acme/teams/eng/backend"},{label:"Services",href:"/orgs/acme/teams/eng/backend/services"},{label:"Auth",href:"/orgs/acme/teams/eng/backend/services/auth"},{label:"Config",href:"/orgs/acme/teams/eng/backend/services/auth/config"},{label:"production.yaml"}],maxItems:5,showIcon:!0})]})]})},h={render:()=>e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Small (sm)"}),e.jsx(a,{items:[{label:"Home",href:"/",icon:e.jsx(n,{})},{label:"Settings",href:"/settings",icon:e.jsx(I,{})},{label:"Profile"}],size:"sm"})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Medium (md)"}),e.jsx(a,{items:[{label:"Home",href:"/",icon:e.jsx(n,{})},{label:"Settings",href:"/settings",icon:e.jsx(I,{})},{label:"Profile"}],size:"md"})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Large (lg)"}),e.jsx(a,{items:[{label:"Home",href:"/",icon:e.jsx(n,{})},{label:"Settings",href:"/settings",icon:e.jsx(I,{})},{label:"Profile"}],size:"lg"})]})]})},f={render:()=>e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Slash separator (/)"}),e.jsx(a,{items:[{label:"Home",href:"/"},{label:"Products",href:"/products"},{label:"Laptops"}],separator:e.jsx("span",{children:"/"})})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Arrow separator (→)"}),e.jsx(a,{items:[{label:"Home",href:"/"},{label:"Products",href:"/products"},{label:"Laptops"}],separator:e.jsx("span",{children:"→"})})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Dot separator (•)"}),e.jsx(a,{items:[{label:"Home",href:"/"},{label:"Products",href:"/products"},{label:"Laptops"}],separator:e.jsx("span",{children:"•"})})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Pipe separator (|)"}),e.jsx(a,{items:[{label:"Home",href:"/"},{label:"Products",href:"/products"},{label:"Laptops"}],separator:e.jsx("span",{children:"|"})})]})]})},x={args:{items:[{label:"Home",href:"/"},{label:"Restricted Area",href:"/restricted",disabled:!0},{label:"Settings",href:"/settings"},{label:"Profile"}],size:"md"}},g={args:{items:[{label:"Home",href:"/",icon:e.jsx(n,{})},{label:"Projects",href:"/projects",icon:e.jsx(l,{})},{label:"Document.pdf",icon:e.jsx(i,{})}],size:"md",showIcon:!1}},v={render:()=>e.jsxs("div",{className:"space-y-8",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-base font-semibold text-neutral-900 mb-3",children:"E-commerce Product Page"}),e.jsx(a,{items:[{label:"Home",href:"/",icon:e.jsx(n,{})},{label:"Electronics",href:"/electronics"},{label:"Computers",href:"/electronics/computers"},{label:"Laptops",href:"/electronics/computers/laptops"},{label:'MacBook Pro 16"'}]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-base font-semibold text-neutral-900 mb-3",children:"Admin Dashboard Settings"}),e.jsx(a,{items:[{label:"Dashboard",href:"/admin",icon:e.jsx(n,{})},{label:"Settings",href:"/admin/settings",icon:e.jsx(I,{})},{label:"Security",href:"/admin/settings/security"},{label:"Two-Factor Authentication"}]})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-base font-semibold text-neutral-900 mb-3",children:"Documentation Navigation"}),e.jsx(a,{items:[{label:"Docs",href:"/docs",icon:e.jsx(i,{})},{label:"Components",href:"/docs/components"},{label:"Navigation",href:"/docs/components/navigation"},{label:"Breadcrumb"}],size:"sm"})]}),e.jsxs("div",{children:[e.jsx("h3",{className:"text-base font-semibold text-neutral-900 mb-3",children:"File Manager"}),e.jsx(a,{items:[{label:"My Files",href:"/files",icon:e.jsx(l,{})},{label:"Documents",href:"/files/documents",icon:e.jsx(l,{})},{label:"Projects",href:"/files/documents/projects",icon:e.jsx(l,{})},{label:"2024",href:"/files/documents/projects/2024",icon:e.jsx(l,{})},{label:"proposal.pdf",icon:e.jsx(i,{})}],maxItems:4})]})]})},j={args:{items:[{label:"Home",href:"/",icon:e.jsx(n,{})},{label:"Category",href:"/category",icon:e.jsx(l,{})},{label:"Subcategory",href:"/category/subcategory"},{label:"Current Page"}],size:"md",showIcon:!0,maxItems:void 0},parameters:{docs:{description:{story:"Experiment with all breadcrumb props using the controls below. Try different item arrays, sizes, separators, and truncation settings."}}}},y={render:()=>e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Semantic HTML Structure"}),e.jsx("p",{className:"text-sm text-neutral-600 mb-4",children:'Uses proper nav, ol, and li elements for screen reader navigation. The last item is marked with aria-current="page".'}),e.jsx(a,{items:[{label:"Home",href:"/"},{label:"Products",href:"/products"},{label:"Laptops"}],"aria-label":"Page navigation"})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Keyboard Navigation"}),e.jsx("p",{className:"text-sm text-neutral-600 mb-4",children:"Press Tab to focus links, Enter to navigate. Links have visible focus indicators."}),e.jsx(a,{items:[{label:"Dashboard",href:"/dashboard"},{label:"Settings",href:"/settings"},{label:"Profile",href:"/settings/profile"},{label:"Edit Profile"}]})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Disabled Items"}),e.jsx("p",{className:"text-sm text-neutral-600 mb-4",children:"Disabled items are marked with aria-disabled and cannot be interacted with."}),e.jsx(a,{items:[{label:"Home",href:"/"},{label:"Restricted",href:"/restricted",disabled:!0},{label:"Public",href:"/public"},{label:"Page"}]})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Screen Reader Support"}),e.jsx("p",{className:"text-sm text-neutral-600 mb-4",children:"Separators are hidden from screen readers with aria-hidden. Icons have proper aria-hidden attributes."}),e.jsx(a,{items:[{label:"Docs",href:"/docs",icon:e.jsx(i,{})},{label:"Guides",href:"/docs/guides",icon:e.jsx(l,{})},{label:"Accessibility Guide",icon:e.jsx(i,{})}],showIcon:!0})]})]}),parameters:{docs:{description:{story:"Breadcrumbs are fully accessible with semantic HTML, ARIA attributes, keyboard support, and screen reader compatibility. All navigation states are properly announced."}}}};var D,C,z,L,R;m.parameters={...m.parameters,docs:{...(D=m.parameters)==null?void 0:D.docs,source:{originalSource:`{
  args: {
    items: [{
      label: 'Home',
      href: '/'
    }, {
      label: 'Settings',
      href: '/settings'
    }, {
      label: 'Profile'
    }],
    size: 'md'
  }
}`,...(z=(C=m.parameters)==null?void 0:C.docs)==null?void 0:z.source},description:{story:`Default breadcrumb with basic navigation path.
Shows the most common usage with a simple three-level hierarchy.`,...(R=(L=m.parameters)==null?void 0:L.docs)==null?void 0:R.description}}};var A,T,M,E,F;d.parameters={...d.parameters,docs:{...(A=d.parameters)==null?void 0:A.docs,source:{originalSource:`{
  args: {
    items: [{
      label: 'Home',
      href: '/',
      icon: <HomeIcon />
    }, {
      label: 'Projects',
      href: '/projects',
      icon: <FolderIcon />
    }, {
      label: 'Document.pdf',
      icon: <DocumentIcon />
    }],
    size: 'md',
    showIcon: true
  }
}`,...(M=(T=d.parameters)==null?void 0:T.docs)==null?void 0:M.source},description:{story:`Breadcrumb with icons for visual clarity.
Icons help users quickly identify page types and hierarchy levels.`,...(F=(E=d.parameters)==null?void 0:E.docs)==null?void 0:F.description}}};var V,W,q,G,_;p.parameters={...p.parameters,docs:{...(V=p.parameters)==null?void 0:V.docs,source:{originalSource:`{
  args: {
    items: [{
      label: 'Home',
      href: '/'
    }, {
      label: 'Dashboard',
      href: '/dashboard'
    }, {
      label: 'Analytics',
      href: '/dashboard/analytics'
    }, {
      label: 'Reports'
    } // Current page - no href
    ],
    size: 'md'
  }
}`,...(q=(W=p.parameters)==null?void 0:W.docs)==null?void 0:q.source},description:{story:`Current page indicator.
The last item without an href is automatically styled as the current page
and marked with aria-current="page" for accessibility.`,...(_=(G=p.parameters)==null?void 0:G.docs)==null?void 0:_.description}}};var K,O,$,Q,U;b.parameters={...b.parameters,docs:{...(K=b.parameters)==null?void 0:K.docs,source:{originalSource:`{
  args: {
    items: [{
      label: 'Home',
      href: '/'
    }, {
      label: 'Documents',
      href: '/documents'
    }, {
      label: 'Work',
      href: '/documents/work'
    }, {
      label: '2024',
      href: '/documents/work/2024'
    }, {
      label: 'Q4',
      href: '/documents/work/2024/q4'
    }, {
      label: 'Reports',
      href: '/documents/work/2024/q4/reports'
    }, {
      label: 'Final Report.pdf'
    }],
    maxItems: 4,
    size: 'md'
  }
}`,...($=(O=b.parameters)==null?void 0:O.docs)==null?void 0:$.source},description:{story:`Truncated breadcrumb for long paths.
When maxItems is set, middle items are collapsed with an ellipsis indicator.
First and last items are always visible.`,...(U=(Q=b.parameters)==null?void 0:Q.docs)==null?void 0:U.description}}};var Y,J,X,Z,ee;u.parameters={...u.parameters,docs:{...(Y=u.parameters)==null?void 0:Y.docs,source:{originalSource:`{
  render: () => <div className="space-y-6">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          7 levels collapsed to 3 items
        </p>
        <Breadcrumb items={[{
        label: 'Root',
        href: '/'
      }, {
        label: 'Level 1',
        href: '/1'
      }, {
        label: 'Level 2',
        href: '/1/2'
      }, {
        label: 'Level 3',
        href: '/1/2/3'
      }, {
        label: 'Level 4',
        href: '/1/2/3/4'
      }, {
        label: 'Level 5',
        href: '/1/2/3/4/5'
      }, {
        label: 'Current Page'
      }]} maxItems={3} />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          10 levels collapsed to 5 items
        </p>
        <Breadcrumb items={[{
        label: 'Home',
        href: '/',
        icon: <HomeIcon />
      }, {
        label: 'Organizations',
        href: '/orgs'
      }, {
        label: 'Acme Corp',
        href: '/orgs/acme'
      }, {
        label: 'Teams',
        href: '/orgs/acme/teams'
      }, {
        label: 'Engineering',
        href: '/orgs/acme/teams/eng'
      }, {
        label: 'Backend',
        href: '/orgs/acme/teams/eng/backend'
      }, {
        label: 'Services',
        href: '/orgs/acme/teams/eng/backend/services'
      }, {
        label: 'Auth',
        href: '/orgs/acme/teams/eng/backend/services/auth'
      }, {
        label: 'Config',
        href: '/orgs/acme/teams/eng/backend/services/auth/config'
      }, {
        label: 'production.yaml'
      }]} maxItems={5} showIcon />
      </div>
    </div>
}`,...(X=(J=u.parameters)==null?void 0:J.docs)==null?void 0:X.source},description:{story:`Collapsed middle section for very long paths.
Shows how the breadcrumb handles deep hierarchies by collapsing
intermediate items while preserving navigation context.`,...(ee=(Z=u.parameters)==null?void 0:Z.docs)==null?void 0:ee.description}}};var ae,se,re,te,ne;h.parameters={...h.parameters,docs:{...(ae=h.parameters)==null?void 0:ae.docs,source:{originalSource:`{
  render: () => <div className="space-y-6">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">Small (sm)</p>
        <Breadcrumb items={[{
        label: 'Home',
        href: '/',
        icon: <HomeIcon />
      }, {
        label: 'Settings',
        href: '/settings',
        icon: <SettingsIcon />
      }, {
        label: 'Profile'
      }]} size="sm" />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">Medium (md)</p>
        <Breadcrumb items={[{
        label: 'Home',
        href: '/',
        icon: <HomeIcon />
      }, {
        label: 'Settings',
        href: '/settings',
        icon: <SettingsIcon />
      }, {
        label: 'Profile'
      }]} size="md" />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">Large (lg)</p>
        <Breadcrumb items={[{
        label: 'Home',
        href: '/',
        icon: <HomeIcon />
      }, {
        label: 'Settings',
        href: '/settings',
        icon: <SettingsIcon />
      }, {
        label: 'Profile'
      }]} size="lg" />
      </div>
    </div>
}`,...(re=(se=h.parameters)==null?void 0:se.docs)==null?void 0:re.source},description:{story:`All breadcrumb sizes for different contexts.
- sm: Compact spaces, tight layouts
- md: Default size for most pages
- lg: Prominent navigation, spacious layouts`,...(ne=(te=h.parameters)==null?void 0:te.docs)==null?void 0:ne.description}}};var le,oe,ie,ce,me;f.parameters={...f.parameters,docs:{...(le=f.parameters)==null?void 0:le.docs,source:{originalSource:`{
  render: () => <div className="space-y-6">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Slash separator (/)
        </p>
        <Breadcrumb items={[{
        label: 'Home',
        href: '/'
      }, {
        label: 'Products',
        href: '/products'
      }, {
        label: 'Laptops'
      }]} separator={<span>/</span>} />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Arrow separator (→)
        </p>
        <Breadcrumb items={[{
        label: 'Home',
        href: '/'
      }, {
        label: 'Products',
        href: '/products'
      }, {
        label: 'Laptops'
      }]} separator={<span>→</span>} />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Dot separator (•)
        </p>
        <Breadcrumb items={[{
        label: 'Home',
        href: '/'
      }, {
        label: 'Products',
        href: '/products'
      }, {
        label: 'Laptops'
      }]} separator={<span>•</span>} />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Pipe separator (|)
        </p>
        <Breadcrumb items={[{
        label: 'Home',
        href: '/'
      }, {
        label: 'Products',
        href: '/products'
      }, {
        label: 'Laptops'
      }]} separator={<span>|</span>} />
      </div>
    </div>
}`,...(ie=(oe=f.parameters)==null?void 0:oe.docs)==null?void 0:ie.source},description:{story:`Custom separators for different visual styles.
You can use any React node as a separator: text, icons, or custom components.`,...(me=(ce=f.parameters)==null?void 0:ce.docs)==null?void 0:me.description}}};var de,pe,be,ue,he;x.parameters={...x.parameters,docs:{...(de=x.parameters)==null?void 0:de.docs,source:{originalSource:`{
  args: {
    items: [{
      label: 'Home',
      href: '/'
    }, {
      label: 'Restricted Area',
      href: '/restricted',
      disabled: true
    }, {
      label: 'Settings',
      href: '/settings'
    }, {
      label: 'Profile'
    }],
    size: 'md'
  }
}`,...(be=(pe=x.parameters)==null?void 0:pe.docs)==null?void 0:be.source},description:{story:`Disabled breadcrumb items.
Disabled items are styled with reduced opacity and cannot be interacted with.`,...(he=(ue=x.parameters)==null?void 0:ue.docs)==null?void 0:he.description}}};var fe,xe,ge,ve,je;g.parameters={...g.parameters,docs:{...(fe=g.parameters)==null?void 0:fe.docs,source:{originalSource:`{
  args: {
    items: [{
      label: 'Home',
      href: '/',
      icon: <HomeIcon />
    }, {
      label: 'Projects',
      href: '/projects',
      icon: <FolderIcon />
    }, {
      label: 'Document.pdf',
      icon: <DocumentIcon />
    }],
    size: 'md',
    showIcon: false
  }
}`,...(ge=(xe=g.parameters)==null?void 0:xe.docs)==null?void 0:ge.source},description:{story:`Without icons for a cleaner look.
Set showIcon to false to hide all icons, even if items have them defined.`,...(je=(ve=g.parameters)==null?void 0:ve.docs)==null?void 0:je.description}}};var ye,Ne,we,Ie,Se;v.parameters={...v.parameters,docs:{...(ye=v.parameters)==null?void 0:ye.docs,source:{originalSource:`{
  render: () => <div className="space-y-8">
      <div>
        <h3 className="text-base font-semibold text-neutral-900 mb-3">
          E-commerce Product Page
        </h3>
        <Breadcrumb items={[{
        label: 'Home',
        href: '/',
        icon: <HomeIcon />
      }, {
        label: 'Electronics',
        href: '/electronics'
      }, {
        label: 'Computers',
        href: '/electronics/computers'
      }, {
        label: 'Laptops',
        href: '/electronics/computers/laptops'
      }, {
        label: 'MacBook Pro 16"'
      }]} />
      </div>

      <div>
        <h3 className="text-base font-semibold text-neutral-900 mb-3">
          Admin Dashboard Settings
        </h3>
        <Breadcrumb items={[{
        label: 'Dashboard',
        href: '/admin',
        icon: <HomeIcon />
      }, {
        label: 'Settings',
        href: '/admin/settings',
        icon: <SettingsIcon />
      }, {
        label: 'Security',
        href: '/admin/settings/security'
      }, {
        label: 'Two-Factor Authentication'
      }]} />
      </div>

      <div>
        <h3 className="text-base font-semibold text-neutral-900 mb-3">
          Documentation Navigation
        </h3>
        <Breadcrumb items={[{
        label: 'Docs',
        href: '/docs',
        icon: <DocumentIcon />
      }, {
        label: 'Components',
        href: '/docs/components'
      }, {
        label: 'Navigation',
        href: '/docs/components/navigation'
      }, {
        label: 'Breadcrumb'
      }]} size="sm" />
      </div>

      <div>
        <h3 className="text-base font-semibold text-neutral-900 mb-3">
          File Manager
        </h3>
        <Breadcrumb items={[{
        label: 'My Files',
        href: '/files',
        icon: <FolderIcon />
      }, {
        label: 'Documents',
        href: '/files/documents',
        icon: <FolderIcon />
      }, {
        label: 'Projects',
        href: '/files/documents/projects',
        icon: <FolderIcon />
      }, {
        label: '2024',
        href: '/files/documents/projects/2024',
        icon: <FolderIcon />
      }, {
        label: 'proposal.pdf',
        icon: <DocumentIcon />
      }]} maxItems={4} />
      </div>
    </div>
}`,...(we=(Ne=v.parameters)==null?void 0:Ne.docs)==null?void 0:we.source},description:{story:"Real-world examples in common application contexts.",...(Se=(Ie=v.parameters)==null?void 0:Ie.docs)==null?void 0:Se.description}}};var Pe,He,ke,Be,De;j.parameters={...j.parameters,docs:{...(Pe=j.parameters)==null?void 0:Pe.docs,source:{originalSource:`{
  args: {
    items: [{
      label: 'Home',
      href: '/',
      icon: <HomeIcon />
    }, {
      label: 'Category',
      href: '/category',
      icon: <FolderIcon />
    }, {
      label: 'Subcategory',
      href: '/category/subcategory'
    }, {
      label: 'Current Page'
    }],
    size: 'md',
    showIcon: true,
    maxItems: undefined
  },
  parameters: {
    docs: {
      description: {
        story: 'Experiment with all breadcrumb props using the controls below. Try different item arrays, sizes, separators, and truncation settings.'
      }
    }
  }
}`,...(ke=(He=j.parameters)==null?void 0:He.docs)==null?void 0:ke.source},description:{story:`Interactive playground to experiment with all props.
Try different combinations of items, sizes, separators, and truncation.`,...(De=(Be=j.parameters)==null?void 0:Be.docs)==null?void 0:De.description}}};var Ce,ze,Le,Re,Ae;y.parameters={...y.parameters,docs:{...(Ce=y.parameters)==null?void 0:Ce.docs,source:{originalSource:`{
  render: () => <div className="space-y-6">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Semantic HTML Structure
        </p>
        <p className="text-sm text-neutral-600 mb-4">
          Uses proper nav, ol, and li elements for screen reader navigation.
          The last item is marked with aria-current="page".
        </p>
        <Breadcrumb items={[{
        label: 'Home',
        href: '/'
      }, {
        label: 'Products',
        href: '/products'
      }, {
        label: 'Laptops'
      }]} aria-label="Page navigation" />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Keyboard Navigation
        </p>
        <p className="text-sm text-neutral-600 mb-4">
          Press Tab to focus links, Enter to navigate. Links have visible focus
          indicators.
        </p>
        <Breadcrumb items={[{
        label: 'Dashboard',
        href: '/dashboard'
      }, {
        label: 'Settings',
        href: '/settings'
      }, {
        label: 'Profile',
        href: '/settings/profile'
      }, {
        label: 'Edit Profile'
      }]} />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Disabled Items
        </p>
        <p className="text-sm text-neutral-600 mb-4">
          Disabled items are marked with aria-disabled and cannot be interacted
          with.
        </p>
        <Breadcrumb items={[{
        label: 'Home',
        href: '/'
      }, {
        label: 'Restricted',
        href: '/restricted',
        disabled: true
      }, {
        label: 'Public',
        href: '/public'
      }, {
        label: 'Page'
      }]} />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Screen Reader Support
        </p>
        <p className="text-sm text-neutral-600 mb-4">
          Separators are hidden from screen readers with aria-hidden. Icons have
          proper aria-hidden attributes.
        </p>
        <Breadcrumb items={[{
        label: 'Docs',
        href: '/docs',
        icon: <DocumentIcon />
      }, {
        label: 'Guides',
        href: '/docs/guides',
        icon: <FolderIcon />
      }, {
        label: 'Accessibility Guide',
        icon: <DocumentIcon />
      }]} showIcon />
      </div>
    </div>,
  parameters: {
    docs: {
      description: {
        story: 'Breadcrumbs are fully accessible with semantic HTML, ARIA attributes, keyboard support, and screen reader compatibility. All navigation states are properly announced.'
      }
    }
  }
}`,...(Le=(ze=y.parameters)==null?void 0:ze.docs)==null?void 0:Le.source},description:{story:`Accessibility features demonstration.
All breadcrumbs support:
- Semantic HTML (nav, ol, li elements)
- aria-label for navigation landmark
- aria-current="page" for current page
- aria-disabled for disabled items
- Keyboard navigation (Tab to focus links, Enter to activate)
- Screen reader friendly structure`,...(Ae=(Re=y.parameters)==null?void 0:Re.docs)==null?void 0:Ae.description}}};const aa=["Default","WithIcon","CurrentPage","Truncated","Collapsed","Sizes","CustomSeparators","DisabledItems","WithoutIcons","RealWorldExamples","Playground","Accessibility"];export{y as Accessibility,u as Collapsed,p as CurrentPage,f as CustomSeparators,m as Default,x as DisabledItems,j as Playground,v as RealWorldExamples,h as Sizes,b as Truncated,d as WithIcon,g as WithoutIcons,aa as __namedExportsOrder,ea as default};
