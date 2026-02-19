import { apiRequest } from "src/lib/apiClient";

export async function fetchUserData(id) {
  const payload = await apiRequest(`/api/users/${id}`, {
    method: "GET",
  });
  return payload.data;
}
