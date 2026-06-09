"use client";

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";
import type { CartItem, MenuItem } from "@/lib/types";

type CartContextValue = {
  restaurantId: string | null;
  restaurantName: string | null;
  items: CartItem[];
  total: number;
  itemCount: number;
  setRestaurant: (id: string, name: string) => void;
  addItem: (menuItem: MenuItem) => void;
  updateQuantity: (menuItemId: string, quantity: number) => void;
  clearCart: () => void;
};

const CartContext = createContext<CartContextValue | null>(null);

export const CartProvider = ({ children }: { children: React.ReactNode }) => {
  const [restaurantId, setRestaurantId] = useState<string | null>(null);
  const [restaurantName, setRestaurantName] = useState<string | null>(null);
  const [items, setItems] = useState<CartItem[]>([]);

  const setRestaurant = useCallback((id: string, name: string) => {
    setRestaurantId(id);
    setRestaurantName(name);
  }, []);

  const addItem = useCallback(
    (menuItem: MenuItem) => {
      if (restaurantId && restaurantId !== menuItem.restaurant_id) {
        setItems([{ menuItem, quantity: 1 }]);
        setRestaurantId(menuItem.restaurant_id);
        return;
      }

      if (!restaurantId) {
        setRestaurantId(menuItem.restaurant_id);
      }

      setItems((prev) => {
        const existing = prev.find((i) => i.menuItem.id === menuItem.id);
        if (existing) {
          return prev.map((i) =>
            i.menuItem.id === menuItem.id
              ? { ...i, quantity: i.quantity + 1 }
              : i,
          );
        }
        return [...prev, { menuItem, quantity: 1 }];
      });
    },
    [restaurantId],
  );

  const updateQuantity = useCallback((menuItemId: string, quantity: number) => {
    if (quantity <= 0) {
      setItems((prev) => prev.filter((i) => i.menuItem.id !== menuItemId));
      return;
    }
    setItems((prev) =>
      prev.map((i) =>
        i.menuItem.id === menuItemId ? { ...i, quantity } : i,
      ),
    );
  }, []);

  const clearCart = useCallback(() => {
    setItems([]);
    setRestaurantId(null);
    setRestaurantName(null);
  }, []);

  const total = useMemo(
    () => items.reduce((sum, i) => sum + i.menuItem.price * i.quantity, 0),
    [items],
  );

  const itemCount = useMemo(
    () => items.reduce((sum, i) => sum + i.quantity, 0),
    [items],
  );

  const value = useMemo(
    () => ({
      restaurantId,
      restaurantName,
      items,
      total,
      itemCount,
      setRestaurant,
      addItem,
      updateQuantity,
      clearCart,
    }),
    [
      restaurantId,
      restaurantName,
      items,
      total,
      itemCount,
      setRestaurant,
      addItem,
      updateQuantity,
      clearCart,
    ],
  );

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
};

export const useCart = () => {
  const ctx = useContext(CartContext);
  if (!ctx) {
    throw new Error("useCart must be used within CartProvider");
  }
  return ctx;
};
