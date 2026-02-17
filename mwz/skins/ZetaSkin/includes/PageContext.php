<?php

namespace ZetaSkin;

use MediaWiki\MediaWikiServices;
use ObjectCache;
use OutputPage;

final class PageContext
{
    private const KEY = 'PageContext';

    private const TTL = 3600;

    public ?array $avatar = null;

    public array $binders = [];

    public array $contributors = [];

    public bool $isView;

    public bool $hasBinders = false;

    public string $revtime = '';

    public int $pageId = 0;

    public static function getInstance(OutputPage $out): self
    {
        $cached = $out->getProperty(self::KEY);
        if ($cached instanceof self) {
            return $cached;
        }

        $ctx = new self($out);
        $out->setProperty(self::KEY, $ctx);

        return $ctx;
    }

    private function __construct(OutputPage $out)
    {
        $this->isView = ((string) $out->getActionName() === 'view');
        $this->revtime = (string) $out->getRevisionTimestamp();

        $title = $out->getTitle();
        $this->pageId = ($title && $title->canExist()) ? (int) $title->getId() : 0;

        if ($this->isView && $this->pageId > 0) {
            $this->binders = $this->fetchBinders($this->pageId);
            $this->hasBinders = ! empty($this->binders);

            $this->contributors = $this->fetchContributors($this->pageId, 10);
        }

        $meId = (int) ($out->getUser()?->getId() ?? 0);
        $this->avatar = $this->fetchUserAvatar($meId);
    }

    private static function dbr()
    {
        $svc = MediaWikiServices::getInstance();

        return $svc->getDBLoadBalancer()->getConnection(DB_REPLICA);
    }

    private static function avatarCacheKey(int $userId): string
    {
        return "avatar:$userId";
    }

    public static function forgetAvatar(int $userId): void
    {
        if ($userId < 1) {
            return;
        }

        $cache = ObjectCache::getLocalClusterInstance();
        $cache->delete(self::avatarCacheKey($userId));
    }

    private function fetchBinders(int $pageId): array
    {
        if ($pageId < 1) {
            return [];
        }

        $dbr = self::dbr();

        $binderIds = $dbr->newSelectQueryBuilder()
            ->select('binder_id')
            ->from('ldb.binder_pages')
            ->where(['page_id' => $pageId])
            ->caller(__METHOD__)
            ->fetchFieldValues();

        if (! $binderIds) {
            return [];
        }

        $res = $dbr->newSelectQueryBuilder()
            ->select(['id', 'data'])
            ->from('ldb.binders')
            ->where(['id' => $binderIds, 'deleted' => 0, 'enabled' => 1])
            ->caller(__METHOD__)
            ->fetchResultSet();

        $out = [];
        foreach ($res as $row) {
            $data = $row->data ? json_decode($row->data, true) : null;
            if (is_array($data) && $data) {
                $out[] = $data;
            }
        }

        return $out;
    }

    private function fetchContributors(int $pageId, int $limit = 10): array
    {
        if ($pageId < 1 || $limit < 1) {
            return [];
        }

        $dbr = self::dbr();

        $rows = $dbr->newSelectQueryBuilder()
            ->select(['U.user_id', 'U.user_name'])
            ->from('revision', 'R')
            ->join('actor', 'AC', 'R.rev_actor = AC.actor_id')
            ->join('user', 'U', 'AC.actor_user = U.user_id')
            ->where(['R.rev_page' => $pageId])
            ->orderBy('R.rev_timestamp', 'DESC')
            ->limit($limit * 5)
            ->caller(__METHOD__)
            ->fetchResultSet();

        $seen = [];
        $out = [];
        foreach ($rows as $row) {
            $id = (int) ($row->user_id ?? 0);
            if ($id < 1 || isset($seen[$id])) {
                continue;
            }
            $seen[$id] = true;
            $out[] = ['id' => $id, 'name' => (string) ($row->user_name ?? '')];
            if (count($out) >= $limit) {
                break;
            }
        }

        return $out;
    }

    private function fetchUserAvatarsByIds(array $userIds): array
    {
        $userIds = array_values(array_unique(array_filter(array_map('intval', $userIds))));
        $userIds = array_values(array_filter($userIds, fn ($v) => $v > 0));
        if (! $userIds) {
            return [];
        }

        $cache = ObjectCache::getLocalClusterInstance();

        $keysById = [];
        foreach ($userIds as $id) {
            $keysById[$id] = self::avatarCacheKey($id);
        }

        $cachedByKey = $cache->getMulti(array_values($keysById)) ?? [];

        $resultById = [];
        $missIds = [];

        foreach ($keysById as $id => $key) {
            $val = $cachedByKey[$key] ?? false;
            if ($val !== false && is_array($val)) {
                $resultById[$id] = $val;
            } else {
                $missIds[] = $id;
            }
        }

        if ($missIds) {
            $dbr = self::dbr();

            $rows = $dbr->newSelectQueryBuilder()
                ->select(['A.user_id', 'A.user_name', 'B.t', 'B.ghash'])
                ->from('user', 'A')
                ->leftJoin('ldb.profiles', 'B', 'A.user_id = B.user_id')
                ->where(['A.user_id' => $missIds])
                ->caller(__METHOD__)
                ->fetchResultSet();

            $toSet = [];

            foreach ($rows as $row) {
                $id = (int) ($row->user_id ?? 0);
                if ($id < 1) {
                    continue;
                }

                $avatar = [
                    'id' => $id,
                    'name' => (string) ($row->user_name ?? ''),
                    't' => (int) ($row->t ?? 1),
                    'ghash' => (string) ($row->ghash ?? ''),
                ];

                $resultById[$id] = $avatar;
                $toSet[self::avatarCacheKey($id)] = $avatar;
            }

            if ($toSet) {
                $cache->setMulti($toSet, self::TTL);
            }
        }

        $out = [];
        foreach ($userIds as $id) {
            if (isset($resultById[$id])) {
                $out[$id] = $resultById[$id];
            }
        }

        return $out;
    }

    private function fetchUserAvatar(int $userId): ?array
    {
        if ($userId < 1) {
            return null;
        }

        $cache = ObjectCache::getLocalClusterInstance();
        $key = self::avatarCacheKey($userId);

        $cached = $cache->get($key);
        if ($cached !== false && is_array($cached)) {
            return $cached;
        }

        $dbr = self::dbr();

        $row = $dbr->newSelectQueryBuilder()
            ->select(['A.user_id', 'A.user_name', 'B.t', 'B.ghash'])
            ->from('user', 'A')
            ->leftJoin('ldb.profiles', 'B', 'A.user_id = B.user_id')
            ->where(['A.user_id' => $userId])
            ->caller(__METHOD__)
            ->fetchRow();

        if (! $row) {
            return null;
        }

        $avatar = [
            'id' => (int) $row->user_id,
            'name' => (string) $row->user_name,
            't' => (int) ($row->t ?? 1),
            'ghash' => (string) ($row->ghash ?? ''),
        ];

        $cache->set($key, $avatar, self::TTL);

        return $avatar;
    }
}
