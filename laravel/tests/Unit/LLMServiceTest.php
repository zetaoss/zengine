<?php

use App\Services\LLMService;
use Illuminate\Http\Client\Request;
use Illuminate\Support\Facades\Http;

it('sends chat completions without a model and returns the response model', function () {
    config([
        'services.llm.enabled' => true,
        'services.llm.endpoint' => 'https://llm.example.test',
        'services.llm.timeout' => 15,
    ]);

    Http::fake([
        'https://llm.example.test/v1/chat/completions' => function (Request $request) {
            expect($request['messages'])->toBe([
                ['role' => 'user', 'content' => 'hello'],
            ]);
            expect($request->data())->not->toHaveKey('model');

            return Http::response([
                'model' => 'gpt-auto-selected',
                'choices' => [
                    [
                        'message' => [
                            'content' => 'world',
                        ],
                    ],
                ],
            ]);
        },
    ]);

    $result = app(LLMService::class)->chatCompletion([
        ['role' => 'user', 'content' => 'hello'],
    ]);

    expect($result)->toBe([
        'content' => 'world',
        'model' => 'gpt-auto-selected',
    ]);
});

it('throws when the llm response does not include a model', function () {
    config([
        'services.llm.enabled' => true,
        'services.llm.endpoint' => 'https://llm.example.test',
        'services.llm.timeout' => 15,
    ]);

    Http::fake([
        'https://llm.example.test/v1/chat/completions' => Http::response([
            'choices' => [
                [
                    'message' => [
                        'content' => 'world',
                    ],
                ],
            ],
        ]),
    ]);

    expect(fn () => app(LLMService::class)->chatCompletion([
        ['role' => 'user', 'content' => 'hello'],
    ]))->toThrow(RuntimeException::class, 'LLM response is missing model.');
});
