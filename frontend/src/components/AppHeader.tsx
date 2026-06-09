"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  AppBar,
  Badge,
  Box,
  Button,
  Container,
  IconButton,
  Toolbar,
  Typography,
} from "@mui/material";
import RestaurantMenuIcon from "@mui/icons-material/RestaurantMenu";
import ShoppingBagOutlinedIcon from "@mui/icons-material/ShoppingBagOutlined";
import ReceiptLongOutlinedIcon from "@mui/icons-material/ReceiptLongOutlined";
import LogoutOutlinedIcon from "@mui/icons-material/LogoutOutlined";
import { useAuth } from "@/contexts/AuthContext";
import { useCart } from "@/contexts/CartContext";

export const AppHeader = () => {
  const { user, logout, token } = useAuth();
  const { itemCount } = useCart();
  const pathname = usePathname();
  const router = useRouter();

  const handleLogout = () => {
    logout();
    router.push("/");
  };

  return (
    <AppBar position="sticky" color="inherit" elevation={0} sx={{ borderBottom: 1, borderColor: "divider" }}>
      <Container maxWidth="lg">
        <Toolbar disableGutters sx={{ gap: 1 }}>
          <RestaurantMenuIcon color="primary" aria-hidden />
          <Typography
            component={Link}
            href="/"
            variant="h6"
            color="text.primary"
            sx={{ flexGrow: 1, textDecoration: "none", fontWeight: 700 }}
          >
            FoodExpress
          </Typography>

          <Button
            component={Link}
            href="/restaurants"
            color={pathname.startsWith("/restaurants") ? "primary" : "inherit"}
          >
            Рестораны
          </Button>

          {token && (
            <Button
              component={Link}
              href="/orders"
              startIcon={<ReceiptLongOutlinedIcon />}
              color={pathname.startsWith("/orders") ? "primary" : "inherit"}
            >
              Заказы
            </Button>
          )}

          {token && itemCount > 0 && (
            <IconButton
              component={Link}
              href="/cart"
              aria-label={`Корзина, ${itemCount} позиций`}
              color="primary"
            >
              <Badge badgeContent={itemCount} color="secondary">
                <ShoppingBagOutlinedIcon />
              </Badge>
            </IconButton>
          )}

          {user ? (
            <Box sx={{ display: "flex", alignItems: "center", gap: 1, ml: 1 }}>
              <Typography variant="body2" color="text.secondary">
                {user.name}
              </Typography>
              <IconButton onClick={handleLogout} aria-label="Выйти" size="small">
                <LogoutOutlinedIcon fontSize="small" />
              </IconButton>
            </Box>
          ) : (
            <Box sx={{ display: "flex", gap: 1, ml: 1 }}>
              <Button component={Link} href="/login" variant="outlined" size="small">
                Войти
              </Button>
              <Button component={Link} href="/register" variant="contained" size="small">
                Регистрация
              </Button>
            </Box>
          )}
        </Toolbar>
      </Container>
    </AppBar>
  );
};
