"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Alert,
  Box,
  Button,
  Container,
  Divider,
  IconButton,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import RemoveIcon from "@mui/icons-material/Remove";
import DeleteOutlineOutlinedIcon from "@mui/icons-material/DeleteOutlineOutlined";
import { useAuth } from "@/contexts/AuthContext";
import { useCart } from "@/contexts/CartContext";
import { api } from "@/lib/api";

export default function CartPage() {
  const { token } = useAuth();
  const router = useRouter();
  const { items, total, restaurantId, updateQuantity, clearCart } = useCart();
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleCheckout = async () => {
    if (!token) {
      router.push("/login");
      return;
    }
    if (!restaurantId || items.length === 0) return;

    setError(null);
    setIsSubmitting(true);
    try {
      const order = await api.createOrder(token, {
        restaurant_id: restaurantId,
        items: items.map((i) => ({
          menu_item_id: i.menuItem.id,
          quantity: i.quantity,
        })),
      });
      clearCart();
      router.push(`/orders/${order.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось оформить заказ");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (items.length === 0) {
    return (
      <Container maxWidth="sm" sx={{ py: 8, textAlign: "center" }}>
        <Typography variant="h5" sx={{ mb: 2 }}>
          Корзина пуста
        </Typography>
        <Button component={Link} href="/restaurants" variant="contained">
          К ресторанам
        </Button>
      </Container>
    );
  }

  return (
    <Container maxWidth="md" sx={{ py: 4 }}>
      <Typography variant="h4" component="h1" sx={{ mb: 3 }}>
        Корзина
      </Typography>

      <Paper sx={{ p: 3 }}>
        <Stack spacing={2}>
          {items.map((item) => (
            <Box key={item.menuItem.id}>
              <Stack direction="row" justifyContent="space-between" alignItems="center">
                <Box>
                  <Typography fontWeight={600}>{item.menuItem.name}</Typography>
                  <Typography variant="body2" color="text.secondary">
                    ${item.menuItem.price.toFixed(2)} × {item.quantity}
                  </Typography>
                </Box>
                <Stack direction="row" alignItems="center" spacing={1}>
                  <IconButton
                    size="small"
                    onClick={() => updateQuantity(item.menuItem.id, item.quantity - 1)}
                    aria-label={`Уменьшить ${item.menuItem.name}`}
                  >
                    <RemoveIcon fontSize="small" />
                  </IconButton>
                  <Typography>{item.quantity}</Typography>
                  <IconButton
                    size="small"
                    onClick={() => updateQuantity(item.menuItem.id, item.quantity + 1)}
                    aria-label={`Увеличить ${item.menuItem.name}`}
                  >
                    <AddIcon fontSize="small" />
                  </IconButton>
                  <IconButton
                    size="small"
                    onClick={() => updateQuantity(item.menuItem.id, 0)}
                    aria-label={`Удалить ${item.menuItem.name}`}
                  >
                    <DeleteOutlineOutlinedIcon fontSize="small" />
                  </IconButton>
                  <Typography fontWeight={600} sx={{ minWidth: 72, textAlign: "right" }}>
                    ${(item.menuItem.price * item.quantity).toFixed(2)}
                  </Typography>
                </Stack>
              </Stack>
              <Divider sx={{ mt: 2 }} />
            </Box>
          ))}

          <Stack direction="row" justifyContent="space-between" alignItems="center">
            <Typography variant="h5">Итого</Typography>
            <Typography variant="h5" color="primary.main">
              ${total.toFixed(2)}
            </Typography>
          </Stack>

          {error && <Alert severity="error">{error}</Alert>}

          <Button
            variant="contained"
            size="large"
            onClick={handleCheckout}
            disabled={isSubmitting}
          >
            {isSubmitting ? "Оформляем…" : "Оформить заказ"}
          </Button>
        </Stack>
      </Paper>
    </Container>
  );
}
