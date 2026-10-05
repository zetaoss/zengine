<?php

namespace ZetaExtension\Auth;

use Redis;

/**
 * Connection to the persist Redis, where goapp keeps auth tokens (OTP, social login bridge).
 * Reads REDIS_PERSIST_HOST/PORT; the default port is 6379.
 */
final class PersistRedis
{
    public static function connect(): Redis
    {
        $host = getenv('REDIS_PERSIST_HOST');
        $port = (int) (getenv('REDIS_PERSIST_PORT') ?: 6379);

        $redis = new Redis;
        $redis->connect($host, $port);

        return $redis;
    }
}
