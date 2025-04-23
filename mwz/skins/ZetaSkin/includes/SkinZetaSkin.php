<?php

namespace ZetaSkin;

use Html;
use SkinTemplate;
use Title;

class SkinZetaSkin extends SkinTemplate
{
    private static $links;

    private static $sidebar;

    private static $hasBinders;

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

    public function generateHTML()
    {
        $this->setupTemplateContext();
        $out = $this->getOutput();

        $html = $out->headElement($this);
        extract($this->getTemplateData());

        ob_start();
        require __DIR__.'/views/app.php';
        $html .= ob_get_clean();

        return $html.$out->tailElement($this);
    }

    public function getTemplateData()
    {
        $data = parent::getTemplateData();
        $data = array_combine(array_map(fn ($x) => str_replace('-', '_', $x), array_keys($data)), $data);
        $out = $this->getOutput();

        $bodyContent = $out->getHTML()."\n".Html::rawElement('div', ['class' => 'printfooter', 'data-nosnippet' => ''], $this->printSource());
        $newTalksHtml = $this->getNewtalks() ?: null;

        $data += [
            'array_indicators' => $this->getIndicatorsData($out->getIndicators()),
            'html_site_notice' => $this->getSiteNotice() ?: null,
            'html_user_message' => $newTalksHtml ? Html::rawElement('div', ['class' => 'usermessage'], $newTalksHtml) : null,
            'html_subtitle' => $this->prepareSubtitle(),
            'html_body_content' => $this->wrapHTML($out->getTitle(), $bodyContent),
            'html_categories' => $this->getCategories(),
            'html_after_content' => $this->afterContentHook(),
            'html_undelete_link' => $this->prepareUndeleteLink(),
            'html_user_language_attributes' => $this->prepareUserLanguageAttributes(),
            'link_mainpage' => Title::newMainPage()->getLocalURL(),
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

        foreach ($this->options['messages'] ?? [] as $message) {
            $data["msg_{$message}"] = $this->msg($message)->text();
        }

        return $data;
    }
}
