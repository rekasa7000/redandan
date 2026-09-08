import { authenticator } from "otplib";
import QRCode from "qrcode";

export function generateSecret(): string {
  return authenticator.generateSecret(32);
}

export function getKeyUri(email: string, secret: string): string {
  return authenticator.keyuri(email, "Reliva", secret);
}

export async function generateQRDataURL(uri: string): Promise<string> {
  return QRCode.toDataURL(uri, { width: 256, margin: 2 });
}

export function verifyCode(code: string, secret: string): boolean {
  return authenticator.verify({ token: code, secret });
}
