import { sendMail } from "#src/modules/shared/mail";
import type { Shop, User } from "#src/modules/shop/shop.repository.ts";
import type { ShopAlert } from "#src/modules/shop/shop.service.ts";
import alertTemplate from "./alert.js";

export async function sendAlert(shop: Shop, user: User, alert: ShopAlert) {
  await sendMail({
    to: user.email,
    subject: `Shop Alert: ${alert.subject} - ${shop.name}`,
    body: alertTemplate({
      FRONTEND_URL: process.env.FRONTEND_URL,
      shop,
      alert,
      user,
    }),
  });
}
