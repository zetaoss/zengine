<?php

namespace App\Services\Stat;

use App\Support\StatWindow;
use Carbon\CarbonImmutable;
use Illuminate\Support\Facades\Http;
use RuntimeException;

class CollectGscApiService
{
    private const SCOPE = 'https://www.googleapis.com/auth/webmasters.readonly';
    private const DEFAULT_TOKEN_URI = 'https://oauth2.googleapis.com/token';
    private const DEFAULT_TIMEZONE = 'America/Los_Angeles';
    private const DEFAULT_AUDIENCE = 'https://oauth2.googleapis.com/token';

    public function resolveCredentials(): array
    {
        $path = $this->env('GOOGLE_SA_FILE');
        if ($path === false || $path === '') {
            throw new RuntimeException('Missing GOOGLE_SA_FILE environment variable.');
        }

        if (! is_file($path)) {
            throw new RuntimeException("Missing Google Search Console credential file: {$path}");
        }

        $json = file_get_contents($path);
        if ($json === false) {
            throw new RuntimeException("Failed to read Google Search Console credential file: {$path}");
        }

        $config = json_decode($json, true);
        if (! is_array($config)) {
            throw new RuntimeException("Google Search Console credential file is not valid JSON: {$path}");
        }

        $siteUrl = $this->firstNonEmpty(
            $this->env('GSC_SITE_URL'),
            $config['gsc_site_url'] ?? null,
            $config['gscSiteUrl'] ?? null,
            $config['site_url'] ?? null,
            $config['siteUrl'] ?? null
        );
        $clientEmail = $config['client_email'] ?? '';
        $privateKey = $config['private_key'] ?? '';
        $tokenUri = $config['token_uri'] ?? self::DEFAULT_TOKEN_URI;
        $type = $config['type'] ?? '';

        if ($type !== '' && $type !== 'service_account') {
            throw new RuntimeException("Google Search Console credential file must be a service account key: {$path}");
        }

        if ($clientEmail === '' || $privateKey === '') {
            throw new RuntimeException(
                "Google Search Console credential file must include client_email and private_key: {$path}"
            );
        }

        if ($siteUrl === '') {
            throw new RuntimeException(
                'Missing Google Search Console site URL. Set GSC_SITE_URL '
                ."or add gsc_site_url to the credential file: {$path}"
            );
        }

        $privateKey = str_replace('\n', "\n", $privateKey);

        if (! openssl_pkey_get_private($privateKey)) {
            throw new RuntimeException("Google Search Console private_key is not a valid private key: {$path}");
        }

        return [$siteUrl, $clientEmail, $privateKey, self::DEFAULT_TIMEZONE, $tokenUri];
    }

    public function propertyDateRangeForDays(int $days): array
    {
        if ($days < 1) {
            throw new RuntimeException('--days must be an integer greater than or equal to 1.');
        }

        $todayStart = CarbonImmutable::now(self::DEFAULT_TIMEZONE)->startOfDay();

        return [$todayStart->subDays($days - 1), $todayStart];
    }

    public function propertyHourAnchoredRangeForDays(int $days): array
    {
        if ($days < 1 || $days > 10) {
            throw new RuntimeException('--days must be an integer between 1 and 10 for Search Console hourly data.');
        }

        // The query range uses an exclusive upper bound in Search Console's timezone.
        $until = StatWindow::hourlyEnd(CarbonImmutable::now(self::DEFAULT_TIMEZONE))->addHour();

        return [$until->subDays($days), $until];
    }

    public function fetchAccessToken(string $clientEmail, string $privateKey, string $tokenUri): string
    {
        $now = time();
        $assertion = $this->buildAssertion($clientEmail, $privateKey, $tokenUri, $now);

        try {
            $response = Http::asForm()
                ->acceptJson()
                ->timeout(20)
                ->post($tokenUri, [
                    'grant_type' => 'urn:ietf:params:oauth:grant-type:jwt-bearer',
                    'assertion' => $assertion,
                ]);
        } catch (\Throwable $e) {
            throw new RuntimeException('Google OAuth token request failed before response: '.$e->getMessage(), 0, $e);
        }

        if (! $response->ok()) {
            throw new RuntimeException("Google OAuth token request failed: HTTP {$response->status()} {$response->body()}");
        }

        $payload = $response->json();
        $accessToken = data_get($payload, 'access_token');
        if ($accessToken === null || $accessToken === '') {
            throw new RuntimeException('Google OAuth token response did not include access_token.');
        }

        return $accessToken;
    }

