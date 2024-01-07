export async function md5(data: string): Promise<string> {
    const msgUint8 = new TextEncoder().encode(data)
    const hashBuffer = await crypto.subtle.digest('MD5', msgUint8)
    const hashArray = Array.from(new Uint8Array(hashBuffer))
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}

export async function encrypt(key: string, payload: string) {
    const secretKey = new TextEncoder().encode(key);

    const cryptoKey = await crypto.subtle.importKey('raw', secretKey, 'AES-GCM', false, ['encrypt']);

    const clientSecret = new TextEncoder().encode(payload);
    const iv = crypto.getRandomValues(new Uint8Array(16));

    const encryptedShit = await crypto.subtle.encrypt({ name: 'AES-GCM', iv: iv }, cryptoKey, clientSecret);

    const buffer = new Uint8Array(encryptedShit);

    return uintToBase64(iv) + ':' + uintToBase64(buffer);
}

function uintToBase64(uint: Uint8Array) {
    return btoa(String.fromCharCode(...uint));
}

function base64ToUint(base64: string) {
    return new Uint8Array(atob(base64).split('').map((c) => c.charCodeAt(0)));
}

export async function decrypt(key: string, payload: string) {
    const secretKey = new TextEncoder().encode(key);

    const cryptoKey = await crypto.subtle.importKey('raw', secretKey, 'AES-GCM', false, ['decrypt']);

    const [iv, encrypted] = payload.split(':').map(base64ToUint);

    const decrypted = await crypto.subtle.decrypt({ name: 'AES-GCM', iv: iv }, cryptoKey, encrypted);

    return new TextDecoder().decode(decrypted);
}
