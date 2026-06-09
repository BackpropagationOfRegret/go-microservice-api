"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Alert,
  Box,
  Chip,
  CircularProgress,
  Container,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import { api } from "@/lib/api";
import type { Order, OrderStatus } from "@/lib/types";
import { useAuth } from "@/contexts/AuthContext";

const STATUS_COLOR: Record<OrderStatus, "default" | "primary" | "success" | "error" | "warning"> = {
  PENDING: "default",
  PAID: "primary",
  PREPARING: "warning",
  READY: "warning",
  DELIVERING: "primary",
  DELIVERED: "success",
  CANCELLED: "error",
};

const STATUS_LABEL: Record<OrderStatus, string> = {
  PENDING: "Создан",
  PAID: "Оплачен",
  PREPARING: "Готовится",
  READY: "Готов",
  DELIVERING: "В пути",
  DELIVERED: "Доставлен",
  CANCELLED: "Отменён",
};

export default function OrdersPage() {
  const { token, isLoading: authLoading } = useAuth();
  const router = useRouter();
  const [orders, setOrders] = useState<Order[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (authLoading) return;
    if (!token) {
      router.push("/login");
      return;
    }

    api
      .listMyOrders(token)
      .then(setOrders)
      .catch((err: Error) => setError(err.message))
      .finally(() => setIsLoading(false));
  }, [token, authLoading, router]);

  return (
    <Container maxWidth="md" sx={{ py: 4 }}>
      <Typography variant="h4" component="h1" sx={{ mb: 3 }}>
        Мои заказы
      </Typography>

      {isLoading && <CircularProgress aria-label="Загрузка заказов" />}
      {error && <Alert severity="error">{error}</Alert>}

      {!isLoading && !error && orders.length === 0 && (
        <Typography color="text.secondary">
          Заказов пока нет.{" "}
          <Link href="/restaurants">Выбрать ресторан</Link>
        </Typography>
      )}

      <Stack spacing={2}>
        {orders.map((order) => (
          <Paper
            key={order.id}
            component={Link}
            href={`/orders/${order.id}`}
            sx={{
              p: 2.5,
              textDecoration: "none",
              color: "inherit",
              "&:hover": { bgcolor: "action.hover" },
            }}
          >
            <Stack direction="row" justifyContent="space-between" alignItems="center">
              <Box>
                <Typography fontWeight={600}>
                  Заказ #{order.id.slice(0, 8)}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {new Date(order.created_at).toLocaleString("ru-RU")} · $
                  {order.total_amount.toFixed(2)}
                </Typography>
              </Box>
              <Chip
                label={STATUS_LABEL[order.status]}
                color={STATUS_COLOR[order.status]}
                size="small"
              />
            </Stack>
          </Paper>
        ))}
      </Stack>
    </Container>
  );
}
