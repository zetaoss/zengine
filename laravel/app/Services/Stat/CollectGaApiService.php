<?php

namespace App\Services\Stat;

use App\Support\StatWindow;
use Carbon\CarbonImmutable;
use Illuminate\Support\Facades\Http;
use RuntimeException;

class CollectGaApiService
{
    private const SCOPE = 'https://www.googleapis.com/auth/analytics.readonly';
    private const DEFAULT_TOKEN_URI = 'https://oauth2.googleapis.com/token';
    private const DEFAULT_TIMEZONE = 'UTC';
    private const DEFAULT_AUDIENCE = 'https://oauth2.googleapis.com/token';

    public function resolveCredentials(): array
    {
        $path = $this->env('GOOGLE_SA_FILE');
        if ($path === false || $path === '') {
            throw new RuntimeException('Missing GOOGLE_SA_FILE environment variable.');
        }

        if (! is_file($path)) {
            throw new RuntimeException("Missing Google Analytics credential file: {$path}");
        }

        $json = file_get_contents($path);
        if ($json === false) {
            throw new RuntimeException("Failed to read Google Analytics credential file: {$path}");
        }

        $config = json_decode($json, true);
        if (! is_array($config)) {
            throw new RuntimeException("Google Analytics credential file is not valid JSON: {$path}");
        }

        $propertyId = $this->firstNonEmpty(
            $config['property_id'] ?? null,
            $config['propertyId'] ?? null,
            $this->env('GA_PROPERTY_ID')
        );
        $clientEmail = $config['client_email'] ?? '';
        $privateKey = $config['private_key'] ?? '';
        $timezone = $this->firstNonEmpty(
            $config['timezone'] ?? null,
            $config['time_zone'] ?? null,
            self::DEFAULT_TIMEZONE
        );
        $tokenUri = $config['token_uri'] ?? self::DEFAULT_TOKEN_URI;
        $type = $config['type'] ?? '';

        if ($type !== '' && $type !== 'service_account') {
            throw new RuntimeException("Google Analytics credential file must be a service account key: {$path}");
        }

        if ($clientEmail === '' || $privateKey === '') {
            throw new RuntimeException(
                "Google Analytics credential file must include client_email and private_key: {$path}"
            );
        }

        if ($propertyId === '') {
            throw new RuntimeException(
                'Missing Google Analytics property ID. Set GA_PROPERTY_ID '
                ."or add property_id to the credential file: {$path}"
            );
        }

        if (! preg_match('/^\d+$/', (string) $propertyId)) {
            throw new RuntimeException("Google Analytics property ID must be numeric: {$propertyId}");
        }

        $privateKey = str_replace('\n', "\n", $privateKey);

        if (! openssl_pkey_get_private($privateKey)) {
            throw new RuntimeException("Google Analytics private_key is not a valid private key: {$path}");
        }

        try {
            CarbonImmutable::now($timezone);
        } catch (\Throwable) {
            throw new RuntimeException("Google Analytics timezone must be a valid timezone identifier: {$path}");
        }

        return [$propertyId, $clientEmail, $privateKey, $timezone, $tokenUri];
    }

    public function propertyDateRangeForDays(int $days, string $timezone): array
    {
        if ($days < 1) {
            throw new RuntimeException('--days must be an integer greater than or equal to 1.');
        }

        $todayStart = CarbonImmutable::now($timezone)->startOfDay();

        return [$todayStart->subDays($days - 1), $todayStart];
    }

    public function propertyHourAnchoredRangeForDays(int $days, string $timezone): array
    {
        if ($days < 1) {
            throw new RuntimeException('--days must be an integer greater than or equal to 1.');
        }

        // The API range is filtered with an exclusive upper bound in local time.
        $until = StatWindow::hourlyEnd(CarbonImmutable::now($timezone))->addHour();

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

    public function runReport(string $accessToken, string $propertyId, array $body): array
    {
        $url = "https://analyticsdata.googleapis.com/v1beta/properties/{$propertyId}:runReport";

        try {
            $response = Http::withToken($accessToken)
                ->acceptJson()
                ->asJson()
                ->timeout(20)
                ->post($url, $body);
        } catch (\Throwable $e) {
            throw new RuntimeException('Google Analytics Data API request failed before response: '.$e->getMessage(), 0, $e);
        }

        if (! $response->ok()) {
            throw new RuntimeException("Google Analytics Data API request failed: HTTP {$response->status()} {$response->body()}");
        }

        $payload = $response->json();
        if (! is_array($payload)) {
            throw new RuntimeException('Google Analytics Data API returned invalid JSON.');
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
        if (! preg_match('/^\d{8}$/', $raw)) {
            return null;
        }

        return substr($raw, 0, 4).'-'.substr($raw, 4, 2).'-'.substr($raw, 6, 2);
    }

    public function parseDateHourDimension(string $raw, string $timezone): ?CarbonImmutable
    {
        if (! preg_match('/^\d{10}$/', $raw)) {
            return null;
        }

        try {
            $timeslot = CarbonImmutable::createFromFormat('YmdH', $raw, $timezone);
        } catch (\Throwable) {
            return null;
        }

        if (! $timeslot) {
            return null;
        }

        return $timeslot->startOfHour();
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
            throw new RuntimeException('Google Analytics private_key is not a valid private key.');
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
