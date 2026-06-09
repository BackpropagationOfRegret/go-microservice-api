"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import {
  Alert,
  Box,
  Breadcrumbs,
  Button,
  Chip,
  CircularProgress,
  Container,
  Grid,
  Stack,
  Typography,
} from "@mui/material";
import ShoppingBagOutlinedIcon from "@mui/icons-material/ShoppingBagOutlined";
import { api } from "@/lib/api";
import type { MenuItem, Restaurant } from "@/lib/types";
import { MenuItemCard } from "@/components/MenuItemCard";
import { useCart } from "@/contexts/CartContext";

export default function RestaurantPage() {
  const params = useParams<{ id: string }>();
  const { itemCount } = useCart();
  const [restaurant, setRestaurant] = useState<Restaurant | null>(null);
  const [menu, setMenu] = useState<MenuItem[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (!params.id) return;

    Promise.all([api.getRestaurant(params.id), api.getMenu(params.id)])
      .then(([rest, items]) => {
        setRestaurant(rest);
        setMenu(items);
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setIsLoading(false));
  }, [params.id]);

  if (isLoading) {
    return (
      <Container sx={{ py: 6 }}>
        <CircularProgress aria-label="Загрузка меню" />
      </Container>
    );
  }

  if (error || !restaurant) {
    return (
      <Container sx={{ py: 6 }}>
        <Alert severity="error">{error ?? "Ресторан не найден"}</Alert>
      </Container>
    );
  }

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Breadcrumbs sx={{ mb: 2 }}>
        <Link href="/restaurants">Рестораны</Link>
        <Typography color="text.primary">{restaurant.name}</Typography>
      </Breadcrumbs>

      <Stack
        direction={{ xs: "column", sm: "row" }}
        justifyContent="space-between"
        alignItems={{ xs: "flex-start", sm: "center" }}
        spacing={2}
        sx={{ mb: 4 }}
      >
        <Box>
          <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }}>
            <Typography variant="h4" component="h1">
              {restaurant.name}
            </Typography>
            <Chip
              size="small"
              label={restaurant.is_open ? "Открыт" : "Закрыт"}
              color={restaurant.is_open ? "success" : "default"}
            />
          </Stack>
          <Typography color="text.secondary">{restaurant.address}</Typography>
        </Box>

        {itemCount > 0 && (
          <Button
            component={Link}
            href="/cart"
            variant="contained"
            startIcon={<ShoppingBagOutlinedIcon />}
          >
            Корзина ({itemCount})
          </Button>
        )}
      </Stack>

      <Grid container spacing={3}>
        {menu.map((item) => (
          <Grid key={item.id} item xs={12} sm={6} md={4}>
            <MenuItemCard item={item} restaurantName={restaurant.name} />
          </Grid>
        ))}
      </Grid>
    </Container>
  );
}
