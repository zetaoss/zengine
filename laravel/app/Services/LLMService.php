<?php

namespace App\Services;

use Illuminate\Support\Facades\Http;
use RuntimeException;

class LLMService
{
    private readonly bool $enabled;
    private readonly string $endpoint;
    private readonly int $timeout;

    public function __construct()
    {
        $this->enabled = config('services.llm.enabled');
        $this->endpoint = config('services.llm.endpoint');
        $this->timeout = config('services.llm.timeout');
    }

    public function chat(array $messages): string
    {
        return $this->chatCompletion($messages)['content'];
    }

    public function chatCompletion(array $messages): array
    {
        if (! $this->enabled) {
            throw new RuntimeException('LLM is disabled. Set LLM_ENABLED=true to enable.');
        }

        $normalized = [];
        foreach ($messages as $message) {
            if (! is_array($message)) {
                continue;
            }

            $role = $message['role'] ?? null;
            $content = $message['content'] ?? null;
            if (! is_string($role) || ! is_string($content)) {
                continue;
            }

            $role = trim($role);
            $content = trim($content);
            if ($role === '' || $content === '') {
                continue;
            }

            $normalized[] = [
                'role' => $role,
                'content' => $content,
            ];
        }

        if ($normalized === []) {
            throw new RuntimeException('LLM messages are empty.');
        }

        $response = Http::acceptJson()
            ->asJson()
            ->timeout($this->timeout)
            ->post($this->url(), [
                'messages' => $normalized,
            ]);

        if (! $response->ok()) {
            throw new RuntimeException($this->formatErrorMessage($response->status(), $response->body()));
        }

        $body = $response->json();
        if (! is_array($body)) {
            throw new RuntimeException('LLM returned invalid JSON.');
        }

        $content = data_get($body, 'choices.0.message.content')
            ?? data_get($body, 'choices.0.text')
            ?? data_get($body, 'output_text');

        if (! is_string($content) || trim($content) === '') {
            throw new RuntimeException('LLM response is missing generated content.');
        }

        $model = data_get($body, 'model');
        if (! is_string($model) || trim($model) === '') {
            throw new RuntimeException('LLM response is missing model.');
        }

        return [
            'content' => trim($content),
            'model' => trim($model),
        ];
    }

    private function url(): string
    {
        if ($this->endpoint === '') {
            throw new RuntimeException('Missing LLM_ENDPOINT environment variable.');
        }

        return $this->endpoint.'/v1/chat/completions';
    }

    private function formatErrorMessage(int $status, string $rawBody): string
    {
        $payload = json_decode($rawBody, true);
        if (is_array($payload)) {
            $error = $payload['error'] ?? null;
            if (is_array($error)) {
                $code = $error['code'] ?? $status;
                $message = $error['message'] ?? null;
                if (is_string($message) && trim($message) !== '') {
                    $codeText = is_scalar($code) ? (string) $code : (string) $status;

                    return "[{$codeText}] ".trim($message);
                }
            }
        }

        return "[{$status}] LLM request failed.";
    }
}
