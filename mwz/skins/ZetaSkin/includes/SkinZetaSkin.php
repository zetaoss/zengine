<?php

namespace ZetaSkin;

class SkinZetaSkin extends SkinBlade
{
    private static $links;

    private static $sidebar;

    private static $hasBinders;

    public static function onBeforePageDisplay($out, $skin)
    {
        $out->addHTMLClasses($_COOKIE['theme'] ?? '');
        if (getenv('APP_ENV') == 'dev') {
            $hash = crc32(rand());
            $out->addHeadItem('css', "<link href='/w/skins/ZetaSkin/resources/dist/app.css?$hash' rel='stylesheet' />");
            $out->addScript("<script src='/w/skins/ZetaSkin/resources/dist/app.js?$hash'></script><script type='module' src='/@vite/client'></script>");
        } else {
            $out->addHeadItem('css', '<link href="/w/skins/ZetaSkin/resources/dist/app.css" rel="stylesheet" />');
            $out->addScript('<script src="/w/skins/ZetaSkin/resources/dist/app.js"></script>');
        }
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
        self::$hasBinders = count($binders) > 0;
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
        if (isset($sidebar['TOOLBOX']['whatlinkshere'])) {
            $sidebar['TOOLBOX']['whatlinkshere']['text'] = '역링크';
        }
        if (isset($sidebar['TOOLBOX']['upload'])) {
            $sidebar['TOOLBOX']['upload']['text'] = '업로드';
        }
        if (isset($sidebar['TOOLBOX']['specialpages'])) {
            $sidebar['TOOLBOX']['specialpages']['text'] = '특수문서';
        }
        self::$sidebar = $sidebar;
    }

    public function getTemplateData()
    {
        return parent::getTemplateData() + [
            // 'links' => self::$links, // TODO: remove this line before release
            'action' => $this->getActionName(),
            'hasBinders' => self::$hasBinders,
            'navs' => array_filter([
                'recentchanges' => array_values(self::$sidebar)[0][0] ?? false,
                'randompage' => array_values(self::$sidebar)[0][1] ?? false,
            ]),
            'userMenu' => array_filter([
                'login' => self::$links['user-menu']['login'] ?? false,
                'createaccount' => self::$links['user-menu']['createaccount'] ?? false,
                'profile' => self::$links['user-menu']['profile'] ?? false,
                'userpage' => self::$links['user-menu']['userpage'] ?? false,
                'mytalk' => self::$links['user-menu']['mytalk'] ?? false,
                'preferences' => self::$links['user-menu']['preferences'] ?? false,
                'watchlist' => self::$links['user-menu']['watchlist'] ?? false,
                'mycontris' => self::$links['user-menu']['mycontris'] ?? false,
                'upload' => self::$sidebar['TOOLBOX']['upload'] ?? false,
                'specialpages' => self::$sidebar['TOOLBOX']['specialpages'] ?? false,
                'logout' => self::$links['user-menu']['logout'] ?? false,
            ]),
            'pageBtns' => array_filter([
                'view' => self::$links['views']['view'] ?? false,
                'edit' => self::$links['views']['edit'] ?? false,
                'whatlinkshere' => self::$sidebar['TOOLBOX']['whatlinkshere'] ?? false,
                'watch' => self::$links['views']['watch'] ?? false,
                'unwatch' => self::$links['views']['unwatch'] ?? false,
                'talk' => self::$links['namespaces']['talk'] ?? false,
            ]),
            'pageMenu' => array_filter([
                'history' => self::$links['views']['history'] ?? false,
                'delete' => self::$links['actions']['delete'] ?? false,
                'move' => self::$links['actions']['move'] ?? false,
                'protect' => self::$links['actions']['protect'] ?? false,
                'print' => self::$sidebar['TOOLBOX']['print'] ?? false,
                'permalink' => self::$sidebar['TOOLBOX']['permalink'] ?? false,
                'info' => self::$sidebar['TOOLBOX']['info'] ?? false,
            ]),
        ];
    }
}
