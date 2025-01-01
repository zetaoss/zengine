<?php
namespace App\Jobs;

use App\Models\CommonReport;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Log;

class CommonReportJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    protected $report;
    protected $result = null;
    public $tries = 1;

    public function __construct($report_id)
    {
        Log::debug("report_id: $report_id");
        $this->report = CommonReport::find($report_id);
    }

    public function handle(): void
    {
        $this->report->state = 1;
        $this->report->save();
        $this->processReport();
        $this->report->state = 2;
        $this->report->save();
    }

    private function processReport(): void
    {
        $items = $this->report->items;

        $qs = [];
        foreach ($items as $item) {
            $qs[] = '"' . $item->name . '"';
            $item->daum_blog = $this->search('daum_blog', $item->name);
            $item->naver_blog = $this->search('naver_blog', $item->name);
            $item->naver_book = $this->search('naver_book', $item->name);
            $item->naver_news = $this->search('naver_news', $item->name);
            $item->google_search = $this->search('google_search', $item->name);
            $item->updateTotal();
            $item->save();
        }
        Log::debug('processReport: done');
    }

    private function search($site_type, $word)
    {
        if ($site_type == 'daum_blog') {
            return self::search_daum('blog', $word);
        }
        if ($site_type == 'naver_blog') {
            return self::search_naver('blog', $word);
        }
        if ($site_type == 'naver_book') {
            return self::search_naver('book', $word);
        }
        if ($site_type == 'naver_news') {
            return self::search_naver('news', $word);
        }
        if ($site_type == 'google_search') {
            return self::search_google($word);
        }
        return false;
    }

    private function search_naver($type, $word)
    {
        $client_id = getenv('NAVER_CLIENT_ID');
        $client_secret = getenv('NAVER_CLIENT_SECRET');

        $encText = urlencode($word);
        $url = "https://openapi.naver.com/v1/search/$type.json?query=" . $encText;
        $is_post = false;
        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_POST, $is_post);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, ["X-Naver-Client-Id: $client_id", "X-Naver-Client-Secret: $client_secret"]);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, 0);
        $response = curl_exec($ch);
        $status_code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        if ($status_code != 200) {
            return false;
        }
        $result = json_decode($response);
        return $result->total;
    }

    private function search_daum($type, $word)
    {
        $apikey = getenv('KAKAO_API_KEY');

        $encText = urlencode($word);
        $url = "https://dapi.kakao.com/v2/search/$type?query=$encText";

        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, ["Authorization: KakaoAK $apikey"]);

        $response = curl_exec($ch);
        $status_code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        if ($status_code != 200) {
            return false;
        }
        $result = json_decode($response);
        return $result->meta->total_count;
    }

    private function search_google($word)
    {
        $endpoint = getenv('EXTERNAL_API_ENDPOINT_V3');

        $q = urlencode("\"$word\"");
        $url = "$endpoint/googling/a.php?q=$q";
        $result = json_decode(file_get_contents($url), true);
        if ($result['code'] != 200) {
            return 0;
        }
        return $result['cnt'];
    }
}
