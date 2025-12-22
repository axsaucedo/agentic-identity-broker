var _e=Object.defineProperty;var Ye=(r,o,t)=>o in r?_e(r,o,{enumerable:!0,configurable:!0,writable:!0,value:t}):r[o]=t;var A=(r,o,t)=>Ye(r,typeof o!="symbol"?o+"":o,t);import{j as e}from"./jsx-runtime-BYYWji4R.js";import{r as c}from"./index-ClcD9ViR.js";import{c as u}from"./cn-JCLedEej.js";import{B as a}from"./Button-DlbfybpJ.js";import"./_commonjsHelpers-Cpj98o6Y.js";const Je=()=>e.jsx("svg",{className:"w-24 h-24",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:1.5,d:"M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"})}),Ke=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"})}),Qe=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"})}),Xe=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"})}),Ze=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"})}),er=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"})});function rr(r){const o=r.message.toLowerCase();return o.includes("network")||o.includes("fetch")||o.includes("timeout")||o.includes("connection")||o.includes("offline")?"network":o.includes("permission")||o.includes("unauthorized")||o.includes("forbidden")||o.includes("access denied")?"permission":"generic"}class d extends c.Component{constructor(t){super(t);A(this,"handleReload",()=>{window.location.reload()});A(this,"handleNavigateHome",()=>{window.location.href="/"});A(this,"handleGracefulMode",()=>{this.setState({isGracefulMode:!0})});this.state={hasError:!1,error:null,errorInfo:null,errorCategory:"generic",errorCount:0,isGracefulMode:!1}}static getDerivedStateFromError(t){return{hasError:!0,error:t,errorCategory:rr(t)}}componentDidCatch(t,s){const{onError:n}=this.props,{errorCount:i}=this.state;if(this.setState({errorInfo:s,errorCount:i+1}),n)try{n(t,s)}catch(l){console.error("Error in GlobalErrorBoundary onError handler:",l)}}getErrorTitle(){const{errorCategory:t}=this.state,{appName:s="Application"}=this.props;switch(t){case"network":return"Connection Problem";case"permission":return"Access Denied";default:return`${s} Encountered an Error`}}getErrorDescription(){const{errorCategory:t}=this.state;switch(t){case"network":return"We're having trouble connecting to our servers. This might be due to your internet connection or a temporary service outage.";case"permission":return"You don't have permission to access this resource. Please contact your administrator if you believe this is an error.";default:return"An unexpected error has occurred. We apologize for the inconvenience. Our team has been notified and is working to resolve the issue."}}renderFooter(){const{supportEmail:t,supportPhone:s,supportUrl:n,appName:i="Application"}=this.props;return!t&&!s&&!n?null:e.jsx("footer",{className:"mt-12 pt-8 border-t border-error-primary/20",children:e.jsxs("div",{className:"max-w-xl mx-auto",children:[e.jsx("h3",{className:"text-sm font-semibold text-error-dark/80 mb-4 text-center",children:"Need Help?"}),e.jsxs("div",{className:"grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4",children:[t&&e.jsxs("a",{href:`mailto:${t}`,className:u("flex items-center gap-3 p-4 rounded-lg","bg-white border border-error-primary/20","hover:border-error-primary/40 hover:bg-error-light/30","transition-colors duration-200","focus:outline-none focus:ring-2 focus:ring-error-primary focus:ring-offset-2"),children:[e.jsx(Xe,{}),e.jsxs("div",{className:"flex-1 min-w-0 text-left",children:[e.jsx("div",{className:"text-xs text-error-dark/60 font-medium",children:"Email Support"}),e.jsx("div",{className:"text-sm text-error-dark font-semibold truncate",children:t})]})]}),s&&e.jsxs("a",{href:`tel:${s}`,className:u("flex items-center gap-3 p-4 rounded-lg","bg-white border border-error-primary/20","hover:border-error-primary/40 hover:bg-error-light/30","transition-colors duration-200","focus:outline-none focus:ring-2 focus:ring-error-primary focus:ring-offset-2"),children:[e.jsx(Ze,{}),e.jsxs("div",{className:"flex-1 min-w-0 text-left",children:[e.jsx("div",{className:"text-xs text-error-dark/60 font-medium",children:"Call Support"}),e.jsx("div",{className:"text-sm text-error-dark font-semibold truncate",children:s})]})]}),n&&e.jsxs("a",{href:n,target:"_blank",rel:"noopener noreferrer",className:u("flex items-center gap-3 p-4 rounded-lg","bg-white border border-error-primary/20","hover:border-error-primary/40 hover:bg-error-light/30","transition-colors duration-200","focus:outline-none focus:ring-2 focus:ring-error-primary focus:ring-offset-2"),children:[e.jsx(er,{}),e.jsxs("div",{className:"flex-1 min-w-0 text-left",children:[e.jsx("div",{className:"text-xs text-error-dark/60 font-medium",children:"Documentation"}),e.jsx("div",{className:"text-sm text-error-dark font-semibold truncate",children:"Help Center"})]})]})]}),e.jsxs("p",{className:"mt-6 text-xs text-error-dark/50 text-center",children:[i," Support Team - We're here to help 24/7"]})]})})}renderFullPageError(){const{error:t,errorInfo:s,errorCount:n,errorCategory:i}=this.state,{showDetails:l,appName:g="Application",recoveryStrategies:m=["reload","navigate","contact-support"],allowGracefulDegradation:x}=this.props;if(!t)return null;const D=m.includes("reload"),Oe=m.includes("navigate"),Ve=x&&m.includes("contact-support");return e.jsx("div",{role:"alert","aria-live":"assertive",className:"min-h-screen w-full bg-gradient-to-br from-error-light via-error-light/70 to-error-light/50 flex items-center justify-center p-6",children:e.jsxs("div",{className:"max-w-2xl w-full",children:[e.jsxs("div",{className:"bg-white border-2 border-error-primary/30 rounded-2xl shadow-2xl p-8 sm:p-12",children:[e.jsxs("div",{className:"flex flex-col items-center text-center mb-8",children:[e.jsx("div",{className:"text-error-primary mb-6",children:e.jsx(Je,{})}),e.jsx("h1",{className:"text-3xl sm:text-4xl font-bold text-error-dark mb-3",children:this.getErrorTitle()}),e.jsx("p",{className:"text-lg text-error-dark/80 leading-relaxed",children:this.getErrorDescription()})]}),l&&e.jsxs("div",{className:"mb-8 space-y-4",children:[e.jsxs("div",{className:"bg-error-dark/5 border border-error-primary/20 rounded-lg p-4",children:[e.jsxs("h3",{className:"text-xs font-semibold text-error-dark uppercase tracking-wide mb-2 flex items-center gap-2",children:[e.jsx("span",{className:"inline-block w-2 h-2 bg-error-primary rounded-full"}),"Error Details"]}),e.jsx("p",{className:"text-sm font-mono text-error-dark break-words",children:t.toString()}),i&&e.jsxs("div",{className:"mt-2 inline-flex items-center gap-2 px-2 py-1 bg-error-primary/10 rounded text-xs font-semibold text-error-dark",children:["Category: ",i]})]}),(s==null?void 0:s.componentStack)&&e.jsxs("details",{className:"bg-error-dark/5 border border-error-primary/20 rounded-lg",children:[e.jsx("summary",{className:"cursor-pointer p-4 text-xs font-semibold text-error-dark uppercase tracking-wide hover:bg-error-dark/10 transition-colors",children:"Component Stack (Click to expand)"}),e.jsx("pre",{className:"p-4 pt-2 text-xs font-mono text-error-dark overflow-x-auto max-h-64 overflow-y-auto",children:s.componentStack})]}),t.stack&&e.jsxs("details",{className:"bg-error-dark/5 border border-error-primary/20 rounded-lg",children:[e.jsx("summary",{className:"cursor-pointer p-4 text-xs font-semibold text-error-dark uppercase tracking-wide hover:bg-error-dark/10 transition-colors",children:"Error Stack (Click to expand)"}),e.jsx("pre",{className:"p-4 pt-2 text-xs font-mono text-error-dark overflow-x-auto max-h-64 overflow-y-auto",children:t.stack})]}),n>1&&e.jsxs("div",{className:"text-xs text-error-dark/70 bg-error-dark/5 rounded-lg px-4 py-3 flex items-center gap-2",children:[e.jsx("span",{className:"inline-block w-2 h-2 bg-amber-500 rounded-full animate-pulse"}),"This error has occurred ",e.jsx("strong",{children:n})," time",n===1?"":"s"," during this session."]})]}),e.jsxs("div",{className:"flex flex-col sm:flex-row gap-3 justify-center",children:[D&&e.jsxs("button",{type:"button",onClick:this.handleReload,className:u("inline-flex items-center justify-center gap-2 px-6 py-3 rounded-lg","bg-error-primary text-white font-semibold text-base","hover:bg-error-dark transition-all duration-200","focus:outline-none focus:ring-2 focus:ring-error-primary focus:ring-offset-2","shadow-lg hover:shadow-xl hover:-translate-y-0.5"),children:[e.jsx(Ke,{}),"Reload ",g]}),Oe&&e.jsxs("button",{type:"button",onClick:this.handleNavigateHome,className:u("inline-flex items-center justify-center gap-2 px-6 py-3 rounded-lg","bg-white text-error-dark font-semibold text-base","border-2 border-error-primary/30","hover:bg-error-light hover:border-error-primary/50 transition-all duration-200","focus:outline-none focus:ring-2 focus:ring-error-primary focus:ring-offset-2","shadow-md hover:shadow-lg"),children:[e.jsx(Qe,{}),"Go to Home"]}),Ve&&e.jsx("button",{type:"button",onClick:this.handleGracefulMode,className:u("inline-flex items-center justify-center gap-2 px-6 py-3 rounded-lg","bg-white text-error-dark font-medium text-base","border border-error-primary/20","hover:bg-error-light/30 hover:border-error-primary/30 transition-all duration-200","focus:outline-none focus:ring-2 focus:ring-error-primary focus:ring-offset-2"),children:"Continue with Limited Features"})]}),l&&e.jsx("div",{className:"mt-8 pt-6 border-t border-error-primary/20",children:e.jsxs("p",{className:"text-xs text-error-dark/50 text-center",children:[e.jsx("strong",{children:"Developer Mode:"})," Detailed error information is shown because"," ",e.jsx("code",{className:"bg-error-dark/10 px-1.5 py-0.5 rounded font-mono",children:"showDetails"})," ","is enabled. Disable in production."]})})]}),this.renderFooter()]})})}render(){const{hasError:t,isGracefulMode:s}=this.state,{children:n,allowGracefulDegradation:i,gracefulFallback:l}=this.props;return t?s&&i&&l?l:this.renderFullPageError():n}}d.__docgenInfo={description:`GlobalErrorBoundary component for catching critical application errors.
Must be a class component as per React requirements.

