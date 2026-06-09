"use client";

import {
  Box,
  Button,
  Card,
  CardContent,
  IconButton,
  Stack,
  Typography,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import RemoveIcon from "@mui/icons-material/Remove";
import type { MenuItem } from "@/lib/types";
import { useCart } from "@/contexts/CartContext";

type MenuItemCardProps = {
  item: MenuItem;
  restaurantName: string;
};

export const MenuItemCard = ({ item, restaurantName }: MenuItemCardProps) => {
  const { items, addItem, updateQuantity, setRestaurant } = useCart();
  const cartItem = items.find((i) => i.menuItem.id === item.id);
  const quantity = cartItem?.quantity ?? 0;

  const handleAdd = () => {
    setRestaurant(item.restaurant_id, restaurantName);
    addItem(item);
  };

  const handleDecrease = () => {
    updateQuantity(item.id, quantity - 1);
  };

  return (
    <Card sx={{ height: "100%" }}>
      <CardContent>
        <Stack spacing={1} height="100%">
          <Typography variant="h6" component="h3">
            {item.name}
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ flexGrow: 1 }}>
            {item.description}
          </Typography>
          <Stack direction="row" justifyContent="space-between" alignItems="center">
            <Typography variant="h6" color="primary.main">
              ${item.price.toFixed(2)}
            </Typography>
            {quantity === 0 ? (
              <Button
                variant="contained"
                size="small"
                startIcon={<AddIcon />}
                onClick={handleAdd}
                disabled={!item.available}
                aria-label={`Добавить ${item.name} в корзину`}
              >
                В корзину
              </Button>
            ) : (
              <Box sx={{ display: "flex", alignItems: "center", gap: 0.5 }}>
                <IconButton
                  size="small"
                  onClick={handleDecrease}
                  aria-label={`Уменьшить количество ${item.name}`}
                >
                  <RemoveIcon fontSize="small" />
                </IconButton>
                <Typography aria-live="polite" sx={{ minWidth: 24, textAlign: "center" }}>
                  {quantity}
                </Typography>
                <IconButton
                  size="small"
                  onClick={handleAdd}
                  aria-label={`Увеличить количество ${item.name}`}
                >
                  <AddIcon fontSize="small" />
                </IconButton>
              </Box>
            )}
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  );
};
