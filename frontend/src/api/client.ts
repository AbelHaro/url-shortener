import axios from "axios";

const baseURL = "/api/v1";

export const apiClient = axios.create({
  baseURL: baseURL,
});
