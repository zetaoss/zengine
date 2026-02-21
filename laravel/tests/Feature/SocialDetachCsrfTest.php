<?php

namespace Tests\Feature;

use Illuminate\Support\Facades\DB;
use Tests\TestCase;

class SocialDetachCsrfTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        config()->set('database.default', 'sqlite');
        config()->set('database.connections.sqlite.database', ':memory:');
        DB::purge('sqlite');
        DB::reconnect('sqlite');

        DB::statement("ATTACH DATABASE ':memory:' AS zetawiki");
        DB::statement('
            CREATE TABLE zetawiki.user_social (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                provider TEXT NOT NULL,
                social_id TEXT NULL,
                user_id INTEGER NULL,
                deletion_code TEXT NULL,
                deauthorized_at TEXT NULL,
                deleted_at TEXT NULL
            )
        ');

        config()->set('services.facebook.client_secret', 'test-facebook-secret');
    }

    public function test_facebook_deauthorize_callback_is_not_blocked_by_csrf(): void
    {
        DB::table('zetawiki.user_social')->insert([
            'provider' => 'facebook',
            'social_id' => 'fb-user-csrf-123',
            'user_id' => 77,
        ]);

        $signedRequest = $this->buildSignedRequest('fb-user-csrf-123', 'test-facebook-secret');

        $response = $this->post('/auth/deauthorize/facebook', [
            'signed_request' => $signedRequest,
        ]);

        $response->assertStatus(200);
        $response->assertJson([
            'status' => 'ok',
        ]);
    }

    private function buildSignedRequest(string $userId, string $secret): string
    {
        $payload = json_encode([
            'algorithm' => 'HMAC-SHA256',
            'user_id' => $userId,
        ], JSON_UNESCAPED_UNICODE);

        $encodedPayload = rtrim(strtr(base64_encode((string) $payload), '+/', '-_'), '=');
        $sig = hash_hmac('sha256', $encodedPayload, $secret, true);
        $encodedSig = rtrim(strtr(base64_encode($sig), '+/', '-_'), '=');

        return $encodedSig.'.'.$encodedPayload;
    }
}
