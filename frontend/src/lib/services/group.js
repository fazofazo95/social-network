import { apiRequest } from "src/lib/apiClient";

export async function createGroup(formData) {
  return apiRequest("/api/groups", {
    method: "POST",
    body: formData,
  });
}
