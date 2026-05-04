<?php

return [
    'postmark' => [
        'token' => env('POSTMARK_TOKEN'),
    ],
    'ses' => [
        'key' => env('AWS_ACCESS_KEY_ID'),
        'secret' => env('AWS_SECRET_ACCESS_KEY'),
        'region' => env('AWS_DEFAULT_REGION', 'us-east-1'),
    ],
    'resend' => [
        'key' => env('RESEND_KEY'),
    ],
    'slack' => [
        'notifications' => [
            'bot_user_oauth_token' => env('SLACK_BOT_USER_OAUTH_TOKEN'),
            'channel' => env('SLACK_BOT_USER_DEFAULT_CHANNEL'),
        ],
    ],
    'facebook' => [
        'client_id' => env('FACEBOOK_CLIENT_ID'),
        'client_secret' => getenv('FACEBOOK_CLIENT_SECRET'),
        'redirect' => env('APP_URL').'/auth/callback/facebook',
    ],
    'github' => [
        'client_id' => env('GITHUB_CLIENT_ID'),
        'client_secret' => env('GITHUB_CLIENT_SECRET'),
        'redirect' => env('APP_URL').'/auth/callback/github',
    ],
    'google' => [
        'client_id' => env('GOOGLE_CLIENT_ID'),
        'client_secret' => env('GOOGLE_CLIENT_SECRET'),
        'redirect' => env('APP_URL').'/auth/callback/google',
    ],
    'google_analytics' => [
        'timezone' => env('GA_TIMEZONE', 'UTC'),
    ],
    'cloudflare' => [
        'api_token' => env('CLOUDFLARE_API_TOKEN'),
        'zone_id' => env('CLOUDFLARE_ZONE_ID'),
    ],
    'mediawiki' => [
        'api_server' => env('API_SERVER'),
    ],
    'internal' => [
        'secret_key' => env('INTERNAL_SECRET_KEY'),
    ],
    'docfac' => [
        'interval_seconds' => (int) env('DOCFAC_INTERVAL_SECONDS'),
        'retry_interval_seconds' => (int) env('DOCFAC_RETRY_INTERVAL_SECONDS'),
        'auto_enqueue' => filter_var(env('DOCFAC_AUTO_ENQUEUE', false), FILTER_VALIDATE_BOOL),
    ],
    'llm' => [
        'enabled' => filter_var(env('LLM_ENABLED', false), FILTER_VALIDATE_BOOL),
        'endpoint' => env('LLM_ENDPOINT'),
        'model' => env('LLM_MODEL'),
        'timeout_seconds' => (int) env('LLM_TIMEOUT_SECONDS'),
    ],
];
