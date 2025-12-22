import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as Ye}from"./index-ClcD9ViR.js";import{c as y,a as Oe}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const I=Oe("flex gap-2 mt-1 text-sm text-error-dark transition-all duration-200",{variants:{size:{sm:"text-xs",md:"text-sm"}},defaultVariants:{size:"md"}}),P=Oe("flex-shrink-0 text-error-primary",{variants:{size:{sm:"w-3.5 h-3.5 mt-0.5",md:"w-4 h-4 mt-0.5"}},defaultVariants:{size:"md"}}),He=({className:s})=>e.jsx("svg",{className:y("w-4 h-4",s),fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"})}),r=Ye.forwardRef(({message:s,icon:We,hideIcon:w=!1,errors:h,suggestions:f,helperText:N,fieldLabel:k,size:a="md",className:Be,...Ve},Ge)=>{const E=We||e.jsx(He,{}),F=h&&h.length>0?h:[s],Ze=F.length>1;return e.jsxs("div",{ref:Ge,role:"alert","aria-live":"polite","aria-atomic":"true",className:y("space-y-1",Be),...Ve,children:[Ze?e.jsx("ul",{className:"space-y-1",children:F.map((v,j)=>e.jsxs("li",{className:I({size:a}),children:[!w&&e.jsx("span",{className:P({size:a}),children:E}),e.jsx("span",{className:"flex-1",children:v})]},j))}):e.jsxs("div",{className:I({size:a}),children:[!w&&e.jsx("span",{className:P({size:a}),children:E}),e.jsx("span",{className:"flex-1",children:s})]}),f&&f.length>0&&e.jsx("div",{className:y("ml-6 text-xs text-error-dark/80 space-y-0.5",a==="sm"&&"ml-5"),children:f.map((v,j)=>e.jsxs("div",{className:"flex items-start gap-1.5",children:[e.jsx("span",{className:"text-error-primary mt-0.5",children:"•"}),e.jsx("span",{children:v})]},j))}),N&&e.jsx("div",{className:y("ml-6 text-xs text-error-dark/70 italic",a==="sm"&&"ml-5"),children:N}),k&&e.jsxs("span",{className:"sr-only",children:["Error for field: ",k]})]})});r.displayName="InlineError";r.__docgenInfo={description:`InlineError component for displaying field-level validation errors.
Typically positioned below form inputs to provide immediate feedback.

@example
\`\`\`tsx
<InlineError message="Email address is required" />

<InlineError
  message="Password is too weak"
  suggestions={["Use at least 8 characters", "Include numbers and symbols"]}
/>

<InlineError
  errors={["Username is required", "Username must be at least 3 characters"]}
/>
\`\`\``,methods:[],displayName:"InlineError",props:{message:{required:!0,tsType:{name:"string"},description:"Error message text"},icon:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Optional custom error icon (overrides default)"},hideIcon:{required:!1,tsType:{name:"boolean"},description:"Hide the default error icon",defaultValue:{value:"false",computed:!1}},errors:{required:!1,tsType:{name:"Array",elements:[{name:"string"}],raw:"string[]"},description:"Array of validation error messages (for multiple errors)"},suggestions:{required:!1,tsType:{name:"Array",elements:[{name:"string"}],raw:"string[]"},description:"Optional suggestion text to help user fix the error"},helperText:{required:!1,tsType:{name:"string"},description:"Optional helper text shown below error"},fieldLabel:{required:!1,tsType:{name:"string"},description:"Associated field label (for context)"},size:{required:!1,tsType:{name:"union",raw:"'sm' | 'md'",elements:[{name:"literal",value:"'sm'"},{name:"literal",value:"'md'"}]},description:"Size variant",defaultValue:{value:"'md'",computed:!1}}},composes:["Omit"]};const Xe={title:"Design System/Feedback/InlineError",component:r,parameters:{layout:"padded"},tags:["autodocs"],argTypes:{message:{control:"text",description:"Error message text"},errors:{control:"object",description:"Array of error messages (for multiple errors)"},suggestions:{control:"object",description:"Array of suggestion texts to help fix the error"},helperText:{control:"text",description:"Optional helper text shown below error"},fieldLabel:{control:"text",description:"Associated field label for accessibility"},hideIcon:{control:"boolean",description:"Hide the default error icon"},size:{control:"select",options:["sm","md"],description:"Size variant"}}},i={args:{message:"This field is required"}},o={render:()=>{const s=()=>e.jsx("svg",{className:"w-4 h-4",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"})});return e.jsxs("div",{className:"space-y-4 max-w-md",children:[e.jsx(r,{message:"Default error icon (X in circle)"}),e.jsx(r,{message:"Custom icon (info circle)",icon:e.jsx(s,{})})]})},args:{message:"Error with icon"}},d={render:()=>e.jsxs("div",{className:"max-w-md space-y-6",children:[e.jsxs("div",{children:[e.jsx("label",{htmlFor:"email-valid",className:"block text-sm font-medium text-gray-700 mb-1",children:"Email Address (Valid)"}),e.jsx("input",{id:"email-valid",type:"email",placeholder:"user@example.com",value:"user@example.com",className:"w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-trust-deep focus:border-transparent",readOnly:!0})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"email-invalid",className:"block text-sm font-medium text-gray-700 mb-1",children:"Email Address (Invalid)"}),e.jsx("input",{id:"email-invalid",type:"email",placeholder:"user@example.com",value:"invalid-email",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"email-error"}),e.jsx("div",{id:"email-error",children:e.jsx(r,{message:"Please enter a valid email address",fieldLabel:"Email Address"})})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"password-empty",className:"block text-sm font-medium text-gray-700 mb-1",children:"Password (Required)"}),e.jsx("input",{id:"password-empty",type:"password",placeholder:"Enter password",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"password-error"}),e.jsx("div",{id:"password-error",children:e.jsx(r,{message:"Password is required",fieldLabel:"Password"})})]})]}),args:{message:"Field-level error"}},l={render:()=>e.jsxs("div",{className:"max-w-md space-y-6",children:[e.jsxs("div",{children:[e.jsx("label",{htmlFor:"username",className:"block text-sm font-medium text-gray-700 mb-1",children:"Username"}),e.jsx("input",{id:"username",type:"text",placeholder:"Enter username",value:"ab",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"username-errors"}),e.jsx("div",{id:"username-errors",children:e.jsx(r,{errors:["Username must be at least 3 characters","Username can only contain letters and numbers","Username is already taken"],message:"Multiple validation errors",fieldLabel:"Username"})})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"phone",className:"block text-sm font-medium text-gray-700 mb-1",children:"Phone Number"}),e.jsx("input",{id:"phone",type:"tel",placeholder:"(555) 123-4567",value:"123",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"phone-errors"}),e.jsx("div",{id:"phone-errors",children:e.jsx(r,{errors:["Phone number must be 10 digits","Phone number format is invalid"],message:"Phone validation errors",fieldLabel:"Phone Number"})})]})]}),args:{errors:["Username must be at least 3 characters","Username can only contain letters and numbers"],message:"Multiple errors"}},t={render:()=>e.jsxs("div",{className:"max-w-md space-y-6",children:[e.jsxs("div",{children:[e.jsx("label",{htmlFor:"password-weak",className:"block text-sm font-medium text-gray-700 mb-1",children:"Password"}),e.jsx("input",{id:"password-weak",type:"password",placeholder:"Enter password",value:"abc123",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"password-suggestions"}),e.jsx("div",{id:"password-suggestions",children:e.jsx(r,{message:"Password is too weak",suggestions:["Use at least 8 characters","Include uppercase and lowercase letters","Add numbers and special symbols","Avoid common words or patterns"],fieldLabel:"Password"})})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"upload",className:"block text-sm font-medium text-gray-700 mb-1",children:"Profile Photo"}),e.jsx("input",{id:"upload",type:"file",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"upload-suggestions"}),e.jsx("div",{id:"upload-suggestions",children:e.jsx(r,{message:"File size is too large (5.2 MB)",suggestions:["Maximum file size is 2 MB","Try compressing your image","Supported formats: JPG, PNG, GIF"],fieldLabel:"Profile Photo"})})]})]}),args:{message:"Password is too weak",suggestions:["Use at least 8 characters","Include uppercase and lowercase letters","Add numbers and special symbols"]}},n={render:()=>e.jsxs("div",{className:"max-w-md space-y-6",children:[e.jsxs("div",{children:[e.jsx("label",{htmlFor:"email-taken",className:"block text-sm font-medium text-gray-700 mb-1",children:"Email Address"}),e.jsx("input",{id:"email-taken",type:"email",placeholder:"user@example.com",value:"existing@example.com",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"email-link-error"}),e.jsx("div",{id:"email-link-error",children:e.jsx(r,{message:e.jsxs(e.Fragment,{children:["This email is already registered."," ",e.jsx("a",{href:"/login",className:"underline hover:text-error-primary font-medium",children:"Sign in instead?"})]}),fieldLabel:"Email Address"})})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"code-invalid",className:"block text-sm font-medium text-gray-700 mb-1",children:"Verification Code"}),e.jsx("input",{id:"code-invalid",type:"text",placeholder:"Enter 6-digit code",value:"123456",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"code-link-error"}),e.jsx("div",{id:"code-link-error",children:e.jsx(r,{message:e.jsxs(e.Fragment,{children:["Invalid verification code."," ",e.jsx("button",{type:"button",onClick:()=>alert("Resending code..."),className:"underline hover:text-error-primary font-medium",children:"Resend code"})]}),fieldLabel:"Verification Code"})})]})]}),args:{message:"Error with link"}},m={render:()=>e.jsxs("div",{className:"max-w-md space-y-6",children:[e.jsxs("div",{children:[e.jsx("label",{htmlFor:"billing-zip",className:"block text-sm font-medium text-gray-700 mb-1",children:"Billing ZIP Code"}),e.jsx("input",{id:"billing-zip",type:"text",placeholder:"12345",value:"ABC",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"zip-error"}),e.jsx("div",{id:"zip-error",children:e.jsx(r,{message:"ZIP code must be 5 digits",fieldLabel:"Billing ZIP Code",helperText:"Example: 12345"})})]}),e.jsx("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:e.jsxs("p",{className:"text-sm text-gray-700",children:["The ",e.jsx("code",{children:"fieldLabel"})," prop adds screen reader context (hidden visually). It helps screen reader users understand which field has an error."]})})]}),args:{message:"ZIP code must be 5 digits",fieldLabel:"Billing ZIP Code"}},c={render:()=>e.jsxs("div",{className:"max-w-md space-y-6",children:[e.jsxs("div",{children:[e.jsx("label",{htmlFor:"amount",className:"block text-sm font-medium text-gray-700 mb-1",children:"Transfer Amount"}),e.jsx("input",{id:"amount",type:"number",placeholder:"0.00",value:"25000",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"amount-error"}),e.jsx("div",{id:"amount-error",children:e.jsx(r,{message:"Transfer amount exceeds daily limit",helperText:"Your daily transfer limit is $10,000. Contact support to increase your limit.",fieldLabel:"Transfer Amount"})})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"domain",className:"block text-sm font-medium text-gray-700 mb-1",children:"Custom Domain"}),e.jsx("input",{id:"domain",type:"text",placeholder:"example.com",value:"invalid..domain",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"domain-error"}),e.jsx("div",{id:"domain-error",children:e.jsx(r,{message:"Invalid domain format",helperText:"Domain must contain only letters, numbers, hyphens, and periods.",fieldLabel:"Custom Domain"})})]})]}),args:{message:"Transfer amount exceeds daily limit",helperText:"Your daily transfer limit is $10,000. Contact support to increase your limit."}},u={render:()=>e.jsxs("div",{className:"max-w-md space-y-6",children:[e.jsxs("div",{children:[e.jsx("label",{htmlFor:"field-sm",className:"block text-xs font-medium text-gray-700 mb-1",children:"Small Field"}),e.jsx("input",{id:"field-sm",type:"text",className:"w-full px-2 py-1 text-sm border-2 border-error-primary rounded-md","aria-invalid":"true","aria-describedby":"error-sm"}),e.jsx("div",{id:"error-sm",children:e.jsx(r,{size:"sm",message:"Small size error message",suggestions:["Suggestion in small size"]})})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"field-md",className:"block text-sm font-medium text-gray-700 mb-1",children:"Medium Field (Default)"}),e.jsx("input",{id:"field-md",type:"text",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md","aria-invalid":"true","aria-describedby":"error-md"}),e.jsx("div",{id:"error-md",children:e.jsx(r,{size:"md",message:"Medium size error message",suggestions:["Suggestion in medium size"]})})]})]}),args:{message:"Size variants",size:"md"}},p={render:()=>e.jsxs("div",{className:"max-w-md space-y-6",children:[e.jsxs("div",{children:[e.jsx("label",{htmlFor:"simple",className:"block text-sm font-medium text-gray-700 mb-1",children:"Simple Field"}),e.jsx("input",{id:"simple",type:"text",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md","aria-invalid":"true","aria-describedby":"simple-error"}),e.jsx("div",{id:"simple-error",children:e.jsx(r,{message:"This field is required",hideIcon:!0})})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"minimal",className:"block text-sm font-medium text-gray-700 mb-1",children:"Minimal Error"}),e.jsx("input",{id:"minimal",type:"text",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md","aria-invalid":"true","aria-describedby":"minimal-error"}),e.jsx("div",{id:"minimal-error",children:e.jsx(r,{message:"Please enter a valid value",hideIcon:!0,helperText:"No icon for cleaner appearance"})})]})]}),args:{message:"Error without icon",hideIcon:!0}},b={render:s=>e.jsxs("div",{className:"max-w-md",children:[e.jsx("label",{htmlFor:"playground-field",className:"block text-sm font-medium text-gray-700 mb-1",children:"Form Field"}),e.jsx("input",{id:"playground-field",type:"text",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"playground-error"}),e.jsx("div",{id:"playground-error",children:e.jsx(r,{...s})})]}),args:{message:"This is an error message",suggestions:["Try this suggestion","Or this one"],helperText:"Additional helper text",fieldLabel:"Form Field",hideIcon:!1,size:"md"}},x={render:()=>e.jsx("div",{className:"max-w-2xl space-y-8",children:e.jsxs("div",{children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-4",children:"Consent Form Validation"}),e.jsxs("div",{className:"space-y-6",children:[e.jsxs("div",{children:[e.jsx("label",{htmlFor:"consent-email",className:"block text-sm font-medium text-gray-700 mb-1",children:"Email Address *"}),e.jsx("input",{id:"consent-email",type:"email",placeholder:"user@example.com",value:"invalid.email@",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md","aria-invalid":"true","aria-describedby":"consent-email-error"}),e.jsx("div",{id:"consent-email-error",children:e.jsx(r,{message:"Please enter a valid email address",suggestions:["Email must contain @ and a domain (e.g., user@example.com)"],fieldLabel:"Email Address"})})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"org-name",className:"block text-sm font-medium text-gray-700 mb-1",children:"Organization Name *"}),e.jsx("input",{id:"org-name",type:"text",placeholder:"Your organization",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md","aria-invalid":"true","aria-describedby":"org-error"}),e.jsx("div",{id:"org-error",children:e.jsx(r,{message:"Organization name is required",fieldLabel:"Organization Name"})})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"retention",className:"block text-sm font-medium text-gray-700 mb-1",children:"Data Retention Period (days) *"}),e.jsx("input",{id:"retention",type:"number",placeholder:"90",value:"500",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md","aria-invalid":"true","aria-describedby":"retention-error"}),e.jsx("div",{id:"retention-error",children:e.jsx(r,{message:"Data retention period exceeds maximum allowed",suggestions:["Maximum retention period is 365 days","For longer retention, contact compliance@example.com"],helperText:"Must comply with GDPR and local privacy regulations",fieldLabel:"Data Retention Period"})})]}),e.jsxs("div",{children:[e.jsxs("div",{className:"flex items-start",children:[e.jsx("input",{id:"terms",type:"checkbox",className:"mt-1 h-4 w-4 text-trust-deep border-2 border-error-primary rounded focus:ring-error-primary","aria-invalid":"true","aria-describedby":"terms-error"}),e.jsx("label",{htmlFor:"terms",className:"ml-2 text-sm text-gray-700",children:"I agree to the Terms of Service and Privacy Policy"})]}),e.jsx("div",{id:"terms-error",className:"ml-6",children:e.jsx(r,{message:"You must accept the terms and conditions to continue",hideIcon:!0,size:"sm",fieldLabel:"Terms Acceptance"})})]})]})]})}),args:{message:"Real-world validation example"}},g={render:()=>e.jsxs("div",{className:"space-y-6 max-w-3xl",children:[e.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Accessibility Features"}),e.jsxs("ul",{className:"text-sm text-gray-700 space-y-1",children:[e.jsxs("li",{children:["• Uses ",e.jsx("code",{children:'role="alert"'})," for immediate screen reader announcement"]}),e.jsxs("li",{children:["• Uses ",e.jsx("code",{children:'aria-live="polite"'})," to avoid interrupting users"]}),e.jsxs("li",{children:["• Uses ",e.jsx("code",{children:'aria-atomic="true"'})," for complete message reading"]}),e.jsxs("li",{children:["• Icons marked with ",e.jsx("code",{children:'aria-hidden="true"'})]}),e.jsxs("li",{children:["• Field association via ",e.jsx("code",{children:"aria-describedby"})," and ",e.jsx("code",{children:"aria-invalid"})]}),e.jsxs("li",{children:["• Optional ",e.jsx("code",{children:"fieldLabel"})," for screen reader context"]}),e.jsx("li",{children:"• Error text color meets WCAG 2.1 AA contrast requirements"}),e.jsx("li",{children:"• Keyboard accessible when containing links/buttons"})]})]}),e.jsxs("div",{children:[e.jsx("label",{htmlFor:"accessible-field",className:"block text-sm font-medium text-gray-700 mb-1",children:"Accessible Form Field"}),e.jsx("input",{id:"accessible-field",type:"text",className:"w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary","aria-invalid":"true","aria-describedby":"accessible-error"}),e.jsx("div",{id:"accessible-error",children:e.jsx(r,{message:"This error will be announced to screen readers",fieldLabel:"Accessible Form Field",helperText:"Try navigating with a screen reader to hear the announcement"})})]})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0},{id:"aria-valid-attr",enabled:!0}]}}},args:{message:"Accessible error message"}};var T,A,L,z,S;i.parameters={...i.parameters,docs:{...(T=i.parameters)==null?void 0:T.docs,source:{originalSource:`{
  args: {
    message: 'This field is required'
  }
}`,...(L=(A=i.parameters)==null?void 0:A.docs)==null?void 0:L.source},description:{story:"Default inline error with basic message",...(S=(z=i.parameters)==null?void 0:z.docs)==null?void 0:S.description}}};var C,M,U,R,D;o.parameters={...o.parameters,docs:{...(C=o.parameters)==null?void 0:C.docs,source:{originalSource:`{
  render: () => {
    const CustomIcon = () => <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>;
    return <div className="space-y-4 max-w-md">
        <InlineError message="Default error icon (X in circle)" />
        <InlineError message="Custom icon (info circle)" icon={<CustomIcon />} />
      </div>;
  },
  args: {
    message: 'Error with icon'
  }
}`,...(U=(M=o.parameters)==null?void 0:M.docs)==null?void 0:U.source},description:{story:"Error with custom icon",...(D=(R=o.parameters)==null?void 0:R.docs)==null?void 0:D.description}}};var q,O,W,B,V;d.parameters={...d.parameters,docs:{...(q=d.parameters)==null?void 0:q.docs,source:{originalSource:`{
  render: () => <div className="max-w-md space-y-6">
      {/* Valid field */}
      <div>
        <label htmlFor="email-valid" className="block text-sm font-medium text-gray-700 mb-1">
          Email Address (Valid)
        </label>
        <input id="email-valid" type="email" placeholder="user@example.com" value="user@example.com" className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-trust-deep focus:border-transparent" readOnly />
      </div>

      {/* Invalid field with error */}
      <div>
        <label htmlFor="email-invalid" className="block text-sm font-medium text-gray-700 mb-1">
          Email Address (Invalid)
        </label>
        <input id="email-invalid" type="email" placeholder="user@example.com" value="invalid-email" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="email-error" />
        <div id="email-error">
          <InlineError message="Please enter a valid email address" fieldLabel="Email Address" />
        </div>
      </div>

      {/* Required field with error */}
      <div>
        <label htmlFor="password-empty" className="block text-sm font-medium text-gray-700 mb-1">
          Password (Required)
        </label>
        <input id="password-empty" type="password" placeholder="Enter password" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="password-error" />
        <div id="password-error">
          <InlineError message="Password is required" fieldLabel="Password" />
        </div>
      </div>
    </div>,
  args: {
    message: 'Field-level error'
  }
}`,...(W=(O=d.parameters)==null?void 0:O.docs)==null?void 0:W.source},description:{story:"Error positioned below a form field",...(V=(B=d.parameters)==null?void 0:B.docs)==null?void 0:V.description}}};var G,Z,Y,H,_;l.parameters={...l.parameters,docs:{...(G=l.parameters)==null?void 0:G.docs,source:{originalSource:`{
  render: () => <div className="max-w-md space-y-6">
      <div>
        <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-1">
          Username
        </label>
        <input id="username" type="text" placeholder="Enter username" value="ab" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="username-errors" />
        <div id="username-errors">
          <InlineError errors={['Username must be at least 3 characters', 'Username can only contain letters and numbers', 'Username is already taken']} message="Multiple validation errors" fieldLabel="Username" />
        </div>
      </div>

      <div>
        <label htmlFor="phone" className="block text-sm font-medium text-gray-700 mb-1">
          Phone Number
        </label>
        <input id="phone" type="tel" placeholder="(555) 123-4567" value="123" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="phone-errors" />
        <div id="phone-errors">
          <InlineError errors={['Phone number must be 10 digits', 'Phone number format is invalid']} message="Phone validation errors" fieldLabel="Phone Number" />
        </div>
      </div>
    </div>,
  args: {
    errors: ['Username must be at least 3 characters', 'Username can only contain letters and numbers'],
    message: 'Multiple errors'
  }
}`,...(Y=(Z=l.parameters)==null?void 0:Z.docs)==null?void 0:Y.source},description:{story:"Multiple validation errors for a single field",...(_=(H=l.parameters)==null?void 0:H.docs)==null?void 0:_.description}}};var $,J,K,X,Q;t.parameters={...t.parameters,docs:{...($=t.parameters)==null?void 0:$.docs,source:{originalSource:`{
  render: () => <div className="max-w-md space-y-6">
      <div>
        <label htmlFor="password-weak" className="block text-sm font-medium text-gray-700 mb-1">
          Password
        </label>
        <input id="password-weak" type="password" placeholder="Enter password" value="abc123" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="password-suggestions" />
        <div id="password-suggestions">
          <InlineError message="Password is too weak" suggestions={['Use at least 8 characters', 'Include uppercase and lowercase letters', 'Add numbers and special symbols', 'Avoid common words or patterns']} fieldLabel="Password" />
        </div>
      </div>

      <div>
        <label htmlFor="upload" className="block text-sm font-medium text-gray-700 mb-1">
          Profile Photo
        </label>
        <input id="upload" type="file" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="upload-suggestions" />
        <div id="upload-suggestions">
          <InlineError message="File size is too large (5.2 MB)" suggestions={['Maximum file size is 2 MB', 'Try compressing your image', 'Supported formats: JPG, PNG, GIF']} fieldLabel="Profile Photo" />
        </div>
      </div>
    </div>,
  args: {
    message: 'Password is too weak',
    suggestions: ['Use at least 8 characters', 'Include uppercase and lowercase letters', 'Add numbers and special symbols']
  }
}`,...(K=(J=t.parameters)==null?void 0:J.docs)==null?void 0:K.source},description:{story:"Error with helpful suggestions",...(Q=(X=t.parameters)==null?void 0:X.docs)==null?void 0:Q.description}}};var ee,re,se,ae,ie;n.parameters={...n.parameters,docs:{...(ee=n.parameters)==null?void 0:ee.docs,source:{originalSource:`{
  render: () => <div className="max-w-md space-y-6">
      <div>
        <label htmlFor="email-taken" className="block text-sm font-medium text-gray-700 mb-1">
          Email Address
        </label>
        <input id="email-taken" type="email" placeholder="user@example.com" value="existing@example.com" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="email-link-error" />
        <div id="email-link-error">
          <InlineError message={<>
                This email is already registered.{' '}
                <a href="/login" className="underline hover:text-error-primary font-medium">
                  Sign in instead?
                </a>
              </>} fieldLabel="Email Address" />
        </div>
      </div>

      <div>
        <label htmlFor="code-invalid" className="block text-sm font-medium text-gray-700 mb-1">
          Verification Code
        </label>
        <input id="code-invalid" type="text" placeholder="Enter 6-digit code" value="123456" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="code-link-error" />
        <div id="code-link-error">
          <InlineError message={<>
                Invalid verification code.{' '}
                <button type="button" onClick={() => alert('Resending code...')} className="underline hover:text-error-primary font-medium">
                  Resend code
                </button>
              </>} fieldLabel="Verification Code" />
        </div>
      </div>
    </div>,
  args: {
    message: 'Error with link'
  }
}`,...(se=(re=n.parameters)==null?void 0:re.docs)==null?void 0:se.source},description:{story:"Error message with actionable links",...(ie=(ae=n.parameters)==null?void 0:ae.docs)==null?void 0:ie.description}}};var oe,de,le,te,ne;m.parameters={...m.parameters,docs:{...(oe=m.parameters)==null?void 0:oe.docs,source:{originalSource:`{
  render: () => <div className="max-w-md space-y-6">
      <div>
        <label htmlFor="billing-zip" className="block text-sm font-medium text-gray-700 mb-1">
          Billing ZIP Code
        </label>
        <input id="billing-zip" type="text" placeholder="12345" value="ABC" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="zip-error" />
        <div id="zip-error">
          <InlineError message="ZIP code must be 5 digits" fieldLabel="Billing ZIP Code" helperText="Example: 12345" />
        </div>
      </div>

      <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
        <p className="text-sm text-gray-700">
          The <code>fieldLabel</code> prop adds screen reader context (hidden visually).
          It helps screen reader users understand which field has an error.
        </p>
      </div>
    </div>,
  args: {
    message: 'ZIP code must be 5 digits',
    fieldLabel: 'Billing ZIP Code'
  }
}`,...(le=(de=m.parameters)==null?void 0:de.docs)==null?void 0:le.source},description:{story:"Error with field label context",...(ne=(te=m.parameters)==null?void 0:te.docs)==null?void 0:ne.description}}};var me,ce,ue,pe,be;c.parameters={...c.parameters,docs:{...(me=c.parameters)==null?void 0:me.docs,source:{originalSource:`{
  render: () => <div className="max-w-md space-y-6">
      <div>
        <label htmlFor="amount" className="block text-sm font-medium text-gray-700 mb-1">
          Transfer Amount
        </label>
        <input id="amount" type="number" placeholder="0.00" value="25000" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="amount-error" />
        <div id="amount-error">
          <InlineError message="Transfer amount exceeds daily limit" helperText="Your daily transfer limit is $10,000. Contact support to increase your limit." fieldLabel="Transfer Amount" />
        </div>
      </div>

      <div>
        <label htmlFor="domain" className="block text-sm font-medium text-gray-700 mb-1">
          Custom Domain
        </label>
        <input id="domain" type="text" placeholder="example.com" value="invalid..domain" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="domain-error" />
        <div id="domain-error">
          <InlineError message="Invalid domain format" helperText="Domain must contain only letters, numbers, hyphens, and periods." fieldLabel="Custom Domain" />
        </div>
      </div>
    </div>,
  args: {
    message: 'Transfer amount exceeds daily limit',
    helperText: 'Your daily transfer limit is $10,000. Contact support to increase your limit.'
  }
}`,...(ue=(ce=c.parameters)==null?void 0:ce.docs)==null?void 0:ue.source},description:{story:"Error with helper text",...(be=(pe=c.parameters)==null?void 0:pe.docs)==null?void 0:be.description}}};var xe,ge,ye,he,fe;u.parameters={...u.parameters,docs:{...(xe=u.parameters)==null?void 0:xe.docs,source:{originalSource:`{
  render: () => <div className="max-w-md space-y-6">
      <div>
        <label htmlFor="field-sm" className="block text-xs font-medium text-gray-700 mb-1">
          Small Field
        </label>
        <input id="field-sm" type="text" className="w-full px-2 py-1 text-sm border-2 border-error-primary rounded-md" aria-invalid="true" aria-describedby="error-sm" />
        <div id="error-sm">
          <InlineError size="sm" message="Small size error message" suggestions={['Suggestion in small size']} />
        </div>
      </div>

      <div>
        <label htmlFor="field-md" className="block text-sm font-medium text-gray-700 mb-1">
          Medium Field (Default)
        </label>
        <input id="field-md" type="text" className="w-full px-3 py-2 border-2 border-error-primary rounded-md" aria-invalid="true" aria-describedby="error-md" />
        <div id="error-md">
          <InlineError size="md" message="Medium size error message" suggestions={['Suggestion in medium size']} />
        </div>
      </div>
    </div>,
  args: {
    message: 'Size variants',
    size: 'md'
  }
}`,...(ye=(ge=u.parameters)==null?void 0:ge.docs)==null?void 0:ye.source},description:{story:"Different size variants",...(fe=(he=u.parameters)==null?void 0:he.docs)==null?void 0:fe.description}}};var ve,je,we,Ne,ke;p.parameters={...p.parameters,docs:{...(ve=p.parameters)==null?void 0:ve.docs,source:{originalSource:`{
  render: () => <div className="max-w-md space-y-6">
      <div>
        <label htmlFor="simple" className="block text-sm font-medium text-gray-700 mb-1">
          Simple Field
        </label>
        <input id="simple" type="text" className="w-full px-3 py-2 border-2 border-error-primary rounded-md" aria-invalid="true" aria-describedby="simple-error" />
        <div id="simple-error">
          <InlineError message="This field is required" hideIcon />
        </div>
      </div>

      <div>
        <label htmlFor="minimal" className="block text-sm font-medium text-gray-700 mb-1">
          Minimal Error
        </label>
        <input id="minimal" type="text" className="w-full px-3 py-2 border-2 border-error-primary rounded-md" aria-invalid="true" aria-describedby="minimal-error" />
        <div id="minimal-error">
          <InlineError message="Please enter a valid value" hideIcon helperText="No icon for cleaner appearance" />
        </div>
      </div>
    </div>,
  args: {
    message: 'Error without icon',
    hideIcon: true
  }
}`,...(we=(je=p.parameters)==null?void 0:je.docs)==null?void 0:we.source},description:{story:"Error without icon",...(ke=(Ne=p.parameters)==null?void 0:Ne.docs)==null?void 0:ke.description}}};var Ee,Fe,Ie,Pe,Te;b.parameters={...b.parameters,docs:{...(Ee=b.parameters)==null?void 0:Ee.docs,source:{originalSource:`{
  render: args => <div className="max-w-md">
      <label htmlFor="playground-field" className="block text-sm font-medium text-gray-700 mb-1">
        Form Field
      </label>
      <input id="playground-field" type="text" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="playground-error" />
      <div id="playground-error">
        <InlineError {...args} />
      </div>
    </div>,
  args: {
    message: 'This is an error message',
    suggestions: ['Try this suggestion', 'Or this one'],
    helperText: 'Additional helper text',
    fieldLabel: 'Form Field',
    hideIcon: false,
    size: 'md'
  }
}`,...(Ie=(Fe=b.parameters)==null?void 0:Fe.docs)==null?void 0:Ie.source},description:{story:"Interactive playground with all controls",...(Te=(Pe=b.parameters)==null?void 0:Pe.docs)==null?void 0:Te.description}}};var Ae,Le,ze,Se,Ce;x.parameters={...x.parameters,docs:{...(Ae=x.parameters)==null?void 0:Ae.docs,source:{originalSource:`{
  render: () => <div className="max-w-2xl space-y-8">
      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-4">
          Consent Form Validation
        </h3>

        <div className="space-y-6">
          {/* Email validation */}
          <div>
            <label htmlFor="consent-email" className="block text-sm font-medium text-gray-700 mb-1">
              Email Address *
            </label>
            <input id="consent-email" type="email" placeholder="user@example.com" value="invalid.email@" className="w-full px-3 py-2 border-2 border-error-primary rounded-md" aria-invalid="true" aria-describedby="consent-email-error" />
            <div id="consent-email-error">
              <InlineError message="Please enter a valid email address" suggestions={['Email must contain @ and a domain (e.g., user@example.com)']} fieldLabel="Email Address" />
            </div>
          </div>

          {/* Organization name */}
          <div>
            <label htmlFor="org-name" className="block text-sm font-medium text-gray-700 mb-1">
              Organization Name *
            </label>
            <input id="org-name" type="text" placeholder="Your organization" className="w-full px-3 py-2 border-2 border-error-primary rounded-md" aria-invalid="true" aria-describedby="org-error" />
            <div id="org-error">
              <InlineError message="Organization name is required" fieldLabel="Organization Name" />
            </div>
          </div>

          {/* Data retention period */}
          <div>
            <label htmlFor="retention" className="block text-sm font-medium text-gray-700 mb-1">
              Data Retention Period (days) *
            </label>
            <input id="retention" type="number" placeholder="90" value="500" className="w-full px-3 py-2 border-2 border-error-primary rounded-md" aria-invalid="true" aria-describedby="retention-error" />
            <div id="retention-error">
              <InlineError message="Data retention period exceeds maximum allowed" suggestions={['Maximum retention period is 365 days', 'For longer retention, contact compliance@example.com']} helperText="Must comply with GDPR and local privacy regulations" fieldLabel="Data Retention Period" />
            </div>
          </div>

          {/* Terms acceptance */}
          <div>
            <div className="flex items-start">
              <input id="terms" type="checkbox" className="mt-1 h-4 w-4 text-trust-deep border-2 border-error-primary rounded focus:ring-error-primary" aria-invalid="true" aria-describedby="terms-error" />
              <label htmlFor="terms" className="ml-2 text-sm text-gray-700">
                I agree to the Terms of Service and Privacy Policy
              </label>
            </div>
            <div id="terms-error" className="ml-6">
              <InlineError message="You must accept the terms and conditions to continue" hideIcon size="sm" fieldLabel="Terms Acceptance" />
            </div>
          </div>
        </div>
      </div>
    </div>,
  args: {
    message: 'Real-world validation example'
  }
}`,...(ze=(Le=x.parameters)==null?void 0:Le.docs)==null?void 0:ze.source},description:{story:"Real-world consent form validation examples",...(Ce=(Se=x.parameters)==null?void 0:Se.docs)==null?void 0:Ce.description}}};var Me,Ue,Re,De,qe;g.parameters={...g.parameters,docs:{...(Me=g.parameters)==null?void 0:Me.docs,source:{originalSource:`{
  render: () => <div className="space-y-6 max-w-3xl">
      <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
        <h4 className="text-sm font-semibold text-gray-900 mb-2">
          Accessibility Features
        </h4>
        <ul className="text-sm text-gray-700 space-y-1">
          <li>• Uses <code>role="alert"</code> for immediate screen reader announcement</li>
          <li>• Uses <code>aria-live="polite"</code> to avoid interrupting users</li>
          <li>• Uses <code>aria-atomic="true"</code> for complete message reading</li>
          <li>• Icons marked with <code>aria-hidden="true"</code></li>
          <li>• Field association via <code>aria-describedby</code> and <code>aria-invalid</code></li>
          <li>• Optional <code>fieldLabel</code> for screen reader context</li>
          <li>• Error text color meets WCAG 2.1 AA contrast requirements</li>
          <li>• Keyboard accessible when containing links/buttons</li>
        </ul>
      </div>

      <div>
        <label htmlFor="accessible-field" className="block text-sm font-medium text-gray-700 mb-1">
          Accessible Form Field
        </label>
        <input id="accessible-field" type="text" className="w-full px-3 py-2 border-2 border-error-primary rounded-md focus:outline-none focus:ring-2 focus:ring-error-primary" aria-invalid="true" aria-describedby="accessible-error" />
        <div id="accessible-error">
          <InlineError message="This error will be announced to screen readers" fieldLabel="Accessible Form Field" helperText="Try navigating with a screen reader to hear the announcement" />
        </div>
      </div>
    </div>,
  parameters: {
    a11y: {
      config: {
        rules: [{
          id: 'color-contrast',
          enabled: true
        }, {
          id: 'aria-valid-attr',
          enabled: true
        }]
      }
    }
  },
  args: {
    message: 'Accessible error message'
  }
}`,...(Re=(Ue=g.parameters)==null?void 0:Ue.docs)==null?void 0:Re.source},description:{story:"Accessibility features demonstration",...(qe=(De=g.parameters)==null?void 0:De.docs)==null?void 0:qe.description}}};const Qe=["Default","WithIcon","FieldLevel","MultipleErrors","WithSuggestions","WithLinks","WithFieldLabel","WithHelperText","Sizes","WithoutIcon","Playground","RealWorldExamples","Accessibility"];export{g as Accessibility,i as Default,d as FieldLevel,l as MultipleErrors,b as Playground,x as RealWorldExamples,u as Sizes,m as WithFieldLabel,c as WithHelperText,o as WithIcon,n as WithLinks,t as WithSuggestions,p as WithoutIcon,Qe as __namedExportsOrder,Xe as default};
