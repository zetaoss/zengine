<?php

namespace App\Http\Controllers;

use App\Models\PageReaction;
use App\Models\PageReactionUser;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class PageReactionController extends MyController
{
    private $emojis = ['👍', '😆', '😢', '😮', '❤️', '❤'];

    private $codes = [128077, 128518, 128546, 128558, 10084, 10084];

    private function emoji2code($emoji)
    {
        return $this->codes[array_search($emoji, $this->emojis)] ?? 0;
    }

    private function code2emoji($code)
    {
        return $this->emojis[array_search($code, $this->codes)] ?? '';
    }

    public function show($pageID)
    {
        $pr = PageReaction::where('page_id', $pageID)->first();
        if (! $pr) {
            return (object) [];
        }
        $emojiCount = $pr->emoji_count;
        $userEmojis = [];
        if ($this->getUserId()) {
            $rows = PageReactionUser::where([['page_id', $pageID], ['user_id', $this->getUserID()]])->get();
            foreach ($rows as $row) {
                $userEmojis[] = $this->code2emoji($row->emoji_code);
            }
        }

        return [compact('emojiCount', 'userEmojis')];
    }

    public function store(Request $request)
    {
        $request->validate([
            'pageid' => 'required|int|min:1',
            'emoji' => 'in:'.implode(',', $this->emojis),
            'enable' => 'required|boolean',
        ]);
        $err = $this->shouldCreatable();
        if ($err !== false) {
            return $err;
        }
        $pageID = request('pageid');
        $userID = $this->getUserID();
        $emojiCode = $this->emoji2code(request('emoji'));
        $row = [
            'page_id' => $pageID,
            'user_id' => $userID,
            'emoji_code' => $emojiCode,
        ];
        if ($request->boolean('enable')) {
            PageReactionUser::updateOrCreate($row)->save();
        } else {
            PageReactionUser::where($row)->delete();
        }
        $rows = PageReactionUser::where('page_id', $pageID)
            ->groupBy('emoji_code')->select('emoji_code', DB::raw('COUNT(*) as cnt'))->get();
        $cnt = 0;
        $emoji_count = [];
        foreach ($rows as $row) {
            $emoji_count[$this->code2emoji($row->emoji_code)] = $row->cnt;
            $cnt += $row->cnt;
        }
        $pr = PageReaction::where('page_id', $pageID)->first();
        if (! $pr) {
            $pr = new PageReaction;
        }
        $pr->page_id = $pageID;
        $pr->cnt = $cnt;
        $pr->emoji_count = $emoji_count;
        $pr->save();

        return ['status' => 'ok'];
    }
}