@example
\`\`\`tsx
// Basic usage - wrap entire app
<GlobalErrorBoundary appName="Identity Broker">
  <App />
</GlobalErrorBoundary>

// With support information
<GlobalErrorBoundary
  appName="Identity Broker"
  supportEmail="support@example.com"
  supportPhone="1-800-SUPPORT"
  onError={(error, errorInfo) => logToSentry(error, errorInfo)}
>
  <App />
</GlobalErrorBoundary>

// With graceful degradation
<GlobalErrorBoundary
  allowGracefulDegradation
  gracefulFallback={<LimitedFunctionalityApp />}
>
  <FullFeaturedApp />
</GlobalErrorBoundary>
\`\`\``,methods:[{name:"handleReload",docblock:null,modifiers:[],params:[],returns:{type:{name:"void"}}},{name:"handleNavigateHome",docblock:null,modifiers:[],params:[],returns:{type:{name:"void"}}},{name:"handleGracefulMode",docblock:null,modifiers:[],params:[],returns:{type:{name:"void"}}},{name:"getErrorTitle",docblock:null,modifiers:[],params:[],returns:{type:{name:"string"}}},{name:"getErrorDescription",docblock:null,modifiers:[],params:[],returns:{type:{name:"string"}}},{name:"renderFooter",docblock:null,modifiers:[],params:[],returns:{type:{name:"ReactNode"}}},{name:"renderFullPageError",docblock:null,modifiers:[],params:[],returns:{type:{name:"ReactNode"}}}],displayName:"GlobalErrorBoundary",props:{children:{required:!0,tsType:{name:"ReactNode"},description:"Child components to protect with error boundary"},appName:{required:!1,tsType:{name:"string"},description:"Application name to display in error messages"},supportEmail:{required:!1,tsType:{name:"string"},description:"Support email for users to contact"},onError:{required:!1,tsType:{name:"signature",type:"function",raw:"(error: Error, errorInfo: ErrorInfo) => void",signature:{arguments:[{type:{name:"Error"},name:"error"},{type:{name:"ErrorInfo"},name:"errorInfo"}],return:{name:"void"}}},description:"Callback when error is caught - use for error reporting services"},showDetails:{required:!1,tsType:{name:"boolean"},description:"Show detailed error information (dev mode)"},recoveryStrategies:{required:!1,tsType:{name:"Array",elements:[{name:"unknown"}],raw:"('reload' | 'navigate' | 'contact-support')[]"},description:"Recovery strategies available to user"},allowGracefulDegradation:{required:!1,tsType:{name:"boolean"},description:"Allow graceful degradation instead of full error page"},gracefulFallback:{required:!1,tsType:{name:"ReactNode"},description:"Fallback content for graceful degradation mode"},supportPhone:{required:!1,tsType:{name:"string"},description:"Custom support phone number"},supportUrl:{required:!1,tsType:{name:"string"},description:"Custom support URL/documentation"}}};const lr={title:"Design System/Feedback/GlobalErrorBoundary",component:d,parameters:{layout:"fullscreen"},tags:["autodocs"],argTypes:{appName:{control:"text",description:"Application name to display in error messages"},supportEmail:{control:"text",description:"Support email for users to contact"},supportPhone:{control:"text",description:"Support phone number"},supportUrl:{control:"text",description:"URL to support documentation or help center"},showDetails:{control:"boolean",description:"Show detailed error information (dev mode)"},allowGracefulDegradation:{control:"boolean",description:"Allow graceful degradation instead of full error page"},recoveryStrategies:{control:"check",options:["reload","navigate","contact-support"],description:"Available recovery strategies"}}},p=({shouldThrow:r})=>{if(r)throw new Error("Critical application error occurred");return e.jsx("div",{className:"min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-6",children:e.jsxs("div",{className:"bg-white rounded-xl shadow-lg p-8 max-w-lg w-full",children:[e.jsx("h2",{className:"text-2xl font-bold text-gray-900 mb-4",children:"Application Running Normally"}),e.jsx("p",{className:"text-gray-700 mb-6",children:"This represents your full application. Click the button below to simulate a critical error that will be caught by the GlobalErrorBoundary."}),e.jsxs("div",{className:"flex gap-3",children:[e.jsx(a,{variant:"primary",size:"md",children:"Normal Action"}),e.jsx(a,{variant:"secondary",size:"md",children:"Another Action"})]})]})})},h=()=>{throw new Error("Application failed to initialize properly")},He=({shouldThrow:r})=>{if(r)throw new Error("Network connection timeout - failed to fetch data from server");return e.jsx(p,{shouldThrow:!1})},$e=({shouldThrow:r})=>{if(r)throw new Error("Permission denied: Unauthorized access to protected resource");return e.jsx(p,{shouldThrow:!1})},f={render:()=>{const[r,o]=c.useState(!1);return e.jsxs("div",{className:"relative",children:[!r&&e.jsx("div",{className:"fixed top-4 left-1/2 transform -translate-x-1/2 z-50",children:e.jsx(a,{variant:"danger",onClick:()=>o(!0),size:"lg",children:"Trigger Critical Error"})}),e.jsx(d,{appName:"Identity Broker",children:e.jsx(p,{shouldThrow:r})})]})}},b={render:()=>e.jsx(d,{appName:"Identity Broker",showDetails:!1,children:e.jsx(h,{})})},v={render:()=>e.jsx(d,{appName:"Identity Broker",supportEmail:"support@identity-broker.com",supportPhone:"1-800-555-0123",supportUrl:"https://docs.identity-broker.com/help",showDetails:!1,children:e.jsx(h,{})})},y={render:()=>{const[r,o]=c.useState([]),t=(i,l)=>{const m=`[${new Date().toLocaleTimeString()}] Error reported: ${i.message}`;o(x=>[...x,m]),console.log("Sending error to monitoring service:",{error:i.message,stack:i.stack,componentStack:l.componentStack})},[s,n]=c.useState(!1);return e.jsxs("div",{className:"relative",children:[r.length>0&&!s&&e.jsxs("div",{className:"fixed top-4 right-4 z-50 bg-gray-900 text-green-400 rounded-lg p-4 max-w-md font-mono text-xs shadow-2xl",children:[e.jsxs("div",{className:"flex justify-between items-center mb-2",children:[e.jsx("span",{className:"font-semibold",children:"Error Reporting Log:"}),e.jsx("button",{onClick:()=>o([]),className:"text-gray-400 hover:text-white underline",children:"Clear"})]}),r.map((i,l)=>e.jsx("div",{className:"mb-1",children:i},l))]}),!s&&e.jsx("div",{className:"fixed top-4 left-1/2 transform -translate-x-1/2 z-50",children:e.jsx(a,{variant:"danger",onClick:()=>n(!0),size:"lg",children:"Trigger Error (with Reporting)"})}),e.jsx(d,{appName:"Identity Broker",onError:t,showDetails:!1,children:e.jsx(p,{shouldThrow:s})})]})}},w={render:()=>{const[r,o]=c.useState(!1),[t,s]=c.useState(1);return e.jsxs("div",{className:"relative",children:[!r&&e.jsxs("div",{className:"fixed top-4 left-1/2 transform -translate-x-1/2 z-50 flex flex-col items-center gap-2",children:[e.jsxs(a,{variant:"danger",onClick:()=>o(!0),size:"lg",children:["Trigger Error (Attempt #",t,")"]}),e.jsx("span",{className:"text-xs text-gray-600 bg-white px-3 py-1 rounded-full shadow",children:"Try the recovery buttons to reload or navigate"})]}),e.jsx(d,{appName:"Identity Broker",recoveryStrategies:["reload","navigate","contact-support"],showDetails:!1,onError:()=>s(t+1),children:e.jsx(p,{shouldThrow:r})})]})}},N={render:()=>{const[r,o]=c.useState(!1),t=()=>e.jsx("div",{className:"min-h-screen bg-gradient-to-br from-amber-50 to-orange-100 flex items-center justify-center p-6",children:e.jsxs("div",{className:"bg-white rounded-xl shadow-lg p-8 max-w-lg w-full border-2 border-amber-400",children:[e.jsxs("div",{className:"flex items-center gap-3 mb-4",children:[e.jsx("div",{className:"w-12 h-12 bg-amber-100 rounded-full flex items-center justify-center",children:e.jsx("svg",{className:"w-6 h-6 text-amber-600",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"})})}),e.jsx("h2",{className:"text-2xl font-bold text-amber-900",children:"Limited Functionality Mode"})]}),e.jsx("p",{className:"text-amber-800 mb-6",children:"Some features are temporarily unavailable, but you can continue using essential functions. Full functionality will be restored shortly."}),e.jsxs("div",{className:"space-y-3",children:[e.jsx(a,{variant:"primary",size:"md",fullWidth:!0,children:"Access Core Features"}),e.jsx(a,{variant:"outline",size:"md",fullWidth:!0,children:"View Available Services"}),e.jsx("button",{onClick:()=>window.location.reload(),className:"w-full text-sm text-amber-700 hover:text-amber-900 underline",children:"Reload to restore full functionality"})]})]})});return e.jsxs("div",{className:"relative",children:[!r&&e.jsx("div",{className:"fixed top-4 left-1/2 transform -translate-x-1/2 z-50",children:e.jsx(a,{variant:"danger",onClick:()=>o(!0),size:"lg",children:"Trigger Error (Graceful Mode)"})}),e.jsx(d,{appName:"Identity Broker",allowGracefulDegradation:!0,gracefulFallback:e.jsx(t,{}),recoveryStrategies:["reload","navigate","contact-support"],showDetails:!1,children:e.jsx(p,{shouldThrow:r})})]})}},k={render:()=>{const[r,o]=c.useState("none");return e.jsxs("div",{className:"relative",children:[r==="none"&&e.jsx("div",{className:"fixed top-4 left-1/2 transform -translate-x-1/2 z-50",children:e.jsxs("div",{className:"bg-white rounded-xl shadow-2xl p-6 border-2 border-gray-200",children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-4 text-center",children:"Simulate Different Error Types"}),e.jsxs("div",{className:"flex flex-col gap-2",children:[e.jsx(a,{variant:"danger",onClick:()=>o("network"),size:"md",fullWidth:!0,children:"Network Error"}),e.jsx(a,{variant:"danger",onClick:()=>o("permission"),size:"md",fullWidth:!0,children:"Permission Error"}),e.jsx(a,{variant:"danger",onClick:()=>o("generic"),size:"md",fullWidth:!0,children:"Generic Error"})]})]})}),e.jsxs(d,{appName:"Identity Broker",supportEmail:"support@identity-broker.com",showDetails:!1,children:[r==="network"&&e.jsx(He,{shouldThrow:!0}),r==="permission"&&e.jsx($e,{shouldThrow:!0}),r==="generic"&&e.jsx(h,{}),r==="none"&&e.jsx(p,{shouldThrow:!1})]})]})}},j={render:()=>e.jsxs("div",{className:"relative",children:[e.jsx("div",{className:"fixed top-4 left-1/2 transform -translate-x-1/2 z-50",children:e.jsx("div",{className:"bg-blue-900 text-white px-4 py-2 rounded-lg shadow-lg text-sm font-semibold",children:"Development Mode - Full Error Details Shown"})}),e.jsx(d,{appName:"Identity Broker",showDetails:!0,supportEmail:"dev-support@identity-broker.com",supportUrl:"https://docs.identity-broker.com/troubleshooting",children:e.jsx(h,{})})]})},E={render:()=>e.jsxs("div",{className:"relative",children:[e.jsx("div",{className:"fixed top-4 left-1/2 transform -translate-x-1/2 z-50",children:e.jsx("div",{className:"bg-green-900 text-white px-4 py-2 rounded-lg shadow-lg text-sm font-semibold",children:"Production Mode - User-Friendly Display"})}),e.jsx(d,{appName:"Identity Broker",showDetails:!1,supportEmail:"support@identity-broker.com",supportPhone:"1-800-555-0123",supportUrl:"https://help.identity-broker.com",children:e.jsx(h,{})})]})},B={render:()=>{const[r,o]=c.useState([]),[t,s]=c.useState(!1),n=(l,g)=>{const x=`[${new Date().toISOString()}] ${l.message}`;o(D=>[...D,x])},i=()=>e.jsx("div",{className:"min-h-screen bg-gradient-to-br from-amber-50 to-orange-100 flex items-center justify-center p-6",children:e.jsxs("div",{className:"bg-white rounded-xl shadow-lg p-8 max-w-lg w-full",children:[e.jsx("h2",{className:"text-2xl font-bold text-amber-900 mb-4",children:"Limited Mode Active"}),e.jsx("p",{className:"text-amber-800 mb-6",children:"Operating with reduced functionality. Core features remain available."}),e.jsx(a,{variant:"primary",size:"md",fullWidth:!0,children:"Access Core Features"})]})});return e.jsxs("div",{className:"relative",children:[r.length>0&&!t&&e.jsxs("div",{className:"fixed top-4 right-4 z-50 bg-gray-900 text-green-400 rounded-lg p-4 max-w-sm font-mono text-xs shadow-2xl",children:[e.jsx("div",{className:"font-semibold mb-2",children:"Error Logs:"}),r.slice(-3).map((l,g)=>e.jsx("div",{className:"truncate",children:l},g))]}),!t&&e.jsx("div",{className:"fixed top-4 left-1/2 transform -translate-x-1/2 z-50",children:e.jsx(a,{variant:"danger",onClick:()=>s(!0),size:"lg",children:"Trigger Error (Full Featured)"})}),e.jsx(d,{appName:"Identity Broker",supportEmail:"support@identity-broker.com",supportPhone:"1-800-555-0123",supportUrl:"https://docs.identity-broker.com/help",onError:n,showDetails:!1,recoveryStrategies:["reload","navigate","contact-support"],allowGracefulDegradation:!0,gracefulFallback:e.jsx(i,{}),children:e.jsx(p,{shouldThrow:t})})]})}},S={render:r=>{const[o,t]=c.useState(!1);return e.jsxs("div",{className:"relative",children:[!o&&e.jsx("div",{className:"fixed top-4 left-1/2 transform -translate-x-1/2 z-50",children:e.jsxs("div",{className:"bg-white rounded-xl shadow-2xl p-6 border-2 border-gray-200",children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-4 text-center",children:"Playground Controls"}),e.jsx("p",{className:"text-sm text-gray-600 mb-4 text-center",children:"Adjust controls in the panel below, then trigger the error"}),e.jsx(a,{variant:"danger",onClick:()=>t(!0),size:"lg",fullWidth:!0,children:"Trigger Error"})]})}),e.jsx(d,{...r,children:e.jsx(p,{shouldThrow:o})})]})},args:{appName:"Identity Broker",supportEmail:"support@identity-broker.com",supportPhone:"1-800-555-0123",supportUrl:"https://docs.identity-broker.com/help",showDetails:!1,recoveryStrategies:["reload","navigate","contact-support"],allowGracefulDegradation:!1}},T={render:()=>{const[r,o]=c.useState("none"),t=()=>{throw new Error("Authentication service failed to initialize")};return e.jsxs("div",{className:"relative",children:[r==="none"&&e.jsx("div",{className:"fixed top-4 left-1/2 transform -translate-x-1/2 z-50",children:e.jsxs("div",{className:"bg-white rounded-xl shadow-2xl p-6 border-2 border-gray-200 max-w-md",children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-4",children:"Real-World Error Scenarios"}),e.jsx("p",{className:"text-sm text-gray-600 mb-4",children:"Simulate common errors in consent management applications"}),e.jsxs("div",{className:"space-y-2",children:[e.jsx(a,{variant:"danger",onClick:()=>o("auth"),size:"md",fullWidth:!0,children:"Authentication Service Failure"}),e.jsx(a,{variant:"danger",onClick:()=>o("network"),size:"md",fullWidth:!0,children:"Network Connectivity Issue"}),e.jsx(a,{variant:"danger",onClick:()=>o("permission"),size:"md",fullWidth:!0,children:"Insufficient Permissions"})]})]})}),e.jsxs(d,{appName:"Consent Manager",supportEmail:"support@consent-manager.com",supportPhone:"1-800-CONSENT",supportUrl:"https://docs.consent-manager.com",showDetails:!1,recoveryStrategies:["reload","navigate","contact-support"],children:[r==="auth"&&e.jsx(t,{}),r==="network"&&e.jsx(He,{shouldThrow:!0}),r==="permission"&&e.jsx($e,{shouldThrow:!0}),r==="none"&&e.jsx(p,{shouldThrow:!1})]})]})}},C={render:()=>e.jsxs("div",{className:"relative",children:[e.jsx("div",{className:"fixed top-4 left-1/2 transform -translate-x-1/2 z-50",children:e.jsxs("div",{className:"bg-white rounded-xl shadow-2xl p-6 border-2 border-blue-200 max-w-lg",children:[e.jsx("h3",{className:"text-lg font-semibold text-blue-900 mb-3",children:"Accessibility Features"}),e.jsxs("ul",{className:"text-sm text-blue-800 space-y-2 list-disc list-inside",children:[e.jsx("li",{children:'role="alert" for immediate screen reader announcement'}),e.jsx("li",{children:'aria-live="assertive" for critical errors'}),e.jsx("li",{children:"Full keyboard navigation support"}),e.jsx("li",{children:"WCAG 2.1 AA compliant color contrast"}),e.jsx("li",{children:"Clear, semantic heading structure"}),e.jsx("li",{children:"Focus indicators on all interactive elements"})]})]})}),e.jsx(d,{appName:"Identity Broker",supportEmail:"accessibility@identity-broker.com",showDetails:!0,children:e.jsx(h,{})})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0},{id:"button-name",enabled:!0},{id:"aria-roles",enabled:!0},{id:"landmark-one-main",enabled:!1}]}}}};var z,G,L,F,I;f.parameters={...f.parameters,docs:{...(z=f.parameters)==null?void 0:z.docs,source:{originalSource:`{
  render: () => {
    const [shouldThrow, setShouldThrow] = useState(false);
    return <div className="relative">
        {!shouldThrow && <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
            <Button variant="danger" onClick={() => setShouldThrow(true)} size="lg">
              Trigger Critical Error
            </Button>
          </div>}

        <GlobalErrorBoundary appName="Identity Broker">
          <BuggyApp shouldThrow={shouldThrow} />
        </GlobalErrorBoundary>
      </div>;
  }
}`,...(L=(G=f.parameters)==null?void 0:G.docs)==null?void 0:L.source},description:{story:"Default - Basic global boundary wrapper",...(I=(F=f.parameters)==null?void 0:F.docs)==null?void 0:I.description}}};var W,P,R,M,U;b.parameters={...b.parameters,docs:{...(W=b.parameters)==null?void 0:W.docs,source:{originalSource:`{
  render: () => {
    return <GlobalErrorBoundary appName="Identity Broker" showDetails={false}>
        <BrokenApp />
      </GlobalErrorBoundary>;
  }
}`,...(R=(P=b.parameters)==null?void 0:P.docs)==null?void 0:R.source},description:{story:"WithFullPageError - Large full-page error display",...(U=(M=b.parameters)==null?void 0:M.docs)==null?void 0:U.description}}};var q,H,$,O,V;v.parameters={...v.parameters,docs:{...(q=v.parameters)==null?void 0:q.docs,source:{originalSource:`{
  render: () => {
    return <GlobalErrorBoundary appName="Identity Broker" supportEmail="support@identity-broker.com" supportPhone="1-800-555-0123" supportUrl="https://docs.identity-broker.com/help" showDetails={false}>
        <BrokenApp />
      </GlobalErrorBoundary>;
  }
}`,...($=(H=v.parameters)==null?void 0:H.docs)==null?void 0:$.source},description:{story:"WithFooterActions - Footer with support/contact information",...(V=(O=v.parameters)==null?void 0:O.docs)==null?void 0:V.description}}};var _,Y,J,K,Q;y.parameters={...y.parameters,docs:{...(_=y.parameters)==null?void 0:_.docs,source:{originalSource:`{
  render: () => {
    const [errorLogs, setErrorLogs] = useState<string[]>([]);
    const handleError = (error: Error, errorInfo: React.ErrorInfo) => {
      const timestamp = new Date().toLocaleTimeString();
      const logEntry = \`[\${timestamp}] Error reported: \${error.message}\`;
      setErrorLogs(prev => [...prev, logEntry]);

      // In production, send to error tracking service
      // Example: Sentry.captureException(error, { contexts: { react: errorInfo } });
      console.log('Sending error to monitoring service:', {
        error: error.message,
        stack: error.stack,
        componentStack: errorInfo.componentStack
      });
    };
    const [shouldThrow, setShouldThrow] = useState(false);
    return <div className="relative">
        {/* Error log display */}
        {errorLogs.length > 0 && !shouldThrow && <div className="fixed top-4 right-4 z-50 bg-gray-900 text-green-400 rounded-lg p-4 max-w-md font-mono text-xs shadow-2xl">
            <div className="flex justify-between items-center mb-2">
              <span className="font-semibold">Error Reporting Log:</span>
              <button onClick={() => setErrorLogs([])} className="text-gray-400 hover:text-white underline">
                Clear
              </button>
            </div>
            {errorLogs.map((log, index) => <div key={index} className="mb-1">
                {log}
              </div>)}
          </div>}

        {!shouldThrow && <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
            <Button variant="danger" onClick={() => setShouldThrow(true)} size="lg">
              Trigger Error (with Reporting)
            </Button>
          </div>}

        <GlobalErrorBoundary appName="Identity Broker" onError={handleError} showDetails={false}>
          <BuggyApp shouldThrow={shouldThrow} />
        </GlobalErrorBoundary>
      </div>;
  }
}`,...(J=(Y=y.parameters)==null?void 0:Y.docs)==null?void 0:J.source},description:{story:"WithErrorReporting - Send error to backend/logging service",...(Q=(K=y.parameters)==null?void 0:K.docs)==null?void 0:Q.description}}};var X,Z,ee,re,te;w.parameters={...w.parameters,docs:{...(X=w.parameters)==null?void 0:X.docs,source:{originalSource:`{
  render: () => {
    const [shouldThrow, setShouldThrow] = useState(false);
    const [attempt, setAttempt] = useState(1);
    return <div className="relative">
        {!shouldThrow && <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50 flex flex-col items-center gap-2">
            <Button variant="danger" onClick={() => setShouldThrow(true)} size="lg">
              Trigger Error (Attempt #{attempt})
            </Button>
            <span className="text-xs text-gray-600 bg-white px-3 py-1 rounded-full shadow">
              Try the recovery buttons to reload or navigate
            </span>
          </div>}

        <GlobalErrorBoundary appName="Identity Broker" recoveryStrategies={['reload', 'navigate', 'contact-support']} showDetails={false} onError={() => setAttempt(attempt + 1)}>
          <BuggyApp shouldThrow={shouldThrow} />
        </GlobalErrorBoundary>
      </div>;
  }
}`,...(ee=(Z=w.parameters)==null?void 0:Z.docs)==null?void 0:ee.source},description:{story:"WithRecovery - Auto-recovery/retry mechanisms",...(te=(re=w.parameters)==null?void 0:re.docs)==null?void 0:te.description}}};var oe,se,ae,ne,ie;N.parameters={...N.parameters,docs:{...(oe=N.parameters)==null?void 0:oe.docs,source:{originalSource:`{
  render: () => {
    const [shouldThrow, setShouldThrow] = useState(false);
    const LimitedApp = () => <div className="min-h-screen bg-gradient-to-br from-amber-50 to-orange-100 flex items-center justify-center p-6">
        <div className="bg-white rounded-xl shadow-lg p-8 max-w-lg w-full border-2 border-amber-400">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-12 h-12 bg-amber-100 rounded-full flex items-center justify-center">
              <svg className="w-6 h-6 text-amber-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <h2 className="text-2xl font-bold text-amber-900">
              Limited Functionality Mode
            </h2>
          </div>
          <p className="text-amber-800 mb-6">
            Some features are temporarily unavailable, but you can continue using
            essential functions. Full functionality will be restored shortly.
          </p>
          <div className="space-y-3">
            <Button variant="primary" size="md" fullWidth>
              Access Core Features
            </Button>
            <Button variant="outline" size="md" fullWidth>
              View Available Services
            </Button>
            <button onClick={() => window.location.reload()} className="w-full text-sm text-amber-700 hover:text-amber-900 underline">
              Reload to restore full functionality
            </button>
          </div>
        </div>
      </div>;
    return <div className="relative">
        {!shouldThrow && <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
            <Button variant="danger" onClick={() => setShouldThrow(true)} size="lg">
              Trigger Error (Graceful Mode)
            </Button>
          </div>}

        <GlobalErrorBoundary appName="Identity Broker" allowGracefulDegradation gracefulFallback={<LimitedApp />} recoveryStrategies={['reload', 'navigate', 'contact-support']} showDetails={false}>
          <BuggyApp shouldThrow={shouldThrow} />
        </GlobalErrorBoundary>
      </div>;
  }
}`,...(ae=(se=N.parameters)==null?void 0:se.docs)==null?void 0:ae.source},description:{story:"GracefulDegradation - App continues with reduced features",...(ie=(ne=N.parameters)==null?void 0:ne.docs)==null?void 0:ie.description}}};var le,de,ce,pe,me;k.parameters={...k.parameters,docs:{...(le=k.parameters)==null?void 0:le.docs,source:{originalSource:`{
  render: () => {
    const [errorType, setErrorType] = useState<'none' | 'network' | 'permission' | 'generic'>('none');
    return <div className="relative">
        {errorType === 'none' && <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
            <div className="bg-white rounded-xl shadow-2xl p-6 border-2 border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900 mb-4 text-center">
                Simulate Different Error Types
              </h3>
              <div className="flex flex-col gap-2">
                <Button variant="danger" onClick={() => setErrorType('network')} size="md" fullWidth>
                  Network Error
                </Button>
                <Button variant="danger" onClick={() => setErrorType('permission')} size="md" fullWidth>
                  Permission Error
                </Button>
                <Button variant="danger" onClick={() => setErrorType('generic')} size="md" fullWidth>
                  Generic Error
                </Button>
              </div>
            </div>
          </div>}

        <GlobalErrorBoundary appName="Identity Broker" supportEmail="support@identity-broker.com" showDetails={false}>
          {errorType === 'network' && <NetworkErrorApp shouldThrow={true} />}
          {errorType === 'permission' && <PermissionErrorApp shouldThrow={true} />}
          {errorType === 'generic' && <BrokenApp />}
          {errorType === 'none' && <BuggyApp shouldThrow={false} />}
        </GlobalErrorBoundary>
      </div>;
  }
}`,...(ce=(de=k.parameters)==null?void 0:de.docs)==null?void 0:ce.source},description:{story:"ErrorCategories - Different error types (network, permission, etc.)",...(me=(pe=k.parameters)==null?void 0:pe.docs)==null?void 0:me.description}}};var ue,he,ge,xe,fe;j.parameters={...j.parameters,docs:{...(ue=j.parameters)==null?void 0:ue.docs,source:{originalSource:`{
  render: () => {
    return <div className="relative">
        <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
          <div className="bg-blue-900 text-white px-4 py-2 rounded-lg shadow-lg text-sm font-semibold">
            Development Mode - Full Error Details Shown
          </div>
        </div>

        <GlobalErrorBoundary appName="Identity Broker" showDetails={true} supportEmail="dev-support@identity-broker.com" supportUrl="https://docs.identity-broker.com/troubleshooting">
          <BrokenApp />
        </GlobalErrorBoundary>
      </div>;
  }
}`,...(ge=(he=j.parameters)==null?void 0:he.docs)==null?void 0:ge.source},description:{story:"Development mode with detailed error information",...(fe=(xe=j.parameters)==null?void 0:xe.docs)==null?void 0:fe.description}}};var be,ve,ye,we,Ne;E.parameters={...E.parameters,docs:{...(be=E.parameters)==null?void 0:be.docs,source:{originalSource:`{
  render: () => {
    return <div className="relative">
        <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
          <div className="bg-green-900 text-white px-4 py-2 rounded-lg shadow-lg text-sm font-semibold">
            Production Mode - User-Friendly Display
          </div>
        </div>

        <GlobalErrorBoundary appName="Identity Broker" showDetails={false} supportEmail="support@identity-broker.com" supportPhone="1-800-555-0123" supportUrl="https://help.identity-broker.com">
          <BrokenApp />
        </GlobalErrorBoundary>
      </div>;
  }
}`,...(ye=(ve=E.parameters)==null?void 0:ve.docs)==null?void 0:ye.source},description:{story:"Production mode with minimal error display",...(Ne=(we=E.parameters)==null?void 0:we.docs)==null?void 0:Ne.description}}};var ke,je,Ee,Be,Se;B.parameters={...B.parameters,docs:{...(ke=B.parameters)==null?void 0:ke.docs,source:{originalSource:`{
  render: () => {
    const [errorLogs, setErrorLogs] = useState<string[]>([]);
    const [shouldThrow, setShouldThrow] = useState(false);
    const handleError = (error: Error, errorInfo: React.ErrorInfo) => {
      const timestamp = new Date().toISOString();
      const logEntry = \`[\${timestamp}] \${error.message}\`;
      setErrorLogs(prev => [...prev, logEntry]);
    };
    const LimitedApp = () => <div className="min-h-screen bg-gradient-to-br from-amber-50 to-orange-100 flex items-center justify-center p-6">
        <div className="bg-white rounded-xl shadow-lg p-8 max-w-lg w-full">
          <h2 className="text-2xl font-bold text-amber-900 mb-4">
            Limited Mode Active
          </h2>
          <p className="text-amber-800 mb-6">
            Operating with reduced functionality. Core features remain available.
          </p>
          <Button variant="primary" size="md" fullWidth>
            Access Core Features
          </Button>
        </div>
      </div>;
    return <div className="relative">
        {errorLogs.length > 0 && !shouldThrow && <div className="fixed top-4 right-4 z-50 bg-gray-900 text-green-400 rounded-lg p-4 max-w-sm font-mono text-xs shadow-2xl">
            <div className="font-semibold mb-2">Error Logs:</div>
            {errorLogs.slice(-3).map((log, index) => <div key={index} className="truncate">
                {log}
              </div>)}
          </div>}

        {!shouldThrow && <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
            <Button variant="danger" onClick={() => setShouldThrow(true)} size="lg">
              Trigger Error (Full Featured)
            </Button>
          </div>}

        <GlobalErrorBoundary appName="Identity Broker" supportEmail="support@identity-broker.com" supportPhone="1-800-555-0123" supportUrl="https://docs.identity-broker.com/help" onError={handleError} showDetails={false} recoveryStrategies={['reload', 'navigate', 'contact-support']} allowGracefulDegradation gracefulFallback={<LimitedApp />}>
          <BuggyApp shouldThrow={shouldThrow} />
        </GlobalErrorBoundary>
      </div>;
  }
}`,...(Ee=(je=B.parameters)==null?void 0:je.docs)==null?void 0:Ee.source},description:{story:"Complete example with all features",...(Se=(Be=B.parameters)==null?void 0:Be.docs)==null?void 0:Se.description}}};var Te,Ce,Ae,De,ze;S.parameters={...S.parameters,docs:{...(Te=S.parameters)==null?void 0:Te.docs,source:{originalSource:`{
  render: args => {
    const [shouldThrow, setShouldThrow] = useState(false);
    return <div className="relative">
        {!shouldThrow && <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
            <div className="bg-white rounded-xl shadow-2xl p-6 border-2 border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900 mb-4 text-center">
                Playground Controls
              </h3>
              <p className="text-sm text-gray-600 mb-4 text-center">
                Adjust controls in the panel below, then trigger the error
              </p>
              <Button variant="danger" onClick={() => setShouldThrow(true)} size="lg" fullWidth>
                Trigger Error
              </Button>
            </div>
          </div>}

        <GlobalErrorBoundary {...args}>
          <BuggyApp shouldThrow={shouldThrow} />
        </GlobalErrorBoundary>
      </div>;
  },
  args: {
    appName: 'Identity Broker',
    supportEmail: 'support@identity-broker.com',
    supportPhone: '1-800-555-0123',
    supportUrl: 'https://docs.identity-broker.com/help',
    showDetails: false,
    recoveryStrategies: ['reload', 'navigate', 'contact-support'],
    allowGracefulDegradation: false
  }
}`,...(Ae=(Ce=S.parameters)==null?void 0:Ce.docs)==null?void 0:Ae.source},description:{story:"Playground - Interactive controls for all props",...(ze=(De=S.parameters)==null?void 0:De.docs)==null?void 0:ze.description}}};var Ge,Le,Fe,Ie,We;T.parameters={...T.parameters,docs:{...(Ge=T.parameters)==null?void 0:Ge.docs,source:{originalSource:`{
  render: () => {
    const [scenario, setScenario] = useState<'none' | 'auth' | 'network' | 'permission'>('none');
    const AuthFailureApp = () => {
      throw new Error('Authentication service failed to initialize');
    };
    return <div className="relative">
        {scenario === 'none' && <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
            <div className="bg-white rounded-xl shadow-2xl p-6 border-2 border-gray-200 max-w-md">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                Real-World Error Scenarios
              </h3>
              <p className="text-sm text-gray-600 mb-4">
                Simulate common errors in consent management applications
              </p>
              <div className="space-y-2">
                <Button variant="danger" onClick={() => setScenario('auth')} size="md" fullWidth>
                  Authentication Service Failure
                </Button>
                <Button variant="danger" onClick={() => setScenario('network')} size="md" fullWidth>
                  Network Connectivity Issue
                </Button>
                <Button variant="danger" onClick={() => setScenario('permission')} size="md" fullWidth>
                  Insufficient Permissions
                </Button>
              </div>
            </div>
          </div>}

        <GlobalErrorBoundary appName="Consent Manager" supportEmail="support@consent-manager.com" supportPhone="1-800-CONSENT" supportUrl="https://docs.consent-manager.com" showDetails={false} recoveryStrategies={['reload', 'navigate', 'contact-support']}>
          {scenario === 'auth' && <AuthFailureApp />}
          {scenario === 'network' && <NetworkErrorApp shouldThrow={true} />}
          {scenario === 'permission' && <PermissionErrorApp shouldThrow={true} />}
          {scenario === 'none' && <BuggyApp shouldThrow={false} />}
        </GlobalErrorBoundary>
      </div>;
  }
}`,...(Fe=(Le=T.parameters)==null?void 0:Le.docs)==null?void 0:Fe.source},description:{story:"Real-world consent management examples",...(We=(Ie=T.parameters)==null?void 0:Ie.docs)==null?void 0:We.description}}};var Pe,Re,Me,Ue,qe;C.parameters={...C.parameters,docs:{...(Pe=C.parameters)==null?void 0:Pe.docs,source:{originalSource:`{
  render: () => <div className="relative">
      <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
        <div className="bg-white rounded-xl shadow-2xl p-6 border-2 border-blue-200 max-w-lg">
          <h3 className="text-lg font-semibold text-blue-900 mb-3">
            Accessibility Features
          </h3>
          <ul className="text-sm text-blue-800 space-y-2 list-disc list-inside">
            <li>role="alert" for immediate screen reader announcement</li>
            <li>aria-live="assertive" for critical errors</li>
            <li>Full keyboard navigation support</li>
            <li>WCAG 2.1 AA compliant color contrast</li>
            <li>Clear, semantic heading structure</li>
            <li>Focus indicators on all interactive elements</li>
          </ul>
        </div>
      </div>

      <GlobalErrorBoundary appName="Identity Broker" supportEmail="accessibility@identity-broker.com" showDetails={true}>
        <BrokenApp />
      </GlobalErrorBoundary>
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
        }, {
          id: 'landmark-one-main',
          enabled: false
        } // Error pages don't need main landmarks
        ]
      }
    }
  }
}`,...(Me=(Re=C.parameters)==null?void 0:Re.docs)==null?void 0:Me.source},description:{story:"Accessibility features demonstration",...(qe=(Ue=C.parameters)==null?void 0:Ue.docs)==null?void 0:qe.description}}};const dr=["Default","WithFullPageError","WithFooterActions","WithErrorReporting","WithRecovery","GracefulDegradation","ErrorCategories","DevelopmentMode","ProductionMode","CompleteExample","Playground","RealWorldScenarios","Accessibility"];export{C as Accessibility,B as CompleteExample,f as Default,j as DevelopmentMode,k as ErrorCategories,N as GracefulDegradation,S as Playground,E as ProductionMode,T as RealWorldScenarios,y as WithErrorReporting,v as WithFooterActions,b as WithFullPageError,w as WithRecovery,dr as __namedExportsOrder,lr as default};
