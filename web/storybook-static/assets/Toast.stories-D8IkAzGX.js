import{j as t}from"./jsx-runtime-BYYWji4R.js";import{R as Bt,r as i}from"./index-ClcD9ViR.js";import{c as k,a as Ct}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const zt=Ct("relative flex gap-3 rounded-lg border p-4 shadow-lg-premium max-w-md w-full transition-all duration-200",{variants:{variant:{info:"bg-info-light text-info-dark border-info-primary/30",success:"bg-success-light text-success-dark border-success-primary/30",warning:"bg-warning-light text-warning-dark border-warning-primary/30",error:"bg-error-light text-error-dark border-error-primary/30"},position:{"top-right":"fixed top-4 right-4","top-left":"fixed top-4 left-4","bottom-right":"fixed bottom-4 right-4","bottom-left":"fixed bottom-4 left-4"}},defaultVariants:{variant:"info",position:"top-right"}}),Ft=Ct("flex-shrink-0 w-5 h-5",{variants:{variant:{info:"text-info-primary",success:"text-success-primary",warning:"text-warning-primary",error:"text-error-primary"}},defaultVariants:{variant:"info"}}),Gt=()=>t.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:t.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"})}),Qt=()=>t.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:t.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"})}),Ut=()=>t.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:t.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"})}),_t=()=>t.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:t.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"})}),Ht=()=>t.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:t.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M6 18L18 6M6 6l12 12"})}),o=Bt.forwardRef(({variant:s="info",icon:e,title:r,duration:d=5e3,onDismiss:n,action:c,position:a="top-right",hideIcon:l=!1,className:St,children:Dt,...At},It)=>{const[Rt,Pt]=i.useState(!0),[Et,qt]=i.useState(!1),[Wt,Vt]=i.useState(!0),Yt=()=>{switch(s){case"success":return t.jsx(Qt,{});case"warning":return t.jsx(Ut,{});case"error":return t.jsx(_t,{});case"info":default:return t.jsx(Gt,{})}},N=()=>{qt(!0),setTimeout(()=>{Pt(!1),n==null||n()},200)};if(i.useEffect(()=>{if(d>0){const T=setTimeout(()=>{N()},d);return()=>clearTimeout(T)}},[d]),i.useEffect(()=>{const T=setTimeout(()=>{Vt(!1)},200);return()=>clearTimeout(T)},[]),!Rt)return null;const Lt=s==="error"?"alert":"status",Mt=s==="error"?"assertive":"polite",Ot=()=>{if(Et)return"opacity-0 scale-95";if(Wt){if(a.includes("right"))return"translate-x-full opacity-0";if(a.includes("left"))return"-translate-x-full opacity-0"}return""};return t.jsxs("div",{ref:It,role:Lt,"aria-live":Mt,className:k(zt({variant:s,position:a}),Ot(),St),...At,children:[!l&&t.jsx("div",{className:Ft({variant:s}),children:e||Yt()}),t.jsxs("div",{className:"flex-1 min-w-0",children:[r&&t.jsx("h4",{className:"text-sm font-semibold mb-1 leading-tight",children:r}),t.jsx("div",{className:"text-sm leading-relaxed",children:Dt})]}),c&&t.jsx("button",{type:"button",onClick:c.onClick,className:k("flex-shrink-0 px-3 py-1.5 text-sm font-medium rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2",s==="info"&&"text-info-primary hover:bg-info-primary/10 focus:ring-info-primary",s==="success"&&"text-success-primary hover:bg-success-primary/10 focus:ring-success-primary",s==="warning"&&"text-warning-primary hover:bg-warning-primary/10 focus:ring-warning-primary",s==="error"&&"text-error-primary hover:bg-error-primary/10 focus:ring-error-primary"),children:c.label}),t.jsx("button",{type:"button",onClick:N,"aria-label":"Dismiss notification",className:k("flex-shrink-0 inline-flex items-center justify-center w-8 h-8 rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2",s==="info"&&"text-info-primary/70 hover:text-info-primary hover:bg-info-primary/10 focus:ring-info-primary",s==="success"&&"text-success-primary/70 hover:text-success-primary hover:bg-success-primary/10 focus:ring-success-primary",s==="warning"&&"text-warning-primary/70 hover:text-warning-primary hover:bg-warning-primary/10 focus:ring-warning-primary",s==="error"&&"text-error-primary/70 hover:text-error-primary hover:bg-error-primary/10 focus:ring-error-primary"),children:t.jsx(Ht,{})})]})});o.displayName="Toast";o.__docgenInfo={description:`Toast component for displaying temporary notification messages.
Supports multiple variants, icons, titles, actions, and auto-dismiss.

