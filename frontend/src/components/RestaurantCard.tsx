"use client";

import Link from "next/link";
import {
  Card,
  CardActionArea,
  CardContent,
  Chip,
  Stack,
  Typography,
} from "@mui/material";
import AccessTimeIcon from "@mui/icons-material/AccessTime";
import PlaceOutlinedIcon from "@mui/icons-material/PlaceOutlined";
import type { Restaurant } from "@/lib/types";

type RestaurantCardProps = {
  restaurant: Restaurant;
};

export const RestaurantCard = ({ restaurant }: RestaurantCardProps) => (
  <Card>
    <CardActionArea
      component={Link}
      href={`/restaurants/${restaurant.id}`}
      aria-label={`Открыть меню ${restaurant.name}`}
    >
      <CardContent>
        <Stack direction="row" justifyContent="space-between" alignItems="flex-start" sx={{ mb: 1 }}>
          <Typography variant="h6" component="h2">
            {restaurant.name}
          </Typography>
          <Chip
            size="small"
            label={restaurant.is_open ? "Открыт" : "Закрыт"}
            color={restaurant.is_open ? "success" : "default"}
          />
        </Stack>
        <Stack direction="row" spacing={0.5} alignItems="center" color="text.secondary" sx={{ mb: 0.5 }}>
          <PlaceOutlinedIcon fontSize="small" aria-hidden />
          <Typography variant="body2">{restaurant.address}</Typography>
        </Stack>
        <Stack direction="row" spacing={0.5} alignItems="center" color="text.secondary">
          <AccessTimeIcon fontSize="small" aria-hidden />
          <Typography variant="body2">
            {restaurant.open_from} – {restaurant.open_to}
          </Typography>
        </Stack>
      </CardContent>
    </CardActionArea>
  </Card>
);
