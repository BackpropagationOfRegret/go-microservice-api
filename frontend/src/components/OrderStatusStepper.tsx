"use client";

import {
  Step,
  StepLabel,
  Stepper,
  Typography,
} from "@mui/material";
import type { OrderStatus } from "@/lib/types";

const STEPS: { status: OrderStatus; label: string }[] = [
  { status: "PENDING", label: "Создан" },
  { status: "PAID", label: "Оплачен" },
  { status: "PREPARING", label: "Готовится" },
  { status: "READY", label: "Готов" },
  { status: "DELIVERING", label: "В пути" },
  { status: "DELIVERED", label: "Доставлен" },
];

const STATUS_ORDER: Record<OrderStatus, number> = {
  PENDING: 0,
  PAID: 1,
  PREPARING: 2,
  READY: 3,
  DELIVERING: 4,
  DELIVERED: 5,
  CANCELLED: -1,
};

type OrderStatusStepperProps = {
  status: OrderStatus;
};

export const OrderStatusStepper = ({ status }: OrderStatusStepperProps) => {
  if (status === "CANCELLED") {
    return (
      <Typography color="error" fontWeight={600}>
        Заказ отменён
      </Typography>
    );
  }

  const activeStep = STATUS_ORDER[status] ?? 0;

  return (
    <Stepper activeStep={activeStep} alternativeLabel sx={{ mt: 2, mb: 2 }}>
      {STEPS.map((step) => (
        <Step key={step.status} completed={STATUS_ORDER[step.status] <= activeStep}>
          <StepLabel>{step.label}</StepLabel>
        </Step>
      ))}
    </Stepper>
  );
};
