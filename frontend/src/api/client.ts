import axios from "axios";

const baseURL = import.meta.env.VITE_API_BASE_URL;

if (!baseURL) {
  console.error("Backend URL is not set.");
}

export const apiClient = axios.create({
  baseURL: baseURL,
});
