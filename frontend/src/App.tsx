import { Routes, Route } from "react-router-dom";
import Landing from "@/pages/Landing";
import Short from "@/pages/Short";
import Redirect from "@/pages/Redirect";
import Register from "@/pages/Register";
import Login from "@/pages/Login";
import { MyURLs } from "@/pages/MyUrls";
import { URLStatistics } from "@/pages/URLStatistics";
import { AuthenticatedLayout } from "@/components/authenticated-layout";

export function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route element={<AuthenticatedLayout />}>
        <Route path="/short" element={<Short />} />
        <Route path="/myurls" element={<MyURLs />} />
        <Route path="/myurls/:id" element={<URLStatistics />} />
      </Route>
      <Route path="/:shortCode" element={<Redirect />} />
      <Route path="*" element={<Landing />} />
    </Routes>
  );
}

export default App;
