import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as f}from"./index-ClcD9ViR.js";import{c as g,a as x}from"./cn-JCLedEej.js";const y=x("inline-flex items-center justify-center font-medium rounded-md transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60",{variants:{variant:{primary:"bg-trust-deep text-white border-transparent hover:bg-trust-hover focus:ring-trust shadow-md hover:shadow-lg hover:-translate-y-px",secondary:"bg-emerald-600 text-white border-transparent hover:bg-emerald-700 focus:ring-emerald-500 shadow-md hover:shadow-lg hover:-translate-y-px",outline:"bg-white text-navy-900 border border-gray-300 hover:bg-gray-50 hover:border-gray-400 focus:ring-navy-700 shadow-sm",ghost:"bg-transparent text-navy-900 hover:bg-gray-100 focus:ring-navy-700",danger:"bg-red-600 text-white border-transparent hover:bg-red-700 focus:ring-red-500 shadow-md hover:shadow-lg hover:-translate-y-px"},size:{sm:"px-3 py-1.5 text-sm h-8",md:"px-4 py-2 text-base h-10",lg:"px-6 py-3 text-lg h-12",xl:"px-8 py-4 text-xl h-14"},fullWidth:{true:"w-full",false:"w-auto"}},defaultVariants:{variant:"primary",size:"md",fullWidth:!1}}),n=f.forwardRef(({variant:s,size:o,fullWidth:i,className:l,children:d,isLoading:t=!1,disabled:c=!1,iconBefore:a,iconAfter:r,type:u="button",...m},p)=>{const h=c||t;return e.jsxs("button",{ref:p,type:u,disabled:h,className:g(y({variant:s,size:o,fullWidth:i}),l),...m,children:[t&&e.jsxs("svg",{className:"animate-spin -ml-1 mr-2 h-4 w-4",xmlns:"http://www.w3.org/2000/svg",fill:"none",viewBox:"0 0 24 24","aria-hidden":"true","data-testid":"button-spinner",children:[e.jsx("circle",{className:"opacity-25",cx:"12",cy:"12",r:"10",stroke:"currentColor",strokeWidth:"4"}),e.jsx("path",{className:"opacity-75",fill:"currentColor",d:"M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"})]}),!t&&a&&e.jsx("span",{className:"mr-2 -ml-1 flex items-center",children:a}),d,r&&e.jsx("span",{className:"ml-2 -mr-1 flex items-center",children:r})]})});n.displayName="Button";n.__docgenInfo={description:`Button component with consistent styling and behavior.
Shows spinner when loading and disables interactions.

@example
\`\`\`tsx
<Button variant="primary" size="md" onClick={handleClick}>
  Click me
</Button>

<Button variant="outline" isLoading>
  Loading...
</Button>

<Button variant="danger" iconBefore={<TrashIcon />}>
  Delete
</Button>
\`\`\``,methods:[],displayName:"Button",props:{isLoading:{required:!1,tsType:{name:"boolean"},description:"Whether button is in loading state",defaultValue:{value:"false",computed:!1}},iconBefore:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Icon to display before children"},iconAfter:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Icon to display after children"},disabled:{defaultValue:{value:"false",computed:!1},required:!1},type:{defaultValue:{value:"'button'",computed:!1},required:!1}},composes:["VariantProps"]};export{n as B};
