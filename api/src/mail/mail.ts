import type { Shop, ShopAlert, User } from "../repository/shops";

interface MaiLRequest {
    from?: string;
    to: string;
    subject: string;
    body: string;
}

async function sendMail(env: Env, mail: MaiLRequest) {
    mail.from = env.MAIL_FROM;

    const formData = new FormData();
    formData.append('from', mail.from);
    formData.append('to', mail.to);
    formData.append('subject', mail.subject);
    formData.append('html', mail.body);

    await fetch(`https://api.eu.mailgun.net/v3/${env.MAILGUN_DOMAIN}/messages`, {
        method: 'POST',
        body: formData,
        headers: {
            'Authorization': 'Basic ' + btoa('api:' + env.MAILGUN_KEY)
        }
    });
}

export async function sendMailConfirmToUser(env: Env, email: string, token: string) {
    const mailBody = `
        <p>Hi,</p>
        <p>Thank you for registering with us. Please click the link below to confirm your email address.</p>

        <p><a href="${env.FRONTEND_URL}/account/confirm/${token}">Confirm email</a></p>

        <p>Best regards,</p>
        <p>FriendsOfShopware</p>
    `;

    await sendMail(env, {
        to: email,
        subject: 'Confirm your email address',
        body: mailBody
    });
}

export async function sendMailResetPassword(env: Env, email: string, token: string) {
    const mailBody = `
        <p>Hi,</p>
        <p>Please click the link below to reset your password.</p>

        <p><a href="${env.FRONTEND_URL}/account/forgot-password/${token}">Reset password</a></p>

        <p>Best regards,</p>
        <p>FriendsOfShopware</p>
    `;

    await sendMail(env, {
        to: email,
        subject: 'Reset your password',
        body: mailBody
    });
}

export async function sendAlert(env: Env, shop: Shop, user: User, alert: ShopAlert) {
    await sendMail(env, {
        to: user.email,
        subject: `Shop Alert: ${alert.subject} - ${shop.name}`,
        body: `Hello ${user.username},<br><br>

        The Shop ${shop.name} has an new alert: <br><br>

        ${alert.message}.
        `
    })
}