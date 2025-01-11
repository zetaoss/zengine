<?php
namespace App\Services;

use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;

class SearchService
{
    private function getServiceConfig(): array
    {
        $naverClientId = getenv('NAVER_CLIENT_ID') ?? throw new \RuntimeException("NAVER_CLIENT_ID is not set.");
        $naverClientSecret = getenv('NAVER_CLIENT_SECRET') ?? throw new \RuntimeException("NAVER_CLIENT_SECRET is not set.");
        $kakaoApiKey = getenv('KAKAO_API_KEY') ?? throw new \RuntimeException("KAKAO_API_KEY is not set.");
        $googleSearchKey = getenv('GOOGLE_SEARCH_KEY') ?? throw new \RuntimeException("GOOGLE_SEARCH_KEY is not set.");
        $googleSearchCx = getenv('GOOGLE_SEARCH_CX') ?? throw new \RuntimeException("GOOGLE_SEARCH_CX is not set.");

        $naverHeaders = [
            'X-Naver-Client-Id' => $naverClientId,
            'X-Naver-Client-Secret' => $naverClientSecret,
        ];

        return [
            'daum_blog' => [
                'url' => "https://dapi.kakao.com/v2/search/blog",
                'headers' => ["Authorization" => "KakaoAK $kakaoApiKey"],
                'queryKey' => 'query',
                'responsePath' => 'meta.total_count',
            ],
            'naver_blog' => [
                'url' => "https://openapi.naver.com/v1/search/blog.json",
                'headers' => $naverHeaders,
                'queryKey' => 'query',
                'responsePath' => 'total',
            ],
            'naver_book' => [
                'url' => "https://openapi.naver.com/v1/search/book.json",
                'headers' => $naverHeaders,
                'queryKey' => 'query',
                'responsePath' => 'total',
            ],
            'naver_news' => [
                'url' => "https://openapi.naver.com/v1/search/news.json",
                'headers' => $naverHeaders,
                'queryKey' => 'query',
                'responsePath' => 'total',
            ],
            'google_search' => [
                'url' => "https://www.googleapis.com/customsearch/v1",
                'headers' => [],
                'queryKey' => 'q',
                'queryParams' => [
                    'key' => $googleSearchKey,
                    'cx' => $googleSearchCx,
                ],
                'responsePath' => 'searchInformation.totalResults',
            ],
        ];
    }

    public function search(string $type, string $word)
    {
        if (empty($type)) {
            throw new \InvalidArgumentException("The search type must be provided.");
        }

        if (empty($word)) {
            throw new \InvalidArgumentException("The search word cannot be empty.");
        }

        $config = $this->getServiceConfig()[$type] ?? null;

        if (!$config) {
            throw new \InvalidArgumentException("Unsupported search type: $type");
        }

        try {
            $response = Http::withHeaders($config['headers'])
                ->get($config['url'], [$config['queryKey'] => $word]);

            if ($response->failed()) {
                Log::error("HTTP request failed: {$response->status()} - {$response->body()}");
                return -1;
            }

            return $this->getValueFromPath($response->json(), $config['responsePath']) ?? -2;
        } catch (\Exception $e) {
            Log::error("An error occurred during the search operation: " . $e->getMessage(), [
                'type' => $type,
                'word' => $word,
                'trace' => $e->getTraceAsString(),
            ]);
            throw $e;
        }
    }

    private function getValueFromPath(array $data, string $path)
    {
        $result = array_reduce(explode('.', $path), function ($carry, $key) {
            return $carry[$key] ?? null;
        }, $data);

        if ($result === null) {
            Log::warning("Failed to traverse response path: $path", ['data' => $data]);
        }

        return $result;
    }
}
