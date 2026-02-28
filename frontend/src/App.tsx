import { Routes, Route } from "react-router-dom";
import Landing from "@/pages/Landing";
import Short from "@/pages/Short";
import Redirect from "@/pages/Redirect";

export function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/short" element={<Short />} />
      <Route path="/:shortCode" element={<Redirect />} />
    </Routes>
  );
}

export default App;
