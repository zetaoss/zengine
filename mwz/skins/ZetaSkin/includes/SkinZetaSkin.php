<?php

namespace ZetaSkin;

use SkinMustache;

class SkinZetaSkin extends SkinMustache
{
    private static $links;

    private static $sidebar;

    private static $isBinder;

    public static function onBeforePageDisplay($out, $skin)
    {
        $out->addHTMLClasses($_COOKIE['theme'] ?? '');
        $out->addHeadItem('css', '<link href="/w/skins/ZetaSkin/resources/dist/app.css" rel="stylesheet" />');
        $out->addScript('<script src="/w/skins/ZetaSkin/resources/dist/app.js"></script>');
    }

    public static function onMakeGlobalVariablesScript(array &$vars, $out)
    {
        $binders = [];
        $contributors = [];
        $lastmod = '';

        if ($vars['wgIsArticle'] && $vars['wgAction'] == 'view') {
            $binders = DataService::getBinders($vars['wgArticleId']) ?? [];
            $contributors = DataService::getContributors($vars['wgPageName']);
            $lastmod = $out->getRevisionTimestamp();
        }

        $vars['binders'] = $binders;
        self::$isBinder = ! empty($binders);
        $vars['contributors'] = $contributors;
        $vars['lastmod'] = $lastmod;
        $vars['avatar'] = DataService::getUserAvatar($vars['wgUserId'] ?? 0);
    }

    public static function onSkinTemplateNavigation__Universal($skinTemplate, &$links)
    {
        if (isset($links['user-menu']['userpage'])) {
            $links['user-menu']['profile'] = ['text' => '프로필', 'href' => '/user/profile'];
            $links['user-menu']['userpage']['text'] = '사용자 문서';
            $links['user-menu']['mytalk']['text'] = '사용자 토론';
        }
        self::$links = $links;
    }

    public static function onSidebarBeforeOutput($skin, &$sidebar)
    {
        $map = [
            'whatlinkshere' => '역링크',
            'upload' => '업로드',
            'specialpages' => '특수문서',
        ];

        foreach ($map as $key => $label) {
            if (isset($sidebar['TOOLBOX'][$key])) {
                $sidebar['TOOLBOX'][$key]['text'] = $label;
            }
        }

        self::$sidebar = $sidebar;
    }

    public function getTemplateData()
    {
        $data = parent::getTemplateData();

        $views = self::$links['views'] ?? [];
        $actions = self::$links['actions'] ?? [];
        $namespaces = self::$links['namespaces'] ?? [];
        $toolbox = self::$sidebar['TOOLBOX'] ?? [];
        $userMenu = self::$links['user-menu'] ?? [];

        $data['isView'] = $this->getActionName() === 'view';
        $data['isBinder'] = self::$isBinder;

        $data['arrayButtons'] = array_values(array_filter([
            'view' => $views['view'] ?? null,
            'edit' => $views['edit'] ?? null,
            'whatlinkshere' => $toolbox['whatlinkshere'] ?? null,
            'watch' => $views['watch'] ?? null,
            'unwatch' => $views['unwatch'] ?? null,
            'talk' => $namespaces['talk'] ?? null,
        ]));
        $data['arrayMenu'] = array_values(array_filter([
            'history' => $views['history'] ?? null,
            'delete' => $actions['delete'] ?? null,
            'move' => $actions['move'] ?? null,
            'protect' => $actions['protect'] ?? null,
            'print' => $toolbox['print'] ?? null,
            'permalink' => $toolbox['permalink'] ?? null,
            'info' => $toolbox['info'] ?? null,
        ]));

        $data['jsonTOC'] = json_encode($data['data-toc'] ?? []);
        $data['jsonUserMenu'] = json_encode(array_filter([
            'login' => $userMenu['login'] ?? null,
            'createaccount' => $userMenu['createaccount'] ?? null,
            'profile' => $userMenu['profile'] ?? null,
            'userpage' => $userMenu['userpage'] ?? null,
            'mytalk' => $userMenu['mytalk'] ?? null,
            'preferences' => $userMenu['preferences'] ?? null,
            'watchlist' => $userMenu['watchlist'] ?? null,
            'mycontris' => $userMenu['mycontris'] ?? null,
            'upload' => $toolbox['upload'] ?? null,
            'specialpages' => $toolbox['specialpages'] ?? null,
            'logout' => $userMenu['logout'] ?? null,
        ]));

        return $data;
    }
}
