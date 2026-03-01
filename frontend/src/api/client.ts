import axios from "axios";

const baseURL = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";

console.log("API Base URL:", baseURL);

export const apiClient = axios.create({
  baseURL: baseURL,
});
