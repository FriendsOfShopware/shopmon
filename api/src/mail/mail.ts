interface MaiLRequest {
    from?: string;
    to: string;
    subject: string;
    body: string;
}

async function sendMail(env: Env, mail: MaiLRequest) {
    mail.from = env.MAIL_FROM;

    await fetch(env.MAIL_URL, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'token': env.MAIL_SECRET,
            },
            body: JSON.stringify(mail)
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