import type { ConnectionOptions } from "bullmq";

let connection: ConnectionOptions | undefined;

export function getRedisConnection(): ConnectionOptions {
  if (connection) {
    return connection;
  }

  const redisUrl = process.env.REDIS_URL || "redis://localhost:6379";
  const url = new URL(redisUrl);

  connection = {
    host: url.hostname,
    port: Number(url.port) || 6379,
    password: url.password || undefined,
    db: url.pathname ? Number(url.pathname.slice(1)) || 0 : 0,
  };

  return connection;
}