    public function query(string $accessToken, string $siteUrl, array $body): array
    {
        $encodedSiteUrl = rawurlencode($siteUrl);
        $url = "https://searchconsole.googleapis.com/webmasters/v3/sites/{$encodedSiteUrl}/searchAnalytics/query";

        try {
            $response = Http::withToken($accessToken)
                ->acceptJson()
                ->asJson()
                ->timeout(20)
                ->post($url, $body);
        } catch (\Throwable $e) {
            throw new RuntimeException('Google Search Console API request failed before response: '.$e->getMessage(), 0, $e);
        }

        if (! $response->ok()) {
            throw new RuntimeException("Google Search Console API request failed: HTTP {$response->status()} {$response->body()}");
        }

        $payload = $response->json();
        if (! is_array($payload)) {
            throw new RuntimeException('Google Search Console API returned invalid JSON.');
        }

        return $payload;
    }

    public function encodeDebugJson(array $payload): string
    {
        $json = json_encode($payload, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);

        return $json === false ? '{}' : $json;
    }

    public function normalizeDateDimension(string $raw): ?string
    {
        if (! preg_match('/^\d{4}-\d{2}-\d{2}$/', $raw)) {
            return null;
        }

        return $raw;
    }

    public function parseDateHourDimension(string $dateRaw, string $hourRaw): ?CarbonImmutable
    {
        if (! preg_match('/^\d{4}-\d{2}-\d{2}$/', $dateRaw) || ! preg_match('/^\d{1,2}$/', $hourRaw)) {
            return null;
        }

        $hour = (int) $hourRaw;
        if ($hour < 0 || $hour > 23) {
            return null;
        }

        try {
            $timeslot = CarbonImmutable::createFromFormat('Y-m-d H', "{$dateRaw} ".str_pad((string) $hour, 2, '0', STR_PAD_LEFT), self::DEFAULT_TIMEZONE);
        } catch (\Throwable) {
            return null;
        }

        if (! $timeslot) {
            return null;
        }

        return $timeslot->startOfHour();
    }

    public function parseHourDimension(string $raw): ?CarbonImmutable
    {
        try {
            $timeslot = CarbonImmutable::parse($raw, self::DEFAULT_TIMEZONE);
        } catch (\Throwable) {
            return null;
        }

        return $timeslot->setTimezone(self::DEFAULT_TIMEZONE)->startOfHour();
    }

    private function buildAssertion(string $clientEmail, string $privateKey, string $tokenUri, int $now): string
    {
        $header = $this->base64UrlEncode((string) json_encode([
            'alg' => 'RS256',
            'typ' => 'JWT',
        ], JSON_UNESCAPED_SLASHES));
        $claims = $this->base64UrlEncode((string) json_encode([
            'iss' => $clientEmail,
            'scope' => self::SCOPE,
            'aud' => $tokenUri !== '' ? $tokenUri : self::DEFAULT_AUDIENCE,
            'iat' => $now,
            'exp' => $now + 3600,
        ], JSON_UNESCAPED_SLASHES));
        $signingInput = $header.'.'.$claims;

        $signature = '';
        $privateKeyResource = openssl_pkey_get_private($privateKey);
        if (! $privateKeyResource) {
            throw new RuntimeException('Google Search Console private_key is not a valid private key.');
        }

        if (! openssl_sign($signingInput, $signature, $privateKeyResource, OPENSSL_ALGO_SHA256)) {
            throw new RuntimeException('Failed to sign Google service account JWT assertion.');
        }

        return $signingInput.'.'.$this->base64UrlEncode($signature);
    }

    private function base64UrlEncode(string $value): string
    {
        return rtrim(strtr(base64_encode($value), '+/', '-_'), '=');
    }

    private function env(string $key): string|false
    {
        $value = getenv($key);
        if ($value !== false) {
            return $value;
        }

        $serverValue = $_SERVER[$key] ?? $_ENV[$key] ?? false;

        return is_string($serverValue) ? $serverValue : false;
    }

    private function firstNonEmpty(string|false|null ...$values): string
    {
        foreach ($values as $value) {
            if (is_string($value) && $value !== '') {
                return $value;
            }
        }

        return '';
    }
}
