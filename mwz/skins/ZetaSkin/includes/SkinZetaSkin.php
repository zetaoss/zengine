<?php

namespace ZetaSkin;

use SkinMustache;

class SkinZetaSkin extends SkinMustache
{
    private static $links;

    private static $sidebar;

    public static function onBeforePageDisplay($out, $skin)
    {
        $out->addHTMLClasses($_COOKIE['theme'] ?? '');
        $out->addHeadItem('adsense', '<script async src="https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client='.getenv('ADSENSE_CLIENT').'" crossorigin="anonymous"></script>');
        $out->addStyle('/w/skins/ZetaSkin/dist/app.css?'.ASSET_HASH);
        $out->addScriptFile('/config.js?'.ASSET_HASH);
        $out->addScriptFile('/w/skins/ZetaSkin/dist/app.js?'.ASSET_HASH);
    }

    public static function onMakeGlobalVariablesScript(array &$vars, $out)
    {
        $ctx = PageContext::getInstance($out);
        $vars['binders'] = $ctx->binders;
        $vars['contributors'] = $ctx->contributors;
        $vars['lastmod'] = $ctx->lastmod;
        $vars['avatar'] = $ctx->avatar;
    }

    public static function onSkinTemplateNavigation__Universal($skinTemplate, &$links)
    {
        $overMap = [
            'user-menu' => [
                'login' => ['accesskey' => 'o'],
                'mytalk' => ['text' => '사용자 토론'],
                'userpage' => ['text' => '사용자 문서'],
                'watchlist' => ['accesskey' => 'l'],
            ],
            'views' => [
                'edit' => ['accesskey' => 'e'],
                'history' => ['accesskey' => 'h'],
            ],
        ];

        foreach ($overMap as $section => $items) {
            foreach ($items as $key => $props) {
                if (isset($links[$section][$key])) {
                    foreach ($props as $prop => $val) {
                        $links[$section][$key][$prop] = $val;
                    }
                }
            }
        }

        if (isset($links['user-menu']['userpage'])) {
            $links['user-menu']['profile'] = [
                'text' => '프로필',
                'href' => '/user/'.rawurlencode($skinTemplate->getUser()->getName()),
            ];
        }

        self::$links = $links;
    }

    public static function onSidebarBeforeOutput($skin, &$sidebar)
    {
        $textMap = [
            'specialpages' => '특수문서',
            'upload' => '업로드',
            'whatlinkshere' => '역링크',
        ];

        foreach ($textMap as $key => $text) {
            if (isset($sidebar['TOOLBOX'][$key])) {
                $sidebar['TOOLBOX'][$key]['text'] = $text;
            }
        }

        self::$sidebar = $sidebar;
    }

    public function getTemplateData()
    {
        $data = parent::getTemplateData();

        $is_article = $data['is-article'];
        $is_specialpage = $data['is-specialpage'];

        $ctx = PageContext::getInstance($this->getOutput());

        $data['hasMeta'] = $is_article && $ctx->isView;
        $data['hasBinders'] = $ctx->hasBinders;

        $views = self::$links['views'] ?? [];
        $data['historyHref'] = $views['history']['href'] ?? null;

        $actions = self::$links['actions'] ?? [];
        $namespaces = self::$links['namespaces'] ?? [];
        $toolbox = self::$sidebar['TOOLBOX'] ?? [];
        $userMenu = self::$links['user-menu'] ?? [];

        $data['arrayButtons'] = array_values(array_filter([
            'view' => $ctx->isView ? null : ($views['view'] ?? null),
            'edit' => $views['edit'] ?? null,
            'whatlinkshere' => $toolbox['whatlinkshere'] ?? null,
            'watch' => $views['watch'] ?? null,
            'unwatch' => $views['unwatch'] ?? null,
            'talk' => $namespaces['talk'] ?? null,
        ]));

        $data['hasToc'] = ! empty($data['data-toc']);

        if ($data['is-anon'] && $is_article && $ctx->isView) {
            $data['ads'] = [
                'client' => getenv('ADSENSE_CLIENT'),
                'slotTop' => getenv('ADSENSE_SLOT_TOP'),
                'slotBottom' => getenv('ADSENSE_SLOT_BOTTOM'),
            ];
        }

        $pageMenu = array_values(array_filter([
            'history' => $views['history'] ?? null,
            'delete' => $actions['delete'] ?? null,
            'move' => $actions['move'] ?? null,
            'protect' => $actions['protect'] ?? null,
            'print' => $toolbox['print'] ?? null,
            'permalink' => $toolbox['permalink'] ?? null,
            'info' => $toolbox['info'] ?? null,
        ]));

        $myMenu = array_filter([
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
        ]);

        $this->getOutput()->addJsConfigVars('dataToc', $data['data-toc'] ?? []);
        $this->getOutput()->addJsConfigVars('pageMenu', $pageMenu);
        $this->getOutput()->addJsConfigVars('myMenu', $myMenu);

        return $data;
    }
}
