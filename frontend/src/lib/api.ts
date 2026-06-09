import type {
  AuthResponse,
  MenuItem,
  Order,
  Restaurant,
  User,
} from "./types";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8090";

type RequestOptions = {
  method?: string;
  body?: unknown;
  token?: string | null;
};

const request = async <T>(path: string, options: RequestOptions = {}): Promise<T> => {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  if (options.token) {
    headers.Authorization = `Bearer ${options.token}`;
  }

  const response = await fetch(`${API_URL}${path}`, {
    method: options.method ?? "GET",
    headers,
    body: options.body ? JSON.stringify(options.body) : undefined,
  });

  if (!response.ok) {
    let message = `Request failed: ${response.status}`;
    try {
      const data = (await response.json()) as { error?: string };
      if (data.error) {
        message = data.error;
      }
    } catch {
      // ignore parse errors
    }
    throw new Error(message);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
};

export const api = {
  register: (body: {
    email: string;
    password: string;
    name: string;
    phone?: string;
  }) => request<AuthResponse>("/api/auth/register", { method: "POST", body }),

  login: (body: { email: string; password: string }) =>
    request<AuthResponse>("/api/auth/login", { method: "POST", body }),

  getMe: (token: string) => request<User>("/api/users/me", { token }),

  listRestaurants: () => request<Restaurant[]>("/api/restaurants"),

  getRestaurant: (id: string) => request<Restaurant>(`/api/restaurants/${id}`),

  getMenu: (restaurantId: string) =>
    request<MenuItem[]>(`/api/restaurants/${restaurantId}/menu`),

  createOrder: (
    token: string,
    body: {
      restaurant_id: string;
      items: { menu_item_id: string; quantity: number }[];
    },
  ) => request<Order>("/api/orders", { method: "POST", token, body }),

  getOrder: (token: string, id: string) =>
    request<Order>(`/api/orders/${id}`, { token }),

  listMyOrders: (token: string) =>
    request<Order[]>("/api/users/me/orders", { token }),
};
