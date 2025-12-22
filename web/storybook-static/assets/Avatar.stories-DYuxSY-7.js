import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as N}from"./index-ClcD9ViR.js";import{c as y,a as Me}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const Pe=Me("inline-flex items-center justify-center overflow-hidden bg-gray-100 text-gray-700 font-medium relative",{variants:{size:{xs:"w-6 h-6 text-xs",sm:"w-8 h-8 text-sm",md:"w-10 h-10 text-base",lg:"w-12 h-12 text-lg",xl:"w-16 h-16 text-2xl"},shape:{circle:"rounded-full",rounded:"rounded-md",square:"rounded-none"}},defaultVariants:{size:"md",shape:"circle"}}),We=Me("absolute bottom-0 right-0 rounded-full border-2 border-white",{variants:{size:{xs:"w-1.5 h-1.5",sm:"w-2 h-2",md:"w-2.5 h-2.5",lg:"w-3 h-3",xl:"w-4 h-4"},status:{online:"bg-success-primary",offline:"bg-gray-400",away:"bg-warning-primary",busy:"bg-error-primary"}}}),s=N.forwardRef(({size:j,shape:Be,className:Re,src:A,alt:Ce="",initials:g,fallbackIcon:b,status:f,onError:v,...Le},Ke)=>{const[qe,Te]=N.useState(!1),[Ue,Fe]=N.useState(!1),Ee=()=>{Te(!0),v==null||v()},Ve=()=>{Fe(!0)},h=A&&!qe,u=!h&&g,w=!h&&!u&&b,Ge=g?g.slice(0,2).toUpperCase():"";return e.jsxs("div",{ref:Ke,className:y(Pe({size:j,shape:Be}),Re),...Le,children:[h&&e.jsx("img",{src:A,alt:Ce,className:y("w-full h-full object-cover",!Ue&&"opacity-0"),onError:Ee,onLoad:Ve}),u&&e.jsx("span",{children:Ge}),w&&e.jsx("span",{className:"w-1/2 h-1/2 text-gray-400",children:b}),!h&&!u&&!w&&e.jsx("svg",{className:"w-1/2 h-1/2 text-gray-400",fill:"currentColor",viewBox:"0 0 20 20","aria-hidden":"true",children:e.jsx("path",{fillRule:"evenodd",d:"M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z",clipRule:"evenodd"})}),f&&e.jsx("span",{className:y(We({size:j,status:f})),"aria-label":`Status: ${f}`})]})});s.displayName="Avatar";s.__docgenInfo={description:`Avatar component for displaying profile pictures or initials.
Automatically falls back to initials or icon if image fails to load.

@example
\`\`\`tsx
<Avatar src="/profile.jpg" alt="John Doe" />

<Avatar initials="JD" status="online" />

<Avatar fallbackIcon={<UserIcon />} size="lg" />

<Avatar src="/agent.png" alt="AI Agent" shape="rounded" />
\`\`\``,methods:[],displayName:"Avatar",props:{src:{required:!1,tsType:{name:"string"},description:"Image source URL"},alt:{required:!1,tsType:{name:"string"},description:"Alt text for the image",defaultValue:{value:"''",computed:!1}},initials:{required:!1,tsType:{name:"string"},description:"Initials to display (1-2 characters)"},fallbackIcon:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Fallback icon element"},status:{required:!1,tsType:{name:"union",raw:"'online' | 'offline' | 'away' | 'busy'",elements:[{name:"literal",value:"'online'"},{name:"literal",value:"'offline'"},{name:"literal",value:"'away'"},{name:"literal",value:"'busy'"}]},description:"Status indicator"},onError:{required:!1,tsType:{name:"signature",type:"function",raw:"() => void",signature:{arguments:[],return:{name:"void"}}},description:"Callback when image fails to load"}},composes:["Omit","VariantProps"]};const $e={title:"Design System/Primitives/Avatar",component:s,parameters:{layout:"centered",docs:{description:{component:"Avatar component for displaying user or agent profile pictures, initials, or fallback icons. Supports multiple sizes, shapes, and status indicators."}}},tags:["autodocs"],argTypes:{size:{control:"select",options:["xs","sm","md","lg","xl"],description:"The size of the avatar"},shape:{control:"select",options:["circle","rounded","square"],description:"The shape of the avatar"},status:{control:"select",options:["online","offline","away","busy"],description:"Status indicator"},src:{control:"text",description:"Image source URL"},alt:{control:"text",description:"Alt text for the image"},initials:{control:"text",description:"Initials to display (1-2 characters)"}}},a={args:{}},t={render:()=>e.jsxs("div",{className:"flex items-end gap-4",children:[e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{size:"xs",initials:"XS"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Extra Small"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{size:"sm",initials:"SM"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Small"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{size:"md",initials:"MD"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Medium"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{size:"lg",initials:"LG"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Large"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{size:"xl",initials:"XL"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Extra Large"})]})]})},i={render:()=>e.jsxs("div",{className:"flex items-center gap-4",children:[e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{shape:"circle",initials:"JD"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Circle"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{shape:"rounded",initials:"JD"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Rounded"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{shape:"square",initials:"JD"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Square"})]})]})},r={render:()=>e.jsxs("div",{className:"flex items-center gap-4",children:[e.jsx(s,{initials:"JD",alt:"John Doe"}),e.jsx(s,{initials:"AS",alt:"Alice Smith"}),e.jsx(s,{initials:"BJ",alt:"Bob Johnson"}),e.jsx(s,{initials:"MK",alt:"Mary King"})]})},n={render:()=>e.jsxs("div",{className:"flex items-center gap-6",children:[e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{initials:"JD",status:"online"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Online"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{initials:"AS",status:"away"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Away"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{initials:"BJ",status:"busy"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Busy"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{initials:"MK",status:"offline"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Offline"})]})]})},l={render:()=>e.jsxs("div",{className:"flex items-center gap-4",children:[e.jsx(s,{src:"https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop",alt:"John Doe"}),e.jsx(s,{src:"https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=100&h=100&fit=crop",alt:"Alice Smith"}),e.jsx(s,{src:"https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=100&h=100&fit=crop",alt:"Bob Johnson"}),e.jsx(s,{src:"https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=100&h=100&fit=crop",alt:"Mary King"})]})},c={render:()=>e.jsxs("div",{className:"flex items-center gap-4",children:[e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{src:"/broken-image.jpg",initials:"JD",alt:"John Doe"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Fallback to initials"})]}),e.jsxs("div",{className:"flex flex-col items-center gap-2",children:[e.jsx(s,{src:"/broken-image.jpg",alt:"Alice Smith"}),e.jsx("span",{className:"text-xs text-gray-600",children:"Fallback to icon"})]})]})},o={render:()=>e.jsxs("div",{className:"flex items-center gap-4",children:[e.jsx(s,{fallbackIcon:e.jsx("svg",{fill:"currentColor",viewBox:"0 0 20 20",className:"w-full h-full",children:e.jsx("path",{fillRule:"evenodd",d:"M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z",clipRule:"evenodd"})}),alt:"Verified agent"}),e.jsx(s,{fallbackIcon:e.jsx("svg",{fill:"currentColor",viewBox:"0 0 20 20",className:"w-full h-full",children:e.jsx("path",{d:"M2 10.5a1.5 1.5 0 113 0v6a1.5 1.5 0 01-3 0v-6zM6 10.333v5.43a2 2 0 001.106 1.79l.05.025A4 4 0 008.943 18h5.416a2 2 0 001.962-1.608l1.2-6A2 2 0 0015.56 8H12V4a2 2 0 00-2-2 1 1 0 00-1 1v.667a4 4 0 01-.8 2.4L6.8 7.933a4 4 0 00-.8 2.4z"})}),alt:"AI Agent",shape:"rounded"})]})},d={render:()=>e.jsxs("div",{className:"flex flex-col gap-6",children:[e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Stacked Group"}),e.jsxs("div",{className:"flex -space-x-2",children:[e.jsx(s,{src:"https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop",alt:"User 1",className:"ring-2 ring-white"}),e.jsx(s,{src:"https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=100&h=100&fit=crop",alt:"User 2",className:"ring-2 ring-white"}),e.jsx(s,{src:"https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=100&h=100&fit=crop",alt:"User 3",className:"ring-2 ring-white"}),e.jsx(s,{initials:"+5",className:"ring-2 ring-white"})]})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Spaced Group"}),e.jsxs("div",{className:"flex gap-2",children:[e.jsx(s,{initials:"JD",status:"online",size:"sm"}),e.jsx(s,{initials:"AS",status:"away",size:"sm"}),e.jsx(s,{initials:"BJ",status:"busy",size:"sm"}),e.jsx(s,{initials:"MK",status:"offline",size:"sm"})]})]})]})},m={render:()=>e.jsxs("div",{className:"flex flex-col gap-6 p-6",children:[e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Profile Card"}),e.jsxs("div",{className:"flex items-center gap-3 p-4 bg-white border border-gray-200 rounded-lg",children:[e.jsx(s,{src:"https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop",alt:"John Doe",status:"online",size:"lg"}),e.jsxs("div",{children:[e.jsx("h4",{className:"text-sm font-medium text-gray-900",children:"John Doe"}),e.jsx("p",{className:"text-xs text-gray-500",children:"john.doe@example.com"})]})]})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"AI Agent"}),e.jsxs("div",{className:"flex items-center gap-3 p-4 bg-white border border-gray-200 rounded-lg",children:[e.jsx(s,{initials:"AI",shape:"rounded",status:"online",size:"lg"}),e.jsxs("div",{children:[e.jsx("h4",{className:"text-sm font-medium text-gray-900",children:"Email Assistant"}),e.jsx("p",{className:"text-xs text-gray-500",children:"Active · 3 permissions granted"})]})]})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Comment Thread"}),e.jsxs("div",{className:"space-y-3",children:[e.jsxs("div",{className:"flex gap-3",children:[e.jsx(s,{initials:"JD",size:"sm"}),e.jsxs("div",{className:"flex-1",children:[e.jsx("p",{className:"text-xs font-medium text-gray-900",children:"John Doe"}),e.jsx("p",{className:"text-xs text-gray-600",children:"This looks great! When can we ship?"})]})]}),e.jsxs("div",{className:"flex gap-3",children:[e.jsx(s,{initials:"AS",size:"sm"}),e.jsxs("div",{className:"flex-1",children:[e.jsx("p",{className:"text-xs font-medium text-gray-900",children:"Alice Smith"}),e.jsx("p",{className:"text-xs text-gray-600",children:"Ready to go live tomorrow."})]})]})]})]})]})},x={args:{initials:"JD",size:"md",shape:"circle",status:"online"}},p={render:()=>e.jsxs("div",{className:"flex flex-col gap-6",children:[e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Avatars with proper alt text"}),e.jsxs("div",{className:"flex gap-4",children:[e.jsx(s,{src:"https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop",alt:"John Doe, Senior Developer"}),e.jsx(s,{initials:"AS",alt:"Alice Smith, Product Manager",status:"online"}),e.jsx(s,{alt:"Anonymous user"})]})]}),e.jsxs("div",{className:"space-y-2",children:[e.jsx("h3",{className:"text-sm font-semibold text-gray-700",children:"Status indicators with aria-label"}),e.jsxs("div",{className:"flex gap-4",children:[e.jsx(s,{initials:"JD",status:"online",alt:"John Doe"}),e.jsx(s,{initials:"AS",status:"away",alt:"Alice Smith"}),e.jsx(s,{initials:"BJ",status:"busy",alt:"Bob Johnson"}),e.jsx(s,{initials:"MK",status:"offline",alt:"Mary King"})]})]})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0}]}}}};var S,J,z,D,I;a.parameters={...a.parameters,docs:{...(S=a.parameters)==null?void 0:S.docs,source:{originalSource:`{
  args: {}
}`,...(z=(J=a.parameters)==null?void 0:J.docs)==null?void 0:z.source},description:{story:"Default avatar with fallback icon",...(I=(D=a.parameters)==null?void 0:D.docs)==null?void 0:I.description}}};var k,M,B,R,C;t.parameters={...t.parameters,docs:{...(k=t.parameters)==null?void 0:k.docs,source:{originalSource:`{
  render: () => <div className="flex items-end gap-4">
      <div className="flex flex-col items-center gap-2">
        <Avatar size="xs" initials="XS" />
        <span className="text-xs text-gray-600">Extra Small</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Avatar size="sm" initials="SM" />
        <span className="text-xs text-gray-600">Small</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Avatar size="md" initials="MD" />
        <span className="text-xs text-gray-600">Medium</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Avatar size="lg" initials="LG" />
        <span className="text-xs text-gray-600">Large</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Avatar size="xl" initials="XL" />
        <span className="text-xs text-gray-600">Extra Large</span>
      </div>
    </div>
}`,...(B=(M=t.parameters)==null?void 0:M.docs)==null?void 0:B.source},description:{story:"All avatar sizes from extra small to extra large",...(C=(R=t.parameters)==null?void 0:R.docs)==null?void 0:C.description}}};var L,K,q,T,U;i.parameters={...i.parameters,docs:{...(L=i.parameters)==null?void 0:L.docs,source:{originalSource:`{
  render: () => <div className="flex items-center gap-4">
      <div className="flex flex-col items-center gap-2">
        <Avatar shape="circle" initials="JD" />
        <span className="text-xs text-gray-600">Circle</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Avatar shape="rounded" initials="JD" />
        <span className="text-xs text-gray-600">Rounded</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Avatar shape="square" initials="JD" />
        <span className="text-xs text-gray-600">Square</span>
      </div>
    </div>
}`,...(q=(K=i.parameters)==null?void 0:K.docs)==null?void 0:q.source},description:{story:"Avatar shapes: circle, rounded, square",...(U=(T=i.parameters)==null?void 0:T.docs)==null?void 0:U.description}}};var F,E,V,G,P;r.parameters={...r.parameters,docs:{...(F=r.parameters)==null?void 0:F.docs,source:{originalSource:`{
  render: () => <div className="flex items-center gap-4">
      <Avatar initials="JD" alt="John Doe" />
      <Avatar initials="AS" alt="Alice Smith" />
      <Avatar initials="BJ" alt="Bob Johnson" />
      <Avatar initials="MK" alt="Mary King" />
    </div>
}`,...(V=(E=r.parameters)==null?void 0:E.docs)==null?void 0:V.source},description:{story:"Avatars displaying user initials",...(P=(G=r.parameters)==null?void 0:G.docs)==null?void 0:P.description}}};var W,O,X,_,H;n.parameters={...n.parameters,docs:{...(W=n.parameters)==null?void 0:W.docs,source:{originalSource:`{
  render: () => <div className="flex items-center gap-6">
      <div className="flex flex-col items-center gap-2">
        <Avatar initials="JD" status="online" />
        <span className="text-xs text-gray-600">Online</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Avatar initials="AS" status="away" />
        <span className="text-xs text-gray-600">Away</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Avatar initials="BJ" status="busy" />
        <span className="text-xs text-gray-600">Busy</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Avatar initials="MK" status="offline" />
        <span className="text-xs text-gray-600">Offline</span>
      </div>
    </div>
}`,...(X=(O=n.parameters)==null?void 0:O.docs)==null?void 0:X.source},description:{story:"Avatars with status indicators",...(H=(_=n.parameters)==null?void 0:_.docs)==null?void 0:H.description}}};var $,Q,Y,Z,ee;l.parameters={...l.parameters,docs:{...($=l.parameters)==null?void 0:$.docs,source:{originalSource:`{
  render: () => <div className="flex items-center gap-4">
      <Avatar src="https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop" alt="John Doe" />
      <Avatar src="https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=100&h=100&fit=crop" alt="Alice Smith" />
      <Avatar src="https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=100&h=100&fit=crop" alt="Bob Johnson" />
      <Avatar src="https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=100&h=100&fit=crop" alt="Mary King" />
    </div>
}`,...(Y=(Q=l.parameters)==null?void 0:Q.docs)==null?void 0:Y.source},description:{story:"Avatars with images",...(ee=(Z=l.parameters)==null?void 0:Z.docs)==null?void 0:ee.description}}};var se,ae,te,ie,re;c.parameters={...c.parameters,docs:{...(se=c.parameters)==null?void 0:se.docs,source:{originalSource:`{
  render: () => <div className="flex items-center gap-4">
      <div className="flex flex-col items-center gap-2">
        <Avatar src="/broken-image.jpg" initials="JD" alt="John Doe" />
        <span className="text-xs text-gray-600">Fallback to initials</span>
      </div>
      <div className="flex flex-col items-center gap-2">
        <Avatar src="/broken-image.jpg" alt="Alice Smith" />
        <span className="text-xs text-gray-600">Fallback to icon</span>
      </div>
    </div>
}`,...(te=(ae=c.parameters)==null?void 0:ae.docs)==null?void 0:te.source},description:{story:"Avatar with broken image fallback to initials",...(re=(ie=c.parameters)==null?void 0:ie.docs)==null?void 0:re.description}}};var ne,le,ce,oe,de;o.parameters={...o.parameters,docs:{...(ne=o.parameters)==null?void 0:ne.docs,source:{originalSource:`{
  render: () => <div className="flex items-center gap-4">
      <Avatar fallbackIcon={<svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
            <path fillRule="evenodd" d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
          </svg>} alt="Verified agent" />
      <Avatar fallbackIcon={<svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
            <path d="M2 10.5a1.5 1.5 0 113 0v6a1.5 1.5 0 01-3 0v-6zM6 10.333v5.43a2 2 0 001.106 1.79l.05.025A4 4 0 008.943 18h5.416a2 2 0 001.962-1.608l1.2-6A2 2 0 0015.56 8H12V4a2 2 0 00-2-2 1 1 0 00-1 1v.667a4 4 0 01-.8 2.4L6.8 7.933a4 4 0 00-.8 2.4z" />
          </svg>} alt="AI Agent" shape="rounded" />
    </div>
}`,...(ce=(le=o.parameters)==null?void 0:le.docs)==null?void 0:ce.source},description:{story:"Avatar with custom fallback icon",...(de=(oe=o.parameters)==null?void 0:oe.docs)==null?void 0:de.description}}};var me,xe,pe,he,ge;d.parameters={...d.parameters,docs:{...(me=d.parameters)==null?void 0:me.docs,source:{originalSource:`{
  render: () => <div className="flex flex-col gap-6">
      {/* Stacked avatars */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">Stacked Group</h3>
        <div className="flex -space-x-2">
          <Avatar src="https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop" alt="User 1" className="ring-2 ring-white" />
          <Avatar src="https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=100&h=100&fit=crop" alt="User 2" className="ring-2 ring-white" />
          <Avatar src="https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=100&h=100&fit=crop" alt="User 3" className="ring-2 ring-white" />
          <Avatar initials="+5" className="ring-2 ring-white" />
        </div>
      </div>

      {/* Spaced avatars */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">Spaced Group</h3>
        <div className="flex gap-2">
          <Avatar initials="JD" status="online" size="sm" />
          <Avatar initials="AS" status="away" size="sm" />
          <Avatar initials="BJ" status="busy" size="sm" />
          <Avatar initials="MK" status="offline" size="sm" />
        </div>
      </div>
    </div>
}`,...(pe=(xe=d.parameters)==null?void 0:xe.docs)==null?void 0:pe.source},description:{story:"Avatar groups showing multiple users",...(ge=(he=d.parameters)==null?void 0:he.docs)==null?void 0:ge.description}}};var fe,ve,ue,Ne,ye;m.parameters={...m.parameters,docs:{...(fe=m.parameters)==null?void 0:fe.docs,source:{originalSource:`{
  render: () => <div className="flex flex-col gap-6 p-6">
      {/* User profile card */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">Profile Card</h3>
        <div className="flex items-center gap-3 p-4 bg-white border border-gray-200 rounded-lg">
          <Avatar src="https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop" alt="John Doe" status="online" size="lg" />
          <div>
            <h4 className="text-sm font-medium text-gray-900">John Doe</h4>
            <p className="text-xs text-gray-500">john.doe@example.com</p>
          </div>
        </div>
      </div>

      {/* Agent card */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">AI Agent</h3>
        <div className="flex items-center gap-3 p-4 bg-white border border-gray-200 rounded-lg">
          <Avatar initials="AI" shape="rounded" status="online" size="lg" />
          <div>
            <h4 className="text-sm font-medium text-gray-900">Email Assistant</h4>
            <p className="text-xs text-gray-500">Active · 3 permissions granted</p>
          </div>
        </div>
      </div>

      {/* Comment thread */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">Comment Thread</h3>
        <div className="space-y-3">
          <div className="flex gap-3">
            <Avatar initials="JD" size="sm" />
            <div className="flex-1">
              <p className="text-xs font-medium text-gray-900">John Doe</p>
              <p className="text-xs text-gray-600">This looks great! When can we ship?</p>
            </div>
          </div>
          <div className="flex gap-3">
            <Avatar initials="AS" size="sm" />
            <div className="flex-1">
              <p className="text-xs font-medium text-gray-900">Alice Smith</p>
              <p className="text-xs text-gray-600">Ready to go live tomorrow.</p>
            </div>
          </div>
        </div>
      </div>
    </div>
}`,...(ue=(ve=m.parameters)==null?void 0:ve.docs)==null?void 0:ue.source},description:{story:"Real-world use cases",...(ye=(Ne=m.parameters)==null?void 0:Ne.docs)==null?void 0:ye.description}}};var je,Ae,be,we,Se;x.parameters={...x.parameters,docs:{...(je=x.parameters)==null?void 0:je.docs,source:{originalSource:`{
  args: {
    initials: 'JD',
    size: 'md',
    shape: 'circle',
    status: 'online'
  }
}`,...(be=(Ae=x.parameters)==null?void 0:Ae.docs)==null?void 0:be.source},description:{story:"Interactive playground for testing all combinations",...(Se=(we=x.parameters)==null?void 0:we.docs)==null?void 0:Se.description}}};var Je,ze,De,Ie,ke;p.parameters={...p.parameters,docs:{...(Je=p.parameters)==null?void 0:Je.docs,source:{originalSource:`{
  render: () => <div className="flex flex-col gap-6">
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">Avatars with proper alt text</h3>
        <div className="flex gap-4">
          <Avatar src="https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop" alt="John Doe, Senior Developer" />
          <Avatar initials="AS" alt="Alice Smith, Product Manager" status="online" />
          <Avatar alt="Anonymous user" />
        </div>
      </div>

      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-gray-700">
          Status indicators with aria-label
        </h3>
        <div className="flex gap-4">
          <Avatar initials="JD" status="online" alt="John Doe" />
          <Avatar initials="AS" status="away" alt="Alice Smith" />
          <Avatar initials="BJ" status="busy" alt="Bob Johnson" />
          <Avatar initials="MK" status="offline" alt="Mary King" />
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
}`,...(De=(ze=p.parameters)==null?void 0:ze.docs)==null?void 0:De.source},description:{story:"Accessibility test - screen reader friendly",...(ke=(Ie=p.parameters)==null?void 0:Ie.docs)==null?void 0:ke.description}}};const Qe=["Default","Sizes","Shapes","WithInitials","WithStatus","WithImage","ImageFallback","CustomFallbackIcon","AvatarGroup","UseCases","Playground","Accessibility"];export{p as Accessibility,d as AvatarGroup,o as CustomFallbackIcon,a as Default,c as ImageFallback,x as Playground,i as Shapes,t as Sizes,m as UseCases,l as WithImage,r as WithInitials,n as WithStatus,Qe as __namedExportsOrder,$e as default};
