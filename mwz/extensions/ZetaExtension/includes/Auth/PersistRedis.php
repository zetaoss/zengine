<?php

namespace ZetaExtension\Auth;

use Redis;

/**
 * Connection to the persist Redis, where goapp keeps auth tokens (OTP, social login bridge).
 * Reads REDIS_PERSIST_HOST/PORT and falls back to the legacy REDIS_HOST/PORT.
 */
final class PersistRedis
{
    public static function connect(): Redis
    {
        $host = getenv('REDIS_PERSIST_HOST') ?: getenv('REDIS_HOST');
        $port = (int) (getenv('REDIS_PERSIST_PORT') ?: getenv('REDIS_PORT') ?: 6379);

        $redis = new Redis;
        $redis->connect($host, $port);

        return $redis;
    }
}
