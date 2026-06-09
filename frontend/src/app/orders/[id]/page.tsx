"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import {
  Alert,
  Box,
  Breadcrumbs,
  Chip,
  CircularProgress,
  Container,
  Divider,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import { api } from "@/lib/api";
import type { Order } from "@/lib/types";
import { useAuth } from "@/contexts/AuthContext";
import { OrderStatusStepper } from "@/components/OrderStatusStepper";

const TERMINAL_STATUSES = new Set(["DELIVERED", "CANCELLED"]);

export default function OrderDetailPage() {
  const params = useParams<{ id: string }>();
  const { token, isLoading: authLoading } = useAuth();
  const router = useRouter();
  const [order, setOrder] = useState<Order | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const fetchOrder = useCallback(async () => {
    if (!token || !params.id) return;
    try {
      const data = await api.getOrder(token, params.id);
      setOrder(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка загрузки");
    } finally {
      setIsLoading(false);
    }
  }, [token, params.id]);

  useEffect(() => {
    if (authLoading) return;
    if (!token) {
      router.push("/login");
      return;
    }
    fetchOrder();
  }, [token, authLoading, router, fetchOrder]);

  useEffect(() => {
    if (!order || TERMINAL_STATUSES.has(order.status)) return;

    const interval = setInterval(fetchOrder, 3000);
    return () => clearInterval(interval);
  }, [order, fetchOrder]);

  if (isLoading) {
    return (
      <Container sx={{ py: 6 }}>
        <CircularProgress aria-label="Загрузка заказа" />
      </Container>
    );
  }

  if (error || !order) {
    return (
      <Container sx={{ py: 6 }}>
        <Alert severity="error">{error ?? "Заказ не найден"}</Alert>
      </Container>
    );
  }

  return (
    <Container maxWidth="md" sx={{ py: 4 }}>
      <Breadcrumbs sx={{ mb: 2 }}>
        <Link href="/orders">Заказы</Link>
        <Typography color="text.primary">#{order.id.slice(0, 8)}</Typography>
      </Breadcrumbs>

      <Paper sx={{ p: 3 }}>
        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
          <Typography variant="h5" component="h1">
            Заказ #{order.id.slice(0, 8)}
          </Typography>
          {!TERMINAL_STATUSES.has(order.status) && (
            <Chip label="Обновляется…" size="small" color="info" variant="outlined" />
          )}
        </Stack>

        <OrderStatusStepper status={order.status} />

        <Divider sx={{ my: 2 }} />

        <Typography variant="subtitle1" fontWeight={600} sx={{ mb: 1 }}>
          Позиции
        </Typography>
        <Stack spacing={1} sx={{ mb: 2 }}>
          {order.items.map((item) => (
            <Stack
              key={item.menu_item_id}
              direction="row"
              justifyContent="space-between"
            >
              <Typography>
                {item.name} × {item.quantity}
              </Typography>
              <Typography>${(item.price * item.quantity).toFixed(2)}</Typography>
            </Stack>
          ))}
        </Stack>

        <Stack direction="row" justifyContent="space-between" sx={{ mb: 2 }}>
          <Typography variant="h6">Итого</Typography>
          <Typography variant="h6" color="primary.main">
            ${order.total_amount.toFixed(2)}
          </Typography>
        </Stack>

        {order.payment && (
          <Box sx={{ bgcolor: "action.hover", p: 2, borderRadius: 2 }}>
            <Typography variant="subtitle2" gutterBottom>
              Оплата
            </Typography>
            <Typography variant="body2">
              Статус: {order.payment.status} · {order.payment.method}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              TX: {order.payment.transaction_id}
            </Typography>
          </Box>
        )}

        {order.courier_id && (
          <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
            Курьер назначен
          </Typography>
        )}
      </Paper>
    </Container>
  );
}
