import { jsx as _jsx } from "react/jsx-runtime";
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
function App() {
    return (_jsx(Router, { basename: "/consent", children: _jsx(Routes, { children: _jsx(Route, { path: "/", element: _jsx("div", { children: "Welcome to Consent Management" }) }) }) }));
}
export default App;
//# sourceMappingURL=App.js.map