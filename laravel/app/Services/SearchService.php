<?php

namespace App\Services;

use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;

class SearchService
{
    private function getConfig(): array
    {
        $naverHeaders = ['X-Naver-Client-Id' => getenv('NAVER_CLIENT_ID'), 'X-Naver-Client-Secret' => getenv('NAVER_CLIENT_SECRET')];
        return [
            'daum_blog' => [
                'url' => "https://dapi.kakao.com/v2/search/blog",
                'headers' => ["Authorization" => "KakaoAK " . getenv('KAKAO_API_KEY')],
                'responsePath' => 'meta.total_count',
                'exactMatch' => false,
            ],
            'naver_blog' => [
                'url' => "https://openapi.naver.com/v1/search/blog.json",
                'headers' => $naverHeaders,
                'responsePath' => 'total',
                'exactMatch' => false,
            ],
            'naver_book' => [
                'url' => "https://openapi.naver.com/v1/search/book.json",
                'headers' => $naverHeaders,
                'responsePath' => 'total',
                'exactMatch' => false,
            ],
            'naver_news' => [
                'url' => "https://openapi.naver.com/v1/search/news.json",
                'headers' => $naverHeaders,
                'responsePath' => 'total',
                'exactMatch' => false,
            ],
            'google_search' => [
                'url' => getenv('GOOGLE_SEARCH_URL'),
                'headers' => [],
                'responsePath' => 'total',
                'exactMatch' => true,
            ],
        ];
    }

    public function getTypes(): array
    {
        return array_keys($this->getConfig());
    }

    public function search(string $type, string $word): array
    {
        if (empty($type) || empty($word)) {
            return [
                'success' => false,
                'error' => 'Search type and word must be provided.',
            ];
        }

        $config = $this->getConfig()[$type] ?? null;
        if (!$config) {
            return [
                'success' => false,
                'error' => "Unsupported search type: $type",
            ];
        }

        if (!empty($config['exactMatch']) && $config['exactMatch'] === true) {
            $word = '"' . $word . '"';
        }

        try {
            $response = Http::withHeaders($config['headers'])
                ->get($config['url'], ['query' => $word]);

            if ($response->failed()) {
                Log::error("HTTP request failed", [
                    'status' => $response->status(),
                    'body' => $response->body(),
                    'type' => $type,
                    'word' => $word,
                ]);
                return [
                    'success' => false,
                    'error' => 'HTTP request failed.',
                ];
            }

            $data = $response->json();
            $value = $this->getValueFromPath($data, $config['responsePath']);

            if ($value === null) {
                return [
                    'success' => false,
                    'error' => 'Failed to retrieve the expected value from the response.',
                ];
            }

            return [
                'success' => true,
                'data' => $value,
            ];
        } catch (\Exception $e) {
            Log::error("An error occurred during the search operation", [
                'type' => $type,
                'word' => $word,
                'exception' => $e->getMessage(),
            ]);
            return [
                'success' => false,
                'error' => 'An unexpected error occurred.',
            ];
        }
    }

    private function getValueFromPath(array $data, string $path)
    {
        return array_reduce(explode('.', $path), function ($carry, $key) {
            return $carry[$key] ?? null;
        }, $data);
    }
}
