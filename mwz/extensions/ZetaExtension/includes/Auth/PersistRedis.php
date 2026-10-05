<?php

namespace ZetaExtension\Auth;

use Redis;

/**
 * Connection to the persist Redis, where goapp keeps auth tokens (OTP, social login bridge).
 * Reads REDIS_PERSIST_HOST and connects on the fixed Redis port 6379.
 */
final class PersistRedis
{
    public static function connect(): Redis
    {
        $host = getenv('REDIS_PERSIST_HOST');

        $redis = new Redis;
        $redis->connect($host, 6379);

        return $redis;
    }
}
