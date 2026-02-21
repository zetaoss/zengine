<?php

namespace App\Http\Controllers;

use App\Models\PageReaction;
use App\Models\PageReactionUser;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;

class PageReactionController extends Controller
{
    private array $emojis = ['👍', '😆', '😢', '😮', '❤️', '❤'];
    private array $codes = [128077, 128518, 128546, 128558, 10084, 10084];

    private function emoji2code(string $emoji): int
    {
        $idx = array_search($emoji, $this->emojis, true);
        if ($idx === false) {
            return 0;
        }

        return (int) ($this->codes[$idx] ?? 0);
    }

    private function code2emoji(int $code): string
    {
        $idx = array_search($code, $this->codes, true);
        if ($idx === false) {
            return '';
        }

        return (string) ($this->emojis[$idx] ?? '');
    }

    public function show(int $pageID)
    {
        $pr = PageReaction::where('page_id', $pageID)->first();

        $emojiCount = $pr?->emoji_count ?? (object) [];
        $userEmojis = [];

        $userID = (int) auth()->id();
        if ($userID > 0) {
            $rows = PageReactionUser::where([
                ['page_id', $pageID],
                ['user_id', $userID],
            ])->get();

            foreach ($rows as $row) {
                $userEmojis[] = $this->code2emoji((int) $row->emoji_code);
            }
        }

        return [compact('emojiCount', 'userEmojis')];
    }

    public function store(Request $request)
    {
        Gate::authorize('unblocked');

        $request->validate([
            'pageid' => 'required|integer|min:1',
            'emoji' => 'required|in:'.implode(',', $this->emojis),
            'enable' => 'required|boolean',
        ]);

        $pageID = (int) $request->input('pageid');
        $userID = (int) auth()->id();
        $emojiCode = $this->emoji2code((string) $request->input('emoji'));

        if ($emojiCode <= 0) {
            abort(422, 'Invalid emoji');
        }

        $row = [
            'page_id' => $pageID,
            'user_id' => $userID,
            'emoji_code' => $emojiCode,
        ];

        if ($request->boolean('enable')) {
            PageReactionUser::updateOrCreate($row);
        } else {
            PageReactionUser::where($row)->delete();
        }

        $rows = PageReactionUser::where('page_id', $pageID)
            ->groupBy('emoji_code')
            ->select('emoji_code', DB::raw('COUNT(*) as cnt'))
            ->get();

        $cnt = 0;
        $emoji_count = [];

        foreach ($rows as $r) {
            $emoji = $this->code2emoji((int) $r->emoji_code);
            if ($emoji === '') {
                continue;
            }

            $emoji_count[$emoji] = (int) $r->cnt;
            $cnt += (int) $r->cnt;
        }

        $pr = PageReaction::firstOrNew(['page_id' => $pageID]);
        $pr->cnt = $cnt;
        $pr->emoji_count = $emoji_count;
        $pr->save();

        return ['ok' => true];
    }
}
