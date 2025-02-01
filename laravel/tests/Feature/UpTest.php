<?php

namespace Tests\Feature;

use Tests\TestCase;

class UpTest extends TestCase
{
    public function test_up(): void
    {
        $response = $this->get('/up');
        $response->assertStatus(200);
    }
}
