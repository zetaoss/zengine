<?php

namespace ZetaSkin;

use SkinMustache;
use ZetaExtension\Disambig\DisambigHooks;

class SkinZetaSkin extends SkinMustache
{
    private static $menu = [];
    private static $action;
    private static $pageId;
    private static $binders;
    private static $dataToc;
    private static $disambig;
    private static $lastModified;
    private static $isArticleView;

    public static function onBeforePageDisplay($out, $skin)
    {
        self::$action = $out->getActionName();
        self::$pageId = (int) $skin->getTitle()->getArticleID();
        $a = self::assets();
        $out->addHeadItem('assets', '<script>if((localStorage.theme||(matchMedia("(prefers-color-scheme: dark)").matches?"dark":""))==="dark")document.documentElement.classList.add("dark","skin-theme-clientpref-night")</script>'
            .'<script>window.ZCONF='.json_encode([
                'avatarBaseUrl' => (string) AVATAR_BASE_URL,
                'gaMeasurementId' => (string) GA_MEASUREMENT_ID,
                'adClient' => (string) AD_CLIENT,
                'adSlots' => array_values(AD_SLOTS),
                'policy' => ($_SERVER['HTTP_X_POLICY'] ?? '') === 'standard' ? 'standard' : 'strict',
            ], JSON_UNESCAPED_SLASHES | JSON_HEX_TAG).'</script>'
            .'<script src="/track.js?'.$a['track'].'" defer></script><script src="/w/skins/ZetaSkin/dist/'.$a['js'].'" defer></script>'
            .'<link rel="stylesheet" href="/w/skins/ZetaSkin/dist/'.$a['css'].'">');
    }

    private static function assets(): array
    {
        return require __DIR__.'/../dist/assets.php';
    }

    public function getTemplateData()
    {
        $data = parent::getTemplateData();
        self::$lastModified = $data['data-last-modified']['timestamp'];
        self::$isArticleView = $data['is-article'] && self::$action == 'view';
        $data['hasMeta'] = self::$isArticleView;
        self::$binders = self::$isArticleView ? PageDataProvider::fetchBinders(self::$pageId) : [];
        self::$disambig = self::$isArticleView ? PageDataProvider::fetchDisambig(self::$pageId) : null;
        $data['hasBinders'] = ! empty(self::$binders);
        $data['hasDisambig'] = ! empty(self::$disambig);

        self::$dataToc = $data['data-toc'] ?? [];
        $data['hasToc'] = ! empty(self::$dataToc);

        $data['html-toc-static'] = $data['hasToc'] ? StaticMarkup::toc(self::$dataToc) : '';
        $registration = $data['hasDisambig'] && class_exists(DisambigHooks::class)
            ? DisambigHooks::getRegistrationData($this->getOutput())
            : null;
        $data['html-disambig-static'] = StaticMarkup::disambig(self::$disambig, $registration, self::$pageId);

        $data['pageButtons'] = $this->getPageButtons();

        if ($this->getUser()->isAnon() && ! empty($data['html-body-content'])) {
            $data['html-body-content'] = preg_replace(
                '/\s*\|\s*<a\s+[^>]*limit=(?:250|500)[^>]*>(?:250|500)<\/a>/',
                '',
                $data['html-body-content']
            );
        }

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
        $vars['disambig'] = self::$disambig;
        $vars['contributors'] = self::$isArticleView ? PageDataProvider::fetchContributors(self::$pageId) : [];
        $vars['lastModified'] = self::$lastModified;
        $vars['menu'] = self::$menu;
    }

    private function getPageButtons(): array
    {
        $buttons = array_filter([
            'view' => self::$action === 'view' ? null : (self::$menu['views']['view'] ?? null),
            'edit' => self::$action === 'edit' ? null : (self::$menu['views']['edit'] ?? null),
            'viewsource' => self::$menu['views']['viewsource'] ?? null,
            'whatlinkshere' => self::$menu['toolbox']['whatlinkshere'] ?? null,
            'talk' => self::$menu['namespaces']['talk'] ?? null,
        ]);

        foreach ($buttons as $k => &$b) {
            switch ($k) {
                case 'edit':
                    $b['id'] = 'ca-edit';
                    break;
                case 'whatlinkshere':
                    $b['text'] = '역링크';
                    break;
                case 'talk':
                    $b['id'] = 'ca-talk';
            }
            $b['title'] = $b['text'];
        }

        return array_values($buttons);
    }
}
