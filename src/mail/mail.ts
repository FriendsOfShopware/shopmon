interface MaiLRequest {
    from?: string;
    to: string;
    subject: string;
    body: string;
}

async function sendMail(mail: MaiLRequest) {
    mail.from = MAIL_FROM;

    await fetch(MAIL_URL, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'token': MAIL_SECRET,
            },
            body: JSON.stringify(mail)
        });
}

export async function sendMailConfirmToUser(email: string, token: string) {
    const mailBody = `
        <p>Hi,</p>
        <p>Thank you for registering with us. Please click the link below to confirm your email address.</p>

        <p><a href="${FRONTEND_URL}/confirm/${token}">Confirm email</a></p>

        <p>Best regards,</p>
        <p>FriendsOfShopware</p>
    `;

    await sendMail({
        to: email,
        subject: 'Confirm your email address',
        body: mailBody
    });
}