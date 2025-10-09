<?php

namespace ZetaSkin;

use MediaWiki\MediaWikiServices;
use ObjectCache;
use OutputPage;

final class PageContext
{
    private static ?PageContext $instance = null;

    public bool $isView = false;

    public int $articleId = 0;

    public array $binders = [];

    public bool $hasBinders = false;

    public array $contributors = [];

    public string $lastmod = '';

    public ?array $avatar = null;

    public static function getInstance(OutputPage $out): PageContext
    {
        return self::$instance ??= new self($out);
    }

    private function __construct(OutputPage $out)
    {
        $title = $out->getTitle();
        $this->isView = ($out->getActionName() === 'view');
        $this->articleId = $title ? (int) $title->getId() : 0;

        if ($this->isView && $this->articleId > 0) {
            $this->binders = $this->fetchBinders($this->articleId);
            $this->hasBinders = ! empty($this->binders);
            $this->contributors = $this->fetchContributors($title->getPrefixedText());
            $this->lastmod = $out->getRevisionTimestamp();
        }

        $this->avatar = $this->fetchUserAvatar((int) ($out->getUser()?->getId() ?? 0));
    }

    private function fetchBinders(int $pageId): array
    {
        $http = MediaWikiServices::getInstance()->getHttpRequestFactory();
        $req = $http->create("http://localhost/w/rest.php/binder/$pageId", [], __METHOD__);
        if (! $req->execute()->isOK()) {
            return [];
        }
        $json = json_decode($req->getContent(), true);

        return is_array($json) ? $json : [];
    }

    private function fetchContributors(string $titleText): array
    {
        $http = MediaWikiServices::getInstance()->getHttpRequestFactory();
        $url = 'http://localhost/w/api.php?format=json&action=query&prop=contributors&titles='.rawurlencode($titleText);
        $req = $http->create($url, [], __METHOD__);
        if (! $req->execute()->isOK()) {
            return [];
        }
        $data = json_decode($req->getContent(), true);
        $pages = $data['query']['pages'] ?? [];
        $list = is_array($pages) ? (array) (array_pop($pages)['contributors'] ?? []) : [];

        return array_map(fn ($u) => $this->fetchUserAvatar((int) ($u['userid'] ?? 0)), $list);
    }

    private function fetchUserAvatar(int $userID): ?array
    {
        if ($userID === 0) {
            return null;
        }

        $cache = ObjectCache::getLocalClusterInstance();
        $key = "userAvatar:$userID";
        $cached = $cache->get($key);
        if ($cached !== false) {
            return $cached;
        }

        $dbr = MediaWikiServices::getInstance()
            ->getDBLoadBalancer()
            ->getConnection(DB_REPLICA);

        $row = $dbr->newSelectQueryBuilder()
            ->select(['user_name', 't', 'ghash'])
            ->from('user', 'A')
            ->leftJoin('ldb.profiles', 'B', 'A.user_id = B.user_id')
            ->where(['A.user_id' => $userID])
            ->caller(__METHOD__)
            ->fetchRow();

        if (! $row) {
            return null;
        }

        $avatar = [
            'id' => $userID,
            'name' => $row->user_name,
            't' => $row->t,
            'ghash' => $row->ghash,
        ];

        $cache->set($key, $avatar);

        return $avatar;
    }
}