@example
\`\`\`tsx
<Toast variant="success" duration={5000}>
  Your changes have been saved successfully.
</Toast>

<Toast variant="warning" title="Warning" position="top-right">
  Please review the following information before proceeding.
</Toast>

<Toast
  variant="error"
  title="Error"
  action={{ label: "Retry", onClick: handleRetry }}
  duration={0}
>
  Failed to process your request.
</Toast>
\`\`\``,methods:[],displayName:"Toast",props:{variant:{required:!1,tsType:{name:"union",raw:"'info' | 'success' | 'warning' | 'error'",elements:[{name:"literal",value:"'info'"},{name:"literal",value:"'success'"},{name:"literal",value:"'warning'"},{name:"literal",value:"'error'"}]},description:"Toast variant based on message severity",defaultValue:{value:"'info'",computed:!1}},icon:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Optional icon to display (overrides default icon)"},title:{required:!1,tsType:{name:"string"},description:"Optional title text (bold)"},duration:{required:!1,tsType:{name:"number"},description:"Auto-dismiss duration in milliseconds (0 = no auto-dismiss)",defaultValue:{value:"5000",computed:!1}},onDismiss:{required:!1,tsType:{name:"signature",type:"function",raw:"() => void",signature:{arguments:[],return:{name:"void"}}},description:"Callback when toast is dismissed"},action:{required:!1,tsType:{name:"signature",type:"object",raw:`{
  label: string;
  onClick: () => void;
}`,signature:{properties:[{key:"label",value:{name:"string",required:!0}},{key:"onClick",value:{name:"signature",type:"function",raw:"() => void",signature:{arguments:[],return:{name:"void"}},required:!0}}]}},description:"Optional action button"},position:{required:!1,tsType:{name:"union",raw:"'top-right' | 'top-left' | 'bottom-right' | 'bottom-left'",elements:[{name:"literal",value:"'top-right'"},{name:"literal",value:"'top-left'"},{name:"literal",value:"'bottom-right'"},{name:"literal",value:"'bottom-left'"}]},description:"Toast position on screen",defaultValue:{value:"'top-right'",computed:!1}},hideIcon:{required:!1,tsType:{name:"boolean"},description:"Hide default icon",defaultValue:{value:"false",computed:!1}}},composes:["Omit"]};const Zt={title:"Design System/Feedback/Toast",component:o,parameters:{layout:"padded"},tags:["autodocs"],argTypes:{variant:{control:"select",options:["info","success","warning","error"],description:"Visual style variant based on message severity"},title:{control:"text",description:"Optional bold title text"},duration:{control:"number",description:"Auto-dismiss duration in milliseconds (0 = no auto-dismiss)"},position:{control:"select",options:["top-right","top-left","bottom-right","bottom-left"],description:"Toast position on screen"},hideIcon:{control:"boolean",description:"Hide the default icon"},children:{control:"text",description:"Toast message content"}}},u={args:{variant:"info",duration:0,children:"This is an informational notification to provide quick feedback."}},m={render:()=>{const[s,e]=i.useState({info:!0,success:!0,warning:!0,error:!0});return t.jsxs("div",{className:"space-y-4 max-w-2xl",children:[s.info&&t.jsx(o,{variant:"info",duration:0,onDismiss:()=>e({...s,info:!1}),children:"This is an informational toast. It provides helpful context or additional information."}),s.success&&t.jsx(o,{variant:"success",duration:0,onDismiss:()=>e({...s,success:!1}),children:"Success! Your changes have been saved and will take effect immediately."}),s.warning&&t.jsx(o,{variant:"warning",duration:0,onDismiss:()=>e({...s,warning:!1}),children:"Warning: This action cannot be undone. Please review your selection carefully."}),s.error&&t.jsx(o,{variant:"error",duration:0,onDismiss:()=>e({...s,error:!1}),children:"Error: Unable to process your request. Please check your connection and try again."}),!s.info&&!s.success&&!s.warning&&!s.error&&t.jsxs("div",{className:"text-center p-8 border border-dashed border-gray-300 rounded-lg",children:[t.jsx("p",{className:"text-gray-600 mb-4",children:"All toasts dismissed"}),t.jsx("button",{onClick:()=>e({info:!0,success:!0,warning:!0,error:!0}),className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors",children:"Show All Toasts"})]})]})},args:{variant:"info",children:"Toast content"}},p={render:()=>{const[s,e]=i.useState({info:!0,success:!0,warning:!0,error:!0});return t.jsxs("div",{className:"space-y-4 max-w-2xl",children:[s.info&&t.jsx(o,{variant:"info",duration:0,onDismiss:()=>e({...s,info:!1}),children:"Each toast variant comes with a default icon that matches its semantic meaning."}),s.success&&t.jsx(o,{variant:"success",duration:0,onDismiss:()=>e({...s,success:!1}),children:"The success icon indicates a positive outcome or completed action."}),s.warning&&t.jsx(o,{variant:"warning",duration:0,onDismiss:()=>e({...s,warning:!1}),children:"The warning icon draws attention to important caution messages."}),s.error&&t.jsx(o,{variant:"error",duration:0,onDismiss:()=>e({...s,error:!1}),children:"The error icon clearly indicates a problem that needs attention."})]})},args:{variant:"info",children:"Toast with icon"}},h={render:()=>{const[s,e]=i.useState({action1:!0,action2:!0,action3:!0});return t.jsxs("div",{className:"space-y-4 max-w-2xl",children:[s.action1&&t.jsx(o,{variant:"warning",duration:0,onDismiss:()=>e({...s,action1:!1}),action:{label:"Review",onClick:()=>alert("Action clicked!")},children:"Your subscription will expire in 3 days. Review your plan to continue using all features."}),s.action2&&t.jsx(o,{variant:"error",duration:0,onDismiss:()=>e({...s,action2:!1}),action:{label:"Retry",onClick:()=>alert("Retrying...")},children:"Payment processing failed. Please verify your payment method and try again."}),s.action3&&t.jsx(o,{variant:"success",duration:0,onDismiss:()=>e({...s,action3:!1}),action:{label:"View",onClick:()=>alert("Opening...")},children:"New features are available! Check out our latest updates."}),!s.action1&&!s.action2&&!s.action3&&t.jsxs("div",{className:"text-center p-8 border border-dashed border-gray-300 rounded-lg",children:[t.jsx("p",{className:"text-gray-600 mb-4",children:"All toasts dismissed"}),t.jsx("button",{onClick:()=>e({action1:!0,action2:!0,action3:!0}),className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors",children:"Show All Toasts"})]})]})},args:{variant:"info",children:"Toast with actions"}},g={render:()=>{const[s,e]=i.useState({toast1:!1,toast2:!1,toast3:!1});return t.jsxs("div",{className:"space-y-4 max-w-2xl",children:[t.jsxs("div",{className:"p-6 border border-dashed border-gray-300 rounded-lg",children:[t.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Auto-Dismiss Demo"}),t.jsx("p",{className:"text-sm text-gray-600 mb-4",children:"Click the buttons below to show toasts that auto-dismiss after 5 seconds. You can also manually dismiss them before the timer ends."}),t.jsxs("div",{className:"flex gap-2 flex-wrap",children:[t.jsx("button",{onClick:()=>e({...s,toast1:!0}),className:"px-4 py-2 bg-info-primary text-white rounded-md hover:bg-info-dark transition-colors",children:"Show Info Toast"}),t.jsx("button",{onClick:()=>e({...s,toast2:!0}),className:"px-4 py-2 bg-success-primary text-white rounded-md hover:bg-success-dark transition-colors",children:"Show Success Toast"}),t.jsx("button",{onClick:()=>e({...s,toast3:!0}),className:"px-4 py-2 bg-warning-primary text-white rounded-md hover:bg-warning-dark transition-colors",children:"Show Warning Toast"})]})]}),s.toast1&&t.jsx(o,{variant:"info",duration:5e3,onDismiss:()=>e({...s,toast1:!1}),children:"This toast will automatically dismiss after 5 seconds."}),s.toast2&&t.jsx(o,{variant:"success",duration:5e3,onDismiss:()=>e({...s,toast2:!1}),children:"Changes saved! This notification will disappear in 5 seconds."}),s.toast3&&t.jsx(o,{variant:"warning",title:"Auto-Dismiss",duration:5e3,onDismiss:()=>e({...s,toast3:!1}),children:"You can still close this manually before the timer ends."})]})},args:{variant:"info",duration:5e3,children:"This toast will auto-dismiss"}},x={render:()=>{const[s,e]=i.useState(4),[r,d]=i.useState([{id:1,variant:"info",message:"First notification in the stack"},{id:2,variant:"success",message:"Successfully saved your changes"},{id:3,variant:"warning",message:"Connection is unstable"}]),n=a=>{const l={info:"New information available",success:"Operation completed successfully",warning:"Please review this warning",error:"An error occurred"};d([...r,{id:s,variant:a,message:l[a]}]),e(s+1)},c=a=>{d(r.filter(l=>l.id!==a))};return t.jsxs("div",{className:"space-y-4",children:[t.jsxs("div",{className:"p-6 border border-dashed border-gray-300 rounded-lg max-w-2xl",children:[t.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Toast Stack Manager"}),t.jsx("p",{className:"text-sm text-gray-600 mb-4",children:"Add multiple toasts to see how they stack. Toasts appear at the top of the stack and stack downward."}),t.jsxs("div",{className:"flex gap-2 flex-wrap",children:[t.jsx("button",{onClick:()=>n("info"),className:"px-3 py-2 text-sm bg-info-primary text-white rounded-md hover:bg-info-dark transition-colors",children:"Add Info"}),t.jsx("button",{onClick:()=>n("success"),className:"px-3 py-2 text-sm bg-success-primary text-white rounded-md hover:bg-success-dark transition-colors",children:"Add Success"}),t.jsx("button",{onClick:()=>n("warning"),className:"px-3 py-2 text-sm bg-warning-primary text-white rounded-md hover:bg-warning-dark transition-colors",children:"Add Warning"}),t.jsx("button",{onClick:()=>n("error"),className:"px-3 py-2 text-sm bg-error-primary text-white rounded-md hover:bg-error-dark transition-colors",children:"Add Error"}),t.jsx("button",{onClick:()=>d([]),className:"px-3 py-2 text-sm bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors",children:"Clear All"})]}),t.jsxs("p",{className:"text-xs text-gray-500 mt-3",children:["Active toasts: ",r.length]})]}),t.jsx("div",{className:"space-y-2 max-w-2xl",children:r.map(a=>t.jsx(o,{variant:a.variant,duration:0,position:void 0,onDismiss:()=>c(a.id),children:a.message},a.id))})]})},args:{variant:"info",children:"Stacked toast"}},f={render:()=>{const[s,e]=i.useState(null);return t.jsxs("div",{className:"h-96 relative border border-dashed border-gray-300 rounded-lg",children:[t.jsx("div",{className:"absolute inset-0 flex items-center justify-center",children:t.jsxs("div",{className:"text-center",children:[t.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Choose a Position"}),t.jsx("p",{className:"text-sm text-gray-600 mb-4 max-w-md",children:"Click a button to show a toast in that corner. Toasts will appear with a slide-in animation from their respective edge."}),t.jsxs("div",{className:"grid grid-cols-2 gap-3",children:[t.jsx("button",{onClick:()=>e("top-left"),className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors",children:"Top Left"}),t.jsx("button",{onClick:()=>e("top-right"),className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors",children:"Top Right"}),t.jsx("button",{onClick:()=>e("bottom-left"),className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors",children:"Bottom Left"}),t.jsx("button",{onClick:()=>e("bottom-right"),className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors",children:"Bottom Right"})]})]})}),s&&t.jsxs(o,{variant:"success",position:s,duration:0,onDismiss:()=>e(null),title:`Toast at ${s}`,children:["This toast appears in the ",s.replace("-"," ")," corner of the screen."]})]})},args:{variant:"info",position:"top-right",children:"Positioned toast"}},v={render:()=>{const[s,e]=i.useState(null),r=(d,n)=>{e(n),setTimeout(()=>e(null),d)};return t.jsxs("div",{className:"space-y-4 max-w-2xl",children:[t.jsxs("div",{className:"p-6 border border-dashed border-gray-300 rounded-lg",children:[t.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-3",children:"Duration Options"}),t.jsx("p",{className:"text-sm text-gray-600 mb-4",children:"Configure how long toasts remain visible before auto-dismissing. Choose from quick (3s), standard (5s), or extended (10s) durations."}),t.jsxs("div",{className:"flex gap-2 flex-wrap",children:[t.jsx("button",{onClick:()=>r(3e3,"3s"),disabled:s==="3s",className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed",children:"Quick (3s)"}),t.jsx("button",{onClick:()=>r(5e3,"5s"),disabled:s==="5s",className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed",children:"Standard (5s)"}),t.jsx("button",{onClick:()=>r(1e4,"10s"),disabled:s==="10s",className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed",children:"Extended (10s)"}),t.jsx("button",{onClick:()=>{e("persistent")},disabled:s==="persistent",className:"px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed",children:"No Auto-Dismiss"})]})]}),s==="3s"&&t.jsx(o,{variant:"info",duration:3e3,onDismiss:()=>e(null),children:"Quick toast - dismisses after 3 seconds"}),s==="5s"&&t.jsx(o,{variant:"success",duration:5e3,onDismiss:()=>e(null),children:"Standard toast - dismisses after 5 seconds"}),s==="10s"&&t.jsx(o,{variant:"warning",duration:1e4,onDismiss:()=>e(null),children:"Extended toast - dismisses after 10 seconds"}),s==="persistent"&&t.jsx(o,{variant:"error",duration:0,onDismiss:()=>e(null),title:"Persistent Toast",children:"This toast requires manual dismissal - no auto-dismiss timer"})]})},args:{variant:"info",duration:5e3,children:"Toast with custom duration"}},b={render:s=>{const[e,r]=i.useState(!0);return e?t.jsx(o,{...s,position:void 0,onDismiss:()=>{var d;r(!1),(d=s.onDismiss)==null||d.call(s)}}):t.jsxs("div",{className:"text-center p-8 border border-dashed border-gray-300 rounded-lg",children:[t.jsx("p",{className:"text-gray-600 mb-4",children:"Toast dismissed"}),t.jsx("button",{onClick:()=>r(!0),className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors",children:"Show Toast"})]})},args:{variant:"info",title:"Toast Title",children:"This is the toast message content. You can customize all properties using the controls below.",duration:0,hideIcon:!1,action:void 0}},y={render:()=>{const[s,e]=i.useState({toast1:!0,toast2:!0,toast3:!0,toast4:!0});return t.jsxs("div",{className:"space-y-4 max-w-2xl",children:[s.toast1&&t.jsx(o,{variant:"info",title:"Information",duration:0,onDismiss:()=>e({...s,toast1:!1}),children:"Titles help organize complex toasts and provide clear context for the message."}),s.toast2&&t.jsx(o,{variant:"success",title:"Access Granted",duration:0,onDismiss:()=>e({...s,toast2:!1}),children:'You now have editor permissions for the "Marketing Assets" workspace.'}),s.toast3&&t.jsx(o,{variant:"warning",title:"Action Required",duration:0,onDismiss:()=>e({...s,toast3:!1}),children:"Your account requires two-factor authentication within 7 days."}),s.toast4&&t.jsx(o,{variant:"error",title:"Authentication Failed",duration:0,onDismiss:()=>e({...s,toast4:!1}),children:"Your session may have expired. Please sign in again."}),!s.toast1&&!s.toast2&&!s.toast3&&!s.toast4&&t.jsxs("div",{className:"text-center p-8 border border-dashed border-gray-300 rounded-lg",children:[t.jsx("p",{className:"text-gray-600 mb-4",children:"All toasts dismissed"}),t.jsx("button",{onClick:()=>e({toast1:!0,toast2:!0,toast3:!0,toast4:!0}),className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors",children:"Show All Toasts"})]})]})},args:{variant:"info",title:"Toast Title",children:"Toast description goes here"}},w={render:()=>{const[s,e]=i.useState([]),[r,d]=i.useState(1),n=a=>{e([...s,{id:r,type:a}]),d(r+1)},c=a=>{e(s.filter(l=>l.id!==a))};return t.jsx("div",{className:"space-y-6 max-w-3xl",children:t.jsxs("div",{children:[t.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-4",children:"Consent Management Scenarios"}),t.jsxs("div",{className:"p-6 border border-dashed border-gray-300 rounded-lg mb-4",children:[t.jsx("p",{className:"text-sm text-gray-600 mb-4",children:"Simulate common consent management notifications:"}),t.jsxs("div",{className:"flex gap-2 flex-wrap",children:[t.jsx("button",{onClick:()=>n("granted"),className:"px-3 py-2 text-sm bg-success-primary text-white rounded-md hover:bg-success-dark transition-colors",children:"Grant Consent"}),t.jsx("button",{onClick:()=>n("revoked"),className:"px-3 py-2 text-sm bg-warning-primary text-white rounded-md hover:bg-warning-dark transition-colors",children:"Revoke Consent"}),t.jsx("button",{onClick:()=>n("expired"),className:"px-3 py-2 text-sm bg-info-primary text-white rounded-md hover:bg-info-dark transition-colors",children:"Consent Expired"}),t.jsx("button",{onClick:()=>n("error"),className:"px-3 py-2 text-sm bg-error-primary text-white rounded-md hover:bg-error-dark transition-colors",children:"Security Alert"})]})]}),t.jsx("div",{className:"space-y-2",children:s.map(a=>a.type==="granted"?t.jsx(o,{variant:"success",title:"Consent Granted",duration:5e3,position:void 0,onDismiss:()=>c(a.id),children:"You've successfully granted access to your profile data. This permission is valid for 30 days."},a.id):a.type==="revoked"?t.jsx(o,{variant:"warning",title:"Consent Revoked",duration:5e3,position:void 0,onDismiss:()=>c(a.id),children:"Access permissions have been revoked. The application will no longer have access to your data."},a.id):a.type==="expired"?t.jsx(o,{variant:"info",title:"Consent Expired",duration:0,position:void 0,onDismiss:()=>c(a.id),action:{label:"Renew",onClick:()=>{alert("Renewing consent..."),c(a.id)}},children:'Your consent for "Marketing Platform" has expired. Renew to continue sharing data.'},a.id):a.type==="error"?t.jsx(o,{variant:"error",title:"Security Alert",duration:0,position:void 0,onDismiss:()=>c(a.id),action:{label:"Review",onClick:()=>alert("Opening security settings...")},children:"Unusual activity detected. Please review your recent consent grants."},a.id):null)})]})})},args:{variant:"info",children:"Real-world example"}};var j,C,S,D,A;u.parameters={...u.parameters,docs:{...(j=u.parameters)==null?void 0:j.docs,source:{originalSource:`{
  args: {
    variant: 'info',
    duration: 0,
    // Disable auto-dismiss for story
    children: 'This is an informational notification to provide quick feedback.'
  }
}`,...(S=(C=u.parameters)==null?void 0:C.docs)==null?void 0:S.source},description:{story:"Default toast with info variant",...(A=(D=u.parameters)==null?void 0:D.docs)==null?void 0:A.description}}};var I,R,P,E,q;m.parameters={...m.parameters,docs:{...(I=m.parameters)==null?void 0:I.docs,source:{originalSource:`{
  render: () => {
    const [toasts, setToasts] = useState({
      info: true,
      success: true,
      warning: true,
      error: true
    });
    return <div className="space-y-4 max-w-2xl">
        {toasts.info && <Toast variant="info" duration={0} onDismiss={() => setToasts({
        ...toasts,
        info: false
      })}>
            This is an informational toast. It provides helpful context or
            additional information.
          </Toast>}

        {toasts.success && <Toast variant="success" duration={0} onDismiss={() => setToasts({
        ...toasts,
        success: false
      })}>
            Success! Your changes have been saved and will take effect
            immediately.
          </Toast>}

        {toasts.warning && <Toast variant="warning" duration={0} onDismiss={() => setToasts({
        ...toasts,
        warning: false
      })}>
            Warning: This action cannot be undone. Please review your selection
            carefully.
          </Toast>}

        {toasts.error && <Toast variant="error" duration={0} onDismiss={() => setToasts({
        ...toasts,
        error: false
      })}>
            Error: Unable to process your request. Please check your connection
            and try again.
          </Toast>}

        {!toasts.info && !toasts.success && !toasts.warning && !toasts.error && <div className="text-center p-8 border border-dashed border-gray-300 rounded-lg">
            <p className="text-gray-600 mb-4">All toasts dismissed</p>
            <button onClick={() => setToasts({
          info: true,
          success: true,
          warning: true,
          error: true
        })} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors">
              Show All Toasts
            </button>
          </div>}
      </div>;
  },
  args: {
    variant: 'info',
    children: 'Toast content'
  }
}`,...(P=(R=m.parameters)==null?void 0:R.docs)==null?void 0:P.source},description:{story:"All toast variants displayed together",...(q=(E=m.parameters)==null?void 0:E.docs)==null?void 0:q.description}}};var W,V,Y,L,M;p.parameters={...p.parameters,docs:{...(W=p.parameters)==null?void 0:W.docs,source:{originalSource:`{
  render: () => {
    const [toasts, setToasts] = useState({
      info: true,
      success: true,
      warning: true,
      error: true
    });
    return <div className="space-y-4 max-w-2xl">
        {toasts.info && <Toast variant="info" duration={0} onDismiss={() => setToasts({
        ...toasts,
        info: false
      })}>
            Each toast variant comes with a default icon that matches its
            semantic meaning.
          </Toast>}

        {toasts.success && <Toast variant="success" duration={0} onDismiss={() => setToasts({
        ...toasts,
        success: false
      })}>
            The success icon indicates a positive outcome or completed action.
          </Toast>}

        {toasts.warning && <Toast variant="warning" duration={0} onDismiss={() => setToasts({
        ...toasts,
        warning: false
      })}>
            The warning icon draws attention to important caution messages.
          </Toast>}

        {toasts.error && <Toast variant="error" duration={0} onDismiss={() => setToasts({
        ...toasts,
        error: false
      })}>
            The error icon clearly indicates a problem that needs attention.
          </Toast>}
      </div>;
  },
  args: {
    variant: 'info',
    children: 'Toast with icon'
  }
}`,...(Y=(V=p.parameters)==null?void 0:V.docs)==null?void 0:Y.source},description:{story:"Toasts with leading icons (default behavior)",...(M=(L=p.parameters)==null?void 0:L.docs)==null?void 0:M.description}}};var O,B,z,F,G;h.parameters={...h.parameters,docs:{...(O=h.parameters)==null?void 0:O.docs,source:{originalSource:`{
  render: () => {
    const [toasts, setToasts] = useState({
      action1: true,
      action2: true,
      action3: true
    });
    return <div className="space-y-4 max-w-2xl">
        {toasts.action1 && <Toast variant="warning" duration={0} onDismiss={() => setToasts({
        ...toasts,
        action1: false
      })} action={{
        label: 'Review',
        onClick: () => alert('Action clicked!')
      }}>
            Your subscription will expire in 3 days. Review your plan to
            continue using all features.
          </Toast>}

        {toasts.action2 && <Toast variant="error" duration={0} onDismiss={() => setToasts({
        ...toasts,
        action2: false
      })} action={{
        label: 'Retry',
        onClick: () => alert('Retrying...')
      }}>
            Payment processing failed. Please verify your payment method and try
            again.
          </Toast>}

        {toasts.action3 && <Toast variant="success" duration={0} onDismiss={() => setToasts({
        ...toasts,
        action3: false
      })} action={{
        label: 'View',
        onClick: () => alert('Opening...')
      }}>
            New features are available! Check out our latest updates.
          </Toast>}

        {!toasts.action1 && !toasts.action2 && !toasts.action3 && <div className="text-center p-8 border border-dashed border-gray-300 rounded-lg">
            <p className="text-gray-600 mb-4">All toasts dismissed</p>
            <button onClick={() => setToasts({
          action1: true,
          action2: true,
          action3: true
        })} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors">
              Show All Toasts
            </button>
          </div>}
      </div>;
  },
  args: {
    variant: 'info',
    children: 'Toast with actions'
  }
}`,...(z=(B=h.parameters)==null?void 0:B.docs)==null?void 0:z.source},description:{story:"Toasts with action buttons",...(G=(F=h.parameters)==null?void 0:F.docs)==null?void 0:G.description}}};var Q,U,_,H,$;g.parameters={...g.parameters,docs:{...(Q=g.parameters)==null?void 0:Q.docs,source:{originalSource:`{
  render: () => {
    const [toasts, setToasts] = useState({
      toast1: false,
      toast2: false,
      toast3: false
    });
    return <div className="space-y-4 max-w-2xl">
        <div className="p-6 border border-dashed border-gray-300 rounded-lg">
          <h4 className="text-sm font-semibold text-gray-900 mb-3">
            Auto-Dismiss Demo
          </h4>
          <p className="text-sm text-gray-600 mb-4">
            Click the buttons below to show toasts that auto-dismiss after 5
            seconds. You can also manually dismiss them before the timer ends.
          </p>
          <div className="flex gap-2 flex-wrap">
            <button onClick={() => setToasts({
            ...toasts,
            toast1: true
          })} className="px-4 py-2 bg-info-primary text-white rounded-md hover:bg-info-dark transition-colors">
              Show Info Toast
            </button>
            <button onClick={() => setToasts({
            ...toasts,
            toast2: true
          })} className="px-4 py-2 bg-success-primary text-white rounded-md hover:bg-success-dark transition-colors">
              Show Success Toast
            </button>
            <button onClick={() => setToasts({
            ...toasts,
            toast3: true
          })} className="px-4 py-2 bg-warning-primary text-white rounded-md hover:bg-warning-dark transition-colors">
              Show Warning Toast
            </button>
          </div>
        </div>

        {toasts.toast1 && <Toast variant="info" duration={5000} onDismiss={() => setToasts({
        ...toasts,
        toast1: false
      })}>
            This toast will automatically dismiss after 5 seconds.
          </Toast>}

        {toasts.toast2 && <Toast variant="success" duration={5000} onDismiss={() => setToasts({
        ...toasts,
        toast2: false
      })}>
            Changes saved! This notification will disappear in 5 seconds.
          </Toast>}

        {toasts.toast3 && <Toast variant="warning" title="Auto-Dismiss" duration={5000} onDismiss={() => setToasts({
        ...toasts,
        toast3: false
      })}>
            You can still close this manually before the timer ends.
          </Toast>}
      </div>;
  },
  args: {
    variant: 'info',
    duration: 5000,
    children: 'This toast will auto-dismiss'
  }
}`,...(_=(U=g.parameters)==null?void 0:U.docs)==null?void 0:_.source},description:{story:"Toasts with auto-dismiss timer (5 seconds)",...($=(H=g.parameters)==null?void 0:H.docs)==null?void 0:$.description}}};var J,K,X,Z,tt;x.parameters={...x.parameters,docs:{...(J=x.parameters)==null?void 0:J.docs,source:{originalSource:`{
  render: () => {
    const [nextId, setNextId] = useState(4);
    const [toasts, setToasts] = useState([{
      id: 1,
      variant: 'info' as const,
      message: 'First notification in the stack'
    }, {
      id: 2,
      variant: 'success' as const,
      message: 'Successfully saved your changes'
    }, {
      id: 3,
      variant: 'warning' as const,
      message: 'Connection is unstable'
    }]);
    const addToast = (variant: 'info' | 'success' | 'warning' | 'error') => {
      const messages = {
        info: 'New information available',
        success: 'Operation completed successfully',
        warning: 'Please review this warning',
        error: 'An error occurred'
      };
      setToasts([...toasts, {
        id: nextId,
        variant,
        message: messages[variant]
      }]);
      setNextId(nextId + 1);
    };
    const removeToast = (id: number) => {
      setToasts(toasts.filter(t => t.id !== id));
    };
    return <div className="space-y-4">
        <div className="p-6 border border-dashed border-gray-300 rounded-lg max-w-2xl">
          <h4 className="text-sm font-semibold text-gray-900 mb-3">
            Toast Stack Manager
          </h4>
          <p className="text-sm text-gray-600 mb-4">
            Add multiple toasts to see how they stack. Toasts appear at the top
            of the stack and stack downward.
          </p>
          <div className="flex gap-2 flex-wrap">
            <button onClick={() => addToast('info')} className="px-3 py-2 text-sm bg-info-primary text-white rounded-md hover:bg-info-dark transition-colors">
              Add Info
            </button>
            <button onClick={() => addToast('success')} className="px-3 py-2 text-sm bg-success-primary text-white rounded-md hover:bg-success-dark transition-colors">
              Add Success
            </button>
            <button onClick={() => addToast('warning')} className="px-3 py-2 text-sm bg-warning-primary text-white rounded-md hover:bg-warning-dark transition-colors">
              Add Warning
            </button>
            <button onClick={() => addToast('error')} className="px-3 py-2 text-sm bg-error-primary text-white rounded-md hover:bg-error-dark transition-colors">
              Add Error
            </button>
            <button onClick={() => setToasts([])} className="px-3 py-2 text-sm bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors">
              Clear All
            </button>
          </div>
          <p className="text-xs text-gray-500 mt-3">
            Active toasts: {toasts.length}
          </p>
        </div>

        <div className="space-y-2 max-w-2xl">
          {toasts.map(toast => <Toast key={toast.id} variant={toast.variant} duration={0} position={undefined as any} // Remove fixed positioning for stacking demo
        onDismiss={() => removeToast(toast.id)}>
              {toast.message}
            </Toast>)}
        </div>
      </div>;
  },
  args: {
    variant: 'info',
    children: 'Stacked toast'
  }
}`,...(X=(K=x.parameters)==null?void 0:K.docs)==null?void 0:X.source},description:{story:"Multiple toasts stacked together",...(tt=(Z=x.parameters)==null?void 0:Z.docs)==null?void 0:tt.description}}};var st,et,ot,at,rt;f.parameters={...f.parameters,docs:{...(st=f.parameters)==null?void 0:st.docs,source:{originalSource:`{
  render: () => {
    const [activePosition, setActivePosition] = useState<'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' | null>(null);
    return <div className="h-96 relative border border-dashed border-gray-300 rounded-lg">
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="text-center">
            <h4 className="text-sm font-semibold text-gray-900 mb-3">
              Choose a Position
            </h4>
            <p className="text-sm text-gray-600 mb-4 max-w-md">
              Click a button to show a toast in that corner. Toasts will appear
              with a slide-in animation from their respective edge.
            </p>
            <div className="grid grid-cols-2 gap-3">
              <button onClick={() => setActivePosition('top-left')} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors">
                Top Left
              </button>
              <button onClick={() => setActivePosition('top-right')} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors">
                Top Right
              </button>
              <button onClick={() => setActivePosition('bottom-left')} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors">
                Bottom Left
              </button>
              <button onClick={() => setActivePosition('bottom-right')} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors">
                Bottom Right
              </button>
            </div>
          </div>
        </div>

        {activePosition && <Toast variant="success" position={activePosition} duration={0} onDismiss={() => setActivePosition(null)} title={\`Toast at \${activePosition}\`}>
            This toast appears in the {activePosition.replace('-', ' ')} corner
            of the screen.
          </Toast>}
      </div>;
  },
  args: {
    variant: 'info',
    position: 'top-right',
    children: 'Positioned toast'
  }
}`,...(ot=(et=f.parameters)==null?void 0:et.docs)==null?void 0:ot.source},description:{story:"Toast positioning variations",...(rt=(at=f.parameters)==null?void 0:at.docs)==null?void 0:rt.description}}};var it,nt,dt,ct,lt;v.parameters={...v.parameters,docs:{...(it=v.parameters)==null?void 0:it.docs,source:{originalSource:`{
  render: () => {
    const [activeToast, setActiveToast] = useState<string | null>(null);
    const showToast = (duration: number, label: string) => {
      setActiveToast(label);
      // Auto-reset for demo purposes
      setTimeout(() => setActiveToast(null), duration);
    };
    return <div className="space-y-4 max-w-2xl">
        <div className="p-6 border border-dashed border-gray-300 rounded-lg">
          <h4 className="text-sm font-semibold text-gray-900 mb-3">
            Duration Options
          </h4>
          <p className="text-sm text-gray-600 mb-4">
            Configure how long toasts remain visible before auto-dismissing.
            Choose from quick (3s), standard (5s), or extended (10s) durations.
          </p>
          <div className="flex gap-2 flex-wrap">
            <button onClick={() => showToast(3000, '3s')} disabled={activeToast === '3s'} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
              Quick (3s)
            </button>
            <button onClick={() => showToast(5000, '5s')} disabled={activeToast === '5s'} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
              Standard (5s)
            </button>
            <button onClick={() => showToast(10000, '10s')} disabled={activeToast === '10s'} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
              Extended (10s)
            </button>
            <button onClick={() => {
            setActiveToast('persistent');
          }} disabled={activeToast === 'persistent'} className="px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
              No Auto-Dismiss
            </button>
          </div>
        </div>

        {activeToast === '3s' && <Toast variant="info" duration={3000} onDismiss={() => setActiveToast(null)}>
            Quick toast - dismisses after 3 seconds
          </Toast>}

        {activeToast === '5s' && <Toast variant="success" duration={5000} onDismiss={() => setActiveToast(null)}>
            Standard toast - dismisses after 5 seconds
          </Toast>}

        {activeToast === '10s' && <Toast variant="warning" duration={10000} onDismiss={() => setActiveToast(null)}>
            Extended toast - dismisses after 10 seconds
          </Toast>}

        {activeToast === 'persistent' && <Toast variant="error" duration={0} onDismiss={() => setActiveToast(null)} title="Persistent Toast">
            This toast requires manual dismissal - no auto-dismiss timer
          </Toast>}
      </div>;
  },
  args: {
    variant: 'info',
    duration: 5000,
    children: 'Toast with custom duration'
  }
}`,...(dt=(nt=v.parameters)==null?void 0:nt.docs)==null?void 0:dt.source},description:{story:"Custom duration timing examples",...(lt=(ct=v.parameters)==null?void 0:ct.docs)==null?void 0:lt.description}}};var ut,mt,pt,ht,gt;b.parameters={...b.parameters,docs:{...(ut=b.parameters)==null?void 0:ut.docs,source:{originalSource:`{
  render: args => {
    const [isVisible, setIsVisible] = useState(true);
    if (!isVisible) {
      return <div className="text-center p-8 border border-dashed border-gray-300 rounded-lg">
          <p className="text-gray-600 mb-4">Toast dismissed</p>
          <button onClick={() => setIsVisible(true)} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors">
            Show Toast
          </button>
        </div>;
    }
    return <Toast {...args} position={undefined as any} // Remove fixed positioning for playground
    onDismiss={() => {
      setIsVisible(false);
      args.onDismiss?.();
    }} />;
  },
  args: {
    variant: 'info',
    title: 'Toast Title',
    children: 'This is the toast message content. You can customize all properties using the controls below.',
    duration: 0,
    hideIcon: false,
    action: undefined
  }
}`,...(pt=(mt=b.parameters)==null?void 0:mt.docs)==null?void 0:pt.source},description:{story:"Interactive playground with all controls",...(gt=(ht=b.parameters)==null?void 0:ht.docs)==null?void 0:gt.description}}};var xt,ft,vt,bt,yt;y.parameters={...y.parameters,docs:{...(xt=y.parameters)==null?void 0:xt.docs,source:{originalSource:`{
  render: () => {
    const [toasts, setToasts] = useState({
      toast1: true,
      toast2: true,
      toast3: true,
      toast4: true
    });
    return <div className="space-y-4 max-w-2xl">
        {toasts.toast1 && <Toast variant="info" title="Information" duration={0} onDismiss={() => setToasts({
        ...toasts,
        toast1: false
      })}>
            Titles help organize complex toasts and provide clear context for
            the message.
          </Toast>}

        {toasts.toast2 && <Toast variant="success" title="Access Granted" duration={0} onDismiss={() => setToasts({
        ...toasts,
        toast2: false
      })}>
            You now have editor permissions for the "Marketing Assets"
            workspace.
          </Toast>}

        {toasts.toast3 && <Toast variant="warning" title="Action Required" duration={0} onDismiss={() => setToasts({
        ...toasts,
        toast3: false
      })}>
            Your account requires two-factor authentication within 7 days.
          </Toast>}

        {toasts.toast4 && <Toast variant="error" title="Authentication Failed" duration={0} onDismiss={() => setToasts({
        ...toasts,
        toast4: false
      })}>
            Your session may have expired. Please sign in again.
          </Toast>}

        {!toasts.toast1 && !toasts.toast2 && !toasts.toast3 && !toasts.toast4 && <div className="text-center p-8 border border-dashed border-gray-300 rounded-lg">
            <p className="text-gray-600 mb-4">All toasts dismissed</p>
            <button onClick={() => setToasts({
          toast1: true,
          toast2: true,
          toast3: true,
          toast4: true
        })} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors">
              Show All Toasts
            </button>
          </div>}
      </div>;
  },
  args: {
    variant: 'info',
    title: 'Toast Title',
    children: 'Toast description goes here'
  }
}`,...(vt=(ft=y.parameters)==null?void 0:ft.docs)==null?void 0:vt.source},description:{story:"Toasts with titles",...(yt=(bt=y.parameters)==null?void 0:bt.docs)==null?void 0:yt.description}}};var wt,Tt,kt,Nt,jt;w.parameters={...w.parameters,docs:{...(wt=w.parameters)==null?void 0:wt.docs,source:{originalSource:`{
  render: () => {
    const [notifications, setNotifications] = useState<Array<{
      id: number;
      type: string;
    }>>([]);
    const [nextId, setNextId] = useState(1);
    const showNotification = (type: string) => {
      setNotifications([...notifications, {
        id: nextId,
        type
      }]);
      setNextId(nextId + 1);
    };
    const removeNotification = (id: number) => {
      setNotifications(notifications.filter(n => n.id !== id));
    };
    return <div className="space-y-6 max-w-3xl">
        <div>
          <h3 className="text-lg font-semibold text-gray-900 mb-4">
            Consent Management Scenarios
          </h3>

          <div className="p-6 border border-dashed border-gray-300 rounded-lg mb-4">
            <p className="text-sm text-gray-600 mb-4">
              Simulate common consent management notifications:
            </p>
            <div className="flex gap-2 flex-wrap">
              <button onClick={() => showNotification('granted')} className="px-3 py-2 text-sm bg-success-primary text-white rounded-md hover:bg-success-dark transition-colors">
                Grant Consent
              </button>
              <button onClick={() => showNotification('revoked')} className="px-3 py-2 text-sm bg-warning-primary text-white rounded-md hover:bg-warning-dark transition-colors">
                Revoke Consent
              </button>
              <button onClick={() => showNotification('expired')} className="px-3 py-2 text-sm bg-info-primary text-white rounded-md hover:bg-info-dark transition-colors">
                Consent Expired
              </button>
              <button onClick={() => showNotification('error')} className="px-3 py-2 text-sm bg-error-primary text-white rounded-md hover:bg-error-dark transition-colors">
                Security Alert
              </button>
            </div>
          </div>

          <div className="space-y-2">
            {notifications.map(notification => {
            if (notification.type === 'granted') {
              return <Toast key={notification.id} variant="success" title="Consent Granted" duration={5000} position={undefined as any} onDismiss={() => removeNotification(notification.id)}>
                    You've successfully granted access to your profile data. This
                    permission is valid for 30 days.
                  </Toast>;
            }
            if (notification.type === 'revoked') {
              return <Toast key={notification.id} variant="warning" title="Consent Revoked" duration={5000} position={undefined as any} onDismiss={() => removeNotification(notification.id)}>
                    Access permissions have been revoked. The application will no
                    longer have access to your data.
                  </Toast>;
            }
            if (notification.type === 'expired') {
              return <Toast key={notification.id} variant="info" title="Consent Expired" duration={0} position={undefined as any} onDismiss={() => removeNotification(notification.id)} action={{
                label: 'Renew',
                onClick: () => {
                  alert('Renewing consent...');
                  removeNotification(notification.id);
                }
              }}>
                    Your consent for "Marketing Platform" has expired. Renew to
                    continue sharing data.
                  </Toast>;
            }
            if (notification.type === 'error') {
              return <Toast key={notification.id} variant="error" title="Security Alert" duration={0} position={undefined as any} onDismiss={() => removeNotification(notification.id)} action={{
                label: 'Review',
                onClick: () => alert('Opening security settings...')
              }}>
                    Unusual activity detected. Please review your recent consent
                    grants.
                  </Toast>;
            }
            return null;
          })}
          </div>
        </div>
      </div>;
  },
  args: {
    variant: 'info',
    children: 'Real-world example'
  }
}`,...(kt=(Tt=w.parameters)==null?void 0:Tt.docs)==null?void 0:kt.source},description:{story:"Real-world use case examples",...(jt=(Nt=w.parameters)==null?void 0:Nt.docs)==null?void 0:jt.description}}};const ts=["Default","Variants","WithIcons","WithActions","WithTimer","Stacked","Positions","CustomDuration","Playground","WithTitle","RealWorldExamples"];export{v as CustomDuration,u as Default,b as Playground,f as Positions,w as RealWorldExamples,x as Stacked,m as Variants,h as WithActions,p as WithIcons,g as WithTimer,y as WithTitle,ts as __namedExportsOrder,Zt as default};
