<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;

class BinderController extends Controller
{
    private const REDIRECT_MAX_HOPS = 9;

    public function index()
    {
        $rows = DB::table('ldb.binders as b')
            ->leftJoin('zetawiki.page as p', 'b.id', '=', 'p.page_id')
            ->select([
                'b.id',
                DB::raw('COALESCE(p.page_title, "") as title'),
                'b.docs',
                'b.links',
                'b.title_doc',
                'b.enabled',
                'b.created_at',
            ])
            ->orderByDesc('b.enabled')
            ->orderByDesc('b.docs')
            ->orderBy('p.page_title')
            ->orderBy('b.id')
            ->get();

        return response()->json($rows->map(static function ($row) {
            return [
                'id' => (int) ($row->id ?? 0),
                'title' => (string) ($row->title ?? ''),
                'docs' => (int) ($row->docs ?? 0),
                'links' => (int) ($row->links ?? 0),
                'title_doc' => (string) ($row->title_doc ?? ''),
                'enabled' => (int) ($row->enabled ?? 0) === 1,
                'created_at' => (string) ($row->created_at ?? ''),
            ];
        }));
    }

    public function update(int $binderId, Request $request)
    {
        Gate::authorize('sysop');

        $validated = $request->validate([
            'enabled' => ['required', 'boolean'],
        ]);

        $exists = DB::table('ldb.binders')
            ->where('id', $binderId)
            ->exists();

        if (! $exists) {
            return response()->json(['message' => 'Binder not found'], 404);
        }

        $realBinderId = $this->resolveRealBinderId($binderId);
        $updatedBinderId = $binderId;
        $deletedBinderId = null;
        $replacementTitle = null;
        $updatedEnabled = (bool) $validated['enabled'];

        DB::transaction(function () use ($binderId, $realBinderId, $validated, &$updatedBinderId, &$deletedBinderId, &$replacementTitle, &$updatedEnabled) {
            if ($realBinderId > 0 && $realBinderId !== $binderId) {
                $deletedBinderId = $binderId;
                $updatedBinderId = $realBinderId;
                $replacementTitle = $this->getPageTitle($realBinderId);

                if ($this->binderExists($realBinderId)) {
                    $this->deleteBinder($binderId);
                    $updatedEnabled = $this->getBinderEnabled($realBinderId);

                    return;
                }

                $this->transferBinder($binderId, $realBinderId, (bool) $validated['enabled']);

                return;
            }

            DB::table('ldb.binders')
                ->where('id', $updatedBinderId)
                ->update([
                    'enabled' => $validated['enabled'] ? 1 : 0,
                ]);
        });

        return [
            'ok' => true,
            'id' => $updatedBinderId,
            'enabled' => $updatedEnabled,
            'deleted_id' => $deletedBinderId,
            'replacement_title' => $replacementTitle,
        ];
    }

    private function resolveRealBinderId(int $binderId): int
    {
        $currentId = $binderId;

        for ($i = 0; $i < self::REDIRECT_MAX_HOPS; $i++) {
            $page = DB::table('zetawiki.page')
                ->select(['page_id'])
                ->where('page_id', $currentId)
                ->first();

            if (! $page) {
                return 0;
            }

            $redirect = DB::table('zetawiki.redirect')
                ->select(['rd_namespace', 'rd_title'])
                ->where('rd_from', $currentId)
                ->first();

            if (! $redirect) {
                return $currentId;
            }

            $targetId = DB::table('zetawiki.page')
                ->where('page_namespace', $redirect->rd_namespace)
                ->where('page_title', $redirect->rd_title)
                ->value('page_id');

            if (! $targetId) {
                return 0;
            }

            $currentId = (int) $targetId;
        }

        return 0;
    }

    private function deleteBinder(int $binderId): void
    {
        DB::table('ldb.binder_pages')
            ->where('binder_id', $binderId)
            ->delete();

        DB::table('ldb.binders')
            ->where('id', $binderId)
            ->delete();
    }

    private function transferBinder(int $fromBinderId, int $toBinderId, bool $enabled): void
    {
        DB::table('ldb.binder_pages')
            ->where('binder_id', $fromBinderId)
            ->update(['binder_id' => $toBinderId]);

        DB::table('ldb.binders')
            ->where('id', $fromBinderId)
            ->update([
                'id' => $toBinderId,
                'enabled' => $enabled ? 1 : 0,
            ]);
    }

    private function binderExists(int $binderId): bool
    {
        return DB::table('ldb.binders')
            ->where('id', $binderId)
            ->exists();
    }

    private function getBinderEnabled(int $binderId): bool
    {
        return (int) DB::table('ldb.binders')
            ->where('id', $binderId)
            ->value('enabled') === 1;
    }

    private function getPageTitle(int $pageId): ?string
    {
        $title = DB::table('zetawiki.page')
            ->where('page_id', $pageId)
            ->value('page_title');

        return is_string($title) ? $title : null;
    }
}
