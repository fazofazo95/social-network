import { apiRequest } from "src/lib/apiClient";

export async function loginUser(credentials) {
  return apiRequest("/api/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(credentials),
  });
}

export async function signupUser(formData) {
  return apiRequest("/api/signup", {
    method: "POST",
    body: formData,
  });
}

export async function verifySession() {
  return apiRequest("/api/verify-session", {
    method: "GET",
  });
}
