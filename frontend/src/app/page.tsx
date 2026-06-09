"use client";

import Link from "next/link";
import {
  Box,
  Button,
  Container,
  Stack,
  Typography,
} from "@mui/material";
import DeliveryDiningIcon from "@mui/icons-material/DeliveryDining";
import RestaurantIcon from "@mui/icons-material/Restaurant";
import SpeedIcon from "@mui/icons-material/Speed";

export default function HomePage() {
  return (
    <Box
      sx={{
        background: "linear-gradient(135deg, #fff8f3 0%, #ffe8d6 100%)",
        minHeight: "calc(100vh - 64px)",
      }}
    >
      <Container maxWidth="lg" sx={{ py: { xs: 6, md: 10 } }}>
        <Stack spacing={4} alignItems="flex-start" maxWidth={640}>
          <Typography variant="h3" component="h1" color="text.primary">
            Еда из любимых ресторанов — с доставкой за минуты
          </Typography>
          <Typography variant="h6" color="text.secondary" fontWeight={400}>
            Микросервисная платформа: заказ, оплата, кухня и курьер работают
            асинхронно через Kafka. Следи за статусом в реальном времени.
          </Typography>
          <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
            <Button
              component={Link}
              href="/restaurants"
              variant="contained"
              size="large"
            >
              Выбрать ресторан
            </Button>
            <Button component={Link} href="/register" variant="outlined" size="large">
              Создать аккаунт
            </Button>
          </Stack>
        </Stack>

        <Stack
          direction={{ xs: "column", md: "row" }}
          spacing={3}
          sx={{ mt: 8 }}
        >
          {[
            {
              icon: <RestaurantIcon color="primary" fontSize="large" />,
              title: "Рестораны рядом",
              text: "Меню и цены в реальном времени",
            },
            {
              icon: <SpeedIcon color="primary" fontSize="large" />,
              title: "Быстрая оплата",
              text: "Автоматическая обработка платежа",
            },
            {
              icon: <DeliveryDiningIcon color="primary" fontSize="large" />,
              title: "Трекинг доставки",
              text: "От кухни до вашей двери",
            },
          ].map((feature) => (
            <Box
              key={feature.title}
              sx={{
                flex: 1,
                p: 3,
                bgcolor: "background.paper",
                borderRadius: 3,
              }}
            >
              {feature.icon}
              <Typography variant="h6" sx={{ mt: 1, mb: 0.5 }}>
                {feature.title}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {feature.text}
              </Typography>
            </Box>
          ))}
        </Stack>
      </Container>
    </Box>
  );
}
