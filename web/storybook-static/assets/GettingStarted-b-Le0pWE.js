import{j as e}from"./jsx-runtime-BYYWji4R.js";import{useMDXComponents as r}from"./index-DUy19JZU.js";import{M as t}from"./index-Dyh1TpE1.js";import"./index-ClcD9ViR.js";import"./_commonjsHelpers-Cpj98o6Y.js";import"./iframe-BkNYAzJV.js";import"./index-BUAr5TKG.js";import"./index-Bhelpi4i.js";import"./index-DrFu-skq.js";function s(i){const n={code:"code",h1:"h1",h2:"h2",h3:"h3",li:"li",p:"p",pre:"pre",strong:"strong",ul:"ul",...r(),...i.components};return e.jsxs(e.Fragment,{children:[e.jsx(t,{title:"Design System/Getting Started"}),`
`,e.jsx(n.h1,{id:"agentic-identity-broker-design-system",children:"Agentic Identity Broker Design System"}),`
`,e.jsxs(n.p,{children:["Welcome to the ",e.jsx(n.strong,{children:"Refined Trust Architecture"})," design system for the consent management frontend."]}),`
`,e.jsx(n.h2,{id:"overview",children:"Overview"}),`
`,e.jsxs(n.p,{children:["This design system provides a comprehensive component library optimized for capturing consent from non-technical end-users who delegate permissions to AI agents. Every component is designed to communicate ",e.jsx(n.strong,{children:"trustworthiness, clarity, and control"}),"."]}),`
`,e.jsx(n.h2,{id:"design-philosophy",children:"Design Philosophy"}),`
`,e.jsx(n.h3,{id:"refined-trust-architecture",children:"Refined Trust Architecture"}),`
`,e.jsxs(n.p,{children:["The aesthetic blends the visual gravity of a ",e.jsx(n.strong,{children:"bank vault"})," with the approachability of ",e.jsx(n.strong,{children:"modern SaaS"}),". This is where users make critical decisions about AI agent permissions, so every pixel must communicate authority and trust."]}),`
`,e.jsx(n.p,{children:e.jsx(n.strong,{children:"Core Principles:"})}),`
`,e.jsxs(n.ul,{children:[`
`,e.jsxs(n.li,{children:[e.jsx(n.strong,{children:"Serif Authority with Warm Humanity"}),": Crimson Pro (serif) for headings creates trust, Manrope (sans) for body text adds approachability"]}),`
`,e.jsxs(n.li,{children:[e.jsx(n.strong,{children:"Color as Semantic Signal"}),": Navy for authority, Emerald for success, Amber for attention"]}),`
`,e.jsxs(n.li,{children:[e.jsx(n.strong,{children:"Elevation Through Shadow"}),": Multi-layer shadows create depth without borders"]}),`
`,e.jsxs(n.li,{children:[e.jsx(n.strong,{children:"Motion That Guides"}),": Purposeful animations (150-500ms) that guide without entertaining"]}),`
`,e.jsxs(n.li,{children:[e.jsx(n.strong,{children:"Whitespace as Luxury"}),": Generous padding and spacing create a premium feel"]}),`
`]}),`
`,e.jsx(n.h2,{id:"typography",children:"Typography"}),`
`,e.jsx(n.pre,{children:e.jsx(n.code,{className:"language-css",children:`Headings: Crimson Pro (serif) - 600/700 weight
Body: Manrope (sans-serif) - 400/500 weight
Monospace: JetBrains Mono - 500 weight
`})}),`
`,e.jsx(n.h2,{id:"color-system",children:"Color System"}),`
`,e.jsx(n.h3,{id:"semantic-tokens",children:"Semantic Tokens"}),`
`,e.jsxs(n.ul,{children:[`
`,e.jsxs(n.li,{children:[e.jsx(n.strong,{children:"trust-deep"})," (#0A2540): Primary actions, critical UI"]}),`
`,e.jsxs(n.li,{children:[e.jsx(n.strong,{children:"trust"})," (#1E4D6B): Medium authority elements"]}),`
`,e.jsxs(n.li,{children:[e.jsx(n.strong,{children:"trust-light"})," (#E8F1F5): Light backgrounds"]}),`
`,e.jsxs(n.li,{children:[e.jsx(n.strong,{children:"cta"})," (#D97706): Call-to-action buttons, warnings"]}),`
`,e.jsxs(n.li,{children:[e.jsx(n.strong,{children:"success-primary"})," (#059669): Success states, granted permissions"]}),`
`,e.jsxs(n.li,{children:[e.jsx(n.strong,{children:"error-primary"})," (#DC2626): Error states"]}),`
`]}),`
`,e.jsx(n.h2,{id:"component-categories",children:"Component Categories"}),`
`,e.jsx(n.h3,{id:"primitives",children:"Primitives"}),`
`,e.jsx(n.p,{children:"Button, Badge, Avatar, Divider, Spinner, Icon"}),`
`,e.jsx(n.h3,{id:"inputs",children:"Inputs"}),`
`,e.jsx(n.p,{children:"TextInput, TextArea, Select, Checkbox, Radio, Switch, DatePicker"}),`
`,e.jsx(n.h3,{id:"feedback",children:"Feedback"}),`
`,e.jsx(n.p,{children:"Alert, Toast, InlineError, EmptyState, Skeleton"}),`
`,e.jsx(n.h3,{id:"overlays",children:"Overlays"}),`
`,e.jsx(n.p,{children:"Modal, Tooltip, Dropdown, Popover"}),`
`,e.jsx(n.h3,{id:"navigation",children:"Navigation"}),`
`,e.jsx(n.p,{children:"Tabs, Breadcrumb, Pagination"}),`
`,e.jsx(n.h3,{id:"data-display",children:"Data Display"}),`
`,e.jsx(n.p,{children:"Table, Card, GrantStatusBadge, ScopeList"}),`
`,e.jsx(n.h3,{id:"layout",children:"Layout"}),`
`,e.jsx(n.p,{children:"Container, Stack, Grid, AppLayout, PageTransition"}),`
`,e.jsx(n.h3,{id:"advanced",children:"Advanced"}),`
`,e.jsx(n.p,{children:"Accordion, Progress"}),`
`,e.jsx(n.h2,{id:"usage",children:"Usage"}),`
`,e.jsx(n.p,{children:"Import components from the design system:"}),`
`,e.jsx(n.pre,{children:e.jsx(n.code,{className:"language-typescript",children:`import { Button } from '@design-system/components/primitives/Button';
import { cn } from '@design-system/utils';
import { semanticColors } from '@design-system/tokens';
`})}),`
`,e.jsx(n.h2,{id:"accessibility",children:"Accessibility"}),`
`,e.jsxs(n.p,{children:["All components follow ",e.jsx(n.strong,{children:"WCAG 2.1 AA"})," guidelines:"]}),`
`,e.jsxs(n.ul,{children:[`
`,e.jsx(n.li,{children:"Keyboard navigation support"}),`
`,e.jsx(n.li,{children:"Screen reader compatibility"}),`
`,e.jsx(n.li,{children:"Visible focus indicators"}),`
`,e.jsx(n.li,{children:"Proper ARIA labels"}),`
`,e.jsx(n.li,{children:"Color contrast compliance"}),`
`]}),`
`,e.jsx(n.h2,{id:"getting-started",children:"Getting Started"}),`
`,e.jsx(n.p,{children:"Browse the component categories in the sidebar to explore interactive examples, view all variants, and experiment with the Playground stories."})]})}function g(i={}){const{wrapper:n}={...r(),...i.components};return n?e.jsx(n,{...i,children:e.jsx(s,{...i})}):s(i)}export{g as default};
