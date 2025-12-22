var zr=Object.defineProperty;var Fr=(s,o,e)=>o in s?zr(s,o,{enumerable:!0,configurable:!0,writable:!0,value:e}):s[o]=e;var B=(s,o,e)=>Fr(s,typeof o!="symbol"?o+"":o,e);import{j as r}from"./jsx-runtime-BYYWji4R.js";import{r as l}from"./index-ClcD9ViR.js";import{c as k}from"./cn-JCLedEej.js";import{B as d}from"./Button-DlbfybpJ.js";import"./_commonjsHelpers-Cpj98o6Y.js";const Lr=()=>r.jsx("svg",{className:"w-6 h-6",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:r.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"})}),Wr=()=>r.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:r.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"})});class i extends l.Component{constructor(e){super(e);B(this,"resetErrorBoundary",()=>{this.setState({hasError:!1,error:null,errorInfo:null})});this.state={hasError:!1,error:null,errorInfo:null,errorCount:0}}static getDerivedStateFromError(e){return{hasError:!0,error:e}}componentDidCatch(e,n){const{onError:t,boundaryName:a}=this.props,{errorCount:m}=this.state;if(this.setState({errorInfo:n,errorCount:m+1}),t)try{t(e,n)}catch(E){console.error("Error in onError handler:",E)}}componentDidUpdate(e){const{hasError:n}=this.state,{resetOnPropsChange:t,children:a}=this.props;n&&t&&a!==e.children&&this.resetErrorBoundary()}renderDefaultFallback(){const{error:e,errorInfo:n,errorCount:t}=this.state,{showDetails:a,boundaryName:m}=this.props;return e?r.jsxs("div",{role:"alert","aria-live":"assertive",className:k("bg-error-light border border-error-primary/30 rounded-lg p-6","text-error-dark"),children:[r.jsxs("div",{className:"flex items-start gap-3 mb-4",children:[r.jsx("div",{className:"flex-shrink-0 text-error-primary mt-0.5",children:r.jsx(Lr,{})}),r.jsxs("div",{className:"flex-1 min-w-0",children:[r.jsx("h3",{className:"text-lg font-semibold text-error-dark mb-1",children:"Something went wrong"}),r.jsx("p",{className:"text-sm text-error-dark/80",children:m?`An error occurred in ${m}.`:"An unexpected error has occurred."})]})]}),a&&r.jsxs("div",{className:"mb-4 space-y-3",children:[r.jsxs("div",{className:"bg-error-dark/5 border border-error-primary/20 rounded-md p-4",children:[r.jsx("h4",{className:"text-xs font-semibold text-error-dark uppercase tracking-wide mb-2",children:"Error Message"}),r.jsx("p",{className:"text-sm font-mono text-error-dark break-words",children:e.toString()})]}),(n==null?void 0:n.componentStack)&&r.jsxs("details",{className:"bg-error-dark/5 border border-error-primary/20 rounded-md",children:[r.jsx("summary",{className:"cursor-pointer p-4 text-xs font-semibold text-error-dark uppercase tracking-wide hover:bg-error-dark/5",children:"Component Stack (Click to expand)"}),r.jsx("pre",{className:"p-4 pt-2 text-xs font-mono text-error-dark overflow-x-auto",children:n.componentStack})]}),e.stack&&r.jsxs("details",{className:"bg-error-dark/5 border border-error-primary/20 rounded-md",children:[r.jsx("summary",{className:"cursor-pointer p-4 text-xs font-semibold text-error-dark uppercase tracking-wide hover:bg-error-dark/5",children:"Error Stack (Click to expand)"}),r.jsx("pre",{className:"p-4 pt-2 text-xs font-mono text-error-dark overflow-x-auto",children:e.stack})]}),t>1&&r.jsxs("div",{className:"text-xs text-error-dark/70 bg-error-dark/5 rounded px-3 py-2",children:["This error has occurred ",r.jsx("strong",{children:t})," time",t===1?"":"s","."]})]}),!a&&r.jsx("div",{className:"mb-4 text-sm text-error-dark/80",children:r.jsx("p",{children:"We apologize for the inconvenience. The error has been logged and our team will investigate."})}),r.jsxs("div",{className:"flex flex-wrap gap-3",children:[r.jsxs("button",{type:"button",onClick:this.resetErrorBoundary,className:k("inline-flex items-center gap-2 px-4 py-2 rounded-md","bg-error-primary text-white font-medium text-sm","hover:bg-error-dark transition-colors duration-200","focus:outline-none focus:ring-2 focus:ring-error-primary focus:ring-offset-2","shadow-sm hover:shadow-md"),children:[r.jsx(Wr,{}),"Try Again"]}),a&&r.jsx("button",{type:"button",onClick:()=>window.location.reload(),className:k("inline-flex items-center px-4 py-2 rounded-md","bg-white text-error-dark font-medium text-sm","border border-error-primary/30","hover:bg-error-light transition-colors duration-200","focus:outline-none focus:ring-2 focus:ring-error-primary focus:ring-offset-2"),children:"Reload Page"})]}),a&&r.jsx("div",{className:"mt-4 pt-4 border-t border-error-primary/20",children:r.jsxs("p",{className:"text-xs text-error-dark/60",children:[r.jsx("strong",{children:"Note:"})," Detailed error information is shown because"," ",r.jsx("code",{className:"bg-error-dark/10 px-1 py-0.5 rounded",children:"showDetails"})," ","is enabled. Disable this in production environments."]})})]}):null}render(){const{hasError:e,error:n}=this.state,{children:t,fallback:a}=this.props;return e&&n?a?typeof a=="function"?a(n,this.resetErrorBoundary):a:this.renderDefaultFallback():t}}i.__docgenInfo={description:`ErrorBoundary component for catching React component errors.
Must be a class component as per React requirements.

@example
\`\`\`tsx
<ErrorBoundary>
  <MyComponent />
</ErrorBoundary>

<ErrorBoundary
  showDetails={true}
  onError={(error, errorInfo) => logErrorToService(error, errorInfo)}
>
  <RiskyComponent />
</ErrorBoundary>

<ErrorBoundary
  fallback={(error, retry) => (
    <CustomErrorUI error={error} onRetry={retry} />
  )}
>
  <FeatureComponent />
</ErrorBoundary>
\`\`\``,methods:[{name:"resetErrorBoundary",docblock:null,modifiers:[],params:[],returns:{type:{name:"void"}}},{name:"renderDefaultFallback",docblock:null,modifiers:[],params:[],returns:{type:{name:"ReactNode"}}}],displayName:"ErrorBoundary",props:{children:{required:!0,tsType:{name:"ReactNode"},description:"Child components to protect with error boundary"},fallback:{required:!1,tsType:{name:"union",raw:"ReactNode | ((error: Error, retry: () => void) => ReactNode)",elements:[{name:"ReactNode"},{name:"unknown"}]},description:"Custom fallback UI - can be static JSX or function receiving (error, retry)"},onError:{required:!1,tsType:{name:"signature",type:"function",raw:"(error: Error, errorInfo: ErrorInfo) => void",signature:{arguments:[{type:{name:"Error"},name:"error"},{type:{name:"ErrorInfo"},name:"errorInfo"}],return:{name:"void"}}},description:"Callback when error is caught"},resetOnPropsChange:{required:!1,tsType:{name:"boolean"},description:"Reset error state when props change (useful for route changes)"},showDetails:{required:!1,tsType:{name:"boolean"},description:"Show detailed error information (dev mode)"},boundaryName:{required:!1,tsType:{name:"string"},description:"Custom error boundary name for identification"}}};const $r={title:"Design System/Feedback/ErrorBoundary",component:i,parameters:{layout:"padded"},tags:["autodocs"],argTypes:{showDetails:{control:"boolean",description:"Show detailed error information (dev mode)"},resetOnPropsChange:{control:"boolean",description:"Reset error state when props change"},boundaryName:{control:"text",description:"Custom name for error boundary identification"}}},c=({shouldThrow:s})=>{if(s)throw new Error("Intentional error from BuggyComponent for testing");return r.jsxs("div",{className:"p-6 bg-green-50 border border-green-200 rounded-lg",children:[r.jsx("h3",{className:"text-lg font-semibold text-green-900 mb-2",children:"Component Working Correctly"}),r.jsx("p",{className:"text-green-700",children:"This component will throw an error when you click the button below."})]})},C=()=>{throw new Error("This component always throws an error during render")},u={render:()=>{const[s,o]=l.useState(!1);return r.jsxs("div",{className:"space-y-4 max-w-2xl",children:[r.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[r.jsx("p",{className:"text-sm text-gray-700 mb-3",children:"Click the button below to trigger an error and see how the ErrorBoundary catches it."}),r.jsx(d,{variant:"danger",onClick:()=>o(!0),size:"sm",children:"Trigger Error"})]}),r.jsx(i,{children:r.jsx(c,{shouldThrow:s})})]})},args:{children:r.jsx("div",{children:"Default content"})}},h={render:()=>{const[s,o]=l.useState(!1);return r.jsxs("div",{className:"space-y-4 max-w-2xl",children:[r.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[r.jsx("p",{className:"text-sm text-gray-700 mb-3",children:"This error boundary has a custom name that appears in the error message."}),r.jsx(d,{variant:"danger",onClick:()=>o(!0),size:"sm",children:"Trigger Error"})]}),r.jsx(i,{boundaryName:"User Profile Section",children:r.jsx(c,{shouldThrow:s})})]})},args:{children:r.jsx("div",{children:"Content"})}},g={render:()=>{const[s,o]=l.useState(!1),[e,n]=l.useState(0),t=()=>{o(!1),n(e+1)};return r.jsxs("div",{className:"space-y-4 max-w-2xl",children:[r.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[r.jsx("p",{className:"text-sm text-gray-700 mb-3",children:'The error boundary provides "Try Again" and "Reload Page" buttons to recover from errors.'}),r.jsxs("div",{className:"flex gap-2",children:[r.jsx(d,{variant:"danger",onClick:()=>o(!0),size:"sm",children:"Trigger Error"}),r.jsx(d,{variant:"outline",onClick:t,size:"sm",children:"Reset Component"})]})]}),r.jsx(i,{showDetails:!0,children:r.jsx(c,{shouldThrow:s})},e)]})},args:{children:r.jsx("div",{children:"Content"})}},p={render:()=>{const[s,o]=l.useState(!1),[e,n]=l.useState(!1);return r.jsxs("div",{className:"space-y-4 max-w-3xl",children:[r.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[r.jsx("p",{className:"text-sm text-gray-700 mb-3",children:"Nested error boundaries allow you to isolate errors to specific parts of your component tree. The inner boundary catches inner errors, while the outer boundary only catches if the inner one fails."}),r.jsxs("div",{className:"flex gap-2",children:[r.jsx(d,{variant:"danger",onClick:()=>o(!0),size:"sm",children:"Trigger Outer Error"}),r.jsx(d,{variant:"danger",onClick:()=>n(!0),size:"sm",children:"Trigger Inner Error"})]})]}),r.jsx(i,{boundaryName:"Outer Container",children:r.jsxs("div",{className:"border-2 border-blue-300 rounded-lg p-4 bg-blue-50",children:[r.jsx("h3",{className:"text-sm font-semibold text-blue-900 mb-3",children:"Outer Error Boundary"}),r.jsx(c,{shouldThrow:s}),r.jsx("div",{className:"mt-4",children:r.jsx(i,{boundaryName:"Inner Widget",children:r.jsxs("div",{className:"border-2 border-purple-300 rounded-lg p-4 bg-purple-50",children:[r.jsx("h4",{className:"text-sm font-semibold text-purple-900 mb-2",children:"Inner Error Boundary"}),r.jsx(c,{shouldThrow:e})]})})})]})})]})},args:{children:r.jsx("div",{children:"Content"})}},x={render:()=>r.jsxs("div",{className:"space-y-4 max-w-3xl",children:[r.jsxs("div",{className:"p-4 bg-blue-50 border border-blue-200 rounded-lg",children:[r.jsx("h4",{className:"text-sm font-semibold text-blue-900 mb-2",children:"Development Mode Features"}),r.jsxs("ul",{className:"text-sm text-blue-800 space-y-1 list-disc list-inside",children:[r.jsx("li",{children:"Shows detailed error message"}),r.jsx("li",{children:"Displays component stack trace"}),r.jsx("li",{children:"Shows full error stack"}),r.jsx("li",{children:'Includes "Reload Page" button'}),r.jsx("li",{children:"Shows error count"}),r.jsx("li",{children:"Logs to browser console"})]})]}),r.jsx(i,{showDetails:!0,boundaryName:"Development Component",children:r.jsx(C,{})})]}),args:{children:r.jsx("div",{children:"Content"})}},b={render:()=>r.jsxs("div",{className:"space-y-4 max-w-3xl",children:[r.jsxs("div",{className:"p-4 bg-amber-50 border border-amber-200 rounded-lg",children:[r.jsx("h4",{className:"text-sm font-semibold text-amber-900 mb-2",children:"Production Mode Features"}),r.jsxs("ul",{className:"text-sm text-amber-800 space-y-1 list-disc list-inside",children:[r.jsx("li",{children:"Shows user-friendly error message"}),r.jsx("li",{children:"Hides technical details"}),r.jsx("li",{children:'Provides "Try Again" button only'}),r.jsx("li",{children:"Logs still sent to error tracking service"})]})]}),r.jsx(i,{showDetails:!1,boundaryName:"Production Component",children:r.jsx(C,{})})]}),args:{children:r.jsx("div",{children:"Content"})}},y={render:()=>{const[s,o]=l.useState(!1),e=(n,t)=>r.jsxs("div",{className:"p-8 bg-gradient-to-br from-purple-50 to-pink-50 border-2 border-purple-200 rounded-xl text-center",children:[r.jsx("div",{className:"inline-flex items-center justify-center w-16 h-16 bg-purple-100 rounded-full mb-4",children:r.jsx("svg",{className:"w-8 h-8 text-purple-600",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:r.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"})})}),r.jsx("h3",{className:"text-xl font-bold text-purple-900 mb-2",children:"Oops! Something broke"}),r.jsx("p",{className:"text-purple-700 mb-1 text-sm",children:"Don't worry, it happens to the best of us."}),r.jsx("p",{className:"text-purple-600 text-xs mb-6 font-mono",children:n.message}),r.jsx(d,{variant:"primary",onClick:t,size:"md",children:"Let's Try That Again"})]});return r.jsxs("div",{className:"space-y-4 max-w-2xl",children:[r.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[r.jsx("p",{className:"text-sm text-gray-700 mb-3",children:"You can provide a custom fallback UI as a function that receives the error and retry callback."}),r.jsx(d,{variant:"danger",onClick:()=>o(!0),size:"sm",children:"Trigger Error"})]}),r.jsx(i,{fallback:e,children:r.jsx(c,{shouldThrow:s})})]})},args:{children:r.jsx("div",{children:"Content"})}},v={render:s=>{const[o,e]=l.useState(!1),[n,t]=l.useState(0),a=()=>{e(!1),t(n+1)};return r.jsxs("div",{className:"space-y-4 max-w-2xl",children:[r.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[r.jsx("p",{className:"text-sm text-gray-700 mb-3",children:"Use the controls below to customize the error boundary behavior. Click the button to trigger an error and see it in action."}),r.jsxs("div",{className:"flex gap-2",children:[r.jsx(d,{variant:"danger",onClick:()=>e(!0),size:"sm",children:"Trigger Error"}),r.jsx(d,{variant:"outline",onClick:a,size:"sm",children:"Reset"})]})]}),r.jsx(i,{...s,children:r.jsx(c,{shouldThrow:o})},n)]})},args:{children:r.jsx("div",{children:"Playground content"}),showDetails:!0,resetOnPropsChange:!1,boundaryName:"Playground Component"}},f={args:{children:r.jsx("div",{children:"Content"})},render:()=>{const[s,o]=l.useState(!1),[e,n]=l.useState(!1);return r.jsxs("div",{className:"space-y-6 max-w-3xl",children:[r.jsxs("div",{children:[r.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-4",children:"Consent Management Error Scenarios"}),r.jsxs("div",{className:"space-y-4",children:[r.jsxs("div",{children:[r.jsx("h4",{className:"text-sm font-semibold text-gray-800 mb-2",children:"1. Consent Form Validation Error"}),r.jsx("div",{className:"mb-2",children:r.jsx(d,{variant:"danger",onClick:()=>o(!0),size:"sm",children:"Simulate Form Error"})}),r.jsx(i,{boundaryName:"Consent Form",showDetails:!1,onError:t=>{console.log("Logging consent form error:",t)},children:r.jsx("div",{className:"p-4 bg-white border border-gray-200 rounded-lg",children:r.jsx(c,{shouldThrow:s})})})]}),r.jsxs("div",{children:[r.jsx("h4",{className:"text-sm font-semibold text-gray-800 mb-2",children:"2. API Integration Error"}),r.jsx("div",{className:"mb-2",children:r.jsx(d,{variant:"danger",onClick:()=>n(!0),size:"sm",children:"Simulate API Error"})}),r.jsx(i,{boundaryName:"API Dashboard",showDetails:!1,fallback:(t,a)=>r.jsxs("div",{className:"p-6 bg-red-50 border border-red-200 rounded-lg text-center",children:[r.jsx("h4",{className:"text-lg font-semibold text-red-900 mb-2",children:"Unable to Load Dashboard"}),r.jsx("p",{className:"text-sm text-red-700 mb-4",children:"We're having trouble connecting to our servers. Please check your connection and try again."}),r.jsx(d,{variant:"primary",onClick:a,size:"sm",children:"Retry Connection"})]}),children:r.jsx("div",{className:"p-4 bg-white border border-gray-200 rounded-lg",children:r.jsx(c,{shouldThrow:e})})})]}),r.jsxs("div",{children:[r.jsx("h4",{className:"text-sm font-semibold text-gray-800 mb-2",children:"3. Feature Widget with Graceful Degradation"}),r.jsxs("div",{className:"grid grid-cols-2 gap-4",children:[r.jsx(i,{boundaryName:"Analytics Widget",fallback:r.jsx("div",{className:"p-4 bg-gray-100 border border-gray-300 rounded-lg text-center",children:r.jsx("p",{className:"text-sm text-gray-600",children:"Analytics temporarily unavailable"})}),children:r.jsxs("div",{className:"p-4 bg-white border border-gray-200 rounded-lg",children:[r.jsx("h5",{className:"font-semibold text-sm mb-2",children:"Analytics"}),r.jsx("p",{className:"text-sm text-gray-600",children:"Widget working correctly"})]})}),r.jsx(i,{boundaryName:"Reports Widget",fallback:r.jsx("div",{className:"p-4 bg-gray-100 border border-gray-300 rounded-lg text-center",children:r.jsx("p",{className:"text-sm text-gray-600",children:"Reports temporarily unavailable"})}),children:r.jsxs("div",{className:"p-4 bg-white border border-gray-200 rounded-lg",children:[r.jsx("h5",{className:"font-semibold text-sm mb-2",children:"Reports"}),r.jsx("p",{className:"text-sm text-gray-600",children:"Widget working correctly"})]})})]})]})]})]}),r.jsx("div",{className:"pt-4 border-t border-gray-200",children:r.jsx("button",{onClick:()=>{o(!1),n(!1)},className:"text-sm text-trust-deep hover:text-trust-hover underline",children:"Reset All Examples"})})]})}},N={args:{children:r.jsx("div",{children:"Content"})},render:()=>{const[s,o]=l.useState([]),[e,n]=l.useState(!1),t=(a,m)=>{const E=`[${new Date().toLocaleTimeString()}] ${a.message}`;o(Pr=>[...Pr,E])};return r.jsxs("div",{className:"space-y-4 max-w-3xl",children:[r.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[r.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Error Logging Integration"}),r.jsx("p",{className:"text-sm text-gray-700 mb-3",children:"The onError callback allows you to integrate with error tracking services like Sentry, LogRocket, or custom logging solutions."}),r.jsx(d,{variant:"danger",onClick:()=>n(!0),size:"sm",children:"Trigger Error"})]}),s.length>0&&r.jsxs("div",{className:"p-4 bg-gray-900 text-green-400 rounded-lg font-mono text-xs",children:[r.jsxs("div",{className:"flex justify-between items-center mb-2",children:[r.jsx("span",{className:"font-semibold",children:"Error Log:"}),r.jsx("button",{onClick:()=>o([]),className:"text-gray-400 hover:text-white text-xs underline",children:"Clear"})]}),s.map((a,m)=>r.jsx("div",{children:a},m))]}),r.jsx(i,{onError:t,showDetails:!1,children:r.jsx(c,{shouldThrow:e})})]})}},j={args:{children:r.jsx("div",{children:"Content"})},render:()=>{const[s,o]=l.useState(1),[e,n]=l.useState(!1);return r.jsxs("div",{className:"space-y-4 max-w-2xl",children:[r.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[r.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Auto-Reset on Route/Prop Changes"}),r.jsx("p",{className:"text-sm text-gray-700 mb-3",children:"When resetOnPropsChange is enabled, the error boundary automatically resets when props change (useful for route transitions)."}),r.jsxs("div",{className:"flex gap-2",children:[r.jsx(d,{variant:"danger",onClick:()=>n(!0),size:"sm",children:"Trigger Error"}),r.jsxs(d,{variant:"outline",onClick:()=>o(s+1),size:"sm",children:["Change User (ID: ",s,")"]})]})]}),r.jsx(i,{resetOnPropsChange:!0,showDetails:!1,children:r.jsxs("div",{className:"p-4 bg-white border border-gray-200 rounded-lg",children:[r.jsxs("h4",{className:"text-sm font-semibold mb-2",children:["User Profile: ",s]}),r.jsx(c,{shouldThrow:e})]})})]})}},w={args:{children:r.jsx("div",{children:"Content"})},render:()=>r.jsxs("div",{className:"space-y-6 max-w-3xl",children:[r.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[r.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Accessibility Features"}),r.jsxs("ul",{className:"text-sm text-gray-700 space-y-1 list-disc list-inside",children:[r.jsxs("li",{children:["Uses ",r.jsx("code",{children:'role="alert"'})," for immediate screen reader announcement"]}),r.jsxs("li",{children:[r.jsx("code",{children:'aria-live="assertive"'})," ensures errors are announced immediately"]}),r.jsx("li",{children:"Buttons are fully keyboard accessible with focus indicators"}),r.jsxs("li",{children:["Error icons are marked with ",r.jsx("code",{children:'aria-hidden="true"'})]}),r.jsx("li",{children:"Clear, semantic heading structure"}),r.jsx("li",{children:"Collapsible sections use proper details/summary elements"}),r.jsx("li",{children:"Color contrast meets WCAG 2.1 AA requirements"})]})]}),r.jsx(i,{showDetails:!0,boundaryName:"Accessible Error Example",children:r.jsx(C,{})})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0},{id:"button-name",enabled:!0},{id:"aria-roles",enabled:!0}]}}}};var T,S,D,I,R;u.parameters={...u.parameters,docs:{...(T=u.parameters)==null?void 0:T.docs,source:{originalSource:`{
  render: () => {
    const [shouldThrow, setShouldThrow] = useState(false);
    return <div className="space-y-4 max-w-2xl">
        <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
          <p className="text-sm text-gray-700 mb-3">
            Click the button below to trigger an error and see how the ErrorBoundary
            catches it.
          </p>
          <Button variant="danger" onClick={() => setShouldThrow(true)} size="sm">
            Trigger Error
          </Button>
        </div>

        <ErrorBoundary>
          <BuggyComponent shouldThrow={shouldThrow} />
        </ErrorBoundary>
      </div>;
  },
  args: {
    children: <div>Default content</div>
  }
}`,...(D=(S=u.parameters)==null?void 0:S.docs)==null?void 0:D.source},description:{story:"Default error boundary catching a component error",...(R=(I=u.parameters)==null?void 0:I.docs)==null?void 0:R.description}}};var A,P,z,F,L;h.parameters={...h.parameters,docs:{...(A=h.parameters)==null?void 0:A.docs,source:{originalSource:`{
  render: () => {
    const [shouldThrow, setShouldThrow] = useState(false);
    return <div className="space-y-4 max-w-2xl">
        <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
          <p className="text-sm text-gray-700 mb-3">
            This error boundary has a custom name that appears in the error message.
          </p>
          <Button variant="danger" onClick={() => setShouldThrow(true)} size="sm">
            Trigger Error
          </Button>
        </div>

        <ErrorBoundary boundaryName="User Profile Section">
          <BuggyComponent shouldThrow={shouldThrow} />
        </ErrorBoundary>
      </div>;
  },
  args: {
    children: <div>Content</div>
  }
}`,...(z=(P=h.parameters)==null?void 0:P.docs)==null?void 0:z.source},description:{story:"Error boundary with custom error message",...(L=(F=h.parameters)==null?void 0:F.docs)==null?void 0:L.description}}};var W,O,U,M,q;g.parameters={...g.parameters,docs:{...(W=g.parameters)==null?void 0:W.docs,source:{originalSource:`{
  render: () => {
    const [shouldThrow, setShouldThrow] = useState(false);
    const [key, setKey] = useState(0);
    const handleReset = () => {
      setShouldThrow(false);
      setKey(key + 1); // Force re-mount
    };
    return <div className="space-y-4 max-w-2xl">
        <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
          <p className="text-sm text-gray-700 mb-3">
            The error boundary provides "Try Again" and "Reload Page" buttons to
            recover from errors.
          </p>
          <div className="flex gap-2">
            <Button variant="danger" onClick={() => setShouldThrow(true)} size="sm">
              Trigger Error
            </Button>
            <Button variant="outline" onClick={handleReset} size="sm">
              Reset Component
            </Button>
          </div>
        </div>

        <ErrorBoundary key={key} showDetails>
          <BuggyComponent shouldThrow={shouldThrow} />
        </ErrorBoundary>
      </div>;
  },
  args: {
    children: <div>Content</div>
  }
}`,...(U=(O=g.parameters)==null?void 0:O.docs)==null?void 0:U.source},description:{story:"Error boundary with retry/reset actions",...(q=(M=g.parameters)==null?void 0:M.docs)==null?void 0:q.description}}};var K,G,$,H,_;p.parameters={...p.parameters,docs:{...(K=p.parameters)==null?void 0:K.docs,source:{originalSource:`{
  render: () => {
    const [outerError, setOuterError] = useState(false);
    const [innerError, setInnerError] = useState(false);
    return <div className="space-y-4 max-w-3xl">
        <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
          <p className="text-sm text-gray-700 mb-3">
            Nested error boundaries allow you to isolate errors to specific parts of
            your component tree. The inner boundary catches inner errors, while the
            outer boundary only catches if the inner one fails.
          </p>
          <div className="flex gap-2">
            <Button variant="danger" onClick={() => setOuterError(true)} size="sm">
              Trigger Outer Error
            </Button>
            <Button variant="danger" onClick={() => setInnerError(true)} size="sm">
              Trigger Inner Error
            </Button>
          </div>
        </div>

        {/* Outer boundary */}
        <ErrorBoundary boundaryName="Outer Container">
          <div className="border-2 border-blue-300 rounded-lg p-4 bg-blue-50">
            <h3 className="text-sm font-semibold text-blue-900 mb-3">
              Outer Error Boundary
            </h3>

            <BuggyComponent shouldThrow={outerError} />

            <div className="mt-4">
              {/* Inner boundary */}
              <ErrorBoundary boundaryName="Inner Widget">
                <div className="border-2 border-purple-300 rounded-lg p-4 bg-purple-50">
                  <h4 className="text-sm font-semibold text-purple-900 mb-2">
                    Inner Error Boundary
                  </h4>
                  <BuggyComponent shouldThrow={innerError} />
                </div>
              </ErrorBoundary>
            </div>
          </div>
        </ErrorBoundary>
      </div>;
  },
  args: {
    children: <div>Content</div>
  }
}`,...($=(G=p.parameters)==null?void 0:G.docs)==null?void 0:$.source},description:{story:"Multiple nested error boundaries",...(_=(H=p.parameters)==null?void 0:H.docs)==null?void 0:_.description}}};var V,Y,J,X,Q;x.parameters={...x.parameters,docs:{...(V=x.parameters)==null?void 0:V.docs,source:{originalSource:`{
  render: () => {
    return <div className="space-y-4 max-w-3xl">
        <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <h4 className="text-sm font-semibold text-blue-900 mb-2">
            Development Mode Features
          </h4>
          <ul className="text-sm text-blue-800 space-y-1 list-disc list-inside">
            <li>Shows detailed error message</li>
            <li>Displays component stack trace</li>
            <li>Shows full error stack</li>
            <li>Includes "Reload Page" button</li>
            <li>Shows error count</li>
            <li>Logs to browser console</li>
          </ul>
        </div>

        <ErrorBoundary showDetails boundaryName="Development Component">
          <BrokenComponent />
        </ErrorBoundary>
      </div>;
  },
  args: {
    children: <div>Content</div>
  }
}`,...(J=(Y=x.parameters)==null?void 0:Y.docs)==null?void 0:J.source},description:{story:"Development mode with detailed error information",...(Q=(X=x.parameters)==null?void 0:X.docs)==null?void 0:Q.description}}};var Z,rr,er,sr,or;b.parameters={...b.parameters,docs:{...(Z=b.parameters)==null?void 0:Z.docs,source:{originalSource:`{
  render: () => {
    return <div className="space-y-4 max-w-3xl">
        <div className="p-4 bg-amber-50 border border-amber-200 rounded-lg">
          <h4 className="text-sm font-semibold text-amber-900 mb-2">
            Production Mode Features
          </h4>
          <ul className="text-sm text-amber-800 space-y-1 list-disc list-inside">
            <li>Shows user-friendly error message</li>
            <li>Hides technical details</li>
            <li>Provides "Try Again" button only</li>
            <li>Logs still sent to error tracking service</li>
          </ul>
        </div>

        <ErrorBoundary showDetails={false} boundaryName="Production Component">
          <BrokenComponent />
        </ErrorBoundary>
      </div>;
  },
  args: {
    children: <div>Content</div>
  }
}`,...(er=(rr=b.parameters)==null?void 0:rr.docs)==null?void 0:er.source},description:{story:"Production mode with minimal error display",...(or=(sr=b.parameters)==null?void 0:sr.docs)==null?void 0:or.description}}};var nr,tr,ar,ir,dr;y.parameters={...y.parameters,docs:{...(nr=y.parameters)==null?void 0:nr.docs,source:{originalSource:`{
  render: () => {
    const [shouldThrow, setShouldThrow] = useState(false);
    const customFallback = (error: Error, retry: () => void) => <div className="p-8 bg-gradient-to-br from-purple-50 to-pink-50 border-2 border-purple-200 rounded-xl text-center">
        <div className="inline-flex items-center justify-center w-16 h-16 bg-purple-100 rounded-full mb-4">
          <svg className="w-8 h-8 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <h3 className="text-xl font-bold text-purple-900 mb-2">
          Oops! Something broke
        </h3>
        <p className="text-purple-700 mb-1 text-sm">
          Don't worry, it happens to the best of us.
        </p>
        <p className="text-purple-600 text-xs mb-6 font-mono">{error.message}</p>
        <Button variant="primary" onClick={retry} size="md">
          Let's Try That Again
        </Button>
      </div>;
    return <div className="space-y-4 max-w-2xl">
        <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
          <p className="text-sm text-gray-700 mb-3">
            You can provide a custom fallback UI as a function that receives the
            error and retry callback.
          </p>
          <Button variant="danger" onClick={() => setShouldThrow(true)} size="sm">
            Trigger Error
          </Button>
        </div>

        <ErrorBoundary fallback={customFallback}>
          <BuggyComponent shouldThrow={shouldThrow} />
        </ErrorBoundary>
      </div>;
  },
  args: {
    children: <div>Content</div>
  }
}`,...(ar=(tr=y.parameters)==null?void 0:tr.docs)==null?void 0:ar.source},description:{story:"Error boundary with custom fallback UI",...(dr=(ir=y.parameters)==null?void 0:ir.docs)==null?void 0:dr.description}}};var lr,cr,mr,ur,hr;v.parameters={...v.parameters,docs:{...(lr=v.parameters)==null?void 0:lr.docs,source:{originalSource:`{
  render: args => {
    const [shouldThrow, setShouldThrow] = useState(false);
    const [key, setKey] = useState(0);
    const handleReset = () => {
      setShouldThrow(false);
      setKey(key + 1);
    };
    return <div className="space-y-4 max-w-2xl">
        <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
          <p className="text-sm text-gray-700 mb-3">
            Use the controls below to customize the error boundary behavior. Click
            the button to trigger an error and see it in action.
          </p>
          <div className="flex gap-2">
            <Button variant="danger" onClick={() => setShouldThrow(true)} size="sm">
              Trigger Error
            </Button>
            <Button variant="outline" onClick={handleReset} size="sm">
              Reset
            </Button>
          </div>
        </div>

        <ErrorBoundary key={key} {...args}>
          <BuggyComponent shouldThrow={shouldThrow} />
        </ErrorBoundary>
      </div>;
  },
  args: {
    children: <div>Playground content</div>,
    showDetails: true,
    resetOnPropsChange: false,
    boundaryName: 'Playground Component'
  }
}`,...(mr=(cr=v.parameters)==null?void 0:cr.docs)==null?void 0:mr.source},description:{story:"Interactive playground with all controls",...(hr=(ur=v.parameters)==null?void 0:ur.docs)==null?void 0:hr.description}}};var gr,pr,xr,br,yr;f.parameters={...f.parameters,docs:{...(gr=f.parameters)==null?void 0:gr.docs,source:{originalSource:`{
  args: {
    children: <div>Content</div>
  },
  render: () => {
    const [consentFormError, setConsentFormError] = useState(false);
    const [apiError, setApiError] = useState(false);
    return <div className="space-y-6 max-w-3xl">
        <div>
          <h3 className="text-lg font-semibold text-gray-900 mb-4">
            Consent Management Error Scenarios
          </h3>

          <div className="space-y-4">
            {/* Consent Form Error */}
            <div>
              <h4 className="text-sm font-semibold text-gray-800 mb-2">
                1. Consent Form Validation Error
              </h4>
              <div className="mb-2">
                <Button variant="danger" onClick={() => setConsentFormError(true)} size="sm">
                  Simulate Form Error
                </Button>
              </div>
              <ErrorBoundary boundaryName="Consent Form" showDetails={false} onError={error => {
              console.log('Logging consent form error:', error);
            }}>
                <div className="p-4 bg-white border border-gray-200 rounded-lg">
                  <BuggyComponent shouldThrow={consentFormError} />
                </div>
              </ErrorBoundary>
            </div>

            {/* API Integration Error */}
            <div>
              <h4 className="text-sm font-semibold text-gray-800 mb-2">
                2. API Integration Error
              </h4>
              <div className="mb-2">
                <Button variant="danger" onClick={() => setApiError(true)} size="sm">
                  Simulate API Error
                </Button>
              </div>
              <ErrorBoundary boundaryName="API Dashboard" showDetails={false} fallback={(error, retry) => <div className="p-6 bg-red-50 border border-red-200 rounded-lg text-center">
                    <h4 className="text-lg font-semibold text-red-900 mb-2">
                      Unable to Load Dashboard
                    </h4>
                    <p className="text-sm text-red-700 mb-4">
                      We're having trouble connecting to our servers. Please check
                      your connection and try again.
                    </p>
                    <Button variant="primary" onClick={retry} size="sm">
                      Retry Connection
                    </Button>
                  </div>}>
                <div className="p-4 bg-white border border-gray-200 rounded-lg">
                  <BuggyComponent shouldThrow={apiError} />
                </div>
              </ErrorBoundary>
            </div>

            {/* Graceful Degradation */}
            <div>
              <h4 className="text-sm font-semibold text-gray-800 mb-2">
                3. Feature Widget with Graceful Degradation
              </h4>
              <div className="grid grid-cols-2 gap-4">
                <ErrorBoundary boundaryName="Analytics Widget" fallback={<div className="p-4 bg-gray-100 border border-gray-300 rounded-lg text-center">
                      <p className="text-sm text-gray-600">
                        Analytics temporarily unavailable
                      </p>
                    </div>}>
                  <div className="p-4 bg-white border border-gray-200 rounded-lg">
                    <h5 className="font-semibold text-sm mb-2">Analytics</h5>
                    <p className="text-sm text-gray-600">
                      Widget working correctly
                    </p>
                  </div>
                </ErrorBoundary>

                <ErrorBoundary boundaryName="Reports Widget" fallback={<div className="p-4 bg-gray-100 border border-gray-300 rounded-lg text-center">
                      <p className="text-sm text-gray-600">
                        Reports temporarily unavailable
                      </p>
                    </div>}>
                  <div className="p-4 bg-white border border-gray-200 rounded-lg">
                    <h5 className="font-semibold text-sm mb-2">Reports</h5>
                    <p className="text-sm text-gray-600">
                      Widget working correctly
                    </p>
                  </div>
                </ErrorBoundary>
              </div>
            </div>
          </div>
        </div>

        <div className="pt-4 border-t border-gray-200">
          <button onClick={() => {
          setConsentFormError(false);
          setApiError(false);
        }} className="text-sm text-trust-deep hover:text-trust-hover underline">
            Reset All Examples
          </button>
        </div>
      </div>;
  }
}`,...(xr=(pr=f.parameters)==null?void 0:pr.docs)==null?void 0:xr.source},description:{story:"Real-world consent management examples",...(yr=(br=f.parameters)==null?void 0:br.docs)==null?void 0:yr.description}}};var vr,fr,Nr,jr,wr;N.parameters={...N.parameters,docs:{...(vr=N.parameters)==null?void 0:vr.docs,source:{originalSource:`{
  args: {
    children: <div>Content</div>
  },
  render: () => {
    const [logs, setLogs] = useState<string[]>([]);
    const [shouldThrow, setShouldThrow] = useState(false);
    const handleError = (error: Error, errorInfo: React.ErrorInfo) => {
      const logEntry = \`[\${new Date().toLocaleTimeString()}] \${error.message}\`;
      setLogs(prev => [...prev, logEntry]);

      // In production, you would send this to your error tracking service
      // Example: Sentry.captureException(error, { contexts: { react: errorInfo } });
    };
    return <div className="space-y-4 max-w-3xl">
        <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
          <h4 className="text-sm font-semibold text-gray-900 mb-2">
            Error Logging Integration
          </h4>
          <p className="text-sm text-gray-700 mb-3">
            The onError callback allows you to integrate with error tracking services
            like Sentry, LogRocket, or custom logging solutions.
          </p>
          <Button variant="danger" onClick={() => setShouldThrow(true)} size="sm">
            Trigger Error
          </Button>
        </div>

        {logs.length > 0 && <div className="p-4 bg-gray-900 text-green-400 rounded-lg font-mono text-xs">
            <div className="flex justify-between items-center mb-2">
              <span className="font-semibold">Error Log:</span>
              <button onClick={() => setLogs([])} className="text-gray-400 hover:text-white text-xs underline">
                Clear
              </button>
            </div>
            {logs.map((log, index) => <div key={index}>{log}</div>)}
          </div>}

        <ErrorBoundary onError={handleError} showDetails={false}>
          <BuggyComponent shouldThrow={shouldThrow} />
        </ErrorBoundary>
      </div>;
  }
}`,...(Nr=(fr=N.parameters)==null?void 0:fr.docs)==null?void 0:Nr.source},description:{story:"Error logging integration example",...(wr=(jr=N.parameters)==null?void 0:jr.docs)==null?void 0:wr.description}}};var Er,kr,Cr,Br,Tr;j.parameters={...j.parameters,docs:{...(Er=j.parameters)==null?void 0:Er.docs,source:{originalSource:`{
  args: {
    children: <div>Content</div>
  },
  render: () => {
    const [userId, setUserId] = useState(1);
    const [shouldThrow, setShouldThrow] = useState(false);
    return <div className="space-y-4 max-w-2xl">
        <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
          <h4 className="text-sm font-semibold text-gray-900 mb-2">
            Auto-Reset on Route/Prop Changes
          </h4>
          <p className="text-sm text-gray-700 mb-3">
            When resetOnPropsChange is enabled, the error boundary automatically
            resets when props change (useful for route transitions).
          </p>
          <div className="flex gap-2">
            <Button variant="danger" onClick={() => setShouldThrow(true)} size="sm">
              Trigger Error
            </Button>
            <Button variant="outline" onClick={() => setUserId(userId + 1)} size="sm">
              Change User (ID: {userId})
            </Button>
          </div>
        </div>

        <ErrorBoundary resetOnPropsChange showDetails={false}>
          <div className="p-4 bg-white border border-gray-200 rounded-lg">
            <h4 className="text-sm font-semibold mb-2">User Profile: {userId}</h4>
            <BuggyComponent shouldThrow={shouldThrow} />
          </div>
        </ErrorBoundary>
      </div>;
  }
}`,...(Cr=(kr=j.parameters)==null?void 0:kr.docs)==null?void 0:Cr.source},description:{story:"Reset on props change demonstration",...(Tr=(Br=j.parameters)==null?void 0:Br.docs)==null?void 0:Tr.description}}};var Sr,Dr,Ir,Rr,Ar;w.parameters={...w.parameters,docs:{...(Sr=w.parameters)==null?void 0:Sr.docs,source:{originalSource:`{
  args: {
    children: <div>Content</div>
  },
  render: () => <div className="space-y-6 max-w-3xl">
      <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
        <h4 className="text-sm font-semibold text-gray-900 mb-2">
          Accessibility Features
        </h4>
        <ul className="text-sm text-gray-700 space-y-1 list-disc list-inside">
          <li>
            Uses <code>role="alert"</code> for immediate screen reader announcement
          </li>
          <li>
            <code>aria-live="assertive"</code> ensures errors are announced
            immediately
          </li>
          <li>Buttons are fully keyboard accessible with focus indicators</li>
          <li>Error icons are marked with <code>aria-hidden="true"</code></li>
          <li>Clear, semantic heading structure</li>
          <li>Collapsible sections use proper details/summary elements</li>
          <li>Color contrast meets WCAG 2.1 AA requirements</li>
        </ul>
      </div>

      <ErrorBoundary showDetails boundaryName="Accessible Error Example">
        <BrokenComponent />
      </ErrorBoundary>
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
        }, {
          id: 'aria-roles',
          enabled: true
        }]
      }
    }
  }
}`,...(Ir=(Dr=w.parameters)==null?void 0:Dr.docs)==null?void 0:Ir.source},description:{story:"Accessibility features demonstration",...(Ar=(Rr=w.parameters)==null?void 0:Rr.docs)==null?void 0:Ar.description}}};const Hr=["Default","WithCustomMessage","WithActions","Nested","DevelopmentMode","ProductionMode","WithFallbackUI","Playground","RealWorldExamples","WithErrorLogging","ResetOnPropsChange","Accessibility"];export{w as Accessibility,u as Default,x as DevelopmentMode,p as Nested,v as Playground,b as ProductionMode,f as RealWorldExamples,j as ResetOnPropsChange,g as WithActions,h as WithCustomMessage,N as WithErrorLogging,y as WithFallbackUI,Hr as __namedExportsOrder,$r as default};
