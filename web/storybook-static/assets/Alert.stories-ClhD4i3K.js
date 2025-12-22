import{j as e}from"./jsx-runtime-BYYWji4R.js";import{R as Ke,r as n}from"./index-ClcD9ViR.js";import{c as A,a as qe}from"./cn-JCLedEej.js";import"./_commonjsHelpers-Cpj98o6Y.js";const Qe=qe("relative flex gap-3 rounded-lg border p-4 transition-all duration-300",{variants:{variant:{info:"bg-info-light text-info-dark border-info-primary/30",success:"bg-success-light text-success-dark border-success-primary/30",warning:"bg-warning-light text-warning-dark border-warning-primary/30",error:"bg-error-light text-error-dark border-error-primary/30"},banner:{true:"rounded-none border-l-0 border-r-0 w-full",false:""}},defaultVariants:{variant:"info",banner:!1}}),Xe=qe("flex-shrink-0 w-5 h-5",{variants:{variant:{info:"text-info-primary",success:"text-success-primary",warning:"text-warning-primary",error:"text-error-primary"}},defaultVariants:{variant:"info"}}),Ze=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"})}),$e=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"})}),er=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"})}),rr=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"})}),tr=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor","aria-hidden":"true",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M6 18L18 6M6 6l12 12"})}),t=Ke.forwardRef(({variant:r="info",icon:s,title:a,dismissible:i=!1,onDismiss:o,action:w,banner:Le=!1,hideIcon:Pe=!1,className:Ye,children:Ee,...Me},Be)=>{const[Oe,Ue]=n.useState(!0),[Fe,ze]=n.useState(!1),Ge=()=>{switch(r){case"success":return e.jsx($e,{});case"warning":return e.jsx(er,{});case"error":return e.jsx(rr,{});case"info":default:return e.jsx(Ze,{})}},He=()=>{ze(!0),setTimeout(()=>{Ue(!1),o==null||o()},300)};if(!Oe)return null;const _e=r==="error"?"alert":"status",Je=r==="error"?"assertive":"polite";return e.jsxs("div",{ref:Be,role:_e,"aria-live":Je,className:A(Qe({variant:r,banner:Le}),Fe&&"opacity-0 scale-95",Ye),...Me,children:[!Pe&&e.jsx("div",{className:Xe({variant:r}),children:s||Ge()}),e.jsxs("div",{className:"flex-1 min-w-0",children:[a&&e.jsx("h4",{className:"text-sm font-semibold mb-1 leading-tight",children:a}),e.jsx("div",{className:"text-sm leading-relaxed",children:Ee})]}),w&&e.jsx("button",{type:"button",onClick:w.onClick,className:A("flex-shrink-0 px-3 py-1.5 text-sm font-medium rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2",r==="info"&&"text-info-primary hover:bg-info-primary/10 focus:ring-info-primary",r==="success"&&"text-success-primary hover:bg-success-primary/10 focus:ring-success-primary",r==="warning"&&"text-warning-primary hover:bg-warning-primary/10 focus:ring-warning-primary",r==="error"&&"text-error-primary hover:bg-error-primary/10 focus:ring-error-primary"),children:w.label}),i&&e.jsx("button",{type:"button",onClick:He,"aria-label":"Dismiss alert",className:A("flex-shrink-0 inline-flex items-center justify-center w-8 h-8 rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2",r==="info"&&"text-info-primary/70 hover:text-info-primary hover:bg-info-primary/10 focus:ring-info-primary",r==="success"&&"text-success-primary/70 hover:text-success-primary hover:bg-success-primary/10 focus:ring-success-primary",r==="warning"&&"text-warning-primary/70 hover:text-warning-primary hover:bg-warning-primary/10 focus:ring-warning-primary",r==="error"&&"text-error-primary/70 hover:text-error-primary hover:bg-error-primary/10 focus:ring-error-primary"),children:e.jsx(tr,{})})]})});t.displayName="Alert";t.__docgenInfo={description:`Alert component for displaying contextual feedback messages.
Supports multiple variants, icons, titles, and actions.

@example
\`\`\`tsx
<Alert variant="success">
  Your changes have been saved successfully.
</Alert>

<Alert variant="warning" title="Warning" dismissible>
  Please review the following information before proceeding.
</Alert>

<Alert
  variant="error"
  title="Error"
  action={{ label: "Retry", onClick: handleRetry }}
>
  Failed to process your request.
</Alert>
\`\`\``,methods:[],displayName:"Alert",props:{variant:{required:!1,tsType:{name:"union",raw:"'info' | 'success' | 'warning' | 'error'",elements:[{name:"literal",value:"'info'"},{name:"literal",value:"'success'"},{name:"literal",value:"'warning'"},{name:"literal",value:"'error'"}]},description:"Alert variant based on message severity",defaultValue:{value:"'info'",computed:!1}},icon:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:"Optional icon to display (overrides default icon)"},title:{required:!1,tsType:{name:"string"},description:"Optional title text (bold)"},dismissible:{required:!1,tsType:{name:"boolean"},description:"Whether alert can be dismissed",defaultValue:{value:"false",computed:!1}},onDismiss:{required:!1,tsType:{name:"signature",type:"function",raw:"() => void",signature:{arguments:[],return:{name:"void"}}},description:"Callback when alert is dismissed"},action:{required:!1,tsType:{name:"signature",type:"object",raw:`{
  label: string;
  onClick: () => void;
}`,signature:{properties:[{key:"label",value:{name:"string",required:!0}},{key:"onClick",value:{name:"signature",type:"function",raw:"() => void",signature:{arguments:[],return:{name:"void"}},required:!0}}]}},description:"Optional action button"},banner:{required:!1,tsType:{name:"boolean"},description:"Display as full-width banner",defaultValue:{value:"false",computed:!1}},hideIcon:{required:!1,tsType:{name:"boolean"},description:"Hide default icon",defaultValue:{value:"false",computed:!1}}},composes:["Omit"]};const or={title:"Design System/Feedback/Alert",component:t,parameters:{layout:"padded"},tags:["autodocs"],argTypes:{variant:{control:"select",options:["info","success","warning","error"],description:"Visual style variant based on message severity"},title:{control:"text",description:"Optional bold title text"},dismissible:{control:"boolean",description:"Whether the alert can be dismissed"},banner:{control:"boolean",description:"Display as full-width banner"},hideIcon:{control:"boolean",description:"Hide the default icon"},children:{control:"text",description:"Alert message content"}}},l={args:{variant:"info",children:"This is an informational message to provide context or guidance to the user."}},c={render:()=>e.jsxs("div",{className:"space-y-4 max-w-2xl",children:[e.jsx(t,{variant:"info",children:"This is an informational alert. It provides helpful context or additional information about the current action."}),e.jsx(t,{variant:"success",children:"Success! Your changes have been saved and will take effect immediately."}),e.jsx(t,{variant:"warning",children:"Warning: This action cannot be undone. Please review your selection carefully before proceeding."}),e.jsx(t,{variant:"error",children:"Error: Unable to process your request. Please check your connection and try again."})]}),args:{variant:"info",children:"Alert content"}},d={render:()=>e.jsxs("div",{className:"space-y-4 max-w-2xl",children:[e.jsx(t,{variant:"info",children:"Each alert variant comes with a default icon that matches its semantic meaning."}),e.jsx(t,{variant:"success",children:"The success icon indicates a positive outcome or completed action."}),e.jsx(t,{variant:"warning",children:"The warning icon draws attention to important caution messages."}),e.jsx(t,{variant:"error",children:"The error icon clearly indicates a problem that needs attention."})]}),args:{variant:"info",children:"Alert with icon"}},u={render:()=>{const[r,s]=n.useState({action:!0,dismiss:!0,both:!0});return e.jsxs("div",{className:"space-y-4 max-w-2xl",children:[r.action&&e.jsx(t,{variant:"warning",action:{label:"Review",onClick:()=>alert("Action clicked!")},children:"Your subscription will expire in 3 days. Review your plan to continue using all features."}),r.dismiss&&e.jsx(t,{variant:"info",dismissible:!0,onDismiss:()=>s({...r,dismiss:!1}),children:"New features are available! Check out our latest updates in the changelog."}),r.both&&e.jsx(t,{variant:"error",dismissible:!0,onDismiss:()=>s({...r,both:!1}),action:{label:"Retry",onClick:()=>alert("Retrying...")},children:"Payment processing failed. Please verify your payment method and try again."}),!r.action&&!r.dismiss&&!r.both&&e.jsxs("div",{className:"text-center p-8 border border-dashed border-gray-300 rounded-lg",children:[e.jsx("p",{className:"text-gray-600 mb-4",children:"All alerts dismissed"}),e.jsx("button",{onClick:()=>s({action:!0,dismiss:!0,both:!0}),className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors",children:"Reset Alerts"})]})]})},args:{variant:"info",children:"Alert with actions"}},m={render:()=>e.jsxs("div",{className:"space-y-4 max-w-2xl",children:[e.jsx(t,{variant:"info",title:"Information",children:"Titles help organize complex alerts and provide clear context for the message. They appear in bold above the description."}),e.jsx(t,{variant:"success",title:"Access Granted",children:'You now have editor permissions for the "Marketing Assets" workspace. You can view, edit, and share all documents.'}),e.jsx(t,{variant:"warning",title:"Action Required",children:"Your account requires two-factor authentication. Please enable 2FA within 7 days to maintain access to sensitive resources."}),e.jsx(t,{variant:"error",title:"Authentication Failed",children:"We couldn't verify your identity. Your session may have expired or your credentials are invalid. Please sign in again."})]}),args:{variant:"info",title:"Alert Title",children:"Alert description goes here"}},p={render:()=>{const[r,s]=n.useState({alert1:!0,alert2:!0,alert3:!0,alert4:!0}),a=()=>{s({alert1:!0,alert2:!0,alert3:!0,alert4:!0})},i=!Object.values(r).some(o=>o);return e.jsxs("div",{className:"space-y-4 max-w-2xl",children:[r.alert1&&e.jsx(t,{variant:"info",title:"Tip",dismissible:!0,onDismiss:()=>s({...r,alert1:!1}),children:"Click the × button to dismiss any alert. It will fade out smoothly."}),r.alert2&&e.jsx(t,{variant:"success",dismissible:!0,onDismiss:()=>s({...r,alert2:!1}),children:"Your profile has been updated successfully."}),r.alert3&&e.jsx(t,{variant:"warning",title:"Reminder",dismissible:!0,onDismiss:()=>s({...r,alert3:!1}),children:"Don't forget to save your work before closing the window."}),r.alert4&&e.jsx(t,{variant:"error",dismissible:!0,onDismiss:()=>s({...r,alert4:!1}),children:"Connection lost. Attempting to reconnect..."}),i&&e.jsxs("div",{className:"text-center p-6 border border-dashed border-gray-300 rounded-lg",children:[e.jsx("p",{className:"text-gray-600 mb-4",children:"Some alerts have been dismissed"}),e.jsx("button",{onClick:a,className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors",children:"Show All Alerts"})]})]})},args:{variant:"info",dismissible:!0,children:"This alert can be dismissed"}},h={render:()=>e.jsxs("div",{className:"space-y-0 -mx-4",children:[e.jsx(t,{variant:"info",banner:!0,action:{label:"Learn More",onClick:()=>alert("Learn more!")},children:"New consent management features are now available. Learn about enhanced privacy controls and compliance tools."}),e.jsx("div",{className:"p-4",children:e.jsx("p",{className:"text-gray-700 mb-4",children:"Banner alerts are typically used at the top of a page or section to display important system-wide messages."})}),e.jsx(t,{variant:"warning",banner:!0,dismissible:!0,onDismiss:()=>alert("Banner dismissed"),children:"Scheduled maintenance on Saturday, Dec 21 from 2:00 AM - 4:00 AM UTC. Services may be temporarily unavailable."}),e.jsx("div",{className:"p-4",children:e.jsx("p",{className:"text-gray-700",children:"They span the full width of their container and have no rounded corners on the sides."})}),e.jsx(t,{variant:"error",banner:!0,title:"System Alert",action:{label:"View Status",onClick:()=>alert("View status")},children:"We're experiencing higher than normal response times. Our team is working to resolve this issue."})]}),args:{variant:"info",banner:!0,children:"This is a full-width banner alert"}},v={render:r=>{const[s,a]=n.useState(!0);return s?e.jsx(t,{...r,onDismiss:()=>{var i;a(!1),(i=r.onDismiss)==null||i.call(r)}}):e.jsxs("div",{className:"text-center p-8 border border-dashed border-gray-300 rounded-lg",children:[e.jsx("p",{className:"text-gray-600 mb-4",children:"Alert dismissed"}),e.jsx("button",{onClick:()=>a(!0),className:"px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors",children:"Show Alert"})]})},args:{variant:"info",title:"Alert Title",children:"This is the alert message content. You can customize all properties using the controls below.",dismissible:!0,banner:!1,hideIcon:!1,action:void 0}},g={render:()=>{const[r,s]=n.useState({pending:!0,granted:!1,expired:!0});return e.jsxs("div",{className:"space-y-6 max-w-3xl",children:[e.jsxs("div",{children:[e.jsx("h3",{className:"text-lg font-semibold text-gray-900 mb-4",children:"Consent Management Scenarios"}),e.jsxs("div",{className:"space-y-4",children:[r.pending&&e.jsx(t,{variant:"warning",title:"Consent Request Pending",dismissible:!0,onDismiss:()=>s({...r,pending:!1}),action:{label:"Review Request",onClick:()=>alert("Opening consent request...")},children:"A third-party application is requesting access to your profile data. Please review and respond to this consent request."}),r.granted&&e.jsx(t,{variant:"success",title:"Access Granted",children:`You've successfully granted "Analytics Dashboard" access to your usage statistics. This permission is valid for 30 days.`}),r.expired&&e.jsx(t,{variant:"info",title:"Consent Expired",dismissible:!0,onDismiss:()=>s({...r,expired:!1}),action:{label:"Renew",onClick:()=>{s({...r,expired:!1,granted:!0}),alert("Consent renewed!")}},children:'Your consent for "Marketing Platform" has expired. Renew to continue sharing data with this service.'}),e.jsx(t,{variant:"error",title:"Security Alert",action:{label:"Secure Account",onClick:()=>alert("Opening security settings...")},children:"We detected unusual activity on your account. Please review your recent consent grants and revoke any suspicious permissions."})]})]}),e.jsx("div",{className:"pt-4 border-t border-gray-200",children:e.jsx("button",{onClick:()=>s({pending:!0,granted:!1,expired:!0}),className:"text-sm text-trust-deep hover:text-trust-hover underline",children:"Reset Examples"})})]})},args:{variant:"info",children:"Real-world example"}},x={render:()=>e.jsxs("div",{className:"space-y-4 max-w-2xl",children:[e.jsx(t,{variant:"info",hideIcon:!0,title:"Plain Information",children:"Sometimes you may want to display alerts without icons for a cleaner look or when the icon doesn't add meaningful context."}),e.jsx(t,{variant:"success",hideIcon:!0,children:"Your settings have been saved."}),e.jsx(t,{variant:"warning",hideIcon:!0,children:"This feature is currently in beta."})]}),args:{variant:"info",hideIcon:!0,children:"Alert without icon"}},f={render:()=>{const r=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"})}),s=()=>e.jsx("svg",{className:"w-5 h-5",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",children:e.jsx("path",{strokeLinecap:"round",strokeLinejoin:"round",strokeWidth:2,d:"M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"})});return e.jsxs("div",{className:"space-y-4 max-w-2xl",children:[e.jsx(t,{variant:"info",icon:e.jsx(r,{}),title:"New Document",children:"You can override the default icon with any custom React component or SVG."}),e.jsx(t,{variant:"success",icon:e.jsx(s,{}),title:"Notification Settings Updated",children:"You will now receive email notifications for important account activities."})]})},args:{variant:"info",icon:void 0,children:"Alert with custom icon"}},y={render:()=>e.jsxs("div",{className:"space-y-6 max-w-3xl",children:[e.jsxs("div",{className:"p-4 bg-gray-50 border border-gray-200 rounded-lg",children:[e.jsx("h4",{className:"text-sm font-semibold text-gray-900 mb-2",children:"Accessibility Features"}),e.jsxs("ul",{className:"text-sm text-gray-700 space-y-1",children:[e.jsxs("li",{children:["• Error alerts use ",e.jsx("code",{children:'role="alert"'})," and"," ",e.jsx("code",{children:'aria-live="assertive"'})]}),e.jsxs("li",{children:["• Other variants use ",e.jsx("code",{children:'role="status"'})," and"," ",e.jsx("code",{children:'aria-live="polite"'})]}),e.jsxs("li",{children:["• Dismiss buttons have ",e.jsx("code",{children:'aria-label="Dismiss alert"'})]}),e.jsxs("li",{children:["• Icons are marked with ",e.jsx("code",{children:'aria-hidden="true"'})]}),e.jsx("li",{children:"• All interactive elements are keyboard accessible"}),e.jsx("li",{children:"• Focus indicators meet WCAG 2.1 AA requirements"}),e.jsx("li",{children:"• Color contrast ratios comply with AA standards"})]})]}),e.jsx(t,{variant:"error",title:"Critical Error",dismissible:!0,children:"This error alert will be announced immediately to screen readers due to its assertive aria-live setting."}),e.jsx(t,{variant:"success",dismissible:!0,action:{label:"Undo",onClick:()=>alert("Undo action")},children:"Changes saved. Use Tab to navigate between action and dismiss buttons."})]}),parameters:{a11y:{config:{rules:[{id:"color-contrast",enabled:!0},{id:"button-name",enabled:!0}]}}},args:{variant:"info",children:"Accessible alert"}},b={render:()=>e.jsxs("div",{className:"space-y-4 max-w-3xl",children:[e.jsx(t,{variant:"info",title:"Terms of Service Update",children:e.jsxs("div",{className:"space-y-2",children:[e.jsx("p",{children:"We've updated our Terms of Service to provide better clarity on data usage and your privacy rights."}),e.jsx("p",{children:"Key changes include enhanced data portability options, clearer consent management workflows, and updated retention policies."}),e.jsx("p",{className:"font-medium",children:"Please review the changes by January 1, 2026 to continue using our services."})]})}),e.jsx(t,{variant:"warning",title:"Data Export Ready",action:{label:"Download",onClick:()=>alert("Downloading...")},children:e.jsxs("div",{className:"space-y-2",children:[e.jsx("p",{children:"Your requested data export is ready for download."}),e.jsxs("ul",{className:"list-disc list-inside text-sm space-y-1",children:[e.jsx("li",{children:"Profile information and settings"}),e.jsx("li",{children:"Consent history and audit logs"}),e.jsx("li",{children:"Connected applications and permissions"})]}),e.jsx("p",{className:"text-xs mt-2",children:"Download link expires in 7 days (December 26, 2025)"})]})})]}),args:{variant:"info",title:"Complex Content",children:"Alert with structured content"}};var j,k,N,C,S;l.parameters={...l.parameters,docs:{...(j=l.parameters)==null?void 0:j.docs,source:{originalSource:`{
  args: {
    variant: 'info',
    children: 'This is an informational message to provide context or guidance to the user.'
  }
}`,...(N=(k=l.parameters)==null?void 0:k.docs)==null?void 0:N.source},description:{story:"Default alert with info variant",...(S=(C=l.parameters)==null?void 0:C.docs)==null?void 0:S.description}}};var D,T,I,R,W;c.parameters={...c.parameters,docs:{...(D=c.parameters)==null?void 0:D.docs,source:{originalSource:`{
  render: () => <div className="space-y-4 max-w-2xl">
      <Alert variant="info">
        This is an informational alert. It provides helpful context or additional
        information about the current action.
      </Alert>

      <Alert variant="success">
        Success! Your changes have been saved and will take effect immediately.
      </Alert>

      <Alert variant="warning">
        Warning: This action cannot be undone. Please review your selection
        carefully before proceeding.
      </Alert>

      <Alert variant="error">
        Error: Unable to process your request. Please check your connection and
        try again.
      </Alert>
    </div>,
  args: {
    variant: 'info',
    children: 'Alert content'
  }
}`,...(I=(T=c.parameters)==null?void 0:T.docs)==null?void 0:I.source},description:{story:"All alert variants displayed together",...(W=(R=c.parameters)==null?void 0:R.docs)==null?void 0:W.description}}};var V,q,L,P,Y;d.parameters={...d.parameters,docs:{...(V=d.parameters)==null?void 0:V.docs,source:{originalSource:`{
  render: () => <div className="space-y-4 max-w-2xl">
      <Alert variant="info">
        Each alert variant comes with a default icon that matches its semantic
        meaning.
      </Alert>

      <Alert variant="success">
        The success icon indicates a positive outcome or completed action.
      </Alert>

      <Alert variant="warning">
        The warning icon draws attention to important caution messages.
      </Alert>

      <Alert variant="error">
        The error icon clearly indicates a problem that needs attention.
      </Alert>
    </div>,
  args: {
    variant: 'info',
    children: 'Alert with icon'
  }
}`,...(L=(q=d.parameters)==null?void 0:q.docs)==null?void 0:L.source},description:{story:"Alerts with leading icons (default behavior)",...(Y=(P=d.parameters)==null?void 0:P.docs)==null?void 0:Y.description}}};var E,M,B,O,U;u.parameters={...u.parameters,docs:{...(E=u.parameters)==null?void 0:E.docs,source:{originalSource:`{
  render: () => {
    const [alerts, setAlerts] = useState({
      action: true,
      dismiss: true,
      both: true
    });
    return <div className="space-y-4 max-w-2xl">
        {alerts.action && <Alert variant="warning" action={{
        label: 'Review',
        onClick: () => alert('Action clicked!')
      }}>
            Your subscription will expire in 3 days. Review your plan to continue
            using all features.
          </Alert>}

        {alerts.dismiss && <Alert variant="info" dismissible onDismiss={() => setAlerts({
        ...alerts,
        dismiss: false
      })}>
            New features are available! Check out our latest updates in the
            changelog.
          </Alert>}

        {alerts.both && <Alert variant="error" dismissible onDismiss={() => setAlerts({
        ...alerts,
        both: false
      })} action={{
        label: 'Retry',
        onClick: () => alert('Retrying...')
      }}>
            Payment processing failed. Please verify your payment method and try
            again.
          </Alert>}

        {!alerts.action && !alerts.dismiss && !alerts.both && <div className="text-center p-8 border border-dashed border-gray-300 rounded-lg">
            <p className="text-gray-600 mb-4">All alerts dismissed</p>
            <button onClick={() => setAlerts({
          action: true,
          dismiss: true,
          both: true
        })} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors">
              Reset Alerts
            </button>
          </div>}
      </div>;
  },
  args: {
    variant: 'info',
    children: 'Alert with actions'
  }
}`,...(B=(M=u.parameters)==null?void 0:M.docs)==null?void 0:B.source},description:{story:"Alerts with action buttons and/or dismiss buttons",...(U=(O=u.parameters)==null?void 0:O.docs)==null?void 0:U.description}}};var F,z,G,H,_;m.parameters={...m.parameters,docs:{...(F=m.parameters)==null?void 0:F.docs,source:{originalSource:`{
  render: () => <div className="space-y-4 max-w-2xl">
      <Alert variant="info" title="Information">
        Titles help organize complex alerts and provide clear context for the
        message. They appear in bold above the description.
      </Alert>

      <Alert variant="success" title="Access Granted">
        You now have editor permissions for the "Marketing Assets" workspace. You
        can view, edit, and share all documents.
      </Alert>

      <Alert variant="warning" title="Action Required">
        Your account requires two-factor authentication. Please enable 2FA within
        7 days to maintain access to sensitive resources.
      </Alert>

      <Alert variant="error" title="Authentication Failed">
        We couldn't verify your identity. Your session may have expired or your
        credentials are invalid. Please sign in again.
      </Alert>
    </div>,
  args: {
    variant: 'info',
    title: 'Alert Title',
    children: 'Alert description goes here'
  }
}`,...(G=(z=m.parameters)==null?void 0:z.docs)==null?void 0:G.source},description:{story:"Alerts with bold titles and descriptions",...(_=(H=m.parameters)==null?void 0:H.docs)==null?void 0:_.description}}};var J,K,Q,X,Z;p.parameters={...p.parameters,docs:{...(J=p.parameters)==null?void 0:J.docs,source:{originalSource:`{
  render: () => {
    const [alertStates, setAlertStates] = useState({
      alert1: true,
      alert2: true,
      alert3: true,
      alert4: true
    });
    const resetAlerts = () => {
      setAlertStates({
        alert1: true,
        alert2: true,
        alert3: true,
        alert4: true
      });
    };
    const anyDismissed = !Object.values(alertStates).some(v => v);
    return <div className="space-y-4 max-w-2xl">
        {alertStates.alert1 && <Alert variant="info" title="Tip" dismissible onDismiss={() => setAlertStates({
        ...alertStates,
        alert1: false
      })}>
            Click the × button to dismiss any alert. It will fade out smoothly.
          </Alert>}

        {alertStates.alert2 && <Alert variant="success" dismissible onDismiss={() => setAlertStates({
        ...alertStates,
        alert2: false
      })}>
            Your profile has been updated successfully.
          </Alert>}

        {alertStates.alert3 && <Alert variant="warning" title="Reminder" dismissible onDismiss={() => setAlertStates({
        ...alertStates,
        alert3: false
      })}>
            Don't forget to save your work before closing the window.
          </Alert>}

        {alertStates.alert4 && <Alert variant="error" dismissible onDismiss={() => setAlertStates({
        ...alertStates,
        alert4: false
      })}>
            Connection lost. Attempting to reconnect...
          </Alert>}

        {anyDismissed && <div className="text-center p-6 border border-dashed border-gray-300 rounded-lg">
            <p className="text-gray-600 mb-4">
              Some alerts have been dismissed
            </p>
            <button onClick={resetAlerts} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors">
              Show All Alerts
            </button>
          </div>}
      </div>;
  },
  args: {
    variant: 'info',
    dismissible: true,
    children: 'This alert can be dismissed'
  }
}`,...(Q=(K=p.parameters)==null?void 0:K.docs)==null?void 0:Q.source},description:{story:"Dismissible alerts with fade-out animation",...(Z=(X=p.parameters)==null?void 0:X.docs)==null?void 0:Z.description}}};var $,ee,re,te,se;h.parameters={...h.parameters,docs:{...($=h.parameters)==null?void 0:$.docs,source:{originalSource:`{
  render: () => <div className="space-y-0 -mx-4">
      <Alert variant="info" banner action={{
      label: 'Learn More',
      onClick: () => alert('Learn more!')
    }}>
        New consent management features are now available. Learn about enhanced
        privacy controls and compliance tools.
      </Alert>

      <div className="p-4">
        <p className="text-gray-700 mb-4">
          Banner alerts are typically used at the top of a page or section to
          display important system-wide messages.
        </p>
      </div>

      <Alert variant="warning" banner dismissible onDismiss={() => alert('Banner dismissed')}>
        Scheduled maintenance on Saturday, Dec 21 from 2:00 AM - 4:00 AM UTC.
        Services may be temporarily unavailable.
      </Alert>

      <div className="p-4">
        <p className="text-gray-700">
          They span the full width of their container and have no rounded corners
          on the sides.
        </p>
      </div>

      <Alert variant="error" banner title="System Alert" action={{
      label: 'View Status',
      onClick: () => alert('View status')
    }}>
        We're experiencing higher than normal response times. Our team is working
        to resolve this issue.
      </Alert>
    </div>,
  args: {
    variant: 'info',
    banner: true,
    children: 'This is a full-width banner alert'
  }
}`,...(re=(ee=h.parameters)==null?void 0:ee.docs)==null?void 0:re.source},description:{story:"Full-width banner style alerts",...(se=(te=h.parameters)==null?void 0:te.docs)==null?void 0:se.description}}};var ae,ie,ne,oe,le;v.parameters={...v.parameters,docs:{...(ae=v.parameters)==null?void 0:ae.docs,source:{originalSource:`{
  render: args => {
    const [isVisible, setIsVisible] = useState(true);
    if (!isVisible) {
      return <div className="text-center p-8 border border-dashed border-gray-300 rounded-lg">
          <p className="text-gray-600 mb-4">Alert dismissed</p>
          <button onClick={() => setIsVisible(true)} className="px-4 py-2 bg-trust-deep text-white rounded-md hover:bg-trust-hover transition-colors">
            Show Alert
          </button>
        </div>;
    }
    return <Alert {...args} onDismiss={() => {
      setIsVisible(false);
      args.onDismiss?.();
    }} />;
  },
  args: {
    variant: 'info',
    title: 'Alert Title',
    children: 'This is the alert message content. You can customize all properties using the controls below.',
    dismissible: true,
    banner: false,
    hideIcon: false,
    action: undefined
  }
}`,...(ne=(ie=v.parameters)==null?void 0:ie.docs)==null?void 0:ne.source},description:{story:"Interactive playground with all controls",...(le=(oe=v.parameters)==null?void 0:oe.docs)==null?void 0:le.description}}};var ce,de,ue,me,pe;g.parameters={...g.parameters,docs:{...(ce=g.parameters)==null?void 0:ce.docs,source:{originalSource:`{
  render: () => {
    const [consents, setConsents] = useState({
      pending: true,
      granted: false,
      expired: true
    });
    return <div className="space-y-6 max-w-3xl">
        <div>
          <h3 className="text-lg font-semibold text-gray-900 mb-4">
            Consent Management Scenarios
          </h3>

          <div className="space-y-4">
            {consents.pending && <Alert variant="warning" title="Consent Request Pending" dismissible onDismiss={() => setConsents({
            ...consents,
            pending: false
          })} action={{
            label: 'Review Request',
            onClick: () => alert('Opening consent request...')
          }}>
                A third-party application is requesting access to your profile
                data. Please review and respond to this consent request.
              </Alert>}

            {consents.granted && <Alert variant="success" title="Access Granted">
                You've successfully granted "Analytics Dashboard" access to your
                usage statistics. This permission is valid for 30 days.
              </Alert>}

            {consents.expired && <Alert variant="info" title="Consent Expired" dismissible onDismiss={() => setConsents({
            ...consents,
            expired: false
          })} action={{
            label: 'Renew',
            onClick: () => {
              setConsents({
                ...consents,
                expired: false,
                granted: true
              });
              alert('Consent renewed!');
            }
          }}>
                Your consent for "Marketing Platform" has expired. Renew to
                continue sharing data with this service.
              </Alert>}

            <Alert variant="error" title="Security Alert" action={{
            label: 'Secure Account',
            onClick: () => alert('Opening security settings...')
          }}>
              We detected unusual activity on your account. Please review your
              recent consent grants and revoke any suspicious permissions.
            </Alert>
          </div>
        </div>

        <div className="pt-4 border-t border-gray-200">
          <button onClick={() => setConsents({
          pending: true,
          granted: false,
          expired: true
        })} className="text-sm text-trust-deep hover:text-trust-hover underline">
            Reset Examples
          </button>
        </div>
      </div>;
  },
  args: {
    variant: 'info',
    children: 'Real-world example'
  }
}`,...(ue=(de=g.parameters)==null?void 0:de.docs)==null?void 0:ue.source},description:{story:"Real-world use case examples",...(pe=(me=g.parameters)==null?void 0:me.docs)==null?void 0:pe.description}}};var he,ve,ge,xe,fe;x.parameters={...x.parameters,docs:{...(he=x.parameters)==null?void 0:he.docs,source:{originalSource:`{
  render: () => <div className="space-y-4 max-w-2xl">
      <Alert variant="info" hideIcon title="Plain Information">
        Sometimes you may want to display alerts without icons for a cleaner look
        or when the icon doesn't add meaningful context.
      </Alert>

      <Alert variant="success" hideIcon>
        Your settings have been saved.
      </Alert>

      <Alert variant="warning" hideIcon>
        This feature is currently in beta.
      </Alert>
    </div>,
  args: {
    variant: 'info',
    hideIcon: true,
    children: 'Alert without icon'
  }
}`,...(ge=(ve=x.parameters)==null?void 0:ve.docs)==null?void 0:ge.source},description:{story:"Alerts without icons",...(fe=(xe=x.parameters)==null?void 0:xe.docs)==null?void 0:fe.description}}};var ye,be,we,Ae,je;f.parameters={...f.parameters,docs:{...(ye=f.parameters)==null?void 0:ye.docs,source:{originalSource:`{
  render: () => {
    const DocumentIcon = () => <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>;
    const BellIcon = () => <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
      </svg>;
    return <div className="space-y-4 max-w-2xl">
        <Alert variant="info" icon={<DocumentIcon />} title="New Document">
          You can override the default icon with any custom React component or
          SVG.
        </Alert>

        <Alert variant="success" icon={<BellIcon />} title="Notification Settings Updated">
          You will now receive email notifications for important account
          activities.
        </Alert>
      </div>;
  },
  args: {
    variant: 'info',
    icon: undefined,
    children: 'Alert with custom icon'
  }
}`,...(we=(be=f.parameters)==null?void 0:be.docs)==null?void 0:we.source},description:{story:"Custom icons (replacing defaults)",...(je=(Ae=f.parameters)==null?void 0:Ae.docs)==null?void 0:je.description}}};var ke,Ne,Ce,Se,De;y.parameters={...y.parameters,docs:{...(ke=y.parameters)==null?void 0:ke.docs,source:{originalSource:`{
  render: () => <div className="space-y-6 max-w-3xl">
      <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
        <h4 className="text-sm font-semibold text-gray-900 mb-2">
          Accessibility Features
        </h4>
        <ul className="text-sm text-gray-700 space-y-1">
          <li>
            • Error alerts use <code>role="alert"</code> and{' '}
            <code>aria-live="assertive"</code>
          </li>
          <li>
            • Other variants use <code>role="status"</code> and{' '}
            <code>aria-live="polite"</code>
          </li>
          <li>
            • Dismiss buttons have <code>aria-label="Dismiss alert"</code>
          </li>
          <li>• Icons are marked with <code>aria-hidden="true"</code></li>
          <li>• All interactive elements are keyboard accessible</li>
          <li>• Focus indicators meet WCAG 2.1 AA requirements</li>
          <li>• Color contrast ratios comply with AA standards</li>
        </ul>
      </div>

      <Alert variant="error" title="Critical Error" dismissible>
        This error alert will be announced immediately to screen readers due to
        its assertive aria-live setting.
      </Alert>

      <Alert variant="success" dismissible action={{
      label: 'Undo',
      onClick: () => alert('Undo action')
    }}>
        Changes saved. Use Tab to navigate between action and dismiss buttons.
      </Alert>
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
    variant: 'info',
    children: 'Accessible alert'
  }
}`,...(Ce=(Ne=y.parameters)==null?void 0:Ne.docs)==null?void 0:Ce.source},description:{story:"Accessibility features demonstration",...(De=(Se=y.parameters)==null?void 0:Se.docs)==null?void 0:De.description}}};var Te,Ie,Re,We,Ve;b.parameters={...b.parameters,docs:{...(Te=b.parameters)==null?void 0:Te.docs,source:{originalSource:`{
  render: () => <div className="space-y-4 max-w-3xl">
      <Alert variant="info" title="Terms of Service Update">
        <div className="space-y-2">
          <p>
            We've updated our Terms of Service to provide better clarity on data
            usage and your privacy rights.
          </p>
          <p>
            Key changes include enhanced data portability options, clearer consent
            management workflows, and updated retention policies.
          </p>
          <p className="font-medium">
            Please review the changes by January 1, 2026 to continue using our
            services.
          </p>
        </div>
      </Alert>

      <Alert variant="warning" title="Data Export Ready" action={{
      label: 'Download',
      onClick: () => alert('Downloading...')
    }}>
        <div className="space-y-2">
          <p>Your requested data export is ready for download.</p>
          <ul className="list-disc list-inside text-sm space-y-1">
            <li>Profile information and settings</li>
            <li>Consent history and audit logs</li>
            <li>Connected applications and permissions</li>
          </ul>
          <p className="text-xs mt-2">
            Download link expires in 7 days (December 26, 2025)
          </p>
        </div>
      </Alert>
    </div>,
  args: {
    variant: 'info',
    title: 'Complex Content',
    children: 'Alert with structured content'
  }
}`,...(Re=(Ie=b.parameters)==null?void 0:Ie.docs)==null?void 0:Re.source},description:{story:"Complex content with multiple paragraphs",...(Ve=(We=b.parameters)==null?void 0:We.docs)==null?void 0:Ve.description}}};const lr=["Default","Variants","WithIcons","WithActions","WithTitle","Dismissible","Banner","Playground","RealWorldExamples","WithoutIcons","CustomIcons","Accessibility","ComplexContent"];export{y as Accessibility,h as Banner,b as ComplexContent,f as CustomIcons,l as Default,p as Dismissible,v as Playground,g as RealWorldExamples,c as Variants,u as WithActions,d as WithIcons,m as WithTitle,x as WithoutIcons,lr as __namedExportsOrder,or as default};
