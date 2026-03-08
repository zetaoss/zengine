<?php

namespace ZetaSkin;

use SkinMustache;

class SkinZetaSkin extends SkinMustache
{
    private static $menu = [];
    private static $action;
    private static $pageId;
    private static $binders;
    private static $dataToc;
    private static $lastModified;
    private static $isArticleView;

    public static function onBeforePageDisplay($out, $skin)
    {
        self::$action = $out->getActionName();
        self::$pageId = (int) $skin->getTitle()->getArticleID();
        $policy = ($_SERVER['HTTP_X_POLICY'] ?? '') === 'standard' ? 'standard' : 'strict';
        $theme = ($_COOKIE['theme'] ?? '') === 'dark' ? 'dark' : '';
        $out->addHTMLClasses($theme);
        $out->addHeadItem('zconf', '<script>window.ZCONF={"avatarBaseUrl":"'.AVATAR_BASE_URL.'","gaMeasurementId":"'.GA_MEASUREMENT_ID.'","adClient":"'.AD_CLIENT.'","adSlots":'.AD_SLOTS.',"policy":"'.$policy.'"};</script>');
        $out->addStyle('/w/skins/ZetaSkin/dist/app.css?'.ASSET_HASH);
        $out->addScriptFile('/w/skins/ZetaSkin/dist/app.js?'.ASSET_HASH);
    }

    public function getTemplateData()
    {
        $data = parent::getTemplateData();
        self::$lastModified = $data['data-last-modified']['timestamp'];
        self::$isArticleView = $data['is-article'] && self::$action == 'view';
        $data['hasMeta'] = self::$isArticleView;
        self::$binders = self::$isArticleView ? PageDataProvider::fetchBinders(self::$pageId) : [];
        $data['hasBinders'] = ! empty(self::$binders);

        self::$dataToc = $data['data-toc'] ?? [];
        $data['hasToc'] = ! empty(self::$dataToc);

        $data['pageButtons'] = $this->getPageButtons();

        return $data;
    }

    public static function onSkinTemplateNavigation__Universal($skinTemplate, &$links)
    {
        self::$menu['actions'] = $links['actions'];
        self::$menu['namespaces'] = $links['namespaces'];
        self::$menu['usermenu'] = $links['user-menu'];
        self::$menu['views'] = $links['views'];
    }

    public static function onSidebarBeforeOutput($skin, &$sidebar)
    {
        self::$menu['toolbox'] = $sidebar['TOOLBOX'];
    }

    public static function onMakeGlobalVariablesScript(array &$vars, $out)
    {
        $vars['binders'] = self::$binders;
        $vars['dataToc'] = self::$dataToc;
        $vars['contributors'] = self::$isArticleView ? PageDataProvider::fetchContributors(self::$pageId) : [];
        $vars['lastModified'] = self::$lastModified;
        $vars['menu'] = self::$menu;
    }

    private function getPageButtons(): array
    {
        $buttons = array_filter([
            'view' => self::$action === 'view' ? null : (self::$menu['views']['view'] ?? null),
            'edit' => self::$action === 'edit' ? null : (self::$menu['views']['edit'] ?? null),
            'viewsource' => self::$menu['actions']['viewsource'] ?? null,
            'whatlinkshere' => self::$menu['toolbox']['whatlinkshere'] ?? null,
            'talk' => self::$menu['namespaces']['talk'] ?? null,
        ]);

        foreach ($buttons as $key => &$button) {
            switch ($key) {
                case 'edit':
                    $button['id'] = 'ca-edit';
                    break;
                case 'whatlinkshere':
                    $button['text'] = '역링크';
                    break;
                case 'talk':
                    $button['id'] = 'ca-talk';
            }
            $button['title'] = $button['text'];
        }

        return array_values($buttons);
    }
}
