<?php
namespace Tests\Feature;

use Tests\TestCase;

class RunboxControllerTest extends TestCase
{
    public function test0()
    {
        $response = $this->getJson('/api/runbox/1/nonexistent-hash');

        $response->assertStatus(200)
            ->assertExactJson(['state' => 0]);
    }
}
