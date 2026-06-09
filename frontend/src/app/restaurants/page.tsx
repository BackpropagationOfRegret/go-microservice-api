"use client";

import { useEffect, useState } from "react";
import {
  Alert,
  CircularProgress,
  Container,
  Grid,
  Typography,
} from "@mui/material";
import { api } from "@/lib/api";
import type { Restaurant } from "@/lib/types";
import { RestaurantCard } from "@/components/RestaurantCard";

export default function RestaurantsPage() {
  const [restaurants, setRestaurants] = useState<Restaurant[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    api
      .listRestaurants()
      .then(setRestaurants)
      .catch((err: Error) => setError(err.message))
      .finally(() => setIsLoading(false));
  }, []);

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Typography variant="h4" component="h1" sx={{ mb: 3 }}>
        Рестораны
      </Typography>

      {isLoading && (
        <CircularProgress aria-label="Загрузка ресторанов" />
      )}

      {error && <Alert severity="error">{error}</Alert>}

      {!isLoading && !error && (
        <Grid container spacing={3}>
          {restaurants.map((restaurant) => (
            <Grid key={restaurant.id} item xs={12} sm={6} md={4}>
              <RestaurantCard restaurant={restaurant} />
            </Grid>
          ))}
        </Grid>
      )}
    </Container>
  );
}
