<?php

namespace Tests\Feature;

use Illuminate\Foundation\Testing\WithoutMiddleware;
use Illuminate\Support\Facades\DB;
use Tests\TestCase;

class SocialDeletionControllerTest extends TestCase
{
    use WithoutMiddleware;

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
                social_id TEXT NOT NULL,
                user_id INTEGER NULL,
                deletion_code TEXT NULL
            )
        ');

        config()->set('services.facebook.client_secret', 'test-facebook-secret');

    }

    public function test_facebook_deletion_rejects_invalid_signed_request(): void
    {
        $response = $this->post('/auth/deletion/facebook', [
            'signed_request' => 'invalid',
        ]);

        $response->assertStatus(400);
        $response->assertJson(['error' => 'invalid_signed_request']);
    }

    public function test_facebook_deletion_updates_user_social_and_returns_confirmation(): void
    {
        DB::table('zetawiki.user_social')->insert([
            'provider' => 'facebook',
            'social_id' => 'fb-user-123',
            'user_id' => 77,
            'deletion_code' => null,
        ]);

        $signedRequest = $this->buildSignedRequest('fb-user-123', 'test-facebook-secret');

        $response = $this->post('/auth/deletion/facebook', [
            'signed_request' => $signedRequest,
        ]);

        $response->assertStatus(200);
        $response->assertJsonStructure([
            'url',
            'confirmation_code',
            'deleted_links',
        ]);
        $response->assertJson(['deleted_links' => 1]);

        $code = (string) $response->json('confirmation_code');
        $this->assertMatchesRegularExpression('/^[a-f0-9]{32}$/', $code);

        $row = DB::table('zetawiki.user_social')
            ->where('provider', 'facebook')
            ->where('social_id', 'fb-user-123')
            ->first();

        $this->assertNotNull($row);
        $this->assertNull($row->user_id);
        $this->assertSame($code, $row->deletion_code);
    }

    public function test_facebook_deletion_status_returns_completed_for_valid_code(): void
    {
        DB::table('zetawiki.user_social')->insert([
            'provider' => 'facebook',
            'social_id' => 'fb-user-xyz',
            'user_id' => null,
            'deletion_code' => 'abc123xyz',
        ]);

        $response = $this->get('/auth/deletion/facebook/status/abc123xyz');

        $response->assertStatus(200);
        $response->assertJson([
            'status' => 'completed',
            'confirmation_code' => 'abc123xyz',
            'deleted_links' => 1,
        ]);
    }

    public function test_deletion_route_rejects_non_facebook_provider(): void
    {
        $response = $this->post('/auth/deletion/google', [
            'signed_request' => 'anything',
        ]);

        $response->assertStatus(404);
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
