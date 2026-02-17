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
        $vars['avatar'] = $ctx->avatar;
        $vars['binders'] = $ctx->binders;
        $vars['contributors'] = $ctx->contributors;
        $vars['revtime'] = $ctx->revtime;
    }

    public static function onSkinTemplateNavigation__Universal($skinTemplate, &$links)
    {
        self::$links = $links;
    }

    public static function onSidebarBeforeOutput($skin, &$sidebar)
    {
        self::$sidebar = $sidebar;
    }

    public function getTemplateData()
    {
        $data = parent::getTemplateData();

        $is_article = $data['is-article'];

        $ctx = PageContext::getInstance($this->getOutput());

        $data['hasMeta'] = $is_article && $ctx->isView;
        $data['hasBinders'] = $ctx->hasBinders;

        $dataToc = $data['data-toc'] ?? [];
        $data['hasToc'] = ! empty($dataToc);

        if ($data['is-anon'] && $is_article && $ctx->isView) {
            $data['ads'] = [
                'client' => getenv('ADSENSE_CLIENT'),
                'slotTop' => getenv('ADSENSE_SLOT_TOP'),
                'slotBottom' => getenv('ADSENSE_SLOT_BOTTOM'),
            ];
        }

        $this->getOutput()->addJsConfigVars('datatoc', $dataToc);
        $this->getOutput()->addJsConfigVars('mm', [
            'actions' => self::$links['actions'],
            'namespaces' => self::$links['namespaces'],
            'toolbox' => self::$sidebar['TOOLBOX'],
            'usermenu' => self::$links['user-menu'],
            'views' => self::$links['views'],
        ]);

        return $data;
    }
}
