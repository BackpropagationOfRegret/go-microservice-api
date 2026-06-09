export type User = {
  id: string;
  email: string;
  name: string;
  phone?: string;
  created_at?: string;
};

export type Restaurant = {
  id: string;
  name: string;
  address: string;
  is_open: boolean;
  open_from: string;
  open_to: string;
};

export type MenuItem = {
  id: string;
  restaurant_id: string;
  name: string;
  description: string;
  price: number;
  available: boolean;
};

export type OrderItem = {
  menu_item_id: string;
  name: string;
  quantity: number;
  price: number;
};

export type OrderStatus =
  | "PENDING"
  | "PAID"
  | "PREPARING"
  | "READY"
  | "DELIVERING"
  | "DELIVERED"
  | "CANCELLED"
  | "REFUNDED";

export type Payment = {
  id: string;
  order_id: string;
  amount: number;
  method: string;
  status: string;
  transaction_id: string;
  created_at: string;
};

export type Order = {
  id: string;
  user_id: string;
  restaurant_id: string;
  status: OrderStatus;
  total_amount: number;
  items: OrderItem[];
  courier_id?: string;
  created_at: string;
  updated_at: string;
  payment?: Payment;
};

export type AuthResponse = {
  token: string;
  user: User;
};

export type CartItem = {
  menuItem: MenuItem;
  quantity: number;
};
