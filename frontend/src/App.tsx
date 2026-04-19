import { Routes, Route } from "react-router-dom";
import Landing from "@/pages/Landing";
import Short from "@/pages/Short";
import Redirect from "@/pages/Redirect";
import Register from "@/pages/Register";
import Login from "@/pages/Login";

export function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/short" element={<Short />} />
      <Route path="/:shortCode" element={<Redirect />} />
    </Routes>
  );
}

export default App;
