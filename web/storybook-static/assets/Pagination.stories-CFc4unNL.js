import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as Oe,r as I}from"./index-ClcD9ViR.js";import{c as p,a as Re}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const $e=Re("flex items-center justify-center gap-1",{variants:{size:{sm:"text-sm",md:"text-base",lg:"text-lg"}},defaultVariants:{size:"md"}}),V=Re("inline-flex items-center justify-center font-medium rounded-md transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-trust disabled:cursor-not-allowed disabled:opacity-40",{variants:{size:{sm:"min-w-[2rem] h-8 px-2 text-sm",md:"min-w-[2.5rem] h-10 px-3 text-base",lg:"min-w-[3rem] h-12 px-4 text-lg"},variant:{default:"bg-white text-navy-900 border border-gray-300 hover:bg-gray-50 hover:border-gray-400",active:"bg-trust-deep text-white border-transparent shadow-md",ghost:"bg-transparent text-navy-900 hover:bg-gray-100"}},defaultVariants:{size:"md",variant:"default"}});function Xe(a,t,n){if(t<=n)return Array.from({length:t},(o,S)=>S+1);const i=[],l=Math.floor(n/2);i.push(1);let g=Math.max(2,a-l),d=Math.min(t-1,a+l);a<=l+1&&(d=Math.min(t-1,n-1)),a>=t-l&&(g=Math.max(2,t-n+2)),g>2&&i.push("ellipsis");for(let o=g;o<=d;o++)i.push(o);return d<t-1&&i.push("ellipsis"),t>1&&i.push(t),i}const s=Oe.forwardRef(({currentPage:a,totalPages:t,onPageChange:n,size:i="md",variant:l="full",maxVisible:g=7,showInfo:d=!1,disabled:o=!1,className:S,...qe},Ue)=>{const k=a===1,z=a===t,M=t<=1,r=i??"md",We=()=>{!k&&!o&&n(a-1)},Be=()=>{!z&&!o&&n(a+1)},_e=c=>{!o&&c!==a&&n(c)},Ke=l==="full"?Xe(a,t,g):[];return e.jsxs("nav",{ref:Ue,role:"navigation","aria-label":"Pagination navigation",className:p("flex flex-col items-center gap-4",S),...qe,children:[e.jsxs("div",{className:$e({size:r}),children:[e.jsx("button",{type:"button",onClick:We,disabled:k||o||M,"aria-label":"Go to previous page",className:V({size:r,variant:"default"}),children:e.jsx("svg",{className:p("shrink-0",r==="sm"&&"w-4 h-4",r==="md"&&"w-5 h-5",r==="lg"&&"w-6 h-6"),fill:"none",stroke:"currentColor",viewBox:"0 0 24 24","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M15 19l-7-7 7-7"})})}),l==="full"&&Ke.map((c,Je)=>{if(c==="ellipsis")return e.jsx("span",{className:p("inline-flex items-center justify-center text-gray-500",r==="sm"&&"w-8 text-sm",r==="md"&&"w-10 text-base",r==="lg"&&"w-12 text-lg"),"aria-hidden":"true",children:"..."},`ellipsis-${Je}`);const m=c,T=m===a;return e.jsx("button",{type:"button",onClick:()=>_e(m),disabled:o,"aria-label":`Go to page ${m}`,"aria-current":T?"page":void 0,className:V({size:r,variant:T?"active":"default"}),children:m},m)}),e.jsx("button",{type:"button",onClick:Be,disabled:z||o||M,"aria-label":"Go to next page",className:V({size:r,variant:"default"}),children:e.jsx("svg",{className:p("shrink-0",r==="sm"&&"w-4 h-4",r==="md"&&"w-5 h-5",r==="lg"&&"w-6 h-6"),fill:"none",stroke:"currentColor",viewBox:"0 0 24 24","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M9 5l7 7-7 7"})})})]}),d&&t>0&&e.jsxs("p",{className:p("text-gray-600",r==="sm"&&"text-xs",r==="md"&&"text-sm",r==="lg"&&"text-base"),"aria-live":"polite","aria-atomic":"true",children:["Page ",a," of ",t]})]})});s.displayName="Pagination";s.__docgenInfo={description:`Pagination component for navigating through dataset pages.
Supports both simple (prev/next only) and full (with page numbers) modes.

@example
\`\`\`tsx
// Full pagination
<Pagination
  currentPage={5}
  totalPages={20}
  onPageChange={(page) => console.log('Go to page:', page)}
/>

// Simple mode with info
<Pagination
  variant="simple"
  showInfo
  currentPage={2}
  totalPages={10}
  onPageChange={handlePageChange}
/>

// Small size
<Pagination
  size="sm"
  currentPage={1}
  totalPages={5}
  onPageChange={handlePageChange}
/>
\`\`\``,methods:[],displayName:"Pagination",props:{currentPage:{required:!0,tsType:{name:"number"},description:"Current active page (1-indexed)"},totalPages:{required:!0,tsType:{name:"number"},description:"Total number of pages"},onPageChange:{required:!0,tsType:{name:"signature",type:"function",raw:"(page: number) => void",signature:{arguments:[{type:{name:"number"},name:"page"}],return:{name:"void"}}},description:"Callback when page changes"},size:{required:!1,tsType:{name:"union",raw:"'sm' | 'md' | 'lg'",elements:[{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"},{name:"literal",value:"'lg'"}]},description:"Size variant",defaultValue:{value:"'md'",computed:!1}},variant:{required:!1,tsType:{name:"union",raw:"'simple' | 'full'",elements:[{name:"literal",value:"'simple'"},{name:"literal",value:"'full'"}]},description:"Display mode",defaultValue:{value:"'full'",computed:!1}},maxVisible:{required:!1,tsType:{name:"number"},description:"Maximum visible page numbers (for full mode)",defaultValue:{value:"7",computed:!1}},showInfo:{required:!1,tsType:{name:"boolean"},description:'Show page info text (e.g., "Page 2 of 10")',defaultValue:{value:"false",computed:!1}},disabled:{required:!1,tsType:{name:"boolean"},description:"Disable all interactions",defaultValue:{value:"false",computed:!1}}},composes:["Omit"]};const ea={title:"Design System/Navigation/Pagination",component:s,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{currentPage:{control:{type:"number",min:1},description:"Current active page (1-indexed)",table:{type:{summary:"number"}}},totalPages:{control:{type:"number",min:1},description:"Total number of pages",table:{type:{summary:"number"}}},size:{control:"select",options:["sm","md","lg"],description:"Size variant",table:{type:{summary:"string"},defaultValue:{summary:"md"}}},variant:{control:"select",options:["simple","full"],description:"Display mode",table:{type:{summary:"string"},defaultValue:{summary:"full"}}},maxVisible:{control:{type:"number",min:3,max:15},description:"Maximum visible page numbers (full mode)",table:{type:{summary:"number"},defaultValue:{summary:7}}},showInfo:{control:"boolean",description:"Show page info text",table:{defaultValue:{summary:!1}}},disabled:{control:"boolean",description:"Disable all interactions",table:{defaultValue:{summary:!1}}}}},u={args:{currentPage:5,totalPages:10,onPageChange:a=>console.log("Go to page:",a),size:"md"}},x={args:{currentPage:2,totalPages:10,showInfo:!0,onPageChange:a=>console.log("Go to page:",a)}},h={render:()=>e.jsxs("div",{className:"space-y-8",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Small"}),e.jsx(s,{size:"sm",currentPage:3,totalPages:10,onPageChange:a=>console.log("Small - page:",a)})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Medium"}),e.jsx(s,{size:"md",currentPage:3,totalPages:10,onPageChange:a=>console.log("Medium - page:",a)})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Large"}),e.jsx(s,{size:"lg",currentPage:3,totalPages:10,onPageChange:a=>console.log("Large - page:",a)})]})]})},b={args:{variant:"simple",currentPage:5,totalPages:20,showInfo:!0,onPageChange:a=>console.log("Go to page:",a)}},P={render:()=>{const[a,t]=I.useState(15),n=50;return e.jsxs("div",{className:"space-y-8",children:[e.jsxs("div",{children:[e.jsxs("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:["Current Page: ",a," / ",n]}),e.jsx(s,{currentPage:a,totalPages:n,onPageChange:t,showInfo:!0})]}),e.jsxs("div",{className:"text-sm text-gray-600 space-y-2",children:[e.jsx("p",{className:"font-medium",children:"Try these pages to see ellipsis behavior:"}),e.jsxs("div",{className:"flex gap-2",children:[e.jsx("button",{onClick:()=>t(1),className:"px-3 py-1 bg-gray-100 hover:bg-gray-200 rounded text-sm",children:"First (1)"}),e.jsx("button",{onClick:()=>t(5),className:"px-3 py-1 bg-gray-100 hover:bg-gray-200 rounded text-sm",children:"Early (5)"}),e.jsx("button",{onClick:()=>t(25),className:"px-3 py-1 bg-gray-100 hover:bg-gray-200 rounded text-sm",children:"Middle (25)"}),e.jsx("button",{onClick:()=>t(45),className:"px-3 py-1 bg-gray-100 hover:bg-gray-200 rounded text-sm",children:"Late (45)"}),e.jsx("button",{onClick:()=>t(50),className:"px-3 py-1 bg-gray-100 hover:bg-gray-200 rounded text-sm",children:"Last (50)"})]})]})]})}},f={render:()=>e.jsxs("div",{className:"space-y-8",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"First Page (Previous disabled)"}),e.jsx(s,{currentPage:1,totalPages:10,onPageChange:a=>console.log("First page - page:",a),showInfo:!0})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Last Page (Next disabled)"}),e.jsx(s,{currentPage:10,totalPages:10,onPageChange:a=>console.log("Last page - page:",a),showInfo:!0})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Single Page (Both disabled)"}),e.jsx(s,{currentPage:1,totalPages:1,onPageChange:a=>console.log("Single page - page:",a),showInfo:!0})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Few Pages (No ellipsis needed)"}),e.jsx(s,{currentPage:2,totalPages:5,onPageChange:a=>console.log("Few pages - page:",a),showInfo:!0})]})]})},v={args:{currentPage:5,totalPages:10,disabled:!0,onPageChange:a=>console.log("Go to page:",a)}},y={render:()=>e.jsxs("div",{className:"space-y-8",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"maxVisible: 5 (Compact)"}),e.jsx(s,{currentPage:10,totalPages:20,maxVisible:5,onPageChange:a=>console.log("maxVisible 5 - page:",a)})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"maxVisible: 7 (Default)"}),e.jsx(s,{currentPage:10,totalPages:20,maxVisible:7,onPageChange:a=>console.log("maxVisible 7 - page:",a)})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"maxVisible: 11 (Extended)"}),e.jsx(s,{currentPage:10,totalPages:20,maxVisible:11,onPageChange:a=>console.log("maxVisible 11 - page:",a)})]})]})},N={render:()=>{const[a,t]=I.useState(1),n=15;return e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{className:"p-6 bg-gray-50 rounded-lg",children:[e.jsxs("h3",{className:"text-lg font-semibold mb-2",children:["Current Page: ",a]}),e.jsx("p",{className:"text-sm text-gray-600",children:"Click on page numbers or use prev/next buttons to navigate."})]}),e.jsx(s,{currentPage:a,totalPages:n,onPageChange:t,showInfo:!0}),e.jsxs("div",{className:"flex gap-2",children:[e.jsx("button",{onClick:()=>t(1),className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors text-sm",children:"Reset to First Page"}),e.jsx("button",{onClick:()=>t(Math.ceil(n/2)),className:"px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors text-sm",children:"Jump to Middle"})]})]})}},w={render:()=>e.jsxs("div",{className:"space-y-8 w-full",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Mobile (320px) - Use simple mode or small size"}),e.jsx("div",{className:"w-80 border-2 border-dashed border-gray-300 p-4",children:e.jsx(s,{size:"sm",variant:"simple",currentPage:5,totalPages:20,showInfo:!0,onPageChange:a=>console.log("Mobile - page:",a)})})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Tablet (768px) - Medium with reduced maxVisible"}),e.jsx("div",{className:"w-[768px] border-2 border-dashed border-gray-300 p-4",children:e.jsx(s,{size:"md",currentPage:5,totalPages:20,maxVisible:5,onPageChange:a=>console.log("Tablet - page:",a)})})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Desktop (1024px+) - Full pagination"}),e.jsx("div",{className:"w-full border-2 border-dashed border-gray-300 p-4",children:e.jsx(s,{size:"md",currentPage:5,totalPages:20,showInfo:!0,onPageChange:a=>console.log("Desktop - page:",a)})})]})]})},C={args:{currentPage:5,totalPages:20,size:"md",variant:"full",maxVisible:7,showInfo:!1,disabled:!1,onPageChange:a=>console.log("Go to page:",a)},parameters:{docs:{description:{story:"Experiment with all pagination props using the controls below. Try different sizes, variants, and page counts to see how the component adapts."}}}},j={render:()=>{const[a,t]=I.useState(3);return e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Keyboard Navigation"}),e.jsx("p",{className:"text-sm text-neutral-600 mb-4",children:'Press Tab to focus buttons, Enter or Space to activate. Current page is announced as "current page" to screen readers.'}),e.jsx(s,{currentPage:a,totalPages:10,onPageChange:t,showInfo:!0})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Focus Indicators"}),e.jsx("p",{className:"text-sm text-neutral-600 mb-4",children:"Visible focus rings meet WCAG 2.1 AA contrast requirements. Try tabbing through the buttons."}),e.jsx(s,{currentPage:5,totalPages:10,onPageChange:n=>console.log("Focus test - page:",n)})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Disabled State"}),e.jsx("p",{className:"text-sm text-neutral-600 mb-4",children:"Disabled pagination prevents all interactions and is announced to screen readers."}),e.jsx(s,{currentPage:5,totalPages:10,disabled:!0,onPageChange:n=>console.log("Disabled - page:",n)})]}),e.jsxs("div",{children:[e.jsx("p",{className:"text-sm font-medium text-neutral-700 mb-3",children:"Live Region Updates"}),e.jsx("p",{className:"text-sm text-neutral-600 mb-4",children:'Page info uses aria-live="polite" to announce changes to screen readers without interrupting.'}),e.jsx(s,{currentPage:a,totalPages:10,showInfo:!0,onPageChange:t})]})]})},parameters:{docs:{description:{story:"Pagination is fully accessible with keyboard support, visible focus indicators, and proper ARIA attributes. All navigation controls are keyboard accessible and properly labeled for screen readers."}}}};var D,A,F,L,E;u.parameters={...u.parameters,docs:{...(D=u.parameters)==null?void 0:D.docs,source:{originalSource:`{
  args: {
    currentPage: 5,
    totalPages: 10,
    onPageChange: page => console.log('Go to page:', page),
    size: 'md'
  }
}`,...(F=(A=u.parameters)==null?void 0:A.docs)==null?void 0:F.source},description:{story:`Default pagination with numbered pages.
Shows standard pagination with page numbers and prev/next buttons.`,...(E=(L=u.parameters)==null?void 0:L.docs)==null?void 0:E.description}}};var G,R,q,U,W;x.parameters={...x.parameters,docs:{...(G=x.parameters)==null?void 0:G.docs,source:{originalSource:`{
  args: {
    currentPage: 2,
    totalPages: 10,
    showInfo: true,
    onPageChange: page => console.log('Go to page:', page)
  }
}`,...(q=(R=x.parameters)==null?void 0:R.docs)==null?void 0:q.source},description:{story:`Pagination with page info display.
Shows "Page X of Y" below the pagination controls.`,...(W=(U=x.parameters)==null?void 0:U.docs)==null?void 0:W.description}}};var B,_,K,J,O;h.parameters={...h.parameters,docs:{...(B=h.parameters)==null?void 0:B.docs,source:{originalSource:`{
  render: () => <div className="space-y-8">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">Small</p>
        <Pagination size="sm" currentPage={3} totalPages={10} onPageChange={page => console.log('Small - page:', page)} />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">Medium</p>
        <Pagination size="md" currentPage={3} totalPages={10} onPageChange={page => console.log('Medium - page:', page)} />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">Large</p>
        <Pagination size="lg" currentPage={3} totalPages={10} onPageChange={page => console.log('Large - page:', page)} />
      </div>
    </div>
}`,...(K=(_=h.parameters)==null?void 0:_.docs)==null?void 0:K.source},description:{story:`All pagination sizes side-by-side.
- sm: Compact tables, mobile layouts
- md: Standard desktop layouts
- lg: Prominent navigation, hero sections`,...(O=(J=h.parameters)==null?void 0:J.docs)==null?void 0:O.description}}};var $,X,Y,H,Q;b.parameters={...b.parameters,docs:{...($=b.parameters)==null?void 0:$.docs,source:{originalSource:`{
  args: {
    variant: 'simple',
    currentPage: 5,
    totalPages: 20,
    showInfo: true,
    onPageChange: page => console.log('Go to page:', page)
  }
}`,...(Y=(X=b.parameters)==null?void 0:X.docs)==null?void 0:Y.source},description:{story:`Simple mode with only prev/next buttons.
Useful when you want minimal UI or don't need direct page access.`,...(Q=(H=b.parameters)==null?void 0:H.docs)==null?void 0:Q.description}}};var Z,ee,ae,te,se;P.parameters={...P.parameters,docs:{...(Z=P.parameters)==null?void 0:Z.docs,source:{originalSource:`{
  render: () => {
    const [currentPage, setCurrentPage] = useState(15);
    const totalPages = 50;
    return <div className="space-y-8">
        <div>
          <p className="text-sm font-medium text-neutral-700 mb-3">
            Current Page: {currentPage} / {totalPages}
          </p>
          <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} showInfo />
        </div>

        <div className="text-sm text-gray-600 space-y-2">
          <p className="font-medium">Try these pages to see ellipsis behavior:</p>
          <div className="flex gap-2">
            <button onClick={() => setCurrentPage(1)} className="px-3 py-1 bg-gray-100 hover:bg-gray-200 rounded text-sm">
              First (1)
            </button>
            <button onClick={() => setCurrentPage(5)} className="px-3 py-1 bg-gray-100 hover:bg-gray-200 rounded text-sm">
              Early (5)
            </button>
            <button onClick={() => setCurrentPage(25)} className="px-3 py-1 bg-gray-100 hover:bg-gray-200 rounded text-sm">
              Middle (25)
            </button>
            <button onClick={() => setCurrentPage(45)} className="px-3 py-1 bg-gray-100 hover:bg-gray-200 rounded text-sm">
              Late (45)
            </button>
            <button onClick={() => setCurrentPage(50)} className="px-3 py-1 bg-gray-100 hover:bg-gray-200 rounded text-sm">
              Last (50)
            </button>
          </div>
        </div>
      </div>;
  }
}`,...(ae=(ee=P.parameters)==null?void 0:ee.docs)==null?void 0:ae.source},description:{story:`Pagination with ellipsis for large datasets.
Smart truncation shows relevant pages with ellipsis.`,...(se=(te=P.parameters)==null?void 0:te.docs)==null?void 0:se.description}}};var ne,re,oe,ie,le;f.parameters={...f.parameters,docs:{...(ne=f.parameters)==null?void 0:ne.docs,source:{originalSource:`{
  render: () => <div className="space-y-8">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          First Page (Previous disabled)
        </p>
        <Pagination currentPage={1} totalPages={10} onPageChange={page => console.log('First page - page:', page)} showInfo />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Last Page (Next disabled)
        </p>
        <Pagination currentPage={10} totalPages={10} onPageChange={page => console.log('Last page - page:', page)} showInfo />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Single Page (Both disabled)
        </p>
        <Pagination currentPage={1} totalPages={1} onPageChange={page => console.log('Single page - page:', page)} showInfo />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Few Pages (No ellipsis needed)
        </p>
        <Pagination currentPage={2} totalPages={5} onPageChange={page => console.log('Few pages - page:', page)} showInfo />
      </div>
    </div>
}`,...(oe=(re=f.parameters)==null?void 0:re.docs)==null?void 0:oe.source},description:{story:`Edge cases: first page, last page, single page, and empty.
Shows how pagination handles boundary conditions.`,...(le=(ie=f.parameters)==null?void 0:ie.docs)==null?void 0:le.description}}};var ge,de,ce,me,pe;v.parameters={...v.parameters,docs:{...(ge=v.parameters)==null?void 0:ge.docs,source:{originalSource:`{
  args: {
    currentPage: 5,
    totalPages: 10,
    disabled: true,
    onPageChange: page => console.log('Go to page:', page)
  }
}`,...(ce=(de=v.parameters)==null?void 0:de.docs)==null?void 0:ce.source},description:{story:`Disabled state prevents all interactions.
Useful during data loading or when pagination should be temporarily locked.`,...(pe=(me=v.parameters)==null?void 0:me.docs)==null?void 0:pe.description}}};var ue,xe,he,be,Pe;y.parameters={...y.parameters,docs:{...(ue=y.parameters)==null?void 0:ue.docs,source:{originalSource:`{
  render: () => <div className="space-y-8">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          maxVisible: 5 (Compact)
        </p>
        <Pagination currentPage={10} totalPages={20} maxVisible={5} onPageChange={page => console.log('maxVisible 5 - page:', page)} />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          maxVisible: 7 (Default)
        </p>
        <Pagination currentPage={10} totalPages={20} maxVisible={7} onPageChange={page => console.log('maxVisible 7 - page:', page)} />
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          maxVisible: 11 (Extended)
        </p>
        <Pagination currentPage={10} totalPages={20} maxVisible={11} onPageChange={page => console.log('maxVisible 11 - page:', page)} />
      </div>
    </div>
}`,...(he=(xe=y.parameters)==null?void 0:xe.docs)==null?void 0:he.source},description:{story:`Different maxVisible values control ellipsis behavior.
Lower values show fewer page numbers, higher values show more.`,...(Pe=(be=y.parameters)==null?void 0:be.docs)==null?void 0:Pe.description}}};var fe,ve,ye,Ne,we;N.parameters={...N.parameters,docs:{...(fe=N.parameters)==null?void 0:fe.docs,source:{originalSource:`{
  render: () => {
    const [currentPage, setCurrentPage] = useState(1);
    const totalPages = 15;
    return <div className="space-y-6">
        <div className="p-6 bg-gray-50 rounded-lg">
          <h3 className="text-lg font-semibold mb-2">
            Current Page: {currentPage}
          </h3>
          <p className="text-sm text-gray-600">
            Click on page numbers or use prev/next buttons to navigate.
          </p>
        </div>

        <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} showInfo />

        <div className="flex gap-2">
          <button onClick={() => setCurrentPage(1)} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors text-sm">
            Reset to First Page
          </button>
          <button onClick={() => setCurrentPage(Math.ceil(totalPages / 2))} className="px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors text-sm">
            Jump to Middle
          </button>
        </div>
      </div>;
  }
}`,...(ye=(ve=N.parameters)==null?void 0:ve.docs)==null?void 0:ye.source},description:{story:`Interactive example with state management.
Shows full pagination behavior with stateful page changes.`,...(we=(Ne=N.parameters)==null?void 0:Ne.docs)==null?void 0:we.description}}};var Ce,je,Se,Ve,Ie;w.parameters={...w.parameters,docs:{...(Ce=w.parameters)==null?void 0:Ce.docs,source:{originalSource:`{
  render: () => <div className="space-y-8 w-full">
      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Mobile (320px) - Use simple mode or small size
        </p>
        <div className="w-80 border-2 border-dashed border-gray-300 p-4">
          <Pagination size="sm" variant="simple" currentPage={5} totalPages={20} showInfo onPageChange={page => console.log('Mobile - page:', page)} />
        </div>
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Tablet (768px) - Medium with reduced maxVisible
        </p>
        <div className="w-[768px] border-2 border-dashed border-gray-300 p-4">
          <Pagination size="md" currentPage={5} totalPages={20} maxVisible={5} onPageChange={page => console.log('Tablet - page:', page)} />
        </div>
      </div>

      <div>
        <p className="text-sm font-medium text-neutral-700 mb-3">
          Desktop (1024px+) - Full pagination
        </p>
        <div className="w-full border-2 border-dashed border-gray-300 p-4">
          <Pagination size="md" currentPage={5} totalPages={20} showInfo onPageChange={page => console.log('Desktop - page:', page)} />
        </div>
      </div>
    </div>
}`,...(Se=(je=w.parameters)==null?void 0:je.docs)==null?void 0:Se.source},description:{story:`Responsive layout example.
Shows how pagination adapts to different container widths.`,...(Ie=(Ve=w.parameters)==null?void 0:Ve.docs)==null?void 0:Ie.description}}};var ke,ze,Me,Te,De;C.parameters={...C.parameters,docs:{...(ke=C.parameters)==null?void 0:ke.docs,source:{originalSource:`{
  args: {
    currentPage: 5,
    totalPages: 20,
    size: 'md',
    variant: 'full',
    maxVisible: 7,
    showInfo: false,
    disabled: false,
    onPageChange: page => console.log('Go to page:', page)
  },
  parameters: {
    docs: {
      description: {
        story: 'Experiment with all pagination props using the controls below. Try different sizes, variants, and page counts to see how the component adapts.'
      }
    }
  }
}`,...(Me=(ze=C.parameters)==null?void 0:ze.docs)==null?void 0:Me.source},description:{story:`Interactive playground to experiment with all props.
Try different combinations of size, variant, and page counts.`,...(De=(Te=C.parameters)==null?void 0:Te.docs)==null?void 0:De.description}}};var Ae,Fe,Le,Ee,Ge;j.parameters={...j.parameters,docs:{...(Ae=j.parameters)==null?void 0:Ae.docs,source:{originalSource:`{
  render: () => {
    const [currentPage, setCurrentPage] = useState(3);
    return <div className="space-y-6">
        <div>
          <p className="text-sm font-medium text-neutral-700 mb-3">
            Keyboard Navigation
          </p>
          <p className="text-sm text-neutral-600 mb-4">
            Press Tab to focus buttons, Enter or Space to activate. Current
            page is announced as "current page" to screen readers.
          </p>
          <Pagination currentPage={currentPage} totalPages={10} onPageChange={setCurrentPage} showInfo />
        </div>

        <div>
          <p className="text-sm font-medium text-neutral-700 mb-3">
            Focus Indicators
          </p>
          <p className="text-sm text-neutral-600 mb-4">
            Visible focus rings meet WCAG 2.1 AA contrast requirements. Try
            tabbing through the buttons.
          </p>
          <Pagination currentPage={5} totalPages={10} onPageChange={page => console.log('Focus test - page:', page)} />
        </div>

        <div>
          <p className="text-sm font-medium text-neutral-700 mb-3">
            Disabled State
          </p>
          <p className="text-sm text-neutral-600 mb-4">
            Disabled pagination prevents all interactions and is announced to
            screen readers.
          </p>
          <Pagination currentPage={5} totalPages={10} disabled onPageChange={page => console.log('Disabled - page:', page)} />
        </div>

        <div>
          <p className="text-sm font-medium text-neutral-700 mb-3">
            Live Region Updates
          </p>
          <p className="text-sm text-neutral-600 mb-4">
            Page info uses aria-live="polite" to announce changes to screen
            readers without interrupting.
          </p>
          <Pagination currentPage={currentPage} totalPages={10} showInfo onPageChange={setCurrentPage} />
        </div>
      </div>;
  },
  parameters: {
    docs: {
      description: {
        story: 'Pagination is fully accessible with keyboard support, visible focus indicators, and proper ARIA attributes. All navigation controls are keyboard accessible and properly labeled for screen readers.'
      }
    }
  }
}`,...(Le=(Fe=j.parameters)==null?void 0:Fe.docs)==null?void 0:Le.source},description:{story:`Accessibility features demonstration.
All pagination controls support:
- Keyboard navigation (Tab to focus, Enter/Space to activate)
- Screen reader labels with ARIA attributes
- Visible focus indicators (ring on focus)
- aria-current for current page
- aria-live for page info updates`,...(Ge=(Ee=j.parameters)==null?void 0:Ee.docs)==null?void 0:Ge.description}}};const aa=["Default","WithInfo","Sizes","SimpleMode","Ellipsis","EdgeCases","Disabled","MaxVisibleVariants","Interactive","Responsive","Playground","Accessibility"];export{j as Accessibility,u as Default,v as Disabled,f as EdgeCases,P as Ellipsis,N as Interactive,y as MaxVisibleVariants,C as Playground,w as Responsive,b as SimpleMode,h as Sizes,x as WithInfo,aa as __namedExportsOrder,ea as default};
